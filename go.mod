module xinu.tv/surfer

go 1.12

require (
	github.com/andybalholm/cascadia v1.0.0
	github.com/beorn7/perks v1.0.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.4.1 // indirect
	github.com/prometheus/procfs v0.0.0-20190528151240-3cb620ac02d0 // indirect
	github.com/wathiede/surfer v0.0.0-20170716220928-7a38cd99e40c
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
)

replace github.com/prometheus/client_golang => ./vendor/github.com/prometheus/client_golang
