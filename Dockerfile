# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend-go/go.mod backend-go/go.sum ./backend-go/
RUN cd backend-go && go mod download
COPY backend-go/ ./backend-go/
COPY --from=frontend /app/dist ./backend-go/dist
RUN cd backend-go && CGO_ENABLED=0 go build -ldflags="-s -w" -o questionnaire-backend .

# Stage 3: Runtime
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=backend /app/backend-go/questionnaire-backend .
ENV PORT=3000
ENV DB_PATH=./data/data.db
EXPOSE 3000
VOLUME ["/app/data"]
CMD ["./questionnaire-backend"]
