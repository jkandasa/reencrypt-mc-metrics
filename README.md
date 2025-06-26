## reencrypt-mc-metrics
This tool is used to reencrypt (or repack) the encrypted metrics that's collected via `mc support telemetry record`.

## Pre-request
The originally encrypted file should be decrypted with `inspect` tool.
```
$ inspect --private-key=private.pem ./metrics-2025-06-26_10-20-14.enc
```

### Usage
```bash 
./reencrypt-mc-metrics --public-key=my-public.pem ./metrics-2025-06-26_10-20-14
```

### Build from source
```bash
CGO_ENABLED=0 go build -o reencrypt-mc-metrics -ldflags "-s -w" main.go
```
