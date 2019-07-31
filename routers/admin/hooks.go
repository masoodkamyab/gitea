// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/base"
	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/modules/setting"
)

const (
	// tplAdminHooks template path for render hook settings
	tplAdminHooks base.TplName = "admin/hooks"
)

// DefaultWebhooks render admin-default webhook list page
func DefaultWebhooks(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.hooks")
	ctx.Data["PageIsAdminHooks"] = true
	ctx.Data["BaseLink"] = setting.AppSubURL + "/admin/hooks"
	ctx.Data["Description"] = ctx.Tr("admin.hooks.desc")

	ws, err := models.GetDefaultWebhooks()
	if err != nil {
		ctx.ServerError("GetWebhooksDefaults", err)
		return
	}

	ctx.Data["Webhooks"] = ws
	ctx.HTML(200, tplAdminHooks)
}

// DeleteDefaultWebhook response for delete admin-default webhook
func DeleteDefaultWebhook(ctx *context.Context) {
	if err := models.DeleteDefaultWebhook(ctx.QueryInt64("id")); err != nil {
		ctx.Flash.Error("DeleteDefaultWebhook: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.webhook_deletion_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": setting.AppSubURL + "/admin/hooks",
	})
}
