## NFC API 定义

### 1. tag标签的保存

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


### 2. tag标签的修改

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

### 3. tag标签的删除

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

### 4. tag标签的查询

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

### 5. 标签任务的保存

请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/tag_tasks/:account_id`

请求参数：**cookie中添加返回的session_id**


``` json
{
  "ops": 1,         // 值为1，代表保存任务 必填 
  "tag_task_lists": [
    {
      "task_name": "task no 1",
      "save_time": 15712458612, // 必填 任务保存时间
      "account_id": 115,      // 选填  调度员id
      "send_email": "123456@qq.com", // 选填  任务执行情况报告发送的邮箱
      "send_time": 1571245677, // 选填 发送报告的时间
      "time_zone": "Asia/Shanghai",  // 必填  任务生成的时区
      "tag_task_nodes": [
        {
          "device_id": 62,   // 设备id  必填
          "tag_id": 98,     // 标签id 必填
          "tag_name":"标签A",   // 标签名字 必填
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

### 6. 标签任务的删除

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

### 7. 单个设备的所有任务的查询

请求协议：`http`

请求方式：`GET`

请求地址：`/web-api/tag_tasks_device/:account-id/:device-id`

请求参数：**cookie中添加返回的session_id** 

返回内容：  

``` json
{
  "single_device_task_list": [
    {
      "tag_task_id":10,  // 任务id
      "task_name":"task A",  // 任务名字
      "save_time": 1571245862,  // 任务保存的时间
      "tag_list":["一楼", "二楼", "大堂"],  // 打卡的标签名字
    },
    {
      "tag_task_id":11,  // 任务id
      "task_name":"task B",  // 任务名字
      "save_time": 1571245861,  // 任务保存的时间
      "tag_list":["一楼", "三楼", "大堂"],  // 打卡的标签名字
    }
  ]
}
```


### 8. 单个任务的详情的查询
请求协议：`http`

请求方式：`GET`

请求地址：`/web-api/tag_tasks/:account-id/:task-id`

请求参数：**cookie中添加返回的session_id** 

返回内容：  

``` json
{
  "task_detail":{
    "tag_nodes":[
      {
          "id": 7, // 标签id
          "tag_name": "标签AAA",  // 标签名字
          "order_start_time": 2,   // 打卡结束时间戳  
          "order_end_time": 3  // 打卡起始时间时间戳
      },
      {
          "id": 8,   // 标签id
          "tag_name": "标签DDD",  // 标签名字
          "order_start_time": 20,   // 打卡结束时间戳  
          "order_end_time": 30  // 打卡起始时间时间戳
      }
    ] 
  } 
}
```

### 9. 查询设备打卡情况
请求协议：`http`

请求方式：`GET`

请求地址：`/web-api/device/clock/:account-id/:device-id/:start-timestamp`

请求参数：**cookie中添加返回的session_id** 

返回内容：  

``` json
{
  "record_list" : [
    {
      "tag_name": "标签A",
      "record_time": 1572416846,
    },
    {
      "tag_name": "标签B",
      "record_time": 1572419848,
    }
  ]
```

### 10. 保存调度员邮箱地址，周报月报时间等
请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/account_clock/:account-id`

请求参数：**cookie中添加返回的session_id** 

请求body内容：  

``` json
{
    "ops":1,    // 必填  1代表保存信息
    "detail_param": {
        "account_id": 115, // 必填 调度员id
        "report_email": "xxxqqq@qq.com",  // 选填  如果不填，默认不发周报月报，默认每个月1号这个时间点发送上个月的打卡情况
        "month_time": "08:00:00",       // 选填 如果不填，默认不发月报
        "day_time": "08:00:00"          // 选填 如果不填，默认不发周报
    }
}
```

### 11. 修改调度员邮箱地址，周报月报时间等
请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/account_clock/:account-id`

请求参数：**cookie中添加返回的session_id** 

请求body内容：  

``` json
{
    "ops":2,    // 必填  2代表更新信息
    "detail_param": {
        "account_id": 115, // 必填 调度员id
        "report_email": "xxxqqq@qq.com",  // 选填  如果不填，默认不发周报月报，默认每个月1号这个时间点发送上个月的打卡情况
        "month_time": "08:00:00",       // 选填 如果不填，默认不发月报
        "day_time": "08:00:00"          // 选填 如果不填，默认不发周报
    }
}
```

### 12. 删除调度员邮箱地址，周报月报时间等
请求协议：`http`

请求方式：`POST`

请求地址：`/web-api/account_clock/:account-id`

请求参数：**cookie中添加返回的session_id** 

请求body内容：  

``` json
{
    "ops":3,    // 必填  3代表删除信息
    "detail_param": {
        "account_id": 115, // 必填 调度员id
        "report_email": "xxxqqq@qq.com",  // 选填  如果不填，默认不发周报月报，默认每个月1号这个时间点发送上个月的打卡情况
        "month_time": "08:00:00",       // 选填 如果不填，默认不发月报
        "day_time": "08:00:00"          // 选填 如果不填，默认不发周报
    }
}
```

### 13. 查询调度员邮箱地址，周报月报时间等
请求协议：`http`

请求方式：`GET`

请求地址：`/web-api/account_clock/:account-id`

请求参数：**cookie中添加返回的session_id** 

返回内容：  

``` json
{
    "detail": {
        "account_id": 115,
        "report_email": "xxxqqq@qq.com", // 如果这个值是空串，这个字段都会没有
        "day_time": "08:00:00"// 如果这个值是空串，这个字段都会没有，比如这个例子中month_time就是空串
    }
}

```
如果没有数据：
``` json
{}
```

