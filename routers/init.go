// Copyright 2016 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"strings"
	"time"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/models/migrations"
	"github.com/masoodkamyab/gitea/modules/cache"
	"github.com/masoodkamyab/gitea/modules/cron"
	"github.com/masoodkamyab/gitea/modules/git"
	"github.com/masoodkamyab/gitea/modules/highlight"
	issue_indexer "github.com/masoodkamyab/gitea/modules/indexer/issues"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/mailer"
	"github.com/masoodkamyab/gitea/modules/markup"
	"github.com/masoodkamyab/gitea/modules/markup/external"
	"github.com/masoodkamyab/gitea/modules/setting"
	"github.com/masoodkamyab/gitea/modules/ssh"

	macaron "gopkg.in/macaron.v1"
)

func checkRunMode() {
	switch setting.Cfg.Section("").Key("RUN_MODE").String() {
	case "prod":
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
		setting.ProdMode = true
	default:
		git.Debug = true
	}
	log.Info("Run Mode: %s", strings.Title(macaron.Env))
}

// NewServices init new services
func NewServices() {
	setting.NewServices()
	mailer.NewContext()
	_ = cache.NewContext()
}

// In case of problems connecting to DB, retry connection. Eg, PGSQL in Docker Container on Synology
func initDBEngine() (err error) {
	log.Info("Beginning ORM engine initialization.")
	for i := 0; i < setting.DBConnectRetries; i++ {
		log.Info("ORM engine initialization attempt #%d/%d...", i+1, setting.DBConnectRetries)
		if err = models.NewEngine(migrations.Migrate); err == nil {
			break
		} else if i == setting.DBConnectRetries-1 {
			return err
		}
		log.Debug("ORM engine initialization attempt #%d/%d failed. Error: %v", i+1, setting.DBConnectRetries, err)
		log.Info("Backing off for %d seconds", int64(setting.DBConnectBackoff/time.Second))
		time.Sleep(setting.DBConnectBackoff)
	}
	models.HasEngine = true
	return nil
}

// GlobalInit is for global configuration reload-able.
func GlobalInit() {
	setting.NewContext()
	if err := git.Init(); err != nil {
		log.Fatal("Git module init failed: %v", err)
	}
	setting.CheckLFSVersion()
	log.Trace("AppPath: %s", setting.AppPath)
	log.Trace("AppWorkPath: %s", setting.AppWorkPath)
	log.Trace("Custom path: %s", setting.CustomPath)
	log.Trace("Log path: %s", setting.LogRootPath)
	models.LoadConfigs()
	NewServices()

	if setting.InstallLock {
		highlight.NewContext()
		external.RegisterParsers()
		markup.Init()
		if err := initDBEngine(); err == nil {
			log.Info("ORM engine initialization successful!")
		} else {
			log.Fatal("ORM engine initialization failed: %v", err)
		}

		if err := models.InitOAuth2(); err != nil {
			log.Fatal("Failed to initialize OAuth2 support: %v", err)
		}

		models.NewRepoContext()

		// Booting long running goroutines.
		cron.NewContext()
		if err := issue_indexer.InitIssueIndexer(false); err != nil {
			log.Fatal("Failed to initialize issue indexer: %v", err)
		}
		models.InitRepoIndexer()
		models.InitSyncMirrors()
		models.InitDeliverHooks()
		models.InitTestPullRequests()
	}
	if models.EnableSQLite3 {
		log.Info("SQLite3 Supported")
	}
	if models.EnableTiDB {
		log.Info("TiDB Supported")
	}
	checkRunMode()

	if setting.InstallLock && setting.SSH.StartBuiltinServer {
		ssh.Listen(setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs)
		log.Info("SSH server started on %s:%d. Cipher list (%v), key exchange algorithms (%v), MACs (%v)", setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs)
	}
}
