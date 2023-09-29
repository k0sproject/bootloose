// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func pathSplit(r rune) bool {
	return r == '.' || r == '[' || r == ']' || r == '"'
}

// GetValueFromConfig returns specific value from object given a string path
func GetValueFromConfig(stringPath string, object interface{}) (interface{}, error) {
	keyPath := strings.FieldsFunc(stringPath, pathSplit)
	v := reflect.ValueOf(object)
	caser := cases.Title(language.English)
	for _, key := range keyPath {
		keyTitle := caser.String(key)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			v = v.FieldByName(keyTitle)
			if !v.IsValid() {
				return nil, fmt.Errorf("%v key does not exist", keyTitle)
			}
		} else if v.Kind() == reflect.Slice {
			index, errConv := strconv.Atoi(keyTitle)
			if errConv != nil {
				return nil, fmt.Errorf("%v is not an index", key)
			}
			v = v.Index(index)
		} else {
			return nil, fmt.Errorf("%v is neither a slice or a struct", v)
		}
	}
	return v.Interface(), nil
}
