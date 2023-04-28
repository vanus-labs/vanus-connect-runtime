// Code generated by go-swagger; DO NOT EDIT.

package healthz

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// HealthzHandlerFunc turns a function with the right signature into a healthz handler
type HealthzHandlerFunc func(HealthzParams) middleware.Responder

// Handle executing the request and returning a response
func (fn HealthzHandlerFunc) Handle(params HealthzParams) middleware.Responder {
	return fn(params)
}

// HealthzHandler interface for that can handle valid healthz params
type HealthzHandler interface {
	Handle(HealthzParams) middleware.Responder
}

// NewHealthz creates a new http.Handler for the healthz operation
func NewHealthz(ctx *middleware.Context, handler HealthzHandler) *Healthz {
	return &Healthz{Context: ctx, Handler: handler}
}

/*
	Healthz swagger:route GET /healthz/ healthz healthz

healthz
*/
type Healthz struct {
	Context *middleware.Context
	Handler HealthzHandler
}

func (o *Healthz) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewHealthzParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
