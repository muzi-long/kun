package profilesvc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// Ensure that ServiceMock does implement Service.
var _ Service = &ServiceMock{}

type ServiceMock struct {
	PostProfileFunc   func(ctx context.Context, profile Profile) (err error)
	GetProfileFunc    func(ctx context.Context, id string) (profile Profile, err error)
	PutProfileFunc    func(ctx context.Context, id string, profile Profile) (err error)
	PatchProfileFunc  func(ctx context.Context, id string, profile Profile) (err error)
	DeleteProfileFunc func(ctx context.Context, id string) (err error)
	GetAddressesFunc  func(ctx context.Context, id string) (addresses []Address, err error)
	GetAddressFunc    func(ctx context.Context, profileID string, addressID string) (address Address, err error)
	PostAddressFunc   func(ctx context.Context, profileID string, address Address) (err error)
	DeleteAddressFunc func(ctx context.Context, profileID string, addressID string) (err error)
}

func (mock *ServiceMock) PostProfile(ctx context.Context, profile Profile) (err error) {
	if mock.PostProfileFunc == nil {
		panic("ServiceMock.PostProfileFunc: not implemented")
	}
	return mock.PostProfileFunc(ctx, profile)
}

func (mock *ServiceMock) GetProfile(ctx context.Context, id string) (profile Profile, err error) {
	if mock.GetProfileFunc == nil {
		panic("ServiceMock.GetProfileFunc: not implemented")
	}
	return mock.GetProfileFunc(ctx, id)
}

func (mock *ServiceMock) PutProfile(ctx context.Context, id string, profile Profile) (err error) {
	if mock.PutProfileFunc == nil {
		panic("ServiceMock.PutProfileFunc: not implemented")
	}
	return mock.PutProfileFunc(ctx, id, profile)
}

func (mock *ServiceMock) PatchProfile(ctx context.Context, id string, profile Profile) (err error) {
	if mock.PatchProfileFunc == nil {
		panic("ServiceMock.PatchProfileFunc: not implemented")
	}
	return mock.PatchProfileFunc(ctx, id, profile)
}

func (mock *ServiceMock) DeleteProfile(ctx context.Context, id string) (err error) {
	if mock.DeleteProfileFunc == nil {
		panic("ServiceMock.DeleteProfileFunc: not implemented")
	}
	return mock.DeleteProfileFunc(ctx, id)
}

func (mock *ServiceMock) GetAddresses(ctx context.Context, id string) (addresses []Address, err error) {
	if mock.GetAddressesFunc == nil {
		panic("ServiceMock.GetAddressesFunc: not implemented")
	}
	return mock.GetAddressesFunc(ctx, id)
}

func (mock *ServiceMock) GetAddress(ctx context.Context, profileID string, addressID string) (address Address, err error) {
	if mock.GetAddressFunc == nil {
		panic("ServiceMock.GetAddressFunc: not implemented")
	}
	return mock.GetAddressFunc(ctx, profileID, addressID)

}
func (mock *ServiceMock) PostAddress(ctx context.Context, profileID string, address Address) (err error) {
	if mock.PostAddressFunc == nil {
		panic("ServiceMock.PostAddressFunc: not implemented")
	}
	return mock.PostAddressFunc(ctx, profileID, address)

}
func (mock *ServiceMock) DeleteAddress(ctx context.Context, profileID string, addressID string) (err error) {
	if mock.DeleteAddressFunc == nil {
		panic("ServiceMock.DeleteAddressFunc: not implemented")
	}
	return mock.DeleteAddressFunc(ctx, profileID, addressID)
}

type (
	request struct {
		method string
		path   string
		body   string
	}

	response struct {
		statusCode int
		body       string
	}
)

