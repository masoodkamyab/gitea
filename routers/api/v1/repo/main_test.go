// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"path/filepath"
	"testing"

	"github.com/masoodkamyab/gitea/models"
)

func TestMain(m *testing.M) {
	models.MainTest(m, filepath.Join("..", "..", "..", ".."))
}
