// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"path"
	"strings"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/base"
	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/modules/git"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/setting"
	"github.com/masoodkamyab/gitea/modules/templates"
)

const (
	tplCommits    base.TplName = "repo/commits"
	tplGraph      base.TplName = "repo/graph"
	tplCommitPage base.TplName = "repo/commit_page"
)

// RefCommits render commits page
func RefCommits(ctx *context.Context) {
	switch {
	case len(ctx.Repo.TreePath) == 0:
		Commits(ctx)
	case ctx.Repo.TreePath == "search":
		SearchCommits(ctx)
	default:
		FileHistory(ctx)
	}
}

// Commits render branch's commits
func Commits(ctx *context.Context) {
	ctx.Data["PageIsCommits"] = true
	if ctx.Repo.Commit == nil {
		ctx.NotFound("Commit not found", nil)
		return
	}
	ctx.Data["PageIsViewCode"] = true

	commitsCount, err := ctx.Repo.GetCommitsCount()
	if err != nil {
		ctx.ServerError("GetCommitsCount", err)
		return
	}

	page := ctx.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	// Both `git log branchName` and `git log commitId` work.
	commits, err := ctx.Repo.Commit.CommitsByRange(page)
	if err != nil {
		ctx.ServerError("CommitsByRange", err)
		return
	}
	commits = models.ValidateCommitsWithEmails(commits)
	commits = models.ParseCommitsWithSignature(commits)
	commits = models.ParseCommitsWithStatus(commits, ctx.Repo.Repository)
	ctx.Data["Commits"] = commits

	ctx.Data["Username"] = ctx.Repo.Owner.Name
	ctx.Data["Reponame"] = ctx.Repo.Repository.Name
	ctx.Data["CommitCount"] = commitsCount
	ctx.Data["Branch"] = ctx.Repo.BranchName

	pager := context.NewPagination(int(commitsCount), git.CommitsRangeSize, page, 5)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.HTML(200, tplCommits)
}

// Graph render commit graph - show commits from all branches.
func Graph(ctx *context.Context) {
	ctx.Data["PageIsCommits"] = true
	ctx.Data["PageIsViewCode"] = true

	commitsCount, err := ctx.Repo.GetCommitsCount()
	if err != nil {
		ctx.ServerError("GetCommitsCount", err)
		return
	}

	graph, err := models.GetCommitGraph(ctx.Repo.GitRepo)
	if err != nil {
		ctx.ServerError("GetCommitGraph", err)
		return
	}

	ctx.Data["Graph"] = graph
	ctx.Data["Username"] = ctx.Repo.Owner.Name
	ctx.Data["Reponame"] = ctx.Repo.Repository.Name
	ctx.Data["CommitCount"] = commitsCount
	ctx.Data["Branch"] = ctx.Repo.BranchName
	ctx.Data["RequireGitGraph"] = true
	ctx.HTML(200, tplGraph)

}

// SearchCommits render commits filtered by keyword
func SearchCommits(ctx *context.Context) {
	ctx.Data["PageIsCommits"] = true
	ctx.Data["PageIsViewCode"] = true

	query := strings.Trim(ctx.Query("q"), " ")
	if len(query) == 0 {
		ctx.Redirect(ctx.Repo.RepoLink + "/commits/" + ctx.Repo.BranchNameSubURL())
		return
	}

	all := ctx.QueryBool("all")
	opts := git.NewSearchCommitsOptions(query, all)
	commits, err := ctx.Repo.Commit.SearchCommits(opts)
	if err != nil {
		ctx.ServerError("SearchCommits", err)
		return
	}
	commits = models.ValidateCommitsWithEmails(commits)
	commits = models.ParseCommitsWithSignature(commits)
	commits = models.ParseCommitsWithStatus(commits, ctx.Repo.Repository)
	ctx.Data["Commits"] = commits

	ctx.Data["Keyword"] = query
	if all {
		ctx.Data["All"] = "checked"
	}
	ctx.Data["Username"] = ctx.Repo.Owner.Name
	ctx.Data["Reponame"] = ctx.Repo.Repository.Name
	ctx.Data["CommitCount"] = commits.Len()
	ctx.Data["Branch"] = ctx.Repo.BranchName
	ctx.HTML(200, tplCommits)
}

