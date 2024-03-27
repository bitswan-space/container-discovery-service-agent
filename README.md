# Container Discovery Service Agent
Service which is getting pipeline containers info from docker socket in configured polling interval and send it to MQTT topic.

## Dependencies
- MQTT Broker (Mosquitto)
- Docker socket

## Config example
### Service config
```yaml
mqtt-broker-url: mqtt://127.0.0.1:1883
polling-interval: 10
topology-topic: /c/local-gitops/running-pipelines/topology
```
### Docker-compose
```yaml
  cds-agent:
    restart: always
    image: <image_url>
    volumes:
        - ./cds-agent/config.yaml:/conf/config.yaml:r
        - /var/run/docker.sock:/var/run/docker.sock
```