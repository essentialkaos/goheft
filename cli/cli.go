package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
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

	"github.com/essentialkaos/goheft/cli/support"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "GoHeft"
	VER  = "0.6.2"
	DESC = "Utility for listing sizes of used static libraries"
)

const (
	OPT_EXTERNAL = "e:external"
	OPT_MIN_SIZE = "m:min-size"
	OPT_RAW      = "r:raw"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_VERB_VER     = "vv:verbose-version"
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
	OPT_HELP:     {Type: options.BOOL},
	OPT_VER:      {Type: options.BOOL},

	OPT_VERB_VER:     {Type: options.BOOL},
	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

var useRawOutput bool

// ////////////////////////////////////////////////////////////////////////////////// //

// Run is main utility function
func Run(gitRev string, gomod []byte) {
	runtime.GOMAXPROCS(1)

	preConfigureUI()

	args, errs := options.Parse(optMap)

	if len(errs) != 0 {
		printError(errs[0].Error())
		os.Exit(1)
	}

	configureUI()

	switch {
	case options.Has(OPT_COMPLETION):
		os.Exit(printCompletion())
	case options.Has(OPT_GENERATE_MAN):
		printMan()
		os.Exit(0)
	case options.GetB(OPT_VER):
		genAbout(gitRev).Print()
		os.Exit(0)
	case options.GetB(OPT_VERB_VER):
		support.Print(APP, VER, gitRev, gomod)
		os.Exit(0)
	case options.GetB(OPT_HELP) || len(args) == 0:
		genUsage().Print()
		os.Exit(0)
	}

	process(args.Get(0).Clean().String())
}

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	term := os.Getenv("TERM")

	fmtc.DisableColors = true

	if term != "" {
		switch {
		case strings.Contains(term, "xterm"),
			strings.Contains(term, "color"),
			term == "screen":
			fmtc.DisableColors = false
		}
	}

	if !fsutil.IsCharacterDevice("/dev/stdout") && os.Getenv("FAKETTY") == "" {
		fmtc.DisableColors = true
		useRawOutput = true
	}

	if os.Getenv("NO_COLOR") != "" {
		fmtc.DisableColors = true
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
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

	if !useRawOutput {
		fmtc.NewLine()
	}

	for _, lib := range libs {
		if lib.Size < minSize {
			continue
		}

		if useRawOutput {
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

	if !useRawOutput && minSize == 0 {
		fmtc.Printf(
			"\n %7s  {*}Total{!} {s-}(packages: %d){!}\n",
			fmtutil.PrettySize(libs.Total()), len(libs),
		)
	}

	if !useRawOutput {
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

			if !useRawOutput {
				fmtc.TPrintf("Building {*}%s{!}…", text)
			}
		}
	}()

	err = cmd.Start()

	if err != nil {
		return "", fmt.Errorf("Can't start build process: %v", err)
	}

	if !useRawOutput {
		fmtc.TPrintf("Processing sources…")
	}

	err = cmd.Wait()

	if !useRawOutput {
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

// printCompletion prints completion for given shell
func printCompletion() int {
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

// printMan prints man page
func printMan() {
	fmt.Println(
		man.Generate(
			genUsage(),
			genAbout(""),
		),
	)
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
func genAbout(gitRev string) *usage.About {
	about := &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2006,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/goheft", update.GitHubChecker},
	}

	if gitRev != "" {
		about.Build = "git:" + gitRev
	}

	return about
}