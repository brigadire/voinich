#!/usr/bin/env python3
"""Build the blinded Gate-A LABEL-to-transcription human review package."""
from __future__ import annotations

import argparse
import csv
from collections import Counter,defaultdict
import hashlib
import json
from pathlib import Path
import sys

from PIL import Image

sys.dont_write_bytecode=True
ROOT=Path(__file__).resolve().parents[3]
PKG=Path(__file__).resolve().parents[1]
AUG=ROOT/'research/astro_spatial_augmented_human_reference'
OCC=ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl'
PANELS=('f68r1','f68r2','f68r3')
EXPANSIONS={'C':'cth','K':'ckh','P':'cph','F':'cfh','N':'iin','A':'ain','H':'ch','S':'sh','E':'ee','I':'in'}
VERSION='hapax-star-label-preparation-1.0'

INPUTS=[
 ('AUGMENTED_OBJECT_REFERENCE','research/astro_spatial_augmented_human_reference/AUGMENTED_HUMAN_OBJECTS.tsv','augmented-human-reference-1.0','NONE'),
 ('RELATION_GROUPS_3G1','research/astro_spatial_augmented_human_reference/RELATION_GROUPS_3G1.tsv','3G1','NONE'),
 ('RELATION_GROUP_MEMBERS_3G1','research/astro_spatial_augmented_human_reference/RELATION_GROUP_MEMBERS_3G1.tsv','3G1','NONE'),
 ('HUMAN_ADDED_OBJECT_AUDIT','research/astro_spatial_augmented_human_reference/HUMAN_ADDED_OBJECT_AUDIT.tsv','C1.0','NONE'),
 ('AUGMENTED_REFERENCE_MANIFEST','research/astro_spatial_augmented_human_reference/AUGMENTED_REFERENCE_MANIFEST.json','augmented-human-reference-1.0','NONE'),
 ('AUGMENTED_REFERENCE_INPUT_MANIFEST','research/astro_spatial_augmented_human_reference/INPUT_MANIFEST.tsv','augmented-human-reference-1.0','NONE'),
 ('AUGMENTED_REFERENCE_CHECKSUMS','research/astro_spatial_augmented_human_reference/SHA256SUMS','augmented-human-reference-1.0','NONE'),
 ('CANONICAL_IMAGE_F68R1','research/astro_spatial_human_adjudication/images/f68r1.jpg','canonical-image','NONE'),
 ('CANONICAL_IMAGE_F68R2','research/astro_spatial_human_adjudication/images/f68r2.jpg','canonical-image','NONE'),
 ('CANONICAL_IMAGE_F68R3','research/astro_spatial_human_adjudication/images/f68r3.jpg','canonical-image','NONE'),
 ('ZL3B_RAW_IVTFF','data/ZL3b-n.txt','ZL3b','ZL3b/EVA'),
 ('ZL3B_X7_EXACT','data_work/ZL3b-x7.txt','ZL3b-x7','ZL3b/EVA'),
 ('ZL3B_X7_CANONICAL','data_work/ZL3b-x7.canonical.txt','ZL3b-x7-canonical','ZL3b/EVA composite atoms'),
 ('ZL3B_OCCURRENCE_METADATA','experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl','Fingerprint-V2.1-compatible','ZL3b/EVA composite atoms'),
 ('ZL3B_HISTORICAL_FREEZE_MANIFEST','experiments/fingerprint-v2-task79-v1/canonical-out/freeze_manifest.json','Fingerprint-V2 historical','ZL3b/EVA composite atoms'),
 ('ZL3B_DETERMINISTIC_REFREEZE_MANIFEST','research/phase2/task83b/TASK83B_RESULTS_MANIFEST.json','Fingerprint-V2.1','ZL3b/EVA composite atoms'),
 ('ZL3B_SOURCE_PROVENANCE','research/phase2/task83b/SOURCE_PROVENANCE.tsv','Fingerprint-V2.1','ZL3b/EVA'),
 ('ZL3B_MULTIRUN_REPRODUCIBILITY','research/phase2/task83b/MULTIRUN_REPRODUCIBILITY.tsv','Fingerprint-V2.1','ZL3b/EVA'),
 ('CORPUS_NORMALIZATION_IMPLEMENTATION','internal/corpusprep/corpusprep.go','Fingerprint-V2.1','ZL3b/EVA'),
 ('PAGE_SECTION_METADATA','research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv','visual-context','NONE'),
 ('PAGE_LINE_METADATA','research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv','visual-context','ZL3b'),
 ('PREVIOUS_HAPAX_REPORT','research/hapax_check/HAPAX_SPATIAL_AND_PAGE_REPORT.md','hapax-check-1','ZL3b canonical'),
 ('PREVIOUS_PAGE_HAPAX','research/hapax_check/PAGE_HAPAX_PREVALENCE.tsv','hapax-check-1','ZL3b canonical'),
 ('PREVIOUS_SECTION_HAPAX','research/hapax_check/SECTION_HAPAX_PREVALENCE.tsv','hapax-check-1','ZL3b canonical'),
 ('PREVIOUS_HAPAX_MANIFEST','research/hapax_check/HAPAX_RESULTS_MANIFEST.json','hapax-check-1','ZL3b canonical'),
 ('PREVIOUS_SPATIAL_BRIDGE','research/astro_spatial_annotation/ASTRO_ZL3B_SPATIAL_BRIDGE.tsv','astro-spatial-1','ZL3b/EVA'),
 ('PREVIOUS_SPATIAL_MANIFEST','research/astro_spatial_annotation/manifest.json','astro-spatial-1','ZL3b/EVA'),
 ('PREVIOUS_SPATIAL_CHECKSUMS','research/astro_spatial_annotation/SHA256SUMS','astro-spatial-1','NONE'),
 ('STOLFI_LABEL_MATCHES','research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv','stolfi-label-inventory','ZL3b/EVA'),
 ('STOLFI_LABEL_AUDIT','research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_AUDIT.md','stolfi-label-inventory','ZL3b/EVA'),
 ('PREVIOUS_LEXICON_EXPERIMENT','research/astro_token_formation/TOKEN_FORMATION_BRUTEFORCE_REPORT.md','token-formation-D0','Voynich/Latin/Arabic'),
 ('PREVIOUS_LEXICON_MANIFEST','research/astro_token_formation/TOKEN_FORMATION_MANIFEST.json','token-formation-D0','Voynich/Latin/Arabic'),
 ('PREVIOUS_LEXICON_CHECKSUMS','research/astro_token_formation/TOKEN_FORMATION_SHA256SUMS','token-formation-D0','NONE'),
 ('PREVIOUS_EXPANDED_LEXICON_REPORT','research/astro_dictionary_expansion_m1/D1_TOKEN_FORMATION_REPORT.md','dictionary-expansion-D1','Voynich/Latin/Arabic'),
 ('PREVIOUS_EXPANDED_LEXICON_AUDIT','research/astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_AUDIT.md','dictionary-expansion-D1','Latin/Arabic/Latinized Arabic'),
 ('PREVIOUS_EXPANDED_LEXICON_MANIFEST','research/astro_dictionary_expansion_m1/manifest.json','dictionary-expansion-D1','Voynich/Latin/Arabic'),
 ('PREVIOUS_EXPANDED_LEXICON_CHECKSUMS','research/astro_dictionary_expansion_m1/SHA256SUMS','dictionary-expansion-D1','NONE'),
]

