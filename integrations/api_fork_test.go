// Copyright 2017 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	api "github.com/masoodkamyab/gitea/modules/structs"
)

func TestCreateForkNoLogin(t *testing.T) {
	prepareTestEnv(t)
	req := NewRequestWithJSON(t, "POST", "/api/v1/repos/user2/repo1/forks", &api.CreateForkOption{})
	MakeRequest(t, req, http.StatusUnauthorized)
}
