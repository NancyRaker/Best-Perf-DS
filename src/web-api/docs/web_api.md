## api 定义

### 1. 创建下级账户   

请求协议：`http`

请求方式：`POST`

请求地址：`/account` 

请求参数：请求body必须包含username和pwd和用户类型
``` json
{
	"confirm_pwd" :"123456",    // 不能为空
	"pid": 5,    				// 不能为空
	"username": "liuyang06",   // 不能为空
	"nick_name": "nana123",   // 不能为空
	"pwd": "123456",     // 不能为空
	"role_id": 3,    // 不能为空
	"email": "123456789@qq.com",
	"privilege_id": 0,
	"contact": "",
	"state": "",
	"phone": "",
	"remark": "",
	"address": ""
}
```

返回参数：body中的session_id
``` json
{
	"success": true,
	"account_id": 1430   // 新创建的用户的id
}
```

### 2. 登录
请求协议：`http`

请求方式：`POST`

请求地址：`/account/login.do/account_name` 

请求参数：请求body中：username和pwd
``` json
{
	"username" : "account_name",
	"pwd" : "123456",
}
```

返回参数：body中：
``` json
{
	"success": true,
	"session_id": "c9f9173c-7cc8-44c3-81a8-7c72d9863f9a"
}
```

### 3. 退出
请求协议：`http`

请求方式：`POST`

请求地址：`/account/logout.do/account_name` 

请求参数：请求body中：username和pwd，请求头中添加返回的session_id
``` json
{
	"username" : "account_name"
}
```

返回参数：body中：
``` json
{
	"success": true,
	"msg": "SignOut is successful",
}
```

### 4. 账户信息以及下面的所有群组
请求协议：`http`

请求方式：`GET`

请求地址：`/account/:account_name`   (`:account_name`代表账户名称，比如访问：`172.16.0.74:8080/account/tiger`,其中tiger为账户名称)  

请求参数：**请求头中添加返回的session_id**  

返回参数：body中：  （返回都是json格式）
``` json
{

	"message" :" 获取用户信息成功",
	"account_info" : "账户的信息",
	"group_list" : "群组的信息",
	"device_list" ： "账户下所有的设备"
	"tree_data": "账户的层级关系"
}
```

### 5. 修改账户信息

请求协议：`http`

请求方式：`POST`

请求地址：`/account/info/update`

请求参数：**请求头中添加返回的session_id**   (以下字段json格式，可以为空)
``` go
	Id       string `json:"id"`   //不能为空
	LoginId  string `json:"login_id"`   //不能为空
	NickName string `json:"nick_name"`  // 不能为空
	Username string `json:"username"`  
	TypeId   string `json:"type_id"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Remark   string `json:"remark"`
	Contact  string `json:"contact"`
```

返回参数：body中： 

``` go
			"success": "true",
			"msg":     "update account success",
```

### 6. 修改账户密码
请求协议：`http`

请求方式：`POST`

请求地址：`/account/pwd/update`

请求参数：**请求头中添加返回的session_id**  
``` go
{
	Id         string `json:"id"`    // 账户id
	OldPwd     string `json:"old_pwd"`
	NewPwd     string `json:"new_pwd"`
	ConfirmPwd string `json:"confirm_pwd"`
}
```

返回参数：body中： 
``` go
{
	"result": "success",
	"msg":    "Password changed successfully",
}
```

### 7. 获取账户下级目录
请求协议：`http`

请求方式：`POST`

请求地址：`/account_class/:accountId/:searchId`

请求参数：accountId是登录者的ip, searchId是需要获取的用户的下级目录的id **请求头中添加返回的session_id**  

返回参数：body中：  (以下字段json格式)

``` go
		"result":    "success",
		"tree_data": resp,

		Id          int             `json:"id"`
		AccountName string          `json:"account_name"`
		Children    []*AccountClass `json:"children"`
