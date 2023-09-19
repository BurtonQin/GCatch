package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/system-pclub/GCatch/GCatch/checkers/bmoc"
	"github.com/system-pclub/GCatch/GCatch/checkers/conflictinglock"
	"github.com/system-pclub/GCatch/GCatch/checkers/doublelock"
	"github.com/system-pclub/GCatch/GCatch/checkers/fatal"
	"github.com/system-pclub/GCatch/GCatch/checkers/forgetunlock"
	"github.com/system-pclub/GCatch/GCatch/checkers/structfield"
	"github.com/system-pclub/GCatch/GCatch/config"
	"github.com/system-pclub/GCatch/GCatch/ssabuild"
	"github.com/system-pclub/GCatch/GCatch/util"
	"github.com/system-pclub/GCatch/GCatch/util/genKill"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/pointer"
)

func main() {

	mainStart := time.Now()
	defer func() {
		mainDur := time.Since(mainStart)
		fmt.Println("\n\nTime of main(): seconds", mainDur.Seconds())
	}()

	pCheckerName := flag.String("checker", "BMOC", "the checker to be used, divided by \":\"")
	pShowCompileError := flag.Bool("compile-error", false, "If fail to compile a package, show the errors of compilation")
	pExcludePath := flag.String("exclude", "vendor", "Name of directories that you want to ignore, divided by \":\"")
	pFnPointerAlias := flag.Bool("pointer", true, "Whether alias analysis is used to figure out function pointers")
	pPrintMod := flag.String("print-mod", "", "Print information like the number of channels, divided by \":\"")
	pGoModulePath := flag.String("mod-module-path", "", "Beta functionality: The module path of the program you want to check, like google.golang.org/grpc")
	pGoModAbsPath := flag.String("mod-abs-path", "", "Beta functionality: The absolute path of the program you want to check, which contains go.mod")

	flag.Parse()

	mapCheckerName := util.SplitStr2Map(*pCheckerName, ":")
	boolShowCompileError := *pShowCompileError
	boolFnPointerAlias := *pFnPointerAlias

	go func() {
		time.Sleep(time.Duration(config.MAX_GCATCH_DDL_SECOND) * time.Second)
		fmt.Println("!!!!")
		fmt.Println("The checker has been running for", config.MAX_GCATCH_DDL_SECOND, "seconds. Now force exit")
		os.Exit(1)
	}()

	for strCheckerName, _ := range mapCheckerName {
		switch strCheckerName {
		case "unlock":
			forgetunlock.Initialize()
		case "double":
			doublelock.Initialize()
		case "conflict", "structfield", "fatal", "BMOC": // no need to initialize these checkers
		case "NBMOC":
			config.BoolChSafety = true
		default:
			fmt.Println("Warning, a not existing checker is in -checker= flag:", strCheckerName)
		}
	}

	var errMsg string
	var bSucc bool

	// A beta functionality: using go.mod to build a program
	config.StrModAbsPath = *pGoModAbsPath
	config.StrModulePath = *pGoModulePath

	if config.StrModAbsPath == "" || config.StrModulePath == "" {
		fmt.Println("mod-module-path or mod-abs-path is empty\nPlease use -help to see what they are and provide valid values")
		return
	}

	config.Prog, config.Pkgs, bSucc, errMsg = ssabuild.BuildWholeProgramGoMod(config.StrModulePath, false, boolShowCompileError, config.StrModAbsPath) // Create SSA packages for the whole program including the dependencies.

	// What about all the other global variables we defined in traditional way and may use later? Let's go through them one by one
	// (0) config.StrEntrancePath: not used later
	// (1) config.StrGOPATH: not used later
	// (2) config.MapExcludePaths: used in IsPathIncluded in config/path.go, should be fine to just use the same code
	//								used in ListAllPkgPaths in config/path.go, let's avoid using it by disabling the -r flag later
	config.MapExcludePaths = util.SplitStr2Map(*pExcludePath, ":")
	// (3) config.StrRelativePath: used in PrintCallGraph in output/callgraph.go; it is not called anyway
	//								used in IsPathIncluded in config/path.go, need to edit this function to use StrModulePath
	//								used in ListAllPkgPaths in config/path.go like (2), let's disable the -r flag later
	// (4) config.StrAbsolutePath: used in multiple functions in config/path.go, but all can be avoided by disabling the -r flag
	//
	// (5) config.BoolDisableFnPointer: should be fine to just use the legacy code
	config.BoolDisableFnPointer = !boolFnPointerAlias
	// (6) config.MapPrintMod:			should be fine to just use the legacy code
	config.MapPrintMod = util.SplitStr2Map(*pPrintMod, ":")
	// (7) config.MapHashOfCheckedCh:			should be fine to just use the legacy code
	config.MapHashOfCheckedCh = make(map[string]struct{})

	if bSucc && len(config.Prog.AllPackages()) > 0 {
		// Step 2.1, Case 1: built SSA successfully, run the checkers in process()
		fmt.Println("Successfully built whole program. Now running checkers")

		detect(mapCheckerName)

	} else {
		// Step 2.1, Case 2: building SSA failed
		fmt.Println("Failed to build the whole program. The entrance package or its dependencies have error.", errMsg)
	}
}

func detect(mapCheckerName map[string]bool) {

	config.Inst2Defers, config.Defer2Insts = genKill.ComputeDeferMap()

	boolNeedCallGraph := mapCheckerName["double"] || mapCheckerName["conflict"] || mapCheckerName["structfield"] ||
		mapCheckerName["fatal"] || mapCheckerName["BMOC"]
	if boolNeedCallGraph {
		config.CallGraph = BuildCallGraph()
		if config.CallGraph == nil {
			return
		}
	}

	for strCheckerName, _ := range mapCheckerName {
		switch strCheckerName {
		case "unlock":
			forgetunlock.Detect()
		case "double":
			doublelock.Detect()
		case "conflict":
			conflictinglock.Detect()
		case "structfield":
			structfield.Detect()
		case "fatal":
			fatal.Detect()
		case "BMOC":
			bmoc.Detect()
		}
	}
}

func BuildCallGraph() *callgraph.Graph {
	cfg := &pointer.Config{
		OLDMains:        nil,
		Prog:            config.Prog,
		Reflection:      config.POINTER_CONSIDER_REFLECTION,
		BuildCallGraph:  true,
		Queries:         nil,
		IndirectQueries: nil,
		Log:             nil,
	}
	result, err := pointer.Analyze(cfg, nil)
	defer func() {
		cfg = nil
		result = nil
	}()
	if err != nil {
		fmt.Println("Error when building callgraph with nil Queries:\n", err.Error())
		return nil
	}
	graph := result.CallGraph
	return graph
}
