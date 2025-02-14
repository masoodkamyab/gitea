// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/setting"

	"github.com/urfave/cli"
)

// CmdConvert represents the available convert sub-command.
var CmdConvert = cli.Command{
	Name:        "convert",
	Usage:       "Convert the database",
	Description: "A command to convert an existing MySQL database from utf8 to utf8mb4",
	Action:      runConvert,
}

func runConvert(ctx *cli.Context) error {
	if err := initDB(); err != nil {
		return err
	}

	log.Trace("AppPath: %s", setting.AppPath)
	log.Trace("AppWorkPath: %s", setting.AppWorkPath)
	log.Trace("Custom path: %s", setting.CustomPath)
	log.Trace("Log path: %s", setting.LogRootPath)
	models.LoadConfigs()

	if models.DbCfg.Type != "mysql" {
		fmt.Println("This command can only be used with a MySQL database")
		return nil
	}

	if err := models.ConvertUtf8ToUtf8mb4(); err != nil {
		log.Fatal("Failed to convert database from utf8 to utf8mb4: %v", err)
		return err
	}

	fmt.Println("Converted successfully, please confirm your database's character set is now utf8mb4")

	return nil
}
