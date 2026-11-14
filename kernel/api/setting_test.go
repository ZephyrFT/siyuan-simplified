// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func TestSetEditorNullRestoresDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousConf, previousReadOnly, previousMarkdown := model.Conf, util.ReadOnly, util.MarkdownSettings.Load()
	util.ReadOnly = true
	t.Cleanup(func() {
		model.Conf, util.ReadOnly = previousConf, previousReadOnly
		util.MarkdownSettings.Store(previousMarkdown)
	})
	model.Conf = model.NewAppConf()
	model.Conf.Editor.Store(conf.NewEditor())
	model.Conf.Editor.Load().DisplayBookmarkIcon = false
	model.Conf.Editor.Load().Emoji = []string{"1f600"}

	code, message, _ := settingContractRequest(t, "setEditor", setEditor, `{"displayBookmarkIcon":null,"emoji":null}`)
	if code != 0 {
		t.Fatalf("save failed: %s", message)
	}
	editor := model.Conf.Editor.Load()
	if !editor.DisplayBookmarkIcon || !reflect.DeepEqual(editor.Emoji, conf.NewEditor().Emoji) {
		t.Fatalf("null did not restore editor defaults: %+v", editor)
	}
}

func TestAPIContractSettingEditorLegacyFont(t *testing.T) {
	for _, route := range []struct {
		name    string
		handler gin.HandlerFunc
	}{{"setEditor", setEditor}, {"patch", patchSetting}} {
		for _, test := range []struct {
			name    string
			body    string
			family  string
			weight  int
			count   int
			display string
		}{
			{"change font", `{"fontFamily":"Verdana","fontWeight":600}`, "Verdana", 600, 1, ""},
			{"change weight", `{"fontWeight":600}`, "Arial", 600, 1, ""},
			{"case-insensitive fields", `{"FontFamily":"Verdana","FontWeight":600}`, "Verdana", 600, 1, ""},
			{"unchanged primary keeps fallbacks", `{"fontFamily":"Arial","fontWeight":400}`, "Arial", 400, 2, "Arial Regular"},
			{"unrelated field keeps fallbacks", `{"fontSize":18}`, "Arial", 400, 2, "Arial Regular"},
			{"explicit list wins", `{"fontFamily":"Verdana","fontWeight":600,"fontFamilies":[{"family":"Georgia","weight":500}]}`, "Georgia", 500, 1, ""},
			{"case-insensitive list wins", `{"fontFamily":"Verdana","FontFamilies":[{"family":"Georgia","weight":500}]}`, "Georgia", 500, 1, ""},
			{"null legacy font resets", `{"fontFamily":null}`, "", 400, 0, ""},
			{"null font list rebuilds", `{"fontFamily":"Verdana","fontFamilies":null}`, "Verdana", 400, 1, ""},
			{"explicit legacy display", `{"fontFamily":"Verdana","fontFamilyDisplay":"Verdana Regular"}`, "Verdana", 400, 1, "Verdana Regular"},
		} {
			t.Run(route.name+"/"+test.name, func(t *testing.T) {
				previousConf, previousReadOnly, previousMarkdown := model.Conf, util.ReadOnly, util.MarkdownSettings.Load()
				util.ReadOnly = true
				t.Cleanup(func() {
					model.Conf, util.ReadOnly = previousConf, previousReadOnly
					util.MarkdownSettings.Store(previousMarkdown)
				})
				model.Conf = model.NewAppConf()
				editor := conf.NewEditor()
				editor.FontFamilies = []*conf.EditorFont{{Family: "Arial", Weight: 400, DisplayName: "Arial Regular"}, {Family: "serif", Weight: 400}}
				editor.CodeFontFamilies = []*conf.EditorFont{{Family: "monospace", Weight: 400}}
				editor.NormalizeFontFamilies()
				model.Conf.Editor.Store(editor)
				body := test.body
				if test.name == "change weight" && route.name == "setEditor" {
					body = `{"fontWeight":600.0}`
				}
				if route.name == "patch" {
					body = `{"editor":` + body + `}`
				}
				code, message, _ := settingContractRequest(t, route.name, route.handler, body)
				editor = model.Conf.Editor.Load()
				if code != 0 || editor.FontFamily != test.family || editor.FontWeight != test.weight || len(editor.FontFamilies) != test.count {
					t.Fatalf("font write failed: %s (%+v)", message, editor)
				}
				if editor.FontFamilyDisplay != test.display {
					t.Fatalf("font display name did not follow the font: %q", editor.FontFamilyDisplay)
				}
				if len(editor.CodeFontFamilies) != 1 || editor.CodeFontFamilies[0].Family != "monospace" {
					t.Fatal("legacy font write changed the code font list")
				}
			})
		}
	}
}

func TestSetFiletreeMergeAndNull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousConf, previousReadOnly := model.Conf, util.ReadOnly
	util.ReadOnly = true
	t.Cleanup(func() { model.Conf, util.ReadOnly = previousConf, previousReadOnly })

	engine := gin.New()
	engine.POST("/api/setting/setFiletree", setFiletree)
	engine.POST("/api/setting/patch", patchSetting)
	for _, test := range []struct {
		name string
		path string
		body string
		want conf.FileTree
	}{
		// 未提交的字段保持当前配置
		{
			name: "omitted keeps current",
			path: "/api/setting/setFiletree",
			body: `{}`,
			want: conf.FileTree{BoxDocEnabled: true, UseSVGDefaultIcon: false, UseSingleLineSave: false, Sort: 3},
		},
		{
			name: "explicit values",
			path: "/api/setting/setFiletree",
			body: `{"boxDocEnabled":false,"useSVGDefaultIcon":true}`,
			want: conf.FileTree{UseSVGDefaultIcon: true, UseSingleLineSave: false, Sort: 3},
		},
		// null 还原构造函数默认值，而不是零值
		{
			name: "null restores defaults",
			path: "/api/setting/setFiletree",
			body: `{"useSingleLineSave":null,"sort":null}`,
			want: conf.FileTree{BoxDocEnabled: true, UseSVGDefaultIcon: false,
				UseSingleLineSave: util.DefaultUseSingleLineSave, Sort: util.SortModeCustom},
		},
		{
			name: "patch null restores defaults",
			path: "/api/setting/patch",
			body: `{"fileTree":{"useSingleLineSave":null,"sort":null}}`,
			want: conf.FileTree{BoxDocEnabled: true, UseSVGDefaultIcon: false,
				UseSingleLineSave: util.DefaultUseSingleLineSave, Sort: util.SortModeCustom},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			model.Conf = model.NewAppConf()
			model.Conf.FileTree.Store(conf.NewFileTree())
			model.Conf.FileTree.Load().BoxDocEnabled = true
			model.Conf.FileTree.Load().UseSVGDefaultIcon = false
			model.Conf.FileTree.Load().UseSingleLineSave = false
			model.Conf.FileTree.Load().Sort = 3

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(recorder, request)
			var response struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Code != 0 {
				t.Fatalf("setting failed: %s (%v)", recorder.Body.String(), err)
			}
			got := model.Conf.FileTree.Load()
			if got.BoxDocEnabled != test.want.BoxDocEnabled || got.UseSVGDefaultIcon != test.want.UseSVGDefaultIcon ||
				got.UseSingleLineSave != test.want.UseSingleLineSave || got.Sort != test.want.Sort {
				t.Fatalf("unexpected file tree settings: %+v", got)
			}
		})
	}
}
