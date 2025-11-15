// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"gopkg.in/yaml.v3"
)

type leaders map[string]Leader

type Leaders struct {
	Leaders leaders `yaml:"leaders"`
}

type Leader struct {
	Keyspace  string   `yaml:"keyspace"`
	Shortcuts []Plugin `yaml:"shortcuts"`
}

func NewLeaders() Leaders {
	return Leaders{
		Leaders: make(map[string]Leader),
	}
}

// Load K9s leaders.
func (p Leaders) LoadLeaders(path string, loadExtra bool) error {
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

func (p *Leaders) load(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	scheme, err := data.JSONValidator.ValidatePlugins(bb)
	if err != nil {
		slog.Warn("Leaders schema validation failed",
			slogs.Path, path,
			slogs.Error, err,
		)
		return fmt.Errorf("leaders validation failed for %s: %w", path, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(bb))
	d.KnownFields(true)

	switch scheme {
	case json.PluginLeadersSchema:
		var oo Leaders
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("leaders unmarshal failed for %s: %w", path, err)
		}
		for k := range oo.Leaders {
			value := oo.Leaders[k]
			p.Leaders[value.Keyspace] = value
		}
	}

	return nil
}

func (p Leaders) loadDir(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	var errs error
	errs = errors.Join(errs, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !isYamlFile(info.Name()) {
			return nil
		}
		errs = errors.Join(errs, p.load(path))
		return nil
	}))

	return errs
}
