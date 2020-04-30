package optimizer

import (
	"context"
	"fmt"
	counter "github.com/yottachain/NodeOptimization/Counter"
	"time"
)

type Optmizer struct {
	counter *counter.Counter
	GetScore func(row counter.NodeCountRow) int64
	ResetInterval time.Duration
}

type ResRow struct {
	ID string
	Score int64
}
type ResRowList []ResRow

func (rrl ResRowList)Sort()ResRowList{

	for i:=0;i<len(rrl);i++{
		for j:=i;j<len(rrl);j++{
			if rrl[j].Score>rrl[i].Score {
				temp := rrl[i]
				rrl[i]=rrl[j]
				rrl[j]=temp
			}
		}
	}

	return rrl
}

func (opt *Optmizer)Feedback(row counter.InRow){
	opt.counter.PushRow(row)
}

func defaultGetScore(row counter.NodeCountRow) int64 {
	return row.Score
}

func (opt *Optmizer)Get(ids ...string) ResRowList{
	res := make([]ResRow,0)

	for k,v:= range opt.counter.CurrentCount(ids...) {
		res = append(res,ResRow{k,opt.GetScore(v)})
	}

	return res
}
func (opt *Optmizer)Get2(ids...string)[]string{
	st:=time.Now()
	defer func() {
		fmt.Println("获取优选列表",time.Now().Sub(st).Milliseconds())
	}()
	ls:=opt.Get(ids...).Sort()
	res :=make([]string,0)
	for _,v :=range ls{
		res=append(res,v.ID)
	}
	return res
}

func New()*Optmizer{
	return &Optmizer{counter.NewCounter(4000),defaultGetScore,time.Minute*30}
}

func(opt *Optmizer) Run(ctx context.Context){
	go func() {
		<-time.After(opt.ResetInterval)
		opt.Reset()
	}()

	opt.counter.Run(ctx)
}

func (opt *Optmizer)CurrentCount(ids ...string) map[string]counter.NodeCountRow  {
	return opt.counter.CurrentCount(ids...)
}

func (opt *Optmizer)Reset(){
	//ls:=opt.Get().Sort()
	opt.counter.Reset()
	//max:=len(ls)
	//for k,v:=range ls{
		//ir:=counter.InRow{v.ID,max-k}
		//opt.Feedback(ir)
	//}
}