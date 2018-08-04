#!/bin/sh

for ((i=1; i<6; i++)); do
	sudo docker run -d \
	--ulimit nofile=98304:98304 \
	--name mqtt-bench-${i} \
	--net=none
	tianzx/mqtt-bench \
	-broker='ssl://msg-dev.app.nio.com:20083' \
	-action='s' \
	-cId="client_ids_${i}" \
	-clients=50000 \
	-tls='client:TlsMobile_10003_dev_dummy1.trustchain,TlsMobile_10003_dev_dummy1.crt,TlsMobile_10003_dev_dummy1.key'
done

for ((i=6; i<11; i++)); do
	sudo docker run -t -i \
	--ulimit nofile=98304:98304 \
	--sysctl net.ipv4.ip_local_port_range="1000 65000" \
	--name mqtt-bench-${i} \
	tianzx/mqtt-bench \
	-broker='ssl://msg-dev.app.nio.com:20083' \
	-action='s' \
	-cId="client_ids_${i}" \
	-clients=50000 \
	-tls='client:TlsMobile_10003_dev_dummy1.trustchain,TlsMobile_10003_dev_dummy1.crt,TlsMobile_10003_dev_dummy1.key'
	sleep 60

done

echo "finish deploy mqtt server performance test ..."
