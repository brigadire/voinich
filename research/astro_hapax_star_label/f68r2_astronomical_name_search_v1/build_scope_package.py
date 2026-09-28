#!/usr/bin/env python3
import csv, hashlib, json
from pathlib import Path

ROOT=Path(__file__).resolve().parent
BASE=ROOT.parent.parent/'astro_spatial_human_completeness'
MIG=ROOT.parent/'legacy_mapping_migration_v1'
PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerows(data)

def main():
    groups=[r for r in rows(BASE/'job_21_group_review_validated/REVIEWED_RELATION_GROUPS.tsv') if r['panel']=='f68r2']
    mapping={r['canonical_label_id']:r for r in rows(MIG/'LABEL_3G1_TRANSCRIPTION_MAPPING.tsv') if r['panel']=='f68r2'}
    cross={r['canonical_label_id']:r for r in rows(MIG/'LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv') if r['panel']=='f68r2'}
    target=[r for r in rows(PREP/'TARGET_STAR_LABELS.tsv') if r['page']=='f68r2']
    scope=[]; cw=[]
    for g in groups:
        lid=g['label_ids']; m=mapping.get(lid,{}); c=cross.get(lid,{})
        scope.append({'canonical_group_id':g['canonical_group_id'],'panel':g['panel'],'cvat_group_id':g['cvat_group_id'],'label_id':lid,'star_ids':g['star_ids'],'member_count':g['member_count'],'relation_semantics':g['relation_semantics'],'human_confidence':g['human_confidence'],'crosswalk_outcome':m.get('crosswalk_outcome','MISSING'),'transcription_locus':m.get('transcription_locus',''),'primary_analysis_inclusion':m.get('primary_analysis_inclusion','NO')})
        cw.append({'canonical_group_id':g['canonical_group_id'],'label_id':lid,'group_crosswalk_outcome':m.get('crosswalk_outcome','MISSING'),'legacy_crosswalk_outcome':c.get('crosswalk_outcome','MISSING'),'legacy_record_ids':c.get('legacy_record_ids',''),'transcription_locus':m.get('transcription_locus',''),'absolute_token_positions':m.get('absolute_token_positions',''),'raw_token_sequence':m.get('raw_token_sequence',''),'normalized_token_sequence':m.get('normalized_token_sequence',''),'token_assignment_status':'UNRESOLVED_NO_FROZEN_BRIDGE' if m.get('crosswalk_outcome')!='UNIQUE_PROVENANCE_CROSSWALK' else 'RESOLVED'})
    write('F68R2_GROUP_SCOPE.tsv',list(scope[0]),scope)
    write('LABEL_OCCURRENCE_CROSSWALK.tsv',list(cw[0]),cw)
    graph_fields=['group_id','label_id','token','identity','lexicon_id','normalized_form','rules','encoded','status']
    write('F68R2_PATH_GRAPH.tsv',graph_fields,[])
    write('BASELINE_30_TABLES.tsv',['table_hash','signature','source_package','use_status'],[])
    write('EXISTENCE_TESTS.tsv',['threshold','status','reason'],[{'threshold':x,'status':'NOT_EVALUATED','reason':'implementation gate: 24 human groups lack frozen label-to-EVA crosswalk'} for x in ('coverage>=4','coverage>=5','coverage>=10','coverage>=15','coverage>=19')])
    write('SOLVER_PROGRESS.tsv',['stage','status','detail'],[{'stage':'scope_crosswalk','status':'FAIL','detail':'AMBIGUOUS_CROSSWALK and HUMAN_ADDED_NOT_IN_LEGACY prevent exact token assignment'},{'stage':'f68r2_search','status':'NOT_EVALUATED','detail':'blocked by scope provenance gate'}])
    for n in ('INCUMBENTS.tsv','BEST_TABLES.tsv','MATCH_ASSIGNMENTS.tsv'):
        write(n,['status','reason'],[])
    write('UNMATCHED_GROUPS.tsv',['canonical_group_id','label_id','reason'],[{'canonical_group_id':x['canonical_group_id'],'label_id':x['label_id'],'reason':'NO_UNIQUE_EVA_OCCURRENCE_CROSSWALK'} for x in scope])
    write('RULE_SUPPORT_AUDIT.tsv',['status','reason'],[])
    write('NULL_RESULTS.tsv',['status','reason'],[{'status':'NOT_RUN','reason':'null protocol requires a frozen real-data selection pipeline'}])
    (ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'human_group_source':str(BASE/'job_21_group_review_validated/REVIEWED_RELATION_GROUPS.tsv'),'mapping_source':str(MIG/'LABEL_3G1_TRANSCRIPTION_MAPPING.tsv'),'target_scope_source':str(PREP/'TARGET_STAR_LABELS.tsv'),'f68r2_human_groups':len(groups),'f68r2_target_occurrences':len(target),'resolved_crosswalks':sum(x['token_assignment_status']=='RESOLVED' for x in cw),'unresolved_crosswalks':sum(x['token_assignment_status']!='RESOLVED' for x in cw)},indent=2,sort_keys=True)+'\n')
    status={'F68R2_GROUP_SCOPE_COUNT':len(groups),'F68R2_TARGET_OCCURRENCE_COUNT':len(target),'RESOLVED_GROUP_TOKEN_CROSSWALKS':sum(x['token_assignment_status']=='RESOLVED' for x in cw),'UNRESOLVED_GROUP_TOKEN_CROSSWALKS':sum(x['token_assignment_status']!='RESOLVED' for x in cw),'F68R2_FULL_TABLE_SEARCH':'NOT_EVALUATED','SUPPORT_VALID_SOLUTION_FOUND':'NOT_EVALUATED','NULL_CONTROLS':'NOT_RUN','IMPLEMENTATION_GATE_FAILED':'YES','SEARCH_INCONCLUSIVE_TIMEOUT':'NO','CROSS_PAGE_REQUIREMENT':'WITHDRAWN_AS_PRIMARY_GATE','SCIENTIFIC_CLAIM':'NONE'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
    print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
