package cmdr
import (
  "os"
  "flag"
  "time"
  "text/tabwriter"
  "fmt"
)


// Default program bound to os.Args and flag.CommandLine
var DefaultProgram = &Program{
  Name:        os.Args[0],
  Options:     flag.CommandLine,
  Usage:       ProgramUsage,
  ExitOnError: true,
}


// Create and add a new command to the DefaultProgram
func Cmd(name, description string, fn interface{}) *Command {
  cmd := NewCommand(name, description, fn)
  DefaultProgram.AddCommand(cmd)
  return cmd
}


// Convenience function, invoking `DefaultProgram.Main(os.Args[1:])`
// Parses args and runs a command. Returns the command run.
func Main() *Command {
  return DefaultProgram.Main(os.Args[1:])
}

// ==============================================================================================

func newTabWriter() *tabwriter.Writer {
  return tabwriter.NewWriter(os.Stderr, 5, 0, 3, ' ', 0)
}

func flagIsString(f *flag.Flag) bool {
  if _, ok := f.Value.(*stringValue); ok {
    return true
  } else if getter, ok := f.Value.(flag.Getter); ok {
    _, ok = getter.Get().(string)
    return ok
  }
  return false
}

func optionsUsage(flagSet *flag.FlagSet) {
  w := newTabWriter()

  flagSet.VisitAll(func(f *flag.Flag) {
    v := f.Value
    if _, ok := v.(*boolValue); ok {
      if f.DefValue == "false" {
        fmt.Fprintf(w, "  -%s\t%s\n", f.Name, f.Usage)
      } else {
        fmt.Fprintf(w, "  -%s=true\t%s\n", f.Name, f.Usage)
      }
    } else {
      if flagIsString(f) {
        fmt.Fprintf(w, "  -%s %q\t%s\n", f.Name, f.DefValue, f.Usage)
      } else {
        fmt.Fprintf(w, "  -%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
      }
    }
  })

  w.Flush()
}

// ==============================================================================================

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage string) {
  DefaultProgram.Options.BoolVar(p, name, value, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, value bool, usage string) *bool {
  return DefaultProgram.Options.Bool(name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
  DefaultProgram.Options.IntVar(p, name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
  return DefaultProgram.Options.Int(name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage string) {
  DefaultProgram.Options.Int64Var(p, name, value, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name string, value int64, usage string) *int64 {
  return DefaultProgram.Options.Int64(name, value, usage)
}


// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint  variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage string) {
  DefaultProgram.Options.UintVar(p, name, value, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func Uint(name string, value uint, usage string) *uint {
  return DefaultProgram.Options.Uint(name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
  DefaultProgram.Options.Uint64Var(p, name, value, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, value uint64, usage string) *uint64 {
  return DefaultProgram.Options.Uint64(name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
  DefaultProgram.Options.StringVar(p, name, value, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
  return DefaultProgram.Options.String(name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
  DefaultProgram.Options.Float64Var(p, name, value, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
  return DefaultProgram.Options.Float64(name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
  DefaultProgram.Options.DurationVar(p, name, value, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func Duration(name string, value time.Duration, usage string) *time.Duration {
  return DefaultProgram.Options.Duration(name, value, usage)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func Var(value flag.Value, name string, usage string) {
  DefaultProgram.Options.Var(value, name, usage)
}
