// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/k0sproject/bootloose/pkg/config"
	"github.com/spf13/cobra"
)

func NewConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get config file information",
		RunE:  getConfig,
	}
}

func getConfig(cmd *cobra.Command, args []string) error {
	c, err := config.NewConfigFromFile(clusterConfigFile(cmd))
	if err != nil {
		return err
	}
	var detail interface{}
	if len(args) > 0 {
		detail, err = config.GetValueFromConfig(args[0], c)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("Failed to get config detail")
		}
	} else {
		detail = c
	}
	if reflect.ValueOf(detail).Kind() != reflect.String {
		res, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			log.Println(err)
			return fmt.Errorf("Cannot convert result to json")
		}
		fmt.Printf("%s", res)
	} else {
		fmt.Printf("%s", detail)
	}
	return nil
}
