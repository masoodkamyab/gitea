// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package private

import (
	"encoding/json"
	"strings"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/git"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/repofiles"

	macaron "gopkg.in/macaron.v1"
)

// PushUpdate update public key updates
func PushUpdate(ctx *macaron.Context) {
	var opt models.PushUpdateOptions
	if err := json.NewDecoder(ctx.Req.Request.Body).Decode(&opt); err != nil {
		ctx.JSON(500, map[string]interface{}{
			"err": err.Error(),
		})
		return
	}

	branch := strings.TrimPrefix(opt.RefFullName, git.BranchPrefix)
	if len(branch) == 0 || opt.PusherID <= 0 {
		ctx.Error(404)
		log.Trace("PushUpdate: branch or secret is empty, or pusher ID is not valid")
		return
	}

	repo, err := models.GetRepositoryByOwnerAndName(opt.RepoUserName, opt.RepoName)
	if err != nil {
		ctx.JSON(500, map[string]interface{}{
			"err": err.Error(),
		})
		return
	}

	err = repofiles.PushUpdate(repo, branch, opt)
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.Error(404)
		} else {
			ctx.JSON(500, map[string]interface{}{
				"err": err.Error(),
			})
		}
		return
	}
	ctx.Status(202)
}
