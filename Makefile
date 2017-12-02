# Shamelessly adapted from the Makefile at vieux/docker-volume-sshfs

PLUGIN_NAME=docker-plugin
PLUGIN_TAG=latest

all: clean docker create enable

clean:
	@echo "### Removing the ./plugin directory"
	@rm -rf ./plugin

docker:
	@echo "### Building the plugin binary"
	@docker build -t builder -f Dockerfile .
	@echo "### Create builder container"
	@docker create --name tmp builder
	@echo "### Copy config.json & roofs to plugin root"
	@mkdir -p ./plugin/rootfs
	@docker cp tmp:/ ./plugin/rootfs/
	@cp config.json ./plugin/
	@docker rm -f tmp
	@docker rmi builder

create:
	@echo "### Remove the ${PLUGIN_NAME} plugin from Docker"
	@docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	@echo "### Create the ${PLUGIN_NAME}:${PLUGIN_TAG} plugin from the contents of ./plugin/"
	@docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./plugin

enable:
	@echo "### Enabling the ${PLUGIN_NAME}:${PLUGIN_TAG} plugin"
	@docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

push: clean docker rootfs create enable
	@echo "### Push the ${PLUGIN_NAME}:${PLUGIN_TAG} plugin to the repository"
	@docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
