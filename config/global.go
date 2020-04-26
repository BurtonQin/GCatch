package config

import (
	"github.com/system-pclub/gochecker/tools/go/callgraph"
	"github.com/system-pclub/gochecker/tools/go/ssa"
	"sync"
)


//-path=/home/song/work/go-workspace/code/src/github.com/etcd-io/etcd -include=github.com/etcd-io/etcd

var StrEntrancePath string // github.com/etcd-io/etcd
var StrGOPATH string // /home/song/work/go-workspace/code
var StrAbsolutePath string // /home/song/work/go-workspace/code/src/
var StrRelativePath string // github.com/etcd-io/etcd

var BoolDisableFnPointer bool


var VecExcludePaths [] string

var Prog *ssa.Program
var Pkgs []*ssa.Package


var BugIndex int
var BugIndexMu sync.Mutex

var VecPathStats [] PathStat

var CallGraph * callgraph.Graph