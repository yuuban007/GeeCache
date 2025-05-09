module example

go 1.22.0

require (
	google.golang.org/protobuf v1.36.6
	mycache v0.0.0
)

require github.com/golang/protobuf v1.5.4 // indirect

replace mycache => ./mycache
