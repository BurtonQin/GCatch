# GCatch (generic support)

In this version of GCatch, I have upgraded the dependency `golang.org/x/tools` to version v0.9.2 in order to support the detection of generic functions. I chose v0.9.2 because it is the latest version that includes `golang.org/x/tools/go/pointer`. Additionally, I have made some changes to pointer to make it compatible with GCatch. These changes are marked with MYCODE. The modified version of `golang.org/x/tools` can be found in the `tools` directory.

This version of GCatch also includes upgrades to other dependencies and fully embraces go modules. It has been built using go mod and can ONLY detect go packages with go mod.

Please note that there may be some bugs in the current version, and the output results may not be completely consistent with the original version of GCatch that uses GOPATH.

## How to build

In the current directory

```
$ go build cmd/GCatch/main.go
```
builds an executable file `main`.

## Test BMOC checker

```
$ ./main -checker BMOC -mod-abs-path testdata/toyprogram/src/bufferCh/ -mod-module-path bufferCh
$ ./main -checker BMOC -mod-abs-path testdata/toyprogram/src/circularWait/ -mod-module-path circularWait
$ ./main -checker BMOC -mod-abs-path testdata/toyprogram/src/doubleClose/ -mod-module-path doubleClose  # no bug report for now
$ ./main -checker BMOC -mod-abs-path testdata/toyprogram/src/sendAfterClose/ -mod-module-path sendAfterClose
$ ./main -checker BMOC -mod-abs-path testdata/grpc-buggy/src/google.golang.org/grpc -mod-module-path google.golang.org/grpc
```

## Test double, conflict, unlock checkers

```
$ ./main -checker double -mod-abs-path /home/boqin/Projects/GCatch/GCatch/tests/double/ -mod-module-path double
$ ./main -checker conflict -mod-abs-path /home/boqin/Projects/GCatch/GCatch/tests/conflict/ -mod-module-path conflict
$ ./main -checker unlock -mod-abs-path /home/boqin/Projects/GCatch/GCatch/tests/unlock -mod-module-path unlock
```
