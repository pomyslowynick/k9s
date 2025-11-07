// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"path/filepath"

	"github.com/adrg/xdg"
)

func NewLeaders() Plugins {
	return Plugins{
		Plugins: make(map[string]Plugin),
	}
}

// Load K9s plugins.
func (p Plugins) LoadLeaders(path string, loadExtra bool) error {
	var errs error

	// Load from global config file
	if err := p.load(AppLeadersFile); err != nil {
		errs = errors.Join(errs, err)
	}

	// Load from cluster/context config
	// if err := p.load(path); err != nil {
	// 	errs = errors.Join(errs, err)
	// }

	if !loadExtra {
		return errs
	}
	// Load from XDG dirs
	const k9sLeadersDir = "k9s/leaders"
	for _, dir := range append(xdg.DataDirs, xdg.DataHome, xdg.ConfigHome) {
		path := filepath.Join(dir, k9sLeadersDir)
		if err := p.loadDir(path); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}
