// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2016 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/setting"

	"github.com/Unknwon/cae/zip"
	"github.com/Unknwon/com"
	"github.com/urfave/cli"
)

// CmdDump represents the available dump sub-command.
var CmdDump = cli.Command{
	Name:  "dump",
	Usage: "Dump Gitea files and database",
	Description: `Dump compresses all related files and database into zip file.
It can be used for backup and capture Gitea server image to send to maintainer`,
	Action: runDump,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: fmt.Sprintf("gitea-dump-%d.zip", time.Now().Unix()),
			Usage: "Name of the dump file which will be created.",
		},
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Show process details",
		},
		cli.StringFlag{
			Name:  "tempdir, t",
			Value: os.TempDir(),
			Usage: "Temporary dir path",
		},
		cli.StringFlag{
			Name:  "database, d",
			Usage: "Specify the database SQL syntax",
		},
		cli.BoolFlag{
			Name:  "skip-repository, R",
			Usage: "Skip the repository dumping",
		},
	},
}

func runDump(ctx *cli.Context) error {
	setting.NewContext()
	setting.NewServices() // cannot access session settings otherwise
	models.LoadConfigs()

	err := models.SetEngine()
	if err != nil {
		return err
	}

	tmpDir := ctx.String("tempdir")
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", tmpDir)
	}
	tmpWorkDir, err := ioutil.TempDir(tmpDir, "gitea-dump-")
	if err != nil {
		log.Fatalf("Failed to create tmp work directory: %v", err)
	}
	log.Printf("Creating tmp work dir: %s", tmpWorkDir)

	// work-around #1103
	if os.Getenv("TMPDIR") == "" {
		os.Setenv("TMPDIR", tmpWorkDir)
	}

	dbDump := path.Join(tmpWorkDir, "gitea-db.sql")

	fileName := ctx.String("file")
	log.Printf("Packing dump files...")
	z, err := zip.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", fileName, err)
	}
	zip.Verbose = ctx.Bool("verbose")

	if ctx.IsSet("skip-repository") {
		log.Printf("Skip dumping local repositories")
	} else {
		log.Printf("Dumping local repositories...%s", setting.RepoRootPath)
		reposDump := path.Join(tmpWorkDir, "gitea-repo.zip")
		if err := zip.PackTo(setting.RepoRootPath, reposDump, true); err != nil {
			log.Fatalf("Failed to dump local repositories: %v", err)
		}
		if err := z.AddFile("gitea-repo.zip", reposDump); err != nil {
			log.Fatalf("Failed to include gitea-repo.zip: %v", err)
		}
	}

	targetDBType := ctx.String("database")
	if len(targetDBType) > 0 && targetDBType != models.DbCfg.Type {
		log.Printf("Dumping database %s => %s...", models.DbCfg.Type, targetDBType)
	} else {
		log.Printf("Dumping database...")
	}

	if err := models.DumpDatabase(dbDump, targetDBType); err != nil {
		log.Fatalf("Failed to dump database: %v", err)
	}

	if err := z.AddFile("gitea-db.sql", dbDump); err != nil {
		log.Fatalf("Failed to include gitea-db.sql: %v", err)
	}

	if len(setting.CustomConf) > 0 {
		log.Printf("Adding custom configuration file from %s", setting.CustomConf)
		if err := z.AddFile("app.ini", setting.CustomConf); err != nil {
			log.Fatalf("Failed to include specified app.ini: %v", err)
		}
	}

	customDir, err := os.Stat(setting.CustomPath)
	if err == nil && customDir.IsDir() {
		if err := z.AddDir("custom", setting.CustomPath); err != nil {
			log.Fatalf("Failed to include custom: %v", err)
		}
	} else {
		log.Printf("Custom dir %s doesn't exist, skipped", setting.CustomPath)
	}

	if com.IsExist(setting.AppDataPath) {
		log.Printf("Packing data directory...%s", setting.AppDataPath)

		var sessionAbsPath string
		if setting.SessionConfig.Provider == "file" {
			sessionAbsPath = setting.SessionConfig.ProviderConfig
		}
		if err := zipAddDirectoryExclude(z, "data", setting.AppDataPath, sessionAbsPath); err != nil {
			log.Fatalf("Failed to include data directory: %v", err)
		}
	}

	if err := z.AddDir("log", setting.LogRootPath); err != nil {
		log.Fatalf("Failed to include log: %v", err)
	}

	if err = z.Close(); err != nil {
		_ = os.Remove(fileName)
		log.Fatalf("Failed to save %s: %v", fileName, err)
	}

	if err := os.Chmod(fileName, 0600); err != nil {
		log.Printf("Can't change file access permissions mask to 0600: %v", err)
	}

	log.Printf("Removing tmp work dir: %s", tmpWorkDir)

	if err := os.RemoveAll(tmpWorkDir); err != nil {
		log.Fatalf("Failed to remove %s: %v", tmpWorkDir, err)
	}
	log.Printf("Finish dumping in file %s", fileName)

	return nil
}

// zipAddDirectoryExclude zips absPath to specified zipPath inside z excluding excludeAbsPath
func zipAddDirectoryExclude(zip *zip.ZipArchive, zipPath, absPath string, excludeAbsPath string) error {
	absPath, err := filepath.Abs(absPath)
	if err != nil {
		return err
	}
	dir, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	zip.AddEmptyDir(zipPath)

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, file := range files {
		currentAbsPath := path.Join(absPath, file.Name())
		currentZipPath := path.Join(zipPath, file.Name())
		if file.IsDir() {
			if currentAbsPath != excludeAbsPath {
				if err = zipAddDirectoryExclude(zip, currentZipPath, currentAbsPath, excludeAbsPath); err != nil {
					return err
				}
			}

		} else {
			if err = zip.AddFile(currentZipPath, currentAbsPath); err != nil {
				return err
			}
		}
	}
	return nil
}
