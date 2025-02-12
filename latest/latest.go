package latest

import (
	"github.com/andrewb1269hg/git-semver/logger"
	"github.com/andrewb1269hg/git-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"io"
	"os/exec"
	"strings"
)

type LatestOptions struct {
	Workdir            string
	IncludePreReleases bool
	MajorVersionFilter int
}

func Latest(options LatestOptions) (*semver.Version, error) {

	repo, err := git.PlainOpenWithOptions(options.Workdir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "Could not open git repository")
	}

	latestReleaseVersion, _, err := FindLatestVersion(repo, options.MajorVersionFilter, options.IncludePreReleases)

	if latestReleaseVersion == nil {
		latestReleaseVersion = &semver.EmptyVersion
	}

	return latestReleaseVersion, err

}

func FindLatestVersion(repo *git.Repository, majorVersionFilter int, preRelease bool) (*semver.Version, *plumbing.Reference, error) {
	latestVersionTag, err := findLatestVersionTag(repo, majorVersionFilter, preRelease)

	if err != nil {
		return nil, nil, err
	}

	if latestVersionTag == nil {
		return nil, nil, nil
	}

	return tagNameToVersion(latestVersionTag.Name().Short()), latestVersionTag, nil
}

func findLatestVersionTag(repo *git.Repository, majorVersionFilter int, includePreReleases bool) (*plumbing.Reference, error) {
	// Use git rev-list to get the latest tag from all branches, not just the current branch
	cmd := exec.Command("git", "rev-list", "--tags", "--max-count=1")
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	cmd.Dir = worktree.Filesystem.Root() // Use the working directory for git commands
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	latestTagHash := strings.TrimSpace(string(output))

	// Retrieve the tag by its hash
	tagIter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	defer tagIter.Close()

	var foundTag *plumbing.Reference
	for tag, err := tagIter.Next(); err != io.EOF; tag, err = tagIter.Next() {
		if err != nil {
			return nil, err
		}

		// Check if this tag matches the latest tag hash
		if tag.Hash().String() == latestTagHash {
			foundTag = tag
			break
		}
	}

	if foundTag == nil {
		return nil, errors.New("no matching tag found")
	}

	// Convert the tag to a semver version and apply filters
	version := tagNameToVersion(foundTag.Name().Short())

	if version == nil || (!includePreReleases && len(version.PreReleaseTag) > 0) {
		return nil, nil
	}

	// Apply the major version filter
	if majorVersionFilter >= 0 && version.Major != majorVersionFilter {
		return nil, nil
	}

	return foundTag, nil
}

func tagNameToVersion(tagName string) *semver.Version {

	version, err := semver.ParseVersion(tagName)

	if err != nil {
		logger.Logger.Debug(err, ": Tag: ", tagName)
		return nil
	}

	return version
}
