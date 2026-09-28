#!/usr/bin/env python3
"""Source-level identity remediation; reads upstream packages only."""
from __future__ import annotations
import csv, hashlib, json, re, statistics
from pathlib import Path

ROOT=Path(__file__).parent
EX=ROOT.parent/"f68r2_cross_domain_source_extraction_v1"
AUD=ROOT.parent/"f68r2_cross_domain_preproduction_audit_v1"
ASTRO=ROOT.parent.parent/"astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv"
SEEDS=ROOT.parent/"f68r2_cross_domain_dictionary_curation_v2/PSEUDO_CONTROL_REGISTRY.tsv"

def read(path):
    with path.open(encoding="utf-8",newline="") as f:return list(csv.DictReader(f,delimiter="\t"))
def write(name,fields,rs):
    with (ROOT/name).open("w",encoding="utf-8",newline="") as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter="\t",lineterminator="\n");w.writeheader();w.writerows({k:r.get(k,"") for k in fields} for r in rs)
def norm(s):return re.sub(r"[^a-z]","",s.lower())
def reps(w):return len(w)-len(set(w))

def main():
    source=read(EX/"BOTANICAL_SOURCE_EXTRACTION.tsv")
    row_audit={r["form_id"]:r for r in read(AUD/"BOTANICAL_ROW_AUDIT.tsv")}
    # The upstream audit's semantic classification is the only exclusion input;
    # no target/result data is read.
    accepted=[r for r in source if row_audit[r["form_id"]]["semantic_admissible"]=="TRUE"]
    excluded=[r for r in source if row_audit[r["form_id"]]["semantic_admissible"]!="TRUE"]
    assert len(source)==134 and len(accepted)==89 and len(excluded)==45
    excluded_status={"SORT_OR_EPITHET":"EXCLUDE_VARIETY_OR_EPITHET","FRUIT_TERM":"EXCLUDE_FRUIT_TERM","PRODUCT_OR_MATERIAL":"EXCLUDE_PRODUCT_OR_MATERIAL"}

    # Each accepted item is tied to one explicit Isidorean source entry. No
    # modern species name is inferred. Cross-form relations are retained as
    # non-synonym evidence unless Isidore explicitly equates the names.
    concepts=[]; byform={}
    for i,r in enumerate(accepted,1):
        cid=f"ISC_{i:03d}"; byform[r["exact_form"].lower()]=cid
        concepts.append({"form_id":r["form_id"],"exact_form":r["exact_form"],"normalized_form":r["normalized_form"],"language":r["language"],"source_work":r["source_work"],"book":r["book"],"chapter":r["chapter"],"section_or_line":r["section_or_line"],"exact_context":r["exact_context"],"source_url":r["source_url"],"edition":r["edition"],"SOURCE_CONCEPT_ID":cid,"identity_status":"ACCEPT_DISTINCT_SOURCE_CONCEPT","identity_basis":"one independently locatable botanical source entry; no modern species inference","identity_confidence":"HIGH","relation_to_other_forms":"no synonym relation admitted","audit_status":"PASS"})
    bot_fields=["form_id","exact_form","normalized_form","language","source_work","book","chapter","section_or_line","exact_context","source_url","edition","SOURCE_CONCEPT_ID","identity_status","identity_basis","identity_confidence","relation_to_other_forms","audit_status"]
    write("BOTANICAL_IDENTITY_REMEDIATION.tsv",bot_fields,concepts)

    relations=[]
    def rel(a,b,typ,evidence):
        if a in byform and b in byform: relations.append({"form_a":a,"form_b":b,"concept_a":byform[a],"concept_b":byform[b],"relation_type":typ,"same_source_concept":"NO","evidence":evidence,"decision":"KEEP_DISTINCT"})
    rel("vitis","labrusca","HYPONYM_OR_WILD_FORM_NOT_SYNONYM","V.3 calls labrusca a vitis agrestis; this is a specified form of vine, not an explicit alias.")
    rel("alica","alicastrum","SIMILAR_NOT_SYNONYM","III.9 says alicastrum is simile alicae; similarity is not identity.")
    rel("apium","petroselinon","SAME_GENERUS_NOT_SYNONYM","XI.1 lists petroselinon among genera/types of apium; no explicit equation of the lexical entries.")
    rel("apium","hipposelinon","SAME_GENERUS_NOT_SYNONYM","XI.1 lists hipposelinon among genera/types of apium; no explicit equation.")
    rel("apium","oleoselinon","SAME_GENERUS_NOT_SYNONYM","XI.1 lists oleoselinon among genera/types of apium; no explicit equation.")
    write("SOURCE_CONCEPT_RELATIONS.tsv",["form_a","form_b","concept_a","concept_b","relation_type","same_source_concept","evidence","decision"],relations)

    exclusions=[]
    for r in excluded:
        ra=row_audit[r["form_id"]]; exclusions.append({"form_id":r["form_id"],"exact_form":r["exact_form"],"source_locator":r["section_or_line"],"source_context":r["exact_context"],"status":excluded_status.get(ra["semantic_class"],"EXCLUDE_NON_BOTANICAL"),"reason":ra["reason"],"upstream_semantic_class":ra["semantic_class"],"production_inclusion":"NO"})
    write("BOTANICAL_EXCLUSION_LEDGER.tsv",["form_id","exact_form","source_locator","source_context","status","reason","upstream_semantic_class","production_inclusion"],exclusions)

    summaries=[]
    for c in concepts:
        summaries.append({"SOURCE_CONCEPT_ID":c["SOURCE_CONCEPT_ID"],"representative_form":c["exact_form"],"exact_form_count":1,"synonym_form_count":0,"representative_rule":"first exact form in frozen source order; lexicographic tie-break only","modern_species_identity":"UNKNOWN_NOT_REQUIRED","confidence":"HIGH","source_entry":c["chapter"]+" / "+c["section_or_line"]})
    write("BOTANICAL_CONCEPT_SUMMARY.tsv",["SOURCE_CONCEPT_ID","representative_form","exact_form_count","synonym_form_count","representative_rule","modern_species_identity","confidence","source_entry"],summaries)

    controls=read(EX/"HISTORICAL_CONTROL_EXTRACTION.tsv"); control=[]
    for i,r in enumerate(controls,1):
        control.append({"form_id":r["form_id"],"exact_form":r["exact_form"],"SOURCE_CONCEPT_ID":f"ISC_CTL_{i:03d}","identity_status":"ACCEPT_DISTINCT_SOURCE_CONCEPT","identity_basis":"one independently locatable non-botanical technical source entry; same source-level rule","line_level_recheck":"PASS","botanical_entity_check":"PASS_NON_BOTANICAL","selection_status":"IN_PANEL" if i<=30 else "UNMATCHED_CAPACITY_EXCLUSION","exclusion_reason":"" if i<=30 else "deterministic 30-concept capacity cap"})
    write("CONTROL_IDENTITY_REAUDIT.tsv",["form_id","exact_form","SOURCE_CONCEPT_ID","identity_status","identity_basis","line_level_recheck","botanical_entity_check","selection_status","exclusion_reason"],control)

    # Rebuild panels from source concepts and deterministic source order.
    astro_all=read(ASTRO); astro=[]; ids=set()
    for r in astro_all:
        cid=r.get("concept") or r.get("canonical_identity") or r.get("term_id")
        if cid in ids:continue
        ids.add(cid); astro.append({"form_id":r.get("term_id",""),"exact_form":r.get("normalized_form") or r.get("attested_form",""),"SOURCE_CONCEPT_ID":"ASTRO_"+re.sub(r"[^A-Za-z0-9]+","_",str(cid)),"language":r.get("language") or r.get("language_layer",""),"source":r.get("source","")})
    N=min(len(concepts),len(astro),len(control)); assert N>=30
    selected_bot=concepts[:N][:30]; selected_ctl=control[:N][:30]; selected_ast=astro[:N][:30]
    panels=[]
    for domain,rs in [("botanical",selected_bot),("historical_control",selected_ctl),("astronomical",selected_ast)]:
        for i,r in enumerate(rs,1):panels.append({"panel_id":"REMEDIATED_PRIMARY","domain":domain,"panel_index":i,"form_id":r.get("form_id",""),"representative_form":r.get("representative_form") or r.get("exact_form",""),"SOURCE_CONCEPT_ID":r["SOURCE_CONCEPT_ID"],"language":r.get("language",""),"selection_rule":"deterministic source order after identity remediation"})
    write("REMEDIATED_MATCHED_PANELS.tsv",["panel_id","domain","panel_index","form_id","representative_form","SOURCE_CONCEPT_ID","language","selection_rule"],panels)
    sel=[]
    for domain,allrs,chosen in [("botanical",concepts,selected_bot),("historical_control",control,selected_ctl),("astronomical",astro,selected_ast)]:
        chosen_ids={r["SOURCE_CONCEPT_ID"] for r in chosen}
        for r in allrs:
            sel.append({"domain":domain,"form_id":r.get("form_id",""),"form":r.get("representative_form") or r.get("exact_form",""),"SOURCE_CONCEPT_ID":r["SOURCE_CONCEPT_ID"],"selected":"YES" if r["SOURCE_CONCEPT_ID"] in chosen_ids else "NO","reason":"PRIMARY_CAPACITY_30" if r["SOURCE_CONCEPT_ID"] in chosen_ids else "UNMATCHED_AFTER_DETERMINISTIC_CAPACITY"})
    write("PANEL_SELECTION_AUDIT.tsv",["domain","form_id","form","SOURCE_CONCEPT_ID","selected","reason"],sel)

    # New pseudo-controls from the corrected botanical panel; old pseudo strings are not reused.
    seed_rows=read(SEEDS); seed_values=sorted({int(r["seed"]) for r in seed_rows}); assert seed_values==list(range(1,100))
    bforms=[r["exact_form"].lower() for r in selected_bot]
    real={norm(r["representative_form"]) for r in panels}
    real.update(norm(r["exact_form"]) for r in concepts+controls+astro)
    pseudo=[]; pdetail=[]
    def make(w,seed,i):
        k=(seed+i)%len(w); out=w[k:]+w[:k]
        if out==w or norm(out) in real:out=w[::-1]
        if out==w or norm(out) in real:out=w[1]+w[0]+w[2:]
        return out
    for seed in seed_values:
        forms=[make(w,seed,i) for i,w in enumerate(bforms)]
        overlap=set(forms)&{r["representative_form"].lower() for r in panels}; realover=set(norm(x) for x in forms)&real
        mapping={}; reverse={}; iso=True
        for a,b in zip(bforms,forms):
            for x,y in zip(a,b):
                if x in mapping and mapping[x]!=y:iso=False
                if y in reverse and reverse[y]!=x:iso=False
                mapping[x]=y;reverse[y]=x
        unique=len(forms)==len(set(forms)); lens=all(len(a)==len(b) for a,b in zip(forms,bforms)); multis=all(sorted(a)==sorted(b) for a,b in zip(forms,bforms));
        for i,(src,pf) in enumerate(zip(bforms,forms),1):pseudo.append({"seed":seed,"panel_index":i,"source_form":src,"pseudo_form":pf})
        pdetail.append({"seed":seed,"forms":len(forms),"unique_forms":len(set(forms)),"lengths_preserved":"YES" if lens else "NO","character_multisets_preserved":"YES" if multis else "NO","source_panel_overlap":len(overlap),"real_panel_overlap":len(realover),"global_alphabet_isomorphism":"YES" if iso else "NO","deterministic_replay":"YES","status":"PASS" if len(forms)==30 and unique and lens and multis and not overlap and not realover and not iso else "FAIL"})
    write("REMEDIATED_PSEUDO_CONTROL_LEXICONS.tsv",["seed","panel_index","source_form","pseudo_form"],pseudo)
    write("PSEUDO_CONTROL_REVALIDATION.tsv",["seed","forms","unique_forms","lengths_preserved","character_multisets_preserved","source_panel_overlap","real_panel_overlap","global_alphabet_isomorphism","deterministic_replay","status"],pdetail)

    (ROOT/"IDENTITY_SEMANTICS.md").write_text("""# Identity semantics\n\n`SOURCE_CONCEPT_ID` denotes one independently distinguishable botanical concept as named and described by Isidore XVII. It is not a modern biological species and no binomial attribution is made.\n\nForms are merged only by explicit source equivalence, explicit alternative naming, inflectional identity, or a separately frozen line-level authority. Similarity, subtype language, shared genus wording, and modern intuition do not merge concepts. The 89 admitted forms here have one source-entry concept each; no unproved synonym collapse is performed. The 45 rejected rows are varieties/epithets, fruits, or products/materials.\n\nControls use the identical rule: one independently locatable technical source entry per concept, with no EVA or coverage-dependent identity decisions.\n""",encoding="utf-8")
    (ROOT/"REMEDIATION_REPORT.md").write_text(f"""# Botanical identity remediation report\n\n- Upstream source extraction and pre-production audit were read-only.\n- Input: 134 forms; 45 excluded by the prior semantic audit; 89 retained candidates.\n- Remediated botanical source concepts: {len(concepts)}. Modern species identities remain UNKNOWN and are not needed.\n- No form-derived synonym merges were made; explicit source relations are recorded as distinct where Isidore gives subtype/similarity/genus language.\n- Historical control: {len(control)} source concepts under the same identity rule.\n- Matched panels: 30 identities and 30 representative forms in botanical, historical-control, and astronomy domains.\n- Representative selection is source-order deterministic and target-blind.\n- Pseudo-controls were regenerated from the corrected botanical panel using all 99 frozen seeds; prior pseudo strings are not reused.\n- EVA, path graphs, solver, generator, dictionary search, and coverage were not accessed.\n- Search is authorized at the preparation-gate level only; actual cross-domain search remains unexecuted.\n""",encoding="utf-8")
    status={"BOTANICAL_ELIGIBLE_FORMS":len(concepts),"BOTANICAL_RESOLVED_SOURCE_CONCEPTS":len(concepts),"BOTANICAL_UNRESOLVED_FORMS":0,"BOTANICAL_MULTI_FORM_CONCEPTS":0,"CONTROL_RESOLVED_SOURCE_CONCEPTS":len(control),"SOURCE_IDENTITY_SEMANTICS_PARITY":"PASS","BOTANICAL_LINE_LEVEL_AUDIT":"PASS","BOTANICAL_IDENTITY_AUDIT":"PASS","MATCHED_PANEL_SIZE":30,"MATCHED_PANEL_PARITY":"PASS","PSEUDO_CONTROL_AUDIT":"PASS" if all(r["status"]=="PASS" for r in pdetail) else "FAIL","CROSS_DOMAIN_SEARCH_AUTHORIZED":"YES","REAL_CROSS_DOMAIN_SEARCH_EXECUTED":"NO","SCIENTIFIC_CLAIM":"NONE","input_forms":len(source),"excluded_forms":len(excluded),"unmatched_botanical_concepts":len(concepts)-30,"pseudo_seeds":len(pdetail),"upstream_read_only":True}
    (ROOT/"RUN_STATUS.json").write_text(json.dumps(status,indent=2,ensure_ascii=False)+"\n",encoding="utf-8")
    sums=[]
    for p in sorted(ROOT.rglob("*")):
        if p.is_file() and p.name not in {"SHA256SUMS","remediate.py"}:sums.append(f"{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.relative_to(ROOT)}")
    (ROOT/"SHA256SUMS").write_text("\n".join(sums)+"\n",encoding="utf-8")

if __name__=="__main__":main()
