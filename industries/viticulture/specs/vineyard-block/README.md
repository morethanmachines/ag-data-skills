# Vineyard block (`viticulture/vineyard-block`)

Status: **draft 0.1.0**. The management unit viticulture records refer to. A document is
`{ "spec": "viticulture/vineyard-block", "spec_version": "0.1.0", "blocks": [...] }`.

The key property is `names`: every name a block is known by, with the role of the party
that uses it (`owner` for the grower, `buyer` for the wine company) and, for renames, the
dates each name was in use. Records from different systems join on these names.
