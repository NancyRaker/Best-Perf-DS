/*
@Time : 2019/4/15 17:53
@Author : yanKoo
@File : message_test
@Software: GoLand
@Description:
*/
package msg

import (
	pb "web-api/proto/talk_cloud"
	"web-api/engine/db"
	"web-api/logger"
	"testing"
)

func testAddMsg(t *testing.T) {
	/*if err := AddMsg(&pb.ImMsgReqData{
		Id:333,
		ReceiverType:0,
		ReceiverId:334,
		ResourcePath:"HELLO WORLD",
		MsgType: 0,
	}, db.DBHandler); err != nil {
			t.Logf("Add Msg error: %v", err)
	}*/
	if err := AddMsg(&pb.ImMsgReqData{
		Id:           333,
		ReceiverType: 1, //group
		ReceiverId:   1,
		ResourcePath: "http://www.baidu.com/1.jpg",
		MsgType:      1,
	}, db.DBHandler); err != nil {
		t.Logf("Add Msg error: %v", err)
	}

}

func testAddMultiMsg(t *testing.T) {
	if err := AddMultiMsg(&pb.ImMsgReqData{
		Id:           333,
		ReceiverType: 2, //group
		ReceiverId:   1,
		ResourcePath: "http://www.baidu.com/1.jpg",
		MsgType:      1,
	}, []int32{335, 336, 337}, db.DBHandler); err != nil {
		t.Logf("Add Msg error: %v", err)
	}
}

func testGetMsg(t *testing.T) {
	if res, err := GetMsg(1503, int32(1), db.DBHandler); err != nil {
		logger.Debugln("Test Get offline msg error")
	} else {
		logger.Debugf("Offline msg: %v", res)
	}
}
