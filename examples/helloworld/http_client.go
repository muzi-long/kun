// Code generated by kun; DO NOT EDIT.
// github.com/RussellLuo/kun

package helloworld

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/RussellLuo/kun/pkg/httpcodec"
)

type HTTPClient struct {
	codecs     httpcodec.Codecs
	httpClient *http.Client
	scheme     string
	host       string
	pathPrefix string
}

func NewHTTPClient(codecs httpcodec.Codecs, httpClient *http.Client, baseURL string) (*HTTPClient, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &HTTPClient{
		codecs:     codecs,
		httpClient: httpClient,
		scheme:     u.Scheme,
		host:       u.Host,
		pathPrefix: strings.TrimSuffix(u.Path, "/"),
	}, nil
}

func (c *HTTPClient) SayHello(ctx context.Context, name string) (message string, err error) {
	codec := c.codecs.EncodeDecoder("SayHello")

	path := "/messages"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return "", err
	}

	_req, err := http.NewRequestWithContext(ctx, "POST", u.String(), reqBodyReader)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return "", err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return "", err
	}

	respBody := &SayHelloResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return "", err
	}
	return respBody.Message, nil
}
