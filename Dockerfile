FROM golang:1.20 as serverbuilder
WORKDIR /code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM scratch
COPY --from=serverbuilder /code/main  /server/main

WORKDIR  /server

CMD ["./main"]
