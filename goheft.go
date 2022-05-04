package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/essentialkaos/ek/v12/env"
	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/usage"
	"github.com/essentialkaos/ek/v12/usage/completion/bash"
	"github.com/essentialkaos/ek/v12/usage/completion/fish"
	"github.com/essentialkaos/ek/v12/usage/completion/zsh"
	"github.com/essentialkaos/ek/v12/usage/man"
	"github.com/essentialkaos/ek/v12/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "GoHeft"
	VER  = "0.6.1"
	DESC = "Utility for listing sizes of used static libraries"
)

const (
	OPT_EXTERNAL = "e:external"
	OPT_MIN_SIZE = "m:min-size"
	OPT_RAW      = "r:raw"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	SIZE_HUGE  uint64 = 5 * 1024 * 1024 // 5Mb
	SIZE_BIG   uint64 = 1024 * 1024     // 1Mb
	SIZE_SMALL uint64 = 25 * 1024       // 25Kb
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
	OPT_RAW:      {Type: options.BOOL},
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
		os.Exit(showAbout())
	}

	if options.GetB(OPT_HELP) || len(args) == 0 {
		os.Exit(showUsage())
	}

	process(args.Get(0).Clean().String())
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

	if options.GetB(OPT_RAW) {
		useRawOuput = true
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

	libsInfo, err := getLibsInfo(workDir)

	if err != nil {
		printErrorAndExit(err.Error())
	}

	if len(libsInfo) == 0 {
		printWarn("No *.a files are found")
		return
	}

	printStats(libsInfo)

	os.RemoveAll(workDir)
}

// getLibsInfo returns slice with info about all used static libs
func getLibsInfo(workDir string) (LibInfoSlice, error) {
	libs := fsutil.List(
		workDir, true,
		fsutil.ListingFilter{Perms: "DRX"},
	)

	// Map package -> path
	pathStore := make(map[string]string)

	for _, libDir := range libs {
		err := scanPkgImports(workDir+"/"+libDir+"/importcfg", pathStore)

		if err != nil {
			return LibInfoSlice{}, err
		}
	}

	var result LibInfoSlice

	for name, lib := range pathStore {
		result = append(result, LibInfo{
			Package: name,
			Size:    uint64(fsutil.GetSize(lib)),
		})
	}

	return result, nil
}

// scanPkgImports extracts packages data from given file
func scanPkgImports(file string, store map[string]string) error {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer fd.Close()

	r := bufio.NewReader(fd)
	s := bufio.NewScanner(r)

	for s.Scan() {
		text := s.Text()

		if !strings.HasPrefix(text, "packagefile ") {
			continue
		}

		pkgInfo := strutil.ReadField(text, 1, false, " ")
		pkgName := strutil.ReadField(pkgInfo, 0, false, "=")

		pkgName = normalizePackageName(pkgName)

		if store[pkgName] == "" {
			store[pkgName] = strutil.ReadField(pkgInfo, 1, false, "=")
		}
	}

	return nil
}

// printStats print statistics
func printStats(libs LibInfoSlice) {
	sort.Sort(sort.Reverse(libs))

	var colorTag string
	var minSize uint64

	if options.Has(OPT_MIN_SIZE) {
		minSize = fmtutil.ParseSize(options.GetS(OPT_MIN_SIZE))
	}

	if !useRawOuput {
		fmtc.NewLine()
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
		case lib.Size > SIZE_HUGE:
			colorTag = "{r}"
		case lib.Size > SIZE_BIG:
			colorTag = "{y}"
		case lib.Size < SIZE_SMALL:
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

	if !useRawOuput && minSize == 0 {
		fmtc.Printf(
			"\n %7s  {*}Total{!} {s-}(packages: %d){!}\n",
			fmtutil.PrettySize(libs.Total()), len(libs),
		)
	}

	if !useRawOuput {
		fmtc.NewLine()
	}
}

// buildBinary run `go build` command and parse output
func buildBinary(file string) (string, error) {
	var workDir string

	cmd := exec.Command("go", "build", "-work", "-a", "-v", file)
	stderrReader, err := cmd.StderrPipe()

	if err != nil {
		return "", fmt.Errorf("Can't redirect 'go build' output: %v", err)
	}

	scanner := bufio.NewScanner(stderrReader)

	go func() {
		for scanner.Scan() {
			text := scanner.Text()

			if workDir == "" {
				workDir = text
				continue
			}

			if strings.HasPrefix(text, "can't load package") {
				return
			}

			text = normalizePackageName(text)

			if !useRawOuput {
				fmtc.TPrintf("Building {*}%s{!}…", text)
			}
		}
	}()

	err = cmd.Start()

	if err != nil {
		return "", fmt.Errorf("Can't start build process: %v", err)
	}

	if !useRawOuput {
		fmtc.TPrintf("Processing sources…")
	}

	err = cmd.Wait()

	if !useRawOuput {
		fmtc.TPrintf("")
	}

	if err != nil {
		return "", fmt.Errorf("Can't start build process: %v", err)
	}

	return strutil.ReadField(workDir, 1, false, "="), nil
}

// normalizePackageName format package name
func normalizePackageName(name string) string {
	if !strings.Contains(name, "vendor/") {
		return name
	}

	return strutil.Substr(name, strings.Index(name, "vendor/")+7, 999999)
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

// Total returns size of all libraries
func (s LibInfoSlice) Total() uint64 {
	var result uint64

	for _, l := range s {
		result += l.Size
	}

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage prints usage info
func showUsage() int {
	genUsage().Render()
	return 0
}

// showAbout prints info about version
func showAbout() int {
	genAbout().Render()
	return 0
}

// genCompletion generates completion for different shells
func genCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Printf(bash.Generate(info, "goheft", "go"))
	case "fish":
		fmt.Printf(fish.Generate(info, "goheft"))
	case "zsh":
		fmt.Printf(zsh.Generate(info, optMap, "goheft", "*.go"))
	default:
		return 1
	}

	return 0
}

// genMan generates man page
func genMan() int {
	fmt.Println(
		man.Generate(
			genUsage(),
			genAbout(),
		),
	)

	return 0
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo("", "go-file")

	info.AddOption(OPT_EXTERNAL, "Shadow internal packages")
	info.AddOption(OPT_MIN_SIZE, "Don't show with size less than defined", "size")
	info.AddOption(OPT_RAW, "Print raw data")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.AddExample("application.go", "Show size of each used library")
	info.AddExample("application.go -m 750kb", "Show size of each used library which greater than 750kb")

	return info
}

// genAbout generates info about version
func genAbout() *usage.About {
	return &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2006,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/goheft", update.GitHubChecker},
	}
}
