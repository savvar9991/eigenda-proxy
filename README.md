![Compiles](https://github.com/Layr-Labs/eigenda-proxy/actions/workflows/build.yml/badge.svg)
![Unit Tests](https://github.com/Layr-Labs/eigenda-proxy/actions/workflows/unit-tests.yml/badge.svg)
![Linter](https://github.com/Layr-Labs/eigenda-proxy/actions/workflows/lint.yml/badge.svg)
![Integration Tests](https://github.com/Layr-Labs/eigenda-proxy/actions/workflows/holesky-test.yml/badge.svg)

# EigenDA Sidecar Proxy

## Introduction

This service wraps the [high-level EigenDA client](https://github.com/Layr-Labs/eigenda/blob/master/api/clients/eigenda_client.go), exposing endpoints for interacting with the EigenDA disperser in conformance to the [OP Alt-DA server spec](https://specs.optimism.io/experimental/alt-da.html), and adding disperser verification logic. This simplifies integrating EigenDA into various rollup frameworks by minimizing the footprint of changes needed within their respective services.

Features:

* Exposes an API for dispersing blobs to EigenDA and retrieving blobs from EigenDA via the EigenDA disperser
* Handles BN254 field element encoding/decoding
* Performs KZG verification during retrieval to ensure that data returned from the EigenDA disperser is correct.
* Performs KZG verification during dispersal to ensure that DA certificates returned from the EigenDA disperser have correct KZG commitments.
* Performs DA certificate verification during dispersal to ensure that DA certificates have been properly bridged to Ethereum by the disperser.
* Performs DA certificate verification during retrieval to ensure that data represented by bad DA certificates do not become part of the canonical chain.
* Compatibility with Optimism's alt-da commitment type with eigenda backend.
* Compatibility with Optimism's keccak-256 commitment type with S3 storage.

In order to disperse to the EigenDA network in production, or at high throughput on testnet, please register your authentication ethereum address through [this form](https://forms.gle/3QRNTYhSMacVFNcU8). Your EigenDA authentication keypair address should not be associated with any funds anywhere.

## Configuration Options
| Option | Default Value | Environment Variable | Description |
|--------|---------------|----------------------|-------------|
| `--addr` | `"127.0.0.1"` | `$EIGENDA_PROXY_ADDR` | Server listening address |
| `--eigenda.cache-path` | `"resources/SRSTables/"` | `$EIGENDA_PROXY_EIGENDA_TARGET_CACHE_PATH` | Directory path to SRS tables for caching. |
| `--eigenda.custom-quorum-ids` |  | `$EIGENDA_PROXY_EIGENDA_CUSTOM_QUORUM_IDS` | Custom quorum IDs for writing blobs. Should not include default quorums 0 or 1. |
| `--eigenda.disable-point-verification-mode` | `false` | `$EIGENDA_PROXY_EIGENDA_DISABLE_POINT_VERIFICATION_MODE` | Disable point verification mode. This mode performs IFFT on data before writing and FFT on data after reading. Disabling requires supplying the entire blob for verification against the KZG commitment. |
| `--eigenda.disable-tls` | `false` | `$EIGENDA_PROXY_EIGENDA_GRPC_DISABLE_TLS` | Disable TLS for gRPC communication with the EigenDA disperser. Default is false. |
| --eigenda.cert-verification-disabled | `false` | `$EIGENDA_PROXY_EIGENDA_CERT_VERIFICATION_DISABLED` | Whether to verify certificates received from EigenDA disperser. |
| `--eigenda.disperser-rpc` |  | `$EIGENDA_PROXY_EIGENDA_DISPERSER_RPC` | RPC endpoint of the EigenDA disperser. |
| `--eigenda.svc-manager-addr` |  | `$EIGENDA_PROXY_EIGENDA_SERVICE_MANAGER_ADDR` | The deployed EigenDA service manager address. The list can be found here: https://github.com/Layr-Labs/eigenlayer-middleware/?tab=readme-ov-file#current-mainnet-deployment |
| `--eigenda.eth-confirmation-depth` | `-1` | `$EIGENDA_PROXY_EIGENDA_ETH_CONFIRMATION_DEPTH` | The number of Ethereum blocks of confirmation that the DA bridging transaction must have before it is assumed by the proxy to be final. If set negative the proxy will always wait for blob finalization. |
| `--eigenda.eth-rpc` |  | `$EIGENDA_PROXY_EIGENDA_ETH_RPC` | JSON RPC node endpoint for the Ethereum network used for finalizing DA blobs. See available list here: https://docs.eigenlayer.xyz/eigenda/networks/ |
| `--eigenda.g1-path` | `"resources/g1.point"` | `$EIGENDA_PROXY_EIGENDA_TARGET_KZG_G1_PATH` | Directory path to g1.point file. |
| `--eigenda.g2-power-of-2-path` | `"resources/g2.point.powerOf2"` | `$EIGENDA_PROXY_EIGENDA_TARGET_KZG_G2_POWER_OF_2_PATH` | Directory path to g2.point.powerOf2 file. |
| `--eigenda.max-blob-length` | `"16MiB"` | `$EIGENDA_PROXY_EIGENDA_MAX_BLOB_LENGTH` | Maximum blob length to be written or read from EigenDA. Determines the number of SRS points loaded into memory for KZG commitments. Example units: '30MiB', '4Kb', '30MB'. Maximum size slightly exceeds 1GB. |
| `--eigenda.put-blob-encoding-version` | `0` | `$EIGENDA_PROXY_EIGENDA_PUT_BLOB_ENCODING_VERSION` | Blob encoding version to use when writing blobs from the high-level interface. |
| `--eigenda.response-timeout` | `60s` | `$EIGENDA_PROXY_EIGENDA_RESPONSE_TIMEOUT` | Total time to wait for a response from the EigenDA disperser. Default is 60 seconds. |
| `--eigenda.signer-private-key-hex` |  | `$EIGENDA_PROXY_EIGENDA_SIGNER_PRIVATE_KEY_HEX` | Hex-encoded signer private key. This key should not be associated with an Ethereum address holding any funds. |
| `--eigenda.status-query-retry-interval` | `5s` | `$EIGENDA_PROXY_EIGENDA_STATUS_QUERY_INTERVAL` | Interval between retries when awaiting network blob finalization. Default is 5 seconds. |
| `--eigenda.status-query-timeout` | `30m0s` | `$EIGENDA_PROXY_EIGENDA_STATUS_QUERY_TIMEOUT` | Duration to wait for a blob to finalize after being sent for dispersal. Default is 30 minutes. |
| `--log.color` | `false` | `$EIGENDA_PROXY_LOG_COLOR` | Color the log output if in terminal mode. |
| `--log.format` | `text` | `$EIGENDA_PROXY_LOG_FORMAT` | Format the log output. Supported formats: 'text', 'terminal', 'logfmt', 'json', 'json-pretty'. |
| `--log.level` | `INFO` | `$EIGENDA_PROXY_LOG_LEVEL` | The lowest log level that will be output. |
| `--log.pid` | `false` | `$EIGENDA_PROXY_LOG_PID` | Show pid in the log. |
| `--memstore.enabled` | `false` | `$EIGENDA_PROXY_MEMSTORE_ENABLED` | Whether to use mem-store for DA logic. |
| `--memstore.expiration` | `25m0s` | `$EIGENDA_PROXY_MEMSTORE_EXPIRATION` | Duration that a mem-store blob/commitment pair are allowed to live. |
| `--memstore.put-latency` | `0` | `$EIGENDA_PROXY_MEMSTORE_PUT_LATENCY` | Artificial latency added for memstore backend to mimic EigenDA's dispersal latency. |
| `--memstore.get-latency` | `0` | `$EIGENDA_PROXY_MEMSTORE_GET_LATENCY` | Artificial latency added for memstore backend to mimic EigenDA's retrieval latency. |
| `--metrics.addr` | `"0.0.0.0"` | `$EIGENDA_PROXY_METRICS_ADDR` | Metrics listening address. |
| `--metrics.enabled` | `false` | `$EIGENDA_PROXY_METRICS_ENABLED` | Enable the metrics server. |
| `--metrics.port` | `7300` | `$EIGENDA_PROXY_METRICS_PORT` | Metrics listening port. |
| `--port` | `3100` | `$EIGENDA_PROXY_PORT` | Server listening port. |
| `--s3.credential-type` |  | `$EIGENDA_PROXY_S3_CREDENTIAL_TYPE` | Static or iam. |
| `--s3.access-key-id` |  | `$EIGENDA_PROXY_S3_ACCESS_KEY_ID` | Access key id for S3 storage. |
| `--s3.access-key-id` |  | `$EIGENDA_PROXY_S3_ACCESS_KEY_ID` | Access key id for S3 storage. |
| `--s3.access-key-secret` |  | `$EIGENDA_PROXY_S3_ACCESS_KEY_SECRET` | Access key secret for S3 storage. |
| `--s3.bucket` |  | `$EIGENDA_PROXY_S3_BUCKET` | Bucket name for S3 storage. |
| `--s3.path` |  | `$EIGENDA_PROXY_S3_PATH` | Bucket path for S3 storage. |
| `--s3.endpoint` |  | `$EIGENDA_PROXY_S3_ENDPOINT` | Endpoint for S3 storage. |
| `--s3.enable-tls` |  | `$EIGENDA_PROXY_S3_ENABLE_TLS` | Enable TLS connection to S3 endpoint. |
| `--routing.fallback-targets` | `[]` | `$EIGENDA_PROXY_FALLBACK_TARGETS` | Fall back backend targets. Supports S3. | Backup storage locations to read from in the event of eigenda retrieval failure. |
| `--routing.cache-targets` | `[]` | `$EIGENDA_PROXY_CACHE_TARGETS` | Caching targets. Supports S3. | Caches data to backend targets after dispersing to DA, retrieved from before trying read from EigenDA. |
| `--routing.concurrent-write-threads` | `0` | `$EIGENDA_PROXY_CONCURRENT_WRITE_THREADS` | Number of threads spun-up for async secondary storage insertions. (<=0) denotes single threaded insertions where (>0) indicates decoupled writes. |
| `--s3.timeout` | `5s` | `$EIGENDA_PROXY_S3_TIMEOUT` | timeout for S3 storage operations (e.g. get, put) |
| `--redis.db` | `0` |  `$EIGENDA_PROXY_REDIS_DB` | redis database to use after connecting to server |
| `--redis.endpoint` | `""` | `$EIGENDA_PROXY_REDIS_ENDPOINT` | redis endpoint url |
| `--redis.password` | `""` | `$EIGENDA_PROXY_REDIS_PASSWORD` | redis password |
| `--redis.eviction` | `24h0m0s`  | `$EIGENDA_PROXY_REDIS_EVICTION` | entry eviction/expiration time |
| `--help, -h` | `false` |  | Show help. |
| `--version, -v` | `false` |  | Print the version. |


### Certificate verification

In order for the EigenDA Proxy to avoid a trust assumption on the EigenDA disperser, the proxy offers a DA cert verification feature which ensures that:

1. The DA cert's batch hash can be computed locally and matches the one persisted on-chain in the `ServiceManager` contract
2. The DA cert's blob inclusion proof can be successfully verified against the blob-batch merkle root
3. The DA cert's quorum params are adequately defined and expressed when compared to their on-chain counterparts
4. The DA cert's quorum ids map to valid quorums

To target this feature, use the CLI flags `--eigenda-svc-manager-addr`, `--eigenda-eth-rpc`.


#### Soft Confirmations

An optional `--eigenda-eth-confirmation-depth` flag can be provided to specify a number of ETH block confirmations to wait before verifying the blob certificate. This allows for blobs to be accredited upon `confirmation` versus waiting (e.g, 25-30m) for `finalization`. The following integer expressions are supported:
`-1`: Wait for blob finalization
`0`: Verify the cert immediately upon blob confirmation and return the blob
`N where N>0`: Wait `N` blocks before verifying the cert and returning the blob

### In-Memory Backend

An ephemeral memory store backend can be used for faster feedback testing when testing rollup integrations. To target this feature, use the CLI flags `--memstore.enabled`, `--memstore.expiration`.

### Asynchronous Secondary Insertions
An optional `--routing.concurrent-write-routines` flag can be provided to enable asynchronous processing for secondary writes - allowing for more efficient dispersals in the presence of a hefty secondary routing layer. This flag specifies the number of write routines spun-up with supported thread counts in range `[1, 100)`.

### Storage Fallback
An optional storage fallback CLI flag `--routing.fallback-targets` can be leveraged to ensure resiliency when **reading**. When enabled, a blob is persisted to a fallback target after being successfully dispersed. Fallback targets use the keccak256 hash of the existing EigenDA commitment as their key, for succinctness. In the event that blobs cannot be read from EigenDA, they will then be retrieved in linear order from the provided fallback targets. 

### Storage Caching
An optional storage caching CLI flag `--routing.cache-targets` can be leveraged to ensure less redundancy and more optimal reading. When enabled, a blob is persisted to each cache target after being successfully dispersed using the keccak256 hash of the existing EigenDA commitment for the fallback target key. This ensure second order keys are succinct. Upon a blob retrieval request, the cached targets are first referenced to read the blob data before referring to EigenDA. 


## Metrics

To the see list of available metrics, run `./bin/eigenda-proxy doc metrics`

To quickly set up monitoring dashboard, add eigenda-proxy metrics endpoint to a reachable prometheus server config as a scrape target, add prometheus datasource to Grafana to, and import the existing [Grafana dashboard JSON file](./grafana_dashboard.json)

## Deployment Guide

### Hardware Requirements

The following specs are recommended for running on a single production server:

* 4 GB RAM
* 1-2 cores CPU

### Deployment Steps

```bash
## Build EigenDA Proxy
$ make
# env GO111MODULE=on GOOS= GOARCH= go build -v -ldflags "-X main.GitCommit=4b7b35bc3770ed5ca809b7ddb8a825c470a00fb4 -X main.GitDate=1719407123 -X main.Version=v0.0.0" -o ./bin/eigenda-proxy ./cmd/server
# github.com/Layr-Labs/eigenda-proxy/server
# github.com/Layr-Labs/eigenda-proxy/cmd/server

## Setup new keypair for EigenDA authentication
$ cast wallet new -j > keypair.json

## Extract keypair ETH address
$ jq -r '.[0].address' keypair.json
# 0x859F0F6D095E18B732FAdc8CD16Ae144F24e2F0D

## If running against mainnet, register the keypair ETH address and wait for approval: https://forms.gle/niMzQqj1JEzqHEny9

## Extract keypair private key and remove 0x prefix
PRIVATE_KEY=$(jq -r '.[0].private_key' keypair.json | tail -c +3)

## Run EigenDA Proxy
$ ./bin/eigenda-proxy \
    --addr 127.0.0.1 \
    --port 3100 \
    --eigenda-disperser-rpc disperser-holesky.eigenda.xyz:443 \
    --eigenda-signer-private-key-hex $PRIVATE_KEY \
    --eigenda-eth-rpc https://ethereum-holesky-rpc.publicnode.com \
    --eigenda-svc-manager-addr 0xD4A7E1Bd8015057293f0D0A557088c286942e84b
# 2024/06/26 09:41:04 maxprocs: Leaving GOMAXPROCS=10: CPU quota undefined
# INFO [06-26|09:41:04.881] Initializing EigenDA proxy server...     role=eigenda_proxy
# INFO [06-26|09:41:04.884]     Reading G1 points (2164832 bytes) takes 2.169417ms role=eigenda_proxy
# INFO [06-26|09:41:04.961]     Parsing takes 76.634042ms            role=eigenda_proxy
# numthread 10
# WARN [06-26|09:41:04.961] Verification disabled                    role=eigenda_proxy
# INFO [06-26|09:41:04.961] Using eigenda backend                    role=eigenda_proxy
# INFO [06-26|09:41:04.962] Starting DA server                       role=eigenda_proxy endpoint=127.0.0.1:5050
# INFO [06-26|09:41:04.973] Started DA Server                        role=eigenda_proxy
...
```

### Env File

We also provide network-specific example env configuration files in `.env.example.holesky` and `.env.example.mainnet` as a place to get started:

1. Copy example env file: `cp .env.example.holesky .env`
2. Update env file, setting `EIGENDA_PROXY_SIGNER_PRIVATE_KEY_HEX`. On mainnet you will also need to set `EIGENDA_PROXY_ETH_RPC`.
3. Pass into binary: `ENV_PATH=.env ./bin/eigenda-proxy --addr 127.0.0.1 --port 3100`

## Running via Docker

Container can be built via running `make docker-build`.

## Commitment Schemas
Currently, there are two commitment modes supported with unique encoding schemas for each. The `version byte` is shared for all modes and denotes which version of the EigenDA certificate is being used/requested. The following versions are currently supported:
* `0x0`: V0 certificate type (i.e, dispersal blob info struct with verification against service manager)

### Optimism Commitment Mode
For `alt-da` clients running on Optimism, the following commitment schema is supported:

```
 0        1        2         3                N
 |--------|--------|---------|----------------|
  commit   da layer  version   raw commitment
  type       type     byte
```

Both `keccak256` (i.e, S3 storage using hash of pre-image for commitment value) and `generic` (i.e, EigenDA) are supported to ensure cross-compatibility with alt-da storage backends if desired by a rollup operator.

OP Stack itself only has a conception of the first byte (`commit type`) and does no semantical interpretation of any subsequent bytes within the encoding. The `da layer type` byte for EigenDA is always `0x00`. However it is currently unused by OP Stack with name space values still being actively [discussed](https://github.com/ethereum-optimism/specs/discussions/135#discussioncomment-9271282).

### Simple Commitment Mode
For simple clients communicating with proxy (e.g, arbitrum nitro), the following commitment schema is supported:

```
 0         1                 N
 |---------|-----------------|
   version   raw commitment
    byte
```

The `raw commitment` is an RLP-encoded [EigenDA certificate](https://github.com/Layr-Labs/eigenda/blob/eb422ff58ac6dcd4e7b30373033507414d33dba1/api/proto/disperser/disperser.proto#L168).

**NOTE:** Commitments are cryptographically verified against the data fetched from EigenDA for all `/get` calls. The server will respond with status `500` in the event where EigenDA were to lie and provide falsified data thats irrespective of the client provided commitment. This feature cannot be disabled and is part of standard operation.

## Testing

### Unit

Unit tests can be ran via invoking `make test`.

### Integration
End-to-end (E2E) tests can be ran via `make e2e-test`.

### Holesky

A holesky integration test can be ran using `make holesky-test` to assert proper dispersal/retrieval against a public network. Please **note** that EigenDA Holesky network which is subject to rate-limiting and slow confirmation times *(i.e, >10 minutes per blob confirmation)*. Please advise EigenDA's [inabox](https://github.com/Layr-Labs/eigenda/tree/master/inabox#readme) if you'd like to spin-up a local DA network for faster iteration testing.

### Optimism

An E2E test exists which spins up a local OP sequencer instance using the [op-e2e](https://github.com/ethereum-optimism/optimism/tree/develop/op-e2e) framework for asserting correct interaction behaviors with batch submission and state derivation. These tests can be ran via `make optimism-test`.

## Resources

* [op-stack](https://github.com/ethereum-optimism/optimism)
* [Alt-DA spec](https://specs.optimism.io/experimental/alt-da.html)
* [eigen da](https://github.com/Layr-Labs/eigenda)
