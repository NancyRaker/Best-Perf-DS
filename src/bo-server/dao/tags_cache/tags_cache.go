/*
@Time : 2019/10/14 18:00 
@Author : yanKoo
@File : tags_cache
@Software: GoLand
@Description:
*/
package tags_cache

import (
	pb "bo-server/api/proto"
	"bo-server/dao/pub"
	"bo-server/engine/cache"
	"bo-server/logger"
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
)

//插入从web导入的Tags信息 TODO 如果序列化中有一个失败？
func SaveTagsInfoToCache(TagsObjs *pb.TagsInfoReq) error {
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

	tagPair = append(tagPair, pub.MakeTagHashKey(TagsObjs.AccountId))
	for _, tag := range TagsObjs.Tags {
		tagsBytes, err = json.Marshal(tag)
		if err != nil {
			logger.Debugf("json marshal error: %s\n", err)
			return err
		}
		tagPair = append(tagPair, pub.MakeTagData(tag.Id))
		tagPair = append(tagPair, tagsBytes)
	}
	_, _ = rd.Do("HMSET", tagPair...)

	return nil
}

// 从缓存删除标签信息
func DelTagsInfoToCache(TagsObjs *pb.TagsInfoReq) error {
	var (
		rd redis.Conn
	)
	if TagsObjs == nil || TagsObjs.Tags == nil || len(TagsObjs.Tags) == 0 {
		return nil
	}

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()

	for _, tag := range TagsObjs.Tags {
		if _, err := rd.Do("HDEL", pub.MakeTagHashKey(tag.AccountId), pub.MakeTagData(tag.Id)); err != nil {
			logger.Errorf("GetTagsInfo From Cache redis.Bytes error : %+v", err)
			return err
		}
	}

	return nil
}

// 更新缓存数据, 因为可以修改mac地址这个原因，不得已只能增加bssid和id这个键值对 TODO 暂时这么来吧
func UpdateTagsInfoToCache(TagsObjs *pb.TagsInfoReq) error {
	var (
		rd        redis.Conn
		TagsBytes []byte
		err       error
	)

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()

	for _, tag := range TagsObjs.Tags {
		// 1. 根据调度员id和标签的id找到对应的Tag数据
		if TagsBytes, err = redis.Bytes(rd.Do("HGET", pub.MakeTagHashKey(tag.AccountId), pub.MakeTagData(tag.Id))); err != nil { //TODO 空针
			logger.Errorf("GetTagsInfo From Cache redis.Bytes error : %+v", err)
			return err
		}
		temp := &pb.Tag{}
		_ = json.Unmarshal(TagsBytes, temp)

		// 1.1 修改tag内容
		temp.TagName = tag.TagName
		temp.TagAddr = tag.TagAddr

		// 2. 序列化，重新保存到缓存
		TagsBytes, err := json.Marshal(tag)
		if err != nil {
			logger.Debugf("json marshal error: %s\n", err)
			return err
		}
		if _, err = rd.Do("HSET", pub.MakeTagHashKey(tag.AccountId), pub.MakeTagData(tag.Id), TagsBytes); err != nil {
			logger.Debugf("Save TagsInfo to cache error: %s\n", err)
			return err
		}
	}

	return nil
}

// 根据account id查询Tags信息
func GetTagsInfoByAccount(aId int32) ([]*pb.Tag, map[int32]*pb.Tag, error) {
	var (
		rd            redis.Conn
		err           error
		tags          = make([]*pb.Tag, 0)
		tagsSourceMap map[string]string
		tagsMap              = make(map[int32]*pb.Tag)
	)

	if rd = cache.GetRedisClient(); rd == nil {
		logger.Debugln("redis Conn is nil")
		return nil, nil, errors.New("redis Conn is nil")
	}
	defer rd.Close()

	logger.Debugf("will get account id: %d", aId)
	if tagsSourceMap, err = redis.StringMap(rd.Do("HGETALL", pub.MakeTagHashKey(aId))); err != nil { //TODO 空针
		logger.Debugf("GetTagsInfoByAccount From Cache redis.Bytes error : %+v", err)
		return nil, nil, err
	}
	for _, tagValue := range tagsSourceMap {
		tag := &pb.Tag{}
		if err = json.Unmarshal([]byte(tagValue), tag); err != nil {
			logger.Debugf("GetTagsInfo From Cache redis.Bytes Unmarshal error : %+v", err)
			return nil, nil, err
		}
		logger.Debugf("GetTagsInfo From Cache redis.Bytes Unmarshal tag: %+v", tag)
		tags = append(tags, tag)
		tagsMap[tag.Id] = tag
	}

	return tags, tagsMap, nil
}
