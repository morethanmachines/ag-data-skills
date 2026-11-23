# Vineyard block: known issues

1. **Block lifecycle.** Splits and merges are out of scope for the first build. Renames are
   held as names with `valid_from` and `valid_to`.
2. **Identifiers across organisations.** A block's identifier differs by party (grower,
   wine company, contractor). `aliases` lists them by scheme, but there is no agreed scheme
   naming convention yet.
3. **Variety codes.** The Wine Australia variety list and codes are not yet in
   `codelists/`, pending confirmation of reuse terms. Variety is free text with an optional
   code and code scheme until then.
4. **Region.** Geographical indications are registered in `external/registry.yaml` but not
   yet constrained here.
