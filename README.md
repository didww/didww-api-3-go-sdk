Go client for DIDWW API v3.

![Tests](https://github.com/didww/didww-api-3-go-sdk/actions/workflows/ci.yml/badge.svg)
![Coverage](https://img.shields.io/endpoint?url=https://didww.github.io/didww-api-3-go-sdk/badge.json)
![Go](https://img.shields.io/badge/go-1.23%2B-blue)

## About DIDWW API v3

The DIDWW API provides a simple yet powerful interface that allows you to fully integrate your own applications with DIDWW services. An extensive set of actions may be performed using this API, such as ordering and configuring phone numbers, setting capacity, creating SIP trunks and retrieving CDRs and other operational data.

The DIDWW API v3 is a fully compliant implementation of the [JSON API specification](http://jsonapi.org/format/).

This SDK implements JSON:API serialization and deserialization without external dependencies, using only the Go standard library.

Read more https://doc.didww.com/api

## API Version

This SDK targets DIDWW API version `2026-10-22` (`didww.DefaultAPIVersion`). The version is not sent as a header: it is requested in the `api_version` of every access token the SDK mints (see [Authentication](#authentication)).

| Module Version | Branch | DIDWW API Version |
|----------------|--------|-------------------|
| **4.x** | [`main`](https://github.com/didww/didww-api-3-go-sdk) | [`2026-10-22`](https://doc.didww.com/api3/2026-10-22/index.html) |
| **3.x** | [`release-3`](https://github.com/didww/didww-api-3-go-sdk/tree/release-3) | [`2026-04-16`](https://doc.didww.com/api3/2026-04-16/index.html) |
| **2.x** | [`2022-05-10`](https://github.com/didww/didww-api-3-go-sdk/tree/2022-05-10) | [`2022-05-10`](https://doc.didww.com/api3/2022-05-10/index.html) |

`didww.WithAPIVersion` requests a later version. `NewClient` returns an error for a value that is not a `YYYY-MM-DD` date and for a version earlier than `2026-10-22`. The resources of earlier API versions are not in this module: use the 3.x line from the [`release-3`](https://github.com/didww/didww-api-3-go-sdk/tree/release-3) branch for API `2026-04-16`.

## Requirements

- Go 1.23+

## Installation

```bash
go get github.com/didww/didww-api-3-go-sdk/v4
```

**Note on module path:** this module follows Go's
[Semantic Import Versioning](https://go.dev/ref/mod#major-version-suffixes):
the major version of the SDK appears in the import path (`/v4`). It is the
version of the Go module, not of the API: requests still go to
`https://api.didww.com/v3`. Projects on `/v3` are unaffected by `go get -u` and
must consciously update their imports to adopt v4; v3 stays on the `release-3`
branch for API version 2026-04-16.

## Usage

```go
package main

import (
    "context"
    "fmt"

    didww "github.com/didww/didww-api-3-go-sdk/v4"
)

func main() {
    client, err := didww.NewClient("YOUR_API_KEY")
    if err != nil {
        panic(err)
    }

    // Check balance
    balance, err := client.Balance().Find(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Println("Balance:", balance.TotalBalance)

    // List DID groups with stock keeping units
    params := didww.NewQueryParams().
        Include("stock_keeping_units").
        Filter("area_name", "Acapulco")
    didGroups, err := client.DIDGroups().List(context.Background(), params)
    if err != nil {
        panic(err)
    }
    fmt.Println("DID groups:", len(didGroups))
}
```

For more examples visit [examples](examples/).

For details on obtaining your API key please visit https://doc.didww.com/api3/configuration.html

## Examples

- Source code: [examples](examples/)
- How to run: [examples/README.md](examples/README.md)

## Configuration

```go
client, err := didww.NewClient("YOUR_API_KEY",
    didww.WithEnvironment(didww.Production),
)
```

| Option | Default | Description |
|--------|---------|-------------|
| `WithEnvironment(env)` | `didww.Sandbox` | `didww.Sandbox` or `didww.Production`. Selects the API, token service and price hosts. See [Environments](#environments). |
| `WithAPIVersion(v)` | `didww.DefaultAPIVersion` | API version requested in every token. See [API Version](#api-version). |
| `WithScope(s)` | none | `didww.ScopeRead` or `didww.ScopeFull`. Without it no scope is requested and the token service grants `full`. See [Scope](#scope). |
| `WithTokenStore(s)` | `didww.NewMemoryTokenStore()` | Cache for access tokens. See [Token cache](#token-cache). |
| `WithLogger(l)` | `slog.Default()` | `*slog.Logger` for the token warnings and debug events. `nil` silences them. See [Logging](#logging). |
| `WithHTTPClient(hc)` | see below | `*http.Client` used for every call. See [Custom HTTP Client](#custom-http-client-proxy-ssl-etc). |
| `WithBaseURL(url)` | API URL of the environment | JSON:API base URL only. |
| `WithAuthURL(url)` | Token service of the environment | Token service base URL. |
| `WithVoiceInPricesURL(url)`, `WithVoiceOutPricesURL(url)`, `WithSMSPricesURL(url)` | Price hosts of the environment | Price service base URLs, with or without the `/v3` suffix. |

The API key is sent only to the token service. An overridden URL, for example to send the traffic through a proxy, does not change the token audiences (see [Authentication](#authentication)). `NewClient` returns an error for an empty API key or one with control characters (such as a trailing newline), a token service or prices URL that is not http or https, an invalid API version, an unknown scope or an environment other than `didww.Sandbox` and `didww.Production`.

### Custom HTTP Client (Proxy, SSL, etc.)

You can pass a custom `*http.Client` for advanced configuration such as proxy support:

```go
import (
    "net/http"
    "net/url"

    didww "github.com/didww/didww-api-3-go-sdk/v4"
)

proxyURL, _ := url.Parse("http://proxy.example.com:8080")
httpClient := &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}

client, err := didww.NewClient("YOUR_API_KEY",
    didww.WithEnvironment(didww.Production),
    didww.WithHTTPClient(httpClient),
)
```

The SDK adds the bearer token to every request; the API key goes only to `POST /token` on the token service. The `*http.Client` is used as given for every call: the token service, the JSON:API, export download, encrypted file upload and prices. The token service, export download and prices calls use a copy of it with redirects disabled, so they never follow a redirect. Any other HTTP settings (timeouts, TLS, proxies, transports) can be configured on the `*http.Client`.

A total `http.Client.Timeout` also bounds the download of a large price or export file, so prefer the transport timeouts. Without `WithHTTPClient` the SDK uses a client with no total timeout, a 10 s connect timeout, a 10 s TLS handshake timeout and a 60 s response header timeout, built once from a copy of `http.DefaultTransport` and shared by those clients. Code that replaces `http.DefaultTransport`, such as an HTTP mock in tests, should pass its client through `WithHTTPClient`. Each token request is also limited to 10 s.

### Environments

| Environment | API | Token service | Voice IN prices | Voice OUT prices | SMS prices |
|-------------|-----|---------------|-----------------|------------------|------------|
| `didww.Production` | `https://api.didww.com/v3` | `https://auth.didww.com` | `https://api-vin.didww.com/v3` | `https://api-vout.didww.com/v3` | `https://api-sms.didww.com/v3` |
| `didww.Sandbox` | `https://sandbox-api.didww.com/v3` | `https://auth-sandbox.didww.com` | `https://api-vin-sandbox.didww.com/v3` | `https://api-vout-sandbox.didww.com/v3` | `https://api-sms-sandbox.didww.com/v3` |

## Authentication

Every request carries `Authorization: Bearer <token>`. The SDK mints the tokens itself with `POST /token` on the token service, sending your API key in the `Api-Key` header. The API key is sent nowhere else. No code change is needed: `didww.NewClient("YOUR_API_KEY")` works as before.

- **Audience.** One token per audience, minted on the first call to a service and reused until it expires. The audience is the production hostname of the service, in the sandbox too:

  | Calls | Token audience |
  |-------|----------------|
  | JSON:API resources, export download, encrypted file upload | `api.didww.com` |
  | Voice IN prices | `api-vin.didww.com` |
  | Voice OUT prices | `api-vout.didww.com` |
  | SMS OUT (A2P, P2P, OTP) and SMS IN prices | `api-sms.didww.com` |

  `client.PublicKeys().List` is sent without a token.
- **Expiry and 401.** A cached token lives for the `expires_in` granted; the SDK adds no local expiry margin. When a request is answered 401, the SDK drops that token, mints a new one and sends the request once more. A second 401 drops the new token too and is returned (see [Error Handling](#error-handling)). A 403, such as `insufficient_scope`, is never retried.
- **Concurrency.** Concurrent calls of one `Client` that need the same token wait for a single mint. A caller whose context ends stops waiting; the mint itself goes on for the other callers.
- **Token service errors** are returned as `*didww.AuthError` from any call that needs a new token:
  - A 429 returns `didww.ErrRateLimited` with `RetryAfter` (the `Retry-After` seconds, at least 1 s). Every further mint for the same token service and API key in the process fails the same way without a request until that time has passed, plus a small random delay so that clients do not retry in step.
  - A 5xx answer, a connection failure or a timeout is retried with exponential backoff and full jitter (0.5 s base, 8 s cap, 4 attempts in all), then returns `didww.ErrAuthServer` or `didww.ErrAuthConnection`.
  - A 400, a 401 or a TLS failure is returned at once, without a retry.

### Scope

```go
client, err := didww.NewClient("YOUR_API_KEY", didww.WithScope(didww.ScopeRead)) // GET and HEAD only
client, err = didww.NewClient("YOUR_API_KEY", didww.WithScope(didww.ScopeFull))
client, err = didww.NewClient("YOUR_API_KEY") // the default: scope is not sent, the token service grants full
```

A `read` token is enough for the price lookups and price files. A write with a `read` token is answered 403, which the JSON:API repositories return as an `*APIError` (see [Error Handling](#error-handling)).

The SDK compares each token it receives with what it asked for:

- a different `scope` (for example `full` granted for a `read` request) or a different `api_version` is logged as a warning through the logger (see [Logging](#logging)), and the token is used;
- a token with an empty access token, or whose `aud` does not include the requested audience, returns `didww.ErrInvalidGrant` and is not cached.

### Token cache

Tokens are cached in a `didww.TokenStore`, keyed by the token service URL, the API key (as a SHA-256 hash), the audience, the API version and the scope. Each entry is stored for the token's `expires_in`.

The default `didww.NewMemoryTokenStore()` keeps the tokens in process memory. Each `Client` built without `WithTokenStore` has its own, so create one `Client` and reuse it: it is safe for concurrent use. To share tokens between clients or processes, pass one store to all of them, for example your own implementation backed by Redis:

```go
type TokenStore interface {
    Get(ctx context.Context, key string) (t Token, ok bool, err error)
    Set(ctx context.Context, key string, t Token, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

client, err := didww.NewClient("YOUR_API_KEY", didww.WithTokenStore(myRedisStore))
```

`didww.Token` marshals to JSON with all its fields so that a store can persist it; its `String`, `GoString` and `LogValue` never show the access token. A shared store holds live bearer tokens, so protect it like the API key. Processes that share a store can each mint once when a token is missing. A store error is logged as a warning, and the call goes on with a newly minted token. Every store call gets a context. The calls the SDK makes on its own behalf, re-reading the store before a mint, writing a minted token, and the read and delete that evict a refused token after a 401, get a 10-second deadline, so a store should return when its context ends. The wait for that eviction follows the context of the call. A panic on the background goroutine that mints a token, for example in the store's `Set`, the logger or a custom transport, is returned as an `*AuthError` instead of crashing the process; a panic in the store's `Get` or `Delete` on the calling goroutine reaches the caller as usual.

### Logging

`WithLogger` sets the `*slog.Logger` for token events; the default is `slog.Default()`, and `WithLogger(nil)` silences them.

- Warnings: `DIDWW token scope mismatch` and `DIDWW token api_version mismatch` (with `requested` and `granted`), and token store failures.
- Debug: `DIDWW token minted`, with `aud`, `scope`, `api_version`, `expires_in` and `duration`.

Neither the token nor the API key is ever logged.

### SDK design decisions

- **Audience strategy.** One token per audience, minted on first use, rather than one token naming every service. An account that never calls the prices services never requests their audiences, so a refused audience cannot break DID management, and every token is valid for one service only.
- **Cache store.** In-process memory by default, because a Go service usually keeps one `Client` for its lifetime. A shared store is opt-in through `WithTokenStore`. Lifetime is decided by the server through the 401 refresh, not by a local expiry margin.
- **Gzip handling.** Price files are streamed to your writer or file exactly as served, still gzipped, without being held in memory. The SDK never decompresses them on its own: `didww.EachPriceRow` reads a downloaded file when you want rows.
- **Upload retry.** The 401 retry sends the request body a second time. `UploadEncryptedFile` takes the encrypted bytes and builds the multipart body in memory, so the retry always resends the complete body; there is no streamed-upload limitation.

## Resources

### Read-Only Resources

```go
ctx := context.Background()

// Countries
countries, _ := client.Countries().List(ctx, nil)
country, _ := client.Countries().Find(ctx, "uuid")

// Regions
regions, _ := client.Regions().List(ctx, nil)

// Cities
cities, _ := client.Cities().List(ctx, nil)

// Areas
areas, _ := client.Areas().List(ctx, nil)

// NANPA Prefixes
prefixes, _ := client.NanpaPrefixes().List(ctx, nil)

// DID Groups (with stock keeping units), filtered by number type
params := didww.NewQueryParams().
    Include("stock_keeping_units").
    Filter("number_type", string(enums.NumberTypeTollFree))
groups, _ := client.DIDGroups().List(ctx, params)

// Available DIDs (with DID group and stock keeping units)
params = didww.NewQueryParams().Include("did_group.stock_keeping_units")
available, _ := client.AvailableDIDs().List(ctx, params)

// Proof Types
proofTypes, _ := client.ProofTypes().List(ctx, nil)

// Public Keys
publicKeys, _ := client.PublicKeys().List(ctx, nil)

// Address Requirements
requirements, _ := client.AddressRequirements().List(ctx, nil)

// Emergency Requirements (2026-04-16)
emergReqs, _ := client.EmergencyRequirements().List(ctx, nil)

// DID History (2026-04-16)
history, _ := client.DIDHistory().List(ctx, nil)

// Supporting Document Templates
templates, _ := client.SupportingDocumentTemplates().List(ctx, nil)

// Balance (singleton)
balance, _ := client.Balance().Find(ctx)
```

`DIDGroup`, `AddressRequirement`, `EmergencyRequirement` and `EmergencyCallingService` carry a `NumberType`, one of the `enums.NumberType` values (`local`, `national`, `mobile`, `toll_free`, `shared_cost`, `global`). `DIDGroup.IsDemo` flags demo groups.

### DIDs

```go
// List DIDs
dids, _ := client.DIDs().List(ctx, nil)

// Update DID - only changed fields are sent (dirty-only PATCH)
did, _ := client.DIDs().Find(ctx, "uuid")
desc := "Updated"
did.Description = &desc
updated, _ := client.DIDs().Update(ctx, did)
```

#### SMS routing

A DID's inbound messages go to an SMS trunk or to an SMS trunk group. The two are exclusive: setting one sends the other as `null` in the same request, as `VoiceInTrunkID` and `VoiceInTrunkGroupID` do.

```go
import "github.com/didww/didww-api-3-go-sdk/v4/resource"

// sends sms_trunk and sms_trunk_group: null
client.DIDs().Update(ctx, &resource.DID{ID: "did-uuid", SmsTrunkID: "sms-trunk-uuid"})

// sends sms_trunk_group and sms_trunk: null
client.DIDs().Update(ctx, &resource.DID{ID: "did-uuid", SmsTrunkGroupID: "sms-trunk-group-uuid"})

// unroute: sends sms_trunk: null and sms_trunk_group: null
client.DIDs().Update(ctx, &resource.DID{ID: "did-uuid", NullifySmsTrunk: true, NullifySmsTrunkGroup: true})

params := didww.NewQueryParams().
    Filter("sms_trunk.id", "sms-trunk-uuid").
    Include("sms_trunk,sms_trunk_group")
dids, _ := client.DIDs().List(ctx, params) // did.SmsTrunk, did.SmsTrunkGroup
```

An SMS trunk cannot be deleted while DIDs are routed to it, nor an SMS trunk group while DIDs are routed to it or to one of its trunks: unroute those DIDs, or route them elsewhere, first.

### Voice In Trunks

```go
import "github.com/didww/didww-api-3-go-sdk/v4/resource/enums"

// Create SIP trunk
ringingTimeout := 30
trunk := &didww.VoiceInTrunk{
    Name:           "My SIP Trunk",
    Priority:       1,
    Weight:         100,
    CliFormat:      enums.CliFormatE164,
    RingingTimeout: &ringingTimeout,
    Configuration: &didww.SIPConfiguration{
        Host:                "sip.example.com",
        Port:                5060,
        CodecIDs:            []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
        TransportProtocolID: enums.TransportProtocolUDP,
    },
}
created, _ := client.VoiceInTrunks().Create(ctx, trunk)

// Update trunk
desc := "Updated"
created.Description = &desc
updated, _ := client.VoiceInTrunks().Update(ctx, created)

// Delete trunk
client.VoiceInTrunks().Delete(ctx, created.ID)
```

`Pop` is a plain `*string` attribute, one of the `enums.Pop*` values (`PopNYC`, `PopFRA`, `PopLAC`, `PopMIA`, `PopSG`, `PopHK`, `PopAMS`). Leave it `nil` to let DIDWW choose; setting it to `nil` on a fetched trunk sends `null`. It is no longer a relationship and cannot be included.

```go
trunk, _ := client.VoiceInTrunks().Find(ctx, "trunk-uuid")
trunk.Pop = didww.Ptr(enums.PopAMS)
updated, _ := client.VoiceInTrunks().Update(ctx, trunk)
```

#### SIP Registration (2026-04-16)

A Voice In Trunk can also authenticate inbound calls via SIP registration
credentials generated by DIDWW. The SDK's `MarshalJSON` auto-cascades the
dependent fields the server requires:

* `EnabledSipRegistration: didww.Ptr(true)` strips `Host` / `Port` from
  the wire (the server rejects them with 422 otherwise);
* `Host: "..."` (non-empty) sends `enabled_sip_registration: false` and
  `use_did_in_ruri: false` on the wire so the server accepts the disable
  PATCH — even when the caller didn't set them on the struct.

The cascade is applied to a local copy inside `MarshalJSON`, so the
caller's struct value is never mutated as a side-effect. The toggle
fields are `*bool` (not plain `bool`) so explicit `false` is preserved
on the wire — use `didww.Ptr(...)` to construct pointers.

The server generates `IncomingAuthUsername` and `IncomingAuthPassword`
and surfaces them in the response when SIP registration is enabled.

```go
trunk := &didww.VoiceInTrunk{
    Name: "Office (registered)",
    Configuration: &didww.SIPConfiguration{
        EnabledSipRegistration: didww.Ptr(true),
        UseDIDInRuri:           didww.Ptr(true),
        CnamLookup:             didww.Ptr(true),
    },
}
created, _ := client.VoiceInTrunks().Create(ctx, trunk)
sip := created.Configuration.(*didww.SIPConfiguration)
// sip.IncomingAuthUsername — server-generated
// sip.IncomingAuthPassword — server-generated
```

To disable SIP registration on an existing trunk, just set `Host` —
`MarshalJSON` flips `enabled_sip_registration` and `use_did_in_ruri` to
`false` on the wire automatically:

```go
update := &didww.VoiceInTrunk{
    ID: "trunk-uuid",
    Configuration: &didww.SIPConfiguration{
        Host: "sip.example.com",
    },
}
client.VoiceInTrunks().Update(ctx, update)
```

### Voice In Trunk Groups

```go
capacityLimit := 50
group := &didww.VoiceInTrunkGroup{
    Name:            "Primary Group",
    CapacityLimit:   &capacityLimit,
    VoiceInTrunkIDs: []string{trunkA.ID, trunkB.ID},
}
created, _ := client.VoiceInTrunkGroups().Create(ctx, group)
```

### Voice Out Trunks

> **Note:** Voice Out Trunks require additional account configuration. Contact DIDWW support to enable.
> The `replace_cli` and `randomize_cli` values of `OnCliMismatchAction` also require account configuration.

Voice Out Trunks use a polymorphic `AuthenticationMethod` (2026-04-16). Three types are supported:

- **`credentials_and_ip`** -- default method; `Username` and `Password` are server-generated and returned in the response.
- **`twilio`** -- requires a `TwilioAccountSid`.
- **`ip_only`** -- read-only; can only be configured by DIDWW staff upon request. Cannot be set via the API.

```go
import (
    "github.com/didww/didww-api-3-go-sdk/v4/resource/authenticationmethod"
    "github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

// NOTE: 203.0.113.0/24 is RFC 5737 TEST-NET-3 documentation space.
// Replace with the real CIDR of your SIP infrastructure.
trunk := &didww.VoiceOutTrunk{
    Name: "My Outbound Trunk",
    AuthenticationMethod: &authenticationmethod.CredentialsAndIp{
        AllowedSipIPs: []string{"203.0.113.0/24"},
    },
    DefaultDstAction:    enums.DefaultDstActionAllowAll,
    OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
    MediaEncryptionMode: enums.MediaEncryptionModeDisabled,
}
created, _ := client.VoiceOutTrunks().Create(ctx, trunk)
// created.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp).Username -- server-generated
// created.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp).Password -- server-generated
```

### SMS Trunks

An SMS trunk delivers inbound messages, accepts outbound ones, or both. Its `Configuration` is one of five types in package `github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration`:

| Type | Struct | Notes |
|------|--------|-------|
| `sms_to_email` | `SmsToEmail` | Inbound messages are forwarded as email. `SendTo`, `Subject` and `Message` are required. |
| `http_in` | `HTTPIn` | Inbound messages are sent as an HTTP request to your `URL`. `HTTPMethod` is one of `HTTPMethodGet`, `HTTPMethodPost`, `HTTPMethodPut`; post and put also need `BodyType` (one of the `BodyType*` constants) and `Body`. |
| `http_out` | `HTTPOut` | You submit outbound messages over HTTP. `AuthName` and `AuthPassword` are generated. The submit URL is `trunk.Meta["host"]`. |
| `smpp_esme` | `SmppEsme` | You bind to DIDWW as an SMPP ESME. `ServerLocation` is required; `SystemID` and `Password` are generated. |
| `smpp_smsc` | `SmppSmsc` | DIDWW binds to your SMSC with your `Host`, `SystemID`, `Password` and `ServerLocation`. |

The read-only `AllowIn` and `AllowOut` tell which directions a trunk handles. `ServerLocation` is `ServerLocationNYC` or `ServerLocationAMS`; the sandbox accepts `nyc` only. A configuration type this SDK does not know is read as `*smstrunkconfiguration.Generic` and sent back unchanged; its numbers are `json.Number`, so they keep their exact value.

```go
import (
    "github.com/didww/didww-api-3-go-sdk/v4/resource"
    "github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration"
)

trunk, _ := client.SmsTrunks().Create(ctx, &resource.SmsTrunk{
    Name: "Inbound webhook",
    Configuration: &smstrunkconfiguration.HTTPIn{
        URL:        "https://example.com/sms",
        HTTPMethod: smstrunkconfiguration.HTTPMethodPost,
        BodyType:   smstrunkconfiguration.BodyTypeJSON,
        Body:       `{"text":"{SMS_TEXT}","from":"{SMS_SRC_ADDR}","to":"{SMS_DST_ADDR}"}`,
    },
})

// On update the configuration type must stay the same; attributes left out keep their values.
trunk.Configuration = &smstrunkconfiguration.HTTPIn{URL: "https://example.com/sms/v2"}
trunk, _ = client.SmsTrunks().Update(ctx, trunk)

// Nullable attributes are cleared with a Generic configuration of the same type: nil is sent as null
trunk.Configuration = &smstrunkconfiguration.Generic{
    Type:       smstrunkconfiguration.TypeHTTPIn,
    Attributes: map[string]any{"http_method": "get", "body_type": nil, "body": nil},
}
trunk, _ = client.SmsTrunks().Update(ctx, trunk)

// List by configuration type, with the group
params := didww.NewQueryParams().
    Filter("configuration.type", "http_in,sms_to_email").
    Include("sms_trunk_group")
trunks, _ := client.SmsTrunks().List(ctx, params)
```

The typed configuration structs cannot clear a nullable attribute: a field at its zero value is omitted, and on update an omitted attribute keeps its stored value. This covers the `QueryParameters`, `Headers`, `BodyType` and `Body` of `http_in`, the `CallbackURL` and `AllowedIPs` of `http_out`, the `SystemType` and `AllowedIPs` of `smpp_esme` and the `SystemType` of `smpp_smsc`. Send such a change as a `*smstrunkconfiguration.Generic` of the same type, as above: its `nil` attribute values are sent as `null`.

`OutgoingDIDIDs` replaces the whole `outgoing_dids` list of the trunk: list every DID to keep. A fetched trunk does not fill it; build it from `trunk.OutgoingDIDs` (include `outgoing_dids`).

The generated credentials of `http_out` and `smpp_esme` trunks are read-only and are never sent back on update. `String` and `GoString` (`%v`, `%+v`, `%#v`) mask them, the `Password` of an `smpp_smsc` configuration, and the `Headers` of an `http_in` configuration, which may carry the credentials of your endpoint. A `*smstrunkconfiguration.Generic` masks every attribute value, since its fields are not known; only the attribute names and explicit nulls are shown. The masking applies to the configuration pointers that `trunk.Configuration` holds; a dereferenced value prints its fields, and the struct fields keep the real values. To rotate the generated credentials:

```go
_, err := client.SmsTrunkRegenerateCredentials().Create(ctx, &resource.SmsTrunkRegenerateCredential{
    SmsTrunkID: outTrunkID,
})
outTrunk, err := client.SmsTrunks().Find(ctx, outTrunkID)
switch cfg := outTrunk.Configuration.(type) {
case *smstrunkconfiguration.HTTPOut:
    // cfg.AuthName and cfg.AuthPassword are the new credentials
case *smstrunkconfiguration.SmppEsme:
    // cfg.SystemID and cfg.Password are the new credentials
}
```

### SMS Trunk Groups

An SMS trunk group spreads a DID's inbound messages over its trunks by their `Priority` (the higher value first). `http_out` trunks cannot join a group. Deleting a group also deletes the trunks in it.

```go
group, _ := client.SmsTrunkGroups().Create(ctx, &resource.SmsTrunkGroup{
    Name:        "Inbound SMS",
    SmsTrunkIDs: []string{"sms-trunk-uuid"},
})
fmt.Println(group.Meta["sms_trunks_count"]) // 1
```

`SmsTrunkIDs` replaces the whole membership of the group: list every trunk to keep. A fetched group does not fill it; build it from the included `SmsTrunks`:

```go
group, _ = client.SmsTrunkGroups().Find(ctx, group.ID, didww.NewQueryParams().Include("sms_trunks"))
ids := []string{"another-sms-trunk-uuid"}
for _, t := range group.SmsTrunks {
    ids = append(ids, t.ID)
}
group, _ = client.SmsTrunkGroups().Update(ctx, &resource.SmsTrunkGroup{ID: group.ID, SmsTrunkIDs: ids})
```

An empty `SmsTrunkIDs` sends nothing; set `ClearSmsTrunks: true` to empty the group.

### Sender ID Verifications

A Sender ID verification covers the sender IDs your outbound messages use. It is created and reviewed outside the API: the API lists and reads verifications, and changes only their `Name` and the SMS trunk their messages are routed through. A closed verification cannot be changed. `Status` is one of the `enums.SenderIDVerificationStatus` values and `SenderType` one of the `enums.SenderType` values; `IsActive()`, `IsClosed()` and the other predicates test the status. `Meta` carries `setup_price` and `monthly_price` as decimal strings.

```go
params := didww.NewQueryParams().
    Filter("status", string(enums.SenderIDVerificationStatusActive)).
    Include("sms_trunk")
verifications, _ := client.SenderIDVerifications().List(ctx, params)

// http_out, smpp_esme or smpp_smsc trunks only
client.SenderIDVerifications().Update(ctx, &resource.SenderIDVerification{
    ID:         verifications[0].ID,
    SmsTrunkID: "sms-trunk-uuid",
})

// Outbound messages stop until a trunk is assigned again
client.SenderIDVerifications().Update(ctx, &resource.SenderIDVerification{
    ID:              verifications[0].ID,
    NullifySmsTrunk: true,
})
```

Deleting an SMS trunk leaves the verifications routed through it without a trunk.

### Orders

```go
// Create order with DID order item
order := &didww.Order{
    Items: []didww.OrderItem{
        {
            Type: "did_order_items",
            Attributes: didww.OrderItemAttributes{
                SkuID: "sku-uuid",
                Qty:   2,
            },
        },
    },
}
created, _ := client.Orders().Create(ctx, order)

// Delete order (cancel)
client.Orders().Delete(ctx, created.ID)
```

### DID Reservations

```go
reservation := &didww.DIDReservation{
    Description:    "Reserved for client",
    AvailableDIDID: "available-did-uuid",
}
created, _ := client.DIDReservations().Create(ctx, reservation)

// Delete reservation
client.DIDReservations().Delete(ctx, created.ID)
```

### Shared Capacity Groups

```go
group := &didww.SharedCapacityGroup{
    Name:                 "Shared Group",
    SharedChannelsCount:  20,
    MeteredChannelsCount: 0,
    CapacityPoolID:       "pool-uuid",
}
created, _ := client.SharedCapacityGroups().Create(ctx, group)
```

### Identities

```go
import "github.com/didww/didww-api-3-go-sdk/v4/resource/enums"

identity := &didww.Identity{
    FirstName:    "John",
    LastName:     "Doe",
    PhoneNumber:  "12125551234",
    IdentityType: enums.IdentityTypePersonal,
    CountryID:    "country-uuid",
}
created, _ := client.Identities().Create(ctx, identity)
```

### Addresses

```go
address := &didww.Address{
    CityName:   "New York",
    PostalCode: "10001",
    Address:    "123 Main St",
    IdentityID: identity.ID,
    CountryID:  "country-uuid",
}
created, _ := client.Addresses().Create(ctx, address)
```

### Address Verifications

```go
cbURL := "http://example.com/callback"
cbMethod := "GET"
verification := &didww.AddressVerification{
    CallbackURL:    &cbURL,
    CallbackMethod: &cbMethod,
    AddressID:      address.ID,
    DIDIDs:         []string{"did-uuid"},
}
created, _ := client.AddressVerifications().Create(ctx, verification)
```

### Emergency Services (2026-04-16)

```go
// List emergency requirements filtered by country
params := didww.NewQueryParams().Filter("country.id", "country-uuid")
reqs, _ := client.EmergencyRequirements().List(ctx, params)

// Create emergency verification
verification := &didww.EmergencyVerification{
    AddressID:  "address-uuid",
    IdentityID: "identity-uuid",
    DIDIDs:     []string{"did-uuid"},
}
created, _ := client.EmergencyVerifications().Create(ctx, verification)

// List emergency calling services
services, _ := client.EmergencyCallingServices().List(ctx, nil)
```

### DID History (2026-04-16)

```go
// List DID history entries
params := didww.NewQueryParams().Filter("did.id", "did-uuid")
history, _ := client.DIDHistory().List(ctx, params)
```

### Exports

```go
import (
    "github.com/didww/didww-api-3-go-sdk/v4/resource"
    "github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

export := &resource.Export{
    ExportType: enums.ExportTypeCdrIn,
    Filters:    map[string]interface{}{"from": "2026-04-01 00:00:00", "to": "2026-04-16 00:00:00"},
}
created, _ := client.Exports().Create(ctx, export)

// SMS export
smsOut, _ := client.Exports().Create(ctx, &resource.Export{
    ExportType: enums.ExportTypeSmsOut,
    Filters: map[string]interface{}{
        "from":         "2026-10-01 00:00:00", // inclusive
        "to":           "2026-10-02 00:00:00", // exclusive
        "service_type": "a2p",
        "sms_trunk.id": []string{"sms-trunk-uuid"},
    },
})

// The export is pending until its file is generated; read it again later
done, _ := client.Exports().Find(ctx, smsOut.ID)
if done.URL != nil {
    var gz bytes.Buffer
    err := client.DownloadExport(ctx, *done.URL, &gz) // gzipped CSV, as served
    var csv bytes.Buffer
    err = client.DownloadAndDecompressExport(ctx, *done.URL, &csv)
}
```

`DownloadExport` and `DownloadAndDecompressExport` send the token only to the origin (scheme, host and port) of the JSON:API base URL or of the environment, so export URLs, which the API builds on the environment's host, are still downloaded when `WithBaseURL` points elsewhere; that request goes directly to the environment's host. A URL on any other origin returns a `*didww.ClientError` and nothing is sent.

`from` and `to` are required for every export type (`cdr_in`, `cdr_out`, `sms_in`, `sms_out`). The other filter keys depend on the type:

- `sms_in` and `sms_out`: `source_address_eq`, `source_address_cont`, `destination_address_eq`, `destination_address_cont`, `status`, `sms_id`, `sms_trunk.id` (one id or a slice of ids);
- `sms_in` only: `trunk_type`, one or more of the `enums.ExportTrunkType*` values;
- `sms_out` only: `service_type` (`a2p` or `p2p`) and `sender_id_verification.id` (one id or a slice of ids).

The doc comment of `resource.Export` lists the keys of every type.

## Prices

Prices are served by three services outside the JSON:API (hosts in [Environments](#environments)). Each price list has a price file and a lookup:

| Accessor | Service | Price file | Lookup | Lookup rows |
|----------|---------|------------|--------|-------------|
| `client.VoiceInPrices()` | Voice IN | `prices.csv.gz` | `lookup/prices` | `resource.VoiceInPrice` |
| `client.VoiceOutPrices()` | Voice OUT | `prices.csv.gz` | `lookup/prices` | `resource.VoiceOutPrice` |
| `client.SMSOutA2PPrices()` | SMS | `sms_out/a2p/prices.csv.gz` | `sms_out/a2p/lookup/prices` | `resource.SMSOutPrice` |
| `client.SMSOutP2PPrices()` | SMS | `sms_out/p2p/prices.csv.gz` | `sms_out/p2p/lookup/prices` | `resource.SMSOutPrice` |
| `client.SMSOutOTPPrices()` | SMS | `sms_out/otp/prices.csv.gz` | `sms_out/otp/lookup/prices` | `resource.SMSOutPrice` |
| `client.SMSInPrices()` | SMS | `sms_in/prices.csv.gz` | `sms_in/lookup/prices` | `resource.SMSInPrice` |

Each accessor returns a `*didww.PricesService[T]`. The account needs a price plan of that service: without one, every call returns `didww.ErrPricePlanNotAssigned`.

### Price lookups

```go
result, err := client.VoiceInPrices().Lookup(ctx, map[string]string{
    "number":          "18005550123",
    "src_country_iso": "US",
})
if err != nil {
    return err
}
fmt.Println(result.TotalRecords)
for _, price := range result.Rows {
    if price.Rate != nil {
        fmt.Printf("%s %s: %s USD/min\n", price.Destination, price.SourceType, *price.Rate)
    }
}

voiceOut, _ := client.VoiceOutPrices().Lookup(ctx, map[string]string{"number": "493012345678"})
fmt.Println(voiceOut.Rows[0].Category) // enums.VoiceOutPriceCategory
a2p, _ := client.SMSOutA2PPrices().Lookup(ctx, map[string]string{"dst_number": "4915123456789"})
smsIn, _ := client.SMSInPrices().Lookup(ctx, map[string]string{"dst_country_iso": "US"})
```

| Service | Filters |
|---------|---------|
| Voice IN | `number` (required), `src_country_iso` |
| Voice OUT | `number` (required); no other filter |
| SMS OUT and SMS IN | `dst_number`, `dst_country_iso`; at least one |

A missing or unknown filter returns an error before any request; empty values are dropped. The server checks the value formats and answers a bad value with 400 (`didww.ErrPricesBadRequest`).

`Lookup` returns a `*didww.PriceLookup[T]` with `Rows` and `TotalRecords`. It always holds at least one row: no match returns `didww.ErrPriceNotFound`, never an empty result. In the rows, prices and fees are decimal strings in USD (`*string`, `nil` when the row rejects the destination; `SMSInPrice.Price` is a `string`), intervals and increments are `int` (`*int` for the scheduled `Next*` increments), `SMSOutPrice.MCC` and `MNC` are `*int`, `VoiceOutPrice.EffectiveAt` is a `*time.Time` and `VoiceOutPrice.Category` is one of the `enums.VoiceOutPriceCategory` values.

The `ID` of an SMS OUT or SMS IN lookup row is derived from its content: it changes when the row or its price changes, so do not use it to track a destination over time. A Voice OUT `ID` stays the same when a re-import changes only the price. A Voice IN `ID` names the answering row of the plan.

### Price files

The files are gzipped CSV with a header row. The SDK writes them exactly as served and never decompresses them.

```go
prices := client.VoiceOutPrices()

// Download unless the copy at the path is current
file, err := prices.Fetch(ctx, "/var/data/voice_out_prices.csv.gz")
if err != nil {
    return err
}
if file.NotModified() {
    fmt.Println("answered 304, the file was kept")
} else {
    fmt.Println("downloaded", file.BytesWritten, "bytes, last modified", file.LastModified)
}

// Read the rows
f, err := os.Open("/var/data/voice_out_prices.csv.gz")
if err != nil {
    return err
}
defer f.Close()
err = didww.EachPriceRow(f, func(row map[string]string) error {
    fmt.Printf("%s: %s\n", row["Prefix"], row["Rate"])
    return nil
})

// Lower-level calls
status, err := prices.Head(ctx, file.LastModified) // sent as If-Modified-Since; a zero time sends none
// status.NotModified(), status.ContentLength (-1 when unknown),
// status.Filename (from Content-Disposition, e.g. "voice_out_prices.csv.gz")

file, err = prices.DownloadToFile(ctx, "/var/data/voice_out_prices.csv.gz")
file, err = prices.Download(ctx, w, time.Time{}) // any io.Writer
```

- `Fetch(ctx, path)` sends the modification time of the file at `path` as `If-Modified-Since`, when the file is there; an unchanged price file is answered 304 and the file is kept. A file written by `DownloadToFile` or `Fetch` gets the `Last-Modified` of the response as its modification time, so the date matches the copy on disk. `Fetch` trusts the modification time of any file at `path`: a file copied or changed by other means can be answered 304 and kept, so replace such a file with `DownloadToFile`.
- `DownloadToFile` and `Fetch` stream the body into a temp file next to `path`, unique to the call, which gets the mode of a new file (`0666` minus the umask) and replaces `path` atomically only on a 200. A 304 or an error leaves an existing file untouched, the temp file is always removed, and concurrent downloads to one path do not mix their bytes.
- `Download(ctx, w, ifModifiedSince)` writes the body into `w` only on a 200, as it arrives, so after a transfer error `w` may hold part of it. An error returned by `w` is returned as is.
- `Head` and the downloads return a `*didww.PriceFile`: `StatusCode`, `LastModified` (zero when absent), `ContentLength`, `Filename`, `BytesWritten` and `NotModified()`.
- `EachPriceRow(r, fn)` gunzips a downloaded file and calls `fn` with each row keyed by the header line. The Voice OUT file has 15 columns. The SMS OUT files have 9: `Source prefix`, `Destination prefix`, `Country`, `MCC`, `MCC name`, `MNC`, `MNC name`, `Reject SMS` (`Yes` or `No`) and `Price (USD, excl. VAT)` with 6 decimals.
- Only `Last-Modified` and `If-Modified-Since` are used. The service sends no `ETag`, and the SDK never sends `If-None-Match`, which would make the service ignore `If-Modified-Since`.
- A 503 on a price file is retried up to 2 times when its `Retry-After` is 60 seconds or less, after waiting that long (at least 1 s) plus a small random delay so that clients do not retry in step. Otherwise it returns `didww.ErrPricesServiceUnavailable` with `RetryAfter`. Lookups are not retried.
- A 404 `price_not_found` returns `didww.ErrPriceNotFound`; any other 404, `price_plan_not_assigned` included, and any 404 on `Head`, which has no body, returns `didww.ErrPricePlanNotAssigned`. `PricesError.Code` holds the code of the answer, or `price_plan_not_assigned` for a 404 without one.
- Redirects are not followed. A status without its own sentinel in [Error Handling](#error-handling), such as a 3xx, a 202 or a 204, returns `didww.ErrPricesUnexpectedStatus`. An error is never turned into an empty file or an empty result.

## Filtering, Sorting, and Pagination

```go
params := didww.NewQueryParams().
    Filter("country.id", "uuid").
    Filter("name", "Arizona").
    Include("country").
    Sort("name").
    Page(1, 25)

regions, _ := client.Regions().List(ctx, params)
```

## Dirty PATCH Serialization

The SDK tracks which fields have been modified and sends only those fields in PATCH requests. This avoids overwriting server-side values that your code hasn't touched.

### Updating a fetched resource

When you fetch a resource and modify it, only the changed fields are sent:

```go
did, _ := client.DIDs().Find(ctx, "uuid")
did.DedicatedChannelsCount = 5
// PATCH payload includes only "dedicated_channels_count", not all attributes
updated, _ := client.DIDs().Update(ctx, did)
```

### Building a resource for update

Create a struct with just the ID to send a PATCH without fetching first:

```go
desc := "New description"
updated, _ := client.DIDs().Update(ctx, &didww.DID{
    ID:          "uuid",
    Description: &desc,
})
// PATCH payload includes only "description"
```

### Clearing a field with explicit null

Setting a pointer field to `nil` after it had a value sends an explicit `null` in the payload:

```go
did, _ := client.DIDs().Find(ctx, "uuid")
did.Description = nil
// PATCH payload: { "description": null }
updated, _ := client.DIDs().Update(ctx, did)
```

### Clearing a relationship

Setting a relationship ID to empty on a built resource sends `"data": null` for to-one relationships:

```go
did, _ := client.DIDs().Find(ctx, "uuid")
did.VoiceInTrunkID = "trunk-uuid"
// PATCH payload includes: "relationships": { "voice_in_trunk": { "data": { ... } }, "voice_in_trunk_group": { "data": null } }
// Mutual exclusion is handled automatically.
updated, _ := client.DIDs().Update(ctx, did)
```

### Included resources

Dirty tracking is automatically enabled on included (sideloaded) resources, so you can fetch with includes and update a related resource directly:

```go
params := didww.NewQueryParams().Include("voice_in_trunk")
did, _ := client.DIDs().Find(ctx, "uuid", params)
trunk := did.VoiceInTrunk
desc := "Updated via include"
trunk.Description = &desc
updated, _ := client.VoiceInTrunks().Update(ctx, trunk)
```

## Trunk Configuration Types

| Type | Struct |
|------|--------|
| SIP | `SIPConfiguration` |
| PSTN | `PSTNConfiguration` |

For the SMS trunk configuration types, see [SMS Trunks](#sms-trunks).

## Order Item Types

| Type | JSON:API type |
|------|---------------|
| DID | `did_order_items` |
| Available DID | `available_did_order_items` |
| Reservation DID | `reservation_did_order_items` |
| Capacity | `capacity_order_items` |
| Emergency | `emergency_order_items` |
| Generic (response only) | `generic_order_items` |

## Filter Reference

> See [`FILTERS.md`](FILTERS.md) for the list of `filter[KEY]` keys accepted by every list endpoint at API version `2026-10-22`.

## Error Handling

```go
import "errors"

trunk, err := client.VoiceInTrunks().Find(ctx, "nonexistent")
if err != nil {
    var apiErr *didww.APIError
    if errors.As(err, &apiErr) {
        fmt.Println("HTTP Status:", apiErr.HTTPStatus)
        for _, e := range apiErr.Errors {
            fmt.Println("Error:", e.Detail)
        }
    }

    var clientErr *didww.ClientError
    if errors.As(err, &clientErr) {
        fmt.Println("Client error:", clientErr.Message)
    }
}
```

For the JSON:API repositories and `DownloadExport`, `APIError` also reports the authentication failures of the API: a 401 still answered after the token was replaced once (`HTTPStatus` 401), and a 403 such as `insufficient_scope` for a write with a `read` token, which is never retried. `UploadEncryptedFile` reports any error status, including a 401 after the retry and a 403, as a `*ClientError` whose `Message` holds the HTTP status. `DownloadExport` does not follow redirects: any status other than 200, such as a 3xx or 204, returns an `*APIError`.

Errors of the token service and of the prices services are told apart with `errors.Is` and their sentinels, and their details read with `errors.As`:

```go
_, err := client.SMSOutA2PPrices().Lookup(ctx, map[string]string{"dst_number": "4915123456789"})

var authErr *didww.AuthError
var pricesErr *didww.PricesError
switch {
case errors.Is(err, didww.ErrRateLimited) && errors.As(err, &authErr):
    // Wait RetryAfter and try again; repeat while the error is ErrRateLimited.
    fmt.Println("Token service rate limit, retry after", authErr.RetryAfter)
case errors.Is(err, didww.ErrInvalidClient):
    fmt.Println("API key refused:", err)
case errors.Is(err, didww.ErrPriceNotFound):
    fmt.Println("No price for this number")
case errors.Is(err, didww.ErrPricesTooManyRequests) && errors.As(err, &pricesErr):
    fmt.Println("Prices rate limit, retry after", pricesErr.RetryAfter)
case err != nil:
    fmt.Println("Error:", err)
}
```

**Token service** (`*didww.AuthError`, with `StatusCode`, `Code`, `Description`, `RetryAfter` and `Err`). It is returned as is from any call that needs a new token:

| Sentinel | Returned when |
|----------|---------------|
| `ErrInvalidClient` | 401: the token service refused the API key. Not retried. |
| `ErrInvalidRequest` | 400, e.g. `invalid_request` or `invalid_scope`. |
| `ErrRateLimited` | 429. `RetryAfter` is the server's wait. Further mints fail without a request until `RetryAfter` plus a small random delay has passed; such a refusal reports the time still left in `RetryAfter`, so retry in a loop rather than once. |
| `ErrAuthServer` | 5xx on the last of the 4 attempts. |
| `ErrAuthConnection` | Connection failure or timeout on the last attempt, or a TLS failure (not retried). |
| `ErrInvalidGrant` | The token is empty or does not cover the requested audience. |
| none | Any other status. |

**Prices** (`*didww.PricesError`, with `StatusCode`, `Code`, `Errors`, `RetryAfter` and `Err`):

| Sentinel | Status |
|----------|--------|
| `ErrPricesBadRequest` | 400, e.g. a filter value of the wrong format |
| `ErrPricesNotAuthorized` | 401 after the token was replaced once |
| `ErrPricesAccessDenied` | 403 |
| `ErrPricePlanNotAssigned` | 404 `price_plan_not_assigned`, and any other 404 except `price_not_found`, including any 404 on `Head`, which has no body |
| `ErrPriceNotFound` | 404 `price_not_found`: no price matches the lookup |
| `ErrPricesNotAcceptable` | 406 |
| `ErrPricesTooManyRequests` | 429, with `RetryAfter`. Not retried. |
| `ErrPricesServiceUnavailable` | 503, with `RetryAfter`. Retried on price files only, see [Price files](#price-files). |
| `ErrPricesUnexpectedStatus` | Any other status |
| `ErrPricesConnection` | Connection failure, timeout, TLS failure or a response body that breaks off. Not retried. An error writing the file or `io.Writer` is returned as is. |

The prices services can also return an `*AuthError`, and `Lookup` a plain error for invalid filters and a `*ClientError` when a 200 answer is not a JSON:API document it can parse or holds no prices. `Download` returns a `*ClientError` for a nil `io.Writer`. `DownloadToFile` and `Fetch` also return the `os` and `io/fs` errors of the file at the path (stat, temp file, times, rename) as is. When the context ends while a 503 `Retry-After` is being waited out, the context's error is returned as is.

## All Supported Resources

| Resource | Repository | Operations |
|----------|-----------|------------|
| Country | `client.Countries()` | list, find |
| Region | `client.Regions()` | list, find |
| City | `client.Cities()` | list, find |
| Area | `client.Areas()` | list, find |
| NanpaPrefix | `client.NanpaPrefixes()` | list, find |
| DIDGroup | `client.DIDGroups()` | list, find |
| AvailableDID | `client.AvailableDIDs()` | list, find |
| ProofType | `client.ProofTypes()` | list, find |
| PublicKey | `client.PublicKeys()` | list |
| AddressRequirement | `client.AddressRequirements()` | list, find |
| EmergencyRequirement | `client.EmergencyRequirements()` | list, find |
| DIDHistory | `client.DIDHistory()` | list |
| SupportingDocumentTemplate | `client.SupportingDocumentTemplates()` | list, find |
| Balance | `client.Balance()` | find |
| DID | `client.DIDs()` | list, find, update, delete |
| VoiceInTrunk | `client.VoiceInTrunks()` | list, find, create, update, delete |
| VoiceInTrunkGroup | `client.VoiceInTrunkGroups()` | list, find, create, update, delete |
| VoiceOutTrunk | `client.VoiceOutTrunks()` | list, find, create, update, delete |
| VoiceOutTrunkRegenerateCredential | `client.VoiceOutTrunkRegenerateCredentials()` | create |
| SmsTrunk | `client.SmsTrunks()` | list, find, create, update, delete |
| SmsTrunkGroup | `client.SmsTrunkGroups()` | list, find, create, update, delete |
| SmsTrunkRegenerateCredential | `client.SmsTrunkRegenerateCredentials()` | create |
| SenderIDVerification | `client.SenderIDVerifications()` | list, find, update |
| DIDReservation | `client.DIDReservations()` | list, find, create, delete |
| CapacityPool | `client.CapacityPools()` | list, find, update |
| SharedCapacityGroup | `client.SharedCapacityGroups()` | list, find, create, update, delete |
| Order | `client.Orders()` | list, find, create, delete |
| Export | `client.Exports()` | list, find, create, update |
| Address | `client.Addresses()` | list, find, create, update, delete |
| AddressVerification | `client.AddressVerifications()` | list, find, create, update |
| EmergencyCallingService | `client.EmergencyCallingServices()` | list, find, delete |
| EmergencyVerification | `client.EmergencyVerifications()` | list, find, create, update |
| EmergencyRequirementValidation | `client.EmergencyRequirementValidations()` | create |
| Identity | `client.Identities()` | list, find, create, update, delete |
| EncryptedFile | `client.EncryptedFiles()` | list, find, delete |
| PermanentSupportingDocument | `client.PermanentSupportingDocuments()` | create |
| Proof | `client.Proofs()` | create |
| AddressRequirementValidation | `client.AddressRequirementValidations()` | create |
| StockKeepingUnit | include on `DIDGroups` | — |
| QtyBasedPricing | include on `CapacityPools` | — |
| VoiceInPrice | `client.VoiceInPrices()` | lookup, head, download, fetch |
| VoiceOutPrice | `client.VoiceOutPrices()` | lookup, head, download, fetch |
| SMSOutPrice | `client.SMSOutA2PPrices()`, `client.SMSOutP2PPrices()`, `client.SMSOutOTPPrices()` | lookup, head, download, fetch |
| SMSInPrice | `client.SMSInPrices()` | lookup, head, download, fetch |

> **Note:** `StockKeepingUnit` and `QtyBasedPricing` have no standalone API endpoints.
> Access them via `include` on `DIDGroups` and `CapacityPools` respectively.
> The price rows are served by the prices services outside the JSON:API, see [Prices](#prices).

## Date and Datetime Fields

The SDK distinguishes between date-only and datetime fields:

- **Datetime fields** are deserialized as `time.Time` (UTC) when always present, or `*time.Time` when optional (nil if the API omits the value):
  - `CreatedAt` — `time.Time`, present on most resources
  - `UpdatedAt` — `time.Time`: `SmsTrunk`, `SmsTrunkGroup`
  - `ExpiresAt` — `*time.Time`: `DID`, `DIDReservation`, `Proof`, `EncryptedFile`
  - `ActivatedAt` — `*time.Time`: `EmergencyCallingService`, `SenderIDVerification` (nullable)
  - `CanceledAt` — `*time.Time`: `EmergencyCallingService` (nullable)
  - `ClosedAt` — `*time.Time`: `SenderIDVerification` (nullable)
  - `EffectiveAt` — `*time.Time`: `VoiceOutPrice` price lookup rows (nullable)
- **Date-only fields** remain as `string` in `"YYYY-MM-DD"` format — Go has no separate date-only type, so the raw string avoids timezone ambiguity:
  - `Identity.BirthDate`
  - `CapacityPool.RenewDate`, `EmergencyCallingService.RenewDate` (nullable)
  - `SenderIDVerification.RenewDate` — `*string`, nil once the verification is closed
  - Order item `BilledFrom` / `BilledTo`
- **String fields** (not numeric):
  - `EmergencyRequirement.EstimateSetupTime` — e.g. `"7-14 days"`, `"1"`
  - `EmergencyRequirement.RequirementRestrictionMessage` — nullable

**Important changes from previous API versions:**
- `ExpiresAt` replaces `ExpireAt` on `DIDReservation` and `EncryptedFile`
- `RenewDate` is a date-only string, NOT a `time.Time`
- `EstimateSetupTime` is a string, NOT an integer

```go
did, _ := client.DIDs().Find(ctx, "uuid")
fmt.Println(did.CreatedAt)   // 2024-01-15 10:00:00 +0000 UTC
fmt.Println(did.ExpiresAt)   // <nil> or &2025-01-15 10:00:00 +0000 UTC

identity, _ := client.Identities().Find(ctx, "uuid")
fmt.Println(identity.BirthDate)  // "1990-05-20"
```

## Enums

The SDK provides enum types in `github.com/didww/didww-api-3-go-sdk/v4/resource/enums`:

`CallbackMethod`, `IdentityType`, `OrderStatus`, `ExportType`, `ExportStatus`, `CliFormat`,
`OnCliMismatchAction`\*, `MediaEncryptionMode`, `DefaultDstAction`, `VoiceOutTrunkStatus`,
`EmergencyCallingServiceStatus`, `EmergencyVerificationStatus`, `DiversionRelayPolicy`,
`TransportProtocol`, `Codec`, `RxDtmfFormat`, `TxDtmfFormat`, `SstRefreshMethod`,
`ReroutingDisconnectCode`, `Feature`, `AreaLevel`, `AddressVerificationStatus`, `StirShakenMode`,
`NumberType`, `SenderIDVerificationStatus`, `SenderType`, `VoiceOutPriceCategory`

\* `replace_cli` and `randomize_cli` require account configuration.

The same package holds the `Pop*` values of `VoiceInTrunk.Pop` and the `ExportTrunkType*` values of the `trunk_type` filter of `sms_in` exports. The SMS trunk configuration values are constants of package `github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration`: `Type*`, `HTTPMethod*`, `BodyType*` and `ServerLocation*`.

## Webhook Signature Validation

Validate incoming webhook callbacks from DIDWW using HMAC-SHA1 signature verification.

Initialize the validator with the callback secret that is enabled in the DIDWW User Panel (**APIs → DIDWW API 3 → Callback Secrets**). DIDWW signs every callback with it and sends callbacks only while a callback secret is enabled.

```go
import didww "github.com/didww/didww-api-3-go-sdk/v4"

validator := didww.NewRequestValidator("YOUR_CALLBACK_SECRET")

// In your webhook handler:
signature := r.Header.Get(didww.SignatureHeaderName) // "X-DIDWW-Signature"
payload := map[string]string{"key": "value"}         // parsed form/query payload
valid := validator.Validate(requestURL, payload, signature)
```

## Releasing a New Major Version

Go enforces [Semantic Import Versioning](https://go.dev/ref/mod#major-version-suffixes):
every major version ≥ 2 must live at a distinct import path (`/v3`, `/v4`, ...).
Bumping the major version is therefore mechanical — follow the same recipe as
established libraries (e.g. [stripe-go's v84→v85 bump](https://github.com/stripe/stripe-go/commit/cc93bb36f7)).

Steps for `/vN` → `/v(N+1)`:

1. Cut a maintenance branch from the current `main` before starting (e.g.
   `release-N`) and push it, so v(N) can still receive patch releases.
2. Create a feature branch from `main` (e.g. `feat/api-YYYY-MM-DD`).
3. Update the module path in `go.mod`:
   ```diff
   - module github.com/didww/didww-api-3-go-sdk/vN
   + module github.com/didww/didww-api-3-go-sdk/v(N+1)
   ```
4. Rewrite every internal import across the repo:
   ```bash
   grep -rln 'github.com/didww/didww-api-3-go-sdk/vN' --include='*.go' \
     | xargs sed -i '' 's|didww-api-3-go-sdk/vN|didww-api-3-go-sdk/v(N+1)|g'
   ```
5. Update this README: installation snippet, usage imports, and the API
   version compatibility table at the top.
6. Update `examples/` imports the same way.
7. Run `go build ./... && go test ./...` — everything must stay green.
8. Commit as a single atomic change (`chore!: bump module path to /v(N+1)`),
   merge to `main`, then tag `vN+1.0.0` and push the tag.

Existing users on `/vN` are unaffected — they must consciously change their
import path to adopt the new major version.

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/didww/didww-api-3-go-sdk

## License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
