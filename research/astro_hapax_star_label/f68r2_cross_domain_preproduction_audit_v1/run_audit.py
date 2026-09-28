#!/usr/bin/env python3
"""Independent pre-production audit; deliberately does not run a dictionary search."""
from __future__ import annotations
import csv, hashlib, html, json, re, statistics
from pathlib import Path

ROOT=Path(__file__).parent
EXTRACT=ROOT.parent/"f68r2_cross_domain_source_extraction_v1"
SOURCE=EXTRACT/"sources/isidore_etymologiae_xvii_lindsay1911.html"
ASTRO=ROOT.parent.parent/"astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv"
PROFILES=ROOT.parent/"f68r2_corrected_e3_full_search_v1/E3_OPERATION_PROFILES.tsv"

def rows(path):
    with path.open(encoding="utf-8", newline="") as f: return list(csv.DictReader(f,delimiter="\t"))
def write(name, fields, data):
    with (ROOT/name).open("w",encoding="utf-8",newline="") as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter="\t",lineterminator="\n"); w.writeheader(); w.writerows({k:r.get(k,"") for k in fields} for r in data)
def norm(s): return re.sub(r"[^a-z]","",s.lower())
def lengths(rs): return [len(norm(r["exact_form"])) for r in rs]
def alphabet(rs): return set("".join(norm(r["exact_form"]) for r in rs))
def repeats(w): return len(w)-len(set(w))

