// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package issues

import (
	"fmt"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/setting"
	"github.com/masoodkamyab/gitea/modules/util"
)

// IndexerData data stored in the issue indexer
type IndexerData struct {
	ID       int64
	RepoID   int64
	Title    string
	Content  string
	Comments []string
	IsDelete bool
	IDs      []int64
}

// Match represents on search result
type Match struct {
	ID     int64   `json:"id"`
	RepoID int64   `json:"repo_id"`
	Score  float64 `json:"score"`
}

// SearchResult represents search results
type SearchResult struct {
	Total int64
	Hits  []Match
}

// Indexer defines an inteface to indexer issues contents
type Indexer interface {
	Init() (bool, error)
	Index(issue []*IndexerData) error
	Delete(ids ...int64) error
	Search(kw string, repoID int64, limit, start int) (*SearchResult, error)
}

var (
	// issueIndexerQueue queue of issue ids to be updated
	issueIndexerQueue Queue
	issueIndexer      Indexer
)

// InitIssueIndexer initialize issue indexer, syncReindex is true then reindex until
// all issue index done.
func InitIssueIndexer(syncReindex bool) error {
	var populate bool
	var dummyQueue bool
	switch setting.Indexer.IssueType {
	case "bleve":
		issueIndexer = NewBleveIndexer(setting.Indexer.IssuePath)
		exist, err := issueIndexer.Init()
		if err != nil {
			return err
		}
		populate = !exist
	case "db":
		issueIndexer = &DBIndexer{}
		dummyQueue = true
	default:
		return fmt.Errorf("unknow issue indexer type: %s", setting.Indexer.IssueType)
	}

	if dummyQueue {
		issueIndexerQueue = &DummyQueue{}
		return nil
	}

	var err error
	switch setting.Indexer.IssueQueueType {
	case setting.LevelQueueType:
		issueIndexerQueue, err = NewLevelQueue(
			issueIndexer,
			setting.Indexer.IssueQueueDir,
			setting.Indexer.IssueQueueBatchNumber)
		if err != nil {
			return err
		}
	case setting.ChannelQueueType:
		issueIndexerQueue = NewChannelQueue(issueIndexer, setting.Indexer.IssueQueueBatchNumber)
	case setting.RedisQueueType:
		addrs, pass, idx, err := parseConnStr(setting.Indexer.IssueQueueConnStr)
		if err != nil {
			return err
		}
		issueIndexerQueue, err = NewRedisQueue(addrs, pass, idx, issueIndexer, setting.Indexer.IssueQueueBatchNumber)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported indexer queue type: %v", setting.Indexer.IssueQueueType)
	}

	go func() {
		err = issueIndexerQueue.Run()
		if err != nil {
			log.Error("issueIndexerQueue.Run: %v", err)
		}
	}()

	if populate {
		if syncReindex {
			populateIssueIndexer()
		} else {
			go populateIssueIndexer()
		}
	}

	return nil
}

// populateIssueIndexer populate the issue indexer with issue data
func populateIssueIndexer() {
	for page := 1; ; page++ {
		repos, _, err := models.SearchRepositoryByName(&models.SearchRepoOptions{
			Page:        page,
			PageSize:    models.RepositoryListDefaultPageSize,
			OrderBy:     models.SearchOrderByID,
			Private:     true,
			Collaborate: util.OptionalBoolFalse,
		})
		if err != nil {
			log.Error("SearchRepositoryByName: %v", err)
			continue
		}
		if len(repos) == 0 {
			return
		}

		for _, repo := range repos {
			is, err := models.Issues(&models.IssuesOptions{
				RepoIDs:  []int64{repo.ID},
				IsClosed: util.OptionalBoolNone,
				IsPull:   util.OptionalBoolNone,
			})
			if err != nil {
				log.Error("Issues: %v", err)
				continue
			}
			if err = models.IssueList(is).LoadDiscussComments(); err != nil {
				log.Error("LoadComments: %v", err)
				continue
			}
			for _, issue := range is {
				UpdateIssueIndexer(issue)
			}
		}
	}
}

// UpdateIssueIndexer add/update an issue to the issue indexer
func UpdateIssueIndexer(issue *models.Issue) {
	var comments []string
	for _, comment := range issue.Comments {
		if comment.Type == models.CommentTypeComment {
			comments = append(comments, comment.Content)
		}
	}
	_ = issueIndexerQueue.Push(&IndexerData{
		ID:       issue.ID,
		RepoID:   issue.RepoID,
		Title:    issue.Title,
		Content:  issue.Content,
		Comments: comments,
	})
}

// DeleteRepoIssueIndexer deletes repo's all issues indexes
func DeleteRepoIssueIndexer(repo *models.Repository) {
	var ids []int64
	ids, err := models.GetIssueIDsByRepoID(repo.ID)
	if err != nil {
		log.Error("getIssueIDsByRepoID failed: %v", err)
		return
	}

	if len(ids) == 0 {
		return
	}

	_ = issueIndexerQueue.Push(&IndexerData{
		IDs:      ids,
		IsDelete: true,
	})
}

// SearchIssuesByKeyword search issue ids by keywords and repo id
func SearchIssuesByKeyword(repoID int64, keyword string) ([]int64, error) {
	var issueIDs []int64
	res, err := issueIndexer.Search(keyword, repoID, 1000, 0)
	if err != nil {
		return nil, err
	}
	for _, r := range res.Hits {
		issueIDs = append(issueIDs, r.ID)
	}
	return issueIDs, nil
}
