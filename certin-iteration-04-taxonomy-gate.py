#!/usr/bin/env python3
"""Read-only Iteration 4 taxonomy gate. Usage: python3 script.py taxonomy.json [report.json]"""
import json,sys,collections,pathlib
p=pathlib.Path(sys.argv[1]); d=json.loads(p.read_text()); ts=d.get('technologies',[]); rs=d.get('technology_relationships',[])
issues=[]; ids=[t.get('id') for t in ts]; counts=collections.Counter(ids)
issues += [f'duplicate technology id: {k}' for k,v in counts.items() if v>1]
lookup={t['id']:t for t in ts if 'id' in t}
for r in rs:
    for key in ('child_id','parent_id'):
        if r.get(key) not in lookup: issues.append(f'unknown relationship {key}: {r.get(key)}')
    if r.get('relation')!='feature_of': issues.append(f'unexpected relationship: {r}')
for t in ts:
    if t.get('review_status')=='curator_approved' and t.get('categories'):
        if not t.get('provenance'): issues.append(f'approved sourced entry missing provenance: {t["id"]}')
        for pr in t.get('provenance',[]):
            for k in ('source','source_commit','license','source_line','modification'):
                if not pr.get(k): issues.append(f'{t["id"]} missing provenance.{k}')
expected={'id':'google-cloud','domains':8,'technologies':94,'relationships':5}
actual={'id':d.get('id'),'domains':len(d.get('domains',[])),'technologies':len(ts),'relationships':len(rs)}
for k,v in expected.items():
    if actual[k]!=v:issues.append(f'{k}: expected {v!r}, got {actual[k]!r}')
firestore=lookup.get('firestore',{})
if 'Cloud Firestore' not in firestore.get('aliases',[]):issues.append('Firestore alias missing')
if not {'Database','Mobile (Firebase)'}.issubset(firestore.get('categories',[])):issues.append('Firestore dual source classification missing')
if 'Cloud Memorystore' not in lookup.get('memorystore',{}).get('aliases',[]):issues.append('Memorystore alias missing')
expected_edges={('bigquery-ml','bigquery'),('bigquery-bi-engine','bigquery'),('bigquery-data-transfer-service','bigquery'),('bigquery-gis','bigquery'),('cloud-sql-insights','cloud-sql')}
actual_edges={(r.get('child_id'),r.get('parent_id')) for r in rs}
if actual_edges!=expected_edges:issues.append('feature_of edge set differs from approved baseline')
a=d.get('attribution',{})
for k,v in {'creator':'Priyanka Vergadia','license':'CC-BY-4.0','source_commit':'6944032502c3f3bad4274352679c313f9f166263'}.items():
    if a.get(k)!=v:issues.append(f'attribution {k} mismatch')
report={'gate':'iteration-04-taxonomy','status':'PASS' if not issues else 'FAIL','source_file':str(p),'counts':actual,'feature_nodes':sum(t.get('kind')=='feature' for t in ts),'pending_reconciliation':sum(t.get('review_status')=='pending_cheat_sheet_reconciliation' for t in ts),'errors':issues,'note':'Read-only artifact gate; does not test Hugo, quiz behavior or compiled question data.'}
print(json.dumps(report,indent=2))
if len(sys.argv)>2:pathlib.Path(sys.argv[2]).write_text(json.dumps(report,indent=2)+'\n')
sys.exit(bool(issues))
