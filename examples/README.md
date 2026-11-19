# Examples

All examples except `webhook` read the API key from the `DIDWW_API_KEY` environment variable and run against the sandbox.

## Prerequisites

- Go 1.23+
- DIDWW API key for sandbox account

## Environment variables

- `DIDWW_API_KEY` (required): your DIDWW API key

## Run an example

```bash
DIDWW_API_KEY=your_api_key go run ./examples/balance/
```

## Available examples

| Directory | Description |
|---|---|
| [`balance`](balance/) | Fetches and prints current account balance and credit. |
| [`countries`](countries/) | Lists countries, demonstrates filtering, and fetches one country by ID. |
| [`regions`](regions/) | Lists regions with filters/includes and fetches a specific region. |
| [`did_groups`](did_groups/) | Fetches DID groups with included SKUs and shows group details. |
| [`did_history`](did_history/) | Lists DID ownership history. |
| [`dids`](dids/) | Lists the most recent DID with its emergency relationships and updates its description and capacity. |
| [`did_trunk_assignment`](did_trunk_assignment/) | Assigns a DID to a new trunk, then to a new trunk group (exclusive assignment), restores the DID's routing and deletes the trunk and group. |
| [`did_sms_routing`](did_sms_routing/) | Routes a DID's inbound SMS to an SMS trunk, then to an SMS trunk group (exclusive assignment), then unroutes it. |
| [`did_reservations`](did_reservations/) | Reserves an available DID, lists and finds reservations, and deletes the reservation. |
| [`capacity_pools`](capacity_pools/) | Lists capacity pools with included shared capacity groups and qty-based pricings, and finds one pool. |
| [`shared_capacity_groups`](shared_capacity_groups/) | Creates a shared capacity group in a capacity pool, lists the groups, updates and deletes it. |
| [`orders`](orders/) | Lists recent orders, orders a DID by SKU and cancels (deletes) that order. |
| [`orders_sku`](orders_sku/) | Orders DIDs by SKU and quantity (one `did_order_items` item). |
| [`orders_available_dids`](orders_available_dids/) | Orders an available DID using included DID group SKU. |
| [`orders_reservation_dids`](orders_reservation_dids/) | Reserves a DID and then places an order from that reservation. |
| [`orders_nanpa`](orders_nanpa/) | Orders a DID number by NPA/NXX prefix. |
| [`orders_capacity`](orders_capacity/) | Purchases capacity by creating a capacity order item. |
| [`orders_all_item_types`](orders_all_item_types/) | Places one order with the three DID item variants: by SKU and quantity, by available DID and by reservation. |
| [`orders_emergency`](orders_emergency/) | Inspects server-created Emergency orders and `emergency_order_items`. |
| [`trunks`](trunks/) | Lists Voice IN trunks with their trunk group, creates a SIP and a PSTN trunk, finds, updates and deletes them. |
| [`voice_in_trunk_groups`](voice_in_trunk_groups/) | Creates two trunks and a trunk group with them, lists groups, renames the group and deletes the group and the trunks. |
| [`voice_in_trunk_sip_registration`](voice_in_trunk_sip_registration/) | End-to-end SIP registration flow: create with `EnabledSipRegistration=true`, rename, disable by setting `Host`, re-enable by toggling the flag. The SDK keeps the dependent fields (`Host`, `Port`, `UseDIDInRuri`) aligned with the server's validation rules automatically. |
| [`voice_out_trunks`](voice_out_trunks/) | CRUD for voice out trunks using the polymorphic authentication_method. |
| [`sms_trunks`](sms_trunks/) | Lists SMS trunks, creates an HTTP IN trunk, updates its configuration, regenerates the credentials of an HTTP OUT trunk and deletes both. |
| [`sms_trunk_groups`](sms_trunk_groups/) | Creates two inbound SMS trunks, groups them by priority, lists groups and deletes the group with its trunks. |
| [`sender_id_verifications`](sender_id_verifications/) | Lists Sender ID verifications with their SMS trunks, routes an active one through an SMS trunk that sends messages and restores its original trunk. |
| [`identity_address_proofs`](identity_address_proofs/) | Creates an identity and an address, uploads an encrypted file, attaches proofs to both, prints their verified flags and deletes them. |
| [`address_verifications`](address_verifications/) | Lists address verifications with reject_comment / external_reference_id. |
| [`upload_file`](upload_file/) | Encrypts a file with the DIDWW public keys, uploads it and prints the encrypted file ID. |
| [`exports`](exports/) | Creates and lists CDR exports with external_reference_id. |
| [`exports_sms`](exports_sms/) | Creates SMS-In and SMS-Out exports with their filters and reads a completed one. |
| [`emergency_requirements`](emergency_requirements/) | Lists emergency service requirements per country/number_type. |
| [`emergency_requirement_validations`](emergency_requirement_validations/) | Validates emergency requirement data. |
| [`emergency_verifications`](emergency_verifications/) | Lists emergency verifications. |
| [`emergency_calling_services`](emergency_calling_services/) | Lists emergency calling services. |
| [`emergency_scenario`](emergency_scenario/) | End-to-end: find DID → check requirements → validate → create verification → get service. |
| [`prices_lookup`](prices_lookup/) | Looks up Voice IN, Voice OUT, SMS OUT (A2P, P2P, OTP) and SMS IN prices for a number. |
| [`prices_download`](prices_download/) | Downloads the gzipped Voice IN price file to the directory given as argument, or to `didww-prices` in the user cache directory, re-checks it with If-Modified-Since and reads a few rows. |
| [`includes`](includes/) | Reads regions, DID groups, available DIDs and DIDs with included relationships: the SDK fills a relationship from the response's included data, so request the ones you read with `Include`. |
| [`webhook`](webhook/) | Signs a sample callback the way DIDWW does and validates it with `RequestValidator`: `valid: true`, then `valid: false` for a tampered payload. No API key needed. |
| [`test_readme`](test_readme/) | Runs the code snippets of the SDK README against the sandbox, cleans up, prints `Results: N passed, M failed` and exits 1 on a failure. |

## Troubleshooting

If `DIDWW_API_KEY` is missing, examples fail fast with:

`DIDWW_API_KEY environment variable is required`