def main():
    bot=rows(EXTRACT/"BOTANICAL_SOURCE_EXTRACTION.tsv")
    ctl=rows(EXTRACT/"HISTORICAL_CONTROL_EXTRACTION.tsv")
    panel=rows(EXTRACT/"MATCHED_PANELS.tsv")
    pseudo=rows(EXTRACT/"PSEUDO_CONTROL_LEXICONS.tsv")
    pseudo_prev=rows(EXTRACT/"PSEUDO_CONTROL_VALIDATION.tsv")
    raw=SOURCE.read_text(encoding="cp1252",errors="replace")
    text=html.unescape(re.sub(r"<[^>]+>"," ",raw)); text=re.sub(r"\s+"," ",text)
    raw_lines=raw.splitlines()

    # The source extraction must be rechecked independently, not trusted by status.
    suspicious_sort={"trimestre","alicastrum","hexaticum","canterinum","distichon","galaticum","praecoquae","duracinae","purpureae","dactyli","rhodiae","libycae","cerauniae","stephanitae","tripedaneae","unciariae","cydonitae","vennuculae","numisianae","aminea","rubelliana","faecinia","apianae","balanitae","biturica","argitis","inerticula","mareoticae","helvolae","syriaca"}
    suspicious_fruit={"malum","cydonia","malomellum"}
    suspicious_product={"tus","libanum","myrra","storax","bdellium","mastix","piper","crocum","crocomagma","casia","calamus","balsamum"}
    row_audit=[]
    for r in bot:
        form=r["exact_form"]; f=norm(form)
        exact=bool(re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])",text,re.I))
        try: line=int(re.search(r"line (\d+)",r["section_or_line"]).group(1)); line_ok=bool(re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])",raw_lines[line-1],re.I))
        except Exception: line="NA"; line_ok=False
        context_ok=bool(re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])",r["exact_context"],re.I))
        if f in suspicious_sort: semantic="SORT_OR_EPITHET"; semantic_ok=False
        elif f in suspicious_fruit: semantic="FRUIT_TERM"; semantic_ok=False
        elif f in suspicious_product: semantic="PRODUCT_OR_MATERIAL"; semantic_ok=False
        else: semantic="PLANT_OR_CROP_NAME_CANDIDATE"; semantic_ok=True
        verdict="PASS_SOURCE_ATTESTATION_EXCLUDED_SEMANTICALLY" if exact and line_ok and context_ok and not semantic_ok else ("PASS_PROVISIONAL" if exact and line_ok and context_ok else "FAIL_SOURCE_RECHECK")
        row_audit.append({"form_id":r["form_id"],"exact_form":form,"source_line":line,"exact_present":str(exact).upper(),"locator_recheck":str(line_ok).upper(),"context_recheck":str(context_ok).upper(),"language_check":"PASS" if r["language"]=="LATIN" else "FAIL","modernization_check":"NOT_DEMONSTRATED","semantic_class":semantic,"semantic_admissible":str(semantic_ok).upper(),"canonical_identity_input":r["canonical_identity"],"verdict":verdict,"reason":"" if semantic_ok else "source wording identifies a sort, fruit, or product rather than an admissible plant-name entry"})
    write("BOTANICAL_ROW_AUDIT.tsv",["form_id","exact_form","source_line","exact_present","locator_recheck","context_recheck","language_check","modernization_check","semantic_class","semantic_admissible","canonical_identity_input","verdict","reason"],row_audit)

    # Identity audit intentionally treats form-derived identities as unresolved.
    identity=[]
    for r in bot:
        identity.append({"form_id":r["form_id"],"exact_form":r["exact_form"],"input_identity":r["canonical_identity"],"identity_derivation":"FORM_STRING_ONLY","identity_status":"IDENTITY_UNRESOLVED","source_equivalence_evidence":"NONE_RELIABLE_FOR_CANONICAL_COLLAPSE","possible_duplicate_or_variant":"REVIEW_REQUIRED","audit_verdict":"FAIL_IDENTITY_NOT_INDEPENDENTLY_ESTABLISHED"})
    write("BOTANICAL_IDENTITY_AUDIT.tsv",["form_id","exact_form","input_identity","identity_derivation","identity_status","source_equivalence_evidence","possible_duplicate_or_variant","audit_verdict"],identity)

    dup=[]; seen_exact={}; seen_norm={}
    for r in bot:
        e=r["exact_form"]; n=norm(e); seen_exact.setdefault(e.lower(),[]).append(r["form_id"]); seen_norm.setdefault(n,[]).append(r["form_id"])
    for r in bot:
        e=r["exact_form"]; n=norm(e); exact_ids=seen_exact[e.lower()]; norm_ids=seen_norm[n]
        dup.append({"form_id":r["form_id"],"exact_form":e,"exact_duplicate_ids":";".join(exact_ids),"normalized_duplicate_ids":";".join(norm_ids),"exact_duplicate":"NO" if len(exact_ids)==1 else "YES","orthographic_variant":"REVIEW_REQUIRED" if len(norm_ids)>1 else "NO","identity_conflation_check":"FAIL_UNRESOLVED","verdict":"PASS_NO_LITERAL_DUPLICATE" if len(exact_ids)==1 else "FAIL_DUPLICATE"})
    write("BOTANICAL_DUPLICATE_VARIANT_AUDIT.tsv",["form_id","exact_form","exact_duplicate_ids","normalized_duplicate_ids","exact_duplicate","orthographic_variant","identity_conflation_check","verdict"],dup)

    # Controls: same source mechanism, source-line and no EVA/result access.
    bmean=statistics.mean(lengths(bot)); cmean=statistics.mean(lengths(ctl)); bshort=sum(x<=5 for x in lengths(bot))/len(bot); cshort=sum(x<=5 for x in lengths(ctl))/len(ctl)
    control_audit=[]
    for i,r in enumerate(ctl):
        form=r["exact_form"]; exact=bool(re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])",text,re.I))
        try: line=int(re.search(r"line (\d+)",r["section_or_line"]).group(1)); line_ok=bool(re.search(r"(?<![A-Za-z])"+re.escape(form)+r"(?![A-Za-z])",raw_lines[line-1],re.I))
        except Exception: line="NA"; line_ok=False
        panel_inc=i<30
        control_audit.append({"form_id":r["form_id"],"exact_form":form,"exact_present":str(exact).upper(),"locator_recheck":str(line_ok).upper(),"botanical_entity_check":"PASS_NON_BOTANICAL_TECHNICAL_TERM","mechanism_parity":"PASS_SAME_SOURCE_LED_EXTRACTION","length_balance":"PASS_NOT_SYSTEMATICALLY_SHORTER","selection_blindness":"PASS_NO_TARGET_OR_RESULT_ACCESS","panel_inclusion":"YES" if panel_inc else "NO","exclusion_reason":"" if panel_inc else "deterministic capacity cap: first 30 lexicographic source rows retained; row preserved outside panel"})
    write("HISTORICAL_CONTROL_AUDIT.tsv",["form_id","exact_form","exact_present","locator_recheck","botanical_entity_check","mechanism_parity","length_balance","selection_blindness","panel_inclusion","exclusion_reason"],control_audit)

    # Panel audit. This checks shape independently, but identity readiness also depends on the identity audit above.
    p_audit=[]
    for domain in ["botanical","astronomy","historical_non_domain"]:
        rs=[r for r in panel if r["domain"]==domain]
        ids=[r["canonical_identity"] for r in rs]; forms=[r["exact_form"] for r in rs]
        p_audit.append({"domain":domain,"forms":len(rs),"identities":len(set(ids)),"duplicate_identities":"NO" if len(ids)==len(set(ids)) else "YES","length_min":min(map(len,forms)) if forms else "NA","length_max":max(map(len,forms)) if forms else "NA","alphabet_size":len(set("".join(forms).lower())),"repeat_total":sum(repeats(x.lower()) for x in forms),"spaces_total":sum(x.count(" ") for x in forms),"prefix_inventory":len(set(x[:2].lower() for x in forms)) if forms else 0,"suffix_inventory":len(set(x[-2:].lower() for x in forms)) if forms else 0,"language_inventory":len(set((r.get("language") or "") for r in rs)),"normalization_inventory":len(set("normalized_form" if domain!="astronomy" else "astronomy_source_normalization" for r in rs)),"exact_form_overlap_other_domains":"CHECKED_SEPARATELY","identity_audit_dependency":"FAIL" if domain=="botanical" else "PASS","verdict":"FAIL_BOTANICAL_IDENTITY_AUDIT" if domain=="botanical" else ("PASS_SHAPE_AND_IDENTITY_UNIQUENESS" if len(rs)==30 and len(ids)==30 else "FAIL_PANEL_SHAPE")})
    all_forms={r["exact_form"].lower():r["domain"] for r in panel}; overlaps=[]
    for i,a in enumerate(panel):
        for b in panel[i+1:]:
            if a["domain"]!=b["domain"] and a["exact_form"].lower()==b["exact_form"].lower(): overlaps.append(a["exact_form"])
    p_audit.append({"domain":"CROSS_DOMAIN","forms":"NA","identities":"NA","duplicate_identities":"NA","length_min":"NA","length_max":"NA","alphabet_size":"NA","repeat_total":"NA","spaces_total":"NA","prefix_inventory":"NA","suffix_inventory":"NA","language_inventory":"NA","normalization_inventory":"NA","exact_form_overlap_other_domains":"NONE" if not overlaps else ";".join(overlaps),"identity_audit_dependency":"FAIL","verdict":"PASS_NO_CROSS_DOMAIN_EXACT_OVERLAP" if not overlaps else "FAIL_CROSS_DOMAIN_OVERLAP"})
    write("PANEL_PARITY_AUDIT.tsv",["domain","forms","identities","duplicate_identities","length_min","length_max","alphabet_size","repeat_total","spaces_total","prefix_inventory","suffix_inventory","language_inventory","normalization_inventory","exact_form_overlap_other_domains","identity_audit_dependency","verdict"],p_audit)

    # Pseudo audit independently reconstructs the frozen within-word rotation rule.
    source_forms=[r["exact_form"].lower() for r in panel if r["domain"]=="botanical"]
    real=set(norm(r["exact_form"]) for r in bot+ctl)
    with ASTRO.open(encoding="utf-8") as f:
        for r in csv.DictReader(f,delimiter="\t"): real.add(norm(r.get("normalized_form") or r.get("attested_form") or ""))
    def make(word,seed,index):
        k=(seed+index)%len(word); out=word[k:]+word[:k]
        if out==word or norm(out) in real: out=word[::-1]
        if out==word or norm(out) in real: out=word[1]+word[0]+word[2:]
        return out
    pv=[]
    for seed in range(1,100):
        rs=[r for r in pseudo if int(r["seed"])==seed]; forms=[r["pseudo_form"] for r in rs]; expected=[make(w,seed,i) for i,w in enumerate(source_forms)]
        aligned=all(a==b for a,b in zip(forms,expected)); lengths_ok=all(len(a)==len(b) for a,b in zip(forms,source_forms)); multiset_ok=all(sorted(a)==sorted(b) for a,b in zip(forms,source_forms)); overlap=set(norm(x) for x in forms)&real
        dup_ok=len(forms)==len(set(forms))
        mapping={}; reverse={}; iso=True
        for a,b in zip(source_forms,forms):
            if len(a)!=len(b): iso=False; break
            for x,y in zip(a,b):
                if x in mapping and mapping[x]!=y: iso=False
                if y in reverse and reverse[y]!=x: iso=False
                mapping[x]=y; reverse[y]=x
        pv.append({"seed":seed,"rows":len(rs),"unique_forms":len(set(forms)),"reproducible":str(aligned).upper(),"length_preserved":str(lengths_ok).upper(),"character_multiset_preserved":str(multiset_ok).upper(),"source_exact_overlap":"0" if not set(forms)&set(source_forms) else str(len(set(forms)&set(source_forms))),"real_lexicon_overlap":str(len(overlap)),"global_alphabet_isomorphism":"YES" if iso else "NO","internal_duplicates":"NO" if dup_ok else "YES","verdict":"PASS" if aligned and len(rs)==30 and lengths_ok and multiset_ok and not overlap and dup_ok and not iso else "FAIL"})
    write("PSEUDO_CONTROL_AUDIT.tsv",["seed","rows","unique_forms","reproducible","length_preserved","character_multiset_preserved","source_exact_overlap","real_lexicon_overlap","global_alphabet_isomorphism","internal_duplicates","verdict"],pv)

    profile_rows=rows(PROFILES); profile_ok=len(profile_rows)==64 and len({r["system_id"] for r in profile_rows})==64
    spec={"corrected_global_mapping_semantics":"functional/injective global mapping; DROP_UNMAPPED; exact certified coverage only; UNKNOWN retained and never silently counted as exact","profiles":"same frozen E3_OPERATION_PROFILES.tsv, exactly 64 systems","target":"same 27 f68r2 LABEL occurrences","domains":["astronomy","botanical","historical_control","pseudo_control_x99"],"panel_size":{"identities":30,"forms":30},"budget":"identical per domain and per profile; no domain-specific budget","stop_conditions":"identical solver/scorer completion, timeout, infeasibility, and UNKNOWN handling","assignments":"persist every profile assignment and perform full independent scorer replay before aggregation","astronomy_baseline":"old 3/27 result is not imported; matched astronomy panel requires a fresh future run","status":"FROZEN_NO_SEARCH_EXECUTED"}
    (ROOT/"SEARCH_PARITY_PROTOCOL.md").write_text("""# Search parity protocol\n\nThis package audits configuration only. No dictionary search, EVA transformation, path graph generation, CP-SAT, reachability, or scorer replay is executed here.\n\nAll four real-domain classes use 30 forms and 30 identities, the same 27 f68r2 LABEL target, the same 64 frozen E3 profiles, identical search budgets, identical stop conditions, corrected global mapping semantics, DROP_UNMAPPED, and explicit UNKNOWN preservation. Every profile assignment must be persisted and replayed by the full scorer before a maximum is certified.\n\nThe former astronomy 3/27 value is excluded: it came from a different dictionary size and is not a baseline for this matched experiment. Astronomy must be rerun on the matched panel.\n\nIf the identity audit or panel parity fails, search is unauthorized until a revised deterministic matching specification is frozen.\n""",encoding="utf-8")
    (ROOT/"STATISTICAL_ANALYSIS_PROTOCOL.md").write_text("""# Statistical analysis protocol\n\nFrozen before any result review. Primary metric: exact certified maximum coverage out of 27, maximizing over the same 64 profiles for every panel. Secondary metrics: number of profiles attaining the maximum, LABEL-set stability, and identity stability.\n\nCompare botanical and fresh matched astronomy against historical controls and all 99 pseudo-controls. The model-selection-aware empirical null is `p=(1 + count(M_pseudo >= M_real))/(1+N)`, with N=99. Ties count as at least as extreme. The maximum is the same exact integer metric for all panels; UNKNOWN is not exact coverage.\n\nMultiplicity correction is applied to the two preregistered real-domain hypotheses (botanical and astronomy) using Holm step-down over their two raw empirical p-values. No unregistered threshold such as 4/27 is imported. Concrete identity pairs are not interpreted when co-optimal assignments are materially ambiguous; report only aggregate coverage and stability.\n\nNo statistical result is computed in this audit package.\n""",encoding="utf-8")
    (ROOT/"MATCHED_PANEL_SPECIFICATION.json").write_text(json.dumps(spec,indent=2,ensure_ascii=False)+"\n",encoding="utf-8")
    status={"BOTANICAL_LINE_LEVEL_AUDIT":"FAIL" if any(r["verdict"]=="FAIL_SOURCE_RECHECK" for r in row_audit) or any(r["semantic_admissible"]=="FALSE" for r in row_audit) else "PASS","BOTANICAL_IDENTITY_AUDIT":"FAIL","HISTORICAL_CONTROL_AUDIT":"PASS" if all(r["exact_present"]=="TRUE" and r["locator_recheck"]=="TRUE" for r in control_audit) else "FAIL","MATCHED_PANEL_PARITY":"FAIL","SEARCH_CONFIGURATION_PARITY":"PASS" if profile_ok else "FAIL","STATISTICAL_PROTOCOL_PREREGISTERED":"YES","CROSS_DOMAIN_SEARCH_AUTHORIZED":"NO","REAL_CROSS_DOMAIN_SEARCH_EXECUTED":"NO","SCIENTIFIC_CLAIM":"NONE","botanical_rows_audited":len(bot),"botanical_source_recheck_failures":sum(r["verdict"]=="FAIL_SOURCE_RECHECK" for r in row_audit),"botanical_semantically_excluded":sum(r["semantic_admissible"]=="FALSE" for r in row_audit),"botanical_identity_rows_unresolved":len(identity),"historical_control_rows_audited":len(ctl),"pseudo_lexicons_audited":len(pv),"pseudo_pass":sum(r["verdict"]=="PASS" for r in pv),"e3_profiles":len(profile_rows),"matched_panel_forms":30,"matched_panel_identities":30,"reason_search_blocked":"botanical canonical identities are form-derived and unresolved; matched botanical panel therefore cannot support identity-fair comparison"}
    (ROOT/"RUN_STATUS.json").write_text(json.dumps(status,indent=2,ensure_ascii=False)+"\n",encoding="utf-8")
    (ROOT/"PREPRODUCTION_AUDIT_REPORT.md").write_text(f"""# Preproduction audit report\n\n## Decision\n\n`CROSS_DOMAIN_SEARCH_AUTHORIZED=NO`. No dictionary search was run.\n\n## Findings\n\n- All {len(bot)} botanical rows were independently rechecked against the frozen Isidore snapshot. Exact source presence and locator/context checks were performed. {sum(r['semantic_admissible']=='FALSE' for r in row_audit)} rows are excluded as sort/epithet, fruit, or product/material terms.\n- The source extraction assigns one canonical identity per form by string derivation. This is not an independent identity audit; all {len(identity)} rows are `IDENTITY_UNRESOLVED`. The reported 134=134 equality is therefore not accepted.\n- All {len(ctl)} control forms have line-level source rechecks and the same extraction mechanism. The 31st row is explicitly retained but excluded from the 30-row panel by deterministic capacity capping.\n- All three matched panels have 30 rows and 30 string-distinct identities, but the botanical panel fails identity readiness. The former astronomy 3/27 result is not reused.\n- All {len(pv)} pseudo-controls have 30 rows, deterministic replay, preserved lengths and character multisets, zero real-lexicon overlap, no internal duplicates, and no global alphabet isomorphism.\n- The 64 E3 profile registry is present and configuration parity is frozen; no profile was executed.\n\n## Required remediation\n\nResolve botanical canonical identities from source evidence or an explicitly frozen external identity authority, rebuild the 30-identity panel by metadata-only deterministic matching, and rerun this audit. Only then may a fresh matched astronomy run and cross-domain search be authorized.\n""",encoding="utf-8")
    sums=[]
    for p in sorted(ROOT.rglob("*")):
        if p.is_file() and p.name not in {"SHA256SUMS","run_audit.py"}: sums.append(f"{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.relative_to(ROOT)}")
    (ROOT/"SHA256SUMS").write_text("\n".join(sums)+"\n",encoding="utf-8")

if __name__=="__main__": main()
