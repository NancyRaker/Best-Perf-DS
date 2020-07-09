### jimi对讲平台服务发现网关服务

1. web-ips is a simple server discovery webserver writen by golang.

## Features
 * Pure Golang
 * Go module manage project

## Architecture
1. 这个服务为旨在为超级管理员、经销商、调度员和对讲机按照需求分发janus、web、grpc的服务地址。
对于对讲机设备，因为通过15位的imei登录，所以维护了一个前缀树，来记录设备属于哪个调度员名下，然后对于每一个服务访问点利用一个bitmap来存储可以登录的超级管理员、经销商、调度员的id。
在发送请求获取服务的时候，调度员等就直接去bitmap查找，时间复杂度是线性的。如果是普通设备，就先到前缀树查找调度员id再到bitmap中查找应该返回的地址。
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/web-gateway-pic/%E9%A1%B9%E7%9B%AE%E9%80%BB%E8%BE%91%E5%A4%84%E7%90%86%E6%B5%81%E7%A8%8B.PNG)
如图所示，整个服务在开始运行的时候就从数据库加载数据，之后就一直维护一个前缀树用于设备的imei匹配调度员id，两个bitmap用来查找调度员的该返回的服务地址，一个map用来映射非设备用户的账号和id之间的关系，账号作为key，id作为value。

2. 使用bitmap独立管理调度员的map，可以在灾备切换服务的时候，很容易控制将一个服务的调度员及其所属设备统一转移到另外的机器上。

3. 将设备和管理员分离，可以更方便的改变调度员登录服务地址转移到其他服务上，设备则不需要与服务地址有联系，调度员应该在哪登录，设备就跟随调度员。

## Quick Start
### Build

1. 到`main.go`文件所在目录下，根据实际情况修改`conf.ini`。
2. 到`main.go`文件所在目录下，指定交叉编译的目标环境，例如win7环境编译linux：`set GOOS=linux` -> `go Build`

### Run
1. 到`main.go`文件所在目录下，根据实际情况修改`conf.ini`。
2. 到`main.go`文件所在目录下，例如win7环境运行: `go run main.go`

### Environment
未用docker部署，所以目前配置项都是使用配置文件方式，其他地方暂时没有用到环境变量。

### Configuration

### Detail Features

#### 一、已完成
1. 新功能

| 接口功能简介 | 详细信息 | 实现办法|
|:--|:--|:--|
| "/server/addr", controllers.GetServerAddr)|对讲机调用，返回janus，web，grpc这些服务地址|根据imei号找到所属调度员，然后根据调度员找到应该返回的地址|
| "/server/change", controllers.ChangeServer)| 改变调度员应该返回的服务地址 | 根据所传参数更新bitmap，以及更新数据库|
| "/server/device/change", controllers.ChangeDispatcher)|  改变某些设备所属的调度员，也就是转移设备 |  更新前缀树 |
| "/server/device/add", controllers.AddDeviceForDispatcher)|  导入设备时候，增加设备节点 | 对于前缀树插入节点 |

2. Bug修复 

| 接口功能简介 | 涉及修改 | 详细信息 | 实现办法|

#### 二、未完成

1. 新功能

| 接口功能简介 | 详细信息 | 实现办法|
| 灾备切换| 发生灾备情况时候，将某台机器的可以登录的用户全部转移到其他bitmap下|更新bitmap|

2. Bug修复 

| 接口功能简介 | 涉及修改 | 详细信息 | 实现办法|
ALTER TABLE `talk_platform`.`customer` ADD COLUMN `server_addr` INT(12) DEFAULT 1 NULL COMMENT '该用户应该访问的服务地址，目前是东莞或者微软云，如果customer是调度员那么名下所有的设备也到该地址，目前1是微软云 2是东莞' AFTER `uid`; 

