/*
@Time : 2019/10/15 10:01 
@Author : yanKoo
@File : tags_test
@Software: GoLand
@Description:
*/
package tags

import (
	"bo-server/api/proto"
	"testing"
)

func testSaveTagsInfos(t *testing.T) {
	var tags []*talk_cloud.Tag
	tags = append(tags, &talk_cloud.Tag{TagAddr:"一楼", TagName:"小白", AccountId:115})
	tags = append(tags, &talk_cloud.Tag{TagAddr:"二楼", TagName:"小红", AccountId:115})
	res := SaveTagsInfo(&talk_cloud.TagsInfoReq{
		Tags:tags,
	})
	t.Log(res)
}

func testUpdateTagsInfoP(t *testing.T) {
	var tags []*talk_cloud.Tag
	tags = append(tags, &talk_cloud.Tag{Id :1, TagAddr:"二+1楼", TagName:"小蓝", AccountId:115})
	res := UpdateTagsInfo(&talk_cloud.TagsInfoReq{
		Tags:tags,
	})
	t.Log(res)
}

func testDelTagsInfo(t *testing.T) {
	var tags []*talk_cloud.Tag
	tags = append(tags, &talk_cloud.Tag{Id :1, TagAddr:"二+1楼", TagName:"小蓝", AccountId:115})
	res := DelTagsInfo(&talk_cloud.TagsInfoReq{
		Tags:tags,
	})
	t.Log(res)
}