package giturl

import (
	"errors"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goaux/stacktrace/v2"
)

type URL struct {
	Remote string

	Host  string
	Owner string
	Name  string

	Commit string
	Path   string
}

var ErrInvalidTarget = errors.New("invalid target")

func Parse(target string) (*URL, error) {
	if isLocal(target) {
		lhs, commit := parseCommit(target)
		remote, path, err := stacktrace.Trace3(findLocalGit(lhs))
		if err != nil {
			return nil, err
		}
		host, err := stacktrace.Trace2(osHostname())
		if err != nil {
			return nil, err
		}
		owner, err := stacktrace.Trace2(userCurrent())
		if err != nil {
			return nil, err
		}
		url := &URL{
			Remote: remote,
			Host:   host,
			Owner:  owner.Name,
			Name:   parseName(remote),
			Commit: commit,
			Path:   path,
		}
		return url, nil
	}

	if m := httpRe.FindStringSubmatch(target); len(m) > 0 {
		left := m[1]
		host := m[2]
		last := m[3]
		lhs, commit := parseCommit(last)
		remote, path, err := stacktrace.Trace3(findRemoteGit(lhs))
		if err != nil {
			return nil, err
		}
		owner, name := parseOwnerName(remote)
		url := &URL{
			Remote: left + remote,
			Host:   host,
			Owner:  owner,
			Name:   name,
			Commit: commit,
			Path:   path,
		}
		return url, nil
	}

	if m := sshRe.FindStringSubmatch(target); len(m) > 0 {
		left := m[1]
		host := m[2]
		last := m[3]
		lhs, commit := parseCommit(last)
		remote, path, err := stacktrace.Trace3(findRemoteGit(lhs))
		if err != nil {
			return nil, err
		}
		owner, name := parseOwnerName(remote)
		url := &URL{
			Remote: left + remote,
			Host:   host,
			Owner:  owner,
			Name:   name,
			Commit: commit,
			Path:   path,
		}
		return url, nil
	}

	return nil, ErrInvalidTarget
}

var pathSeparator = string([]rune{os.PathSeparator})
var httpRe = regexp.MustCompile(`^(https?://(?:([^:/]+)(?::[0-9]+)?))(.*)$`)
var sshRe = regexp.MustCompile(`^((?:[^@]+@)?([^:]+):)(.*)$`)

func isLocal(target string) bool {
	switch {
	case strings.HasPrefix(target, pathSeparator):
		return true
	case strings.HasPrefix(target, "."):
		return true
	case filepath.VolumeName(target) != "":
		return true
	}
	return false
}

func parseCommit(target string) (lhs, commit string) {
	if i := strings.LastIndexByte(target, '@'); i >= 0 {
		return target[:i], target[i+1:]
	}
	return target, "HEAD"
}

func findLocalGit(target string) (remote, dir string, err error) {
	target = filepath.Clean(target)
	if target == "" {
		err = ErrInvalidTarget
		return
	}
	if filepath.Ext(target) == ".git" {
		return target, "/", nil
	}
	if _, err := osStat(filepath.Join(target, ".git")); err == nil { // if NO error
		return target, "/", nil
	}
	parent, sub := filepath.Split(target)
	remote, dir, err = findLocalGit(parent)
	if err != nil {
		return "", "", err
	}
	return remote, filepath.Join(dir, sub), nil
}

func parseName(target string) string {
	s := filepath.Base(target)
	return strings.TrimSuffix(s, filepath.Ext(s))
}

func findRemoteGit(target string) (remote, dir string, err error) {
	remote, dir, err = findRemoteGitR(target)
	if err != nil {
		if errors.Is(err, ErrInvalidTarget) {
			return target, "/", nil
		}
		return "", "", err
	}
	return remote, dir, nil
}

func findRemoteGitR(target string) (remote, dir string, err error) {
	target = path.Clean(target)
	if target == "" || target == "." || target == "/" {
		err = ErrInvalidTarget
		return
	}
	if path.Ext(target) == ".git" {
		return target, "/", nil
	}
	parent, sub := path.Split(target)
	remote, dir, err = findRemoteGitR(parent)
	if err != nil {
		return "", "", err
	}
	return remote, path.Join(dir, sub), nil
}

func parseOwnerName(remote string) (owner, name string) {
	owner = path.Base(path.Dir(remote))
	name = strings.TrimSuffix(path.Base(remote), path.Ext(remote))
	return
}

var osHostname = os.Hostname

var userCurrent = user.Current

var osStat = os.Stat
