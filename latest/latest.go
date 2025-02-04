package latest

import (
  "github.com/blang/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/psanetra/git-semver/logger"
	"github.com/psanetra/git-semver/semver"
	"io"
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

	tagIter, err := repo.Tags()

	if err != nil {
		return nil, err
	}

	defer tagIter.Close()

	var maxVersionTag *plumbing.Reference
	var maxVersion = &semver.EmptyVersion

	for tag, err := tagIter.Next(); err != io.EOF; tag, err = tagIter.Next() {
		if err != nil {
			return nil, err
		}

		version := tagNameToVersion(tag.Name().Short())

		if version == nil || !includePreReleases && len(version.PreReleaseTag) > 0 {
			continue
		}

		if (majorVersionFilter < 0 || majorVersionFilter == version.Major) && semver.CompareVersions(version, maxVersion) > 0 {
			maxVersion = version
			maxVersionTag = tag
		}
	}

	return maxVersionTag, nil
}

func tagNameToVersion(tagName string) *semver.Version {

	version, err := semver.ParseVersion(tagName)

	if err != nil {
		logger.Logger.Debug(err, ": Tag: ", tagName)
		return nil
	}

	return version
}

func FindLatestVersionOnBranch(repo *git.Repository, majorVersionFilter int, branchName string, preRelease bool) (*semver.Version, *plumbing.Reference, error) {
    iter, err := repo.Tags()
    if err != nil {
        return nil, nil, err
    }

    var latestVersion *semver.Version
    var latestTag *plumbing.Reference

    err = iter.ForEach(func(ref *plumbing.Reference) error {
        commit, err := repo.CommitObject(ref.Hash())
        if err != nil {
            return nil
        }

        // Ensure the commit is on the specified branch
        isMerged, err := repo.getMergeBase(commit, branchName)
        if err != nil || !isMerged {
            return nil // Skip tags not in the default branch history
        }

        version, err := semver.ParseTolerant(ref.Name().Short())
        if err != nil {
            return nil
        }

        if latestVersion == nil || version.GreaterThan(latestVersion) {
            latestVersion = &version
            latestTag = ref
        }

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    return latestVersion, latestTag, nil
}

func getMergeBase(repoPath, branchA, branchB string) (string, error) {
    cmd := exec.Command("git", "-C", repoPath, "merge-base", branchA, branchB)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

