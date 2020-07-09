/*
@Time : 2019/4/10 14:16
@Author : yanKoo
@File : deviceTest
@Software: GoLand
@Description:
*/
package device

import (
	"bo-server/model"
	"testing"
)

func TestMultiUpdateDevice(t *testing.T) {
	devices := make([]*model.Device, 0)
	devices = append(devices, &model.Device{
		IMei: "123456789111111",
	})
	//_ = MultiUpdateDevice(&model.AccountDeviceTransReq{
	//	Devices: devices,
	//	Receiver: model.DeviceReceiver{
	//		AccountId: 31,
	//	},
	//})
}