// FileHistory show a file's reversions
func FileHistory(ctx *context.Context) {
	ctx.Data["IsRepoToolbarCommits"] = true

	fileName := ctx.Repo.TreePath
	if len(fileName) == 0 {
		Commits(ctx)
		return
	}

	branchName := ctx.Repo.BranchName
	commitsCount, err := ctx.Repo.GitRepo.FileCommitsCount(branchName, fileName)
	if err != nil {
		ctx.ServerError("FileCommitsCount", err)
		return
	} else if commitsCount == 0 {
		ctx.NotFound("FileCommitsCount", nil)
		return
	}

	page := ctx.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	commits, err := ctx.Repo.GitRepo.CommitsByFileAndRange(branchName, fileName, page)
	if err != nil {
		ctx.ServerError("CommitsByFileAndRange", err)
		return
	}
	commits = models.ValidateCommitsWithEmails(commits)
	commits = models.ParseCommitsWithSignature(commits)
	commits = models.ParseCommitsWithStatus(commits, ctx.Repo.Repository)
	ctx.Data["Commits"] = commits

	ctx.Data["Username"] = ctx.Repo.Owner.Name
	ctx.Data["Reponame"] = ctx.Repo.Repository.Name
	ctx.Data["FileName"] = fileName
	ctx.Data["CommitCount"] = commitsCount
	ctx.Data["Branch"] = branchName

	pager := context.NewPagination(int(commitsCount), git.CommitsRangeSize, page, 5)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.HTML(200, tplCommits)
}

// Diff show different from current commit to previous commit
func Diff(ctx *context.Context) {
	ctx.Data["PageIsDiff"] = true
	ctx.Data["RequireHighlightJS"] = true

	userName := ctx.Repo.Owner.Name
	repoName := ctx.Repo.Repository.Name
	commitID := ctx.Params(":sha")

	commit, err := ctx.Repo.GitRepo.GetCommit(commitID)
	if err != nil {
		if git.IsErrNotExist(err) {
			ctx.NotFound("Repo.GitRepo.GetCommit", err)
		} else {
			ctx.ServerError("Repo.GitRepo.GetCommit", err)
		}
		return
	}
	if len(commitID) != 40 {
		commitID = commit.ID.String()
	}

	statuses, err := models.GetLatestCommitStatus(ctx.Repo.Repository, commitID, 0)
	if err != nil {
		log.Error("GetLatestCommitStatus: %v", err)
	}

	ctx.Data["CommitStatus"] = models.CalcCommitStatus(statuses)

	diff, err := models.GetDiffCommit(models.RepoPath(userName, repoName),
		commitID, setting.Git.MaxGitDiffLines,
		setting.Git.MaxGitDiffLineCharacters, setting.Git.MaxGitDiffFiles)
	if err != nil {
		ctx.NotFound("GetDiffCommit", err)
		return
	}

	parents := make([]string, commit.ParentCount())
	for i := 0; i < commit.ParentCount(); i++ {
		sha, err := commit.ParentID(i)
		parents[i] = sha.String()
		if err != nil {
			ctx.NotFound("repo.Diff", err)
			return
		}
	}

	ctx.Data["CommitID"] = commitID
	ctx.Data["Username"] = userName
	ctx.Data["Reponame"] = repoName
	ctx.Data["IsImageFile"] = commit.IsImageFile
	ctx.Data["Title"] = commit.Summary() + " · " + base.ShortSha(commitID)
	ctx.Data["Commit"] = commit
	ctx.Data["Verification"] = models.ParseCommitWithSignature(commit)
	ctx.Data["Author"] = models.ValidateCommitWithEmail(commit)
	ctx.Data["Diff"] = diff
	ctx.Data["Parents"] = parents
	ctx.Data["DiffNotAvailable"] = diff.NumFiles() == 0
	ctx.Data["SourcePath"] = setting.AppSubURL + "/" + path.Join(userName, repoName, "src", "commit", commitID)

	note := &git.Note{}
	err = git.GetNote(ctx.Repo.GitRepo, commitID, note)
	if err == nil {
		ctx.Data["Note"] = string(templates.ToUTF8WithFallback(note.Message))
		ctx.Data["NoteCommit"] = note.Commit
		ctx.Data["NoteAuthor"] = models.ValidateCommitWithEmail(note.Commit)
	}

	if commit.ParentCount() > 0 {
		ctx.Data["BeforeSourcePath"] = setting.AppSubURL + "/" + path.Join(userName, repoName, "src", "commit", parents[0])
	}
	ctx.Data["RawPath"] = setting.AppSubURL + "/" + path.Join(userName, repoName, "raw", "commit", commitID)
	ctx.Data["BranchName"], err = commit.GetBranchName()
	if err != nil {
		ctx.ServerError("commit.GetBranchName", err)
	}
	ctx.HTML(200, tplCommitPage)
}

// RawDiff dumps diff results of repository in given commit ID to io.Writer
func RawDiff(ctx *context.Context) {
	if err := models.GetRawDiff(
		models.RepoPath(ctx.Repo.Owner.Name, ctx.Repo.Repository.Name),
		ctx.Params(":sha"),
		models.RawDiffType(ctx.Params(":ext")),
		ctx.Resp,
	); err != nil {
		ctx.ServerError("GetRawDiff", err)
		return
	}
}
