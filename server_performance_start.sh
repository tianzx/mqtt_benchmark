#!/bin/sh
echo "init mqtt server performance test..."

echo "1 step build exec file :"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mqtt-bench ./tls/main.go ./tls/clientId.go
echo "2 step build image :"
sudo docker build -t tianzx/mqtt-bench ./
echo "3 step pull image :"
sudo docker push tianzx/mqtt-bench
echo "4 step pull image :"
sudo docker pull tianzx/mqtt-bench
echo "5 step start container :"

for ((i=1; i<3; i++)); do
	sudo docker run -t -i \
	--ulimit nofile=98304:98304 \
	--name mqtt-bench-${i} \
	tianzx/mqtt-bench \
	-broker='ssl://msg-dev.app.nio.com:20083' \
	-action='s' \
	-cId="client_ids_${i}" \
	-clients=10 \
	-tls='client:TlsMobile_10003_dev_dummy1.trustchain,TlsMobile_10003_dev_dummy1.crt,TlsMobile_10003_dev_dummy1.key'
done

echo "finish deploy mqtt server performance test ..."
