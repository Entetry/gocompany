compose:
	docker-compose build && docker-compose up -d
compose-down:
	docker-compose down

proto:
	protoc --proto_path=protocol protocol/*.proto --go_out=./protocol --go-grpc_out=./protocol --govalidators_out=./protocol
