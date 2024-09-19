package router

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type Validatable interface {
	Validate() error
}

// type HandlerFunc[S any, U any] func(ctx *Context[S, U])
type HandlerFunc[S any, U any] func(ctx *Context[S, U], reqData *S) (*U, error)

type Context[S any, U any] struct {
	context.Context
	cancelFunc context.CancelFunc
	W          http.ResponseWriter
	Req        *http.Request
	handlers   []HandlerFunc[S, U]
	index      int
	ReqData    *S
	RespData   *U
	Err        error
	Log        *log.Logger
}

func CreateContext[S any, U any](w http.ResponseWriter, req *http.Request) *Context[S, U] {
	ctx, cancel := context.WithCancel(req.Context())
	return &Context[S, U]{
		Context:    ctx,
		cancelFunc: cancel,
		W:          w,
		Req:        req,
		index:      -1,
	}
}

// func (c *Context[S, U]) Next(reqData S) (*U, error) {
// 	c.index++
// 	var respData *U
// 	var err error
// 	for c.index < len(c.handlers) {
// 		//c.handlers[c.index](c)
// 		respData, err = c.handlers[c.index](c, &reqData)
// 		if err != nil {
// 			return nil, err
// 		}
// 		// if respData != nil {
// 		// 	return respData, nil
// 		// }
// 		c.index++
// 	}
// 	return respData, nil
// }

func (c *Context[S, U]) Next(reqData *S) (*U, error) {
	c.index++
	if c.index < len(c.handlers) {
		return c.handlers[c.index](c, reqData)
	}
	return nil, nil
}

func JSON[U any](w http.ResponseWriter, respData *U) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(respData)
}

func (c *Context[S, U]) Decode(req *http.Request, data *S) error {
	if c.ReqData != nil {
		return nil
	}

	//var reqData S
	decoder := json.NewDecoder(c.Req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return err
	}
	c.ReqData = data

	if validator, ok := any(data).(Validatable); ok {
		return validator.Validate()
	}
	return nil
}

// func Decode[S any](req *http.Request, data *S) error {
// 	decoder := json.NewDecoder(req.Body)
// 	decoder.DisallowUnknownFields()
// 	return decoder.Decode(data)
// }
