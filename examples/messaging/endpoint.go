// Code generated by kun; DO NOT EDIT.
// github.com/RussellLuo/kun

package messaging

import (
	"context"

	"github.com/RussellLuo/kun/pkg/httpoption"
	"github.com/RussellLuo/validating/v3"
	"github.com/go-kit/kit/endpoint"
)

type GetMessageRequest struct {
	UserID    string `json:"-"`
	MessageID string `json:"-"`
}

// ValidateGetMessageRequest creates a validator for GetMessageRequest.
func ValidateGetMessageRequest(newSchema func(*GetMessageRequest) validating.Schema) httpoption.Validator {
	return httpoption.FuncValidator(func(value interface{}) error {
		req := value.(*GetMessageRequest)
		return httpoption.Validate(newSchema(req))
	})
}

type GetMessageResponse struct {
	Text string `json:"text"`
	Err  error  `json:"-"`
}

func (r *GetMessageResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *GetMessageResponse) Failed() error { return r.Err }

// MakeEndpointOfGetMessage creates the endpoint for s.GetMessage.
func MakeEndpointOfGetMessage(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*GetMessageRequest)
		text, err := s.GetMessage(
			ctx,
			req.UserID,
			req.MessageID,
		)
		return &GetMessageResponse{
			Text: text,
			Err:  err,
		}, nil
	}
}
