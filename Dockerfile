FROM golang:latest
EXPOSE 8887
RUN mkdir /root/.config
ADD .config /root/.config
ENV SRC_DIR=${PWD}:/app
COPY . ${SRC_DIR}
WORKDIR ${SRC_DIR}
CMD go run app/main.go
