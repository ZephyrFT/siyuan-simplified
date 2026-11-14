// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/88250/gulu"
	"github.com/gin-gonic/gin"
	"github.com/siyuan-note/siyuan/kernel/apicontract"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
)

type settingConfigInput interface {
	ConfigJSON() []byte
	ConfigError() error
	HasField(string) bool
}

func compareSettingConfig[Request settingConfigInput, Data, Config any](t *testing.T, endpoint apicontract.Endpoint[Request, Data], factory func() Config, bodies []string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	for _, body := range bodies {
		t.Run(endpoint.Definition().Name+body, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("POST", endpoint.Definition().Path, strings.NewReader(body))
			result := gulu.Ret.NewResult()
			fields, ok := legacyJSONArg(context, result)
			if !ok {
				t.Fatal(result.Msg)
			}
			encoded, err := gulu.JSON.MarshalJSON(fields)
			if err != nil {
				t.Fatal(err)
			}
			before, after := factory(), factory()
			beforeErr := gulu.JSON.UnmarshalJSON(encoded, before)
			request, err := endpoint.Decode(strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			afterErr := request.ConfigError()
			if (beforeErr == nil) != (afterErr == nil) || beforeErr != nil && beforeErr.Error() != afterErr.Error() {
				t.Fatalf("config validation changed: before=%v after=%v", beforeErr, afterErr)
			}
			if beforeErr != nil {
				return
			}
			if err = gulu.JSON.UnmarshalJSON(request.ConfigJSON(), after); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("partial configuration changed: before=%#v after=%#v", before, after)
			}
			for _, key := range []string{"fontFamilies", "codeFontFamilies", "bodyGradient", "globalFontFamilies"} {
				_, exists := fields[key]
				if request.HasField(key) != exists {
					t.Fatalf("key presence changed: %s", key)
				}
			}
		})
	}
}

func TestAPIContractSettingConfigCompatibility(t *testing.T) {
	common := []string{`null`, `{}`, `{"unknown":{"nested":[null,1.0]}}`}
	compareSettingConfig(t, apicontract.SetConfSnippet, func() *conf.Snpt { return &conf.Snpt{} }, append(common, `{"enabledCSS":null}`, `{"enabledCSS":true}`))
	compareSettingConfig(t, apicontract.SetSecrets, func() *conf.Secrets { return &conf.Secrets{} }, append(common, `{"items":[null,{"allowedHosts":[null,"localhost"]}]}`))
	compareSettingConfig(t, apicontract.SetVariables, func() *conf.Variables { return &conf.Variables{} }, append(common, `{"items":[null,{"name":null}]}`))
	compareSettingConfig(t, apicontract.SetFlashcard, conf.NewFlashcard, append(common, `{"newCardLimit":1e0}`, `{"NewCardLimit":7,"newCardLimit":2}`, `{"newCardLimit":1.5}`))
	compareSettingConfig(t, apicontract.SetEditor, conf.NewEditor, append(common, `{"backlinkBlockSort":1}`, `{"backlinkBlockSort":2}`, `{"backlinkBlockSort":null}`, `{"backlinkBlockSort":"1"}`, `{"markdown":{"inlineMath":null}}`, `{"markdown":null,"fontFamilies":null}`, `{"FontFamilies":[],"fontSize":16.0}`, `{"fontSize":1.5}`, `{"markdown":false}`, `{"emoji":[null,"x"]}`))
	compareSettingConfig(t, apicontract.SetExport, func() *conf.Export { return &conf.Export{} }, append(common, `{"blockRefMode":1.0}`, `{"blockRefMode":1.5}`))
	compareSettingConfig(t, apicontract.SetFiletree, conf.NewFileTree, append(common, `{"maxOpenTabCount":1e0,"boxDocEnabled":null}`, `{"useSVGDefaultIcon":null,"tabStartupMode":null}`, `{"maxOpenTabCount":"x"}`))
	compareSettingConfig(t, apicontract.SetSearch, func() *conf.Search { return &conf.Search{} }, append(common, `{"hanSensitive":null,"limit":3.0}`, `{"limit":3.2}`, `{"customBlock":true}`, `{"customBlock":false}`, `{"customBlock":null}`, `{"customBlock":"bad"}`))
	compareSettingConfig(t, apicontract.SetAppearance, func() *conf.Appearance { return &conf.Appearance{} }, append(common, `{"bodyGradient":null}`, `{"BodyGradient":null,"globalFontFamilies":[]}`, `{"statusBar":false}`, `{"notifications":{"selectAllTip":null}}`))
	compareSettingConfig(t, apicontract.SetEntryVisibility, func() *conf.EntryVisibility { return &conf.EntryVisibility{} }, append(common, `{"profiles":[null,{"orders":{"menu":[null,"id"]}}]}`))
}

