module github.com/wathiede/surfer

go 1.12

require (
	github.com/andybalholm/cascadia v1.3.1
	github.com/go-kit/kit v0.8.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/glog v1.0.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515 // indirect
	github.com/pkg/errors v0.8.0 // indirect
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/sirupsen/logrus v1.2.0 // indirect
	golang.org/x/net v0.7.0
)

replace github.com/prometheus/client_golang => ./vendor/github.com/prometheus/client_golang
