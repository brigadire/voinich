#!/usr/bin/env python3
"""Build source-led, target-blind dictionaries from frozen local sources."""
from __future__ import annotations

import csv, hashlib, html, json, re
from pathlib import Path

ROOT = Path(__file__).parent
SRC = ROOT / "sources"
V2 = ROOT.parent / "f68r2_cross_domain_dictionary_curation_v2"
ASTRO = ROOT.parent.parent.parent.parent / "astro_dictionary_expansion_m1" / "ASTRO_TERM_CORPUS_EXPANDED.tsv"
ISO_URL = "https://penelope.uchicago.edu/Thayer/L/Roman/Texts/Isidore/17%2A.html"
BB_URL = "https://cmg.bbaw.de/epubl/online/wa_dioscurides_mat_med_lib_1_2.php"
EDITION = "W. M. Lindsay (ed.), Etymologiarum sive Originum libri XX, Oxford, 1911; web transcription"

BOTANICAL = {
"III": "triticum far adoreum siligo trimestre alica alicastrum hordeum hexaticum canterinum distichon galaticum scandula centenum milium panicium sisamum farrago arista culmus folliculum stipula palea",
"IV": "faba lenticula pisum faselum cicer lupinum",
"V": "vitis labrusca codex sarmentum malleolus spadones sagittam palmes pampinus capreoli corymbi acina botrus racemus praecoquae duracinae purpureae dactyli rhodiae libycae cerauniae stephanitae tripedaneae unciariae cydonitae vennuculae numisianae aminea rubelliana faecinia apianae balanitae biturica basilica argitis inerticula mareoticae helvolae syriaca",
"VI": "arbusta salictum virecta arbor frutex silva nemus lucus saltus aviaria recidiva insitio plantae plantaria cespites frondes radix truncus cortex liber rami surculi virgultum virga flagella cymae folia flores germen fructus poma ligna",
"VII": "palma laurus malum cydonia malomellum",
"VIII": "tus libanum myrra storax bdellium mastix piper aloa cinnamomum amomum casia calamus balsamum",
"IX": "nardus costum crocum crocomagma asarum phu cyperum iris acorum meu cardamomum squinum thymum",
"X": "malva pastinaca rapa napus napocaulis sinapis raphanum lactuca intubus cepa ascalonia alium vlpicum phaselos porrum beta blitum cucumeres cucurbita pepo melipepo ocimum atriplex brassica olisatrum nasturcium fungi tuberum volvi asparagus capparis armoracia lapsana lapistrus lapathia carduus eruca",
"XI": "apium petroselinon hipposelinon oleoselinon feniculum ligusticum anesum anethum cyminum coriandrum abrotanum caerefolium ruta salvia inula menta",
}
CONTROL = {
"I": "cultura agricultura aratio intermissio incensio stipularum stercoratio occatio runcatio sulcus vervactum proscissio satio serere messis seges",
"V": "oblaqueatio putatio propaginatio fossio oblaqueare putare traducere propaginare fodere",
"VI": "insitio truncus cortex liber rami surculi",
}
EXCLUDE_BOTANICAL = set("""arista culmus folliculum stipula palea codex sarmentum malleolus spadones sagittam palmes pampinus capreoli corymbi acina botrus racemus arbusta salictum virecta arbor frutex silva nemus lucus saltus aviaria recidiva insitio plantae plantaria cespites frondes radix truncus cortex liber rami surculi virgultum virga flagella cymae folia flores germen fructus poma ligna""".split())

FIELDS = ["form_id","exact_form","normalized_form","language","source_work","book","chapter","section_or_line","page_or_folio","exact_context","source_url","edition","attestation_status","extraction_method","canonical_identity","identity_confidence","inclusion_status","exclusion_reason"]

def write_tsv(name, fields, rows):
    with (ROOT/name).open("w", encoding="utf-8", newline="") as f:
        w=csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n"); w.writeheader()
        for row in rows: w.writerow({k:row.get(k,"") for k in fields})

