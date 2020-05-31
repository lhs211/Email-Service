FROM golang:latest
MAINTAINER Luke Sykes
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get github.com/gorilla/mux
EXPOSE 8888
CMD ["go", "run", "/app/bluebook.go"]