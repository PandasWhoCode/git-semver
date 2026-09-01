package next

import (
	"github.com/PandasWhoCode/git-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_Next_should_return_first_version_if_repository_has_no_tags(t *testing.T) {
	workdir := initRepo(t, commit{message: "feat: initial commit"})

	version, err := Next(NextOptions{
		Workdir:            workdir,
		Stable:             true,
		MajorVersionFilter: -1,
	})

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", version.ToString())
}

func Test_Next_should_return_next_version_if_repository_only_has_pre_release_tags(t *testing.T) {
	workdir := initRepo(
		t,
		commit{message: "feat: initial commit", tags: []string{"v0.1.0-alpha.1"}},
		commit{message: "fix: a fix"},
	)

	version, err := Next(NextOptions{
		Workdir:            workdir,
		Stable:             false,
		MajorVersionFilter: -1,
	})

	require.NoError(t, err)
	assert.Equal(t, "0.1.0", version.ToString())
}

func Test_Next_should_increment_pre_release_counter_if_repository_only_has_pre_release_tags(t *testing.T) {
	workdir := initRepo(
		t,
		commit{message: "feat: initial commit", tags: []string{"v0.1.0-alpha.1"}},
		commit{message: "fix: a fix", tags: []string{"v0.1.0-alpha.2"}},
		commit{message: "fix: another fix"},
	)

	version, err := Next(NextOptions{
		Workdir:            workdir,
		Stable:             false,
		MajorVersionFilter: -1,
		PreReleaseOptions: semver.PreReleaseOptions{
			Label:         "alpha",
			AppendCounter: true,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "0.1.0-alpha.3", version.ToString())
}

type commit struct {
	message string
	tags    []string
}

// initRepo creates a repository containing the given commits and returns its working directory.
func initRepo(t *testing.T, commits ...commit) string {
	t.Helper()

	workdir := t.TempDir()

	repo, err := git.PlainInit(workdir, false)
	assert.NoError(t, err)

	worktree, err := repo.Worktree()
	assert.NoError(t, err)

	for i, c := range commits {
		var hash plumbing.Hash

		hash, err = worktree.Commit(c.message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "git-semver",
				Email: "git-semver@example.com",
				When:  time.Now().Add(time.Duration(i) * time.Second),
			},
			AllowEmptyCommits: true,
		})
		assert.NoError(t, err)

		for _, tagName := range c.tags {
			_, err = repo.CreateTag(tagName, hash, nil)
			assert.NoError(t, err)
		}
	}

	return workdir
}
