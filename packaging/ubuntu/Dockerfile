# This file describes how to build buildbot into a runnable linux container with all dependencies installed
# To build:
#
# Install docker (http://docker.io)
#
# # Download buildbot Dockerfile
# wget https://raw.github.com/mzdaniel/buildbot/master/Dockerfile
#
# # Build buildbot image
# IMAGE_ID=$(docker build < Dockerfile | tail -1)
#
# # Run buildbot
# CONTAINER_ID=$(docker run -d $IMAGE_ID)
# CONTAINER_IP=$(docker inspect $CONTAINER_ID | sed -En 's/.+"IpAddress": "(.+)".+/\1/p')
# wget -qO- $CONTAINER_IP:8010
#
# VERSION       0.1
# DOCKER-VERSION    0.3.2

# Base docker image
from    ubuntu:12.04

# Make dpkg happy with the upstart issue
run dpkg-divert --local --rename --add /sbin/initctl
run ln -s /bin/true /sbin/initctl

# Install buildbot and its dependencies
run echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list
run apt-get update
run DEBIAN_FRONTEND=noninteractive apt-get install -y python-pip python-dev supervisor git sudo ssh
run pip install sqlalchemy==0.7.9 buildbot buildbot_slave

# Set ssh superuser buildbot with password buildbot
run mkdir /data /var/run/sshd
run useradd -m -d /data/buildbot -p sacH4vLqY4BaM buildbot
run sed -Ei 's/adm:x:4:/admin:x:4:buildbot/' /etc/group
run sed -Ei 's/(\%admin ALL=\(ALL\) )ALL/\1 NOPASSWD:ALL/' /etc/sudoers

# Create buildbot configuration
run cd /data/buildbot; sudo -u buildbot sh -c "buildbot create-master master"
run cp -a /data/buildbot/master/master.cfg.sample /data/buildbot/master/master.cfg
run cd /data/buildbot; sudo -u buildbot sh -c "buildslave create-slave slave localhost:9989 example-slave pass"

# Set supervisord buildbot and sshd processes
run /bin/echo -e "[program:buildmaster]\ncommand=twistd --nodaemon --no_save -y buildbot.tac\ndirectory=/data/buildbot/master\nuser=buildbot\n\n[program:buildworker]\ncommand=twistd --nodaemon --no_save -y buildbot.tac\ndirectory=/data/buildbot/slave\nuser=buildbot\n" >/etc/supervisor/conf.d/buildbot.conf

run /bin/echo -e "[program:sshd]\ncommand=/usr/sbin/sshd -D\n" >/etc/supervisor/conf.d/sshd.conf

# Setup running docker container buildbot process
expose 8010
cmd ["/usr/bin/supervisord", "-n"]

# https://raw.github.com/mzdaniel/buildbot/master/Dockerfile
