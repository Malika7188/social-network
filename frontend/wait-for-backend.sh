#!/bin/sh

# Wait for the backend to be available
echo "Waiting for backend at http://backend:8080..."

while ! wget --spider -q http://backend:8080/api/auth/health; do
  sleep 4
done

echo "Backend is up. Starting frontend..."

# Start the frontend app
npm run dev:frontend
