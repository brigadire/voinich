#!/usr/bin/env python3
"""Validate a complete blinded export and freeze a new mapping snapshot."""
from __future__ import annotations

import argparse
import csv
from datetime import datetime
import json
from pathlib import Path
import sys

sys.dont_write_bytecode=True
PKG=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PKG/'scripts'))
import build_preparation as b

FIELDS=('review_id','label_id','panel','candidate_snapshot_sha256','review_action','mapping_outcome','selected_occurrence_ids','reading_order','mapping_confidence','reviewer_id','decision_timestamp','ambiguity_status','notes','completion_status')
OUTCOMES={'SINGLE_TOKEN','MULTI_TOKEN','RING_SEQUENCE','AMBIGUOUS','UNREADABLE','NO_TRANSCRIPTION_MATCH','OUT_OF_TRANSCRIPTION_SCOPE'}
ACTIONS={'CONFIRMED_TOKEN','CONFIRMED_SEQUENCE','CORRECTED_MAPPING','AMBIGUOUS','UNREADABLE','NO_MATCH'}
ORDERS={'CORPUS_ORDER','CLOCKWISE','COUNTERCLOCKWISE','VISUAL_LEFT_TO_RIGHT','UNDETERMINED'}
CONFIDENCE={'LOW','MEDIUM','HIGH'}
AMBIGUITY={'CLEAR','AMBIGUOUS','UNREADABLE','NO_MATCH','OUT_OF_SCOPE'}

def iso8601(value):
    try:datetime.fromisoformat(value.replace('Z','+00:00'))
    except ValueError as e:raise ValueError('invalid ISO-8601 timestamp: '+value) from e

def load_export(path):
    with Path(path).open(encoding='utf-8',newline='') as f:
        reader=csv.DictReader(f,delimiter='\t');rows=list(reader)
    b.check(tuple(reader.fieldnames or ())==FIELDS,'human export schema mismatch')
    return rows

def validate_rows(rows,reviewer_id='',timestamp=''):
    candidates={r['review_id']:r for r in b.read_tsv(PKG/'LABEL_TOKEN_CANDIDATES.tsv')}
    occurrence={r['occurrence_id']:r for r in b.read_tsv(PKG/'TRANSCRIPTION_TOKEN_CANDIDATES.tsv')}
    snapshot=b.sha(PKG/'LABEL_TOKEN_CANDIDATES.tsv')
    b.check(len(rows)==len(candidates)==92 and len({r['review_id'] for r in rows})==92,'export must contain each of 92 review IDs once')
    b.check({r['review_id'] for r in rows}==set(candidates),'export review ID coverage mismatch')
    verified=[]
    for source in rows:
        r=dict(source);cand=candidates[r['review_id']]
        if reviewer_id:r['reviewer_id']=reviewer_id
        if timestamp:r['decision_timestamp']=timestamp
        b.check(r['label_id']==cand['label_id'] and r['panel']==cand['panel'],'label identity/panel changed')
        b.check(r['candidate_snapshot_sha256']==snapshot,'candidate snapshot mismatch')
        b.check(r['completion_status']=='COMPLETE','incomplete row: '+r['review_id'])
        b.check(r['review_action'] in ACTIONS and r['mapping_outcome'] in OUTCOMES,'invalid action/outcome')
        b.check(r['mapping_confidence'] in CONFIDENCE and r['ambiguity_status'] in AMBIGUITY,'confidence/ambiguity unset')
        b.check(bool(r['reviewer_id']) and bool(r['decision_timestamp']),'reviewer provenance unset');iso8601(r['decision_timestamp'])
        ids=[x for x in r['selected_occurrence_ids'].split(';') if x]
        b.check(len(ids)==len(set(ids)),'duplicate selected occurrence')
        b.check(all(x in occurrence and occurrence[x]['panel']==r['panel'] for x in ids),'unknown/cross-page selected occurrence')
        outcome=r['mapping_outcome'];token_bearing=outcome in {'SINGLE_TOKEN','MULTI_TOKEN','RING_SEQUENCE'}
        if outcome=='SINGLE_TOKEN':b.check(len(ids)==1 and r['review_action'] in {'CONFIRMED_TOKEN','CORRECTED_MAPPING'},'single-token action/count mismatch')
        elif outcome in {'MULTI_TOKEN','RING_SEQUENCE'}:b.check(len(ids)>=2 and r['review_action'] in {'CONFIRMED_SEQUENCE','CORRECTED_MAPPING'},'sequence action/count mismatch')
        elif outcome=='AMBIGUOUS':b.check(r['review_action']=='AMBIGUOUS' and r['ambiguity_status']=='AMBIGUOUS','ambiguous outcome mismatch')
        elif outcome=='UNREADABLE':b.check(r['review_action']=='UNREADABLE' and not ids and r['ambiguity_status']=='UNREADABLE','unreadable outcome mismatch')
        elif outcome=='NO_TRANSCRIPTION_MATCH':b.check(r['review_action']=='NO_MATCH' and not ids and r['ambiguity_status']=='NO_MATCH','no-match outcome mismatch')
        else:b.check(r['review_action']=='NO_MATCH' and not ids and r['ambiguity_status']=='OUT_OF_SCOPE','out-of-scope outcome mismatch')
        if token_bearing:b.check(r['reading_order'] in ORDERS and r['ambiguity_status']=='CLEAR','token-bearing reading-order/ambiguity mismatch')
        elif outcome!='AMBIGUOUS':b.check(r['reading_order'] in {'','UNDETERMINED'},'non-token outcome cannot assert reading order')
        toks=[occurrence[x] for x in ids]
        line_refs=[]
        for x in toks:
            if x['line_ref'] not in line_refs:line_refs.append(x['line_ref'])
        verified.append({'label_id':r['label_id'],'panel':r['panel'],'geometry_type':cand['geometry_type'],
          **{k:cand[k] for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},'group_id':cand['group_id'],'group_size':cand['group_size'],
          'star_count':cand['star_count'],'human_added':cand['human_added'],'mapping_outcome':outcome,
          'transcription_line_refs':';'.join(line_refs) if token_bearing else 'NONE','token_occurrence_ids':';'.join(ids) if token_bearing else 'NONE',
          'raw_token_forms':' '.join(x['readable_eva'] for x in toks) if token_bearing else 'NONE',
          'normalized_token_keys':';'.join(x['canonical_token_key'] for x in toks) if token_bearing else 'NONE',
          'alternative_occurrence_ids':';'.join(ids) if outcome=='AMBIGUOUS' and ids else 'NONE','reading_order':r['reading_order'] or 'UNDETERMINED',
          'mapping_confidence':r['mapping_confidence'],'reviewer_id':r['reviewer_id'],'decision_timestamp':r['decision_timestamp'],
          'ambiguity_status':r['ambiguity_status'],'review_action':r['review_action'],'review_notes':r['notes'],
          'provenance':'HSL-1.0_BLINDED_HUMAN_MAPPING','candidate_snapshot_sha256':snapshot,'final_mapping_status':'HUMAN_VERIFIED'})
    return verified

