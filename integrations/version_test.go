// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	"github.com/masoodkamyab/gitea/modules/setting"
	"github.com/masoodkamyab/gitea/modules/structs"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	prepareTestEnv(t)

	setting.AppVer = "test-version-1"
	req := NewRequest(t, "GET", "/api/v1/version")
	resp := MakeRequest(t, req, http.StatusOK)

	var version structs.ServerVersion
	DecodeJSON(t, resp, &version)
	assert.Equal(t, setting.AppVer, string(version.Version))
}