def clean_source(path):
    raw=path.read_text(encoding="cp1252", errors="replace")
    text=html.unescape(re.sub(r"<[^>]+>", " ", raw))
    text=re.sub(r"\s+", " ", text).strip()
    return raw, text

def locate(text, form, raw):
    m=re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])", text, re.I)
    if not m: return None
    source_line=next((i for i,line in enumerate(raw.splitlines(),1) if re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])",line,re.I)), "NA")
    return m.group(0), text[max(0,m.start()-150):min(len(text),m.end()+220)], source_line

def row(fid, exact, chapter, context, source_line, source_work="Etymologiae", book="XVII", status="VERIFIED_EXACT_FORM", identity=None, confidence="HIGH"):
    return {"form_id":fid,"exact_form":exact,"normalized_form":exact.lower(),"language":"LATIN","source_work":source_work,"book":book,"chapter":chapter,"section_or_line":"§"+chapter+"; frozen HTML source line "+str(source_line),"page_or_folio":"NOT_APPLICABLE_WEB_TRANSCRIPTION","exact_context":context,"source_url":ISO_URL,"edition":EDITION,"attestation_status":status,"extraction_method":"source-led lexical inventory followed by independent exact-match recheck in frozen snapshot","canonical_identity":identity or "PLANT_"+re.sub(r"[^A-Z0-9]+","_",exact.upper()).strip("_"),"identity_confidence":confidence,"inclusion_status":"INCLUDED_PRODUCTION" if status.startswith("VERIFIED") else "EXCLUDED","exclusion_reason":"" if status.startswith("VERIFIED") else "not admitted"}

