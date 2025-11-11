# Swish Analytics Test Publisher

The publisher streams the data sets from their minio buckets and into kafka.

### Deployment
Deployment and mode selection is done through Github Actions of this repository. This deploys a job which publishes a data set to its respective kafka topic for 30 seconds and then it ends.

### Modes
This application runs in 2 modes, 1 and 2.

Mode 1: Publishes messages at a rate of 5 messages per second to topic `data-set-1`.
Mode 2: Publishes messages at a rate of 10 messages per second to topic `data-set-2`.

### Under the hood
Configuration and deployment of the publisher is done using [Helm charts](https://helm.sh/docs). The mode value is supplied through the environment variable `APP_MODE`, which is set by the `mode` field in the helm chart's values.yaml. This field is what is configured by the Github Actions mode selection.

