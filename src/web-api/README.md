### jimi对讲平台后台管理系统

1. web-api is a web dispatcher system server writen by golang.


## Features
 * Pure Golang
 * Go module manage project

## Architecture

#### 1. app之间，以及与web用户之间的交互过程：
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/server%20proj%20picture/imserver1.0.0.jpg)

#### 2. Im 推送的处理流程：
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/server%20proj%20picture/impush_model.jpg)

#### 3. Im 推送的数据来源：

1. App用户自己创建的stream模式的grpc的client发送的心跳、聊天文字以及上线下线通知等消息。
2. Web用户通过web-api后台服务创建的stream模式的grpc的client发送的心跳、聊天文字、上线下线通知以及Web用户主动下线等消息。
3. 对讲机对讲的时候由janus服务产生并推送到redis里面的对讲音频消息。

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

1. 到`main.go`文件所在目录下，根据实际情况修改`web_conf.ini`。
2. 到`main.go`文件所在目录下，指定交叉编译的目标环境，例如win7环境编译linux：`set GOOS=linux` -> `go Build`

### Run
1. 到`main.go`文件所在目录下，根据实际情况修改`web_conf.ini`。
2. 到`main.go`文件所在目录下，例如win7环境运行: `go run main.go`

### Environment
未用docker部署，所以目前配置项都是使用配置文件方式，其他地方暂时没有用到环境变量。

### Configuration

### Detail Features

#### 一、已完成
1. 新功能

| 接口功能简介 | 涉及修改 | 详细信息 | 实现办法|
|:----------|:---:|:---:|:--|
|web界面刷新，im重新连接|imMessagePublishDispatcher|sdfg|当web界面刷新，web通过发送消息号为10001，webapi模块把消息转发到这里的imserver，</br>然后imserver把该用户的在线session删除。|
|web用户异地登录，断开原有连接|gfd|
|web用户直接关闭web页面，通知imserver更新用户在线状态|sfdg|
|...|...|...|...|
|轨迹回放数据的获取|GetGpsForTrace接口|根据开始结束时间戳返回经纬度数据|fdgsfdg|
|wifi录入页面gps信息时不做小数点限制|数据库wifi表和location表|默认小数点后15位|ffd|
| web界面临时组| WebCreateGroup |</br>1. 通过地图全选，创建临时组，多次创建只会存在一个临时组。</br>2. 当web页面刷新，临时组解散。</br>3 .当登录用户session过期，强制关闭web界面，临时组解散。</br>4. 同一个用户异地登录，临时组解散。|1. 使用`user_group`表原有的`stat`字段，1表示该组为常规群组，2表示临时组。</br>2.分开了原先写好的app和web创建群组用一个方法，web单独用自己的方法去创建。|
| 增加janus状态|获取群组列表|增加janus在线|根据锁定组找出janus在线||
| web导入设备可以添加到指定账号下方|-|-|-|
|web转移设备时，需要一个新接口，去展示可以转移的账号|GetJuniorAccount|web转移设备时，需要一个新接口，去展示可以转移的账号|新增接口|
2. Bug修复 

| 接口功能简介 | 涉及修改 | 详细信息 | 实现办法|
|:----------|:---:|:---:|:--|
|【Bug转需求】微软云 发送sos报警im链接断开|因为有些设备没有发送sos消息的时候，可能缓存并没有经纬度，wifi description等信息，就会报空针，</br>web-api进程panic，导致im断开|在发送sos消息给web前端时候，增加参数判空校验|web回复sos数据的时候校验数据就可以|
| sos数据返回，wifi des没有会报空针|impush|sos数据返回，wifi des没有会报空针|sos数据返回检测数据|
|【Bug转需求】在线状态获取失败，因为grpc修改了redis对应的key|user_cache中获取状态信息函数|修改redis取数据的key|修改key||
| 修改群组名称通知app |这里修改内容不多，主要是调用rpc方法那边处理|修改群组名称需要通知app|通过grpc通知|

#### 二、未完成

| 接口功能简介 | 涉及修改 | 详细描述 | 实现办法| 
|:----------|:---:|:---|:--|

优化前：getaccountinfo
```
time="2019-08-13 17:13:50.973804211" level=info msg="start get group info: 1565687630973794304, before now use: 124 ms"
time="2019-08-13 17:13:50.974413661" level=info msg="start get group list info: 1565687630974408557, before now use: 0 ms"
time="2019-08-13 17:14:54.956914468" level=info msg="start sort device list : 1565687694956910335, before now use: 63982 ms"
time="2019-08-13 17:14:58.610647342" level=info msg="start sort device list : 1565687698610644260, before now use: 3653 ms"
time="2019-08-13 17:14:58.683305488" level=info msg="end request : 1565687698683285373, before now use: 72 ms"

time="2019-08-13 17:17:10.775696141" level=info msg="start get group info: 1565687830775689518, before now use: 74 ms"
time="2019-08-13 17:17:10.776306688" level=info msg="start get group list info: 1565687830776288697, before now use: 0 ms"
time="2019-08-13 17:17:55.880717317" level=info msg="start sort device list : 1565687875880713549, before now use: 45104 ms"
time="2019-08-13 17:17:59.029781367" level=info msg="start sort device list : 1565687879029777950, before now use: 3149 ms"
time="2019-08-13 17:17:59.065420859" level=info msg="end request : 1565687879065408239, before now use: 35 ms"
```

现在数据库的io和数据传输是耗时大户了，因为垮机房了
```
time="2019-08-16 07:00:25.845546155" level=info msg="start get group info: 1565938825845538955, before now use: 954 ms"
time="2019-08-16 07:00:26.138596938" level=info msg="start get group list info: 1565938826138589738, before now use: 293 ms"
time="2019-08-16 07:00:26.519560577" level=info msg="start sort device list : 1565938826519558577, before now use: 380 ms"
time="2019-08-16 07:00:26.638089560" level=info msg="start get account tree : 1565938826638088060, before now use: 118 ms"
time="2019-08-16 07:00:26.970527137" level=info msg="end request : 1565938826970517136, before now use: 332 ms"

time="2019-08-16 07:13:00.340032312" level=info msg="start get group info: 1565939580340025212, before now use: 779 ms"
time="2019-08-16 07:13:00.462717626" level=info msg="start get group list info: 1565939580462695026, before now use: 122 ms"
time="2019-08-16 07:13:00.847711695" level=info msg="start sort device list : 1565939580847710195, before now use: 385 ms"
time="2019-08-16 07:13:01.002543649" level=info msg="start get account tree : 1565939581002541149, before now use: 154 ms"
time="2019-08-16 07:13:01.218408057" level=info msg="end request : 1565939581218401457, before now use: 215 ms"

time="2019-08-16 07:16:48.394210430" level=info msg="start get group info: 1565939808394202730, before now use: 825 ms"
time="2019-08-16 07:16:48.611952753" level=info msg="start get group list info: 1565939808611944153, before now use: 217 ms"
time="2019-08-16 07:16:48.919283743" level=info msg="start sort device list : 1565939808919281643, before now use: 307 ms"
time="2019-08-16 07:16:49.040778148" level=info msg="start get account tree : 1565939809040776148, before now use: 121 ms"
time="2019-08-16 07:16:49.590536045" level=info msg="end request : 1565939809590527645, before now use: 549 ms"
```