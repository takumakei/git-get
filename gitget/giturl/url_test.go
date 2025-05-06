package giturl

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrl(t *testing.T) {
	assert.Equal(t, 42, 42)
	require.Equal(t, 42, 42)

	osHostname = func() (string, error) {
		return "git.example", nil
	}
	userCurrent = func() (*user.User, error) {
		return &user.User{Name: "gecos"}, nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		name = filepath.Base(filepath.Dir(filepath.Clean(name)))
		if name == "repo" {
			return nil, nil
		}
		return nil, os.ErrNotExist
	}

	testCases := []struct {
		Name   string
		Target string
		Expect *URL
	}{
		{
			"local root",
			"/home/abcdefg/documents/repo.git",
			&URL{
				Remote: "/home/abcdefg/documents/repo.git",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/",
				Commit: "HEAD",
			},
		},

		{
			"local root commit",
			"/home/abcdefg/documents/repo.git@1a2b3c5",
			&URL{
				Remote: "/home/abcdefg/documents/repo.git",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/",
				Commit: "1a2b3c5",
			},
		},

		{
			"local dir",
			"/home/abcdefg/documents/repo.git/path/to/dir",
			&URL{
				Remote: "/home/abcdefg/documents/repo.git",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "HEAD",
			},
		},

		{
			"local dir commit",
			"/home/abcdefg/documents/repo.git/path/to/dir@1a2b3c5",
			&URL{
				Remote: "/home/abcdefg/documents/repo.git",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "1a2b3c5",
			},
		},

		{
			"local stat root",
			"/home/abcdefg/documents/repo",
			&URL{
				Remote: "/home/abcdefg/documents/repo",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/",
				Commit: "HEAD",
			},
		},

		{
			"local stat root commit",
			"/home/abcdefg/documents/repo@1a2b3c5",
			&URL{
				Remote: "/home/abcdefg/documents/repo",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/",
				Commit: "1a2b3c5",
			},
		},

		{
			"local stat dir",
			"/home/abcdefg/documents/repo/path/to/dir",
			&URL{
				Remote: "/home/abcdefg/documents/repo",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "HEAD",
			},
		},

		{
			"local stat dir commit",
			"/home/abcdefg/documents/repo/path/to/dir@1a2b3c5",
			&URL{
				Remote: "/home/abcdefg/documents/repo",
				Host:   "git.example",
				Owner:  "gecos",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "1a2b3c5",
			},
		},

		{
			"http root",
			"https://github.com/abcdefg/repo.git",
			&URL{
				Remote: "https://github.com/abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/",
				Commit: "HEAD",
			},
		},

		{
			"http root",
			"https://github.com/abcdefg/repo.git@1a2b3c5",
			&URL{
				Remote: "https://github.com/abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/",
				Commit: "1a2b3c5",
			},
		},

		{
			"http root",
			"https://github.com/abcdefg/repo.git/path/to/dir",
			&URL{
				Remote: "https://github.com/abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "HEAD",
			},
		},

		{
			"http root",
			"https://github.com/abcdefg/repo.git/path/to/dir@1a2b3c5",
			&URL{
				Remote: "https://github.com/abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "1a2b3c5",
			},
		},

		{
			"ssh root",
			"git@github.com:abcdefg/repo.git",
			&URL{
				Remote: "git@github.com:abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/",
				Commit: "HEAD",
			},
		},

		{
			"ssh root",
			"git@github.com:abcdefg/repo.git@1a2b3c5",
			&URL{
				Remote: "git@github.com:abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/",
				Commit: "1a2b3c5",
			},
		},

		{
			"ssh root",
			"git@github.com:abcdefg/repo.git/path/to/dir",
			&URL{
				Remote: "git@github.com:abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "HEAD",
			},
		},

		{
			"ssh root",
			"git@github.com:abcdefg/repo.git/path/to/dir@1a2b3c5",
			&URL{
				Remote: "git@github.com:abcdefg/repo.git",
				Host:   "github.com",
				Owner:  "abcdefg",
				Name:   "repo",
				Path:   "/path/to/dir",
				Commit: "1a2b3c5",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := Parse(tc.Target)
			require.NoError(t, err)
			require.Equal(t, tc.Expect, actual)
		})
	}
}
