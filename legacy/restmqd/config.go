// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
)

type ConfigData struct {
	XMLName xml.Name `xml:"Server"`
	Debug   bool

	// http
	Addr     string `xml:",attr"`
	XHeaders bool   `xml:",attr"`

	SSL struct {
		Addr     string `xml:",attr"`
		CertFile string
		KeyFile  string
	}

	// databases
	Redis string
}

// ReadConfig reads and parses the XML configuration file.
func ReadConfig(filename string, cfg *ConfigData) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(buf, cfg); err != nil {
		return err
	}
	// Make file paths relative to the config file's dir.
	basedir := filepath.Dir(filename)
	relativePath(basedir, &cfg.SSL.CertFile)
	relativePath(basedir, &cfg.SSL.KeyFile)
	return nil
}

func relativePath(basedir string, path *string) {
	p := *path
	if p != "" && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
