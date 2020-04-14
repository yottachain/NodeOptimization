package optimizer

import (
	counter "github.com/yottachain/NodeOptimization/Counter"
)

type Optmizer struct {
	*counter.Counter
	GetScore func(row counter.NodeCountRow) int64
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
	opt.PushRow(row)
}

func defaultGetScore(row counter.NodeCountRow) int64 {
	var w1 int64 = 2
	var w2 int64 = -1
	if (row[0]+row[1])==0 {
		return 500
	}
	return row[0] * w1 + row[1]* w2
}

func (opt *Optmizer)Get(ids ...string) ResRowList{

	res := make([]ResRow,0)

	for k,v:= range opt.CurrentCount(ids...) {
		res = append(res,ResRow{k,opt.GetScore(v)})
	}

	return res
}
func (opt *Optmizer)Get2(ids...string)[]string{
	ls:=opt.Get(ids...).Sort()
	res :=make([]string,0)
	for _,v :=range ls{
		res=append(res,v.ID)
	}
	return res
}

func New()*Optmizer{
	return &Optmizer{counter.NewCounter(4000),defaultGetScore}
}