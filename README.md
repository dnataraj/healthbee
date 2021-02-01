[![Build Status](https://travis-ci.com/dnataraj/healthbee.svg?token=g7PAjZdpVPTj6UWnsEsA&branch=main)](https://travis-ci.com/dnataraj/healthbee)

#### Development Notes

* TODO: Complete testing - unit and integration
* TODO: Possibly combine monitor and auditor for simpler configuration and running
* TODO: Add stop and deregistration

#### Usage

* Start HealthBee API service
    - First start the monitoring API
      - ```go run ./cmd/web --brokers "<kafka service endpoint>" --dsn "<postgres connection string>"```
      - ```go run ./cmd/web -h``` provides usage information
      - Local brokers and databases can be used, just provide the local service addressed and use the ```-local``` flag
    - Then, start the auditor service
      - ``go run ./cmd/auditors --brokers "<kafka service endpoint>" --dsn "<postgres connection string>"```
      - Currently there are 2 consumers for the auditor service
    
* Register a site for monitoring
    - This can be done with curl and triggers site monitoring at the configured interval e.g.

```
   curl --request POST \
  --url http://localhost:8000/monitor \
  --header 'content-type: application/json' \
  --data '{
	"url": "https://www.duckduckgo.com",
	"interval": "5s",
	"pattern": "privacy"
  }'

```
    
    
* A clean shut down of both the monitor and the auditor can be achieved by simply doing a Ctrl-C
