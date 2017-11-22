# Shamelessly adapted from the Makefile at vieux/docker-volume-sshfs

PLUGIN_NAME=LoggingDriverTest
PLUGIN_TAG=master

all: push

clean:
	@echo "### Removing the ./plugin directory"
	@rm -rf ./plugin

docker:
	@echo "### Building the plugin binary"
	@docker build -t builder -f Dockerfile.binary .
	@echo "### Copy the plugin binary"
	@docker create --name tmp builder
	@docker cp tmp:/go/bin/LoggingDriverTest .
	@docker rm -fv tmp
	@docker rmi builder
	@echo "### Create the rootfs image"
	@docker build -t ${PLUGIN_NAME}:rootfs .

rootfs:
	@echo "### Create rootfs directory in ./plugin/rootfs"
	@mkdir -p ./plugin/rootfs
	@docker create --name rootfs-tmp ${PLUGIN_NAME}:rootfs
	@docker export rootfs-tmp | tar -x -C ./plugin/rootfs
	@echo "### Copy config.json to ./plugin/"
	@cp config.json ./plugin/
	@docker rm -fv rootfs-tmp

create:
	@echo "### Remove the ${PLUGIN_NAME} plugin from Docker"
	@docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	@echo "### Create the ${PLUGIN_NAME}:${PLUGIN_TAG} plugin from the contents of ./plugin/"
	@docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./plugin

enable:
	@echo "### Enabling the ${PLUGIN_NAME}:${PLUGIN_TAG} plugin"
	@mkdir -p /var/log/LoggingDriverTest/
	@docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

push: clean docker rootfs create enable
	@echo "### Push the ${PLUGIN_NAME}:${PLUGIN_TAG} plugin to the repository"
	@docker tag ${PLUGIN_NAME}:${PLUGIN_TAG} localhost:5000/${PLUGIN_NAME}:${PLUGIN_TAG}
	@docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
