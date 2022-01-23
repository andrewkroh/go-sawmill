package protobuf

//go:generate go install github.com/gogo/protobuf/protoc-gen-gofast
//go:generate protoc --gofast_out=. event.proto
