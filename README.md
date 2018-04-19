# maxmind
A GeoIP server and lookup tool written in GO, and pulling ideas from:

- https://github.com/klauspost/geoip-service
- https://github.com/twisted1919/geoip-go
- https://github.com/oschwald/geoip2-golang    
- https://github.com/mitchellh/cli

### Install  
```bash
$ go ge t -u github.com/rabbitt/maxmind
$ cd $GOPATH/src/github.com/rabbitt/maxmind
$ make bootstrap
$ make maxmind
```

### Runtime Requirements
- [GeoLite2-City.mmdb](https://dev.maxmind.com/geoip/geoip2/geolite2/#Downloads) from MaxMind

### Usage

This package provides a single binary `maxmind` providing both, `lookup`, and `server` functionality. Basic help
information can be gleaned by typing `maxmind --help`. More specific help information can be found by issuing
`--help` to the relevant subcommand.

#### The Lookup tool

To list GeoIP data for one, or more, IPs, use the following:
```bash
$ maxmind lookup -f <path to GeoLite2-City.mmdb> 123.123.123.123 8.8.8.8 8.8.4.4

[ 123.123.123.123 ]---------------------->
  Continent:      [Asia (AS)]
  Country:        [China (CN)]
  Subdivision:    [Beijing (BJ)]
  City:           [Beijing]
  Location:
    Coordinates:  [39.9289, 116.3883 (20)]
    Timezone:     [Asia/Shanghai]

[         8.8.8.8 ]---------------------->
  Continent:      [North America (NA)]
  Country:        [United States (US)]
  Location:
    Coordinates:  [37.7510, -97.8220 (1000)]

[         8.8.4.4 ]---------------------->
  Continent:      [North America (NA)]
  Country:        [United States (US)]
  Location:
    Coordinates:  [37.7510, -97.8220 (1000)]
```

#### The Server

The server has three routes that it listens for requests on:
* `GET /ping`   - responds with 200 and pong
* `HEAD /ping`  - responds with 200 only
* `GET /ip/:ip` - responds with geodata (as JSON) for the requested ip

##### Configuration

Configuration can be passed on the command line, or put into a JSON encoded file
using the same long form names as those of the options listed below:

<dl>
  <dt>-c, --config.file <file></dt>
  <dd>Path to JSON encoded file containing server configuration.</dd>

  <dt>-i, --sever.ip <file></dt>
  <dd>IP Address for the service to bind to (default: 127.0.0.1)</dd>

  <dt>-p, --server.port <file></dt>
  <dd>Port for service to bind to (default: 8000)</dd>

  <dt>-f, --database.file <file></dt>
  <dd>Path to the MaxMind database file (default: /var/lib/maxminddb/GeoLite2-City.mmdb)</dd>

  <dt>-t, --cache.ttl <file></dt>
  <dd>How long to rcache response data before refetching it from the database (0 disables, default: 3600.0)</dd>

  <dt>-T, --worker.threads <file></dt>
  <dd>Number of worker threads to handle incoming requests (default: Number of CPU processes)</dd>
</dl>

Starting up requires, at a minimum, the path to the MaxMind database file:

```bash
$ bin/maxmind server -f /var/lib/maxminddb/GeoLite2-City.mmdb
INFO: Configuration:
INFO:     Bind Address:   [ 127.0.0.1:8000 ]
INFO:     Cache TTL:      [ 3600.00 seconds ]
INFO:     Worker Threads: [ 8 ]
INFO:     Database File:  [ /private/var/lib/maxminddb/GeoLite2-City.mmdb ]
INFO: Caching enabled; will cache requests for 3600.00 seconds
INFO: Listening on 127.0.0.1:8000 ...
```

###### Example Output
Example output from a client request for ip `123.123.123.123`:

```javascript
# $ curl -s http://127.0.0.1:8000/ip/123.123.123.123 | jq
{
  "status": "success",
  "message": "OK",
  "data": {
    "city": {
      "name": "Beijing"
    },
    "continent": {
      "code": "AS",
      "name": "Asia"
    },
    "country": {
      "iso_code": "CN",
      "name": "China"
    },
    "location": {
      "accuracy_radius": 20,
      "latitude": 39.9289,
      "longitude": 116.3883,
      "time_zone": "Asia/Shanghai"
    },
    "postal": {},
    "registered_country": {
      "iso_code": "CN",
      "name": "China"
    },
    "represented_country": {},
    "subdivisions": [
      {
        "iso_code": "BJ",
        "name": "Beijing"
      }
    ],
    "subdivision": {
      "iso_code": "BJ",
      "name": "Beijing"
    },
    "traits": {}
  }
}
```
