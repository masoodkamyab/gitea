// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	"github.com/masoodkamyab/gitea/modules/setting"
)

func TestSignup(t *testing.T) {
	prepareTestEnv(t)

	setting.Service.EnableCaptcha = false

	req := NewRequestWithValues(t, "POST", "/user/sign_up", map[string]string{
		"user_name": "exampleUser",
		"email":     "exampleUser@example.com",
		"password":  "examplePassword",
		"retype":    "examplePassword",
	})
	MakeRequest(t, req, http.StatusFound)

	// should be able to view new user's page
	req = NewRequest(t, "GET", "/exampleUser")
	MakeRequest(t, req, http.StatusOK)
}
