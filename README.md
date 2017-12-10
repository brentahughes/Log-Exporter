# Log-Exporter
Simple service for collecting metrics on log files

CURRENTLY ONLY SUPPORT auth.log

### NOTICE
This will add a label for each hostname, ip_address, process, type, and user which can result in a very large number of metrics to track in prometheus. If your server gets a ton of auth attempts you may want to give prometheus more resources or lower the data retention.


## Usage

`./log-exporter -auth.path /path/to/auth.log -request.path /path/to/access.log`

By default metrics will be available at localhost:9090/metrics. This can be changed by using the `-prometheus.port` and `-prometheus.endpoint` flags for your needs.

### Request Log Format
I peronsally proxy all http reqeusts through caddy resulting in a single access.log. This also means my access log format will likely be different from yours. You can use the `-request.regexMatch` flag to set your parser.

*My Access Log Format* [{when}] [{host}] [{remote}] [{status}] [{method}] {uri}"
*The Parser I use* ^\\[.* .0000\\] \\[(?P<domain>.*)\\] \\[(?P<ip_address>[0-9\\.]+)\\] \\[(?P<status>\\d{3})\\] \\[(?P<method>\\w+)\\] .*$
    - Notice I am using named groups in my regex. Yours will require the same for at least `domain`, `ip_address`, `status`, and `method`. Any others will be ignored.

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
