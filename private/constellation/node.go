package constellation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/tv42/httpunix"
)

func launchNode(workDir, nodeCommand string) (*exec.Cmd, error) {
	cmd := exec.Command(nodeCommand, workDir)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go io.Copy(os.Stderr, stderr)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	return cmd, nil
}

func unixTransport(socketPath string) *httpunix.Transport {
	t := &httpunix.Transport{
		DialTimeout:           1 * time.Second,
		RequestTimeout:        5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	t.RegisterLocation("c", socketPath)
	return t
}

func unixClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: unixTransport(socketPath),
	}
}

type SendRequest struct {
	Payload string   `json:"payload"`
	From    string   `json:"from"`
	To      []string `json:"to"`
}

type SendResponse struct {
	Key string `json:"key"`
}

type ReceiveRequest struct {
	Key string `json:"key"`
	To  string `json:"to"`
}

type ReceiveResponse struct {
	Payload string `json:"payload"`
}

type Client struct {
	httpClient   *http.Client
	publicKey    [32]byte
	b64PublicKey string
	baseURL      string
}

func (c *Client) do(path string, apiReq interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(apiReq)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.baseURL+path, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err == nil && res.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 status code: %+v", res)
	}
	return res, err
}

func (c *Client) UpCheck() error {
	res, err := c.httpClient.Get(c.baseURL + "upcheck")
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New("Constellation Node API did not respond to upcheck request")
	}
	return nil
}

func (c *Client) SendPayload(pl []byte, b64From string, b64To []string) ([]byte, error) {
	var from string
	if b64From == "" {
		from = c.b64PublicKey
	} else {
		from = b64From
	}
	req := &SendRequest{
		Payload: base64.StdEncoding.EncodeToString(pl),
		From:    from,
		To:      b64To,
	}
	res, err := c.do("send", req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	sres := new(SendResponse)
	err = json.NewDecoder(res.Body).Decode(sres)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(sres.Key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (c *Client) ReceivePayload(key []byte) ([]byte, error) {
	b64Key := base64.StdEncoding.EncodeToString(key)
	req := &ReceiveRequest{
		Key: b64Key,
		To:  c.b64PublicKey,
	}
	res, err := c.do("receive", req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	rres := new(ReceiveResponse)
	err = json.NewDecoder(res.Body).Decode(rres)
	if err != nil {
		return nil, err
	}
	pl, err := base64.StdEncoding.DecodeString(rres.Payload)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func NewClient(config *Config) (*Client, error) {
	// read public key
	b64PublicKey, err := ioutil.ReadFile(config.PublicKeys[0])
	if err != nil {
		return nil, err
	}

	// set httpClient; if socket is specified, used unix domain socket, else http
	var httpClient *http.Client
	var baseURL string
	if len(config.Socket) > 0 {
		httpClient = unixClient(config.Socket)
		baseURL = "http+unix://c/"
	} else {
		httpClient = http.DefaultClient
		baseURL = config.URL
		if baseURL[len(baseURL)-1:] != "/" {
			baseURL += "/"
		}
	}

	return &Client{
		httpClient:   httpClient,
		baseURL:      baseURL,
		b64PublicKey: string(b64PublicKey),
	}, nil
}
