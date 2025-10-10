package httpx

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func NewClient(c *http.Client) Client {
	return Client{cli: c}
}

// Client http 客户端。
type Client struct {
	cli *http.Client
}

// GetJSON GET request and response JSON
func (c Client) GetJSON(ctx context.Context, rawURL string, header http.Header, result any) error {
	resp, err := c.sendJSON(ctx, http.MethodGet, rawURL, header, nil, true)
	if err != nil {
		return err
	}

	return c.unmarshalJSON(resp.Body, result)
}

func (c Client) SendJSON(ctx context.Context, method, rawURL string, header http.Header, body, result any) error {
	if method == "" {
		method = http.MethodPost
	}

	resp, err := c.sendJSON(ctx, method, rawURL, header, body, false)
	if err != nil {
		return err
	}

	return c.unmarshalJSON(resp.Body, result)
}

func (c Client) SendURLEncoded(ctx context.Context, method, rawURL string, header http.Header, body url.Values, result any) error {
	if method == "" {
		method = http.MethodPost
	}
	if header == nil {
		header = make(http.Header, 4)
	}
	header.Set("Content-Type", "application/x-www-form-urlencoded")

	var r io.Reader
	if len(body) != 0 {
		raw := body.Encode()
		r = strings.NewReader(raw)
	}
	req, err := c.newRequest(ctx, method, rawURL, header, r)
	if err != nil {
		return err
	}

	resp, err := c.send(req)
	if err != nil {
		return err
	}

	return c.unmarshalJSON(resp.Body, result)
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	return c.send(req)
}

func (c Client) sendJSON(ctx context.Context, method, rawURL string, header http.Header, body any, nobody bool) (*http.Response, error) {
	if header == nil {
		header = make(http.Header, 4)
	}
	header.Set("Accept", "application/json")

	var r io.Reader
	if !nobody {
		rd, err := c.marshalJSON(body)
		if err != nil {
			return nil, err
		}
		r = rd
		header.Set("Content-Type", "application/json; charset=utf-8")
	}

	req, err := c.newRequest(ctx, method, rawURL, header, r)
	if err != nil {
		return nil, err
	}

	return c.send(req)
}

func (Client) newRequest(ctx context.Context, method, rawURL string, header http.Header, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if len(header) != 0 {
		req.Header = header
	}

	return req, nil
}

func (c Client) send(req *http.Request) (*http.Response, error) {
	h := req.Header
	if host := h.Get("Host"); host != "" {
		req.Host = host
	}
	if h.Get("Accept") == "" {
		h.Set("Accept", "*/*")
	}
	if h.Get("User-Agent") == "" {
		chrome141 := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"
		h.Set("User-Agent", chrome141)
	}
	if h.Get("Accept-Language") == "" {
		h.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	}

	resp, err := c.getClient().Do(req)
	if err != nil {
		return nil, err
	}

	code := resp.StatusCode
	rem := code / 100
	if rem == 2 || rem == 3 {
		return resp, nil
	}

	e := &Error{
		Code:    code,
		Header:  resp.Header,
		Request: req,
	}
	buf := make([]byte, 1024)
	n, _ := io.ReadFull(resp.Body, buf)
	_ = resp.Body.Close()
	e.Body = buf[:n]

	return nil, e
}

func (c Client) marshalJSON(v any) (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := json.MarshalWrite(buf, v); err != nil {
		return nil, err
	}

	return buf, nil
}

func (c Client) getClient() *http.Client {
	if cli := c.cli; cli != nil {
		return cli
	}
	return http.DefaultClient
}

func (c Client) unmarshalJSON(rc io.ReadCloser, result any) error {
	//goland:noinspection GoUnhandledErrorResult
	defer rc.Close()
	if rc == http.NoBody {
		return nil
	}
	if result == nil {
		_, err := io.Copy(io.Discard, rc)
		return err
	}

	return json.UnmarshalRead(rc, result)
}
