up-build:
	docker-compose up --build

up:
	docker-compose up

lint:
	golangci-lint run

.PHONY: easyjson
easyjson:
	@echo "Генерация easyjson..."
	@for file in $$(find ./internal/entity/dto -name '*.go' | grep -v "_easyjson.go"); do \
		easyjson -all $$file; \
	done
	@echo "Генерация завершена"

.PHONY: perf-test-report
perf-test-report:
	wrk -t4 -d60m -s db/perf_test/load_data.lua http://localhost:8000 > db/perf_test/result.csv 
	
	wrk -t4 -d60m -s perf_test/get_data.lua http://localhost:8000 > db/perf_test/getresult.csv
	