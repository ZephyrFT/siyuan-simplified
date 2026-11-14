// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/88250/lute/ast"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	ignore "github.com/sabhiram/go-gitignore"

	"github.com/88250/lute/parse"
	"github.com/siyuan-note/dejavu"
	"github.com/siyuan-note/siyuan/kernel/apicontract"
	"github.com/siyuan-note/siyuan/kernel/cache"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/filesys"
	"github.com/siyuan-note/siyuan/kernel/treenode"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func writeAssetDownloadDocumentTest(t *testing.T, tree *parse.Tree) {
	t.Helper()
	if _, err := filesys.WriteTree(tree); err != nil {
		t.Fatal(err)
	}
	treenode.UpsertBlockTree(tree)
	t.Cleanup(func() { cache.RemoveTreeData(tree.ID); cache.RemoveDocIAL(tree.Path) })
}

func prepareSyncDirectoryTest(t *testing.T) {
	t.Helper()
	originalConf := Conf
	originalData, originalRepo, originalConfDir := util.DataDir, util.RepoDir, util.ConfDir
	originalHistory, originalTemp := util.HistoryDir, util.TempDir
	t.Cleanup(func() {
		Conf = originalConf
		util.DataDir, util.RepoDir, util.ConfDir = originalData, originalRepo, originalConfDir
		util.HistoryDir, util.TempDir = originalHistory, originalTemp
	})
	base := t.TempDir()
	util.DataDir, util.RepoDir, util.ConfDir = filepath.Join(base, "data"), filepath.Join(base, "repo"), filepath.Join(base, "conf")
	util.HistoryDir, util.TempDir = filepath.Join(base, "history"), filepath.Join(base, "temp")
	if err := os.MkdirAll(util.ConfDir, 0755); err != nil {
		t.Fatal(err)
	}
	Conf = NewAppConf()
	{
		confSectionValue0, confSectionValue1, confSectionValue2 := conf.NewSync(), conf.NewRepo(), conf.NewSystem()
		Conf.Sync.Store(confSectionValue0)
		Conf.Repo.Store(confSectionValue1)
		Conf.System.Store(confSectionValue2)
	}
	Conf.Sync.Load().Provider = conf.ProviderS3
	Conf.Sync.Load().S3 = &conf.S3{Bucket: "notes.backup", Region: "us-east-1", Timeout: 5, PathStyle: true,
		AccessKey: "test-key", SecretKey: "test-secret"}
	Conf.Repo.Load().Key = []byte("0123456789abcdef0123456789abcdef")
}

func prepareAssetDownloadDocumentTest(t *testing.T) (*fileOperationTestFixture, *dejavu.Repo, *dejavu.Repo, string) {
	fixture := setupFileOperationTest(t)
	full, partial, fullData := prepareAssetDownloadRepoTest(t)
	{
		confSectionValue0, confSectionValue1, confSectionValue2 := conf.NewExport(), conf.NewEditor(), conf.NewAppearance()
		Conf.Export.Store(confSectionValue0)
		Conf.Editor.Store(confSectionValue1)
		Conf.Appearance.Store(confSectionValue2)
	}
	Conf.FileTree.Store(conf.NewFileTree())
	Conf.NotebookCrypto = conf.NewNotebookCrypto()
	for range 3 {
		writeAssetDownloadDocumentTest(t, treenode.NewTree(fixture.box.ID, "/"+ast.NewNodeID()+".sy", "/Seed", "Seed"))
	}
	if err := fixture.box.SaveConf(conf.NewBoxConf()); err != nil {
		t.Fatal(err)
	}
	return fixture, full, partial, fullData
}

func mustSyncIgnoreRules(t testing.TB) ([]string, *ignore.GitIgnore) {
	t.Helper()
	lines, matcher, err := getSyncIgnoreRules()
	if err != nil {
		t.Fatal(err)
	}
	return lines, matcher
}

func mustSyncIgnoreLines(t testing.TB) []string {
	t.Helper()
	lines, _ := mustSyncIgnoreRules(t)
	return lines
}

func connectSyncCheckoutClient(t *testing.T, id string) *websocket.Conn {
	t.Helper()
	push := melody.New()
	connected := make(chan *melody.Session, 1)
	push.HandleConnect(func(session *melody.Session) { util.AddPushChan(session); connected <- session })
	push.HandleDisconnect(util.RemovePushChan)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = push.HandleRequest(w, r) }))
	connection, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+
		"/?app=checkout-test&id="+id+"&type=main", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	select {
	case session := <-connected:
		t.Cleanup(func() {
			util.RemovePushChan(session)
			_ = connection.Close()
			_ = push.Close()
			server.Close()
		})
	case <-time.After(5 * time.Second):
		t.Fatal("client did not register")
	}
	return connection
}

func readSyncCheckoutState(t *testing.T, connection *websocket.Conn) apicontract.SyncCheckoutState {
	t.Helper()
	_ = connection.SetReadDeadline(time.Now().Add(5 * time.Second))
	var event struct {
		Cmd  string
		Data apicontract.SyncCheckoutState
	}
	if err := connection.ReadJSON(&event); err != nil || event.Cmd != "syncProgress" {
		t.Fatalf("unexpected progress event: %+v %v", event, err)
	}
	return event.Data
}
