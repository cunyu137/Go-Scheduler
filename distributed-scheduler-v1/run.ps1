docker compose up -d
Start-Sleep -Seconds 8
go mod tidy
go run ./cmd/server
