# RHOSAK-CONSUMER-LAG-EXPORTER

Exports the consumer lag from the RHOSAK Kafka Admin API using the [RHOAS Go SDK](https://github.com/redhat-developer/app-services-sdk-go) to the Prometheus format.

It can either export to stdout on http, the default serving on port for running locally is `7843`. There is a `/data` endpoint for the data, and a health endpoint is included on `/health`.

To build it, run `go build`. To run it, either use `go run` or execute the binary created by `go build`. All paramters and commands are documented, and the help can be viewed using `--help`.

The provided `Dockerfile` will build and run the exporter on port 8080. 

The exporter accepts the bootstrap server(s), the client id and client secret to use a either parameters or environment variables (`BOOTSTRAP_SERVERS`, `CLIENT_ID`, `CLIENT_SECRET`) - these environment variables can be passed into the Docker container.
