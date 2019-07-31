// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mail

import (
	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/notification/base"
)

type mailNotifier struct {
	base.NullNotifier
}

var (
	_ base.Notifier = &mailNotifier{}
)

// NewNotifier create a new mailNotifier notifier
func NewNotifier() base.Notifier {
	return &mailNotifier{}
}

func (m *mailNotifier) NotifyCreateIssueComment(doer *models.User, repo *models.Repository,
	issue *models.Issue, comment *models.Comment) {
	var act models.ActionType
	if comment.Type == models.CommentTypeClose {
		act = models.ActionCloseIssue
	} else if comment.Type == models.CommentTypeReopen {
		act = models.ActionReopenIssue
	} else if comment.Type == models.CommentTypeComment {
		act = models.ActionCommentIssue
	} else if comment.Type == models.CommentTypeCode {
		act = models.ActionCommentIssue
	}

	if err := comment.MailParticipants(act, issue); err != nil {
		log.Error("MailParticipants: %v", err)
	}
}

func (m *mailNotifier) NotifyNewIssue(issue *models.Issue) {
	if err := issue.MailParticipants(issue.Poster, models.ActionCreateIssue); err != nil {
		log.Error("MailParticipants: %v", err)
	}
}

func (m *mailNotifier) NotifyIssueChangeStatus(doer *models.User, issue *models.Issue, isClosed bool) {
	var actionType models.ActionType
	if issue.IsPull {
		if isClosed {
			actionType = models.ActionClosePullRequest
		} else {
			actionType = models.ActionReopenPullRequest
		}
	} else {
		if isClosed {
			actionType = models.ActionCloseIssue
		} else {
			actionType = models.ActionReopenIssue
		}
	}

	if err := issue.MailParticipants(doer, actionType); err != nil {
		log.Error("MailParticipants: %v", err)
	}
}

func (m *mailNotifier) NotifyNewPullRequest(pr *models.PullRequest) {
	if err := pr.Issue.MailParticipants(pr.Issue.Poster, models.ActionCreatePullRequest); err != nil {
		log.Error("MailParticipants: %v", err)
	}
}

func (m *mailNotifier) NotifyPullRequestReview(pr *models.PullRequest, r *models.Review, comment *models.Comment) {
	var act models.ActionType
	if comment.Type == models.CommentTypeClose {
		act = models.ActionCloseIssue
	} else if comment.Type == models.CommentTypeReopen {
		act = models.ActionReopenIssue
	} else if comment.Type == models.CommentTypeComment {
		act = models.ActionCommentIssue
	}
	if err := comment.MailParticipants(act, pr.Issue); err != nil {
		log.Error("MailParticipants: %v", err)
	}
}
