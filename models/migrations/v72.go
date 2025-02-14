// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"fmt"

	"github.com/masoodkamyab/gitea/modules/util"

	"github.com/go-xorm/xorm"
)

func addReview(x *xorm.Engine) error {
	// Review see models/review.go
	type Review struct {
		ID          int64 `xorm:"pk autoincr"`
		Type        string
		ReviewerID  int64 `xorm:"index"`
		IssueID     int64 `xorm:"index"`
		Content     string
		CreatedUnix util.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix util.TimeStamp `xorm:"INDEX updated"`
	}

	if err := x.Sync2(new(Review)); err != nil {
		return fmt.Errorf("Sync2: %v", err)
	}
	return nil
}
