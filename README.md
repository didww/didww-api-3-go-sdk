Go client for DIDWW API v3.

![Tests](https://github.com/didww/didww-api-3-go-sdk/actions/workflows/ci.yml/badge.svg)
![Coverage](https://img.shields.io/endpoint?url=https://didww.github.io/didww-api-3-go-sdk/badge.json)
![Go](https://img.shields.io/badge/go-1.23%2B-blue)

## About DIDWW API v3

The DIDWW API provides a simple yet powerful interface that allows you to fully integrate your own applications with DIDWW services. An extensive set of actions may be performed using this API, such as ordering and configuring phone numbers, setting capacity, creating SIP trunks and retrieving CDRs and other operational data.

The DIDWW API v3 is a fully compliant implementation of the [JSON API specification](http://jsonapi.org/format/).

This SDK implements JSON:API serialization and deserialization without external dependencies, using only the Go standard library.

Read more https://doc.didww.com/api

This SDK targets DIDWW API v3 documentation version:
[https://doc.didww.com/api3/2026-04-16/index.html](https://doc.didww.com/api3/2026-04-16/index.html)

The client sends the `X-DIDWW-API-Version: 2026-04-16` header with each request.

Version **3.x** targets API version `2026-04-16`.
Version **2.x** (branch `release-2`) targets API version `2022-05-10`.

## Requirements

- Go 1.23+

## Installation

```bash
go get github.com/didww/didww-api-3-go-sdk
```

## Usage

```go
package main

import (
    "context"
    "fmt"

    didww "github.com/didww/didww-api-3-go-sdk"
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
    didww.WithTimeout(30000), // 30 seconds
)
```

### Custom HTTP Client (Proxy, SSL, etc.)

You can pass a custom `*http.Client` for advanced configuration such as proxy support:

```go
import (
    "net/http"
    "net/url"

    didww "github.com/didww/didww-api-3-go-sdk"
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

The API key header is added automatically. Any other HTTP settings (timeouts, TLS, proxies, transports) can be configured on the `*http.Client`.

### Environments

| Environment | Base URL |
|-------------|----------|
| `didww.Production` | `https://api.didww.com/v3` |
| `didww.Sandbox` | `https://sandbox-api.didww.com/v3` |

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

// POPs (Points of Presence)
pops, _ := client.Pops().List(ctx, nil)

// DID Group Types
types, _ := client.DIDGroupTypes().List(ctx, nil)

// DID Groups (with stock keeping units)
params := didww.NewQueryParams().Include("stock_keeping_units")
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

### Voice In Trunks

```go
import "github.com/didww/didww-api-3-go-sdk/resource/enums"

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

```go
import (
    "github.com/didww/didww-api-3-go-sdk/resource/authenticationmethod"
    "github.com/didww/didww-api-3-go-sdk/resource/enums"
)

trunk := &didww.VoiceOutTrunk{
    Name: "My Outbound Trunk",
    AuthenticationMethod: &authenticationmethod.IpOnly{
        AllowedSipIPs: []string{"203.0.113.0/24"},
    },
    AllowedRtpIPs:       []string{"203.0.113.1"},
    DstPrefixes:         []string{},
    DefaultDstAction:    enums.DefaultDstActionAllowAll,
    OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
    MediaEncryptionMode: enums.MediaEncryptionModeDisabled,
}
created, _ := client.VoiceOutTrunks().Create(ctx, trunk)
```

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
import "github.com/didww/didww-api-3-go-sdk/resource/enums"

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

### Exports

```go
import "github.com/didww/didww-api-3-go-sdk/resource/enums"

export := &didww.Export{
    ExportType: enums.ExportTypeCdrIn,
    Filters:    map[string]interface{}{"from": "2026-04-01 00:00:00", "to": "2026-04-16 00:00:00"},
}
created, _ := client.Exports().Create(ctx, export)
```

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

## Order Item Types

| Type | JSON:API type |
|------|---------------|
| DID | `did_order_items` |
| Available DID | `available_did_order_items` |
| Reservation DID | `reservation_did_order_items` |
| Capacity | `capacity_order_items` |
| Emergency | `emergency_order_items` |
| Generic (response only) | `generic_order_items` |

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

## All Supported Resources

| Resource | Repository | Operations |
|----------|-----------|------------|
| Country | `client.Countries()` | list, find |
| Region | `client.Regions()` | list, find |
| City | `client.Cities()` | list, find |
| Area | `client.Areas()` | list, find |
| NanpaPrefix | `client.NanpaPrefixes()` | list, find |
| Pop | `client.Pops()` | list, find |
| DIDGroupType | `client.DIDGroupTypes()` | list, find |
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

> **Note:** `StockKeepingUnit` and `QtyBasedPricing` have no standalone API endpoints.
> Access them via `include` on `DIDGroups` and `CapacityPools` respectively.

## Date and Datetime Fields

The SDK distinguishes between date-only and datetime fields:

- **Datetime fields** are deserialized as `time.Time` (UTC) when always present, or `*time.Time` when optional (nil if the API omits the value):
  - All `CreatedAt` fields — `time.Time`, present on most resources
  - Expiry fields — `*time.Time`: `DID.ExpiresAt`, `Proof.ExpiresAt`, `EncryptedFile.ExpiresAt`; `DIDReservation.ExpiresAt` is `time.Time` (always present)
- **Date-only fields** (`Identity.BirthDate`, `CapacityPool.RenewDate`, order item `BilledFrom`/`BilledTo`) remain as `string` in `"YYYY-MM-DD"` format — Go has no separate date-only type, so the raw string avoids timezone ambiguity.

```go
did, _ := client.DIDs().Find(ctx, "uuid")
fmt.Println(did.CreatedAt)   // 2024-01-15 10:00:00 +0000 UTC
fmt.Println(did.ExpiresAt)   // <nil> or &2025-01-15 10:00:00 +0000 UTC

identity, _ := client.Identities().Find(ctx, "uuid")
fmt.Println(identity.BirthDate)  // "1990-05-20"
```

## Enums

The SDK provides enum types in `github.com/didww/didww-api-3-go-sdk/resource/enums`:

`CallbackMethod`, `IdentityType`, `OrderStatus`, `ExportType`, `ExportStatus`, `CliFormat`,
`OnCliMismatchAction`\*, `MediaEncryptionMode`, `DefaultDstAction`, `VoiceOutTrunkStatus`,
`TransportProtocol`, `Codec`, `RxDtmfFormat`, `TxDtmfFormat`, `SstRefreshMethod`,
`ReroutingDisconnectCode`, `Feature`, `AreaLevel`, `AddressVerificationStatus`, `StirShakenMode`,
`DiversionRelayPolicy`

\* `replace_cli` and `randomize_cli` require account configuration.

## Webhook Signature Validation

Validate incoming webhook callbacks from DIDWW using HMAC-SHA1 signature verification.

```go
import didww "github.com/didww/didww-api-3-go-sdk"

validator := didww.NewRequestValidator("YOUR_API_KEY")

// In your webhook handler:
signature := r.Header.Get(didww.SignatureHeaderName) // "X-DIDWW-Signature"
payload := map[string]string{"key": "value"}         // parsed form/query payload
valid := validator.Validate(requestURL, payload, signature)
```

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/didww/didww-api-3-go-sdk

## License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
