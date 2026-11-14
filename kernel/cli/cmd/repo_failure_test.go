// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"testing"

	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/spf13/cobra"
)

func TestRepoCommandsReportUnavailableOperations(t *testing.T) {
	previousConf, previousDryRun := model.Conf, dryRun
	model.Conf = &model.AppConf{Lang: "en", Repo: model.NewConfSection(conf.NewRepo()), Sync: model.NewConfSection(conf.NewSync())}
	dryRun = false
	t.Cleanup(func() { model.Conf, dryRun = previousConf, previousDryRun })
	for _, command := range []*cobra.Command{repoCheckoutCmd} {
		t.Run(command.Name(), func(t *testing.T) {
			input := &cobra.Command{}
			input.Flags().String("id", "missing", "")
			if err := command.RunE(input, nil); err == nil {
				t.Fatal("unavailable operation reported success")
			}
		})
	}
}
