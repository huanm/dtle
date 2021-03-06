/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/hashicorp/nomad, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package config

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	consul "github.com/hashicorp/consul/api"

	"github.com/actiontech/dtle/internal"
)

// ConsulConfig contains the configuration information necessary to
// communicate with a Consul Agent in order to:
//
// - Register services and their checks with Consul
//
// - Bootstrap this Udup Client with the list of Udup Servers registered
//   with Consul
//
// Both the Agent and the executor need to be able to import ConsulConfig.
type ConsulConfig struct {
	// ServerServiceName is the name of the service that Udup uses to register
	// servers with Consul
	ServerServiceName string `mapstructure:"server_service_name"`

	// ClientServiceName is the name of the service that Udup uses to register
	// clients with Consul
	ClientServiceName string `mapstructure:"client_service_name"`

	// AutoAdvertise determines if this Udup Agent will advertise its
	// services via Consul.  When true, Udup Agent will register
	// services with Consul.
	AutoAdvertise *bool `mapstructure:"auto_advertise"`

	// ChecksUseAdvertise specifies that Consul checks should use advertise
	// address instead of bind address
	ChecksUseAdvertise *bool `mapstructure:"checks_use_advertise"`

	// Addr is the address of the local Consul agent
	Addr string `mapstructure:"address"`

	// Timeout is used by Consul HTTP Client
	Timeout time.Duration `mapstructure:"timeout"`

	// Token is used to provide a per-request ACL token. This options overrides
	// the agent's default token
	Token string `mapstructure:"token"`

	// Auth is the information to use for http access to Consul agent
	Auth string `mapstructure:"auth"`

	// EnableSSL sets the transport scheme to talk to the Consul agent as https
	EnableSSL *bool `mapstructure:"ssl"`

	// VerifySSL enables or disables SSL verification when the transport scheme
	// for the consul api client is https
	VerifySSL *bool `mapstructure:"verify_ssl"`

	// CAFile is the path to the ca certificate used for Consul communication
	CAFile string `mapstructure:"ca_file"`

	// CertFile is the path to the certificate for Consul communication
	CertFile string `mapstructure:"cert_file"`

	// KeyFile is the path to the private key for Consul communication
	KeyFile string `mapstructure:"key_file"`

	// ServerAutoJoin enables Udup servers to find peers by querying Consul and
	// joining them
	ServerAutoJoin *bool `mapstructure:"server_auto_join"`

	// ClientAutoJoin enables Udup servers to find addresses of Udup servers
	// and register with them
	ClientAutoJoin *bool `mapstructure:"client_auto_join"`
}

// DefaultConsulConfig() returns the canonical defaults for the Udup
// `consul` configuration.
func DefaultConsulConfig() *ConsulConfig {
	return &ConsulConfig{
		ServerServiceName:  "server",
		ClientServiceName:  "client",
		AutoAdvertise:      internal.BoolToPtr(true),
		ChecksUseAdvertise: internal.BoolToPtr(false),
		EnableSSL:          internal.BoolToPtr(false),
		VerifySSL:          internal.BoolToPtr(false),
		ServerAutoJoin:     internal.BoolToPtr(true),
		ClientAutoJoin:     internal.BoolToPtr(true),
		Timeout:            5 * time.Second,
	}
}

// Merge merges two Consul Configurations together.
func (a *ConsulConfig) Merge(b *ConsulConfig) *ConsulConfig {
	result := a.Copy()

	if b.ServerServiceName != "" {
		result.ServerServiceName = b.ServerServiceName
	}
	if b.ClientServiceName != "" {
		result.ClientServiceName = b.ClientServiceName
	}
	if b.AutoAdvertise != nil {
		result.AutoAdvertise = internal.BoolToPtr(*b.AutoAdvertise)
	}
	if b.Addr != "" {
		result.Addr = b.Addr
	}
	if b.Timeout != 0 {
		result.Timeout = b.Timeout
	}
	if b.Token != "" {
		result.Token = b.Token
	}
	if b.Auth != "" {
		result.Auth = b.Auth
	}
	if b.EnableSSL != nil {
		result.EnableSSL = internal.BoolToPtr(*b.EnableSSL)
	}
	if b.VerifySSL != nil {
		result.VerifySSL = internal.BoolToPtr(*b.VerifySSL)
	}
	if b.CAFile != "" {
		result.CAFile = b.CAFile
	}
	if b.CertFile != "" {
		result.CertFile = b.CertFile
	}
	if b.KeyFile != "" {
		result.KeyFile = b.KeyFile
	}
	if b.ServerAutoJoin != nil {
		result.ServerAutoJoin = internal.BoolToPtr(*b.ServerAutoJoin)
	}
	if b.ClientAutoJoin != nil {
		result.ClientAutoJoin = internal.BoolToPtr(*b.ClientAutoJoin)
	}
	if b.ChecksUseAdvertise != nil {
		result.ChecksUseAdvertise = internal.BoolToPtr(*b.ChecksUseAdvertise)
	}
	return result
}

// ApiConfig() returns a usable Consul config that can be passed directly to
// hashicorp/consul/api.  NOTE: datacenter is not set
func (c *ConsulConfig) ApiConfig() (*consul.Config, error) {
	config := consul.DefaultConfig()
	if c.Addr != "" {
		config.Address = c.Addr
	}
	if c.Token != "" {
		config.Token = c.Token
	}
	if c.Timeout != 0 {
		config.HttpClient.Timeout = c.Timeout
	}
	if c.Auth != "" {
		var username, password string
		if strings.Contains(c.Auth, ":") {
			split := strings.SplitN(c.Auth, ":", 2)
			username = split[0]
			password = split[1]
		} else {
			username = c.Auth
		}

		config.HttpAuth = &consul.HttpBasicAuth{
			Username: username,
			Password: password,
		}
	}
	if c.EnableSSL != nil && *c.EnableSSL {
		config.Scheme = "https"
		tlsConfig := consul.TLSConfig{
			Address:  config.Address,
			CAFile:   c.CAFile,
			CertFile: c.CertFile,
			KeyFile:  c.KeyFile,
		}
		if c.VerifySSL != nil {
			tlsConfig.InsecureSkipVerify = !*c.VerifySSL
		}

		tlsClientCfg, err := consul.SetupTLSConfig(&tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating tls client config for consul: %v", err)
		}
		config.HttpClient.Transport = &http.Transport{
			TLSClientConfig: tlsClientCfg,
		}
	}

	return config, nil
}

// Copy returns a copy of this Consul config.
func (c *ConsulConfig) Copy() *ConsulConfig {
	if c == nil {
		return nil
	}

	nc := new(ConsulConfig)
	*nc = *c

	// Copy the bools
	if nc.AutoAdvertise != nil {
		nc.AutoAdvertise = internal.BoolToPtr(*nc.AutoAdvertise)
	}
	if nc.ChecksUseAdvertise != nil {
		nc.ChecksUseAdvertise = internal.BoolToPtr(*nc.ChecksUseAdvertise)
	}
	if nc.EnableSSL != nil {
		nc.EnableSSL = internal.BoolToPtr(*nc.EnableSSL)
	}
	if nc.VerifySSL != nil {
		nc.VerifySSL = internal.BoolToPtr(*nc.VerifySSL)
	}
	if nc.ServerAutoJoin != nil {
		nc.ServerAutoJoin = internal.BoolToPtr(*nc.ServerAutoJoin)
	}
	if nc.ClientAutoJoin != nil {
		nc.ClientAutoJoin = internal.BoolToPtr(*nc.ClientAutoJoin)
	}

	return nc
}
