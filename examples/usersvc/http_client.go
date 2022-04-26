// Code generated by kun; DO NOT EDIT.
// github.com/RussellLuo/kun

package usersvc

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

func (c *HTTPClient) CreateUser(ctx context.Context, user User) (result User, err error) {
	codec := c.codecs.EncodeDecoder("CreateUser")

	path := "/users"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	q := u.Query()
	for _, v := range codec.EncodeRequestParam("user", user) {
		q.Add("name", v)
	}
	for _, v := range codec.EncodeRequestParam("user", user) {
		q.Add("age", v)
	}
	u.RawQuery = q.Encode()

	_req, err := http.NewRequestWithContext(ctx, "POST", u.String(), nil)
	if err != nil {
		return User{}, err
	}
	for _, v := range codec.EncodeRequestParam("user", user) {
		_req.Header.Add("X-Forwarded-For", v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return User{}, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return User{}, err
	}

	respBody := &CreateUserResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return User{}, err
	}
	return respBody.Result, nil
}
