cmdr helps writing command-line programs with subcommands in Go.

example.go:

    package main
    import (
      "github.com/rsms/cmdr"
    )

    var quiet = false
    var rootDir = "."

    var version = cmdr.Cmd("version", "Show version", func () {
      println("cmdr example v1.2.3")
    })

    var foo = cmdr.Cmd("foo", "Example command", func (opt *struct {
      FooBar      bool     `        Bar the foo with some bar`
      FirstName   string   `="John" Name of a cool person`
      Dir         string   `?"."    Directory to list`
      File        []string `!       Some files`
    }, cmd *cmdr.Command) {
      if !quiet {
        cmd.Logf("foo command run with opt=%+v", opt)
      }
      // Do something useful
    })

    func main() {
      cmdr.BoolVar(&quiet, "quiet", false, "Suppress status messages")
      cmdr.StringVar(&rootDir, "root-dir", rootDir, "Directory to start in")
      cmdr.Main()
    }

Output:

    $ go build
    $ ./example -h
    Usage: ./example [options] <command>
    Options:
      -quiet false    Suppress status messages
      -root-dir "."   Directory to start in
    Commands:
      foo [<dir>] <file>...  Example command
      version                Show version
      help <cmd>             More information about a command
    
    $ ./example version
    cmdr example v1.2.3
    $ ./example help foo
    Example command
    Usage: ./example foo [options] [<dir>] <file>...
    Options:
      -first-name "John"   Name of a cool person
      -foo-bar             Bar the foo with some bar
    Arguments:
      <dir>       Directory to list (default: ".")
      <file>...   Some files 
    $ ./example foo --first-name "Lisa" -foo-bar=false hello *.go
    foo command run with opt=&{FooBar:false FirstName:Lisa Dir:hello File:[example.go]}
    $


## MIT license

Copyright (c) 2015 Rasmus Andersson <http://rsms.me/>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