PLAN="""# Hapax–STAR/LABEL analysis plan (preregistered)

Version: `HSL-1.0`. This plan is frozen before any human-verified LABEL/token cohort, enrichment statistic, or new lexicon match is produced.

## Primary hypothesis and units

The directional primary hypothesis is `P(hapax | grouped LABEL token occurrence) > P(hapax | ungrouped LABEL token occurrence)` on f68r1, f68r2 and f68r3. The primary analytical unit is a verified transcription token occurrence, with LABEL as the dependence block and panel as the randomization/cluster stratum. A multi-token LABEL contributes its verified occurrences but is never treated as independent tokens for resampling. The eight-member 3G1 group contributes one LABEL/group observation, never seven pairs.

## Hapax and denominator

Primary `HAPAX` means an exact frozen normalized token key with frequency exactly one in the complete ZL3b occurrence corpus, not within these pages or within the Astronomical section. The authoritative prepared corpus is `data_work/ZL3b-x7.canonical.txt` (SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`); occurrence identities and metadata come from the byte-registered ZL3b occurrence JSONL. Metadata and comments are not tokens. Upstream token boundaries are retained. Composite EVA sequences (`cth/ckh/cph/cfh/iin/ain/ch/sh/ee/in`) remain frozen atomic keys; readable expansion is display-only. Apostrophe, `?`, `@NNN;`, digits and other retained literal transcription markers are not silently removed. Upstream selection of the first explicit IVTFF alternative is not reopened.

## Mapping and uncertainty

Every one of the 92 LABEL objects must receive exactly one explicit human outcome: `SINGLE_TOKEN`, `MULTI_TOKEN`, `RING_SEQUENCE`, `AMBIGUOUS`, `UNREADABLE`, `NO_TRANSCRIPTION_MATCH`, or `OUT_OF_TRANSCRIPTION_SCOPE`. Empty is incomplete, never NO_MATCH. Primary analysis includes only verified SINGLE_TOKEN/MULTI_TOKEN/RING_SEQUENCE mappings whose selected occurrence IDs exist byte-for-byte in the frozen registry; ambiguous, unreadable and unmatched records are excluded. Ring start and order may be `UNDETERMINED`; no order is invented. Human UI is blind to group membership, origin, hapax/frequency, prior lexicon results and per-token hypothesis metadata.

## Primary test and effect sizes

Report exact numerators/denominators, token-occurrence hapax risk difference, risk ratio and odds ratio with explicit zero-cell handling. Use a one-sided panel-stratified permutation that reallocates grouped status at LABEL-block level while preserving within-panel grouped counts and token bundles; enumerate when feasible, otherwise 100,000 deterministic draws with seed `20260915`. Report a two-sided 95% label-block bootstrap CI stratified by panel (100,000 replicates, seed `20260916`). Because there are only three panels, panel-aware inference is primary and uncertainty is described as weak-cluster pilot inference.

## Label- and group-level analyses

Separately report `LABEL_CONTAINS_HAPAX`, controlling descriptively for token count, token length, confidence and panel; use label-block stratified permutation. Each 3G1 group is one row with its single LABEL, group size, STAR count, initial/new/extended category and human-added membership. No Cartesian expansion is permitted.

## Secondary family and multiplicity

Secondary analyses are: panel-specific effects; new versus unchanged groups; human-added LABEL; the one eight-member group descriptively; thresholds frequency <=2 and <=3; token length; initial/final glyph sequences; and concentration by panel. Apply Benjamini–Hochberg at q=0.05 across this secondary family. The primary test is a single preregistered family and is not combined with secondary p-values. Controls requiring unavailable verified spatial mappings are reported unavailable, not synthesized.

## Sensitivity policies

Run (1) exact readable EVA form, (2) frozen composite normalized key, and (3) exclusion of tokens containing uncertain/unreadable markers or nonalphabetic atoms. An alternative normalization is allowed only if already registered here or in a frozen upstream specification; none may be invented after results. Compare grouped LABEL against same-page ungrouped LABEL primarily, then eligible astronomical and length-matched other-section controls only when reproducibly constructible.

## Lexicon gate and tiers

No lexicon stage runs before complete frozen human mapping and nonempty grouped-hapax subset. Matching rules must be frozen first: Tier 0 exact minimal normalization; Tier 1 only source-documented orthographic variants; Tier 2 only preregistered historically attested Arabic-to-Latin transliterations; Tier 3 constrained structural similarity with frozen threshold, length controls, pseudo-token controls and FDR. If primary enrichment is not supported, any permitted search is `EXPLORATORY`; otherwise it may be `HYPOTHESIS_FOLLOWUP`. Single similarities never establish etymology, a named star, or decipherment.

## Blocking and stopping

Stop before enrichment if any frozen checksum fails, fewer or more than 92 LABEL outcomes exist, a canonical ID/geometry/group membership changes, selected occurrences do not exist, UI blindness fails, or mapping is incomplete. Stop before lexicon matching if provenance or matching-plan freeze is absent. Do not alter hypotheses, hapax definition, thresholds, controls, or normalization after mapping/results inspection.
"""

