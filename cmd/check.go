/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/storage"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Проверяет конфиг и базу данных",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.ReadConfig()
		if err != nil {
			return err
		}

		db, err := storage.NewFileMetaStorage(conf.StoragePath)
		if err != nil {
			return err
		}

		spew.Dump(conf, db.GetTags())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
