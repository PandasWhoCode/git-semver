package latest

import (
	"github.com/PandasWhoCode/git-semver/logger"
	"github.com/PandasWhoCode/git-semver/semver"
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
	// We no longer need to capture the output of the command
	_, err = cmd.Output()

	if err != nil {
		return nil, err
	}

	// Retrieve the tags from the repository
	tagIter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	defer tagIter.Close()

	var foundTags []*plumbing.Reference

	// Loop through all tags and filter those that match the semantic versioning format
	for tag, err := tagIter.Next(); err != io.EOF; tag, err = tagIter.Next() {
		if err != nil {
			return nil, err
		}

		// Skip non-semantic version tags (i.e., tags that don't start with 'v')
		if !strings.HasPrefix(tag.Name().Short(), "v") {
			continue
		}

		// Convert the tag to a semver version
		version, err := semver.ParseVersion(tag.Name().Short())
		if err != nil {
			// If parsing fails, skip this tag
			continue
		}

		// Apply pre-release filter and major version filter
		if (!includePreReleases && len(version.PreReleaseTag) > 0) || (majorVersionFilter >= 0 && version.Major != majorVersionFilter) {
			continue
		}

		// Add the tag to the list of valid semantic version tags
		foundTags = append(foundTags, tag)
	}

	// If no valid tags were found
	if len(foundTags) == 0 {
		return nil, errors.New("no matching semantic version tags found")
	}

	// Find the highest version tag
	var latestTag *plumbing.Reference
	var maxVersion *semver.Version
	for _, tag := range foundTags {
		version, err := semver.ParseVersion(tag.Name().Short())
		if err != nil {
			// Skip invalid versions
			continue
		}
		if maxVersion == nil || semver.CompareVersions(version, maxVersion) > 0 {
			maxVersion = version
			latestTag = tag
		}
	}

	if latestTag == nil {
		return nil, errors.New("no valid semantic version tag found")
	}

	return latestTag, nil
}

func tagNameToVersion(tagName string) *semver.Version {

	version, err := semver.ParseVersion(tagName)

	if err != nil {
		logger.Logger.Debug(err, ": Tag: ", tagName)
		return nil
	}

	return version
}
