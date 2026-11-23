#!/usr/bin/env python3
"""Validate a mapped document against a commons schema or profile from the bundled snapshot.

  python3 validate.py SCHEMA_OR_PROFILE.json DOCUMENT.json

Uses the `jsonschema` package (JSON Schema 2020-12) if it is installed, resolving every
commons $id to the local snapshot so nothing is fetched. Without it, falls back to a basic
check of required properties and says so.
"""
import json, os, sys

BASE = "https://raw.githubusercontent.com/morethanmachines/ag-data-skills/main/"
HERE = os.path.dirname(os.path.abspath(__file__))
SNAPSHOT = os.path.join(os.path.dirname(HERE), "snapshot")


def load(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def local_path(url):
    rel = url[len(BASE):].split("#")[0]
    return os.path.join(SNAPSHOT, rel)


def basic_check(schema, doc, root, path="$"):
    """Required-property check only, following local $defs refs."""
    problems = []
    def res(s):
        while isinstance(s, dict) and isinstance(s.get("$ref"), str) and s["$ref"].startswith("#/"):
            node = root
            for part in s["$ref"][2:].split("/"):
                node = node[part]
            s = node
        return s
    s = res(schema)
    if not isinstance(s, dict):
        return problems
    if isinstance(doc, dict):
        for r in s.get("required", []):
            if r not in doc:
                problems.append(f"{path}: missing required property {r!r}")
        for k, sub in (s.get("properties") or {}).items():
            if k in doc:
                problems += basic_check(sub, doc[k], root, f"{path}.{k}")
    elif isinstance(doc, list) and isinstance(s.get("items"), dict):
        for i, x in enumerate(doc):
            problems += basic_check(s["items"], x, root, f"{path}[{i}]")
    return problems


def main(argv):
    if len(argv) != 3:
        raise SystemExit(__doc__)
    schema, doc = load(argv[1]), load(argv[2])
    try:
        from jsonschema import Draft202012Validator
        try:
            from referencing import Registry, Resource
            def retrieve(uri):
                if uri.startswith(BASE):
                    return Resource.from_contents(load(local_path(uri)))
                raise LookupError(f"refusing to fetch {uri}")
            registry = Registry(retrieve=retrieve)
            v = Draft202012Validator(schema, registry=registry)
        except ImportError:
            from jsonschema import RefResolver
            store = {}
            for d, _, files in os.walk(SNAPSHOT):
                for f in files:
                    if f.endswith(".json"):
                        try:
                            s = load(os.path.join(d, f))
                        except Exception:
                            continue
                        if isinstance(s, dict) and "$id" in s:
                            store[s["$id"]] = s
            v = Draft202012Validator(schema, resolver=RefResolver.from_schema(schema, store=store))
        errors = sorted(v.iter_errors(doc), key=lambda e: list(e.absolute_path))
        for e in errors[:50]:
            print("/".join(str(p) for p in e.absolute_path) or "(root)", ":", e.message[:300])
        print(f"{len(errors)} error(s)" if errors else "valid (JSON Schema 2020-12)")
        sys.exit(1 if errors else 0)
    except ImportError:
        root = schema
        targets = [schema]
        for sub in schema.get("allOf", []):
            ref = sub.get("$ref", "")
            if ref and not ref.startswith("#"):
                p = os.path.normpath(os.path.join(os.path.dirname(os.path.abspath(argv[1])), ref))
                targets.append(load(p))
        problems = []
        for t in targets:
            problems += basic_check(t, doc, t)
        for p in problems[:50]:
            print(p)
        print("jsonschema is not installed: only required properties were checked.")
        sys.exit(1 if problems else 0)


if __name__ == "__main__":
    main(sys.argv)