GOALS="""# Hapax–STAR/LABEL goals

Test whether tokens in human-verified visual LABEL groups are enriched for corpus-wide hapax forms relative to same-page ungrouped LABEL, without treating visual association as astronomical semantics. First obtain a complete blinded human mapping from 92 physical LABEL objects to frozen ZL3b transcription occurrences. Only a later gated stage may calculate enrichment or revisit historical lexicons.
"""

TRANSCRIPTION_POLICY="""# Transcription policy

The project uses ZL3b/EVA. The primary corpus file is the deterministic-refreeze-supported `data_work/ZL3b-x7.canonical.txt`, SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`. Token occurrence IDs, folio, locus/line and positions come from the registered occurrence metadata whose three target pages are reproduced byte-identically in the Task83b multirun audit.

Candidate display expands frozen composite atoms through the already implemented bijection `C→cth`, `K→ckh`, `P→cph`, `F→cfh`, `N→iin`, `A→ain`, `H→ch`, `S→sh`, `E→ee`, `I→in`. This expansion is for readability; `canonical_token_key` and absolute occurrence ID determine identity. Case is preserved in the key. Separators are upstream token boundaries. Comments/metadata are absent from occurrences. Literal uncertain/unreadable markers remain literal. Alternative readings are not adjudicated automatically; the upstream first-reading convention remains fixed, while human ambiguity may be recorded without inventing a form.

No page-local frequency, hapax status or dictionary information enters candidate construction or the review UI.
"""

HAPAX_DEFINITION="""# Frozen hapax definition

Primary: `HAPAX = YES` iff a frozen composite normalized token key occurs exactly once across every occurrence in the registered complete ZL3b corpus. This is a token-type frequency evaluated corpus-wide; the 92 mappings only select occurrences and never define frequency. Metadata, comments and missing records do not enter the occurrence denominator. Literal markers retained by upstream preprocessing remain part of the exact key.

Sensitivity analyses are preregistered for readable expanded EVA exact forms and for exclusion of keys containing `?`, `@`, digits or nonalphabetic atoms. Frequency <=2 and <=3 are rare-token sensitivities, not alternative hapax definitions. No frequency or hapax field is included in the blind review assets.
"""

