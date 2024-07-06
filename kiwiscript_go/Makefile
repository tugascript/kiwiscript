# Copyright (C) 2024 Afonso Barracha
# 
# This file is part of KiwiScript.
# 
# KiwiScript is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# 
# KiwiScript is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.

BUILD_DIR=target
DATABASE_URL=postgres://postgres:postgres@localhost:5432/kiwiscript?sslmode=disable
.DEFAULT_GOAL := run

build:
	GOARCH=amd64 GOOS=darwin go build -o ./$(BUILD_DIR)/kiwiscript_macos_x86 main.go
	GOARCH=arm64 GOOS=darwin go build -o ./$(BUILD_DIR)/kiwiscript_macos_arm main.go
	GOARCH=amd64 GOOS=linux go build -o ./$(BUILD_DIR)/kiwiscript_linux main.go
	GOARCH=amd64 GOOS=windows go build -o ./$(BUILD_DIR)/kiwiscript_windows.exe main.go

dc-up:
	docker-compose -f ../docker-compose.yaml up -d

pm-up:
	podman-compose -f ../docker-compose.yaml up -d

dc-down:
	docker-compose -f ../docker-compose.yaml down

pm-down:
	podman-compose -f ../docker-compose.yaml down

create-db:
	docker exec -it postgres psql -U postgres -c "CREATE DATABASE kiwiscript;"

create-pm-db:
	podman exec -it postgres psql -U postgres -c "CREATE DATABASE kiwiscript;"

drop-db:
	docker exec -it postgres psql -U postgres -c "DROP DATABASE kiwiscript;"

drop-pm-db:
	podman exec -it postgres psql -U postgres -c "DROP DATABASE kiwiscript;"

migrate-up:
	migrate -path ./providers/database/migrations -database "$(DATABASE_URL)" -verbose up

migrate-down:
	migrate -path ./providers/database/migrations -database "$(DATABASE_URL)" -verbose down

sqlc:
	sqlc generate

server:
	go run main.go

dev:
	air

keygen:
	DEBUG=false go run ../keygen/main.go

keygen-debug:
	DEBUG=true go run ../keygen/main.go

test-verbose:
	go test -v ./tests

test:
	go test ./tests