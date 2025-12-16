package middleware

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"log"
	"net/http"
	"runtime/debug"
)

type RequestId = struct{}

func AddRequestId(f nethttp.StrictHttpHandlerFunc, operationID string) nethttp.StrictHttpHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		rId := r.Header.Get("X-Request-Id")
		// добавление request-id в контекст и в ответ
		if rId == "" {
			rId = uuid.New().String()
			ctx = context.WithValue(ctx, RequestId{}, uuid.New())
		} else {
			ctx = context.WithValue(ctx, RequestId{}, rId)
		}

		// добавление request-id в ответ
		w.Header().Add("X-Request-Id", rId)

		return f(ctx, w, r, request)
	}
}

func PanicRecover(f nethttp.StrictHttpHandlerFunc, operationID string) nethttp.StrictHttpHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %s, %v,\n%s", operationID, rec, string(debug.Stack()))
				err = fmt.Errorf("%v", r)
			}
		}()
		return f(ctx, w, r, request)
	}
}
