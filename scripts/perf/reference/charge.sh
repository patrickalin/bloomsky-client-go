echo go-wrk
go get github.com/adjust/go-wrk
go-wrk -n 5 http://localhost:1111 > perf.0
# 1 minute
