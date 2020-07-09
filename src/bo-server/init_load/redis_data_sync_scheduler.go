/*
@Time : 2019/5/5 16:01
@Author : yanKoo
@File : redis_data_sync_scheduler
@Software: GoLand
@Description:
*/
package init_load

type SimpleScheduler struct {
	workerChan chan int32
}

func (s *SimpleScheduler) Submit(r int32) {
	s.workerChan <- r
}

func (s *SimpleScheduler) ConfigureMasterWorkerChan(c chan int32) {
	s.workerChan = c
}
