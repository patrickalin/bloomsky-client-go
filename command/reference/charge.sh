echo go-wrk
go get github.com/adjust/go-wrk
go-wrk -n 5 http://localhost:1111 > command/reference/perf.0
# 1 minute