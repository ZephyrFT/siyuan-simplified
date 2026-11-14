// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siyuan-note/dejavu"
	"github.com/siyuan-note/dejavu/cloud"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func prepareAssetDownloadRepoTest(t *testing.T) (*dejavu.Repo, *dejavu.Repo, string) {
	t.Helper()
	originalConf := Conf
	originalWorkspace, originalData, originalRepo := util.WorkspaceDir, util.DataDir, util.RepoDir
	originalHistory, originalTemp, originalConfDir := util.HistoryDir, util.TempDir, util.ConfDir
	originalStatusBar := util.StatusBarCfg.Load()
	t.Cleanup(func() {
		util.StatusBarCfg.Store(originalStatusBar)
		Conf = originalConf
		util.WorkspaceDir, util.DataDir, util.RepoDir = originalWorkspace, originalData, originalRepo
		util.HistoryDir, util.TempDir, util.ConfDir = originalHistory, originalTemp, originalConfDir
	})
	base := t.TempDir()
	remote, fullDir := filepath.Join(base, "cloud"), filepath.Join(base, "full")
	key := []byte("0123456789abcdef0123456789abcdef")
	fullData := filepath.Join(fullDir, "data")
	backend := cloud.NewLocal(&cloud.BaseCloud{Conf: &cloud.Conf{Dir: "main", RepoPath: filepath.Join(fullDir, "repo"),
		AvailableSize: 1 << 30, Local: &cloud.ConfLocal{Endpoint: remote}}})
	full, err := dejavu.NewRepo(fullData, filepath.Join(fullDir, "repo"), filepath.Join(fullDir, "history"),
		filepath.Join(fullDir, "temp"), "full", "full", "windows", key, nil, backend)
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"seed.txt": "seed", "assets/file.bin": "version one"} {
		p := filepath.Join(fullData, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	Conf = NewAppConf()
	util.StatusBarCfg.Store(&util.StatusBar{MsgDataSyncDisabled: true})
	{
		confSectionValue0, confSectionValue1, confSectionValue2, confSectionValue3 := conf.NewSync(), conf.NewRepo(), conf.NewSystem(), conf.NewSearch()
		Conf.Sync.Store(confSectionValue0)
		Conf.Repo.Store(confSectionValue1)
		Conf.System.Store(confSectionValue2)
		Conf.Search.Store(confSectionValue3)
	}
	Conf.Sync.Load().Local = &conf.Local{Endpoint: remote}
	Conf.SetUser(&conf.User{UserId: "asset-test", UserSiYuanOneTimePayStatus: 1})
	Conf.Sync.Load().Provider, Conf.Sync.Load().CloudName, Conf.Sync.Load().Enabled = conf.ProviderLocal, "main", true
	Conf.Sync.Load().Local.Endpoint, Conf.Sync.Load().AssetDownloadMode = remote, 1
	Conf.Repo.Load().Key = key
	util.WorkspaceDir = filepath.Join(base, "partial")
	util.DataDir, util.RepoDir = filepath.Join(util.WorkspaceDir, "data"), filepath.Join(util.WorkspaceDir, "repo")
	util.HistoryDir, util.TempDir = filepath.Join(util.WorkspaceDir, "history"), filepath.Join(util.WorkspaceDir, "temp")
	util.ConfDir = filepath.Join(util.WorkspaceDir, "conf")
	if _, err = full.Index("first", true, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err = full.Sync(nil); err != nil {
		t.Fatal(err)
	}
	partial, err := newRepository()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(util.DataDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(util.DataDir, "seed.txt"), []byte("seed"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = partial.Index("partial", true, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err = partial.Sync(nil); err != nil {
		t.Fatal(err)
	}
	id, _, err := partial.AssetDownloadChanges()
	if err != nil {
		t.Fatal(err)
	}
	if err = partial.AcknowledgeAssetDownloadChanges(id); err != nil {
		t.Fatal(err)
	}
	return full, partial, fullData
}
