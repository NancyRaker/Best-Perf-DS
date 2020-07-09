/*
@Time : 2019/6/17 17:02 
@Author : yanKoo
@File : wifi_cache
@Software: GoLand
@Description:
*/
package wifi_cache

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	pb "bo-server/api/proto"
	"bo-server/engine/cache"
	"bo-server/logger"
	"strconv"
)

func makeWifiKey(accountId int32) string {
	return "acc:" + strconv.Itoa(int(accountId)) + ":wifi"
}

func makeUserWifiMap(accountId int32) string {
	return "usr:" + strconv.Itoa(int(accountId)) + ":bssId"
}

//插入从web导入的wifi信息 TODO 如果序列化中有一个失败？
func SaveWifiInfoToCache(wifiObjs *pb.WifiInfoReq) error {
	var (
		rd        redis.Conn
		wifiBytes []byte
		err       error
	)

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()

	_ = rd.Send("MULTI")
	for _, v := range wifiObjs.Wifis {
		wifiBytes, err = json.Marshal(v)
		if err != nil {
			logger.Debugf("json marshal error: %s\n", err)
			return err
		}
		//key := "wifi:"+ v.BssId +":info" // 直接bssid做key
		_ = rd.Send("SET", v.BssId, wifiBytes)
		_ = rd.Send("SET", makeUserWifiMap(v.Id), v.BssId)
		_ = rd.Send("SADD", makeWifiKey(wifiObjs.AccountId), wifiBytes)
	}

	if _, err = rd.Do("EXEC"); err != nil {
		logger.Debugf("Save WifiInfo to cache error: %s\n", err)
		return err
	}
	return nil
}

func CheckWifiInfoFromCache(bssId string) (bool, error) {
	var (
		rd      redis.Conn
		err     error
		ifExist bool
	)

	if rd = cache.GetRedisClient(); rd == nil {
		logger.Debugln("redis Conn is nil")
		return false, errors.New("redis Conn is nil")
	}
	defer rd.Close()
	if ifExist, err = redis.Bool(rd.Do("EXISTS", bssId)); err != nil {
		return false, err
	}

	return ifExist, nil
}

// 根据bssId查询wifi信息
func GetWifiInfoFromCache(bssId string) (*pb.Wifi, error) {
	var (
		rd       redis.Conn
		err      error
		wifi     = &pb.Wifi{}
		wifiByte []byte
	)

	if rd = cache.GetRedisClient(); rd == nil {
		logger.Debugln("redis Conn is nil")
		return nil, errors.New("redis Conn is nil")
	}
	defer rd.Close()

	if wifiByte, err = redis.Bytes(rd.Do("GET", bssId)); err != nil { //TODO 空针
		logger.Debugf("GetWifiInfo From Cache redis.Bytes error : %+v", err)
		return nil, err
	}

	if err = json.Unmarshal([]byte(wifiByte), wifi); err != nil {
		logger.Debugf("GetWifiInfo From Cache redis.Bytes Unmarshal error : %+v", err)
		return nil, err
	}

	return wifi, nil
}

// 根据account id查询wifi信息
func GetWifiInfoByAccount(aId int32) ([]*pb.Wifi, error) {
	var (
		rd      redis.Conn
		err     error
		wifiS   = make([]*pb.Wifi, 0)
		wifiStr []string
	)

	if rd = cache.GetRedisClient(); rd == nil {
		logger.Debugln("redis Conn is nil")
		return nil, errors.New("redis Conn is nil")
	}
	defer rd.Close()

	logger.Debugf("will get account id: %d", aId)
	if wifiStr, err = redis.Strings(rd.Do("SMEMBERS", makeWifiKey(aId))); err != nil { //TODO 空针
		logger.Debugf("GetWifiInfoByAccount From Cache redis.Bytes error : %+v", err)
		return nil, nil
	}
	for _, w := range wifiStr {
		wifi := &pb.Wifi{}
		if err = json.Unmarshal([]byte(w), wifi); err != nil {
			logger.Debugf("GetWifiInfo From Cache redis.Bytes Unmarshal error : %+v", err)
			return nil, err
		}
		logger.Debugf("GetWifiInfo From Cache redis.Bytes Unmarshal wifi: %+v", wifi)
		wifiS = append(wifiS, wifi)
	}

	return wifiS, nil
}