func TestAPIContractSettingPreferencesCompatibility(t *testing.T) {
	compareSettingConfig(t, apicontract.SetAppearance, conf.NewAppearance, []string{
		`{}`, `{"compactMode":true,"menuMouseScroll":false,"outlineKeepCurrentExpand":true}`,
		`{"pdfTheme":{"light":"dark","dark":"light","annoColor":"red"},"mobileBars":{"autoHide":false}}`,
		`{"showFloatingDockBar":true}`, `{"showFloatingDockBar":false}`,
		`{"showFloatingDockBar":null}`, `{"ShowFloatingDockBar":false}`, `{"showFloatingDockBar":"bad"}`,
		`{"showFloatingStatusBar":true}`, `{"showFloatingStatusBar":false}`, `{"showFloatingStatusBar":null}`,
		`{"ShowFloatingStatusBar":true}`, `{"showFloatingStatusBar":"bad"}`,
		`{"pdfTheme":null,"mobileBars":null}`, `{"mobileBars":{"sidebarSwipe":null,"sidebarButtons":false}}`,
		`{"compactMode":"true"}`, `{"pdfTheme":{"light":false}}`, `{"mobileBars":{"autoHide":1}}`,
	})
	compareSettingConfig(t, apicontract.SetFiletree, conf.NewFileTree, []string{
		`{}`, `{"recentDocsSort":"updated"}`, `{"recentDocsSort":null}`, `{"recentDocsSort":true}`,
	})
}

func TestAPIContractSettingMindmapCompatibility(t *testing.T) {
	compareSettingConfig(t, apicontract.SetSearch, func() *conf.Search { return &conf.Search{} }, []string{
		`{}`, `{"mindmap":true,"mindmapItem":false}`, `{"mindmap":false,"mindmapItem":true}`,
		`{"mindmap":null,"mindmapItem":null}`, `{"mindmap":"bad"}`, `{"mindmapItem":1}`,
	})
}

func TestAPIContractSettingFileTreeClickModeCompatibility(t *testing.T) {
	compareSettingConfig(t, apicontract.SetFiletree, conf.NewFileTree, []string{
		`{}`, `{"docIconClickMode":1,"parentDocTitleClickMode":2}`,
		`{"docIconClickMode":0}`, `{"docIconClickMode":null}`, `{"DocIconClickMode":1}`,
		`{"parentDocTitleClickMode":0}`, `{"parentDocTitleClickMode":1}`, `{"parentDocTitleClickMode":null}`,
		`{"ParentDocTitleClickMode":2}`, `{"docIconClickMode":2}`, `{"parentDocTitleClickMode":3}`,
		`{"docIconClickMode":"1"}`, `{"docIconClickMode":1.5}`, `{"parentDocTitleClickMode":true}`,
		// 已废弃的布尔字段不再被读取
		`{"docIconClickExpand":true}`, `{"parentDocClickExpand":true,"parentDocDoubleClickOpen":false}`,
	})
}

func TestAPIContractSettingImageDisplayCompatibility(t *testing.T) {
	compareSettingConfig(t, apicontract.SetEditor, conf.NewEditor, []string{
		`{}`, `{"displayImgName":true}`, `{"displayImgName":false}`,
		`{"displayImgName":null}`, `{"DisplayImgName":true}`, `{"displayImgName":"bad"}`,
		`{"displayImgAlt":true}`, `{"displayImgAlt":false}`, `{"displayImgAlt":null}`,
		`{"DisplayImgAlt":true}`, `{"displayImgAlt":"bad"}`,
	})
}

