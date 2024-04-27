# v8 for OMS latency


## How to build

```
go build
```

## How to use

```
./v8 oms_20240411.log ./0411.csv
```

## Cost
- OmsCostTime1: Delay in processing orders from clients.
- OmsCostTime2: Delay in processing returns from the matching engine to clients.
