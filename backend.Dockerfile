FROM golang:1.25-alpine

WORKDIR /app

# Copier les fichiers de dépendances
COPY iot-backend/go.mod iot-backend/go.sum ./
RUN go mod download

# Copier tout le code
COPY iot-backend/ .

# Compiler
RUN go build -o main ./cmd/api/main_secure.go

EXPOSE 8080

CMD ["./main"]
