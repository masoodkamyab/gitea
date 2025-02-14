// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"time"

	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/setting"

	"github.com/go-xorm/xorm"
)

func migrateProtectedBranchStruct(x *xorm.Engine) error {
	type ProtectedBranch struct {
		ID          int64  `xorm:"pk autoincr"`
		RepoID      int64  `xorm:"UNIQUE(s)"`
		BranchName  string `xorm:"UNIQUE(s)"`
		CanPush     bool
		Created     time.Time `xorm:"-"`
		CreatedUnix int64
		Updated     time.Time `xorm:"-"`
		UpdatedUnix int64
	}

	var pbs []ProtectedBranch
	err := x.Find(&pbs)
	if err != nil {
		return err
	}

	for _, pb := range pbs {
		if pb.CanPush {
			if _, err = x.ID(pb.ID).Delete(new(ProtectedBranch)); err != nil {
				return err
			}
		}
	}

	switch {
	case setting.UseSQLite3:
		log.Warn("Unable to drop columns in SQLite")
	case setting.UseMySQL, setting.UsePostgreSQL, setting.UseMSSQL, setting.UseTiDB:
		if _, err := x.Exec("ALTER TABLE protected_branch DROP COLUMN can_push"); err != nil {
			// Ignoring this error in case we run this migration second time (after migration reordering)
			log.Warn("DROP COLUMN can_push (skipping): %v", err)
		}
	default:
		log.Fatal("Unrecognized DB")
	}

	return nil
}
