package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2017 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"pkg.re/essentialkaos/ek.v6/arg"
	"pkg.re/essentialkaos/ek.v6/env"
	"pkg.re/essentialkaos/ek.v6/fmtc"
	"pkg.re/essentialkaos/ek.v6/fmtutil"
	"pkg.re/essentialkaos/ek.v6/fsutil"
	"pkg.re/essentialkaos/ek.v6/strutil"
	"pkg.re/essentialkaos/ek.v6/usage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "GoHeft"
	VER  = "0.0.1"
	DESC = "Utility for listing sizes of used static libraries"
)

const (
	ARG_EXTERNAL = "e:external"
	ARG_MIN_SIZE = "m:min-size"
	ARG_NO_COLOR = "nc:no-color"
	ARG_HELP     = "h:help"
	ARG_VER      = "v:version"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type LibInfo struct {
	Package string
	Size    uint64
}

type LibInfoSlice []LibInfo

func (s LibInfoSlice) Len() int           { return len(s) }
func (s LibInfoSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s LibInfoSlice) Less(i, j int) bool { return s[i].Size < s[j].Size }

// ////////////////////////////////////////////////////////////////////////////////// //

var argMap = arg.Map{
	ARG_EXTERNAL: {Type: arg.BOOL},
	ARG_MIN_SIZE: {},
	ARG_NO_COLOR: {Type: arg.BOOL},
	ARG_HELP:     {Type: arg.BOOL, Alias: "u:usage"},
	ARG_VER:      {Type: arg.BOOL, Alias: "ver"},
}

var useRawOuput bool

// ////////////////////////////////////////////////////////////////////////////////// //

func main() {
	runtime.GOMAXPROCS(1)

	args, errs := arg.Parse(argMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	configureUI()

	if arg.GetB(ARG_VER) {
		showAbout()
		return
	}

	if arg.GetB(ARG_HELP) || len(args) == 0 {
		showUsage()
		return
	}

	process(args[0])
}

// configureUI configure user interface
func configureUI() {
	envVars := env.Get()
	term := envVars.GetS("TERM")

	fmtc.DisableColors = true

	if term != "" {
		switch {
		case strings.Contains(term, "xterm"),
			strings.Contains(term, "color"),
			term == "screen":
			fmtc.DisableColors = false
		}
	}

	if arg.GetB(ARG_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if !fsutil.IsCharacterDevice("/dev/stdout") && envVars.GetS("FAKETTY") == "" {
		fmtc.DisableColors = true
		useRawOuput = true
	}
}

// process start build
func process(file string) {
	if !fsutil.IsExist(file) {
		printError("Can't build binary - file %s is not exist", file)
		os.Exit(1)
	}

	workDir, err := buildBinary(file)

	if err != nil {
		printError(err.Error())
		os.RemoveAll(workDir)
		os.Exit(1)
	}

	libsInfo := getLibsInfo(workDir)

	if len(libsInfo) == 0 {
		printWarn("No *.a files are found")
		return
	}

	printStats(libsInfo)

	os.RemoveAll(workDir)
}

// getLibsInfo remove slice with info about all used static libs
func getLibsInfo(workDir string) LibInfoSlice {
	libs := fsutil.ListAllFiles(
		workDir, true,
		&fsutil.ListingFilter{
			MatchPatterns: []string{"*.a"},
		},
	)

	if len(libs) == 0 {
		return nil
	}

	var result LibInfoSlice

	for _, lib := range libs {
		libName := strutil.Substr(lib, 0, len(lib)-2)
		libSize := uint64(fsutil.GetSize(workDir + "/" + lib))
		result = append(result, LibInfo{libName, libSize})
	}

	return result
}

// printStats print statistics
func printStats(libs LibInfoSlice) {
	sort.Sort(sort.Reverse(libs))

	var colorTag string
	var minSize uint64

	if arg.Has(ARG_MIN_SIZE) {
		minSize = fmtutil.ParseSize(arg.GetS(ARG_MIN_SIZE))
	}

	for _, lib := range libs {
		if lib.Size < minSize {
			continue
		}

		if useRawOuput {
			fmt.Println(lib.Size, lib.Package)
			continue
		}

		switch {
		case lib.Size > 5*1024*1024:
			colorTag = "{r}"
		case lib.Size > 1024*1024:
			colorTag = "{y}"
		case lib.Size < 25*1024:
			colorTag = "{s}"
		default:
			colorTag = ""
		}

		if arg.GetB(ARG_EXTERNAL) && !strings.Contains(lib.Package, ".") {
			fmtc.Printf(" {s-}%7s  %s{!}\n", fmtutil.PrettySize(lib.Size), lib.Package)
		} else {
			fmtc.Printf(" "+colorTag+"%7s{!}  %s\n", fmtutil.PrettySize(lib.Size), lib.Package)
		}
	}
}

// buildBinary run `go build` command and parse output
func buildBinary(file string) (string, error) {
	cmd := exec.Command(
		"go",
		"build",
		"-work",
		"-a",
		file,
	)

	output, _ := cmd.CombinedOutput()

	return parseBuildOutput(output)
}

// parseBuildOutput parse `go build` command output
func parseBuildOutput(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("\"go build\" output is empty")
	}

	// Remove empty line at the end
	data = data[:len(data)-1]

	dataSlice := strings.Split(string(data), "\n")
	workDir := strutil.Substr(dataSlice[0], 5, 9999)

	if len(dataSlice) == 1 {
		return workDir, nil
	}

	return workDir, fmt.Errorf(strings.Join(dataSlice[1:], "\n"))
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Printf("{r}"+f+"{!}\n", a...)
}

// printWarn prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Printf("{y}"+f+"{!}\n", a...)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	usage.Breadcrumbs = true

	info := usage.NewInfo("", "file")

	info.AddOption(ARG_EXTERNAL, "Shadow internal packages")
	info.AddOption(ARG_MIN_SIZE, "Don't show with size less than defined", "size")
	info.AddOption(ARG_NO_COLOR, "Disable colors in output")
	info.AddOption(ARG_HELP, "Show this help message")
	info.AddOption(ARG_VER, "Show version")

	info.AddExample("application.go", "Show size of each used library")
	info.AddExample("application.go -m 750kb", "Show size of each used library which greater than 750kb")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:        APP,
		Version:    VER,
		Desc:       DESC,
		Year:       2006,
		Owner:      "ESSENTIAL KAOS",
		License:    "Essential Kaos Open Source License <https://essentialkaos.com/ekol>",
		Repository: "essentialkaos/goheft",
	}

	about.Render()
}