SCHEMA="""# LABEL/token mapping schema

`LABEL_TOKEN_CANDIDATES.tsv` contains one row per physical LABEL and is not a verified mapping. `proposed_*` fields are deliberately `NONE`: no frozen 2D-to-transcription alignment exists. `alternative_line_refs` is the exhaustive same-page candidate line set. Internal group and provenance fields are retained outside the blinded UI for later cohort construction.

Human export fields:

- `review_id`, `label_id`, `panel`, `candidate_snapshot_sha256`: immutable identifiers;
- `review_action`: `CONFIRMED_TOKEN`, `CONFIRMED_SEQUENCE`, `CORRECTED_MAPPING`, `AMBIGUOUS`, `UNREADABLE`, or `NO_MATCH`;
- `mapping_outcome`: `SINGLE_TOKEN`, `MULTI_TOKEN`, `RING_SEQUENCE`, `AMBIGUOUS`, `UNREADABLE`, `NO_TRANSCRIPTION_MATCH`, or `OUT_OF_TRANSCRIPTION_SCOPE`;
- `selected_occurrence_ids`: semicolon-separated absolute occurrence IDs in intended order;
- `reading_order`: `CORPUS_ORDER`, `CLOCKWISE`, `COUNTERCLOCKWISE`, `VISUAL_LEFT_TO_RIGHT`, or `UNDETERMINED`;
- confidence, reviewer, ISO-8601 timestamp, ambiguity and notes;
- `completion_status=COMPLETE` only after an explicit outcome. Blank fields never mean NO_MATCH.

For token-bearing outcomes every selected occurrence must exist on the LABEL panel. SINGLE_TOKEN requires one occurrence; MULTI_TOKEN and RING_SEQUENCE require at least two. Ambiguous/unreadable/no-match/out-of-scope outcomes cannot carry asserted token forms. The importer derives exact raw/normalized forms from occurrence IDs; reviewers cannot type a novel token into the frozen mapping.
"""

GUIDE="""# Blind physical-LABEL transcription review guide

Review each of 92 physical text regions. The interface shows a canonical crop, its geometry on the full page, and every frozen transcription occurrence on that page. It intentionally does not show research categories or token statistics.

1. Inspect the crop and page context.
2. Select the exact occurrence(s) that transcribe the marked physical region. Select only what lies within the region; a physical run may contain several tokens.
3. Choose SINGLE_TOKEN, MULTI_TOKEN, or RING_SEQUENCE as appropriate. Do not choose a circular starting point or direction unless visible/transcription evidence supports it; otherwise set reading order UNDETERMINED.
4. If more than one mapping remains plausible, choose AMBIGUOUS and optionally retain alternatives. If the marks cannot be read, choose UNREADABLE. If the region is legible but absent from the candidate transcription, explicitly choose NO_TRANSCRIPTION_MATCH. Use OUT_OF_TRANSCRIPTION_SCOPE only when the physical region is not represented by the registered transcription scope.
5. Set confidence, reviewer ID and decision timestamp. Save in the browser and export TSV only when all rows are complete.

Do not infer meaning, astronomical identity or spelling from nearby drawings. Do not use external dictionaries. Empty is incomplete and must not be used as NO_MATCH.
"""

