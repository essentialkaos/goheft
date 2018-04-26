package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2018 ESSENTIAL KAOS                         //
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

	"pkg.re/essentialkaos/ek.v9/env"
	"pkg.re/essentialkaos/ek.v9/fmtc"
	"pkg.re/essentialkaos/ek.v9/fmtutil"
	"pkg.re/essentialkaos/ek.v9/fsutil"
	"pkg.re/essentialkaos/ek.v9/options"
	"pkg.re/essentialkaos/ek.v9/strutil"
	"pkg.re/essentialkaos/ek.v9/usage"
	"pkg.re/essentialkaos/ek.v9/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "GoHeft"
	VER  = "0.3.1"
	DESC = "Utility for listing sizes of used static libraries"
)

const (
	OPT_EXTERNAL = "e:external"
	OPT_MIN_SIZE = "m:min-size"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"
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

var optMap = options.Map{
	OPT_EXTERNAL: {Type: options.BOOL},
	OPT_MIN_SIZE: {},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VER:      {Type: options.BOOL, Alias: "ver"},
}

var useRawOuput bool

// ////////////////////////////////////////////////////////////////////////////////// //

func main() {
	runtime.GOMAXPROCS(1)

	args, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	configureUI()

	if options.GetB(OPT_VER) {
		showAbout()
		return
	}

	if options.GetB(OPT_HELP) || len(args) == 0 {
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

	if options.GetB(OPT_NO_COLOR) {
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
		printErrorAndExit("Can't build binary - file %s does not exist", file)
	}

	workDir, err := buildBinary(file)

	if err != nil {
		os.RemoveAll(workDir)
		printErrorAndExit(err.Error())
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
		fsutil.ListingFilter{
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

	if options.Has(OPT_MIN_SIZE) {
		minSize = fmtutil.ParseSize(options.GetS(OPT_MIN_SIZE))
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

		if options.GetB(OPT_EXTERNAL) && !strings.Contains(lib.Package, ".") {
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
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
}

// printErrorAndExit print error mesage and exit with exit code 1
func printErrorAndExit(f string, a ...interface{}) {
	printError(f, a...)
	os.Exit(1)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func showUsage() {
	info := usage.NewInfo("", "file")

	info.AddOption(OPT_EXTERNAL, "Shadow internal packages")
	info.AddOption(OPT_MIN_SIZE, "Don't show with size less than defined", "size")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.AddExample("application.go", "Show size of each used library")
	info.AddExample("application.go -m 750kb", "Show size of each used library which greater than 750kb")

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2006,
		Owner:         "ESSENTIAL KAOS",
		License:       "Essential Kaos Open Source License <https://essentialkaos.com/ekol>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/goheft", update.GitHubChecker},
	}

	about.Render()
}
