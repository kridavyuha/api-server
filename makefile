VOLUME_PATH=$(shell pwd)/volume

start_db:
	docker run --name db -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres \
	-v $(VOLUME_PATH):/var/lib/postgresql/data \
	-p 5432:5432 \
	-d postgres

stop_db:
	docker stop db
	docker rm db

restart_db: stop_db start_db

start_redis:
	docker run --name redis -p 6379:6379 -d redis

stop_redis:
	docker stop redis
	docker rm redis

restart_redis: stop_redis start_redis

start_mq:
	docker run -d --name rabbitmq -p 5672:5672-p 15672:15672 rabbitmq:management

stop_mq:
	docker stop rabbitmq
	docker rm rabbitmq

restart_mq: stop_mq start_mq

build_amd64:
	GOOS=linux GOARCH=amd64 go build -o myapp_amd64 cmd/api-server/*.go


