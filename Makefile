run:
	./docker/run.dev.sh

stop:
	./docker/stop.dev.sh

prom:
	docker run --name prometheus --rm -v $(pwd)/prometheus.yml:/etc/prometheus -p 9090:9090 prom/prometheus

migrate:
	 migrate -path=./migrations -database="mongodb://admin:adminpwd@localhost:27017/cdn?authsource=admin" up

test-ci:
	go test -race -short ./...

#mocks
REPO_SRC=./internal/cdn/cdn_repo.go
SERVICE_SRC=./internal/cdn/cdn_service.go
CDN_MOCKS_DST=./internal/cdn/mocks

MODULE_CONTROLLER_SRC=./internal/modules/controller.go
MODULES_MOCKS_DST=./internal/modules/mocks
mocks:
	mockgen -source ${REPO_SRC} -destination ${CDN_MOCKS_DST}/cdn_repo_mock.go && \
	mockgen -source=${SERVICE_SRC} -destination=${CDN_MOCKS_DST}/cdn_service_mock.go && \
	mockgen -source=${MODULE_CONTROLLER_SRC} -destination=${MODULES_MOCKS_DST}/controller_mock.go
