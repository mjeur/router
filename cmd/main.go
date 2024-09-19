package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mjeur/router"
)

type CalculateReq struct {
	FirstNumber   int    `json:"first_number"`
	SecondNumber  int    `json:"second_number"`
	OperationName string `json:"operation_name"`
}

func (req *CalculateReq) Validate() error {
	if req.FirstNumber < 0 {
		return errors.New("first_number must be non-negative")
	}
	if req.SecondNumber < 0 {
		return errors.New("second_number must be non-negative")
	}
	if req.OperationName != "plus" && req.OperationName != "multiply" {
		return errors.New("invalid operation_name")
	}
	return nil
}

type CalculateResp struct {
	Result int `json:"result"`
}

func main() {
	fmt.Println(">>>")
	logger := log.New(os.Stdout, "[LOG] ", log.Lshortfile)
	r := router.New(logger)

	//router.POST("/calc", loggingMiddleware[CalculateReq, CalculateResp], calcHandler)
	router.POST(r, "/calc", authMiddleware, loggingMiddleware, handler)
	logger.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatal(err)
	}
}

func authMiddleware[S any, U any](ctx *router.Context[S, U], reqData *S) (*U, error) {
	token := ctx.Req.Header.Get("Authorization")
	if !validateToken(token) {
		return nil, errors.New("unauthorized")
	}
	return ctx.Next(reqData)
}

func validateToken(token string) bool {
	return token == "admin"
}

func loggingMiddleware[S any, U any](ctx *router.Context[S, U], reqData *S) (*U, error) {
	r := ctx.Req
	ctx.Log.Printf("Received %s request for %s", r.Method, r.URL.Path)
	return ctx.Next(ctx.ReqData)
}

func handler(ctx *router.Context[CalculateReq, CalculateResp], reqData *CalculateReq) (*CalculateResp, error) {
	if err := ctx.Decode(ctx.Req, reqData); err != nil {
		ctx.Err = err
		return nil, err
	}

	//reqData := ctx.ReqData

	if reqData.FirstNumber < 0 {
		ctx.Err = errors.New("first_number must be non-negative")
		return nil, ctx.Err
	}
	if reqData.SecondNumber < 0 {
		ctx.Err = errors.New("second_number must be non-negative")
		return nil, ctx.Err
	}
	if reqData.OperationName != "plus" && reqData.OperationName != "multiply" {
		ctx.Err = errors.New("invalid operation_name")
		return nil, ctx.Err
	}

	var result int
	switch reqData.OperationName {
	case "plus":
		result = reqData.FirstNumber + reqData.SecondNumber
	case "multiply":
		result = reqData.FirstNumber * reqData.SecondNumber
	}

	ctx.RespData = &CalculateResp{Result: result}
	return &CalculateResp{Result: result}, nil
}
