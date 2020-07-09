### 文件转存服务

1. file server is a file storage server writen by golang.


## Features
 * Pure Golang
 * Go module manage project

## Architecture

#### 1.消息处理流程：
![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/file_server_pic/%E6%B6%88%E6%81%AF%E5%A4%84%E7%90%86%E6%B5%81%E7%A8%8B.PNG)

如上图所示，整个服务由engine挂起，engine启动的时候，首先会创建很多个worker，然后创建调度器来负责管理消息和worker之间的关系。
最后，挂起listener去获取ptt消息，listener挂起的时候，又会运行起来一个kafka的consumer。(listener的消息来源是一个接口，这个接口可以是各种消息队列，或者redis，http请求等，这里使用的是kafka消息队列。)

运行起来之后，消息就先从kafka的consumer获得，然后转发给listener，然后提交给engine，然后由engine交付给scheduler，托付给scheduler根据worker工作完成情况和队列情况进行消息分发处理。

## Quick Start
### Build
1. 到`main.go`文件所在目录下，根据实际情况修改配置文件`file_storage_conf.ini`。
2. 到`main.go`文件所在目录下，指定交叉编译的目标环境，例如win7环境编译linux：`set GOOS=linux` -> `go Build`

### Run
1. 到`main.go`文件所在目录下，根据实际情况修改配置文件`file_storage_conf.ini`。
2. 到`main.go`文件所在目录下，例如win7环境运行: `go run main.go`

### Environment
未用docker部署，所以目前配置项都是使用配置文件方式，其他地方暂时没有用到环境变量。

### Configuration

### Detail Features

#### 一、已完成
1. 新功能

| 接口功能简介 | 涉及修改 | 详细信息 | 实现办法|
|:----------|:---:|:---:|:--|

#### 二、未完成
