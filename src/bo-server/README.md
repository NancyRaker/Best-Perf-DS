### jimi对讲平台im和后台管理服务

1. bo-server is a im server writen by golang.


## Features
 * Pure Golang
 * Go module manage project

## Architecture

#### 1. app之间，以及与web用户之间的交互过程：
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/server%20proj%20picture/imserver1.0.0.jpg)

#### 2. app登录流程：
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/server%20proj%20picture/app_login_schduler.jpg)

#### 3. Im 推送的处理流程：
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/server%20proj%20picture/impush_model.jpg)

#### 4. Im 推送的数据来源：

1. App用户自己创建的stream模式的grpc的client发送的心跳、聊天文字以及上线下线通知等消息。
2. Web用户通过web-api后台服务创建的stream模式的grpc的client发送的心跳、聊天文字、上线下线通知以及Web用户主动下线等消息。
3. 对讲机对讲的时候由janus服务产生并推送到redis里面的对讲音频消息。

### 5. janus与web进行消息队列通信流程
***
### IM Push Pro:

**判断app下线的情况：**

只通过心跳消息来确定，即时发送消息失败，按照离线消息处理（// 这里就会出现重复消息，如果对方其实已经收到）。

**<font color="ff0000">TODO:</font>**
**如何保证消息不丢不重?**
**消息重复**：客户端发送消息，由于网络原因或者其他原因，客户端没有收到服务器的发送消息回执，这时客户端会重复发送一次，假设这种情况，客户端发送的消息服务器已经处理了，但是回执消息丢失了，客户端再次发送，服务器就会重复处理，这样，接收方就会看到两条重复的消息，解决这个问题的方法是，客户端对发送的消息增加一个递增的序列号，服务器保存客户端发送的最大序列号，当客户端发送的消息序列号大于序列号，服务器正常处理这个消息，否则，不处理，回执消息一律返回服务器此时最大的序列号。

**消息丢失**：服务器投递消息后，如果客户端网络不好，或者已经断开连接，或者是已经切换到其他网络，服务器并不是立马可以感知到的，服务器只能感知到消息已经投递出去了，所以这个时候，就会造成消息丢失。

**怎样解决：**
1. 消息持久化：
2. 投递消息增加序列号
3. 增加消息投递的ACK机制，就是客户端收到消息之后，需要回执服务器，自己收到了消息。
这样，服务器就可以知道，客户端那些消息已经收到了，当客户端检测到stream不可用（使用心跳机制），重新建立连接，带上自己收到的消息的最大序列号，触发一次消息同步，服务器根据这个序列号，将客户端未收到再次投递给客户端，这样就可以保证消息不丢失了。
***

## Quick Start
### Build

1. 到`main.go`文件所在目录下，根据实际情况修改`bo_conf.ini`。
2. 到`main.go`文件所在目录下，指定交叉编译的目标环境，例如win7环境编译linux：`set GOOS=linux` -> `go Build`
3. vendor中已经打包了所有编译会用到的库，所以也可以直接编译：`go build -mod vendor`

### Run
1. 到`main.go`文件所在目录下，根据实际情况修改`bo_conf.ini`。
2. 到`main.go`文件所在目录下，例如win7环境运行: `go run main.go`

### Environment
未用docker部署，所以目前配置项都是使用配置文件方式，其他地方暂时没有用到环境变量。

### Configuration

### Detail Features

#### 一、已完成
1. 新功能

| 接口功能简介 | 涉及修改 | 详细信息 |
|:----------:|:---|:---|
| 单个用户之间的消息转发| DataPublish接口，ImMessagePublish接口|消息全部由服务进程转发|
| 群聊消息转发|DataPublish接口，ImMessagePublish接口|消息全部由服务进程转发|
|...|...|...|
| 轨迹回放数据的获取|GetGpsForTrace接口|根据开始结束时间戳返回经纬度数据|
| wifi录入页面gps信息时不做小数点限制|数据库wifi表和location表|默认小数点后15位|
| web界面临时组 | WebCreateGroup ImMessagePublish, 数据库user_group的stat字段 |分开了原先写好的app和web创建群组用一个方法，web单独用自己的方法去创建|
| app自动升级| Login |增加版本号之后，受影响的还有web的im登录时，也需要这个版本号|
| supervisor增加短信提醒|supervisor配置文件|新增event listener，当有标准输出中有预定的event就调用公司的接口发送短信|
| 把一个调度员下面的设备转移到另外一个调度员下面，之前的的调度员已经不存在这个设备了，但是群组里还是有这一个设备。||||
|增加app上传apk升级服务|web-interval中的web_update_apk服务|提供一个服务来更新apk的版本|

2. Bug修复 

| 缺陷描述 | 涉及修改 | 详细信息 |
|:----------:|:---|:---|
|【Bug转需求】web端在修改了设备的昵称后，im的上线提醒的名字没有同步还是使用之前的名字|修改设备信息接口UpdateDeviceInfo| 因为之前前端传来的参数只有imei号，没有id，所以缓存里的数据没有同时更改|
| 推送sos消息的时候只推送设备所在janus当前组|修改涉及proto文件和sos推送| 之前的sos是推送给web所在群组的所有在线人 |
|【Bug转需求】微软云轨迹回放丢失数据|dao中的location，数据库location表的local_time直接改成时间戳字符串|数据库location表的local_time直接改成时间戳字符串|
|redis连接池取连接的时候应该加上上下文去限制等待时间|使用context，在引擎里面做了修改||
|ptt只发给当前群组|imInstance|将ptt的dispatcher和普通im分开，因为这个ptt语音消息只能发送给janus当前房间|
|响应服务地址|web-gateway|单独运行一个服务，将服务地址信息写在配置文件用来分发|
|创建组流程出现问题|webCreateGroup函数|通过允许创建只含有调度员一个人的群组|
|login登录优化|getGroupList|同一个设备的信息就只获取一次|
|redis sentinel 配置还有修改|redis引擎，redis.ini, redis的配置加载文件|使用sentinel模式，自己选择redis地址|
|增加janus状态为3之后引起sos判断在线人数出现问题，导致sos发送人员缺少|notifySosToOther|增加if条件判断|

#### 二、未完成
| 接口功能简介 | 涉及修改 | 详细信息 |
|:----------:|:---|:---|
|增加消息回执过程，遇到问题，消耗流量，消息重复| | |
|mysql主备切换|db/engine/change接口，track.sh||

