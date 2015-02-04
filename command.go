package cmdr
import (
  "reflect"
  "fmt"
  "flag"
  "strconv"
  "strings"
  "bytes"
  "unicode"
  "unicode/utf8"
  "regexp"
  "os"
)


type Command struct {
  Name        string
  Description string
  Options     *flag.FlagSet
  OptionCount int
  Args        []Argument
  VarArgs     *Argument
  Program     *Program
  main        func(*Command)
}

type Argument struct {
  Name        string
  Description string
  Optional    bool
  Value       ValueBinding
}

type ValueBinding interface {
  String() string
  Set(string) error
}

type SliceBinding interface {
  ValueBinding
  Setv([]string) error
}


func (cmd *Command) Logf(format string, a ...interface{}) {
  if cmd.IsQuiet() { return }
  fmt.Fprintf(os.Stdout, format + "\n", a...)
}


func (cmd *Command) Log(a ...interface{}) {
  if cmd.IsQuiet() { return }
  fmt.Fprint(os.Stdout, a...)
  os.Stdout.WriteString("\n")
}


func (cmd *Command) IsQuiet() bool {
  if cmd.Program != nil {
    return cmd.Program.IsQuiet
  }
  return false
}


// Prints a message from `fmt.Sprint(msg...)` together with info on how to invoke help,
// finally calling os.Exit(1)
func (cmd *Command) Fail(msg ...interface{}) {
  if cmd.Program != nil {
    var see string
    if cmd.Options.Parsed() {
      see = cmd.Program.Name + " " + cmd.Name + " -help"
    } else {
      see = cmd.Program.Name + " -help"
    }
    fmt.Fprintf(os.Stderr,
      "%s %s: %v. See '%s'\n",
      cmd.Program.Name, cmd.Name, fmt.Sprint(msg...), see)
  } else {
    fmt.Fprint(os.Stderr, msg...)
  }
  os.Exit(1)
}


func (cmd *Command) Usage() {
  if len(cmd.Description) != 0 {
    os.Stderr.WriteString(cmd.Description + "\n")
  }
  os.Stderr.WriteString("Usage: ")
  if cmd.Program != nil {
    os.Stderr.WriteString(cmd.Program.Name + " ")
  }

  if cmd.OptionCount != 0 {
    fmt.Fprintf(os.Stderr, "%s [options]%s\n", cmd.Name, cmd.argsString())
    os.Stderr.WriteString("Options:\n")
    optionsUsage(cmd.Options)
  } else {
    fmt.Fprintf(os.Stderr, "%s%s\n", cmd.Name, cmd.argsString())
  }

  if len(cmd.Args) != 0 || cmd.VarArgs != nil {
    os.Stderr.WriteString("Arguments:\n")
    w := newTabWriter()

    for _, arg := range cmd.Args {
      defaultValue := arg.Value.String()
      if defaultValue != "" {
        if _, ok := arg.Value.(*stringValue); ok {
          fmt.Fprintf(w, "  <%s>\t%s (default: %q)\n", arg.Name, arg.Description, defaultValue)
        } else {
          fmt.Fprintf(w, "  <%s>\t%s (default: %s)\n", arg.Name, arg.Description, defaultValue)
        }
      } else {
        fmt.Fprintf(w, "  <%s>\t%s \n", arg.Name, arg.Description)
      }
    }

    if cmd.VarArgs != nil {
      fmt.Fprintf(w, "  <%s>...\t%s \n", cmd.VarArgs.Name, cmd.VarArgs.Description)
    }

    w.Flush()
  }
}


func (cmd *Command) NameAndArgs() string {
  return fmt.Sprintf("%s%s", cmd.Name, cmd.argsString())
}


func (cmd *Command) argsString() string {
  var buf bytes.Buffer
  for _, arg := range cmd.Args {
    if arg.Optional {
      fmt.Fprintf(&buf, " [<%s>]", arg.Name)
    } else {
      fmt.Fprintf(&buf, " <%s>", arg.Name)
    }
  }
  if cmd.VarArgs != nil {
    if cmd.VarArgs.Optional {
      fmt.Fprintf(&buf, " [<%s>...]", cmd.VarArgs.Name)
    } else {
      fmt.Fprintf(&buf, " <%s>...", cmd.VarArgs.Name)
    }
  }
  return buf.String()
}


