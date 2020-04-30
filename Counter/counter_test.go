package counter

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCounter_CurrentCount(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	ctx,cancel :=context.WithCancel(context.Background())

	c := NewCounter(4000)

	go func() {
		for  {
			id:=rand.Intn(10)
			status:=rand.Intn(15)
			if status > id {
				status = 1
			} else {
				status = 0
			}
			c.PushRow(InRow{fmt.Sprintf("id-%d",id),status, 1000})
			time.After(time.Microsecond * 50)
		}
	}()
	go c.Run(ctx)


	//cancel()
	time.Sleep(3*time.Second)
	//c.CurrentCount().Print()
	defer cancel()
}
