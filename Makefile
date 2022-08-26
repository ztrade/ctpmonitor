API_PROTO_FILES=$(shell find pb -name *.proto)
.PHONY: api
# generate api proto
api:
	protoc --proto_path=./pb \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./pb \
 	       --go-http_out=paths=source_relative:./pb \
 	       --go-grpc_out=paths=source_relative:./pb \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)
