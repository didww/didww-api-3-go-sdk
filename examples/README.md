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
| [`dids`](dids/) | Updates DID routing/capacity by assigning trunk and capacity pool. |
| [`trunks`](trunks/) | Lists trunks, creates SIP and PSTN trunks, updates and deletes them. |
| [`shared_capacity_groups`](shared_capacity_groups/) | Creates a shared capacity group in a capacity pool. |
| [`orders`](orders/) | Lists orders and creates/cancels a DID order using live SKU lookup. |
| [`orders_sku`](orders_sku/) | Creates a DID order by SKU resolved from DID groups. |
| [`orders_nanpa`](orders_nanpa/) | Orders a DID number by NPA/NXX prefix. |
| [`orders_capacity`](orders_capacity/) | Purchases capacity by creating a capacity order item. |
| [`orders_available_dids`](orders_available_dids/) | Orders an available DID using included DID group SKU. |
| [`orders_reservation_dids`](orders_reservation_dids/) | Reserves a DID and then places an order from that reservation. |
| [`voice_in_trunk_groups`](voice_in_trunk_groups/) | CRUD for trunk groups with trunk relationships. |
| [`voice_out_trunks`](voice_out_trunks/) | CRUD for voice out trunks using 2026-04-16 polymorphic authentication_method. |
| [`did_reservations`](did_reservations/) | Creates, lists, finds and deletes DID reservations. |
| [`exports`](exports/) | Creates and lists CDR exports with 2026-04-16 external_reference_id. |
| [`capacity_pools`](capacity_pools/) | Lists capacity pools with included shared capacity groups. |
| [`did_history`](did_history/) | Lists DID ownership history (2026-04-16). |
| [`identities`](identities/) | Lists identities with country and birth_country (2026-04-16). |
| [`emergency_requirements`](emergency_requirements/) | Lists emergency service requirements (2026-04-16). |
| [`emergency_calling_services`](emergency_calling_services/) | Lists emergency calling services (2026-04-16). |
| [`emergency_verifications`](emergency_verifications/) | Lists emergency verifications (2026-04-16). |
| [`emergency_requirement_validations`](emergency_requirement_validations/) | Validates emergency requirement data (2026-04-16). |
| [`emergency_scenario`](emergency_scenario/) | End-to-end: find DID → check requirements → validate → create verification → get service. |
| [`address_verifications`](address_verifications/) | Lists address verifications with reject_comment / external_reference_id (2026-04-16). |
| [`orders_emergency`](orders_emergency/) | Creates an emergency order with EmergencyOrderItem (2026-04-16). |

## Troubleshooting

If `DIDWW_API_KEY` is missing, examples fail fast with:

`DIDWW_API_KEY environment variable is required`
