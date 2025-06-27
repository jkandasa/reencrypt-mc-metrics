## repack-mc-metrics
This tool is used to repack the metrics that's collected via `mc support telemetry record` and unencrypted with `inspect` tool.
Optionally repack can be encrypted

## Pre-request
The originally encrypted file should be decrypted with `inspect` tool.
```
$ inspect --private-key=private.pem ./metrics-2025-06-26_10-20-14.enc
```

### Usage
```bash
# repack without encryption
./repack-mc-metrics ./metrics-2025-06-26_10-20-14

# source can be a zip file
./repack-mc-metrics ./metrics-2025-06-26_10-20-14.zip

# repack with encryption
./repack-mc-metrics --encrypt --public-key=my-public.pem ./metrics-2025-06-26_10-20-14
```

### Build from source
```bash
CGO_ENABLED=0 go build -o repack-mc-metrics -ldflags "-s -w" main.go
```
