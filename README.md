# Log-Exporter
Simple service for collecting metrics on log files

CURRENTLY ONLY SUPPORT auth.log

### NOTICE
This will add a label for each hostname, ip_address, process, type, and user which can result in a very large number of metrics to track in prometheus. If your server gets a ton of auth attempts you may want to give prometheus more resources or lower the data retention.


## Usage

`./log-exporter -auth /path/to/auth.log`

By default metrics will be available at localhost:9090/metrics. This can be changed by using the `-port` and `-endpoint` flags for your needs.

### Geo IP
For location metrics based in the IP addresses found in the log you must have the geoip2 db downloaded somehwere the app can see it.

![GeoIP2 Lite](https://dev.maxmind.com/geoip/geoip2/geolite2/)

Extract mmdb file into the same directory as log-exporter

`./log-exporter -auth /path/to/auth.log -geodb /path/to/geoip2.mmdb`


### Debugging
Use the `-debug` flag to proccess the entire log. This will help scan full file and identify any issues

## Screenshots

![GeoIP Map](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/geoip_map.png)
![Auth Attempts](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/auth_attempts.png)
![Invalid Users](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/invalid_users.png)
![By IP Address](https://raw.githubusercontent.com/bah2830/Log-Exporter/master/images/by_ip_address.png)
