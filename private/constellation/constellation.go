package constellation

import (
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"
)

type Constellation struct {
	node *Client
	c    *cache.Cache
}

func (g *Constellation) Send(data []byte, from string, to []string) (out []byte, err error) {
	out, err = g.node.SendPayload(data, from, to)
	if err != nil {
		return nil, err
	}
	g.c.Set(string(out), data, cache.DefaultExpiration)
	return out, nil
}

func (g *Constellation) Receive(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	// Ignore this error since not being a recipient of
	// a payload isn't an error.
	// TODO: Return an error if it's anything OTHER than
	// 'you are not a recipient.'
	dataStr := string(data)
	x, found := g.c.Get(dataStr)
	if found {
		return x.([]byte), nil
	}
	pl, _ := g.node.ReceivePayload(data)
	g.c.Set(dataStr, pl, cache.DefaultExpiration)
	return pl, nil
}

func New(configPath string) (*Constellation, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// start the private transaction node
	if config.NodeAutostart {
		launchNode(configPath, config.NodeCommand)
	}

	// build the client
	nodeClient, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	// ensure node is up
	err = nodeClient.UpCheck()
	if err != nil {
		return nil, err
	}

	return &Constellation{
		node: nodeClient,
		c:    cache.New(5*time.Minute, 5*time.Minute),
	}, nil
}

func MustNew(configPath string) *Constellation {
	g, err := New(configPath)
	if err != nil {
		panic(fmt.Sprintf("MustNew error: %v", err))
	}
	return g
}

func MaybeNew(configPath string) *Constellation {
	if configPath == "" {
		return nil
	}
	return MustNew(configPath)
}
