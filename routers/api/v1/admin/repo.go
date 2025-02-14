// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	api "github.com/masoodkamyab/gitea/modules/structs"

	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/routers/api/v1/repo"
	"github.com/masoodkamyab/gitea/routers/api/v1/user"
)

// CreateRepo api for creating a repository
func CreateRepo(ctx *context.APIContext, form api.CreateRepoOption) {
	// swagger:operation POST /admin/users/{username}/repos admin adminCreateRepo
	// ---
	// summary: Create a repository on behalf a user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: username
	//   in: path
	//   description: username of the user. This user will own the created repository
	//   type: string
	//   required: true
	// - name: repository
	//   in: body
	//   required: true
	//   schema: { "$ref": "#/definitions/CreateRepoOption" }
	// responses:
	//   "201":
	//     "$ref": "#/responses/Repository"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "422":
	//     "$ref": "#/responses/validationError"
	owner := user.GetUserByParams(ctx)
	if ctx.Written() {
		return
	}

	repo.CreateUserRepo(ctx, owner, form)
}
