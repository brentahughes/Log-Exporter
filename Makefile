APP_NAME := log-exporter
HTTP_PORT := 8990
RUN_FLAG = -d --restart=always

LOG_DIR := /var/log

build:
	docker build -t $(APP_NAME) .

debug: build remove_container
	$(eval RUN_FLAG = --rm)
	$(call run)

deploy: build remove_container
	$(call run)

remove_container:
	-docker rm -f $(APP_NAME)

define run
	docker run $(RUN_FLAG) --name $(APP_NAME) -p $(HTTP_PORT):9090 \
		-v $(LOG_DIR):/logs
		-v /etc/localtime:/etc/localtime:ro \


		$(APP_NAME)
endef
