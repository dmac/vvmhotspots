vvmhotspots is a tool for drilling down into VisualVM call stacks.

# Install

```
$ go get github.com/dmac/vvmhotspots
```

# Usage

```
$ vvmhotspots -h
Usage: ./vvmhotspots [OPTIONS] FILE

FILE is the path to an exported VisualVM call tree in XML format.
Exported subtree call stacks work in addition to the full call stack,
and will process more quickly.

OPTIONS are:
  -ignore=[]: Ignore matching function names (may specify multiple)
  -n=50: Report first n results
  -root="": Treat first matching function name as root node
```
