// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"github.com/masoodkamyab/gitea/modules/context"

	api "github.com/masoodkamyab/gitea/modules/structs"
	"github.com/masoodkamyab/gitea/routers/api/v1/convert"
)

// ListStargazers list a repository's stargazers
func ListStargazers(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/stargazers repository repoListStargazers
	// ---
	// summary: List a repo's stargazers
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
	//     "$ref": "#/responses/UserList"
	stargazers, err := ctx.Repo.Repository.GetStargazers(-1)
	if err != nil {
		ctx.Error(500, "GetStargazers", err)
		return
	}
	users := make([]*api.User, len(stargazers))
	for i, stargazer := range stargazers {
		users[i] = convert.ToUser(stargazer, ctx.IsSigned, ctx.User != nil && ctx.User.IsAdmin)
	}
	ctx.JSON(200, users)
}
