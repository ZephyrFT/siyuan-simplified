// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package apicontract

type LoadedPluginRequest struct {
	Name string `json:"name"`
}

type LoadedPlugin struct {
	Name      string             `json:"name"`
	State     string             `json:"state"`
	StateCode int                `json:"stateCode"`
	Methods   []*PluginRPCMethod `json:"methods"`
}

type PluginRPCMethod struct {
	Name         string   `json:"name"`
	Descriptions []string `json:"descriptions"`
}

type pluginNameError struct {
	code    int
	message string
}

func (e *pluginNameError) Error() string { return e.message }
