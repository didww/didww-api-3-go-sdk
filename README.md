Go client for DIDWW API v3.

## About DIDWW API v3

The DIDWW API provides a simple yet powerful interface that allows you to fully integrate your own applications with DIDWW services. An extensive set of actions may be performed using this API, such as ordering and configuring phone numbers, setting capacity, creating SIP trunks and retrieving CDRs and other operational data.

The DIDWW API v3 is a fully compliant implementation of the [JSON API specification](http://jsonapi.org/format/).

Read more https://doc.didww.com/api

This SDK targets DIDWW API v3 documentation version:
[https://doc.didww.com/api3/2022-05-10/index.html](https://doc.didww.com/api3/2022-05-10/index.html)

The client sends the `X-DIDWW-API-Version: 2022-05-10` header with each request.

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

// Requirements
requirements, _ := client.Requirements().List(ctx, nil)

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
trunk := &didww.VoiceInTrunk{
    Name:           "My SIP Trunk",
    Priority:       1,
    Weight:         100,
    CliFormat:      enums.CliFormatE164,
    RingingTimeout: 30,
    Configuration: &didww.SIPConfiguration{
        Host:                "sip.example.com",
        Port:                5060,
        CodecIDs:            []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
        TransportProtocolID: enums.TransportProtocolUDP,
    },
}
created, _ := client.VoiceInTrunks().Create(ctx, trunk)

// Update trunk
created.Description = "Updated"
updated, _ := client.VoiceInTrunks().Update(ctx, created)

// Delete trunk
client.VoiceInTrunks().Delete(ctx, created.ID)
```

### Voice In Trunk Groups

```go
group := &didww.VoiceInTrunkGroup{
    Name:            "Primary Group",
    CapacityLimit:   50,
    VoiceInTrunkIDs: []string{trunkA.ID, trunkB.ID},
}
created, _ := client.VoiceInTrunkGroups().Create(ctx, group)
```

### Voice Out Trunks

> **Note:** Voice Out Trunks require additional account configuration. Contact DIDWW support to enable.
> The `replace_cli` and `randomize_cli` values of `OnCliMismatchAction` also require account configuration.

```go
import "github.com/didww/didww-api-3-go-sdk/resource/enums"

trunk := &didww.VoiceOutTrunk{
    Name:                "My Outbound Trunk",
    AllowedSipIPs:       []string{"0.0.0.0/0"},
    DefaultDstAction:    enums.DefaultDstActionAllowAll,
    OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
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
    CapacityPoolID:       "pool-uuid",
}
created, _ := client.SharedCapacityGroups().Create(ctx, group)
```

### Identities

```go
identity := &didww.Identity{
    FirstName:    "John",
    LastName:     "Doe",
    PhoneNumber:  "12125551234",
    IdentityType: "Personal",
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
export := &didww.Export{
    ExportType: "cdr_in",
    Filters:    map[string]interface{}{"year": 2025, "month": 1},
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
| Requirement | `client.Requirements()` | list, find |
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
| Export | `client.Exports()` | list, find, create |
| Address | `client.Addresses()` | list, find, create, update, delete |
| AddressVerification | `client.AddressVerifications()` | list, find, create |
| Identity | `client.Identities()` | list, find, create, update, delete |
| EncryptedFile | `client.EncryptedFiles()` | list, find, delete |
| PermanentSupportingDocument | `client.PermanentSupportingDocuments()` | create |
| Proof | `client.Proofs()` | create |
| RequirementValidation | `client.RequirementValidations()` | create |
| StockKeepingUnit | include on `DIDGroups` | — |
| QtyBasedPricing | include on `CapacityPools` | — |

> **Note:** `StockKeepingUnit` and `QtyBasedPricing` have no standalone API endpoints.
> Access them via `include` on `DIDGroups` and `CapacityPools` respectively.

## Enums

The SDK provides enum types in `github.com/didww/didww-api-3-go-sdk/resource/enums`:

`CallbackMethod`, `IdentityType`, `OrderStatus`, `ExportType`, `ExportStatus`, `CliFormat`,
`OnCliMismatchAction`\*, `MediaEncryptionMode`, `DefaultDstAction`, `VoiceOutTrunkStatus`,
`TransportProtocol`, `Codec`, `RxDtmfFormat`, `TxDtmfFormat`, `SstRefreshMethod`,
`ReroutingDisconnectCode`, `Feature`, `AreaLevel`, `AddressVerificationStatus`, `StirShakenMode`

\* `replace_cli` and `randomize_cli` require account configuration.

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/didww/didww-api-3-go-sdk

## License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