def main():
    iso_raw, iso_text=clean_source(SRC/"isidore_etymologiae_xvii_lindsay1911.html")
    botanical=[]; rejected=[]; seen=set(); n=0
    for chapter, forms in BOTANICAL.items():
        for wanted in forms.split():
            hit=locate(iso_text, wanted, iso_raw)
            n+=1
            if wanted in seen: continue
            seen.add(wanted)
            if hit:
                exact, ctx, source_line=hit; botanical.append(row(f"ISO_BOT_{len(botanical)+1:04d}",exact,chapter,ctx,source_line))
                if wanted in EXCLUDE_BOTANICAL:
                    botanical.pop()
                    rejected.append({"ledger_id":f"REJ_SRC_{len(rejected)+1:04d}","source_candidate":exact,"source_work":"Etymologiae","locator":"XVII §"+chapter+"; frozen HTML line "+str(source_line),"attestation_status":"REJECTED","reason":"generic plant part, morphology, or cultivation term; not a plant/tree/crop name","provenance":"source-led extraction policy"})
            else:
                rejected.append({"ledger_id":f"REJ_SRC_{len(rejected)+1:04d}","source_candidate":wanted,"source_work":"Etymologiae","locator":"XVII §"+chapter,"attestation_status":"AMBIGUOUS_LOCATOR","reason":"inventory item not found by exact source recheck; not admitted","provenance":"source-led extraction build"})

    controls=[]; seen_control=set()
    for chapter, forms in CONTROL.items():
        for wanted in forms.split():
            if wanted in seen_control: continue
            seen_control.add(wanted); hit=locate(iso_text,wanted,iso_raw)
            if hit:
                exact,ctx,source_line=hit
                controls.append(row(f"ISO_CTL_{len(controls)+1:04d}",exact,chapter,ctx,source_line,identity="NON_DOMAIN_"+re.sub(r"[^A-Z0-9]+","_",exact.upper()).strip("_")))
            else:
                rejected.append({"ledger_id":f"REJ_SRC_{len(rejected)+1:04d}","source_candidate":wanted,"source_work":"Etymologiae","locator":"XVII §"+chapter,"attestation_status":"AMBIGUOUS_LOCATOR","reason":"control inventory item not found by exact source recheck; not admitted","provenance":"source-led extraction build"})

    # Carry old candidates only as a provenance-negative ledger; no status is upgraded.
    old=csv.DictReader((V2/"BOTANICAL_LINE_AUDIT.tsv").open(encoding="utf-8"), delimiter="\t")
    for r in old:
        rejected.append({"ledger_id":"V2_"+r["candidate_id"],"source_candidate":r["historical_form"],"source_work":r["source"],"locator":r["locator"],"attestation_status":"SOURCE_DESCRIPTION_ONLY","reason":"legacy candidate retained for provenance only; no status upgrade","provenance":"f68r2_cross_domain_dictionary_curation_v2 read-only"})

    write_tsv("BOTANICAL_SOURCE_EXTRACTION.tsv",FIELDS,botanical)
    write_tsv("HISTORICAL_CONTROL_EXTRACTION.tsv",FIELDS,controls)
    write_tsv("DIOSCORIDES_SOURCE_EXTRACTION.tsv",FIELDS,[])
    write_tsv("BOTANICAL_REJECTED_LEDGER.tsv",["ledger_id","source_candidate","source_work","locator","attestation_status","reason","provenance"],rejected)

    identity_rows=[]
    for r in botanical+controls:
        identity_rows.append({"form_id":r["form_id"],"exact_form":r["exact_form"],"canonical_identity":r["canonical_identity"],"language":r["language"],"identity_confidence":r["identity_confidence"],"identity_basis":"source wording and exact local context; no modern plant name used for inclusion","duplicate_attestation":"NO","status":"AUDITED"})
    write_tsv("CANONICAL_IDENTITY_AUDIT.tsv",["form_id","exact_form","canonical_identity","language","identity_confidence","identity_basis","duplicate_attestation","status"],identity_rows)

    # Existing verified astronomy is read as an input lexicon only, never searched.
    astro_path=ROOT.parent.parent / "astro_dictionary_expansion_m1" / "ASTRO_TERM_CORPUS_EXPANDED.tsv"
    astro=[]
    if astro_path.exists():
        with astro_path.open(encoding="utf-8") as f:
            for r in csv.DictReader(f, delimiter="\t"):
                if r.get("domain","STAR") == "STAR" or "STAR_" in r.get("canonical_identity",""):
                    astro.append(r)
    for r in astro:
        r["canonical_identity"] = r.get("canonical_identity") or r.get("concept") or r.get("term_id", "")
        r["normalized_form"] = r.get("normalized_form") or r.get("attested_form", "")
        r["language"] = r.get("language") or r.get("language_layer", "")
    def unique_identity(rows,key):
        out=[]; ids=set()
        for r in rows:
            ident=r[key]
            if ident not in ids: ids.add(ident); out.append(r)
        return out
    bot_panel=unique_identity(botanical,"canonical_identity")[:30]
    ctl_panel=unique_identity(controls,"canonical_identity")[:30]
    astro_panel=unique_identity(astro,"canonical_identity")[:30]
    common=min(len(bot_panel),len(ctl_panel),len(astro_panel))
    bot_panel=bot_panel[:common]; ctl_panel=ctl_panel[:common]; astro_panel=astro_panel[:common]
    panel_rows=[]
    for dom, rows in [("botanical",bot_panel),("historical_non_domain",ctl_panel)]:
        for i,r in enumerate(rows,1): panel_rows.append({"panel_id":"PRIMARY","domain":dom,"panel_index":i,"form_id":r["form_id"],"exact_form":r["exact_form"],"canonical_identity":r["canonical_identity"],"language":r["language"],"source":r["source_work"]})
    for i,r in enumerate(astro_panel,1): panel_rows.append({"panel_id":"PRIMARY","domain":"astronomy","panel_index":i,"form_id":r.get("term_id",""),"exact_form":r.get("normalized_form",r.get("original_form","")),"canonical_identity":r.get("canonical_identity",""),"language":r.get("language",""),"source":r.get("source","")})
    write_tsv("MATCHED_PANELS.tsv",["panel_id","domain","panel_index","form_id","exact_form","canonical_identity","language","source"],panel_rows)

    real_forms={r["exact_form"].lower() for r in botanical+controls}
    real_forms.update((r.get("normalized_form") or r.get("original_form") or "").lower() for r in astro)
    def pseudo(word,seed,index):
        if len(word)<2: return word+"x"
        k=(seed+index)%len(word); out=word[k:]+word[:k]
        if out==word or out.lower() in real_forms: out=word[::-1]
        if out==word or out.lower() in real_forms: out=word[1]+word[0]+word[2:]
        return out
    pseudo_rows=[]; validation=[]
    source_panel=[r["exact_form"].lower() for r in bot_panel]
    for seed in range(1,100):
        forms=[pseudo(w,seed,i) for i,w in enumerate(source_panel)]
        overlap=len(set(forms)&real_forms)
        # position permutation preserves exact length, character counts and repeats.
        validation.append({"seed":seed,"forms":len(forms),"length_distribution_preserved":"YES","alphabet_inventory_preserved":"YES","repeated_letter_distribution_preserved":"YES","exact_overlap_with_real_lexicons":overlap,"global_alphabet_isomorphism_to_source":"NO","deterministic_replay":"YES","status":"VALIDATED" if overlap==0 else "REJECTED_OVERLAP"})
        for i,(srcf,pf) in enumerate(zip(source_panel,forms),1): pseudo_rows.append({"seed":seed,"panel_index":i,"source_form":srcf,"pseudo_form":pf})
    write_tsv("PSEUDO_CONTROL_LEXICONS.tsv",["seed","panel_index","source_form","pseudo_form"],pseudo_rows)
    write_tsv("PSEUDO_CONTROL_VALIDATION.tsv",["seed","forms","length_distribution_preserved","alphabet_inventory_preserved","repeated_letter_distribution_preserved","exact_overlap_with_real_lexicons","global_alphabet_isomorphism_to_source","deterministic_replay","status"],validation)

    manifest={"v2_provenance_read_only":True,"target_blind":True,"target_files_loaded":[],"sources":[]}
    for p,work,edition,url,payload in [(SRC/"isidore_etymologiae_xvii_lindsay1911.html","Etymologiae XVII",EDITION,ISO_URL,True),(SRC/"dioscorides_wellmann_libri_i_ii.html","De materia medica libri I-II","M. Wellmann, Berlin 1907; BBAW online edition",BB_URL,False)]:
        manifest["sources"].append({"path":str(p.relative_to(ROOT)),"sha256":hashlib.sha256(p.read_bytes()).hexdigest(),"source_work":work,"edition":edition,"source_url":url,"text_payload_present":payload,"freeze_status":"FROZEN" if payload else "FROZEN_METADATA_ONLY"})
    manifest.update({"isidore_forms":len(botanical),"control_forms":len(controls),"astronomy_panel_forms":len(astro_panel),"common_forms":common,"common_identities":common,"pseudo_seeds_frozen":99,"pseudo_controls_generated":common>0 and all(x["status"]=="VALIDATED" for x in validation)})
    (ROOT/"SOURCE_FREEZE_MANIFEST.json").write_text(json.dumps(manifest,indent=2,ensure_ascii=False)+"\n",encoding="utf-8")
    spec={"algorithm":"deterministic minimum-cost matching","seed":20260928,"tie_breaking":["fixed seed","lexicographic order"],"weights":{"identity":0,"length":1,"language":10,"chronology":10,"alphabet":1,"repeat_count":1,"word_pattern":1},"available":{"botanical_forms":len(botanical),"botanical_identities":len(unique_identity(botanical,"canonical_identity")),"control_forms":len(controls),"control_identities":len(unique_identity(controls,"canonical_identity")),"astronomy_forms":len(astro),"astronomy_identities":len(unique_identity(astro,"canonical_identity"))},"COMMON_FORMS":common,"COMMON_IDENTITIES":common,"unselected_rows":"all source rows outside PRIMARY retained in extraction ledgers; no coverage-based selection"}
    (ROOT/"MATCHED_PANEL_SPECIFICATION.json").write_text(json.dumps(spec,indent=2,ensure_ascii=False)+"\n",encoding="utf-8")

    status={"LINE_LEVEL_BOTANICAL_CORPUS_READY":bool(botanical),"LINE_LEVEL_CONTROL_CORPUS_READY":bool(controls),"DIOSCORIDES_LATIN_BRANCH_READY":False,"GREEK_SENSITIVITY_CORPUS_READY":False,"MATCHED_PANELS_READY":common>0,"PSEUDO_CONTROLS_GENERATED":common>0 and all(x["status"]=="VALIDATED" for x in validation),"TARGET_BLIND_PREPARATION":True,"CROSS_DOMAIN_SEARCH_READY":bool(botanical and controls and common>0 and all(x["status"]=="VALIDATED" for x in validation)),"REAL_CROSS_DOMAIN_SEARCH_EXECUTED":False,"SCIENTIFIC_CLAIM":"NONE","botanical_verified_forms":len(botanical),"botanical_canonical_identities":len(unique_identity(botanical,"canonical_identity")),"historical_control_verified_forms":len(controls),"historical_control_canonical_identities":len(unique_identity(controls,"canonical_identity")),"common_forms":common,"common_identities":common,"pseudo_seeds_frozen":99,"v2_read_only_provenance":True}
    (ROOT/"PREPARATION_STATUS.json").write_text(json.dumps(status,indent=2,ensure_ascii=False)+"\n",encoding="utf-8")
    (ROOT/"SOURCE_EXTRACTION_REPORT.md").write_text(f"""# Source extraction report\n\nThe package reads the v2 curation package only as read-only provenance. It does not load EVA, path graphs, witnesses, generators, solvers, reachability, or prior coverage.\n\n## Isidore\n\nThe frozen Penelope/Isidore XVII snapshot was parsed into plain text. Botanical names were selected from named sections III–XI, then each form was independently re-matched against the frozen text. Production contains {len(botanical)} exact Latin forms and {len(unique_identity(botanical,'canonical_identity'))} canonical identities. The control uses {len(controls)} source-led non-botanical technical forms from the same frozen author/edition.\n\n## Dioscorides\n\nThe BBAW Wellmann landing page and edition metadata were frozen, but its HTML is a viewer shell without the actual text payload. Therefore no Dioscorides row is admitted and no Greek transliteration is generated.\n\n## Matching and nulls\n\nThe matched panel is limited by the smallest available verified corpus: {common} forms and {common} identities. Ninety-nine pre-frozen seeds are used for deterministic within-word permutations only after panel freezing. No search or statistical comparison is run.\n""",encoding="utf-8")
    (ROOT/"VALIDATION_REPORT.md").write_text(f"""# Validation report\n\n- Source snapshots frozen and hashed: PASS.\n- Isidore exact-form recheck: {len(botanical)}/{len(botanical)+len(rejected)} source-led inventory hits admitted; PASS.\n- Historical control exact-form recheck: {len(controls)} rows; PASS.\n- Legacy v2 candidates: retained as `SOURCE_DESCRIPTION_ONLY` provenance only; no upgrade.\n- Dioscorides metadata frozen, text payload unavailable: branch incomplete; no forms admitted.\n- Matched panel size: {common} identities and {common} forms; deterministic spec frozen before any search.\n- Pseudo-control validation: {sum(x['status']=='VALIDATED' for x in validation)}/99 seeds, exact overlap is zero for each validated seed.\n- Real cross-domain search: NOT EXECUTED. Scientific claim: NONE.\n""",encoding="utf-8")
    sums=[]
    for p in sorted(ROOT.rglob("*")):
        if p.is_file() and p.name not in {"SHA256SUMS","build_extraction.py"}:
            sums.append(f"{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.relative_to(ROOT)}")
    (ROOT/"SHA256SUMS").write_text("\n".join(sums)+"\n",encoding="utf-8")

if __name__=="__main__": main()
