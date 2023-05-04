// Code generated by go-swagger; DO NOT EDIT.

package connector

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vanus-labs/vanus-connect-runtime/api/models"
)

// ChataiOKCode is the HTTP code returned for type ChataiOK
const ChataiOKCode int = 200

/*
ChataiOK OK

swagger:response chataiOK
*/
type ChataiOK struct {

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewChataiOK creates ChataiOK with default headers values
func NewChataiOK() *ChataiOK {

	return &ChataiOK{}
}

// WithPayload adds the payload to the chatai o k response
func (o *ChataiOK) WithPayload(payload *models.APIResponse) *ChataiOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the chatai o k response
func (o *ChataiOK) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ChataiOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}