package memory

type MemoryPrivateTransactionManger struct {
	nodes map[string][]byte
}

// Send payload to list of to addresses
func (g *MemoryPrivateTransactionManger) Send(data []byte, from string, to []string) (out []byte, err error) {
	for i := 0; i < len(to); i++ {
		g.nodes[to[i]] = data
	}

	//Out is response code
	return
}

// Receive payload.
func (g *MemoryPrivateTransactionManger) Receive(data []byte) ([]byte, error) {

	pl := g.nodes[string(data)]

	return pl, nil
}

// MustNew Instantiates the in memory fake nodes
func MustNew(configPath string) *MemoryPrivateTransactionManger {
	return &MemoryPrivateTransactionManger{
		nodes: make(map[string][]byte),
	}
}
