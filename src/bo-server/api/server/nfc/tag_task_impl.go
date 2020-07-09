/*
@Time : 2019/10/16 14:38 
@Author : yanKoo
@File : web_tag_taskimpl
@Software: GoLand
@Description:标签任务相关操作
*/
package nfc

import (
	pb "bo-server/api/proto"
	"bo-server/dao/tag_task"
	"bo-server/dao/tag_task_cache"
	"bo-server/dao/tags_cache"
	"bo-server/logger"
	"context"
	"errors"
	"net/http"
)

// web 保存，修改，删除标签任务
func (wssu *NFCServiceServerImpl) PostTagTasksList(ctx context.Context, req *pb.TagTasksListReq) (*pb.TagTasksListResp, error) {
	logger.Debugf("receive data from client: %+v", req)
	var (
		res *pb.TagTasksListResp
		//errResp = &pb.TagTasksListResp{Res:&pb.Result{Msg:"process error, please try again later"}}
		err        error
		saveTask   int32 = 1
		updateTask int32 = 2
		deleteTask int32 = 3
	)
	// 1. TODO 校验数据
	if len(req.TagTaskLists) == 0 || req.TagTaskLists == nil {
		return &pb.TagTasksListResp{Res: &pb.Result{Code: http.StatusOK}}, nil
	}
	switch req.Ops {
	case saveTask:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = saveTaskTasksList(req)
		if err != nil {
			logger.Debugf("post save Tag info error: %+v", err)
			return nil, err
		}
	case updateTask:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = updateTaskTasksList(req)
		if err != nil {
			logger.Debugf("post del Tag info error: %+v", err)
			return nil, err
		}
	case deleteTask:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = delTagTasksList(req)
		if err != nil {
			logger.Debugf("post del Tag info error: %+v", err)
			return nil, err
		}

	}
	return res, nil
}

// 保存task信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func saveTaskTasksList(TagTasksList *pb.TagTasksListReq) (*pb.TagTasksListResp, error) {
	// 1. 保存到mysql
	if err := tag_task.SaveTagTasksList(TagTasksList); err != nil {
		logger.Errorf("tw.SaveTaskTasksList to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 添加到缓存
	if err := tag_task_cache.SaveTagTasksListToCache(TagTasksList); err != nil {
		logger.Errorf("twc.SaveTaskTasksListToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.TagTasksListResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func delTagTasksList(TagTasksList *pb.TagTasksListReq) (*pb.TagTasksListResp, error) {
	// 1. 从mysql删除
	if err := tag_task.DelTagTasksList(TagTasksList); err != nil {
		logger.Errorf("tw.DelTagTasksList to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 从缓存删除
	if err := tag_task_cache.DelTagTasksListToCache(TagTasksList); err != nil {
		logger.Errorf("twc.DelTagTasksListToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.TagTasksListResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 更新task信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func updateTaskTasksList(TagTasksList *pb.TagTasksListReq) (*pb.TagTasksListResp, error) {
	/*// 1. 更新mysql
	if err := tt.UpdateTaskTasksList(TagTasksList); err != nil {
		logger.Errorf("tw.SaveTaskTasksList to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 更新缓存
	if err := tTc.UpdateTaskTasksListToCache(TagTasksList); err != nil {
		logger.Errorf("twc.SaveTaskTasksListToCache to redis fail with error : %+v", err)
		return nil, err
	}*/
	return &pb.TagTasksListResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 调度员获取task信息
func (wssu *NFCServiceServerImpl) GetTagTasksList(ctx context.Context, req *pb.GetTagTasksListReq) (*pb.GetTagTasksListResp, error) {
	/*	if req.AccountId < 0 {
			return nil, errors.New("account id is invalid")
		}

		var (
			Tags = make([]*pb.Tag, 0)
			err  error
		)

		if Tags, err = tTc.GetTagTasksListByAccount(req.AccountId); err != nil {
			return nil, err
		}

		if Tags != nil && len(Tags) > 0 {
			tagsSorter{}.randomQuickSort(&Tags, 0, len(Tags))
		}
	*/
	return &pb.GetTagTasksListResp{}, nil
}

// 查询设备有多少任务需要执行
func (wssu *NFCServiceServerImpl) QueryTasksByDevice(ctx context.Context, req *pb.DeviceTasksReq) (*pb.DeviceTasksResp, error) {
	// 0. 参数校验

	// 1. 查询调度员名下所有的任务
	tasks, err := tag_task_cache.QueryTagTask(req.AccountId)
	if err != nil {
		return nil, errors.New("QueryTasksByDevice QueryTagTask err" + err.Error())
	}
	// 2. 查询调度员名下所有的tag标签
	_, tags, err := tags_cache.GetTagsInfoByAccount(req.AccountId)
	if err != nil {
		return nil, errors.New("QueryTasksByDevice GetTagsInfoByAccount err" + err.Error())
	}

	// 3. 组装web前端需要的json体
	var res = &pb.DeviceTasksResp{}
	for _, task := range tasks {
		var tagNames []string

		// 该设备在当前任务中需要打的标签点
		for _, taskNode := range task.TagTaskNodes {
			if taskNode.DeviceId == req.DeviceId {
				if tag, ok := tags[taskNode.TagId]; !ok {
					return nil, errors.New("can't find tag")
				} else {
					tagNames = append(tagNames, tag.TagName)
				}
			}
		}

		var node *pb.DeviceTaskNode
		if len(tagNames) != 0 { // 说明有设备
			// 该任务的一些任务信息
			node = &pb.DeviceTaskNode{
				TagTaskId: task.TagTaskId,
				TaskName:  task.TaskName,
				SaveTime:  task.SaveTime,
				TagList:   tagNames,
			}
			res.SingleDeviceTaskList = append(res.SingleDeviceTaskList, node)
		}
	}

	return res, nil
}

// 查询任务的详细信息
func (wssu *NFCServiceServerImpl) QueryTaskDetail(ctx context.Context, req *pb.TaskDetailReq) (*pb.TaskDetailResp, error) {
	// 0. 参数校验
	logger.Debugf("QueryTaskDetail param: %+v", req)
	// 1. 查询调度员名某个task_id
	task, err := tag_task_cache.QuerySingleTagTask(req.AccountId, req.TaskId)
	if err != nil {
		return nil, errors.New("QueryTaskDetail QuerySingleTagTask err" + err.Error())
	}
	// 2. 查询调度员名下所有的tag标签
	_, tagMap, err := tags_cache.GetTagsInfoByAccount(req.AccountId)
	if err != nil {
		return nil, errors.New("QueryTaskDetail GetTagsInfoByAccount err" + err.Error())
	}

	var res = &pb.TaskDetailResp{TaskDetail: &pb.TaskDetail{}}
	if task.TagTaskNodes != nil {
		for _, taskNode := range task.TagTaskNodes {
			res.TaskDetail.TagNodes = append(res.TaskDetail.TagNodes, &pb.TagNode{
				Id:             taskNode.Id,
				TagName:        tagMap[taskNode.TagId].TagName,
				OrderStartTime: taskNode.OrderStartTime,
				OrderEndTime:   taskNode.OrderEndTime,
			})
		}
	}

	return res, nil
}
