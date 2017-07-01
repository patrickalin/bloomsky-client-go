echo go-torch
go get github.com/uber/go-torch
# git clone git@github.com:brendangregg/FlameGraph.git
go-torch -t 5 -u http://localhost:1111
open torch.svg
