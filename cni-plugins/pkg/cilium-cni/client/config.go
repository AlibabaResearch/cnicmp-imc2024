// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package client

import (
	"time"

	"github.com/cilium/cilium/api/v1/client/daemon"
	"github.com/cilium/cilium/api/v1/models"
)

const (
	ClientTimeout = 90 * time.Second
)

// ConfigGet returns a daemon configuration.
func (c *Client) ConfigGet() (*models.DaemonConfiguration, error) {
	resp, err := c.Daemon.GetConfig(nil)
	if err != nil {
		return nil, Hint(err)
	}
	return resp.Payload, nil
}

// ConfigPatch modifies the daemon configuration.
func (c *Client) ConfigPatch(cfg models.DaemonConfigurationSpec) error {
	fullCfg, err := c.ConfigGet()
	if err != nil {
		return err
	}

	for opt, value := range cfg.Options {
		fullCfg.Spec.Options[opt] = value
	}
	if cfg.PolicyEnforcement != "" {
		fullCfg.Spec.PolicyEnforcement = cfg.PolicyEnforcement
	}

	params := daemon.NewPatchConfigParams().WithConfiguration(fullCfg.Spec).WithTimeout(ClientTimeout)
	_, err = c.Daemon.PatchConfig(params)
	return Hint(err)
}
