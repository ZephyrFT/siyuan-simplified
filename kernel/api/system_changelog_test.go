// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/siyuan-note/siyuan/kernel/apicontract"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func changelogContractHandler(version string) gin.HandlerFunc {
	return contractHandler(apicontract.SystemGetChangelog, func(c *gin.Context, request apicontract.SystemChangelogRequest) apicontract.Response[apicontract.SystemChangelogData] {
		return loadChangelog(c.Request.Context(), request.Force, version)
	})
}

func TestAPIContractStableChangelogStaysLocal(t *testing.T) {
	oldDir, oldConf := util.WorkingDir, model.Conf
	t.Cleanup(func() {
		util.WorkingDir, model.Conf = oldDir, oldConf
	})
	util.WorkingDir, model.Conf = t.TempDir(), model.NewAppConf()
	model.Conf.Lang = "en"
	directory := filepath.Join(util.WorkingDir, "changelogs", "v3.8.9")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "v3.8.9.en.md"), []byte("## Local stable changes"), 0600); err != nil {
		t.Fatal(err)
	}
	recorder := systemContractRequest(t, "POST", "getChangelog", changelogContractHandler("3.8.9"), strings.NewReader(`{"force":true}`))
	var response struct {
		Code int                             `json:"code"`
		Data apicontract.SystemChangelogData `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Code != 0 || !response.Data.Show ||
		response.Data.Version != "3.8.9" || !strings.Contains(response.Data.HTML, "Local stable changes") {
		t.Fatalf("stable local changelog changed: %s, %v", recorder.Body.String(), err)
	}
}

// TestAPIContractChangelogRequiresWriteRole 验证只读角色无法借读取更新日志持久化配置。
// https://github.com/siyuan-note/siyuan/security/advisories/GHSA-cvcw-64hw-hvhf
func TestAPIContractChangelogRequiresWriteRole(t *testing.T) {
	oldDir, oldConf, oldConfDir := util.WorkingDir, model.Conf, util.ConfDir
	t.Cleanup(func() { util.WorkingDir, model.Conf, util.ConfDir = oldDir, oldConf, oldConfDir })
	util.WorkingDir, util.ConfDir = t.TempDir(), t.TempDir()
	model.Conf = model.NewAppConf()
	model.Conf.Lang = "en"
	model.Conf.ShowChangelog = true
	directory := filepath.Join(util.WorkingDir, "changelogs", "v3.8.9")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "v3.8.9.en.md"), []byte("## Changes"), 0600); err != nil {
		t.Fatal(err)
	}

	recorder := changelogRouteRequest(t, model.RoleReader)
	if http.StatusForbidden != recorder.Code {
		t.Fatalf("reader reached the changelog handler: %d %s", recorder.Code, recorder.Body.String())
	}
	if !model.Conf.ShowChangelog {
		t.Fatal("reader persisted the changelog state")
	}

	if recorder = changelogRouteRequest(t, model.RoleVisitor); http.StatusOK == recorder.Code {
		t.Fatalf("visitor reached the changelog handler: %s", recorder.Body.String())
	}
	if !model.Conf.ShowChangelog {
		t.Fatal("visitor persisted the changelog state")
	}

	recorder = changelogRouteRequest(t, model.RoleAdministrator)
	if http.StatusOK != recorder.Code {
		t.Fatalf("administrator was rejected: %d %s", recorder.Code, recorder.Body.String())
	}
	if model.Conf.ShowChangelog {
		t.Fatal("administrator did not persist the changelog state")
	}
}

func changelogRouteRequest(t *testing.T, role model.Role) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(model.RoleContextKey, role) })
	handlers := contractRouteHandlers(apicontract.SystemGetChangelog, changelogContractHandler("3.8.9"))
	router.POST("/api/system/getChangelog", handlers...)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/api/system/getChangelog", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}
