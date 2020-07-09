/*
@Time : 2019/5/4 13:48
@Author : yanKoo
@File : redis_data_sync
@Software: GoLand
@Description:
*/
package init_load

import (
	"database/sql"
	tg "bo-server/dao/group"
	tgc "bo-server/dao/group_cache"
	"bo-server/dao/pub"
	tu "bo-server/dao/user"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"sync"
)

type ConcurrentEngine struct {
	Scheduler   Scheduler
	WorkerCount int
}

type Scheduler interface {
	Submit(int32)
	ConfigureMasterWorkerChan(chan int32)
}

func (e ConcurrentEngine) Run() {
	in := make(chan int32)
	var wg sync.WaitGroup
	e.Scheduler.ConfigureMasterWorkerChan(in)

	for i := 0; i < e.WorkerCount; i++ {
		createWorker(in, &wg)
	}

	// 查找所有的用户id
	uIds, _ := tu.SelectAllUserId()
	for _, v := range uIds {
		//log.log.Debugf("# uid %d", id)
		wg.Add(1)
		go func(id int32) { e.Scheduler.Submit(id) }(v)
	}
	wg.Wait()
	logger.Debugf("**********************redis data sync done*****************************")
}

func createWorker(in chan int32, wg *sync.WaitGroup) {
	go func() {
		for {
			uId := <-in
			err := UserData(uId)
			if err != nil {
				//continue
			}
			err = GroupData(uId)
			if err != nil {
				//continue
			}
			wg.Done()
		}
	}()
}

func DataInit() {
	uIds, _ := tu.SelectAllUserId()
	for _, v := range uIds {
		_ = UserData(v)
		_ = GroupData(v)

	}
}

func UserData(uId int32) error {
	// 根据用户id去获取每一位的信息，放进缓存
	res, err := tu.SelectUserByKey(int(uId))
	if err != nil && err != sql.ErrNoRows {
		logger.Debugf("UserData SelectUserByKey error : %s", err)
		return err
	}

	if err := tuc.AddUserDataInCache(int32(res.Id), []interface{}{
		pub.USER_Id, res.Id,
		pub.IMEI, res.IMei,
		pub.BATTERY, 0, // 系统启动默认电量为0
		pub.ACCOUNT_ID, res.AccountId,
		pub.USER_NAME, res.UserName,
		pub.NICK_NAME, res.NickName,
		pub.USER_TYPE, res.UserType,
		pub.LOCK_GID, res.LockGroupId,
		pub.DEVICE_TYPE, res.DeviceType,
		pub.ONLINE, pub.USER_OFFLINE, // 加载数据默认全部离线
	}, cache.GetRedisClient()); err != nil {
		logger.Error("Add user information to cache with error: ", err)
	}
	logger.Debugln("Add User Info into cache done")
	return nil
}

func GroupData(uid int32) error {
	gl, _, err := tg.GetGroupListFromDB(int32(uid), db.DBHandler)
	if err != nil {
		return err
	}
	logger.Debugln("GroupData GetGroupListFromDB start update redis")
	// 新增到缓存 更新两个地方，首先，每个组的信息要更新，就是group data，记录了群组的id和名字
	if err := tgc.AddGroupInCache(gl, cache.GetRedisClient()); err != nil {
		return err
	}

	// 其次更新一个userSet  就是一个组里有哪些用户
	if err := tuc.AddUserInGroupToCache(gl, cache.GetRedisClient()); err != nil {
		return err
	}

	// 每个用户的信息
	for _, g := range gl.GroupList {
		for _, u := range g.UserList {
			logger.Debugf("%+v", u)
			if err := tuc.AddUserDataInCache(u.Id, []interface{}{
				pub.USER_Id, u.Id,
				pub.IMEI, u.Imei,
				pub.NICK_NAME, u.NickName,
				pub.ONLINE, u.Online,
				pub.LOCK_GID, u.LockGroupId,
			}, cache.GetRedisClient()); err != nil {
				logger.Error("GroupData Add user information to cache with error: ", err)
			}
		}
	}

	// 每一个群组拥有的成员
	for _, v := range gl.GroupList {
		if err := tgc.AddGroupCache(v.UserList, v, cache.GetRedisClient()); err != nil {
			return err
		}
	}
	return nil
}
