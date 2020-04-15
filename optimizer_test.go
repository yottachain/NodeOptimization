package optimizer

import (
	"context"
	"fmt"
	counter "github.com/yottachain/NodeOptimization/Counter"
	"math/rand"
	"testing"
	"time"
)

func TestOptmizer_Get(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	ctx,cancel :=context.WithCancel(context.Background())

	o:=New()

	go func() {
		for  {
			id:=rand.Intn(2000)
			status:=rand.Intn(2000)
			if status > id {
				status = 1
			} else {
				status = 0
			}
			o.Feedback(counter.InRow{fmt.Sprintf("id-%d",id),status})
			time.After(time.Microsecond * 50)
		}
	}()
	go o.Run(ctx)


	//cancel()
	time.Sleep(5*time.Second)
	ids := make([]string,0)
	for i:=0;i<2000;i++{
		ids=append(ids,fmt.Sprintf("id-%d",i))
	}
	rl:=o.Get(ids...)
	for  k,v:=range rl.Sort(){
		fmt.Println(k,v)
	}
	defer cancel()
}
