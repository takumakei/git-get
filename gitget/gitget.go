package gitget

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/goaux/results"
	"github.com/goaux/slog/logger"
	"github.com/goaux/stacktrace/v2"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/takumakei/git-get/gitget/excludes"
	"github.com/takumakei/git-get/gitget/flags"
	"github.com/takumakei/git-get/gitget/giturl"
)

var log = results.Must1(logger.New())

//go:embed gitget.example.txt
var exampleTxt string

var GitGet = &cobra.Command{
	Use:     "git-get [flags] <url>[@<commit>] [<dir>]",
	Short:   "Extract specific directories or files from any commit within your Git repository",
	Example: strings.TrimRightFunc(exampleTxt, unicode.IsSpace),

	RunE: runE,

	ValidArgsFunction: validArgs,

	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	fl := GitGet.Flags()
	fl.StringP("cache-dir", "d", cacheDir(), "Cache `dir`")
	fl.BoolP("force", "f", false, "Force overwrite")
	fl.StringSliceP("exclude", "e", nil, "Exclude `patterns`")

	setup(GitGet, "GIT_GET_")
}

func cacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "git-get")
	}
	if dir, err := os.UserCacheDir(); err == nil { // if NO error
		return filepath.Join(dir, "git-get")
	}
	if dir, err := os.UserHomeDir(); err == nil { // if NO error
		return filepath.Join(dir, ".cache", "git-get")
	}
	return filepath.Join(".cache", "git-get")
}

func setup(cmd *cobra.Command, prefix string) {
	flags.Init(cmd.PersistentFlags(), prefix)
	flags.Init(cmd.Flags(), prefix)
	for _, cmd := range cmd.Commands() {
		setup(cmd, fmt.Sprintf("%s_%s_", prefix, strings.ToUpper(cmd.Name())))
	}
}

func validArgs(
	cmd *cobra.Command, args []string, toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) == 0 {
		list := []cobra.Completion{
			"git@github.com:",
			"https://",
			"https://github.com/",
		}
		if toComplete != "" {
			list = slices.DeleteFunc(list, func(s string) bool {
				return !strings.HasPrefix(s, toComplete)
			})
		}
		return list, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		return nil, cobra.ShellCompDirectiveFilterDirs
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func runE(cmd *cobra.Command, args []string) error {
	log.Info(cmd.Name(), flags.Log(cmd), slog.Any("args", args))
	exclude, err := excludes.New(results.Must1(cmd.LocalFlags().GetStringSlice("exclude")))
	if err != nil {
		return err
	}
	cacheDir := results.Must1(cmd.LocalFlags().GetString("cache-dir"))
	if len(args) == 0 {
		return pflag.ErrHelp
	}
	repo := args[0]
	url, err := giturl.Parse(repo)
	if err != nil {
		return err
	}
	log.Debug("url", "url", url)
	var target string
	if len(args) >= 2 {
		target = args[1]
	}
	if target == "" {
		if url.Path == "/" {
			target = url.Name
		} else {
			target = filepath.Base(url.Path)
		}
	}
	target = filepath.Clean(filepath.Join(".", target))
	if _, err := stacktrace.Trace2(os.Stat(target)); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	} else {
		if !results.Must1(cmd.LocalFlags().GetBool("force")) {
			return fmt.Errorf("%q, file already exists", target)
		}
	}
	repoDir := filepath.Join(cacheDir, "repos", url.Host, url.Owner, url.Name)
	log.Info("url", "url", url, "repoDir", repoDir)
	if _, err := stacktrace.Trace2(os.Lstat(repoDir)); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err := gitClone(cmd, url, repoDir); err != nil {
			return err
		}
	} else {
		if err := gitFetch(cmd, url, repoDir); err != nil {
			return err
		}
	}
	repoCO := filepath.Join(cacheDir, "works", ulid.Make().String())
	log.Info("url", "url", url, "repoDir", repoDir, "repoCO", repoCO)
	if err := gitCheckout(cmd, url, repoDir, repoCO); err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(repoCO)
	}()
	fmt.Fprintf(cmd.OutOrStdout(), "get %q to %q\n", repo, target)
	if err := stacktrace.Trace(os.RemoveAll(target)); err != nil {
		return err
	}
	return copyTree(cmd, url, repoCO, target, exclude)
}

func gitClone(cmd *cobra.Command, url *giturl.URL, repoDir string) error {
	git := exec.Command("git", "clone", "--bare", url.Remote, repoDir)
	git.Stdout = cmd.OutOrStdout()
	git.Stderr = cmd.ErrOrStderr()
	return git.Run()
}

func gitFetch(cmd *cobra.Command, url *giturl.URL, repoDir string) error {
	_ = url
	// TODO: url.Remote を clone したディレクトリ？

	git := exec.Command("git", "--git-dir", repoDir, "fetch", "--all")
	git.Stdout = cmd.OutOrStdout()
	git.Stderr = cmd.ErrOrStderr()
	return git.Run()
}

func gitCheckout(cmd *cobra.Command, url *giturl.URL, repoDir, repoCO string) error {
	git := exec.Command("git", "clone", "--no-checkout", repoDir, repoCO)
	git.Stdout = cmd.OutOrStdout()
	git.Stderr = cmd.ErrOrStderr()
	if err := git.Run(); err != nil {
		return err
	}
	git = exec.Command("git", "switch", "--detach", url.Commit)
	git.Stdout = cmd.OutOrStdout()
	git.Stderr = cmd.ErrOrStderr()
	git.Dir = repoCO
	return git.Run()
}

func copyTree(_ *cobra.Command, url *giturl.URL, repoCO, target string, exclude *excludes.Matcher) error {
	log.Debug("copyTree", "url", url, "repoCO", repoCO, "target", target)
	src := filepath.Join(repoCO, url.Path)
	if _, err := stacktrace.Trace2(os.Stat(src)); err != nil {
		return err
	}
	srcFS := os.DirFS(src)
	return fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if exclude.PathMatch(path) {
			return nil
		}
		if path == "." {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			return nil
		}
		if d.IsDir() {
			if err := os.MkdirAll(filepath.Join(target, path), 0755); err != nil {
				return err
			}
		} else {
			if err := copyFile(filepath.Join(target, path), srcFS, path); err != nil {
				return err
			}
		}
		return nil
	})
}

func copyFile(dst string, srcFS fs.FS, src string) error {
	file, err := stacktrace.Trace2(srcFS.Open(src))
	if err != nil {
		return err
	}
	defer file.Close()

	ofile, err := stacktrace.Trace2(os.Create(dst))
	if err != nil {
		return err
	}
	defer ofile.Close()

	if _, err := stacktrace.Trace2(io.Copy(ofile, file)); err != nil {
		return err
	}
	return ofile.Close()
}
