# init mqtt server performance test...

# 1 step build exec file :
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mqtt-bench ./tls/main.go ./tls/clientId.go

# 2 step build image :
sudo docker build -t tianzx/mqtt-bench ./

# 3 step pull  :
sudo docker push tianzx/mqtt-bench

# 4 step pull image :
sudo docker pull tianzx/mqtt-bench

# 5 step start container :
sudo docker run -d  --ulimit nofile=98304:98304  --sysctl net.ipv4.ip_local_port_range="1000 65000" tianzx/mqtt-bench  -broker='ssl://msg-dev.app.nio.com://:20083' -action='s' -cId='client_ids_1' -clients=100 -tls='client:TlsMobile_10003_dev_dummy1.trustchain,TlsMobile_10003_dev_dummy1.crt,TlsMobile_10003_dev_dummy1.key'


#others useful command:

netstat -nat|grep -i "20083"|wc -l

sudo yum install docker

curl -sSL https://get.daocloud.io/daotools/set_mirror.sh | sh -s http://76cafde5.m.daocloud.io
(then you need edit /etc/docker/daemon.json to delete comma)

sudo service docker start

sudo docker ps -a | awk '{print $1}'|sudo xargs  docker stop

sudo docker ps -a | awk '{print $1}'|sudo xargs docker rm

sudo systemctl restart  message_server.service

sudo systemctl start  message_server.service

sudo systemctl stop  message_server.service

sudo systemctl status  message_server.service

sudo nohup /usr/bin/java -server -XX:+HeapDumpOnOutOfMemoryError -Djava.awt.headless=true -Xms3g -Xmx5g -Dlogback.configurationFile="/data/app/greatwall_messaging_server/conf/msg_server_logback.xml" -Dnextev_msg.home="/data/app/greatwall_messaging_server" -cp "/data/app/greatwall_messaging_server/lib/*" com.nio.message.server.MessageServer &

sudo docker exec -it ${containerId} /bin/bash
#ref:

This is the TCP server-client suit to help you test if your OS supports c1000k(1 million connections).
(https://github.com/ideawu/c1000k)