func TestAPIContractSettingImageDisplayUpdates(t *testing.T) {
	previousConf, previousReadOnly, previousMarkdown := model.Conf, util.ReadOnly, util.MarkdownSettings.Load()
	model.Conf = model.NewAppConf()
	model.Conf.Editor.Store(conf.NewEditor())
	util.ReadOnly = true
	t.Cleanup(func() {
		{
			derivedValue0, derivedValue1, derivedValue2 := previousConf, previousReadOnly, previousMarkdown
			model.Conf = derivedValue0
			util.ReadOnly = derivedValue1
			util.MarkdownSettings.Store(derivedValue2)
		}
	})
	for _, entry := range []struct {
		body string
		want bool
	}{
		{`{"displayImgName":true}`, true},
		{`{}`, true},
		{`{"displayImgName":false}`, false},
		{`{"DisplayImgName":true}`, true},
		{`{"displayImgName":null}`, false},
	} {
		code, message, data := settingContractRequest(t, "setEditor", setEditor, entry.body)
		var response apicontract.SettingEditor
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatal(err)
		}
		if code != 0 || model.Conf.Editor.Load().DisplayImgName != entry.want || response.DisplayImgName != entry.want {
			t.Fatalf("image display update %s: code=%d message=%s data=%s", entry.body, code, message, data)
		}
	}
	for _, entry := range []struct {
		body string
		want bool
	}{
		{`{"displayImgAlt":true}`, true}, {`{}`, true}, {`{"displayImgAlt":false}`, false},
		{`{"DisplayImgAlt":true}`, true}, {`{"displayImgAlt":null}`, false},
	} {
		code, message, data := settingContractRequest(t, "setEditor", setEditor, entry.body)
		var response apicontract.SettingEditor
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatal(err)
		}
		if code != 0 || model.Conf.Editor.Load().DisplayImgAlt != entry.want || response.DisplayImgAlt != entry.want {
			t.Fatalf("image tooltip text update %s: code=%d message=%s data=%s", entry.body, code, message, data)
		}
	}
}

func TestAPIContractSettingParserCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{"", `{`, `[]`, `false`, `{"unknown":1e1000}`, `{"app":1e1000,"app":1}`, `{"fontSize":16}{"extra":true}`} {
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest("POST", "/api/setting/setEditor", strings.NewReader(body))
		result := gulu.Ret.NewResult()
		_, ok := legacyJSONArg(context, result)
		_, err := apicontract.SetEditor.Decode(strings.NewReader(body))
		if ok != (err == nil) || err != nil && err.Error() != result.Msg {
			t.Fatalf("parser changed for %s: before=%s after=%v", body, result.Msg, err)
		}
	}
}

func settingContractRequest(t *testing.T, route string, handler gin.HandlerFunc, body string) (int, string, json.RawMessage) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	path := "/api/setting/" + route
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("POST", path, strings.NewReader(body))
	handler(context)
	requireAPIContract(t, "POST", path, recorder)
	var result struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result.Code, result.Msg, result.Data
}

func TestAPIContractSettingErrorsAndAdmission(t *testing.T) {
	for _, entry := range []struct {
		route               string
		handler             gin.HandlerFunc
		body, message, data string
	}{
		{"setTheme", setTheme, `{"theme":"theme","modes":[null]}`, "[modes] is required ([0] for light, [1] for dark, [0,1] for both)", "null"},
		{"setExport", setExport, `{"blockRefMode":1.5}`, "json: cannot unmarshal number 1.5 into Go struct field Export.blockRefMode of type int", `{"closeTimeout":5000}`},
	} {
		code, msg, data := settingContractRequest(t, entry.route, entry.handler, entry.body)
		if code != -1 || msg != entry.message || string(data) != entry.data {
			t.Fatalf("%s changed: %d %s %s", entry.route, code, msg, data)
		}
	}
}

