package latest

import (
	"github.com/PandasWhoCode/git-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_tagNameToVersion_should_return_version(t *testing.T) {
	version := tagNameToVersion("1.2.3")

	assert.Equal(
		t,
		&semver.Version{
			Major:         1,
			Minor:         2,
			Patch:         3,
			PreReleaseTag: []interface{}{},
		},
		version,
	)
}

func Test_tagNameToVersion_should_return_version_if_tag_has_v_prefix(t *testing.T) {
	version := tagNameToVersion("v1.2.3")

	assert.Equal(
		t,
		&semver.Version{
			Major:         1,
			Minor:         2,
			Patch:         3,
			PreReleaseTag: []interface{}{},
		},
		version,
	)
}

func Test_FindLatestVersion_should_return_no_version_if_repository_has_no_tags(t *testing.T) {
	repo, _ := initRepo(t)

	version, tag, err := FindLatestVersion(repo, -1, false)

	require.NoError(t, err)
	assert.Nil(t, version)
	assert.Nil(t, tag)
}

func Test_FindLatestVersion_should_return_no_version_if_repository_only_has_pre_release_tags(t *testing.T) {
	repo, _ := initRepo(t, "v0.1.0-alpha.1", "v0.1.0-alpha.2")

	version, tag, err := FindLatestVersion(repo, -1, false)

	require.NoError(t, err)
	assert.Nil(t, version)
	assert.Nil(t, tag)
}

func Test_FindLatestVersion_should_return_latest_pre_release_if_pre_releases_are_included(t *testing.T) {
	repo, _ := initRepo(t, "v0.1.0-alpha.1", "v0.1.0-alpha.2")

	version, tag, err := FindLatestVersion(repo, -1, true)

	require.NoError(t, err)
	assert.Equal(t, "0.1.0-alpha.2", version.ToString())
	assert.Equal(t, "v0.1.0-alpha.2", tag.Name().Short())
}

func Test_FindLatestVersion_should_ignore_pre_release_tags_if_a_release_tag_exists(t *testing.T) {
	repo, _ := initRepo(t, "v0.1.0", "v0.2.0-alpha.1")

	version, tag, err := FindLatestVersion(repo, -1, false)

	require.NoError(t, err)
	assert.Equal(t, "0.1.0", version.ToString())
	assert.Equal(t, "v0.1.0", tag.Name().Short())
}

func Test_Latest_should_return_empty_version_if_repository_has_no_tags(t *testing.T) {
	_, workdir := initRepo(t)

	version, err := Latest(LatestOptions{Workdir: workdir})

	require.NoError(t, err)
	assert.Equal(t, "0.0.0", version.ToString())
}

func Test_Latest_should_return_empty_version_if_repository_only_has_pre_release_tags(t *testing.T) {
	_, workdir := initRepo(t, "v0.1.0-alpha.1")

	version, err := Latest(LatestOptions{Workdir: workdir})

	require.NoError(t, err)
	assert.Equal(t, "0.0.0", version.ToString())
}

// initRepo creates a repository with a single commit which is tagged with all given tag names.
func initRepo(t *testing.T, tagNames ...string) (*git.Repository, string) {
	t.Helper()

	workdir := t.TempDir()

	repo, err := git.PlainInit(workdir, false)
	assert.NoError(t, err)

	worktree, err := repo.Worktree()
	assert.NoError(t, err)

	commit, err := worktree.Commit("feat: initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "git-semver",
			Email: "git-semver@example.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: true,
	})
	assert.NoError(t, err)

	for _, tagName := range tagNames {
		_, err = repo.CreateTag(tagName, commit, nil)
		assert.NoError(t, err)
	}

	return repo, workdir
}
