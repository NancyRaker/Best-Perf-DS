/*
@Time : 2019/10/16 14:59 
@Author : yanKoo
@File : tag_task_cache
@Software: GoLand
@Description:
*/
package tag_task_cache

import (
	pb "bo-server/api/proto"
	"bo-server/dao/pub"
	"bo-server/engine/cache"
	"bo-server/logger"
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
)

// 保存标签任务到缓存
func SaveTagTasksListToCache(tasksListReq *pb.TagTasksListReq) error {
	var (
		rd        redis.Conn
		tagsBytes []byte
		err       error
		tagPair   []interface{}
	)

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()

	tagPair = append(tagPair, pub.MakeTagTaskHashKey(tasksListReq.AccountId))
	for _, taskList := range tasksListReq.TagTaskLists {
		tagsBytes, err = json.Marshal(taskList)
		if err != nil {
			logger.Debugf("json marshal error: %s\n", err)
			return err
		}
		tagPair = append(tagPair, pub.MakeTaskData(taskList.TagTaskId))
		tagPair = append(tagPair, tagsBytes)
	}
	_, _ = rd.Do("HMSET", tagPair...)

	return nil
}

// 从缓存删除标签任务
func DelTagTasksListToCache(tasksListReq *pb.TagTasksListReq) error {
	if tasksListReq == nil || tasksListReq.TagTaskLists == nil || len(tasksListReq.TagTaskLists) == 0 {
		return nil
	}

	var rd redis.Conn
	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()
	for _, tag := range tasksListReq.TagTaskLists {
		if _, err := rd.Do("HDEL", pub.MakeTagTaskHashKey(tasksListReq.AccountId), pub.MakeTaskData(tag.TagTaskId)); err != nil {
			logger.Errorf("GetTagsInfo From Cache redis.Bytes error : %+v", err)
			return err
		}
	}

	return nil
}

// 获取调度员名下所有任务
func QueryTagTask(accountId int32) ([]*pb.TagTaskListNode, error) {
	var (
		rd            redis.Conn
		err           error
		taskSourceMap map[string]string

		tagTasks = make([]*pb.TagTaskListNode, 0)
	)

	if rd = cache.GetRedisClient(); rd == nil {
		logger.Debugln("redis Conn is nil")
		return nil, errors.New("redis Conn is nil")
	}
	defer rd.Close()

	logger.Debugf("will get account id: %d", accountId)
	if taskSourceMap, err = redis.StringMap(rd.Do("HGETALL", pub.MakeTagTaskHashKey(accountId))); err != nil { //TODO 空针
		logger.Debugf("QueryTagTask From Cache redis.Bytes error : %+v", err)
		return nil, nil
	}
	for _, tagValue := range taskSourceMap {
		node := &pb.TagTaskListNode{}
		if err = json.Unmarshal([]byte(tagValue), node); err != nil {
			logger.Debugf("QueryTagTask From Cache redis.Bytes Unmarshal error : %+v", err)
			return nil, err
		}
		logger.Debugf("QueryTagTask From Cache redis.Bytes Unmarshal node: %+v", node)
		tagTasks = append(tagTasks, node)
	}

	return tagTasks, nil
}

func QuerySingleTagTask(accountId int32, taskId int32) (*pb.TagTaskListNode, error) {
	var (
		rd  redis.Conn
		tagTask = &pb.TagTaskListNode{}
	)

	if rd = cache.GetRedisClient(); rd == nil {
		logger.Debugln("redis Conn is nil")
		return nil, errors.New("redis Conn is nil")
	}
	defer rd.Close()

	logger.Debugf("will get account id: %d", accountId)
	if taskSource, err := redis.Bytes(rd.Do("HGET", pub.MakeTagTaskHashKey(accountId), taskId)); err != nil { //TODO 空针
		logger.Debugf("QueryTagTask From Cache redis.Bytes error : %+v", err)
		return nil, err
	} else {
		if err = json.Unmarshal([]byte(taskSource), tagTask); err != nil {
			logger.Debugf("QueryTagTask From Cache redis.Bytes Unmarshal error : %+v", err)
			return nil, err
		}
		logger.Debugf("QueryTagTask From Cache redis.Bytes Unmarshal node: %+v", tagTask)
	}
	return tagTask, nil
}
