[![progress-banner](https://backend.codecrafters.io/progress/shell/315e8342-2f17-4d99-a405-0ab392e40752)](https://app.codecrafters.io/users/memreunlucan15?r=2qF)

# Build Your Own Shell — Go

My solution to the [CodeCrafters "Build Your Own Shell" Challenge](https://app.codecrafters.io/courses/shell/overview), written from scratch in Go.

It's a small POSIX-style shell: it reads a line, parses it, and either runs a builtin, executes an external program from `PATH`, or wires up a pipeline. I wrote every part by hand to actually understand how a shell works under the hood — the tokenizer, process management, job control and completion are all my own code rather than library glue.

## Features

- **REPL** — read, evaluate, print, loop
- **Builtins** — `echo`, `exit`, `type`, `pwd`, `cd`, `history`, `jobs`, `complete`, `declare`
- **External commands** — resolved through `PATH` and run as child processes
- **Quoting** — single quotes, double quotes and backslash escaping, handled by a hand-written state-machine tokenizer
- **Redirection** — `>`, `1>`, `2>`, `>>`, `1>>`, `2>>`
- **Pipelines** — two-command and multi-command pipelines, including pipelines that mix builtins and external programs
- **Background jobs** — `cmd &`, the `jobs` builtin, and automatic reaping before each prompt
- **Tab completion** — builtins, `PATH` executables, files and directories, longest-common-prefix completion, multiple-match listing, and programmable completion via `complete -C`
- **History** — in-memory history, `history -r/-w/-a`, and persistence through `HISTFILE`
- **Parameter expansion** — `$VAR` and `${VAR}`

## Running it locally

You need Go (1.26+) installed. Then just run the `app` package:

```sh
go run ./app
```

On Linux/macOS you can also use the build-and-run script:

```sh
./my_shell.sh
```

## Notes

This is a learning project — the goal was understanding, not building a production shell. Working through it stage by stage taught me a lot about parsing, processes, pipes, concurrency and how much is really going on every time you hit Enter in a terminal.
