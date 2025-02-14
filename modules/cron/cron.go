// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cron

import (
	"time"

	"github.com/gogits/cron"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/modules/setting"
)

var c = cron.New()

// NewContext begins cron tasks
func NewContext() {
	var (
		entry *cron.Entry
		err   error
	)
	if setting.Cron.UpdateMirror.Enabled {
		entry, err = c.AddFunc("Update mirrors", setting.Cron.UpdateMirror.Schedule, models.MirrorUpdate)
		if err != nil {
			log.Fatal("Cron[Update mirrors]: %v", err)
		}
		if setting.Cron.UpdateMirror.RunAtStart {
			entry.Prev = time.Now()
			entry.ExecTimes++
			go models.MirrorUpdate()
		}
	}
	if setting.Cron.RepoHealthCheck.Enabled {
		entry, err = c.AddFunc("Repository health check", setting.Cron.RepoHealthCheck.Schedule, models.GitFsck)
		if err != nil {
			log.Fatal("Cron[Repository health check]: %v", err)
		}
		if setting.Cron.RepoHealthCheck.RunAtStart {
			entry.Prev = time.Now()
			entry.ExecTimes++
			go models.GitFsck()
		}
	}
	if setting.Cron.CheckRepoStats.Enabled {
		entry, err = c.AddFunc("Check repository statistics", setting.Cron.CheckRepoStats.Schedule, models.CheckRepoStats)
		if err != nil {
			log.Fatal("Cron[Check repository statistics]: %v", err)
		}
		if setting.Cron.CheckRepoStats.RunAtStart {
			entry.Prev = time.Now()
			entry.ExecTimes++
			go models.CheckRepoStats()
		}
	}
	if setting.Cron.ArchiveCleanup.Enabled {
		entry, err = c.AddFunc("Clean up old repository archives", setting.Cron.ArchiveCleanup.Schedule, models.DeleteOldRepositoryArchives)
		if err != nil {
			log.Fatal("Cron[Clean up old repository archives]: %v", err)
		}
		if setting.Cron.ArchiveCleanup.RunAtStart {
			entry.Prev = time.Now()
			entry.ExecTimes++
			go models.DeleteOldRepositoryArchives()
		}
	}
	if setting.Cron.SyncExternalUsers.Enabled {
		entry, err = c.AddFunc("Synchronize external users", setting.Cron.SyncExternalUsers.Schedule, models.SyncExternalUsers)
		if err != nil {
			log.Fatal("Cron[Synchronize external users]: %v", err)
		}
		if setting.Cron.SyncExternalUsers.RunAtStart {
			entry.Prev = time.Now()
			entry.ExecTimes++
			go models.SyncExternalUsers()
		}
	}
	if setting.Cron.DeletedBranchesCleanup.Enabled {
		entry, err = c.AddFunc("Remove old deleted branches", setting.Cron.DeletedBranchesCleanup.Schedule, models.RemoveOldDeletedBranches)
		if err != nil {
			log.Fatal("Cron[Remove old deleted branches]: %v", err)
		}
		if setting.Cron.DeletedBranchesCleanup.RunAtStart {
			entry.Prev = time.Now()
			entry.ExecTimes++
			go models.RemoveOldDeletedBranches()
		}
	}
	c.Start()
}

// ListTasks returns all running cron tasks.
func ListTasks() []*cron.Entry {
	return c.Entries()
}
