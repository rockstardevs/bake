// bake - release management tool
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/singhsaysdotcom/cobra"
	"github.com/singhsaysdotcom/gdrive/gdrive"
	"github.com/singhsaysdotcom/shlog"
	"google.golang.org/api/drive/v2"
)

var (
	// Flags
	NAME    string // populated via ldflags
	VERSION string // populated via ldflags

	GOOS   = runtime.GOOS
	GOARCH = runtime.GOARCH

	versionFile       string
	enable_uploads    bool
	enable_git_push   bool
	enable_git_tag    bool
	enable_git_commit bool
	enable_git_tasks  bool

	logger          = shlog.NewLogger()
	is_versioned    bool
	current_version *Version

	logFile    *os.File
	errLogFile *os.File
	goog       *gdrive.Drive
)

func SetEnv() {
	if os.Getenv("GOOS") != "" {
		GOOS = os.Getenv("GOOS")
	}
	if os.Getenv("GOARCH") != "" {
		GOARCH = os.Getenv("GOARCH")
	}
}

func Build(args []string) bool {
	SetEnv()
	logger.Message("GOOS")
	logger.Status(shlog.Grey, GOOS)
	logger.Message("GOARCH")
	logger.Status(shlog.Grey, GOARCH)
	logger.Message("building %s ...", PkgName(""))
	_, err := exec.LookPath("go")
	if err != nil {
		logger.Err()
		return false
	}
	c := exec.Command("go", "build", "-v", "-ldflags", "-X \"main.VERSION="+current_version.String()+"\"", "-o", PkgName(""))
	if len(args) > 0 {
		c.Args = append(c.Args, args...)
	}
	if !CaptureLogs("build", c) {
		logger.Err()
		return false
	} else {
		logger.Ok()
		return true
	}
}

// git commits version file
func CommitVersion() bool {
	var ok bool
	logger.Message("git commit new version ...")
	ok = CaptureLogs("git add", exec.Command("git", "add", versionFile))
	if !ok {
		logger.Err()
		return false
	}
	ok = CaptureLogs("commitversion", exec.Command("git", "commit", versionFile, "-m", "Built new version "+current_version.String()))
	if ok {
		logger.Ok()
	} else {
		logger.Err()
	}
	return ok
}

// adds a git tag
func TagVersion() bool {
	logger.Message("adding git tag ...")
	ok := CaptureLogs("git tag", exec.Command("git", "tag", "v"+current_version.String()))
	if ok {
		logger.Ok()
	} else {
		logger.Err()
	}
	return ok
}

// git push to remote
func Push() bool {
	logger.Message("git push to remote ...")
	changes, err := exec.Command("git", "status", "-s").Output()
	if err != nil {
		logger.Status(shlog.Orange, "skipped - unknown status")
		return false
	}
	if len(changes) > 0 {
		logger.Status(shlog.Orange, "skipped - uncommited changes")
		return false
	}
	remotes, err := exec.Command("git", "remote").Output()
	if err != nil || len(remotes) == 0 {
		logger.Status(shlog.Orange, "skipped - no remotes")
		return false
	}
	err = exec.Command("git", "push").Run()
	if err != nil {
		logger.Err()
		return false
	}
	logger.Ok()
	return true
}

// Uploads the binary to google drive
func Upload() bool {
	var (
		err error
	)
	logger.Message("uploading new binary ...")
	goog, err = gdrive.New("", false, false)
	if err != nil {
		logger.Status(shlog.Orange, "not configured")
		return false
	}
	// TODO: check it file already exists and remove it
	// TODO: add support for uploads to folders
	filename := PkgName("")
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		logger.Err()
		return false
	}
	fileRef := &drive.File{Title: path.Base(filename), MimeType: "application/octet-stream"}
	reader := bufio.NewReader(f)
	_, err = goog.Files.Insert(fileRef).Media(reader).Do()
	if err != nil {
		logger.Err()
		return false
	}
	logger.Ok()
	return true
}

