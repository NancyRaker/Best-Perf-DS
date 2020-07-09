### 一、janus与bo-server和file-server之间的消息交互流程

如图所示，项目中三个服务之间使用kafka中间件来传递消息。

![](https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/server%20proj%20picture/bo-server%E4%B8%8Ejanus%E4%B9%8B%E9%97%B4%E7%9A%84%E6%B6%88%E6%81%AF%E9%98%9F%E5%88%97%E4%BA%A4%E4%BA%92%E8%BF%87%E7%A8%8B.PNG)

#### 1、操作类消息处理流程：
1. **创建房间**：首先web前端发送http请求通知web-api服务，然后web-api服务通过grpc调用，通知bo-server创建群组，数据写入数据库和redis，随后通过kafka消息队列通知janus创建房间，阻塞等待消息结果，这里如果过了超时时间还没有返回，就认为操作失败，bo-server就删除刚刚创建的群组。返回web-api错误，通知web前端操作失败。
2. **删除房间**：// TODO  考虑使用一个单独的异步服务，统一往数据库读取群组id，通知janus删除房间。

#### 2、对讲消息处理：
1. janus生产对讲消息，bo-server转发消息给调度员，file-server负责转存文件和存储文件元信息。


### 二、janus与grpc消息协议定义

消息传递目前全部只使用kafka官方消息体中的value字段来传递json字符串，来作为消息信令传递的方式。

#### 1、janus往kafka写入的消息体
1.  **ptt对讲消息**

- json结构：

```json
{
    "msg_type" : "int 类型，// 1是对讲消息，2是后台信令控制类消息的响应结果",
    "ip" : "string 类型，//消息的ip，其实就目前的架构，这个ip永远只有本机，因为grpc和janus都在同一台服务器",
    "ptt_msg" : {
        "uid" :  "int 类型，设备id",
        "m_type" : "string 类型， 消息类型，默认ptt，可以不发送",
        "md5" :  "string 类型， 可以不发送，有file-server自己做文件md5",
        "grp_id" : "string 类型， 对讲消息要发送的群组",
        "file_path" :  "string 类型，文件地址",
        "timestamp" : "string 类型"
    },
}
```

- example：

``` json
{
    "msg_type": 1,
    "ip":"127.0.0.1", 
    "ptt_msg": {
        "uid": "14",
        "md5": "md5",
        "grp_id": "1",
        "timestamp": "1569569697",
        "file_path": "https://dev.yunptt.com:9666/room_1_user_14_1569569695.mp3"
    }
}

```

2. **当janus收到bo-server发送的操作类消息之后，对应的是后台信令控制类消息的响应结果：**

（1）创建房间回复

- json结构：

```json
{
    "msg_type" : "int 类型，// 1是对讲消息，2是后台信令控制类消息的响应结果",
    "ip" : "string 类型，//消息的ip，值是grpc传递过去的值",
    "signal_msg" : {
        "signal_type" : "int 类型，// 目前1代表创建群组消息的响应",
        "create_group_resp" : {
            "gid": "int类型, 值就是grpc传过去的",
            "group_name": "xxx group",
            "dispatcherId" : "int类型, 值就是grpc传过去的调度员的id",
        },
        "res": {
            "code": "消息处理结果代码，目前暂时定义0是成功，其他的是错误码...",
            "msg": "string 错误的内容"
        }
    }
}
```

- example：

```json
{
    "msg_type": 2,
    "ip": "127.0.0.1",
    "signal_msg": {
        "signal_type": 1,
        "create_group_resp": {
            "gid": 470,
            "group_name": "0929group1128",
            "dispatcher_id": 7
        },
        "res": {
            "code": 0,
            "msg": "command executed successfully"
        }
    }
}
```

3. **对讲机对讲时候，janus发送出来的通话质量的消息：**

（1）创建房间回复

- json结构：

```json
{
    "msg_type" : "int 类型，// 1是对讲消息，2是后台信令控制类消息的响应结果, 3是通话质量消息",
    "ip" : "string 类型，//消息的ip，值是grpc传递过去的值",
    "quality_msg": {
        "user_id": "设备id",
        "user_status": "1是talker 2是listener",
        "in_link_quality": "xx质量",
        "in_link_media_quality": "xx质量",
        "out_link_quality": "xx质量",
        "out_link_media_quality": "xx质量",
        "rtt":"xx质量"
    }
}
```

- example：

```json
{
    "msg_type": 3,
    "ip": "127.0.0.1",
    "quality_msg": {
        "user_id": 62,
        "user_status": 1,
        "in_link_quality": 100,
        "in_media_link_quality": 100,
        "out_link_quality": 100,
        "out_media_link_quality": 100,
        "rtt": 100
    }
}
```

### 2. grpc发送给janus的请求类消息

（1）、创建房间
- json结构：

```json
{
  "msg_type" : "int 消息类型，永远只会为2，因为grpc只会通知janus请求类的消息",
  "ip" : "string 类型，//消息的ip，值是grpc传递过去的本机的值",
  "signal_msg" : {
        "signal_type" : "int 类型，// 目前1代表创建群组的请求消息",
        "create_group_req" : {
            "注释": "当signal_type为1，才有这个对象",
            "gid": "int类型, 群组id",
            "group_name": "xxx group",
            "dispatcher_id" : "int类型, 调度员id"
        }
  }
}
```

- example：

``` json 
{
    "msg_type": 2,
    "ip": "127.0.0.1",
    "signal_msg": {
        "signal_type": 1,
        "create_group_req": {
            "gid": 470,
            "group_name": "0929group1128",
            "dispatcher_id": 7
        }
    }
}
```

（2）、删除房间
- json结构：

```json
{
  "msg_type" : "int 消息类型，永远只会为2，因为grpc只会通知janus请求类的消息",
  "ip" : "string 类型，//消息的ip，值是grpc传递过去的本机的值",
  "signal_msg" : {
        "signal_type" : "int 类型，// 目前2代表创建群组的请求消息",
        "destroy_group_req" : {
            "注释": "当signal_type为2，才有这个对象",
            "gid": "int类型, 群组id",
            "dispatcher_id" : "int类型, 调度员id"
        }
  }
}
```

- example：

``` json 
{
    "msg_type": 2,
    "ip": "127.0.0.1",
    "signal_msg": {
        "signal_type": 2,
        "destroy_group_req" : {
            "gid": 470,
            "dispatcher_id": 7
        }
    }
}
```

（3）、修改成员角色（优先级）
- json结构：
注意，用户一定要在房间才可成功修改对应的角色，另外，只要redis的群组管理员信息已经写入，后台会在用户join或者changeroom的时候主动从reids查询并更新用户的角色信息。

```json
{
  "msg_type" : "int 消息类型，永远只会为2，因为grpc只会通知janus请求类的消息",
  "ip" : "string 类型，//消息的ip，值是grpc传递过去的本机的值",
  "signal_msg" : {
        "signal_type" : "int 类型，// 目前3代表修改角色的请求消息",
        "alter_participant_req" : {
            "注释": "当signal_type为3，才有这个对象",
            "gid": "int类型, 群组id",
            "uid": "int类型, 用户id",
            "role": "int类型, 用户角色（0:普通成员, 1:管理员, 2:拥有者, 其他值非法,权限依次递增）",
            "dispatcher_id" : "int类型, 调度员id"
        }
  }
}
```

- example：

``` json 
{
    "msg_type": 2,
    "ip": "127.0.0.1",
    "signal_msg": {
        "signal_type": 3,
        "alter_participant_req" : {
            "gid": 470,
            "uid": 17,
            "role": 1,
            "dispatcher_id": 7
        }
    }
}
```
