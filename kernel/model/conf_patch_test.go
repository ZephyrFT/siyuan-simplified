// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"encoding/json"
	"testing"

	"github.com/siyuan-note/siyuan/kernel/conf"
)

func TestMergeConfSectionJSONFieldCaseDeterministic(t *testing.T) {
	for range 20 {
		merged, err := MergeConfSectionJSON(conf.NewEditor(), []byte(`{"BlockGutterMode":false,"blockGutterMode":true}`))
		if err != nil {
			t.Fatal(err)
		}
		var next conf.Editor
		if err = json.Unmarshal(merged, &next); err != nil || !next.BlockGutterMode {
			t.Fatalf("canonical field did not win consistently: %s (%v)", merged, err)
		}
	}
}

func TestMergeConfJSONKeepsExactKeys(t *testing.T) {
	// 没有结构体类型的通用合并只按精确键名匹配，由调用方决定是否需要大小写兼容。
	merged, err := MergeConfJSON([]byte(`{"openFile":"F1"}`), []byte(`{"openfile":"F2"}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(merged) != `{"openFile":"F1","openfile":"F2"}` {
		t.Fatalf("untyped merge conflated independently named keys: %s", merged)
	}
}
