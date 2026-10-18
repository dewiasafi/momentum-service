# momentum-service

## migrations database in folder cmd/migrations
export $(grep -v '^#' .env | xargs)
migrate -path ./cmd/migrations -database "$DATABASE_URL" up