def freeze(input_path,out,reviewer_id='',timestamp=''):
    rows=load_export(input_path);verified=validate_rows(rows,reviewer_id,timestamp)
    out=Path(out);b.check(not out.exists(),'output directory already exists; use a new freeze directory');out.mkdir(parents=True)
    b.write_tsv(out/'HUMAN_VERIFIED_LABEL_TOKEN_MAPPING.tsv',verified)
    report=f"""# LABEL/token mapping validation

All 92 canonical LABEL IDs have one explicit, schema-valid human outcome. Selected token occurrences, if any, exist on the same panel in the frozen candidate registry; raw and normalized forms were derived by the importer and not typed by the reviewer. Geometry and 3G1 metadata were rejoined from the frozen candidate snapshot.

LABEL_TOKEN_MAPPING_STATUS=FROZEN_COMPLETE
LABELS_VALIDATED=92
HAPAX_ENRICHMENT_RUN_AUTHORIZED=YES
"""
    (out/'LABEL_TOKEN_MAPPING_VALIDATION.md').write_text(report,encoding='utf-8')
    manifest={'version':'HSL-MAPPING-1.0','status':'FROZEN_COMPLETE','labels':92,'review_export_sha256':b.sha(input_path),
      'candidate_snapshot_sha256':b.sha(PKG/'LABEL_TOKEN_CANDIDATES.tsv'),'mapping_sha256':b.sha(out/'HUMAN_VERIFIED_LABEL_TOKEN_MAPPING.tsv'),
      'analysis_plan_sha256':b.sha(PKG/'HAPAX_STAR_LABEL_ANALYSIS_PLAN.md'),'hapax_enrichment_authorized':True}
    b.write_json(out/'LABEL_TOKEN_MAPPING_MANIFEST.json',manifest)
    files=sorted(x for x in out.iterdir() if x.is_file() and x.name!='SHA256SUMS')
    (out/'SHA256SUMS').write_text('\n'.join(f'{b.sha(x)}  {x.name}' for x in files)+'\n',encoding='utf-8')
    print('LABEL_TOKEN_MAPPING_STATUS=FROZEN_COMPLETE; LABELS=92')

def main():
    p=argparse.ArgumentParser();p.add_argument('--input',type=Path,required=True);p.add_argument('--output-dir',type=Path,required=True)
    p.add_argument('--reviewer-id',default='');p.add_argument('--timestamp',default='');args=p.parse_args();freeze(args.input,args.output_dir,args.reviewer_id,args.timestamp)
if __name__=='__main__':main()
