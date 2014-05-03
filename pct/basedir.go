/*
   Copyright (c) 2014, Percona LLC and/or its affiliates. All rights reserved.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package pct

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	DEFAULT_BASEDIR    = "/usr/local/percona/percona-agent"
	CONFIG_FILE_SUFFIX = ".conf"
	// Relative to Basedir.path:
	CONFIG_DIR      = "config"
	DATA_DIR        = "data"
	BIN_DIR         = "bin"
	START_LOCK_FILE = "start.lock"
)

type basedir struct {
	path      string
	configDir string
	dataDir   string
	binDir    string
}

var Basedir basedir

func (b *basedir) Init(path string) error {
	var err error
	b.path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := MakeDir(b.path); err != nil && !os.IsExist(err) {
		return err
	}

	b.configDir = filepath.Join(b.path, CONFIG_DIR)
	if err := MakeDir(b.configDir); err != nil && !os.IsExist(err) {
		return err
	}

	b.dataDir = filepath.Join(b.path, DATA_DIR)
	if err := MakeDir(b.dataDir); err != nil && !os.IsExist(err) {
		return err
	}

	b.binDir = filepath.Join(b.path, BIN_DIR)
	if err := MakeDir(b.binDir); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (b *basedir) Path() string {
	return b.path
}

func (b *basedir) Dir(service string) string {
	switch service {
	case "config":
		return b.configDir
	case "data":
		return b.dataDir
	case "bin":
		return b.binDir
	default:
		log.Panic("Invalid service: " + service)
	}
	return ""
}

func (b *basedir) ConfigFile(service string) string {
	return filepath.Join(b.configDir, service+CONFIG_FILE_SUFFIX)
}

func (b *basedir) ReadConfig(service string, v interface{}) error {
	configFile := filepath.Join(b.configDir, service+CONFIG_FILE_SUFFIX)
	data, err := ioutil.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		// There's an error and it's not "file not found".
		return err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &v)
	}
	return err
}

func (b *basedir) WriteConfig(service string, config interface{}) error {
	configFile := filepath.Join(b.configDir, service+CONFIG_FILE_SUFFIX)
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, data, 0644)
}

func (b *basedir) WriteConfigString(service, config string) error {
	configFile := filepath.Join(b.configDir, service+CONFIG_FILE_SUFFIX)
	return ioutil.WriteFile(configFile, []byte(config), 0644)
}

func (b *basedir) RemoveConfig(service string) error {
	configFile := filepath.Join(b.configDir, service+CONFIG_FILE_SUFFIX)
	return RemoveFile(configFile)
}

func (b *basedir) File(file string) string {
	switch file {
	case "start-lock":
		file = START_LOCK_FILE
	default:
		log.Panicf("Unknown basedir file: %s", file)
	}
	return filepath.Join(b.Path(), file)
}
