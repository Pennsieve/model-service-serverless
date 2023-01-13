
mod:
	# This make rule requires Go 1.11+
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

test:
	docker-compose -f docker-compose.test.yml up --build --exit-code-from api_tests
	docker-compose -f docker-compose.test.yml down --volumes