package axe

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GitInfo holds git repository metadata
type GitInfo struct {
	Version    string    `json:"version"`
	Repository string    `json:"repository"`
	Modified   time.Time `json:"modified"`
}

func BuildGitInfo(path string, gitbin string) (*GitInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("Path must not be empty")
	}
	isDir, err := fileIsDirOrDirLink(path)
	if err != nil {
		return nil, fmt.Errorf("Check path: %w", err)
	}
	if !isDir {
		return nil, fmt.Errorf("Path must be a directory")
	}

	useGit := false
	gitInfo := &GitInfo{}

	// check for git bin available
	if useGit {
		// check for dir is a repo
		if err = gitInfo.ReadRepository(path, gitbin); err == nil {
			useGit = true
		}
	}

	now := time.Now() // MAYBE: last change of dir content?
	if !useGit {
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("Get absolute path: %w", err)
		}
		gitInfo.Repository = "file://" + abs
		gitInfo.Version = "v0.0.0-" + now.Format("20060102150405")
		gitInfo.Modified = now
		return gitInfo, nil
	}

	if err = gitInfo.ReadVersion(path, gitbin); err != nil {
		gitInfo.Version = "v0.0.0-" + now.Format("20060102150405")
	}
	if err = gitInfo.ReadModified(path, gitbin); err != nil {
		gitInfo.Modified = now
	}
	return gitInfo, nil
}

// Version fills rv with package version from git
func (gi *GitInfo) ReadVersion(path string, gitbin string) error {
	out, err := exec.Command(gitbin, "-C", path, "describe", "--tags", "--always").Output()
	if err != nil {
		return fmt.Errorf("Git describe: %w", err)
	}
	gi.Version = strings.TrimSuffix(string(out), "\n")
	return nil
}

// Repository fills rv with package repo from git
func (gi *GitInfo) ReadRepository(path string, gitbin string) error {
	out, err := exec.Command(gitbin, "-C", path, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return fmt.Errorf("Git config: %w", err)
	}
	gi.Repository = strings.TrimSuffix(string(out), "\n")
	return nil
}

func MkTime(in []byte) (time.Time, error) {
	tm, err := strconv.ParseInt(string(in), 10, 64)
	if err != nil {
		return time.Now(), err
	}
	return time.Unix(tm, 0), nil
}

// Modified fills rv with package last commit timestamp
func (gi *GitInfo) ReadModified(path string, gitbin string) error {
	out, err := exec.Command(gitbin, "-C", path, "show", "-s", "--format=format:%ct", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("Git show: %w", err)
	}
	gi.Modified, err = MkTime(out)
	return nil
}

// fileIsDirOrDirLink returns true if path is a dir or symlink to dir
func fileIsDirOrDirLink(path string) (bool, error) {
	file, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if file.IsDir() {
		return true, nil
	}
	if file.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}
	// check symlink
	var linkSrc string
	linkDst := filepath.Join(path, file.Name())
	linkSrc, err = filepath.EvalSymlinks(linkDst)
	if err == nil {
		var fi os.FileInfo
		fi, err = os.Lstat(linkSrc)
		if err == nil {
			return fi.IsDir(), nil
		}
	}
	return false, err
}
