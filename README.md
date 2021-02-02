[![Build Status](https://travis-ci.com/dnataraj/healthbee.svg?token=g7PAjZdpVPTj6UWnsEsA&branch=main)](https://travis-ci.com/dnataraj/healthbee)

HealthBee is a self-hosted website availability monitoring service.

#### System Requirements
* HealthBee has been tested to work on Ubuntu 20.04 LTS with Go 15.2

#### User Guide

The HealthBee API can be used to register any (web) site for availability by periodic monitoring.
Monitoring results are collected and eventually stored for analysis. The API is specified in detail below.

* ```GET /sites``` : Lists all the currently registered sites
* ```POST /sites/{id}/stop``` : Stops monitoring activity for a particular site
* ```POST /sites``` : Register a new site for monitoring
    * The request for site registration can be specified in JSON as follows :
        ```
            {   
                "url": "https://www.google.com", <-- the site address 
                "interval": "4s",  <-- a monitoring interval, in seconds
                "pattern": "content"  <-- an optional regular expression that is searched for in the returned page
            }
        ```
* ```GET /sites/{id}``` will return the last 20 metrics for the given site in JSON 

##### Installation and setup
* HealthBee can be installed on your system using the ```go get``` [command](https://golang.org/pkg/cmd/go/internal/get/), for example
    ```$>go get -u -v github.com/dnataraj/healthbee/cmd/healthbee```
* This will install the latest HealthBee binary to a directory specified in the ```$GOBIN``` env var
  
##### Starting & configuring HealthBee 
Since HealthBee relies on Apache Kafka and PostgreSQL to do its work efficiently and securely, various 
configurations need to be provided to the application at boot time. 

* After a successful installation, the HealthBee application should be in your ```$GOBIN``` or equivalent path
* Executing the application with the -h flag, like so ```$> ./healthbee -h``` will produce a usage description
* The following flags are mandatory
    * ```--brokers``` : The endpoint address for the Kafka service 
    * ```--dsn-string``` : The PostgreSQL database connection string
    * ```--service-cert``` : (For secure communication with Kafka) The Kafka provider public key certificate
    * ```--service-key``` : (For secure communication with Kafka) The Kafka provider private key
    * ```--ca-cert``` : (For secure communication with Kafka) The CA certificate
  
#### Shutting down
* A clean shutdown of HealthBee can be performed by simple hitting Ctrl-C on the foreground process or sending a ```SIGINT``` to
the running process
  
#### Development Notes
* TODO: (High) Move the consumer/auditor functionality into pkg, where it belongs
* TODO: Highlight testing strategy and possibilities - both unit and integration
* TODO: Add support for site & metrics removal

#### Attributions and credits
* Attributions have been mentioned in the source code wherever appropriate but are also listed here
    * https://tylerchr.blog/golang-18-whats-coming/ (on HTTP connection draining)
    * https://github.com/segmentio/kafka-go (on Kafka producer and consumer configuration - esply. for TLS)
    * https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations (working with Durations and JSON)