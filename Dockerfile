FROM ubuntu
MAINTAINER tianzx@aliyun.com

RUN mkdir  -p /data/app
CMD sysctl -w net.ipv4.ip_local_port_range="1000 65500";
#RUN echo 'net.ipv4.ip_local_port_range = 8001 65000' >> /etc/sysctl.conf
WORKDIR /data/app
RUN mkdir /data/app/log
VOLUME /log:/data/app/log
COPY ./mqtt-bench /data/app/
COPY ./tls/clientIds/client_ids_* /data/app/
COPY ./tls/TlsMobile_10003_dev_dummy1.* /data/app/
RUN ls -l /data/app
# install dependent packages
# RUN go get github.com/eclipse/paho.mqtt.golang
# RUN go get github.com/golang/net
# RUN mkdir  -p /data/app
#begin compile
RUN cd /data/app
#RUN go build -o mqtt-bench
#end compile

#docker 启动参数

ENTRYPOINT ["./mqtt-bench"]
