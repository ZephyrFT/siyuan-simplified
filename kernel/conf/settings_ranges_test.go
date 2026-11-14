// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package conf

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"testing"
)

func checkNormalizedIntValues[T any](t *testing.T, constructor func() *T, normalize func(*T) *T,
	field string, cases [][2]int) {
	t.Helper()
	for _, entry := range cases {
		t.Run(field+"/"+strconv.Itoa(entry[0]), func(t *testing.T) {
			value := constructor()
			reflect.ValueOf(value).Elem().FieldByName(field).SetInt(int64(entry[0]))
			value = normalize(value)
			actual := reflect.ValueOf(value).Elem().FieldByName(field).Int()
			if actual != int64(entry[1]) {
				t.Fatalf("%s: input %d, got %d, want %d", field, entry[0], actual, entry[1])
			}
			before, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			after, err := json.Marshal(normalize(value))
			if err != nil || string(before) != string(after) {
				t.Fatalf("normalization is not idempotent: %s -> %s (%v)", before, after, err)
			}
		})
	}
}

func TestNormalizeSettingsEnumRanges(t *testing.T) {
	binary := [][2]int{{-9, 0}, {-1, 0}, {0, 0}, {1, 1}, {2, 0}, {7, 0}, {math.MaxInt, 0}}
	ternary := [][2]int{{-9, 0}, {-1, 0}, {0, 0}, {1, 1}, {2, 2}, {3, 0}, {42, 0}, {math.MaxInt, 0}}
	for _, field := range []string{"BacklinkGlobalSort", "BacklinkBlockSort", "FloatWindowMode", "HeadingEmbedMode"} {
		checkNormalizedIntValues(t, NewEditor, NormalizeEditor, field, ternary)
	}
	for _, field := range []string{"DatabaseAttrClickMode", "DatabaseAttrViewMode"} {
		checkNormalizedIntValues(t, NewEditor, NormalizeEditor, field, binary)
	}
	checkNormalizedIntValues(t, NewEditor, NormalizeEditor, "CodeTabSpaces",
		[][2]int{{-1, 0}, {0, 0}, {1, 0}, {2, 2}, {3, 0}, {4, 4}, {5, 0}, {6, 6}, {7, 0}, {8, 8}, {9, 0}, {math.MaxInt, 0}})
	checkNormalizedIntValues(t, NewEditor, NormalizeEditor, "FontSize",
		[][2]int{{8, 16}, {9, 9}, {16, 16}, {72, 72}, {73, 16}, {200, 16}})
	checkNormalizedIntValues(t, NewExport, NormalizeExport, "BlockEmbedMode",
		[][2]int{{-5, 1}, {-1, 1}, {0, 0}, {1, 1}, {2, 1}, {9, 1}, {math.MaxInt, 1}})
	checkNormalizedIntValues(t, NewExport, NormalizeExport, "FileAnnotationRefMode", binary)
	checkNormalizedIntValues(t, NewAppearance, func(value *Appearance) *Appearance {
		return NormalizeAppearance(value, false)
	}, "CloseButtonBehavior", binary)
	checkNormalizedIntValues(t, NewSync, NormalizeSync, "AssetDownloadMode", binary)
	checkNormalizedIntValues(t, NewSystem, NormalizeSystem, "AutoLaunch2", ternary)
	checkNormalizedIntValues(t, NewSystem, NormalizeSystem, "LockScreenMode", binary)
	checkNormalizedIntValues(t, NewFlashcard, func(value *Flashcard) *Flashcard {
		value, warning := NormalizeFlashcard(value)
		if warning != "" {
			t.Fatal(warning)
		}
		return value
	}, "ReviewMode", ternary)
}

func TestNormalizeSettingsNumericRanges(t *testing.T) {
	checkNormalizedIntValues(t, NewSearch, NormalizeSearch, "Limit",
		[][2]int{{-1, 64}, {0, 64}, {1, 32}, {31, 32}, {32, 32}, {64, 64}, {10240, 10240}, {10241, 10240}, {1000000, 10240}})
	checkNormalizedIntValues(t, NewFileTree, NormalizeFileTree, "MaxListCount",
		[][2]int{{-1, 512}, {0, 512}, {1, 1}, {512, 512}, {10240, 10240}, {10241, 10240}, {1000000, 10240}})
	checkNormalizedIntValues(t, NewFileTree, NormalizeFileTree, "LargeFileWarningSize",
		[][2]int{{-1, 8}, {0, 8}, {1, 8}, {2, 2}, {8, 8}, {10240, 10240}, {10241, 10240}, {1000000, 10240}})
	checkNormalizedIntValues(t, NewRepo, NormalizeRepo, "IndexRetentionDays",
		[][2]int{{0, 180}, {1, 1}, {180, 180}, {3650, 3650}, {100000, 100000}})
}