func makeRequest(r request, handler http.Handler) *httptest.ResponseRecorder {
	var req *http.Request
	if r.body != "" {
		reqBody := strings.NewReader(r.body)
		req = httptest.NewRequest(r.method, r.path, reqBody)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	} else {
		req = httptest.NewRequest(r.method, r.path, nil)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	return w
}

func checkResponse(w *httptest.ResponseRecorder, want response) string {
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	gotStatusCode := resp.StatusCode
	if gotStatusCode != want.statusCode {
		return fmt.Sprintf("StatusCode: got (%d), want (%d)", gotStatusCode, want.statusCode)
	}

	gotContentType := resp.Header.Get("Content-Type")
	wantContentType := "application/json; charset=utf-8"
	if gotContentType != wantContentType {
		return fmt.Sprintf("ContentType: got (%q), want (%q)", gotContentType, wantContentType)
	}

	gotBody := string(body)
	if gotBody != want.body {
		return fmt.Sprintf("Body: got (%q), want (%q)", gotBody, want.body)
	}

	return ""
}

func TestHTTP_PostProfile(t *testing.T) {
	type in struct {
		profile Profile
	}
	type out struct {
		err error
	}

	requestCases := []struct {
		name    string
		request request
		wantIn  in
	}{
		{
			name: "request",
			request: request{
				method: "POST",
				path:   "/profiles",
				body:   `{"profile": {"id": "1234", "name": "kok"}}`,
			},
			wantIn: in{
				profile: Profile{
					ID:   "1234",
					Name: "kok",
				},
			},
		},
	}
	for _, c := range requestCases {
		t.Run(c.name, func(t *testing.T) {
			var gotIn in
			makeRequest(c.request, NewHTTPHandler(&ServiceMock{
				PostProfileFunc: func(ctx context.Context, profile Profile) (err error) {
					gotIn = in{
						profile: profile,
					}
					return nil
				},
			}))
			if !reflect.DeepEqual(gotIn, c.wantIn) {
				t.Fatalf("In: got (%v), want (%v)", gotIn, c.wantIn)
			}
		})
	}

	responseCases := []struct {
		name         string
		request      request
		out          out
		wantResponse response
	}{
		{
			name: "response-ok",
			request: request{
				method: "POST",
				path:   "/profiles",
				body:   `{}`,
			},
			out: out{
				err: nil,
			},
			wantResponse: response{
				statusCode: http.StatusOK,
				body:       `{}` + "\n",
			},
		},
		{
			name: "response-err",
			request: request{
				method: "POST",
				path:   "/profiles",
				body:   `{}`,
			},
			out: out{
				err: ErrAlreadyExists,
			},
			wantResponse: response{
				statusCode: http.StatusBadRequest,
				body:       `{"error":"already exists"}` + "\n",
			},
		},
	}
	for _, c := range responseCases {
		t.Run(c.name, func(t *testing.T) {
			w := makeRequest(c.request, NewHTTPHandler(&ServiceMock{
				PostProfileFunc: func(ctx context.Context, profile Profile) (err error) {
					return c.out.err
				},
			}))
			if errStr := checkResponse(w, c.wantResponse); errStr != "" {
				t.Fatal(errStr)
			}
		})
	}
}

func TestHTTP_GetProfile(t *testing.T) {
	type in struct {
		id string
	}
	type out struct {
		profile Profile
		err     error
	}

	requestCases := []struct {
		name    string
		request request
		wantIn  in
	}{
		{
			name: "request",
			request: request{
				method: "GET",
				path:   "/profiles/1234",
			},
			wantIn: in{
				id: "1234",
			},
		},
	}
	for _, c := range requestCases {
		t.Run(c.name, func(t *testing.T) {
			var gotIn in
			makeRequest(c.request, NewHTTPHandler(&ServiceMock{
				GetProfileFunc: func(ctx context.Context, id string) (profile Profile, err error) {
					gotIn = in{
						id: id,
					}
					return Profile{}, nil
				},
			}))
			if !reflect.DeepEqual(gotIn, c.wantIn) {
				t.Fatalf("In: got (%v), want (%v)", gotIn, c.wantIn)
			}
		})
	}

	responseCases := []struct {
		name         string
		request      request
		out          out
		wantResponse response
	}{
		{
			name: "response-ok",
			request: request{
				method: "GET",
				path:   "/profiles/1234",
			},
			out: out{
				profile: Profile{
					ID:   "1234",
					Name: "kok",
				},
				err: nil,
			},
			wantResponse: response{
				statusCode: http.StatusOK,
				body:       `{"profile":{"id":"1234","name":"kok"}}` + "\n",
			},
		},
		{
			name: "response-err",
			request: request{
				method: "GET",
				path:   "/profiles/1234",
			},
			out: out{
				profile: Profile{},
				err:     ErrNotFound,
			},
			wantResponse: response{
				statusCode: http.StatusNotFound,
				body:       `{"error":"not found"}` + "\n",
			},
		},
	}
	for _, c := range responseCases {
		t.Run(c.name, func(t *testing.T) {
			w := makeRequest(c.request, NewHTTPHandler(&ServiceMock{
				GetProfileFunc: func(ctx context.Context, id string) (profile Profile, err error) {
					return c.out.profile, c.out.err
				},
			}))
			if errStr := checkResponse(w, c.wantResponse); errStr != "" {
				t.Fatal(errStr)
			}
		})
	}
}

func TestHTTP_PutProfile(t *testing.T) {
	type in struct {
		id      string
		profile Profile
	}
	type out struct {
		err error
	}

	requestCases := []struct {
		name    string
		request request
		wantIn  in
	}{
		{
			name: "request",
			request: request{
				method: "PUT",
				path:   "/profiles/1234",
				body:   `{"profile": {"id": "5678", "name": "kok", "addresses": [{"id": "0", "location": "here"}]}}`,
			},
			wantIn: in{
				id: "1234",
				profile: Profile{
					ID:   "5678",
					Name: "kok",
					Addresses: []Address{
						{
							ID:       "0",
							Location: "here",
						},
					},
				},
			},
		},
	}
	for _, c := range requestCases {
		t.Run(c.name, func(t *testing.T) {
			var gotIn in
			makeRequest(c.request, NewHTTPHandler(&ServiceMock{
				PutProfileFunc: func(ctx context.Context, id string, profile Profile) (err error) {
					gotIn = in{
						id:      id,
						profile: profile,
					}
					return nil
				},
			}))
			if !reflect.DeepEqual(gotIn, c.wantIn) {
				t.Fatalf("In: got (%v), want (%v)", gotIn, c.wantIn)
			}
		})
	}

	responseCases := []struct {
		name         string
		request      request
		out          out
		wantResponse response
	}{
		{
			name: "response-ok",
			request: request{
				method: "PUT",
				path:   "/profiles/1234",
				body:   `{}`,
			},
			out: out{
				err: nil,
			},
			wantResponse: response{
				statusCode: http.StatusOK,
				body:       `{}` + "\n",
			},
		},
		{
			name: "response-err",
			request: request{
				method: "PUT",
				path:   "/profiles/1234",
				body:   `{}`,
			},
			out: out{
				err: ErrInconsistentIDs,
			},
			wantResponse: response{
				statusCode: http.StatusBadRequest,
				body:       `{"error":"inconsistent IDs"}` + "\n",
			},
		},
	}
	for _, c := range responseCases {
		t.Run(c.name, func(t *testing.T) {
			w := makeRequest(c.request, NewHTTPHandler(&ServiceMock{
				PutProfileFunc: func(ctx context.Context, id string, profile Profile) (err error) {
					return c.out.err
				},
			}))
			if errStr := checkResponse(w, c.wantResponse); errStr != "" {
				t.Fatal(errStr)
			}
		})
	}
}
