#!/bin/sh

# Wait for database to be ready
echo "Waiting for database to be ready..."
while ! nc -z db 5432; do
  sleep 1
done
echo "Database is ready!"

# Give database a moment to fully initialize
sleep 2

# Run migrations
echo "Running database migrations..."
migrate -path /app/migrations -database "postgresql://user:password@db:5432/dashtrack?sslmode=disable" up

if [ $? -eq 0 ]; then
    echo "Migrations completed successfully!"
else
    echo "Migration failed!"
    exit 1
fi

# Start the application
echo "Starting application..."
exec "$@"
