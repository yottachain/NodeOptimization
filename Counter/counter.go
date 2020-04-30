package counter

import (
	"context"
	"sync"
)

//type NodeCountRow map[int]int64
type NodeCountRow struct {
	SuccTimes int64
	FailTimes int64
	AvgDelayTimes int64
	Score int64
}

type CountRowTable map[string]NodeCountRow

type InRow struct {
	ID string
	Status int
	DelayTimes int64
}

// 计数器，用于统计各种状态次数，并归档历史数据
type Counter struct {
	NodeCountTable sync.Map
	InQueue chan InRow
	mtx sync.RWMutex
	//// 归档队列
	//ARCountTable map[string][]NodeCountRow
	//// 归档间隔
	//ARInterval time.Duration
	//// 最大归档长度
	//ARMax int
}

func NewCounter(inQueueLen uint) *Counter {
	o := Counter{
		sync.Map{},
		// 同时最大处理请求次数
		make(chan InRow,inQueueLen),
		sync.RWMutex{},
		//make(map[string][]NodeCountRow),
		//// 默认5分钟归档一次
		//time.Minute*5,
		//// 默认最大归档55分钟
		//11,
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
	var nodeCountRow NodeCountRow
	counter.mtx.RLock()
	ac,_:=counter.NodeCountTable.LoadOrStore(row.ID, nodeCountRow)
	counter.mtx.RUnlock()
	nodeCountRow = ac.(NodeCountRow)

	totalTiems := nodeCountRow.SuccTimes + nodeCountRow.FailTimes
	if totalTiems == 0 {
		return
	}
	totalDelay := float64(nodeCountRow.AvgDelayTimes*(totalTiems))+ float64(row.DelayTimes)
	nodeCountRow.AvgDelayTimes = int64(totalDelay/float64(totalTiems))
	if row.Status == 0 {
		nodeCountRow.SuccTimes = nodeCountRow.SuccTimes + 1
	}else {
		nodeCountRow.FailTimes = nodeCountRow.FailTimes + 1
	}
	counter.NodeCountTable.Store(row.ID, nodeCountRow)
}

func (counter *Counter)Calc_score(ids ...string) {
	defScore := 1000
	cSW := 0.7
	cUW := 0.3
	cW := 0.8
	dW := 0.2
	connScore := float64(defScore)*cW
	delayScore := float64(defScore)*dW

	csScore := connScore*cSW
	cuScore := connScore*cUW

	var TotalDelays int64
	var TotalConnTimes int64
	for _,v:= range ids{
		counter.mtx.RLock()
		var nodeCountRow NodeCountRow
		ac,_:=counter.NodeCountTable.LoadOrStore(v, nodeCountRow)
		counter.mtx.RUnlock()
		nodeCountRow = ac.(NodeCountRow)
		TotalDelays = TotalDelays + nodeCountRow.AvgDelayTimes
		TotalConnTimes = TotalConnTimes + nodeCountRow.SuccTimes + nodeCountRow.FailTimes
	}
	allAvgDelayTimes := int64(float64(TotalDelays) / float64(len(ids)))
	for _,v:= range ids{
		counter.mtx.RLock()
		var nodeCountRow NodeCountRow
		ac,_:=counter.NodeCountTable.LoadOrStore(v, nodeCountRow)
		counter.mtx.RUnlock()
		nodeCountRow = ac.(NodeCountRow)
		connTimes := nodeCountRow.FailTimes + nodeCountRow.SuccTimes
		if connTimes == 0 {
			nodeCountRow.Score = int64(defScore)
			continue
		}
		cScore := csScore * float64(nodeCountRow.SuccTimes) / float64(connTimes) + cuScore * float64(TotalConnTimes) / float64(connTimes)
		dScore := delayScore * float64(allAvgDelayTimes) / float64(nodeCountRow.AvgDelayTimes)
		nodeCountRow.Score = int64(cScore + dScore)
	}
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

	res := make(map[string]NodeCountRow)

	for _,v:= range ids{
		counter.mtx.RLock()
		var nodeCountRow NodeCountRow
		ac,_:=counter.NodeCountTable.LoadOrStore(v, nodeCountRow)
		counter.mtx.RUnlock()
		res[v]=ac.(NodeCountRow)
	}
	return res
}

func(counter *Counter)Reset(){
	counter.mtx.Lock()
	counter.NodeCountTable = sync.Map{}
	counter.mtx.Unlock()
}

