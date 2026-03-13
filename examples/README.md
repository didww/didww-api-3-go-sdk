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
| [`voice_out_trunks`](voice_out_trunks/) | CRUD for voice out trunks (requires account config). |
| [`did_reservations`](did_reservations/) | Creates, lists, finds and deletes DID reservations. |
| [`exports`](exports/) | Creates and lists CDR exports. |
| [`capacity_pools`](capacity_pools/) | Lists capacity pools with included shared capacity groups. |

## Troubleshooting

If `DIDWW_API_KEY` is missing, examples fail fast with:

`DIDWW_API_KEY environment variable is required`
