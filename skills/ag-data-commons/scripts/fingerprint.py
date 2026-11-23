#!/usr/bin/env python3
"""Fingerprint helpers for the Ag Data Skills commons. Standard library only.

  python3 fingerprint.py columns FILE            normalised column names of a CSV, JSON or xlsx header
  python3 fingerprint.py id FINGERPRINT.json     fingerprint id (same algorithm as the commons ci)
  python3 fingerprint.py match FILE [INDEX_DIR]  score FILE's columns against known fingerprints

Nothing here reads or prints data values: only header names.
"""
import csv, hashlib, json, os, re, sys, zipfile

HERE = os.path.dirname(os.path.abspath(__file__))
SNAPSHOT = os.path.join(os.path.dirname(HERE), "snapshot")


def normalise(name):
    s = name.strip().lower().replace("%", " pct ").replace("°", "")
    return re.sub(r"[^a-z0-9]+", "_", s).strip("_")


def fingerprint_id(fp):
    req = sorted(normalise(c["name"]) for c in fp["columns"] if not c.get("optional"))
    return "fp-" + hashlib.sha256("\n".join(req).encode()).hexdigest()[:12]


def match(file_cols, fp_cols):
    """Same formula as ci/internal/fingerprint.Match (without value patterns)."""
    known = {normalise(c["name"]): c for c in fp_cols}
    required = {n for n, c in known.items() if not c.get("optional")}
    file = {normalise(c) for c in file_cols}
    matched = len(file & set(known))
    unknown = len(file - set(known))
    denom = len(required) + unknown + len({n for n in file if n in known and known[n].get("optional")})
    return matched / denom if denom else 0.0


def xlsx_header_rows(path, rows=4):
    """First few rows of the first sheet, as text (no third-party libraries)."""
    ns = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
    import xml.etree.ElementTree as ET
    with zipfile.ZipFile(path) as z:
        shared = []
        if "xl/sharedStrings.xml" in z.namelist():
            for si in ET.fromstring(z.read("xl/sharedStrings.xml")).iter(ns + "si"):
                shared.append("".join(t.text or "" for t in si.iter(ns + "t")))
        sheets = sorted(n for n in z.namelist() if n.startswith("xl/worksheets/sheet"))
        out = []
        for row in ET.fromstring(z.read(sheets[0])).iter(ns + "row"):
            vals = []
            for c in row.iter(ns + "c"):
                v = c.find(ns + "v")
                t = c.get("t")
                if t == "s" and v is not None:
                    vals.append(shared[int(v.text)])
                elif t == "inlineStr":
                    vals.append("".join(x.text or "" for x in c.iter(ns + "t")))
                else:
                    vals.append(v.text if v is not None else "")
            out.append(vals)
            if len(out) >= rows:
                break
        return out


def columns(path):
    ext = os.path.splitext(path)[1].lower()
    if ext in (".csv", ".tsv", ".txt"):
        with open(path, newline="", encoding="utf-8-sig") as f:
            return next(csv.reader(f, delimiter="\t" if ext == ".tsv" else ","))
    if ext == ".json":
        with open(path, encoding="utf-8") as f:
            doc = json.load(f)
        def first_records(d):
            if isinstance(d, list) and d and isinstance(d[0], dict):
                return d
            if isinstance(d, dict):
                for v in d.values():
                    r = first_records(v)
                    if r:
                        return r
            return None
        recs = first_records(doc) or []
        return list(recs[0].keys()) if recs else list(doc.keys())
    if ext == ".xlsx":
        rows = xlsx_header_rows(path)
        # Report the first row with several non-empty cells; layouts with title rows or
        # merged headers need describing by hand in the fingerprint.
        for r in rows:
            if sum(1 for v in r if v) >= 3:
                return [v for v in r if v]
        return rows[0] if rows else []
    raise SystemExit(f"unsupported file type: {ext}")


def main(argv):
    if len(argv) < 3:
        raise SystemExit(__doc__)
    cmd, path = argv[1], argv[2]
    if cmd == "columns":
        for c in columns(path):
            print(normalise(c))
    elif cmd == "id":
        with open(path) as f:
            print(fingerprint_id(json.load(f)))
    elif cmd == "match":
        index_dir = argv[3] if len(argv) > 3 else os.path.join(SNAPSHOT, "index")
        cols = columns(path)
        hits = []
        for name in sorted(os.listdir(index_dir)) if os.path.isdir(index_dir) else []:
            if not name.endswith(".json"):
                continue
            with open(os.path.join(index_dir, name)) as f:
                ix = json.load(f)
            for fp in ix.get("fingerprints") or []:
                fpc = [{"name": n} for n in fp.get("required_columns", [])] + \
                      [{"name": n, "optional": True} for n in fp.get("optional_columns") or []]
                hits.append((match(cols, fpc), fp["id"], fp.get("crosswalk") or "(sighting only)"))
        print("file columns:", ", ".join(normalise(c) for c in cols))
        if not hits:
            print("no known fingerprints: map directly to the spec")
        for score, fid, cw in sorted(hits, reverse=True):
            print(f"{score:.2f} {fid} {cw}")
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main(sys.argv)
