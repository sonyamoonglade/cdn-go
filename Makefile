run:
	docker-compose -f docker-compose.dev.yaml --env-file ./.env up --build
stop:
	docker-compose -f docker-compose.dev.yaml down

migrate:
	 migrate -path=./migrations -database="mongodb://admin:adminpwd@localhost:27017/cdn?authsource=admin" up

#mocks
REPO_SRC=./internal/cdn/cdn_repo.go
SERVICE_SRC=./internal/cdn/cdn_service.go
CDN_MOCKS_DST=./internal/cdn/mocks
mocks:
	mockgen -source ${REPO_SRC} -destination ${CDN_MOCKS_DST}/cdn_repo_mock.go && \
	mockgen -source=${SERVICE_SRC} -destination=${CDN_MOCKS_DST}/cdn_service_mock.go 
