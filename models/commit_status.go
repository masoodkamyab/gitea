// Copyright 2017 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"container/list"
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/setting"
	api "github.com/masoodkamyab/gitea/modules/structs"
	"github.com/masoodkamyab/gitea/modules/util"
)

// CommitStatusState holds the state of a Status
// It can be "pending", "success", "error", "failure", and "warning"
type CommitStatusState string

// IsWorseThan returns true if this State is worse than the given State
func (css CommitStatusState) IsWorseThan(css2 CommitStatusState) bool {
	switch css {
	case CommitStatusError:
		return true
	case CommitStatusFailure:
		return css2 != CommitStatusError
	case CommitStatusWarning:
		return css2 != CommitStatusError && css2 != CommitStatusFailure
	case CommitStatusSuccess:
		return css2 != CommitStatusError && css2 != CommitStatusFailure && css2 != CommitStatusWarning
	default:
		return css2 != CommitStatusError && css2 != CommitStatusFailure && css2 != CommitStatusWarning && css2 != CommitStatusSuccess
	}
}

const (
	// CommitStatusPending is for when the Status is Pending
	CommitStatusPending CommitStatusState = "pending"
	// CommitStatusSuccess is for when the Status is Success
	CommitStatusSuccess CommitStatusState = "success"
	// CommitStatusError is for when the Status is Error
	CommitStatusError CommitStatusState = "error"
	// CommitStatusFailure is for when the Status is Failure
	CommitStatusFailure CommitStatusState = "failure"
	// CommitStatusWarning is for when the Status is Warning
	CommitStatusWarning CommitStatusState = "warning"
)

// CommitStatus holds a single Status of a single Commit
type CommitStatus struct {
	ID          int64             `xorm:"pk autoincr"`
	Index       int64             `xorm:"INDEX UNIQUE(repo_sha_index)"`
	RepoID      int64             `xorm:"INDEX UNIQUE(repo_sha_index)"`
	Repo        *Repository       `xorm:"-"`
	State       CommitStatusState `xorm:"VARCHAR(7) NOT NULL"`
	SHA         string            `xorm:"VARCHAR(64) NOT NULL INDEX UNIQUE(repo_sha_index)"`
	TargetURL   string            `xorm:"TEXT"`
	Description string            `xorm:"TEXT"`
	ContextHash string            `xorm:"char(40) index"`
	Context     string            `xorm:"TEXT"`
	Creator     *User             `xorm:"-"`
	CreatorID   int64

	CreatedUnix util.TimeStamp `xorm:"INDEX created"`
	UpdatedUnix util.TimeStamp `xorm:"INDEX updated"`
}

func (status *CommitStatus) loadRepo(e Engine) (err error) {
	if status.Repo == nil {
		status.Repo, err = getRepositoryByID(e, status.RepoID)
		if err != nil {
			return fmt.Errorf("getRepositoryByID [%d]: %v", status.RepoID, err)
		}
	}
	if status.Creator == nil && status.CreatorID > 0 {
		status.Creator, err = getUserByID(e, status.CreatorID)
		if err != nil {
			return fmt.Errorf("getUserByID [%d]: %v", status.CreatorID, err)
		}
	}
	return nil
}

// APIURL returns the absolute APIURL to this commit-status.
func (status *CommitStatus) APIURL() string {
	_ = status.loadRepo(x)
	return fmt.Sprintf("%sapi/v1/%s/statuses/%s",
		setting.AppURL, status.Repo.FullName(), status.SHA)
}

// APIFormat assumes some fields assigned with values:
// Required - Repo, Creator
func (status *CommitStatus) APIFormat() *api.Status {
	_ = status.loadRepo(x)
	apiStatus := &api.Status{
		Created:     status.CreatedUnix.AsTime(),
		Updated:     status.CreatedUnix.AsTime(),
		State:       api.StatusState(status.State),
		TargetURL:   status.TargetURL,
		Description: status.Description,
		ID:          status.Index,
		URL:         status.APIURL(),
		Context:     status.Context,
	}
	if status.Creator != nil {
		apiStatus.Creator = status.Creator.APIFormat()
	}

	return apiStatus
}

// CalcCommitStatus returns commit status state via some status, the commit statues should order by id desc
func CalcCommitStatus(statuses []*CommitStatus) *CommitStatus {
	var lastStatus *CommitStatus
	var state CommitStatusState
	for _, status := range statuses {
		if status.State.IsWorseThan(state) {
			state = status.State
			lastStatus = status
		}
	}
	if lastStatus == nil {
		if len(statuses) > 0 {
			lastStatus = statuses[0]
		} else {
			lastStatus = &CommitStatus{}
		}
	}
	return lastStatus
}