HTML=r'''<!doctype html><meta charset="utf-8"><title>Physical LABEL transcription review</title>
<style>body{font:14px sans-serif;margin:0;background:#eee}header{position:sticky;top:0;background:#17324d;color:white;padding:10px;z-index:2}main{max-width:1250px;margin:auto;background:white;padding:16px}.cols{display:grid;grid-template-columns:1fr 1fr;gap:14px}img,canvas{max-width:100%;border:1px solid #888}.tokens{max-height:330px;overflow:auto;border:1px solid #aaa;padding:8px}.line{margin:8px 0}.tok{display:inline-block;margin:2px;padding:3px 5px;border:1px solid #bbb;border-radius:4px}label{display:block;margin:7px 0}input,select,textarea{font:inherit;padding:4px}textarea{width:95%}.nav button{margin-right:6px}.bad{color:#b00}.ok{color:#075}</style>
<header><span id="progress"></span> <span class="nav"><button onclick="move(-1)">Previous</button><button onclick="move(1)">Next</button><button onclick="save()">Save</button><button onclick="exportTSV()">Export TSV</button></span></header>
<main><h2 id="title"></h2><div class="cols"><div><h3>Region crop</h3><img id="crop"><h3>Page context</h3><canvas id="page"></canvas></div><div><h3>Candidate transcription occurrences on this page</h3><div class="tokens" id="tokens"></div><label>Review action <select id="action"><option></option><option>CONFIRMED_TOKEN</option><option>CONFIRMED_SEQUENCE</option><option>CORRECTED_MAPPING</option><option>AMBIGUOUS</option><option>UNREADABLE</option><option>NO_MATCH</option></select></label><label>Mapping outcome <select id="outcome"><option></option><option>SINGLE_TOKEN</option><option>MULTI_TOKEN</option><option>RING_SEQUENCE</option><option>AMBIGUOUS</option><option>UNREADABLE</option><option>NO_TRANSCRIPTION_MATCH</option><option>OUT_OF_TRANSCRIPTION_SCOPE</option></select></label><label>Reading order <select id="order"><option></option><option>CORPUS_ORDER</option><option>CLOCKWISE</option><option>COUNTERCLOCKWISE</option><option>VISUAL_LEFT_TO_RIGHT</option><option>UNDETERMINED</option></select></label><label>Confidence <select id="confidence"><option></option><option>HIGH</option><option>MEDIUM</option><option>LOW</option></select></label><label>Reviewer ID <input id="reviewer"></label><label>Decision timestamp (ISO-8601) <input id="timestamp" placeholder="2026-09-15T12:00:00Z"></label><label>Ambiguity status <select id="ambiguity"><option></option><option>CLEAR</option><option>AMBIGUOUS</option><option>UNREADABLE</option><option>NO_MATCH</option><option>OUT_OF_SCOPE</option></select></label><label>Notes <textarea id="notes"></textarea></label><p id="status"></p></div></div></main>
<script src="review_data.js"></script><script>
let i=0,storageKey='labelTokenReview:'+REVIEW_DATA.snapshot,state=JSON.parse(localStorage.getItem(storageKey)||'{}'); const $=x=>document.getElementById(x);
function current(){return REVIEW_DATA.labels[i]} function valid(s){let n=(s.selected||[]).length;if(!s.action||!s.outcome||!s.confidence||!s.reviewer||!s.timestamp||!s.ambiguity)return false;if(s.outcome==='SINGLE_TOKEN')return n===1&&['CONFIRMED_TOKEN','CORRECTED_MAPPING'].includes(s.action)&&s.ambiguity==='CLEAR'&&!!s.order;if(['MULTI_TOKEN','RING_SEQUENCE'].includes(s.outcome))return n>=2&&['CONFIRMED_SEQUENCE','CORRECTED_MAPPING'].includes(s.action)&&s.ambiguity==='CLEAR'&&!!s.order;if(s.outcome==='AMBIGUOUS')return s.action==='AMBIGUOUS'&&s.ambiguity==='AMBIGUOUS';if(s.outcome==='UNREADABLE')return n===0&&s.action==='UNREADABLE'&&s.ambiguity==='UNREADABLE';if(s.outcome==='NO_TRANSCRIPTION_MATCH')return n===0&&s.action==='NO_MATCH'&&s.ambiguity==='NO_MATCH';return n===0&&s.outcome==='OUT_OF_TRANSCRIPTION_SCOPE'&&s.action==='NO_MATCH'&&s.ambiguity==='OUT_OF_SCOPE'}
function save(){let x=current(),s=state[x.review_id]||{};s.selected=[...document.querySelectorAll('.tok input:checked')].map(x=>x.value);for(let k of ['action','outcome','order','confidence','reviewer','timestamp','ambiguity','notes'])s[k]=$(k).value;state[x.review_id]=s;localStorage.setItem(storageKey,JSON.stringify(state));renderProgress();$('status').textContent=valid(s)?'Saved; row complete':'Saved; required fields remain';$('status').className=valid(s)?'ok':'bad'}
function move(d){save();i=Math.max(0,Math.min(REVIEW_DATA.labels.length-1,i+d));render()}
function renderProgress(){let n=REVIEW_DATA.labels.filter(x=>valid(state[x.review_id]||{})).length;$('progress').textContent=`${i+1}/${REVIEW_DATA.labels.length}; complete ${n}/${REVIEW_DATA.labels.length}`}
function render(){let x=current(),s=state[x.review_id]||{};$('title').textContent=x.label_id+' — '+x.panel;$('crop').src=x.crop;let box=x.geometry,im=new Image;im.onload=()=>{let c=$('page'),displayScale=Math.min(1,1100/im.width);c.width=im.width*displayScale;c.height=im.height*displayScale;let sx=c.width/x.page_size[0],sy=c.height/x.page_size[1],q=c.getContext('2d');q.drawImage(im,0,0,c.width,c.height);q.strokeStyle='#e6007e';q.lineWidth=4;q.strokeRect(box[0]*sx,box[1]*sy,(box[2]-box[0])*sx,(box[3]-box[1])*sy)};im.src=x.page_image;let lines=REVIEW_DATA.lines[x.panel],h='';for(let line of lines){h+=`<div class="line"><b>${line.line_ref}</b> [${line.locus_type}]<br>`;for(let t of line.tokens)h+=`<label class="tok"><input type="checkbox" value="${t.occurrence_id}" ${(s.selected||[]).includes(t.occurrence_id)?'checked':''}>${t.eva} <small>#${t.occurrence_id}</small></label>`;h+='</div>'}$('tokens').innerHTML=h;for(let k of ['action','outcome','order','confidence','reviewer','timestamp','ambiguity','notes'])$(k).value=s[k]||'';$('status').textContent='';renderProgress()}
function esc(v){return '"'+String(v??'').replaceAll('"','""')+'"'} function exportTSV(){save();let f=['review_id','label_id','panel','candidate_snapshot_sha256','review_action','mapping_outcome','selected_occurrence_ids','reading_order','mapping_confidence','reviewer_id','decision_timestamp','ambiguity_status','notes','completion_status'],out=[f.join('\t')];for(let x of REVIEW_DATA.labels){let s=state[x.review_id]||{},r=[x.review_id,x.label_id,x.panel,REVIEW_DATA.snapshot,s.action||'',s.outcome||'',(s.selected||[]).join(';'),s.order||'',s.confidence||'',s.reviewer||'',s.timestamp||'',s.ambiguity||'',s.notes||'',valid(s)?'COMPLETE':'INCOMPLETE'];out.push(r.map(esc).join('\t'))}let b=new Blob([out.join('\n')+'\n'],{type:'text/tab-separated-values'}),a=document.createElement('a');a.href=URL.createObjectURL(b);a.download='HUMAN_LABEL_TOKEN_MAPPING_EXPORT.tsv';a.click();URL.revokeObjectURL(a.href)}render();
</script>'''

def check(ok,msg):
    if not ok:raise ValueError(msg)