func (cmd *Command) Parse(args []string) error {
  if err := cmd.Options.Parse(args); err != nil {
    return err
  }
  args = cmd.Options.Args()
  argVarCount := len(cmd.Args)
  argEnd := argVarCount
  if len(args) < argEnd {
    argEnd = len(args)
  }
  i := 0
  for ; i != argEnd; i++ {
    if err := cmd.Args[i].Value.Set(args[i]); err != nil {
      return err
    }
  }

  // varargs
  if cmd.VarArgs != nil && i != len(args) {
    if sliceValue, ok := cmd.VarArgs.Value.(SliceBinding); ok {
      if err := sliceValue.Setv(args[i:]); err != nil {
        return err
      }
    }
    // else return errors.New() maybe?
  }

  return nil
}


func (cmd *Command) Run(p *Program, args []string) {
  if cmd.main != nil {
    pp := cmd.Program
    cmd.Program = p
    defer func(){ cmd.Program = pp }()
    cmd.Parse(args)
    cmd.main(cmd)
  }
}


func (cmd *Command) countOptions() {
  cmd.OptionCount = 0
  cmd.Options.VisitAll(func(f *flag.Flag) { cmd.OptionCount++ })
}


var commandType = reflect.TypeOf(new(Command)).Elem()


func NewCommand(name, description string, f interface{}) *Command {
  fnv := reflect.ValueOf(f)
  fnt := fnv.Type()
  if fnt.Kind() != reflect.Func {
    panic("not a function")
  }

  cmd := &Command{
    Name:        name,
    Description: description,
    Options:     flag.NewFlagSet(name, flag.ExitOnError),
  }
  cmd.Options.Usage = func() { cmd.Usage() }

  if fnt.NumIn() == 0 {
    // taking no arguments and having no options
    if fn, ok := f.(func()); ok {
      cmd.main = func(_ *Command){ fn() }
      return cmd
    }
    panic("command function shouldn't return any values")
  }

  T := fnt.In(0).Elem()
  if T.Kind() != reflect.Struct {
    panic("Command.Options must be a struct")
  }

  // accepts command as a second argument?
  // if fnt.NumIn() == 2 {
  //   T2 := fnt.In(0).Elem()
  //   if T2 == commandType {
  //     println("TODO T2 is commandType")
  //   }
  // }

  cmdStructVPtr := reflect.New(T)
  cmd.main = func(cmd *Command) {
    fnv.Call([]reflect.Value{cmdStructVPtr, reflect.ValueOf(cmd)})
  }

  for fieldIndex := 0; fieldIndex != T.NumField(); fieldIndex++ {
    field := T.Field(fieldIndex)
    rune0, _ := utf8.DecodeRuneInString(field.Name)
    if unicode.IsUpper(rune0) {
      cmd.processField(&cmdStructVPtr, fieldIndex)
    }
  }

  cmd.countOptions()

  return cmd
}


func (cmd *Command) processField(stValuePtr *reflect.Value, fieldIndex int) {
  fieldV := stValuePtr.Elem().Field(fieldIndex)
  field := stValuePtr.Elem().Type().Field(fieldIndex)
  name := translateFieldName(field.Name)
  defaultValue, descr, prefix := parseFieldTag(string(field.Tag))
  if prefix == '!' {
    optional := false
    cmd.addArg(&fieldV, optional, name, defaultValue, descr)
  } else if prefix == '?' {
    optional := true
    cmd.addArg(&fieldV, optional, name, defaultValue, descr)
  } else {
    cmd.addOption(&fieldV, name, defaultValue, descr)
  }
}


func (cmd *Command) addArg(field *reflect.Value, optional bool, name, defaultValue, descr string) {
  // fmt.Printf("addArg(field=%v, optional=%v, name=%q, defaultValue=%q descr=%q)\n",
  //            field, optional, name, defaultValue, descr)
  val := NewValueBinding(field, defaultValue)
  if val != nil {
    if _, ok := val.(SliceBinding); ok {
      if cmd.VarArgs != nil {
        panic("multiple varargs in command struct")
      }
      cmd.VarArgs = &Argument{name, descr, optional, val}
    } else {
      cmd.Args = append(cmd.Args, Argument{name, descr, optional, val})
    }
  }
}


func (cmd *Command) addOption(field *reflect.Value, name, defaultValue, descr string) {
  // fmt.Printf("addOption(field=%v, name=%q, defaultValue=%q descr=%q)\n",
  //            field, name, defaultValue, descr)
  val := NewValueBinding(field, defaultValue)
  if val != nil {
    cmd.Options.Var(val, name, descr)
  }
}

// ===============================================================================================

