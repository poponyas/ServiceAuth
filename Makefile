grpc-codegen:
	@protoc --proto_path=api/grpc \
		--go_out=./gen/grpc/auth --go_opt=paths=source_relative \
		--go-grpc_out=./gen/grpc/auth --go-grpc_opt=paths=source_relative \
		auth.proto