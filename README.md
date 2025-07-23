# EdgeOne for `libdns`

This package implements the [libdns](https://github.com/libdns/libdns) interfaces for the [EdgeOne API](https://www.tencentcloud.com/zh/document/product/1145/50453)

## Code example

```go
import "github.com/libdns/edgeone"
provider := &edgeone.Provider{
    SecretId:  "YOUR_Secret_ID",
    SecretKey: "YOUR_Secret_Key",
}
```

## Security Credentials

To authenticate you need to supply a [TencentCloud API Key](https://console.tencentcloud.com/cam/capi).

## Other instructions

`libdns/tencentcloud` is based on the new version of Tencentcloud api, uses secret Id and key as authentication methods, supports permission settings, and supports DNSPod international version.

`libdns/edgeone` is based on the Edgeone DNS API, uses token as the authentication method, supports granular permission settings, and is designed for Edgeone global CDN integration.

`libdns/dnspod` is based on the old version of dnspod.cn api, uses token as the authentication method, does not support permission settings, and does not support DNSPod international version.
