# Examples

All examples read the API key from the `DIDWW_API_KEY` environment variable.

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
| [`did_trunk_assignment`](did_trunk_assignment/) | Demonstrates exclusive trunk/trunk group assignment on DIDs. |
| [`did_sms_routing`](did_sms_routing/) | Routes a DID's inbound SMS to an SMS trunk, then to an SMS trunk group (exclusive assignment), then unroutes it. |
| [`dids`](dids/) | Updates DID routing/capacity by assigning trunk and capacity pool. |
| [`trunks`](trunks/) | Lists trunks with their POP, creates SIP and PSTN trunks, updates and deletes them. |
| [`shared_capacity_groups`](shared_capacity_groups/) | Creates a shared capacity group in a capacity pool. |
| [`orders`](orders/) | Lists orders and creates/cancels a DID order using live SKU lookup. |
| [`orders_sku`](orders_sku/) | Creates a DID order by SKU resolved from DID groups. |
| [`orders_nanpa`](orders_nanpa/) | Orders a DID number by NPA/NXX prefix. |
| [`orders_capacity`](orders_capacity/) | Purchases capacity by creating a capacity order item. |
| [`orders_available_dids`](orders_available_dids/) | Orders an available DID using included DID group SKU. |
| [`orders_reservation_dids`](orders_reservation_dids/) | Reserves a DID and then places an order from that reservation. |
| [`voice_in_trunk_groups`](voice_in_trunk_groups/) | CRUD for trunk groups with trunk relationships. |
| [`voice_in_trunk_sip_registration`](voice_in_trunk_sip_registration/) | End-to-end SIP registration flow: create with `EnabledSipRegistration=true`, rename, disable by setting `Host`, re-enable by toggling the flag. The SDK keeps the dependent fields (`Host`, `Port`, `UseDIDInRuri`) aligned with the server's validation rules automatically. |
| [`voice_out_trunks`](voice_out_trunks/) | CRUD for voice out trunks using the polymorphic authentication_method. |
| [`did_reservations`](did_reservations/) | Creates, lists, finds and deletes DID reservations. |
| [`exports`](exports/) | Creates and lists CDR exports with external_reference_id. |
| [`exports_sms`](exports_sms/) | Creates SMS-In and SMS-Out exports with their filters and reads a completed one. |
| [`sms_trunks`](sms_trunks/) | Lists SMS trunks, creates an HTTP IN trunk, updates its configuration, regenerates the credentials of an HTTP OUT trunk and deletes both. |
| [`sms_trunk_groups`](sms_trunk_groups/) | Creates two inbound SMS trunks, groups them by priority, lists groups and deletes the group with its trunks. |
| [`sender_id_verifications`](sender_id_verifications/) | Lists Sender ID verifications with their SMS trunks, routes an active one through an SMS trunk that sends messages and restores its original trunk. |
| [`prices_lookup`](prices_lookup/) | Looks up Voice IN, Voice OUT, SMS OUT (A2P, P2P, OTP) and SMS IN prices for a number. |
| [`prices_download`](prices_download/) | Downloads the gzipped Voice IN price file to the directory given as argument, or to `didww-prices` in the user cache directory, re-checks it with If-Modified-Since and reads a few rows. |
| [`capacity_pools`](capacity_pools/) | Lists capacity pools with included shared capacity groups. |
| [`did_history`](did_history/) | Lists DID ownership history. |
| [`identities`](identities/) | Lists identities with country and birth_country. |
| [`emergency_requirements`](emergency_requirements/) | Lists emergency service requirements per country/number_type. |
| [`emergency_calling_services`](emergency_calling_services/) | Lists emergency calling services. |
| [`emergency_verifications`](emergency_verifications/) | Lists emergency verifications. |
| [`emergency_requirement_validations`](emergency_requirement_validations/) | Validates emergency requirement data. |
| [`emergency_scenario`](emergency_scenario/) | End-to-end: find DID → check requirements → validate → create verification → get service. |
| [`address_verifications`](address_verifications/) | Lists address verifications with reject_comment / external_reference_id. |
| [`orders_emergency`](orders_emergency/) | Creates an emergency order with EmergencyOrderItem. |

## Troubleshooting

If `DIDWW_API_KEY` is missing, examples fail fast with:

`DIDWW_API_KEY environment variable is required`
