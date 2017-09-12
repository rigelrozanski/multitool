package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//nolint
var (
	GoPath = os.Getenv("GOPATH")
	GoSrc  = path.Join(GoPath, "src")

	GotCmd = &cobra.Command{
		Use:   "got",
		Short: "Ebuchmans got tool",
	}
	ReplaceCmd = &cobra.Command{
		Use:   "replace [oldStr] [newStr]",
		Short: "String replace on all files in the directory tree",
		RunE:  replaceCmd,
	}
	PullCmd = &cobra.Command{
		Use:   "pull [oldStr] [newStr]",
		Short: "Swap paths, pull changes, swap back",
		RunE:  pullCmd,
	}
	BranchCmd = &cobra.Command{
		Use:   "branch",
		Short: "List the branch every directory is on",
		RunE:  branchCmd,
	}
	CheckoutCmd = &cobra.Command{
		Use:   "checkout",
		Short: "Checkout a git branch across all repos in the current dir. Add arguments like <repo>:<branch> to specify excpetions and <repo> to specify which repos to run checkout in, if not all.",
		RunE:  checkoutCmd,
	}
	DepCmd = &cobra.Command{
		Use:   "dep",
		Short: "Toggle the import statement for a godep managed dependency",
		RunE:  depCmd,
	}

	flagPath   = "path"
	flagDepth  = "depth"
	flagLocal  = "local"
	flagVendor = "vendor"
)

func init() {

	fsReplace := flag.NewFlagSet("", flag.ContinueOnError)
	fsDep := flag.NewFlagSet("", flag.ContinueOnError)

	fsReplace.StringP(flagPath, "p", ".", "specify the path to act upon")
	fsReplace.IntP(flagDepth, "d", -1, "specify the recursion depth")
	fsDep.AddFlagSet(fsReplace)
	fsDep.BoolP(flagLocal, "l", false, "set the import path tot he proper $GOPATH location")
	fsDep.BoolP(flagVendor, "v", false,
		"set the import path to the vendored location (a mirror of $GOPATH within the current repo)")

	ReplaceCmd.Flags().AddFlagSet(fsReplace)
	DepCmd.Flags().AddFlagSet(fsDep)

	viper.BindPFlag(flagPath, fsDep.Lookup(flagPath))
	viper.BindPFlag(flagDepth, fsDep.Lookup(flagDepth))
	viper.BindPFlag(flagLocal, fsDep.Lookup(flagLocal))
	viper.BindPFlag(flagVendor, fsDep.Lookup(flagVendor))

	GotCmd.AddCommand(
		ReplaceCmd,
		PullCmd,
		CheckoutCmd,
		BranchCmd,
		DepCmd,
	)

	RootCmd.AddCommand(GotCmd)
	RootCmd.AddCommand(ReplaceCmd)
}

// replace a line of text in every file with another
func replaceCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("command needs 2 args")
	}
	oldS := args[0]
	newS := args[1]
	dir := viper.GetString(flagPath)
	depth := viper.GetInt(flagDepth)
	return replace(dir, oldS, newS, depth)
}

// replace import paths with host, pull, replace back
func pullCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("command needs 2 args")
	}
	remote := args[0]
	branch := args[1]
	remotePath, err := resolveRemoteRepo(remote)
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	localPath, err := resolveLocalRepo(wd)
	if err != nil {
		return err
	}
	localFullPath := path.Join(GoSrc, localPath)

	err = replace(localFullPath, localPath, remotePath, -1)
	if err != nil {
		return err
	}
	addCommit("change to upstream paths")
	gitPull(remote, branch)
	return replace(localFullPath, remotePath, localPath, -1)
}

func branchCmd(cmd *cobra.Command, args []string) error {
	var dir string
	if len(args) == 1 {
		dir = args[0]
	} else {
		dir, _ = os.Getwd()
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range dirFiles {
		name := f.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		p := path.Join(dir, name)
		if f.IsDir() {
			branch, err := gitGetBranch(p)
			if err == ErrNotGitRepo {
				continue
			}
			if err != nil {
				return err
			}
			fmt.Printf("%s : %s\n", name, branch)
		}
	}
	return nil
}

// update the go-import paths in a directory
func depCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("command needs 1 arg")
	}
	repo := args[0]

	depth := viper.GetInt(flagDepth)
	dir := viper.GetString(flagPath)
	oldS, newS := "", ""
	current, _ := os.Getwd()

	if !strings.HasPrefix(current, GoSrc) {
		return fmt.Errorf("Directory is not on the $GOPATH")
	}

	remains := current[len(GoSrc)+1:] // consume the slash too
	spl := strings.Split(remains, "/")
	if len(spl) < 3 {
		return fmt.Errorf("Invalid positioned repo on the $GOPATH")
	}
	currentRepo := strings.Join(spl[:3], "/")

	if viper.GetBool(flagLocal) {
		oldS = path.Join(currentRepo, "Godeps", "_workspace", "src", repo)
		newS = repo
	} else if viper.GetBool(flagVendor) {
		oldS = repo
		newS = path.Join(currentRepo, "Godeps", "_workspace", "src", repo)
	} else {
		return fmt.Errorf("Specify the --local or --vendor flag to toggle the import statement")
	}

	// now run the replace
	// but avoid the Godeps/ dir
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if !f.IsDir() {
			continue
		}
		if f.Name() != "Godeps" {
			err := replace(f.Name(), oldS, newS, depth-1)
			if err != nil {
				return err
			}
		}
	}
	return replace(dir, oldS, newS, 1) // replace in any files in the root
}

// checkout a branch across every repository in a directory
func checkoutCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("command needs at least 1 arg")
	}
	branch := args[0]
	var repos []string
	if len(args) > 1 {
		repos = args[1:]
	}

	var nonColon bool
	repoMap := make(map[string]string)
	for _, r := range repos {
		sp := strings.Split(r, ":")
		repo := sp[0]
		var b string
		if len(sp) != 2 {
			nonColon = true
			b = branch
			//ifExit(fmt.Errorf("Additional arguments must be of the form <repo>:<branch>"))
		} else {
			b = sp[1]
		}
		repoMap[repo] = b
	}

	dir, _ := os.Getwd()

	// if nonColon, we only loop through dirs in the repoMap
	if nonColon {
		for r, b := range repoMap {
			p := path.Join(dir, r)
			f, err := os.Stat(p)
			if err != nil {
				log.Println("Unknown repo:", r)
				continue
			}
			if !f.IsDir() {
				log.Println(r, " is not a directory")
			}
			gitCheckout(p, b)
		}
		return nil
	}

	// otherwise, we loop through all dirs in the current one
	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range dirFiles {
		name := f.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		p := path.Join(dir, name)
		if f.IsDir() {
			if b, ok := repoMap[name]; ok {
				gitCheckout(p, b)
			} else {
				gitCheckout(p, branch)
			}
		}
	}
	return nil
}
