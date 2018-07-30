FROM ubuntu

# 设置locale
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8
ENV TZ=Asia/Shanghai
ENV GO_ENV=prod

RUN mkdir  -p /data/app
RUN echo 'net.ipv4.ip_local_port_range = 8001 65000' >> /etc/sysctl.conf
WORKDIR /data/app
COPY ./tls/mqtt-bench /data/app/
COPY ./tls/clientIds/client_ids_* /data/app/
COPY ./tls/TlsMobile_10003_dev_dummy1.* /data/app/
#COPY ./tls/config/sysctl.conf /etc/
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