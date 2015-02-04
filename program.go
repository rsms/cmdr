package cmdr
import (
  "os"
  "flag"
  "fmt"
  "text/tabwriter"
  "sort"
)

type Program struct {
  // Name of the program
  Name           string

  // Program-global options
  Options        *flag.FlagSet

  // Commands provided by this program
  Commands       map[string]*Command

  // Optional command to be run when the program is invoked w/o a command name
  DefaultCommand *Command

  // Invoked to display usage for the program
  Usage          func(*Program)

  // The value of this variable affects Command.IsQuiet.
  // Can be used to provide a global -quiet or -verbose option.
  IsQuiet        bool

  // When true, os.Exit(1) is called upon failure
  ExitOnError    bool
}


// A command that shows program usage
var HelpCommand = NewCommand("help", "Show help", func (opt *struct {
  Command string `?`
}, cmd *Command) {
  if opt.Command == "help" {
    fmt.Fprintf(os.Stderr, "Usage: help <command>\n")
  } else {
    cmd2 := cmd.Program.Commands[opt.Command]
    if cmd2 != nil {
      cmd2.Run(cmd.Program, []string{"-h"})
    } else if opt.Command == "" {
      ProgramUsage(cmd.Program)
    } else {
      fmt.Fprintf(os.Stderr, "%s help: unknown command \"%s\". See '%s help'\n",
                  cmd.Program.Name, opt.Command, cmd.Program.Name)
    }
  }
})


func (p *Program) CommandNames() []string {
  keys := make([]string, len(p.Commands))
  i := 0
  for k, _ := range p.Commands {
    keys[i] = k
    i++
  }
  sort.Strings(keys)
  return keys
}


func ProgramUsage(p *Program) {
  nflags := 0
  p.Options.VisitAll(func(f *flag.Flag) { nflags++ })
  if nflags == 0 {
    if len(p.Commands) == 0 {
      fmt.Fprintf(os.Stderr, "Usage: %s\n", p.Name)
    } else {
      fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", p.Name)
    }
  } else {
    if len(p.Commands) == 0 {
      fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", p.Name)
    } else {
      fmt.Fprintf(os.Stderr, "Usage: %s [options] <command>\nOptions:\n", p.Name)
    }
    p.OptionsUsage()
  }
  if len(p.Commands) != 0 {
    os.Stderr.WriteString("Commands:\n")
    w := tabwriter.NewWriter(os.Stderr, 5, 0, 2, ' ', 0)
    for _, cmdName := range p.CommandNames() {
      cmd := p.Commands[cmdName]
      fmt.Fprintf(w, "  %s\t%s\n", cmd.NameAndArgs(), cmd.Description)
    }
    fmt.Fprint(w, "  help <cmd>\tMore information about a command\n")
    w.Flush()
  }
}


// Print options with their default values
func (p *Program) OptionsUsage() {
  optionsUsage(p.Options)
}


// Add a new command to this program. If there's already a command with the same name, that
// command is replaced with `cmd`.
func (p *Program) AddCommand(cmd *Command) {
  if p.Commands == nil {
    p.Commands = make(map[string]*Command)
  }
  p.Commands[cmd.Name] = cmd
}


// Parses args and returns the matching command
func (p *Program) Parse(args []string) (cmd *Command, cmdArgs []string) {
  p.Options.Usage = func() { p.Usage(p) }
  defer func(){ p.Options.Usage = nil }()
  if err := p.Options.Parse(args); err != nil {
    panic(err)
  }
  remainingArgs := p.Options.Args()
  if len(remainingArgs) == 0 || len(p.Commands) == 0 {
    // No command specified
    if p.DefaultCommand != nil {
      return p.DefaultCommand, remainingArgs[1:]
    }
    fmt.Fprintf(os.Stderr, "%s: no command specified\n", p.Name)
    p.Usage(p)
  } else {
    cmdName := remainingArgs[0]
    cmd := p.Commands[cmdName]
    if cmd == nil && cmdName == "help" {
      cmd = HelpCommand
    }
    if cmd != nil {
      return cmd, remainingArgs[1:]
    }
    fmt.Fprintf(os.Stderr, "%s: unknown command \"%s\". See '%s help'\n", p.Name, cmdName, p.Name)
  }
  if p.ExitOnError {
    os.Exit(1)
  }
  return nil, remainingArgs
}


// Parses args and runs a command. Returns the command run.
func (p *Program) Main(args []string) *Command {
  cmd, cmdArgs := p.Parse(args)
  if cmd != nil {
    cmd.Run(p, cmdArgs)
  }
  return cmd
}
