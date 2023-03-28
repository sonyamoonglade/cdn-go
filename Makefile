run-dev:
	./docker/run.dev.sh

stop-dev:
	./docker/stop.dev.sh

test-ci: mocks
	go test -race -short -count=3 ./...

REPO_SRC=./internal/cdn/cdn_repo.go
SERVICE_SRC=./internal/cdn/cdn_service.go
MODULE_CONTROLLER_SRC=./internal/modules/controller.go

CDN_MOCKS_DST=./internal/cdn/mocks
MODULES_MOCKS_DST=./internal/modules/mocks

premock:
	rm -rf ${CDN_MOCKS_DST} ${MODULES_MOCKS_DST} && \
	mkdir -p ${CDN_MOCKS_DST} && \
	mkdir -p ${MODULES_MOCKS_DST}

mocks: premock
	mockgen -source ${REPO_SRC} -destination ${CDN_MOCKS_DST}/cdn_repo_mock.go && \
	mockgen -source=${SERVICE_SRC} -destination=${CDN_MOCKS_DST}/cdn_service_mock.go && \
	mockgen -source=${MODULE_CONTROLLER_SRC} -destination=${MODULES_MOCKS_DST}/controller_mock.go