func BuildCommon(save_version bool, upload bool, args []string) {
	var ok bool
	if !Build(args) {
		return
	}
	if save_version {
		if ok, _ = SaveVersion(&versionFile); !ok {
			return
		}
	}
	if IsGitRepo() && enable_git_tasks {
		// Skip all git changes if no new version
		if !save_version {
			logger.Message("git commit new version ...")
			logger.Status(shlog.Orange, "skipped - no new version")
			logger.Message("adding git tag ...")
			logger.Status(shlog.Orange, "skipped - no new version")
			logger.Message("git push to remote ...")
			logger.Status(shlog.Orange, "skipped - no new version")
		} else {
			if enable_git_commit && !CommitVersion() {
				return
			}
			if enable_git_tag && !TagVersion() {
				return
			}
			if enable_git_push && !Push() {
				return
			}
		}
	}
	if enable_uploads && upload {
		Upload()
	}
	logger.Done()
}

// PrintVersion prints the current version of bake.
func PrintVersion(cmd *cobra.Command, args []string) {
	logger.Message("Bake Version")
	logger.Status(shlog.Green, VERSION)
}

// Build a new major version
func BuildMajor(cmd *cobra.Command, args []string) {
	if !is_versioned {
		current_version = NewVersion()
		logger.Message("new major version")
		logger.Status(shlog.Green, current_version.String())
	} else {
		current_version.IncMajor()
	}
	BuildCommon(true, true, args)
}

// Build a new minor version
func BuildMinor(cmd *cobra.Command, args []string) {
	if !is_versioned {
		current_version = NewVersion()
		logger.Message("new minor version")
		logger.Status(shlog.Green, current_version.String())
	} else {
		current_version.IncMinor()
	}
	BuildCommon(true, true, args)
}

// Build at the next build number
func BuildNext(cmd *cobra.Command, args []string) {
	if !is_versioned {
		current_version = NewVersion()
		logger.Message("new build version")
		logger.Status(shlog.Green, current_version.String())
	} else {
		current_version.IncBuild()
	}
	BuildCommon(true, true, args)
}

// Rebuild at current version
func Rebuild(cmd *cobra.Command, args []string) {
	BuildCommon(false, false, args)
}

// Rebuilds at the current version and reuploads to drive
func Reupload(cmd *cobra.Command, args []string) {
	BuildCommon(false, true, args)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "bake",
		Short: "minimal build and release tool",
	}
	rootCmd.PersistentFlags().StringVarP(&versionFile, "version_file", "f", "VERSION", "name of the version file")
	rootCmd.PersistentFlags().BoolVar(&enable_uploads, "enable_uploads", true, "enable uploads to Google Drive")
	rootCmd.PersistentFlags().BoolVar(&enable_git_tasks, "enable_git_tasks", true, "enable git related tasks")
	rootCmd.PersistentFlags().BoolVar(&enable_git_commit, "enable_git_commit", true, "enable git commits for version changes")
	rootCmd.PersistentFlags().BoolVar(&enable_git_tag, "enable_git_tag", true, "enable git tagging for version changes.")
	rootCmd.PersistentFlags().BoolVar(&enable_git_push, "enable_git_push", true, "enable git push to remotes")

	is_versioned, current_version, _ = GetVersion(&versionFile)

	// Log files
	err := os.Mkdir(".log", 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Printf("Error creating logs directory\n")
		os.Exit(1)
	}
	logFile, err := os.Create(".log/bake.log")
	defer logFile.Close()
	if err != nil {
		fmt.Printf("Error creating log file\n")
		os.Exit(1)
	}
	errLogFile, err = os.Create(".log/bake.err.log")
	defer errLogFile.Close()
	if err != nil {
		fmt.Printf("Error creating log file\n")
		os.Exit(1)
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "prints current version of bake itself.",
		Run:   PrintVersion,
	}

	majorCmd := &cobra.Command{
		Use:   "major",
		Short: "build a new major version",
		Run:   BuildMajor,
	}

	minorCmd := &cobra.Command{
		Use:   "minor",
		Short: "build a new minor version",
		Run:   BuildMinor,
	}

	nextCmd := &cobra.Command{
		Use:   "next",
		Short: "build at the next build number",
		Run:   BuildNext,
	}

	rebuildCmd := &cobra.Command{
		Use:   "rebuild",
		Short: "rebuilds at the current version",
		Run:   Rebuild,
	}

	reuploadCmd := &cobra.Command{
		Use:   "reupload",
		Short: "rebuilds at the current version and reuploads to drive",
		Run:   Reupload,
	}

	rootCmd.AddCommand(versionCmd, majorCmd, minorCmd, nextCmd, rebuildCmd, reuploadCmd)
	rootCmd.Execute()
}
