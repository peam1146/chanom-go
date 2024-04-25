FROM --platform=amd64 golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/main .

CMD [ "/app/main" ]