func NewValueBinding(v *reflect.Value, defaultValue string) ValueBinding {
  var val ValueBinding
  switch v.Type().Kind() {
    case reflect.Bool:   val = (*boolValue)(v)
    case reflect.String: val = (*stringValue)(v)
    case reflect.Slice:  val = (*sliceValue)(v)
    default: return nil
  }
  if defaultValue != "" {
    val.Set(defaultValue)
  }
  return val
}


type ValueBinder func(*reflect.Value) ValueBinding

func NewValueBinder(T reflect.Type) ValueBinder {
  switch T.Kind() {
  case reflect.String: return func(v *reflect.Value) ValueBinding { return (*stringValue)(v) }
  case reflect.Bool: return func(v *reflect.Value) ValueBinding { return (*boolValue)(v) }
  case reflect.Slice: return func(v *reflect.Value) ValueBinding { return (*sliceValue)(v) }
  default: panic("unexpected value type")
  }
}


type stringValue reflect.Value
func (v *stringValue) rv() *reflect.Value { return (*reflect.Value)(v) }
func (v *stringValue) String() string { return v.rv().String() }
func (v *stringValue) Set(s string) error {
  v.rv().SetString(s)
  return nil
}


type boolValue reflect.Value
func (v *boolValue) rv() *reflect.Value { return (*reflect.Value)(v) }
func (v *boolValue) IsBoolFlag() bool { return true }
func (v *boolValue) String() string {
  if v.rv().Bool() {
    return "true"
  } else {
    return "false"
  }
}
func (v *boolValue) Set(s string) error {
  bv, err := strconv.ParseBool(s)
  v.rv().SetBool(bv)
  return err
}


type sliceValue reflect.Value
func (sv *sliceValue) rv() *reflect.Value { return (*reflect.Value)(sv) }
func (sv *sliceValue) String() string { return sv.rv().String() }
func (sv *sliceValue) Set(s string) error { panic("Set() on slice") }
func (sv *sliceValue) Setv(args []string) error {
  v := sv.rv()
  T := v.Type()
  sliceV := reflect.MakeSlice(T, len(args), len(args))
  valueBinder := NewValueBinder(T.Elem())
  for i, s := range args {
    val := sliceV.Index(i)
    if err := valueBinder(&val).Set(s); err != nil {
      return err
    }
  }
  v.Set(sliceV)
  return nil
}

// ===============================================================================================

func parseFieldTag(tag string) (defaultValue string, tail string, prefix byte) {
  tail = tag
  for tag != "" {
    prefix = tag[0]
    if prefix != '!' && prefix != '?' && prefix != '=' {
      break
    }

    i := 1

    if prefix == '=' {
      if len(tag) < 3 {
        break
      }
    }

    // TODO:
    //   - break out the below "Unquote string" code
    //   - add check for "[...]" and when found, build defaultValue=[]string by calling
    //     the broken-out "Unquote string" code on each component of the list

    if len(tag) > 1 && tag[1] == '"' {
      // Unquote value
      i++
      for i < len(tag) && tag[i] != '"' {
        if tag[i] == '\\' {
          i++
        }
        i++
      }
      if i >= len(tag) {
        break
      }
      i++
      defaultValue, _ = strconv.Unquote(tag[1:i])
    }

    tail = tag[i:]
    break
  }
  tail = strings.TrimSpace(tail)
  return defaultValue, tail, prefix
}


var optionNameRegex1 = regexp.MustCompile(
  `(\p{Lu}+[\p{Ll}\p{Lt}\p{Lm}\p{Lo}\p{Nd}]+)_|([\p{Lu}]+)([\p{Lu}])`)
var optionNameRegex2 = regexp.MustCompile(
  `(?:([^\-_])[\-_]+|(\p{Lu}[\p{Ll}\p{Lt}\p{Lm}\p{Lo}\p{Nd}]+)[\-_]*)`)

// e.g. "FooBar"                  => "foo-bar"
//      "Lol"                     => "lol"
//      "FOO"                     => "foo"
//      "FirstNameLOLCat"         => "first-name-lol-cat"
//      "FooBar_baz_CATz_LOLCaT"  => "foo-bar-baz-catz-lol-ca-t"
//      "Plan9From800Outer_space" => "plan9-from800-outer-space"
func translateFieldName(s string) string {
  s = optionNameRegex1.ReplaceAllString(s, "-$1$2-$3")
  s = optionNameRegex2.ReplaceAllString(s, "$1$2-")
  if s[len(s)-1] == '-' {
    s = s[:len(s)-1]
  }
  return strings.ToLower(s)
}
