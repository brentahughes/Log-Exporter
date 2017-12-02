# Log-Exporter
Simple service for collecting metrics on log files

CURRENTLY ONLY SUPPORT auth.log

### NOTICE
This will add a label for each hostname, ip_address, process, type, and user which can result in a very large number of metrics to track in prometheus. If your server gets a ton of auth attempts you may want to give prometheus more resources or lower the data retention.
