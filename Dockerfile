FROM golang:1.23 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o migrate ./migrate/
RUN go build -o script ./scripts/
RUN go build -o main


FROM golang:1.23 AS runtime

WORKDIR /app
COPY --from=build /app/main /app/migrate /app/script /app/.env ./

CMD ./migrate && ./script auto && ./main
# CMD tail -f /dev/null