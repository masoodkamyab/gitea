// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/modules/git"
	api "github.com/masoodkamyab/gitea/modules/structs"
	"github.com/masoodkamyab/gitea/routers/api/v1/convert"
	"github.com/masoodkamyab/gitea/routers/api/v1/utils"
)

// ListHooks list all hooks of a repository
func ListHooks(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/hooks repository repoListHooks
	// ---
	// summary: List the hooks in a repository
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/HookList"
	hooks, err := models.GetWebhooksByRepoID(ctx.Repo.Repository.ID)
	if err != nil {
		ctx.Error(500, "GetWebhooksByRepoID", err)
		return
	}

	apiHooks := make([]*api.Hook, len(hooks))
	for i := range hooks {
		apiHooks[i] = convert.ToHook(ctx.Repo.RepoLink, hooks[i])
	}
	ctx.JSON(200, &apiHooks)
}

// GetHook get a repo's hook by id
func GetHook(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/hooks/{id} repository repoGetHook
	// ---
	// summary: Get a hook
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the hook to get
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Hook"
	repo := ctx.Repo
	hookID := ctx.ParamsInt64(":id")
	hook, err := utils.GetRepoHook(ctx, repo.Repository.ID, hookID)
	if err != nil {
		return
	}
	ctx.JSON(200, convert.ToHook(repo.RepoLink, hook))
}

// TestHook tests a hook
func TestHook(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/hooks/{id}/tests repository repoTestHook
	// ---
	// summary: Test a push webhook
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the hook to test
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	if ctx.Repo.Commit == nil {
		// if repo does not have any commits, then don't send a webhook
		ctx.Status(204)
		return
	}

	hookID := ctx.ParamsInt64(":id")
	hook, err := utils.GetRepoHook(ctx, ctx.Repo.Repository.ID, hookID)
	if err != nil {
		return
	}

	if err := models.PrepareWebhook(hook, ctx.Repo.Repository, models.HookEventPush, &api.PushPayload{
		Ref:    git.BranchPrefix + ctx.Repo.Repository.DefaultBranch,
		Before: ctx.Repo.Commit.ID.String(),
		After:  ctx.Repo.Commit.ID.String(),
		Commits: []*api.PayloadCommit{
			convert.ToCommit(ctx.Repo.Repository, ctx.Repo.Commit),
		},
		Repo:   ctx.Repo.Repository.APIFormat(models.AccessModeNone),
		Pusher: convert.ToUser(ctx.User, ctx.IsSigned, false),
		Sender: convert.ToUser(ctx.User, ctx.IsSigned, false),
	}); err != nil {
		ctx.Error(500, "PrepareWebhook: ", err)
		return
	}
	go models.HookQueue.Add(ctx.Repo.Repository.ID)
	ctx.Status(204)
}

// CreateHook create a hook for a repository
func CreateHook(ctx *context.APIContext, form api.CreateHookOption) {
	// swagger:operation POST /repos/{owner}/{repo}/hooks repository repoCreateHook
	// ---
	// summary: Create a hook
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateHookOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Hook"
	if !utils.CheckCreateHookOption(ctx, &form) {
		return
	}
	utils.AddRepoHook(ctx, &form)
}

// EditHook modify a hook of a repository
func EditHook(ctx *context.APIContext, form api.EditHookOption) {
	// swagger:operation PATCH /repos/{owner}/{repo}/hooks/{id} repository repoEditHook
	// ---
	// summary: Edit a hook in a repository
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: index of the hook
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/EditHookOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Hook"
	hookID := ctx.ParamsInt64(":id")
	utils.EditRepoHook(ctx, &form, hookID)
}

// DeleteHook delete a hook of a repository
func DeleteHook(ctx *context.APIContext) {
	// swagger:operation DELETE /repos/{owner}/{repo}/hooks/{id} repository repoDeleteHook
	// ---
	// summary: Delete a hook in a repository
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the hook to delete
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "404":
	//     "$ref": "#/responses/notFound"
	if err := models.DeleteWebhookByRepoID(ctx.Repo.Repository.ID, ctx.ParamsInt64(":id")); err != nil {
		if models.IsErrWebhookNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(500, "DeleteWebhookByRepoID", err)
		}
		return
	}
	ctx.Status(204)
}
