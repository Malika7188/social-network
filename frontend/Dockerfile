FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application
COPY . .


# Expose port 3000
EXPOSE 3000

# Copy the wait script
COPY wait-for-backend.sh .

# Make it executable
RUN chmod +x wait-for-backend.sh

# Start using the wait script
CMD ["./wait-for-backend.sh"]
