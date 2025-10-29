package corepkg

import (
	"bufio"
	"os/exec"
	"strings"
)

type GitVersionInfo struct {
	Branch     string
	Commit     string
	Tag        string
	Author     string
	CommitDate string
}

func NewGitVersionInfo(absPath string) *GitVersionInfo {
	vi := &GitVersionInfo{Branch: "unknown", Commit: "unknown", Tag: "unknown", CommitDate: "unknown"}

	// Run 'git log -n 1 --pretty=fuller' and parse the output
	// git log -n 1 --pretty='BRANCH=%d%nCOMMITNAME=%cn%nCOMMITDATE=%cd%nCOMMITHASH=%H'
	cmd := exec.Command("git", "log", "-n", "1", "--pretty=BRANCH=%d%nCOMMITNAME=%cn%nCOMMITDATE=%cd%nCOMMITHASH=%H%")
	cmd.Dir = absPath
	raw, err := cmd.Output()
	output := string(raw)

	if err != nil {
		return vi
	}

	// Example output:
	// BRANCH=(HEAD -> master, tag: BEAR, origin/master, origin/HEAD)
	// COMMITHASH=71c422a84b7b2d4c1af58538ce902b0fef646902
	// COMMITNAME: jurgen.kluft
	// COMMITDATE: Sat May 17 20:06:22 2025 +0800

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "COMMITDATE") {
			vi.CommitDate = strings.TrimPrefix(line, "COMMITDATE=")
		} else if strings.HasPrefix(line, "COMMITNAME") {
			vi.Author = strings.TrimPrefix(line, "COMMITNAME=")
		} else if strings.HasPrefix(line, "COMMITHASH") {
			vi.Commit = strings.TrimPrefix(line, "COMMITHASH=")
		} else if strings.HasPrefix(line, "BRANCH") {
			branchDescr := strings.TrimPrefix(line, "BRANCH=")
			branchDescr = strings.TrimSpace(branchDescr)
			branchDescr = strings.TrimSpace(strings.Trim(branchDescr, "()"))
			if strings.HasPrefix(branchDescr, "HEAD") {
				branchDescr = strings.TrimPrefix(branchDescr, "HEAD -> ")
				branchFields := strings.Split(branchDescr, ",")
				for i, field := range branchFields {
					field = strings.TrimSpace(field)
					if i == 0 {
						vi.Branch = strings.TrimSpace(field)
					} else if strings.HasPrefix(field, "tag:") {
						vi.Tag = strings.TrimSpace(strings.TrimPrefix(field, "tag:"))
					}
				}
			}
		}
	}

	return vi
}
