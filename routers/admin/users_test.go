// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"testing"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/auth"
	"github.com/masoodkamyab/gitea/modules/test"
	"github.com/stretchr/testify/assert"
)

func TestNewUserPost_MustChangePassword(t *testing.T) {

	models.PrepareTestEnv(t)
	ctx := test.MockContext(t, "admin/users/new")

	u := models.AssertExistsAndLoadBean(t, &models.User{
		IsAdmin: true,
		ID:      2,
	}).(*models.User)

	ctx.User = u

	username := "gitea"
	email := "gitea@gitea.io"

	form := auth.AdminCreateUserForm{
		LoginType:          "local",
		LoginName:          "local",
		UserName:           username,
		Email:              email,
		Password:           "xxxxxxxx",
		SendNotify:         false,
		MustChangePassword: true,
	}

	NewUserPost(ctx, form)

	assert.NotEmpty(t, ctx.Flash.SuccessMsg)

	u, err := models.GetUserByName(username)

	assert.NoError(t, err)
	assert.Equal(t, username, u.Name)
	assert.Equal(t, email, u.Email)
	assert.True(t, u.MustChangePassword)
}

func TestNewUserPost_MustChangePasswordFalse(t *testing.T) {

	models.PrepareTestEnv(t)
	ctx := test.MockContext(t, "admin/users/new")

	u := models.AssertExistsAndLoadBean(t, &models.User{
		IsAdmin: true,
		ID:      2,
	}).(*models.User)

	ctx.User = u

	username := "gitea"
	email := "gitea@gitea.io"

	form := auth.AdminCreateUserForm{
		LoginType:          "local",
		LoginName:          "local",
		UserName:           username,
		Email:              email,
		Password:           "xxxxxxxx",
		SendNotify:         false,
		MustChangePassword: false,
	}

	NewUserPost(ctx, form)

	assert.NotEmpty(t, ctx.Flash.SuccessMsg)

	u, err := models.GetUserByName(username)

	assert.NoError(t, err)
	assert.Equal(t, username, u.Name)
	assert.Equal(t, email, u.Email)
	assert.False(t, u.MustChangePassword)
}
