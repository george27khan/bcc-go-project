package detach_context

import (
	"context"
	"time"
)

/*
Написать DetachContext(ctx context.Context) context.Context , который прокидывает
все ключи родительского контекста, но не отменяется при отмене родительского контекста.
*/

type MyDetachContext struct {
	ctx context.Context
}

func (c MyDetachContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}
func (c MyDetachContext) Done() <-chan struct{} {
	return nil
}
func (c MyDetachContext) Err() error {
	return nil
}
func (c MyDetachContext) Value(key any) any {
	return c.ctx.Value(key)
}

func DetachContext(ctx context.Context) context.Context {
	return MyDetachContext{ctx: ctx}
}
