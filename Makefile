# Tentukan variabel-variabel yang mungkin Anda perlukan
APP_NAME=./tmp/main
MAIN_PACKAGE=./webcore/main.go
DOCKER_COMPOSE_FILE=docker-compose.dev-min.yml
DOCKER_PACKAGE=kemenkesri/konsolidator-ckg-tb:latest
DOCKER_CR=docker.io
TEST_FLAGS=-v -race

# Target default: menjalankan aplikasi dalam mode development dengan air dan docker-compose
all: clean sync vet run

# Target untuk sinkronisasi workspace Go
sync:
	@echo "Running go work sync..."
	@go work sync

# Target untuk debig aplikasi
debug:
	@echo "Debug the application with air..."
	@air -c .air-direct.toml

# Target untuk menjalankan aplikasi
run: sync
	@echo "Running the application..."
	@go run webcore/main.go

# Target untuk test pengujian
test:
	@echo "Running tests..."
	@go test $(TEST_FLAGS) ./...

# Target untuk cek vurnerability analisis kode
vet:
	@echo "Running VET..."
	@go vet $(MAIN_PACKAGE)

# Target untuk membangun aplikasi
build: sync
	@echo "Building the application..."
	@go build -o $(APP_NAME) $(MAIN_PACKAGE)

# Target untuk menjalankan docker-compose.dev-min.yml
docker-up:
	@echo "Starting services with docker-compose..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

# Target untuk menghentikan container docker-compose
docker-down:
	@echo "Stopping services with docker-compose..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down

# Target untuk build docker image siap production
docker-build:
	@echo "Build docker image $(DOCKER_PACKAGE)..."
	@docker build --target=production --platform linux/amd64 -t $(DOCKER_PACKAGE) --no-cache .

# Target untuk push docker image ke container registry
docker-push:
	@echo "Push docker image $(DOCKER_PACKAGE) to $(DOCKER_CR)..."
	@docker push $(DOCKER_CR)/$(DOCKER_PACKAGE)

# Target untuk membersihkan build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(APP_NAME)

# Target phony: mendeklarasikan target yang bukan file
.PHONY: all sync debug run test build docker-up docker-down docker-build docker-push clean
