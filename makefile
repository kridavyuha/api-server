VOLUME_PATH=/Users/rithvik/Documents/Strategic\ Fantasy\ League/App/backend/volumes


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

