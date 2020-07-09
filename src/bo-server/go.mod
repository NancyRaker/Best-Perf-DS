module bo-server

go 1.12

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.39.0

	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190513172903-22d7a77e9e5f
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190510132918-efd6b22b2522
	golang.org/x/image => github.com/golang/image v0.0.0-20190523035834-f03afa92d3ff
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20190509164839-32b2708ab171
	golang.org/x/net => github.com/golang/net v0.0.0-20190522155817-f3200d17e092
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20190523182746-aaccbc9213b0
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190526052359-791d8a0f4d09
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/time => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => github.com/golang/tools v0.0.0-20190525145741-7be61e1b0e51

	google.golang.org/api => github.com/googleapis/google-api-go-client v0.5.0
	google.golang.org/appengine => github.com/golang/appengine v1.6.0
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20190522204451-c2c4e71fbf69
	google.golang.org/grpc => github.com/grpc/grpc-go v1.21.0
)

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/CanalClient/canal-go v0.0.0-20191011031409-2db66ad8cdb3
	github.com/FZambia/sentinel v1.0.0
	github.com/Shopify/sarama v1.23.1
	github.com/ceshihao/ratelimiter v0.0.0-20190523025428-602ec2aa595f
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/go-ini/ini v1.42.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.0
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/jordan-wright/email v0.0.0-20190819015918-041e0cec78b0
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/olivere/elastic v6.2.21+incompatible // indirect
	github.com/olivere/elastic/v7 v7.0.4 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/panjf2000/ants v1.0.0
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.4.2
	github.com/smartwalle/dbs v1.0.0
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a // indirect
	github.com/tebeka/strftime v0.0.0-20140926081919-3f9c7761e312 // indirect
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	google.golang.org/grpc v1.19.0
	gopkg.in/ini.v1 v1.44.0 // indirect
)
