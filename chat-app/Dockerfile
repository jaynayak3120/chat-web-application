FROM golang:1.22.1-alpine3.18
WORKDIR /test
COPY . /test
RUN go build /test
EXPOSE 8080
ENTRYPOINT [ "./chat-app" ]