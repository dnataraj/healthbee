[![Build Status](https://travis-ci.com/dnataraj/healthbee.svg?token=g7PAjZdpVPTj6UWnsEsA&branch=main)](https://travis-ci.com/dnataraj/healthbee)

HealthBee is a website availability monitoring service.

#### System Requirements
* HealthBee has been tested to work on Ubuntu 20.04 LTS with Golang 15.2

#### User Guide

The HealthBee API can be used to register any (web) site for monitoring. The API is further specified in detail below.

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

* HealthBee can be installed on your system using the ```go install``` command, for example
    ```> go install github.com/dnataraj/healthbee```
* This will install HealthBee binary to a directory specified in the ```GOBIN``` env var
* A clean shut down of the service can be achieved by simply doing a Ctrl-C

#### Development Notes

* TODO: Highlight testing strategy and possibilities - both unit and integration
* TODO: Support API 
