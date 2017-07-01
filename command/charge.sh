echo go-wrk
go get github.com/adjust/go-wrk
go-wrk -n 10000 http://localhost:1111
# 1 minute