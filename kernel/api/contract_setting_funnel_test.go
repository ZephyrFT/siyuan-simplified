// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"testing"

	"github.com/siyuan-note/siyuan/kernel/apicontract"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func TestAPIContractSettingPatchDiagnosticsAndAdditionalSections(t *testing.T) {
	previous, readonly, directory := model.Conf, util.ReadOnly, util.ConfDir
	t.Cleanup(func() { model.Conf, util.ReadOnly, util.ConfDir = previous, readonly, directory })
	model.Conf = model.NewAppConf()
	util.ConfDir, util.ReadOnly = t.TempDir(), false
	model.NormalizeAppConf(model.Conf)
	code, message, data := settingContractRequest(t, "patch", patchSetting, `{"flashcard":{"weights":"invalid"}}`)
	if code != 0 {
		t.Fatalf("warning blocked valid normalized configuration: %s", message)
	}
	var diagnostics apicontract.SettingPatchData
	if err := json.Unmarshal(data, &diagnostics); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics.Issues) != 1 || diagnostics.Issues[0].Field != "flashcard.weights" || diagnostics.Issues[0].Severity != "warning" ||
		model.Conf.Flashcard.Load().Weights != conf.DefaultFSRSWeights() {
		t.Fatal("normalization warning was not returned")
	}
	for _, namespace := range []string{"ai", "bazaar", "publish"} {
		if code, _, _ := settingContractRequest(t, "patch", patchSetting, `{"`+namespace+`":{}}`); code == 0 {
			t.Fatalf("removed settings namespace %s was accepted", namespace)
		}
	}
	for _, body := range []string{`{"system":{"settingsWindow":true}}`, `{"repo":{"indexRetentionDays":42}}`, `{"snippet":{"enabledCSS":false}}`} {
		code, message, data := settingContractRequest(t, "patch", patchSetting, body)
		if code != 0 || string(data) != "null" {
			t.Fatalf("additional patch failed or changed the null response: %s %d %s %s", body, code, message, data)
		}
	}
	if !model.Conf.System.Load().SettingsWindow || model.Conf.Repo.Load().IndexRetentionDays != 42 || model.Conf.Snippet.Load().EnabledCSS {
		t.Fatal("additional patch did not apply")
	}
	for _, body := range []string{`{"system":{"id":"replacement"}}`, `{"repo":{"key":"replacement"}}`} {
		code, _, _ := settingContractRequest(t, "patch", patchSetting, body)
		if code == 0 {
			t.Fatalf("restricted field accepted: %s", body)
		}
	}
}
