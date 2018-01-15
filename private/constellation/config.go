package constellation

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Socket     string   `toml:"socket"`
	PublicKeys []string `toml:"publickeys"`

	NodeCommand   string `toml:"nodeCommand"`
	NodeAutostart bool   `toml:"nodeAutostart"`
	URL           string `toml:"url"`

	// Deprecated
	SocketPath    string `toml:"socketPath"`
	PublicKeyPath string `toml:"publicKeyPath"`
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := new(Config)
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, err
	}
	// Fall back to Constellation 0.0.1 config format if necessary
	if cfg.Socket == "" {
		cfg.Socket = cfg.SocketPath
	}
	if len(cfg.PublicKeys) == 0 {
		cfg.PublicKeys = append(cfg.PublicKeys, cfg.PublicKeyPath)
	}

	// Default the Node command to constellation-node
	if len(cfg.NodeCommand) == 0 {
		cfg.NodeCommand = "constellation-node"
	}
	return cfg, nil
}
