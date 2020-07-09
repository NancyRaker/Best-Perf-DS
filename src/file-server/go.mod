module file-server

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
	github.com/Shopify/sarama v1.23.1
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/go-ini/ini v1.42.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.4.2
	github.com/ugorji/go/codec v0.0.0-20190320090025-2dc34c0b8780 // indirect
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c // indirect
	google.golang.org/appengine v1.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)
