// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	"github.com/masoodkamyab/gitea/models"
)

func TestAPIReposGitCommits(t *testing.T) {
	prepareTestEnv(t)
	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session)

	for _, ref := range [...]string{
		"commits/master", // Branch
		"commits/v1.1",   // Tag
	} {
		req := NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/git/%s?token="+token, user.Name, ref)
		session.MakeRequest(t, req, http.StatusOK)
	}

	// Test getting non-existent refs
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/git/commits/unknown?token="+token, user.Name)
	session.MakeRequest(t, req, http.StatusNotFound)
}
