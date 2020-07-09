### 项目简介
服务实现了调度员调度平台相关的rpc调用、对讲机设备相关的rpc调用以及NFC功能的rpc调用、以及im相关的stream模式的rpc调用。

### 项目目录介绍
``` shell
|__ api
|   |__ proto     // grpc 的proto文件和使用protoc和protoc-gen-go工具生成的go文件
|   |   |__ *.proto
|   |   |__ *.pb.proto
|   |__ server  // 整个服务所有的rpc调用的实现
|       |__ app  // 对讲机相关的rpc调用
|       |   |__ *.go 
|       |__ im   // im推送的实现
|       |   |__ *.go    
|       |__ location   // 对讲机上传gps信息的rpc实现
|       |   |__ *.go    
|       |__ nfc   // NFC功能相关的rpc实现
|       |   |__ factory
|       |   |   |__ common_manager  // nfc邮件生成器和邮件发送器的父级管理者
|       |   |   |   |__ base_manager.go  // 父类
|       |   |   |__ report_builder  // 报告生成工厂
|       |   |   |   |__ report_builder.go  // 报告生成器
|       |   |   |   |__ report_builder_manager.go  // 报告生成器的管理者
|       |   |   |__ report_sender  // 报告发送工厂
|       |   |   |   |__ report_builder.go  // 报告发送器
|       |   |   |   |__ report_builder_manager.go  // 报告发送器的管理者
|       |   |   |__ report_task_generator  // 报告任务生成器
|       |   |   |   |__ report_task_generator.go  // 报告任务生成器
|       |   |   |__ scheduler  // 通用队列调度器
|       |   |   |   |__ queue_scheduler.go  // 队列调度器
|       |   |   |__ runner.go // NFC工厂runner
|       |   |__ *.go // NFC相关的增删改查的rpc调用的实现
|       |__ web
|       |   |__ *.go // 调度平台相关的rpc调用
|       |__ core.go // server服务的对外接口
|__ conf  // 配置文件
|__ dao   // 数据库的操作
|__ docs  // 项目的一些文档说明
|__ engine // redi缓存和mysql使用的连接引擎以及grpc客户端连接池
|__ init_load  // 服务启动的时候一些初始化操作
|__ internal   // 一些服务内部用到的公共服务
|   |__ event_hub  // 与janus通信的时候，对消息队列里的回复消息进行管理
|   |__ mq   // 消息队列
|       |__ mq_provider // 消息队列的提供者，比如kafka或者redis作为消息队列，提供消息的消费者和生产者接口
|       |__ mq_receiver // 从mq_provider获取消息，暴露接收消息接口供上层rpc调用服务使用
|       |__ mq_sender // 向mq_provider发送消息，暴露发送消息接口供上层rpc调用服务调用
|__ logger // 日志
|__ model  // 项目中会用到的公共消息结构体
|__ test   // 一些测试代码
|__ utils  // 一些工具类
|__ vendor  // 打包了项目中用到的所有第三方库
|__ bo_conf.ini // 整个项目的配置文件
|__ go.mod  // go mod文件
|__ go.sum  // mod管理模式下的sum文件
|__ main.go  // 项目入口
|__ README.md 
```