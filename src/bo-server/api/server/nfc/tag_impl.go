/*
@Time : 2019/10/14 17:33 
@Author : yanKoo
@File : web_tag_impl
@Software: GoLand
@Description:
*/
package nfc

import (
	pb "bo-server/api/proto"
	tt "bo-server/dao/tags"
	tTc "bo-server/dao/tags_cache"
	"bo-server/logger"
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"
)
type NFCServiceServerImpl struct {
}
// web 保存，修改，删除Tag信息
func (nssu *NFCServiceServerImpl) PostTagsInfo(ctx context.Context, req *pb.TagsInfoReq) (*pb.TagsInfoResp, error) {
	logger.Debugf("receive data from client: %+v", req)
	var (
		res *pb.TagsInfoResp
		//errResp = &pb.TagsInfoResp{Res:&pb.Result{Msg:"process error, please try again later"}}
		err error
		saveTag   int32 = 1
		updateTag int32 = 2
		deleteTag int32 = 3
	)
	// 1. TODO 校验数据
	if len(req.Tags) == 0 || req.Tags == nil {
		return &pb.TagsInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
	}
	switch req.Ops {
	case saveTag:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = saveTagsInfo(req)
		if err != nil {
			logger.Debugf("post save Tag info error: %+v", err)
			return nil, err
		}
	case updateTag:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = updateTagsInfo(req)
		if err != nil {
			logger.Debugf("post del Tag info error: %+v", err)
			return nil, err
		}
	case deleteTag:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = delTagsInfo(req)
		if err != nil {
			logger.Debugf("post del Tag info error: %+v", err)
			return nil, err
		}

	}
	return res, nil
}

// 更新Tag信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func updateTagsInfo(TagsInfo *pb.TagsInfoReq) (*pb.TagsInfoResp, error) {
	// 1. 更新mysql
	if err := tt.UpdateTagsInfo(TagsInfo); err != nil {
		logger.Errorf("tw.SaveTagsInfo to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 更新缓存
	if err := tTc.UpdateTagsInfoToCache(TagsInfo); err != nil {
		logger.Errorf("twc.SaveTagsInfoToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.TagsInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 保存Tag信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func saveTagsInfo(TagsInfo *pb.TagsInfoReq) (*pb.TagsInfoResp, error) {
	// 1. 保存到mysql
	if err := tt.SaveTagsInfo(TagsInfo); err != nil {
		logger.Errorf("tw.SaveTagsInfo to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 添加到缓存
	if err := tTc.SaveTagsInfoToCache(TagsInfo); err != nil {
		logger.Errorf("twc.SaveTagsInfoToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.TagsInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func delTagsInfo(TagsInfo *pb.TagsInfoReq) (*pb.TagsInfoResp, error) {
	// 1. 从mysql删除
	if err := tt.DelTagsInfo(TagsInfo); err != nil {
		logger.Errorf("tw.DelTagsInfo to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 从缓存删除
	if err := tTc.DelTagsInfoToCache(TagsInfo); err != nil {
		logger.Errorf("twc.DelTagsInfoToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.TagsInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 调度员获取Tag信息
func (wssu *NFCServiceServerImpl) GetTagsInfo(ctx context.Context, req *pb.GetTagsInfoReq) (*pb.GetTagsInfoResp, error) {
	if req.AccountId < 0 {
		return nil, errors.New("account id is invalid")
	}

	var (
		tags = make([]*pb.Tag, 0)
		err  error
	)

	if tags, _, err = tTc.GetTagsInfoByAccount(req.AccountId); err != nil {
		return nil, err
	}

	if tags != nil && len(tags) > 0 {
		tagsSorter{}.randomQuickSort(&tags, 0, len(tags))
	}

	return &pb.GetTagsInfoResp{Tags: tags}, nil
}

type tagsSorter struct {}

func (t tagsSorter)randomQuickSort(list *[]*pb.Tag, start, end int) {
	if end-start > 1 {
		// get the pivot
		mid := t.randomPartition(list, start, end)
		t.randomQuickSort(list, start, mid)
		t.randomQuickSort(list, mid+1, end)
	}
}

func (t tagsSorter)randomPartition(list *[]*pb.Tag, begin, end int) int {
	// 生成真随机数
	i := t.randInt(begin, end)
	// 下面这行是核心部分，随机选择主元， 如果没有此次交换，就是普通快排
	(*list)[i], (*list)[begin] = (*list)[begin], (*list)[i]
	return t.partition(list, begin, end)
}

func (t tagsSorter)partition(list *[]*pb.Tag, begin, end int) (i int) {
	cValue := (*list)[begin].ImportTimestamp
	i = begin
	for j := i + 1; j < end; j++ {
		if (*list)[j].ImportTimestamp > cValue {
			i++
			(*list)[j], (*list)[i] = (*list)[i], (*list)[j]
		}
	}
	(*list)[i], (*list)[begin] = (*list)[begin], (*list)[i]
	return i
}

// 真随机数
func (t tagsSorter)randInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}