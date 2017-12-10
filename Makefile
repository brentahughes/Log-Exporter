APP_NAME := log-exporter
HTTP_PORT := 8990
RUN_FLAG = -d --restart=always

LOG_DIR := /var/log

LOG_DIR := $(PWD)/test_logs/:/logs/
ifeq ($(shell if [ -f "/var/log/auth.log" ]; then echo yes; fi),yes)
    LOG_DIR := /var/log/:/logs/
endif

REQUEST_LOG_DIR := $(PWD)/test_logs/:/requestLogs/
ifeq ($(shell if [ -f "/var/log/auth.log" ]; then echo yes; fi),yes)
    REQUEST_LOG_DIR := $(HOME)/caddy-web-server/logs:/requestLogs/
endif

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
		-v $(LOG_DIR) \
		-v $(REQUEST_LOG_DIR) \
		$(APP_NAME) \
		-auth /logs/auth.log \
		-geodb /app/geoip.mmdb \
		-request /requestLogs/access.log \
		-exluded.ips 63.143.42.243,63.143.42.242
endef