func TestAPIContractSettingEmojiAndPublishValidation(t *testing.T) {
	previous, readonly, confDir := model.Conf, util.ReadOnly, util.ConfDir
	util.ReadOnly, util.ConfDir = true, t.TempDir()
	model.Conf = model.NewAppConf()
	model.Conf.Editor.Store(conf.NewEditor())
	model.Conf.Publish.Store(conf.NewPublish())
	t.Cleanup(func() { model.Conf, util.ReadOnly, util.ConfDir = previous, readonly, confDir })
	code, _, data := settingContractRequest(t, "setEmoji", setEmoji, `{"emoji":[]}`)
	if code != 0 || string(data) != "null" || model.Conf.Editor.Load().Emoji == nil || len(model.Conf.Editor.Load().Emoji) != 0 {
		t.Fatal("empty emoji list changed")
	}
}

func fillSettingPayload(value reflect.Value) {
	switch value.Kind() {
	case reflect.Pointer:
		value.Set(reflect.New(value.Type().Elem()))
		fillSettingPayload(value.Elem())
	case reflect.Struct:
		for _, field := range value.Fields() {
			if field.CanSet() {
				fillSettingPayload(field)
			}
		}
		if value.Type() == reflect.TypeFor[conf.EntryVisibilityProfile]() {
			value.FieldByName("Defaults").SetString(conf.EntryVisibilityProfileSimple)
		}
	case reflect.Slice:
		value.Set(reflect.MakeSlice(value.Type(), 2, 2))
		fillSettingPayload(value.Index(0))
	case reflect.Map:
		value.Set(reflect.MakeMap(value.Type()))
		key, item := reflect.New(value.Type().Key()).Elem(), reflect.New(value.Type().Elem()).Elem()
		fillSettingPayload(key)
		fillSettingPayload(item)
		value.SetMapIndex(key, item)
	case reflect.String:
		value.SetString("value")
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(2)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value.SetUint(2)
	case reflect.Float32, reflect.Float64:
		value.SetFloat(2.5)
	}
}

func compareSettingPayload[Source, Data any](t *testing.T, route string, convert func(*Source) *Data) {
	t.Helper()
	for _, populate := range []bool{false, true} {
		source := new(Source)
		if populate {
			fillSettingPayload(reflect.ValueOf(source).Elem())
		}
		before, err := json.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		after, err := json.Marshal(convert(source))
		if err != nil {
			t.Fatal(err)
		}
		var beforeValue, afterValue any
		if err = json.Unmarshal(before, &beforeValue); err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(after, &afterValue); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(beforeValue, afterValue) {
			t.Fatalf("%s response mapping changed:\n%s\n%s", route, before, after)
		}
		if route == "" {
			continue
		}
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		payload, err := json.Marshal(apicontract.Success(convert(source)))
		if err != nil {
			t.Fatal(err)
		}
		recorder.Write(payload)
		requireAPIContract(t, "POST", "/api/setting/"+route, recorder)
	}
}

func TestAPIContractSettingCompletePayloads(t *testing.T) {
	compareSettingPayload(t, "setSnippet", settingSnptPayload)
	compareSettingPayload(t, "setSecrets", settingSecretsPayload)
	compareSettingPayload(t, "setVariables", settingVariablesPayload)
	compareSettingPayload(t, "setFlashcard", settingFlashcardPayload)
	compareSettingPayload(t, "setEditor", settingEditorPayload)
	compareSettingPayload(t, "setExport", settingExportPayload)
	compareSettingPayload(t, "setFiletree", settingFileTreePayload)
	compareSettingPayload(t, "setSearch", settingSearchPayload)
	compareSettingPayload(t, "setAppearance", settingAppearancePayload)
	compareSettingPayload(t, "setEntryVisibility", settingEntryVisibilityPayload)
	compareSettingPayload(t, "setBootAppearance", settingBootAppearanceSelectionPayload)
	compareSettingPayload(t, "", settingBootAppearancePayload)
}