// GetCommitStatuses returns all statuses for a given commit.
func GetCommitStatuses(repo *Repository, sha string, page int) ([]*CommitStatus, error) {
	statuses := make([]*CommitStatus, 0, 10)
	return statuses, x.Limit(10, page*10).Where("repo_id = ?", repo.ID).And("sha = ?", sha).Find(&statuses)
}

// GetLatestCommitStatus returns all statuses with a unique context for a given commit.
func GetLatestCommitStatus(repo *Repository, sha string, page int) ([]*CommitStatus, error) {
	ids := make([]int64, 0, 10)
	err := x.Limit(10, page*10).
		Table(&CommitStatus{}).
		Where("repo_id = ?", repo.ID).And("sha = ?", sha).
		Select("max( id ) as id").
		GroupBy("context_hash").OrderBy("max( id ) desc").Find(&ids)
	if err != nil {
		return nil, err
	}
	statuses := make([]*CommitStatus, 0, len(ids))
	if len(ids) == 0 {
		return statuses, nil
	}
	return statuses, x.In("id", ids).Find(&statuses)
}

// NewCommitStatusOptions holds options for creating a CommitStatus
type NewCommitStatusOptions struct {
	Repo         *Repository
	Creator      *User
	SHA          string
	CommitStatus *CommitStatus
}

// NewCommitStatus save commit statuses into database
func NewCommitStatus(opts NewCommitStatusOptions) error {
	if opts.Repo == nil {
		return fmt.Errorf("NewCommitStatus[nil, %s]: no repository specified", opts.SHA)
	}

	repoPath := opts.Repo.RepoPath()
	if opts.Creator == nil {
		return fmt.Errorf("NewCommitStatus[%s, %s]: no user specified", repoPath, opts.SHA)
	}

	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("NewCommitStatus[repo_id: %d, user_id: %d, sha: %s]: %v", opts.Repo.ID, opts.Creator.ID, opts.SHA, err)
	}

	opts.CommitStatus.Description = strings.TrimSpace(opts.CommitStatus.Description)
	opts.CommitStatus.Context = strings.TrimSpace(opts.CommitStatus.Context)
	opts.CommitStatus.TargetURL = strings.TrimSpace(opts.CommitStatus.TargetURL)
	opts.CommitStatus.SHA = opts.SHA
	opts.CommitStatus.CreatorID = opts.Creator.ID
	opts.CommitStatus.RepoID = opts.Repo.ID

	// Get the next Status Index
	var nextIndex int64
	lastCommitStatus := &CommitStatus{
		SHA:    opts.SHA,
		RepoID: opts.Repo.ID,
	}
	has, err := sess.Desc("index").Limit(1).Get(lastCommitStatus)
	if err != nil {
		if err := sess.Rollback(); err != nil {
			log.Error("NewCommitStatus: sess.Rollback: %v", err)
		}
		return fmt.Errorf("NewCommitStatus[%s, %s]: %v", repoPath, opts.SHA, err)
	}
	if has {
		log.Debug("NewCommitStatus[%s, %s]: found", repoPath, opts.SHA)
		nextIndex = lastCommitStatus.Index
	}
	opts.CommitStatus.Index = nextIndex + 1
	log.Debug("NewCommitStatus[%s, %s]: %d", repoPath, opts.SHA, opts.CommitStatus.Index)

	opts.CommitStatus.ContextHash = hashCommitStatusContext(opts.CommitStatus.Context)

	// Insert new CommitStatus
	if _, err = sess.Insert(opts.CommitStatus); err != nil {
		if err := sess.Rollback(); err != nil {
			log.Error("Insert CommitStatus: sess.Rollback: %v", err)
		}
		return fmt.Errorf("Insert CommitStatus[%s, %s]: %v", repoPath, opts.SHA, err)
	}

	return sess.Commit()
}

// SignCommitWithStatuses represents a commit with validation of signature and status state.
type SignCommitWithStatuses struct {
	Status *CommitStatus
	*SignCommit
}

// ParseCommitsWithStatus checks commits latest statuses and calculates its worst status state
func ParseCommitsWithStatus(oldCommits *list.List, repo *Repository) *list.List {
	var (
		newCommits = list.New()
		e          = oldCommits.Front()
	)

	for e != nil {
		c := e.Value.(SignCommit)
		commit := SignCommitWithStatuses{
			SignCommit: &c,
		}
		statuses, err := GetLatestCommitStatus(repo, commit.ID.String(), 0)
		if err != nil {
			log.Error("GetLatestCommitStatus: %v", err)
		} else {
			commit.Status = CalcCommitStatus(statuses)
		}

		newCommits.PushBack(commit)
		e = e.Next()
	}
	return newCommits
}

// hashCommitStatusContext hash context
func hashCommitStatusContext(context string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(context)))
}
