run:
	docker-compose -f docker-compose.dev.yaml --env-file ./.env up --build
stop:
	docker-compose -f docker-compose.dev.yaml down

migrate:
	 migrate -path=./migrations -database="mongodb://admin:adminpwd@localhost:27017/cdn?authsource=admin" up

repo-mock:
	mockgen -source=./internal/cdn/cdn_repo.go -destination=./internal/cdn/mocks/cdn_repo_mock.go

service-mock:
	mockgen -source ./internal/cdn/cdn_service.go -destination=./internal/cdn/mocks/cdn_service_mock.go

gen-mocks: repo-mock service-mock
