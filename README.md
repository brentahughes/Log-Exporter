# Log-Exporter
Simple service for collecting metrics on log files

CURRENTLY ONLY SUPPORT auth.log

### NOTICE
This will add a label for each hostname, ip_address, process, type, and user which can result in a very large number of metrics to track in prometheus. If your server gets a ton of auth attempts you may want to give prometheus more resources or lower the data retention.


## Usage

`./log-exporter -auth /path/to/auth.log`

By default metrics will be available at localhost:9090/metrics. This can be changed by using the `-port` and `-endpoint` flags for your needs.


## Screenshots

![Auth Attempts](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/auth_attempts.png)
![Invalid Users](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/invalid_users.png)
![By IP Address](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/by_ip_address.png)
