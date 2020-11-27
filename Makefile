TAG := $(shell git log -1 --pretty=%h)

build:
	@echo "Building docker image..."
	docker build -t ardenn/feedr:${TAG} --network host .

migrate-up:
	migrate -database ${DATABASE_URL} -path migrations up

migrate-down:
	migrate -database ${DATABASE_URL} -path migrations down