def sha(path):
    h=hashlib.sha256()
    with Path(path).open('rb') as f:
        for b in iter(lambda:f.read(1<<20),b''):h.update(b)
    return h.hexdigest()
def read_tsv(path):
    with Path(path).open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write_tsv(path,rows,fields=None):
    path=Path(path);path.parent.mkdir(parents=True,exist_ok=True)
    if fields is None:fields=list(rows[0])
    with path.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fields,delimiter='\t',lineterminator='\n',extrasaction='raise');w.writeheader();w.writerows(rows)
def write_json(path,obj):Path(path).write_text(json.dumps(obj,sort_keys=True,indent=2,ensure_ascii=False)+'\n',encoding='utf-8')
def verify_ledger(path):
    path=Path(path)
    for line in path.read_text().splitlines():
        if not line.strip():continue
        expected,rel=line.split('  ',1);local=(path.parent/rel).resolve();rooted=(ROOT/rel).resolve()
        target=local if local.is_file() else rooted
        check(target.is_relative_to(ROOT.resolve()) and target.is_file(),f'unsafe/missing ledger target {rel}')
        check(sha(target)==expected,f'checksum mismatch {target}')
def row_count(path):
    if path.suffix=='.tsv':return len(read_tsv(path))
    if path.suffix=='.jsonl':
        with path.open() as f:return sum(1 for line in f if line.strip())
    if path.suffix=='.txt':
        with path.open(errors='replace') as f:return sum(1 for _ in f)
    return ''
def expand(key):return ''.join(EXPANSIONS.get(x,x) for x in key.split('\x1f'))

def verify_inputs(manifest=None):
    for ledger in (AUG/'SHA256SUMS',ROOT/'research/astro_spatial_annotation/SHA256SUMS',ROOT/'research/astro_token_formation/TOKEN_FORMATION_SHA256SUMS',ROOT/'research/astro_dictionary_expansion_m1/SHA256SUMS'):
        verify_ledger(ledger)
    refreeze=json.loads((ROOT/'research/phase2/task83b/TASK83B_RESULTS_MANIFEST.json').read_text())
    check(refreeze['status']=='AUTHORITATIVE' and refreeze['version']=='FINGERPRINT_V2.1_DETERMINISTIC_SCIENTIFIC_REFREEZE','ZL3b refreeze not authoritative')
    for rel,d in refreeze['output_artifact_checksums'].items():check(sha(ROOT/'research/phase2/task83b'/rel)==d,'Task83b manifest mismatch '+rel)
    multirun=read_tsv(ROOT/'research/phase2/task83b/MULTIRUN_REPRODUCIBILITY.tsv')
    zcor=next(r for r in multirun if r['artifact']=='reconstruction_input/data_work/ZL3b-x7.canonical.txt')
    zocc=next(r for r in multirun if r['artifact']=='zl/occurrence_metadata.jsonl')
    check(zcor['byte_identical']=='true' and zcor['sha256_run_a']==sha(ROOT/'data_work/ZL3b-x7.canonical.txt'),'ZL3b corpus refreeze binding failed')
    check(zocc['byte_identical']=='true' and zocc['sha256_run_a']==sha(OCC),'occurrence metadata refreeze binding failed')
    result=[]
    for role,rel,version,system in INPUTS:
        p=ROOT/rel;check(p.is_file(),'missing input '+rel)
        result.append({'logical_role':role,'path':rel,'version':version,'transcription_system':system,'row_count':row_count(p),'file_size':p.stat().st_size,'sha256':sha(p),'frozen_status':'FROZEN_VERIFIED'})
    if manifest:
        registered=read_tsv(manifest)
        for r in result:r['row_count']=str(r['row_count']);r['file_size']=str(r['file_size'])
        check(result==registered,'input manifest changed')
    return result

def occurrence_registry():
    all_rows=[]
    for line in OCC.open(encoding='utf-8'):
        o=json.loads(line)
        if o['folio'] in PANELS:
            key=o['token'];oid=str(o['absolute_token_position'])
            all_rows.append({'occurrence_id':oid,'panel':o['folio'],'line_ref':o['line_identifier'].split('\0')[-1],
                'locus_ref':o['locus_identifier'].split('\0')[-1],'locus_type':o['locus_type'],'index_in_line':o['index_in_line'],
                'index_in_locus':o['index_in_locus'],'canonical_token_key':'/'.join(key.split('\x1f')),
                'readable_eva':expand(key),'label_text_status':o['label_text_status'],'missing_status':o['missing_status'] or 'NONE',
                'source_path':'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl'})
    check(Counter(r['panel'] for r in all_rows)=={'f68r1':69,'f68r2':89,'f68r3':110},'target-page occurrence counts changed')
    return all_rows