// 从缓存删除MAC地址
func DelWifiInfoToCache(wifiObjs *pb.WifiInfoReq) error {
	var (
		rd        redis.Conn
		wifiBytes []byte
		err       error
	)

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()

	for _, v := range wifiObjs.Wifis {
		if wifiBytes, err = redis.Bytes(rd.Do("GET", v.BssId)); err != nil { //TODO 空针
			logger.Errorf("GetWifiInfo From Cache redis.Bytes error : %+v", err)
			return err
		}
		_ = rd.Send("MULTI")
		//key := "wifi:"+ v.BssId +":info" // 直接bssid做key
		_ = rd.Send("DEL", v.BssId, makeUserWifiMap(v.Id))
		_ = rd.Send("SREM", makeWifiKey(wifiObjs.AccountId), string(wifiBytes))
		if _, err = rd.Do("EXEC"); err != nil {
			logger.Debugf("Save WifiInfo to cache error: %s\n", err)
			return err
		}
	}

	return nil
}

// 更新缓存数据, 因为可以修改mac地址这个原因，不得已只能增加bssid和id这个键值对 TODO 暂时这么来吧
func UpdateWifiInfoToCache(wifiObjs *pb.WifiInfoReq) error {
	var (
		rd        redis.Conn
		wifiBytes []byte
		bssIdKey  string
		err       error
	)

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("redis Conn is nil")
	}
	defer rd.Close()

	for _, v := range wifiObjs.Wifis {
		if bssIdKey, err = redis.String(rd.Do("GET", makeUserWifiMap(v.Id))); err != nil { //TODO 空针
			logger.Errorf("GetWifiInfo From Cache bssIdKey error : %+v", err)
			return err
		}
		if wifiBytes, err = redis.Bytes(rd.Do("GET", bssIdKey)); err != nil { //TODO 空针
			logger.Errorf("GetWifiInfo From Cache redis.Bytes error : %+v", err)
			return err
		}

		// 1. 先删除原来的
		_ = rd.Send("MULTI")
		_ = rd.Send("DEL", bssIdKey, makeUserWifiMap(v.Id))
		_ = rd.Send("SREM", makeWifiKey(wifiObjs.AccountId), string(wifiBytes))
		if _, err = rd.Do("EXEC"); err != nil {
			logger.Debugf("Save WifiInfo to cache error: %s\n", err)
			return err
		}
	}


	// 2. 反序列化，重新保存到缓存
	for _, v := range wifiObjs.Wifis {
		_ = rd.Send("MULTI")
		wifiBytes, err := json.Marshal(v)
		if err != nil {
			logger.Debugf("json marshal error: %s\n", err)
			return err
		}
		_ = rd.Send("SET", v.BssId, wifiBytes)
		_ = rd.Send("SET", makeUserWifiMap(v.Id), v.BssId)
		_ = rd.Send("SADD", makeWifiKey(wifiObjs.AccountId), wifiBytes)
		if _, err = rd.Do("EXEC"); err != nil {
			logger.Debugf("Save WifiInfo to cache error: %s\n", err)
			return err
		}
	}
	return nil
}

// 找出信号最强的那个wifi的bssid
func FindMostMatchWifiBssId(wifis []*pb.Wifi) string {
	logger.Debugf("findMostMatchWifiBssId slice :%+v", wifis)
	var (
		bssId    = wifis[0].BssId
		maxLevel = wifis[0].Level
	)
	if len(wifis) == 1 {
		return bssId
	}
	for i := 1; i < len(wifis); i++ {
		if maxLevel <= wifis[i].Level {
			maxLevel = wifis[i].Level
			bssId = wifis[i].BssId
		}
	}
	logger.Debugf("find max level bassId: %s", bssId)
	return bssId
}
