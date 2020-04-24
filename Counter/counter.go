package counter

import (
	"context"
	"sync"
)

type NodeCountRow map[int]int64

type CountRowTable sync.Map

type InRow struct {
	ID string
	Status int
}

// 计数器，用于统计各种状态次数，并归档历史数据
type Counter struct {
	NodeCountTable map[string]NodeCountRow
	InQueue chan InRow
	//// 归档队列
	//ARCountTable map[string][]NodeCountRow
	//// 归档间隔
	//ARInterval time.Duration
	//// 最大归档长度
	//ARMax int

	sync.RWMutex
}

func NewCounter(inQueueLen uint) *Counter {
	o := Counter{
		make(map[string]NodeCountRow),
		// 同时最大处理请求次数
		make(chan InRow,inQueueLen),
		//make(map[string][]NodeCountRow),
		//// 默认5分钟归档一次
		//time.Minute*5,
		//// 默认最大归档55分钟
		//11,
		sync.RWMutex{},
	}
	return &o
}

func (counter *Counter)PushRow(inrow InRow) bool {
	select {
	case counter.InQueue<-inrow:
		return true
	default:
		return false
	}
}

func (counter *Counter)Run(ctx context.Context){
	// 在后台归档
	//go counter.ar(ctx)

	for {
		select {
		case <-ctx.Done():
			break
		default:
			counter.inOne()
		}
	}
}

func (counter *Counter)inOne(){

	// 根据传入操作状态给各个操作状态计次
	row:=<-counter.InQueue

	counter.Lock()
	defer counter.Unlock()
	nodeCountRow,ok:=counter.NodeCountTable[row.ID]
	if !ok {
		nodeCountRow = make(NodeCountRow)
	}
	nodeCountRow[row.Status]=nodeCountRow[row.Status]+1
	counter.NodeCountTable[row.ID]=nodeCountRow
}

//func (counter *Counter)ar(ctx context.Context){
//	for {
//		<-time.After(counter.ARInterval)
//		select {
//		case <-ctx.Done():
//			break
//		default:
//			counter.arOnce()
//		}
//	}
//}
//
//// 归档一次
//func (counter *Counter)arOnce(){
//
//	for id,arcountRow := range counter.ARCountTable {
//		func(){
//			counter.Lock(id)
//			defer counter.UnLock(id)
//
//			if arcountRow == nil {
//				counter.ARCountTable[id]=[]NodeCountRow{counter.NodeCountTable[id]}
//			} else if len(arcountRow)<counter.ARMax{
//				counter.ARCountTable[id]=append(arcountRow,counter.NodeCountTable[id])
//			} else {
//				counter.ARCountTable[id]=append(arcountRow[1:],counter.NodeCountTable[id])
//			}
//
//			delete(counter.NodeCountTable,id)
//		}()
//	}
//}

// 返回现有计次状态，id为空返回所有
func (counter *Counter)CurrentCount(ids ...string) map[string]NodeCountRow  {
	counter.Lock()
	defer counter.Unlock()

	if len(ids)==0{
		for k,_:=range counter.NodeCountTable{
			ids = append(ids, k)
		}
	}

	res := make(map[string]NodeCountRow)

	for _,v:= range ids{
		nodeCountRow,ok:=counter.NodeCountTable[v]
		if !ok {
			nodeCountRow=make(NodeCountRow)
		}
		res[v]=nodeCountRow
	}
	return res
}
func(counter *Counter)Reset(){
	counter.NodeCountTable = make(map[string]NodeCountRow)
}

