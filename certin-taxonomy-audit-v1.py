#!/usr/bin/env python3
"""Read-only structural audit of a Certin taxonomy JSON pack."""
import argparse
import json
import sys
from collections import defaultdict
from pathlib import Path


def audit(data):
    errors = []
    warnings = []
    techs = data.get('technologies', [])
    ids = {}
    labels = {}
    for i, item in enumerate(techs):
        ident = item.get('id', '')
        where = f'technologies[{i}]'
        if not isinstance(ident, str) or not ident.strip():
            errors.append(f'{where}: missing ID')
            continue
        if ident in ids:
            errors.append(f'{where}: duplicate ID {ident!r}')
        ids[ident] = item
        for label in [item.get('label', '')] + item.get('aliases', []):
            if not isinstance(label, str) or not label.strip():
                errors.append(f'{where}: empty or invalid label/alias')
                continue
            normalized = label.strip().casefold()
            if normalized in labels and labels[normalized] != ident:
                errors.append(f'{where}: ambiguous label {label!r} shared by {labels[normalized]} and {ident}')
            labels[normalized] = ident
        categories = item.get('categories', [])
        if not isinstance(categories, list) or any(not isinstance(c, str) or not c.strip() for c in categories):
            errors.append(f'{where}: invalid categories')
        elif len(set(c.casefold().strip() for c in categories)) != len(categories):
            errors.append(f'{where}: duplicate category')
        for p in item.get('provenance', []):
            if not isinstance(p, dict) or any(not p.get(k) for k in ('source', 'source_commit', 'license', 'modification')):
                errors.append(f'{where}: incomplete provenance')
    edges = data.get('technology_relationships', [])
    graph = defaultdict(list)
    seen_edges = set()
    for i, edge in enumerate(edges):
        child, parent, relation = (edge.get(k) for k in ('child_id', 'parent_id', 'relation'))
        if relation != 'feature_of':
            errors.append(f'relationship[{i}]: unsupported relation {relation!r}')
        if child not in ids or parent not in ids:
            errors.append(f'relationship[{i}]: missing endpoint {child!r} -> {parent!r}')
        if child == parent:
            errors.append(f'relationship[{i}]: self relationship {child!r}')
        key = (child, parent, relation)
        if key in seen_edges:
            errors.append(f'relationship[{i}]: duplicate edge {key!r}')
        seen_edges.add(key)
        if child in ids and parent in ids:
            graph[child].append(parent)
    colors = {}
    def visit(node):
        if colors.get(node) == 1:
            errors.append(f'feature_of cycle involving {node!r}')
            return
        if colors.get(node) == 2:
            return
        colors[node] = 1
        for parent in graph[node]:
            visit(parent)
        colors[node] = 2
    for ident in ids:
        visit(ident)
    firestore = ids.get('firestore')
    if firestore:
        if 'Cloud Firestore' not in firestore.get('aliases', []):
            warnings.append('firestore missing Cloud Firestore alias')
        if not {'Database', 'Mobile (Firebase)'}.issubset(set(firestore.get('categories', []))):
            warnings.append('firestore missing shared source categories')
    return errors, warnings, len(ids), len(edges)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('taxonomy', type=Path)
    args = parser.parse_args()
    try:
        data = json.loads(args.taxonomy.read_text(encoding='utf-8'))
        errors, warnings, count, edges = audit(data)
    except (OSError, ValueError, TypeError) as exc:
        print(f'ERROR: {exc}', file=sys.stderr)
        return 2
    for message in errors:
        print(f'ERROR: {message}')
    for message in warnings:
        print(f'WARNING: {message}')
    print(f'Audit: {count} technologies, {edges} relationships, {len(errors)} errors, {len(warnings)} warnings')
    return 1 if errors else 0

if __name__ == '__main__':
    sys.exit(main())