def build(out=PKG):
    out=Path(out);out.mkdir(parents=True,exist_ok=True)
    for d in ('human_review/assets/crops','human_review/assets/pages'): (out/d).mkdir(parents=True,exist_ok=True)
    (out/'HAPAX_STAR_LABEL_GOALS.md').write_text(GOALS,encoding='utf-8')
    (out/'HAPAX_STAR_LABEL_ANALYSIS_PLAN.md').write_text(PLAN,encoding='utf-8')
    plan_sha=sha(out/'HAPAX_STAR_LABEL_ANALYSIS_PLAN.md')
    write_json(out/'ANALYSIS_PLAN_FREEZE.json',{'version':'HSL-1.0','sha256':plan_sha,'frozen_before_verified_mapping':True,'frozen_before_enrichment':True,'frozen_before_lexicon_matching':True})
    (out/'TRANSCRIPTION_POLICY.md').write_text(TRANSCRIPTION_POLICY,encoding='utf-8')
    (out/'HAPAX_DEFINITION.md').write_text(HAPAX_DEFINITION,encoding='utf-8')
    (out/'LABEL_TOKEN_MAPPING_SCHEMA.md').write_text(SCHEMA,encoding='utf-8')
    (out/'HUMAN_LABEL_TOKEN_MAPPING_GUIDE.md').write_text(GUIDE,encoding='utf-8')
    inputs=verify_inputs();write_tsv(out/'INPUT_MANIFEST.tsv',inputs,['logical_role','path','version','transcription_system','row_count','file_size','sha256','frozen_status'])

    objects=[r for r in read_tsv(AUG/'AUGMENTED_HUMAN_OBJECTS.tsv') if r['panel'] in PANELS and r['object_type']=='LABEL']
    check(len(objects)==92 and Counter(r['panel'] for r in objects)=={'f68r1':37,'f68r2':33,'f68r3':22},'relation LABEL set changed')
    group_rows=read_tsv(AUG/'RELATION_GROUPS_3G1.tsv');members=read_tsv(AUG/'RELATION_GROUP_MEMBERS_3G1.tsv')
    group={r['canonical_group_id']:r for r in group_rows};group_of={r['object_id']:r['canonical_group_id'] for r in members if r['object_type']=='LABEL'}
    check(len(group_of)==64,'grouped LABEL count changed')
    occ=occurrence_registry();write_tsv(out/'TRANSCRIPTION_TOKEN_CANDIDATES.tsv',occ)
    by_panel=defaultdict(list)
    for r in occ:by_panel[r['panel']].append(r)
    line_rows=[];ui_lines={}
    for panel in PANELS:
        lines=defaultdict(list)
        for r in by_panel[panel]:lines[r['line_ref']].append(r)
        ui_lines[panel]=[]
        for line_ref,tokens in lines.items():
            token_ids=';'.join(r['occurrence_id'] for r in tokens);raw=' '.join(r['readable_eva'] for r in tokens)
            line_rows.append({'panel':panel,'line_ref':line_ref,'locus_type':tokens[0]['locus_type'],'token_count':len(tokens),'occurrence_ids':token_ids,'readable_eva_sequence':raw,'candidate_scope':'SAME_PAGE_EXHAUSTIVE'})
            ui_lines[panel].append({'line_ref':line_ref,'locus_type':tokens[0]['locus_type'],'tokens':[{'occurrence_id':r['occurrence_id'],'eva':r['readable_eva']} for r in tokens]})
    write_tsv(out/'TRANSCRIPTION_LINE_CANDIDATES.tsv',line_rows)

    candidates=[];ui_labels=[]
    for r in sorted(objects,key=lambda x:(x['panel'],x['object_id'])):
        gid=group_of.get(r['object_id']);g=group.get(gid,{})
        rid='MAP_'+hashlib.sha256((VERSION+'|'+r['object_id']).encode()).hexdigest()[:16].upper()
        bbox=[float(r[k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
        image=Image.open(AUG.parent/'astro_spatial_human_adjudication/images'/f"{r['panel']}.jpg").convert('RGB')
        pad=max(50,int(max(bbox[2]-bbox[0],bbox[3]-bbox[1])*.18));crop_box=(max(0,int(bbox[0])-pad),max(0,int(bbox[1])-pad),min(image.width,int(bbox[2])+pad),min(image.height,int(bbox[3])+pad))
        crop=image.crop(crop_box);crop.thumbnail((1400,650),Image.Resampling.LANCZOS);crop_rel=f'human_review/assets/crops/{rid}.jpg';crop.save(out/crop_rel,quality=90,optimize=False,progressive=False)
        line_refs=';'.join(x['line_ref'] for x in line_rows if x['panel']==r['panel'])
        candidates.append({'review_id':rid,'label_id':r['object_id'],'panel':r['panel'],'geometry_type':r['geometry_type'],
            **{k:r[k] for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},
            'group_id':gid or 'UNGROUPED','group_size':g.get('member_count','0'),'star_count':g.get('star_count','0'),
            'human_added':'YES' if r['origin']=='HUMAN_ADDED_COMPLETENESS' else 'NO','proposed_transcription_line':'NONE',
            'proposed_token_position':'NONE','proposed_raw_token_form':'NONE','proposed_normalized_token_form':'NONE',
            'mapping_method':'PAGE_LEVEL_EXHAUSTIVE_CANDIDATE_SET_NO_2D_TEXT_ALIGNMENT','mapping_confidence':'LOW',
            'ambiguity_status':'AMBIGUOUS_REQUIRES_HUMAN','alternative_line_refs':line_refs,'candidate_line_count':len(line_refs.split(';')),
            'visual_crop_reference':crop_rel,'candidate_mapping_outcome':'AMBIGUOUS'})
        ui_labels.append({'review_id':rid,'label_id':r['object_id'],'panel':r['panel'],'geometry':bbox,'rotation':r['rotation'],'page_size':[image.width,image.height],
            'crop':'assets/crops/'+rid+'.jpg','page_image':'assets/pages/'+r['panel']+'.jpg'})
    write_tsv(out/'LABEL_TOKEN_CANDIDATES.tsv',candidates)
    snapshot=sha(out/'LABEL_TOKEN_CANDIDATES.tsv')
    export_fields=['review_id','label_id','panel','candidate_snapshot_sha256','review_action','mapping_outcome','selected_occurrence_ids','reading_order','mapping_confidence','reviewer_id','decision_timestamp','ambiguity_status','notes','completion_status']
    template=[{'review_id':r['review_id'],'label_id':r['label_id'],'panel':r['panel'],'candidate_snapshot_sha256':snapshot,
        'review_action':'','mapping_outcome':'','selected_occurrence_ids':'','reading_order':'','mapping_confidence':'','reviewer_id':'','decision_timestamp':'','ambiguity_status':'','notes':'','completion_status':'INCOMPLETE'} for r in candidates]
    write_tsv(out/'HUMAN_REVIEW_EXPORT_TEMPLATE.tsv',template,export_fields)
    for panel in PANELS:
        im=Image.open(AUG.parent/'astro_spatial_human_adjudication/images'/f'{panel}.jpg').convert('RGB');im.thumbnail((1600,1600),Image.Resampling.LANCZOS);im.save(out/f'human_review/assets/pages/{panel}.jpg',quality=88,optimize=False,progressive=False)
    (out/'human_review/index.html').write_text(HTML,encoding='utf-8')
    js='const REVIEW_DATA='+json.dumps({'snapshot':snapshot,'labels':ui_labels,'lines':ui_lines},sort_keys=True,ensure_ascii=False,separators=(',',':'))+';\n'
    (out/'human_review/review_data.js').write_text(js,encoding='utf-8')

    report=f"""# Preparation report

Gate A preparation is complete. No frozen human-verified mapping was found: the previous spatial bridge marks token/geometry links `UNMAPPED`, and no artifact binds the 92 current canonical LABEL IDs to transcription occurrences. Therefore enrichment and lexicon stages were not run.

The candidate table covers 92 LABEL exactly: 37 f68r1, 33 f68r2 and 22 f68r3; 64 are grouped and 28 ungrouped internally. The blind browser package contains 92 crops, three page-context images, and the exhaustive same-page registry of 268 frozen ZL3b occurrences in 90 lines. It makes no automatic token choice: every row is `AMBIGUOUS_REQUIRES_HUMAN` with explicit `NONE` proposed fields. Group/origin and all frequency/hapax fields are absent from UI assets.

The analysis plan was frozen as SHA-256 `{plan_sha}` before candidate output. The primary ZL3b canonical corpus checksum is confirmed against the authoritative deterministic Task83b refreeze. Prior hapax and Latin/Arabic experiments are registered as historical inputs only and were not reused as answers.

```text
HAPAX_STAR_LABEL_PREPARATION=COMPLETE
LABELS_IN_RELATION_SCOPE=92
LABEL_TOKEN_MAPPING_STATUS=READY_FOR_HUMAN_VERIFICATION
HAPAX_DEFINITION_FROZEN=YES
ANALYSIS_PLAN_FROZEN=YES
HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO
LEXICON_MATCH_RUN_AUTHORIZED=NO
FROZEN_INPUTS_UNCHANGED=YES
RESULTS_REPRODUCIBLE=YES
```
"""
    (out/'PREPARATION_REPORT.md').write_text(report,encoding='utf-8')
    repro="""# Reproducibility

From this directory:

```bash
python3 -B scripts/build_preparation.py
python3 -B scripts/validate_preparation.py
python3 -B -m unittest discover -s tests -v
sha256sum -c --quiet SHA256SUMS
```

Open `human_review/index.html` locally, complete all 92 rows and export TSV. Validate/import later with `python3 -B scripts/import_human_mapping.py --input EXPORT.tsv --reviewer-id ID --timestamp ISO8601 --output-dir NEW_DIRECTORY`. The importer refuses incomplete or invented-token mappings. Validation rebuilds generated preparation files in a temporary directory and compares them byte-for-byte.
"""
    (out/'REPRODUCIBILITY.md').write_text(repro,encoding='utf-8')
    write_json(out/'PREPARATION_MANIFEST.json',{'version':VERSION,'status':'READY_FOR_HUMAN_VERIFICATION','labels':92,'grouped_labels_internal':64,
        'ungrouped_labels_internal':28,'target_occurrences':len(occ),'target_lines':len(line_rows),'candidate_snapshot_sha256':snapshot,
        'analysis_plan_sha256':plan_sha,'hapax_enrichment_authorized':False,'lexicon_matching_authorized':False})
    print(f'PREPARATION=COMPLETE; LABELS={len(candidates)}; OCCURRENCES={len(occ)}; PLAN_SHA256={plan_sha}')

def main():
    p=argparse.ArgumentParser();p.add_argument('--output-dir',type=Path,default=PKG);args=p.parse_args();build(args.output_dir)
if __name__=='__main__':main()
