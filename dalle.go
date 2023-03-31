package dalle

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	buf "github.com/kklab-com/goth-bytebuf"
	"github.com/kklab-com/goth-kkutil/value"
)

const DefaultAPIEndpoint = "https://api.openai.com/v1/images/generations"

type ResponseFormat string
type ResponseSize string

const (
	ResponseFormatURL     = ResponseFormat("url")
	ResponseFormatB64JSON = ResponseFormat("b64_json")
	ResponseSize256x256   = ResponseSize("256x256")
	ResponseSize512x512   = ResponseSize("512x512")
	ResponseSize1024x1024 = ResponseSize("1024x1024")
)

type Opts struct {
	N              int            `json:"n,omitempty"`
	Size           ResponseSize   `json:"size,omitempty"`
	ResponseFormat ResponseFormat `json:"response_format,omitempty"`
	User           string         `json:"user,omitempty"`
}

type Request struct {
	Prompt string `json:"prompt"`
	Opts
}

type Data struct {
	Url     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}

func (d *Data) buf() (io.Reader, error) {
	buffer := buf.EmptyByteBuf()
	if d.Url != "" {
		if resp, err := http.Get(d.Url); err == nil {
			buffer.WriteReader(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return buffer, nil
			} else {
				return nil, fmt.Errorf(string(buffer.Bytes()))
			}
		} else {
			return nil, err
		}
	} else if d.B64JSON != "" {
		if decoded, err := base64.StdEncoding.DecodeString(d.B64JSON); err == nil {
			return buffer.WriteBytes(decoded), nil
		} else {
			return nil, err
		}
	}

	return buffer, nil

}

func (d *Data) Binary() (io.Reader, error) {
	return d.buf()
}

type Response struct {
	Created int    `json:"created"`
	Data    []Data `json:"data"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
	Type    string `json:"type"`
}

func (e *Error) Error() string {
	return value.JsonMarshal(e)
}

type Client struct {
	ApiEndpoint string
	apiKey      string
	opts        Opts
}

func NewClient(apiKey string) *Client {
	return NewClientWithOpts(apiKey, Opts{})
}

func NewClientWithOpts(apiKey string, opts Opts) *Client {
	if apiKey == "" {
		return nil
	}

	return &Client{apiKey: apiKey, ApiEndpoint: DefaultAPIEndpoint, opts: opts}
}

func (c *Client) Opts() *Opts {
	return &c.opts
}

func (c *Client) Request(prompt string) (*Response, error) {
	request, _ := http.NewRequest("POST", c.ApiEndpoint, bytes.NewBufferString(value.JsonMarshal(Request{
		Prompt: prompt,
		Opts:   c.opts,
	})))

	request.Header = http.Header{}
	request.Header.Set("authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	request.Header.Set("content-type", "application/json")
	request.Header.Set("user-agent", "curl/7.79.1")
	request.Header.Set("accept", "application/json")
	httpResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	byteBuf := buf.EmptyByteBuf().WriteReader(httpResponse.Body)
	response := &Response{}
	if err := json.Unmarshal(byteBuf.Bytes(), response); err != nil {
		return nil, err
	} else if response.Error != nil {
		return nil, response.Error
	} else {
		return response, nil
	}
}