```

### 8. 获取下级某个用户的所有设备
请求协议：`http`

请求方式：`GET`

请求地址：`/account_device/:accountId/:getAdviceId`

请求参数：accountId是登录者的id, searchId是需要获取的用户的下级目录的id **请求头中添加返回的session_id**  

返回参数：body中：  (以下字段json格式)
```
		"account_info": ai,
		"devices":      deviceAll,
```

### 9. 转移设备

请求协议：`http`

请求方式：`POST`

请求地址：`/account_device/:accountId`

请求参数：accountId是登录者的id, **请求头中添加返回的session_id**  

返回参数：body中：  (以下字段json格式)
```
		"result": "success",
		"msg":    "trans successful"
```


### 10. 导入设备  grpc

请求协议：`http`

请求方式：`POST`

请求地址：`/device/import/SurpeRoot`

请求参数：**请求头中添加返回的session_id** 

请求body：一个string数组
``` go
	DeviceIMei []string `json:"device_imei"`
```

返回参数：body中：  (以下字段json格式)
```
		"msg":"import device successful."
```


### 11. 创建组  grpc
请求协议：`http`

请求方式：`POST`

请求地址：`/group`

请求参数：**cookie中添加返回的session_id**  // 调试的时候暂时注释


``` json
{
  "device_ids": [ -1 ],
  "device_infos": [设备对象],
  "group_info": {
    "group_name": "重庆组",    // 必须的
    "account_id": 6,         // 必须的
    "status": "1",
    "c_time": "2019-03-18 10:28:26"
  }
}
```
返回参数：  

``` json
		"result": "success",
		"msg":    "Create group successfully",
```

### 12. 更新组名  目前web只用改组名

请求协议：`http`

请求方式：`POST`

请求地址：`/group/update`

请求参数：**cookie中添加返回的session_id**  // 调试的时候暂时注释


``` json
{
  "group_info": {
    "group_name": "重庆组",    // 必须的
    "account_id": 6,         // 必须的
    "status": "1",
    "c_time": "2019-03-18 10:28:26"
  }
}
```
返回参数：  

``` json
		"result": "success",
		"msg":    "Update group successfully",
```

### 13. 删除组 grpc
请求协议：`http`

请求方式：`POST`

请求地址：`/group/delete`

请求参数：**cookie中添加返回的session_id**  // 调试的时候暂时注释


``` json
{
  "group_info": {
    "group_name": "重庆组",    // 必须的
    "account_id": 6,         // 必须的
    "status": "1",
    "c_time": "2019-03-18 10:28:26"
  }
}
```
返回参数：  

``` json
		"result": "success",
		"msg":    "Delete group successfully",
```

### 14. 更新组成员  grpc

请求协议：`http`

请求方式：`POST`

请求地址：`/group/devices/update`

请求参数：**cookie中添加返回的session_id**  // 调试的时候暂时注释


``` json
{
  "group_info": {
    "group_name": "重庆组",    // 必须的
    "account_id": 6,         // 必须的
    "status": "1",
    "c_time": "2019-03-18 10:28:26"
  }
}
```
返回参数：  

``` json
		"result": "success",
		"msg":    resUpd.ResultMsg.Msg,

```

***
NFC
***
### 15. tag标签的保存

请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/tags/:account_id`

请求参数：**cookie中添加返回的session_id**


``` json
{
  "ops": 1,  // 1代表录入
  "tags": [
    {
      "tag_name":"标签AAA",
	  "tag_addr":"兴东AAA",
	  "import_timestamp":1570215487,
	  "uuid":"a7dbe513-ea08-4e32-a52c-74632bb89841"
    }
  ],
  "account_id":115
}
```
返回参数：  

``` json
{
    "msg": "The process successfully!"
}
```


### 16. tag标签的修改

