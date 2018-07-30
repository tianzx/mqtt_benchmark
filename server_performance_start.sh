#!/bin/sh
echo "init mqtt server performance ..."

for ((i=1; i<10; i++)); do
	sudo docker run -t \
	--ulimit nofile=98304:98304 \
	--name mqtt-bench-${i} \
	tianzx/mqtt-bench \
	-broker='ssl://msg-dev.app.nio.com:20083' \
	-action='s' \
	-cId="client_ids_${i}" \
	-clients=50000 \
	-tls='client:TlsMobile_10003_dev_dummy1.trustchain,TlsMobile_10003_dev_dummy1.crt,TlsMobile_10003_dev_dummy1.key'
done

echo "finish deploy mqtt server performance ..."
