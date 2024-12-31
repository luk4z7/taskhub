FROM golang:1.23.1 as build

ARG APP_NAME_ARG
ENV APP_NAME=$APP_NAME_ARG

WORKDIR /home

ADD . .

RUN go build -o bin ${APP_NAME}

CMD ["/home/bin"]
