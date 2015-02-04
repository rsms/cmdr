package main
import (
  "github.com/rsms/cmdr"
  "os"
  "fmt"
)


var version = cmdr.Cmd("version", "Show version", func () {
  println("cmdr example v1.2.3")
})


var foo = cmdr.Cmd("foo", "Example command", func (opt *struct {
  FooBar      bool     `        Bar the foo with some bar`
  FirstName   string   `="John" Name of a cool person`
  Dir         string   `?"."    Directory to list`
  File        []string `!       Some files`
}, cmd *cmdr.Command) {
  cmd.Logf("foo command run with opt=%+v", opt)
  // Do something useful
})


var ls = cmdr.Cmd("ls", "List files", func (opt *struct {
  Long  bool    `      List in long format`
  Dir   string  `?"."  Directory to list`
}, cmd *cmdr.Command) {
  f, err := os.Open(opt.Dir)
  if err != nil {
    cmd.Fail(err)
  }
  defer f.Close()

  if opt.Long {
    fiv, _ := f.Readdir(0)
    for _, fi := range fiv {
      name := fi.Name()
      if fi.IsDir() {
        fmt.Printf("%10d     %s/\n", fi.Size(), name)
      } else if fi.Size() > 1024 {
        fmt.Printf("%10d kB  %s\n", fi.Size()/1024, name)
      } else {
        fmt.Printf("%10d B   %s\n", fi.Size(), name)
      }
    }
  } else {
    names, _ := f.Readdirnames(0)
    for _, name := range names {
      println(name)
    }
  }
})


func main() {
  cmdr.BoolVar(&cmdr.DefaultProgram.IsQuiet, "quiet", false, "Suppress status messages")
  cmdr.Main()
}
