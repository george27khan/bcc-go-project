package detach_context

import (
	"context"
	"fmt"
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

func someFuncWithContext(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	fmt.Println("someFuncWithContext done!")
	return nil
}

//func main() {
//	ctx := context.WithValue(context.Background(), "key1", "value1")
//	ctx = context.WithValue(ctx, "key2", "value2")
//	//ctx, cancel := context.WithCancel(ctx)
//	//ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Second))
//	ctx, _ = context.WithTimeout(ctx, time.Second)
//
//	defer func() {
//		dctx := DetachContext(ctx)
//		dctx, cancel := context.WithTimeout(ctx, 2*time.Second)
//		defer cancel()
//		if err := someFuncWithContext(dctx); err != nil {
//			fmt.Println("someFuncWithContext not work")
//		}
//	}()
//	time.Sleep(2 * time.Second)
//	fmt.Println("ctx ", ctx.Err())
//}
