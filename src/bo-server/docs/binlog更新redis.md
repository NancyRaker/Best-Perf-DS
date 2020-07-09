### canal中间件：

1. 运行canal服务，使用客户端去访问canal服务，获取数据库操作信息。
2. canal使用protobuf协议，客户端从canal服务得到的消息有很多类型，用得到的主要是DELETE、INSERT、UPDATE。
3. 经过canal服务使用protobuf协议包装后的消息，只能获取到某一行记录或者某一行记录修改前和修改后的数据。
形如：
```
binlog[mysql-bin.000001 : 585578],name[talk_platform,location], eventType: INSERT

uid :     8  : sqlType:-5 name:"uid" updated:true isNull:false value:"8" mysqlType:"bigint(12)"   
lng : 113.927944000  : index:1 sqlType:3 name:"lng" updated:true isNull:false value:"113.927944000" mysqlType:"decimal(19,15)"   
lat : 22.579768000  : index:2 sqlType:3 name:"lat" updated:true isNull:false value:"22.579768000" mysqlType:"decimal(19,15)" 
cse_sp :   0,0  : index:3 sqlType:12 name:"cse_sp" updated:true isNull:false value:"0,0" mysqlType:"varchar(128)" 
local_time : 1572243973  : index:4 sqlType:12 name:"local_time" updated:true isNull:false value:"1572243973" mysqlType:"varchar(128)" 
country :   460  : index:5 sqlType:12 name:"country" updated:true isNull:false value:"460" mysqlType:"varchar(255)" 
operator :     1  : index:6 sqlType:12 name:"operator" updated:true isNull:false value:"1" mysqlType:"varchar(255)"   
lac :  9550  : index:7 sqlType:-5 name:"lac" updated:true isNull:false value:"9550" mysqlType:"bigint(12)"   
cid : 115108868  : index:8 sqlType:-5 name:"cid" updated:true isNull:false value:"115108868" mysqlType:"bigint(12)" 
bs_sth : 0,0,0,0  : index:9 sqlType:12 name:"bs_sth" updated:true isNull:false value:"0,0,0,0" mysqlType:"varchar(255)" 
wifi_sth : -46,a8:0c:ca:04:7e:f7|-48,a8:0c:ca:0c:7e:f7|-56,a8:0c:ca:84:7e:f7  : index:10 sqlType:12 name:"wifi_sth" updated:true isNull:false value:"-46,a8:0c:ca:04:7e:f7|-48,a8:0c:ca:0c:7e:f7|-56,a8:0c:ca:84:7e:f7" mysqlType:"varchar(255)" 
bt_sth :        : index:11 sqlType:12 name:"bt_sth" updated:true isNull:true mysqlType:"varchar(255)" 
create_time : 2019-10-28 14:26:15  : index:12 sqlType:93 name:"create_time" updated:true isNull:false value:"2019-10-28 14:26:15" mysqlType:"timestamp" 
```

### bo-server更换缓存同步方式需要的改动：

更新方式涉及的mysql表:

#### 1. `location` 存储设备上传的定位数据

描述：只有一个插入操作需要同步缓存
需要注意： 根据mysql的插入记录数据去反推测，该调记录是wifi定位还是gps定位。
涉及修改：整个gps定位存储逻辑

#### 2. `user_group` 存储群组

描述：创建群组，修改群组名和删除群组需要同步缓存。
需要注意：创建群组和删除群组，与`group_member`表相关，在更新缓存的时候，需要配合来自canal服务关于`group_member`表的消息来同步
涉及修改：临时组和普通组的创建销毁


#### 3. `group_member` 群id和设备id对应关系

描述：添加进群组和移动群成员以及和修改群成员的类型需要同步缓存。
需要注意：根据`group_member`这个表的单条记录去反推测，web前端过来的实际请求然后再去更新redis
涉及修改：调度员转移群成员和修改群成员的类型。

#### 4. `user` 账号信息

描述： 创建调度员、导入设备、转移设备、修改设备信息，修改调度员信息、设备切换房间设置默认组需要同步缓存
需要注意：根据canal消息更新缓存的时候，需要根据那一行的记录添加和更新缓存。
涉及修改：创建调度员、导入设备、转移设备、修改设备信息，修改调度员信息

#### 5. `tags`

描述：NFC功能的标签的增删改查需要同步缓存
需要注意：这里的实现是最繁琐的，因为功能实现的时候是用的hash表，得要通过单条mysql记录，查得出hash表，然后再去做增删改查操作
涉及修改：NFC功能的标签的增删改查

#### 6. `task`

描述：NFC功能的标签任务的增删改查需要同步缓存
需要注意：同`tag`表类似这里的实现是最繁琐的，因为功能实现的时候是用的hash表，得要通过单条mysql记录，查得出hash表，然后再去做增删改查操作
涉及修改：NFC功能的标签任务的增删改查

### other

其他关于redis的操作：

1. bo-server服务重启的时候从mysql加载数据到redis。
2. im功能和设备、web用户登录对session的过期时间的管理。