请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/tags/:account_id`

请求参数：**cookie中添加返回的session_id** 


``` json
{
  "ops": 2,  // 2代表修改
  "tags": [
    {
      "id":8,
      "tag_name":"标签AAA",
	  "tag_addr":"兴东AAA",
	  "import_timestamp":1570215487,
	  "uuid":"a7dbe513-ea08-4e32-a52c-74632bb89841",
	  "account_id":115
    }
  ],
  "account_id":115
}
```
返回参数：  

``` json
{
    "msg": "The process successfully!"
}
```

### 17. tag标签的删除

请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/tags/:account_id`

请求参数：**cookie中添加返回的session_id**  


``` json
{
  "ops": 3,  //  3代表删除
  "tags": [
    {
      "id":8,
      "tag_name":"标签AAA",
	  "tag_addr":"兴东AAA",
	  "import_timestamp":1570215487,
	  "uuid":"a7dbe513-ea08-4e32-a52c-74632bb89841",
	  "account_id":115
    }
  ],
  "account_id":115
}
```
返回参数：  

``` json
{
    "msg": "The process successfully!"
}
```

### 18. tag标签的查询

请求协议：`http`

请求方式：`GET`

请求地址：`/web-api/tags/:account_id`

请求参数：**cookie中添加返回的session_id** 

返回参数：  

``` json
{
    "tags": [
        {
            "id": 7,
            "tag_name": "标签DDD",
            "tag_addr": "兴东DDD",
            "account_id": 115,
            "import_timestamp": 1570215487,
            "uuid":"a7dbe513-ea08-4e32-a52c-74632bb89841"
        }
    ]
}
```

### 19. 标签任务的保存

请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/tag_tasks/:account_id`

请求参数：**cookie中添加返回的session_id**


``` json
{
  "ops": 1,         // 值为1，代表保存任务 必填 
  "tag_task_lists": [
    {
      "account_id": 115,      // 选填  调度员id
      "send_email": "123456@qq.com", // 选填  任务执行情况报告发送的邮箱
      "send_time": 1571245677, // 选填 发送报告的时间
      "time_zone": "Asia/Shanghai",  // 必填  任务生成的时区
      "tag_task_nodes": [
        {
          "device_id": 62,   // 设备id  必填
          "tag_id": 98,     // 标签id 必填
          "account_id":115,   // 调度员id 必填
          "order_end_time": 3,  // 打卡起始时间时间戳 必填
          "order_start_time": 2   // 打卡结束时间戳  必填
        },
        {
          "device_id": 72,   // 设备id  必填
          "tag_id": 100,     // 标签id 必填
          "account_id":115,   // 调度员id 必填
          "order_end_time": 3,  // 打卡起始时间时间戳 必填
          "order_start_time": 2   // 打卡结束时间戳  必填
        }
      ]
    }
   ],
  "account_id":115  // 调度员的id 必填 
}
```
返回参数：  

``` json
{
    "msg": "The process successfully!"
}
```

### 20. 标签任务的删除

请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/tag_tasks/:account_id`

请求参数：**cookie中添加返回的session_id**  


``` json
{
  "ops": 3,         // 值为3，代表删除任务 必填 
  "tag_task_lists": [
    {
      "tag_task_id":10,  // 必填 任务id
      "account_id": 115,      // 选填  调度员id
      "send_email": "123456@qq.com", // 选填  任务执行情况报告发送的邮箱
      "send_time": 1571245677, // 选填 发送报告的时间
      "tag_task_nodes": [
        {
          "device_id": 62,   // 设备id  选填
          "tag_id": 98,     // 标签id 选填
          "account_id":115,   // 调度员id 选填
          "order_end_time": 3,  // 打卡起始时间时间戳 选填
          "order_start_time": 2   // 打卡结束时间戳  选填
        },
        {
          "device_id": 72,   // 设备id  选填
          "tag_id": 100,     // 标签id 选填
          "account_id":115,   // 调度员id 选填
          "order_end_time": 3,  // 打卡起始时间时间戳 选填
          "order_start_time": 2   // 打卡结束时间戳  选填
        }
      ]
    }
   ],

  "account_id":115  // 调度员的id必填
}
```
返回参数：  

``` json
{
    "msg": "The process successfully!"
}
```