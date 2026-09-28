#!/usr/bin/env python3
"""Content-blind preparation of a provenance-audited botanical/control corpus.

This script never reads LABEL occurrences, EVA tokens, paths, or coverage results.
"""
import csv, hashlib, json, random, string
from pathlib import Path
from collections import Counter

ROOT=Path(__file__).resolve().parent
BASE=ROOT.parent
ASTRO=BASE/'exploratory_astronomical_dictionary_search_v1'/'ASTRONOMICAL_NAMES_LEXICON.tsv'
PROFILES=BASE/'f68r2_corrected_e3_resumable_run_v1'/'E3_OPERATION_PROFILES.tsv'
SOURCE='https://penelope.uchicago.edu/Thayer/L/Roman/Texts/Isidore/17%2A.html'
DIOSC='https://www.dioscorides.org/about/'

BOT=[
('PLANT_ABELLANA','abellana','LATIN','Isidore Etymologiae XVII'),('PLANT_AMYGDALA','amygdala','LATIN','Isidore Etymologiae XVII'),('PLANT_ANETHUM','anethum','LATIN','Isidore Etymologiae XVII'),('PLANT_ANTHEMIS','anthemis','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_APIASTRUM','apiastrum','LATIN','Isidore Etymologiae XVII'),('PLANT_APIUM','apium','LATIN','Isidore Etymologiae XVII'),('PLANT_ARTEMISIA','artemisia','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_ASPHODELUS','asphodelus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_AVENA','avena','LATIN','Isidore Etymologiae XVII'),('PLANT_BASILICUM','basilicum','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_BETONICA','betonica','LATIN','Isidore Etymologiae XVII'),('PLANT_BRASSICA','brassica','LATIN','Isidore Etymologiae XVII'),('PLANT_CALAMINTHA','calamintha','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CALENDULA','calendula','LATIN','Isidore Etymologiae XVII'),('PLANT_CANNABIS','cannabis','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CAPPARIS','capparis','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CARDUUS','carduus','LATIN','Isidore Etymologiae XVII'),('PLANT_CARYOPHYLLUS','caryophyllus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CASSIA','cassia','ARABIC_LATIN','Dioscorides De materia medica'),('PLANT_CEDRUS','cedrus','LATIN','Isidore Etymologiae XVII'),('PLANT_CENTAURIUM','centaurium','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CERASUS','cerasus','LATIN','Isidore Etymologiae XVII'),('PLANT_CICHORIUM','cichorium','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CINNAMOMUM','cinnamomum','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CISTUS','cistus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_COLOCASIA','colocasia','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CORIANDRUM','coriandrum','LATIN','Isidore Etymologiae XVII'),('PLANT_CROCUS','crocus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_CUCUMIS','cucumis','LATIN','Isidore Etymologiae XVII'),('PLANT_CUMINUM','cuminum','LATIN','Isidore Etymologiae XVII'),('PLANT_CYPERUS','cyperus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_DAUCUS','daucus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_ELEBORUS','eleborus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_ENDIVIA','endivia','LATIN','Isidore Etymologiae XVII'),('PLANT_FOENICULUM','feniculum','LATIN','Isidore Etymologiae XVII'),('PLANT_GALBANUM','galbanum','ARABIC_LATIN','Dioscorides De materia medica'),('PLANT_GENTIANA','gentiana','LATIN','Isidore Etymologiae XVII'),('PLANT_HYOSCYAMUS','hyoscyamus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_HYSSOPUS','hyssopus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_IRIS','iris','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_LAUREL','laurus','LATIN','Isidore Etymologiae XVII'),('PLANT_LAVANDULA','lavandula','LATIN','Isidore Etymologiae XVII'),('PLANT_LENTICULA','lenticularis','LATIN','Isidore Etymologiae XVII'),('PLANT_LIGUSTICUM','ligusticum','LATIN','Dioscorides De materia medica'),('PLANT_LILIUM','lilium','LATIN','Isidore Etymologiae XVII'),('PLANT_MALVA','malva','LATIN','Isidore Etymologiae XVII'),('PLANT_MANDRAGORA','mandragora','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_MARUBIUM','marubium','LATIN','Isidore Etymologiae XVII'),('PLANT_MELISSA','melissa','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_MENTHA','mentha','LATIN','Isidore Etymologiae XVII'),('PLANT_MILLEFOLIUM','millefolium','LATIN','Isidore Etymologiae XVII'),('PLANT_MORUS','morus','LATIN','Isidore Etymologiae XVII'),('PLANT_NARCISSUS','narcissus','GREEK_LATIN','Isidore Etymologiae XVII'),('PLANT_NIGELLA','nigella','LATIN','Dioscorides De materia medica'),('PLANT_OLIVA','oliva','LATIN','Isidore Etymologiae XVII'),('PLANT_ORIGANUM','origanum','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_PAPAVER','papaver','LATIN','Isidore Etymologiae XVII'),('PLANT_PASTINACA','pastinaca','LATIN','Isidore Etymologiae XVII'),('PLANT_PENTAPHYLLON','pentaphyllon','GREEK_LATIN','Isidore Etymologiae XVII'),('PLANT_PETROSELINUM','petroselinum','LATIN','Isidore Etymologiae XVII'),('PLANT_PLANTAGO','plantago','LATIN','Isidore Etymologiae XVII'),('PLANT_PORRUM','porrum','LATIN','Isidore Etymologiae XVII'),('PLANT_RAPHANUS','raphanus','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_ROSA','rosa','LATIN','Isidore Etymologiae XVII'),('PLANT_RUTA','ruta','LATIN','Isidore Etymologiae XVII'),('PLANT_SALVIA','salvia','LATIN','Isidore Etymologiae XVII'),('PLANT_SAMBUCUS','sambucus','LATIN','Isidore Etymologiae XVII'),('PLANT_SATUREIA','satureia','LATIN','Isidore Etymologiae XVII'),('PLANT_SENECIO','senecio','LATIN','Isidore Etymologiae XVII'),('PLANT_SISYMBRIUM','sisymbrium','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_STAPHISAGRIA','staphisagria','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_TAMARIX','tamarix','LATIN','Isidore Etymologiae XVII'),('PLANT_THYMUM','thymum','GREEK_LATIN','Dioscorides De materia medica'),('PLANT_TRIFOLIUM','trifolium','LATIN','Isidore Etymologiae XVII'),('PLANT_VERBENA','verbena','LATIN','Isidore Etymologiae XVII'),('PLANT_VIOLA','viola','LATIN','Isidore Etymologiae XVII'),('PLANT_VITIS','vitis','LATIN','Isidore Etymologiae XVII'),
]
REJECT=[('modern_linnaean','Solanum tuberosum','modern binomial; postdates target period'),('modern_linnaean','Mentha × piperita','modern hybrid designation'),('late_transliteration','aloe vera','modernized scientific form; source form not frozen'),('reconstructed','*herba_example','editorial reconstruction, not attested'),('ambiguous_identity','papaver','identity requires botanical adjudication in context')]
CONTROL=['ars','arithmetica','astrologia','grammatica','geometria','musica','rhetorica','dialectica','philosophia','medicina','astronomia','instrumentum','astrolabium','calamus','charta','codex','figura','numerus','pagina','scriptura','theoria','canon','magisterium','speculum','vas','unguentum','pulvis','aqua','oleum','vinum']
def row(i,ident,form,lang,source):
 return {'lexicon_id':f'BOT_{i:04d}','canonical_plant_identity':ident,'historical_form':form,'language':lang,'attestation_date':'antiquity–7th c. source tradition','source':source,'source_type':'SECONDARY_EDITION_OR_CRITICAL_SOURCE','confidence':'MEDIUM_PENDING_LINE_AUDIT','normalized_form':form.lower().replace(' ','') ,'historically_attested':'CANDIDATE','modern_or_anachronistic':'NO','reconstructed':'NO','provenance_status':'PROVISIONAL_LINE_LEVEL_AUDIT_REQUIRED','frozen_preprocessing_applicability':'TO_BE_EVALUATED_WITHOUT_TARGET','source_reference':SOURCE if 'Isidore' in source else DIOSC}
def stats(rows):
 forms=[r['normalized_form'] for r in rows]; chars=Counter(''.join(forms)); lens=Counter(map(len,forms)); ids=Counter(r['canonical_plant_identity'] for r in rows)
 return len(rows),len(ids),sum(chars.values()),len(chars),','.join(f'{k}:{v}' for k,v in sorted(lens.items()))
def main():
 headers=list(row(1,*BOT[0][0:1],*BOT[0][1:2],*BOT[0][2:3],*BOT[0][3:4]).keys())
 bot=[row(i+1,*x) for i,x in enumerate(BOT)]
 with (ROOT/'BOTANICAL_LEXICON_FULL.tsv').open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=headers,delimiter='\t');w.writeheader();w.writerows(bot)
 write_rej=['reject_id','category','form','reason','disposition'];
 with (ROOT/'BOTANICAL_LEXICON_REJECTED.tsv').open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=write_rej,delimiter='\t');w.writeheader();w.writerows({'reject_id':f'REJ_{i:03d}','category':a,'form':b,'reason':c,'disposition':'REJECTED_FROM_PRIMARY_CORPUS'} for i,(a,b,c) in enumerate(REJECT))
 audit=[{'lexicon_id':r['lexicon_id'],'form':r['historical_form'],'source':r['source'],'source_reference':r['source_reference'],'source_type':r['source_type'],'audit_status':'PENDING_LINE_LEVEL_VERIFICATION','historical_cutoff':'before 1500 not yet individually demonstrated'} for r in bot]
 with (ROOT/'BOTANICAL_SOURCE_AUDIT.tsv').open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=list(audit[0]),delimiter='\t');w.writeheader();w.writerows(audit)
 panels=[]
 for seed in (101,202,303):
  ids=list(range(len(bot)));random.Random(seed).shuffle(ids)
  panels.append({'panel_id':f'BOT_P{seed}','seed':seed,'source_corpus':'BOTANICAL_LEXICON_FULL.tsv','requested_forms':299,'actual_forms':len(bot),'canonical_identities':len(set(r['canonical_plant_identity'] for r in bot)),'status':'NOT_READY_SIZE_AND_LINE_AUDIT','selection':'deterministic permutation; no target access'})
 with (ROOT/'BOTANICAL_PANELS.tsv').open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=list(panels[0]),delimiter='\t');w.writeheader();w.writerows(panels)
 hc=[{'control_id':f'HCTRL_{i:03d}','form':x,'normalized_form':x,'domain':'historical_non_domain','source':'Isidore Etymologiae or medieval technical vocabulary','source_reference':SOURCE,'status':'PROVISIONAL_SOURCE_AUDIT_REQUIRED'} for i,x in enumerate(CONTROL)]
 with (ROOT/'HISTORICAL_CONTROL_LEXICONS.tsv').open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=list(hc[0]),delimiter='\t');w.writeheader();w.writerows(hc)
 pseudo=[]
 for i,r in enumerate(bot):
  s=r['normalized_form']; out=''.join(chr(97+(ord(c)-97+((i%5)+1))%26) for c in s)
  pseudo.append({'pseudo_id':f'PSEUDO_{i:04d}','source_lexicon_id':r['lexicon_id'],'pseudo_form':out,'length':len(out),'construction':'deterministic Caesar shift mod 26; preserves length, destroys spelling/identity','target_blind':'YES'})
 with (ROOT/'PSEUDO_LEXICON_CONTROLS.tsv').open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=list(pseudo[0]),delimiter='\t');w.writeheader();w.writerows(pseudo)
 a=stats(bot); (ROOT/'LEXICON_MATCHING_STATISTICS.tsv').write_text('dataset\tforms\tidentities\ttotal_characters\tunique_characters\tlength_distribution\n'+'botanical_seed\t'+'\t'.join(map(str,a))+'\n'+'astronomical_reference\t299\t94\tNOT_USED\tNOT_USED\tNOT_USED\n')
 frozen={'target_access':'FORBIDDEN','target_files_loaded':[],'profile_sha256':hashlib.sha256(PROFILES.read_bytes()).hexdigest(),'astronomical_baseline':'3/27','operation_profiles':'64','real_search_executed':False,'preparation_seed':101}
 (ROOT/'INPUT_FREEZE.json').write_text(json.dumps(frozen,indent=2)+'\n')
 status={'ASTRONOMICAL_BASELINE':'3/27','BOTANICAL_FULL_LEXICON_READY':'NO_SOURCE_AUDIT_INCOMPLETE','BOTANICAL_HISTORICAL_AUDIT':'PARTIAL','MATCHED_BOTANICAL_PANELS_READY':'NO_SIZE_MISMATCH','HISTORICAL_CONTROLS_READY':'PARTIAL','PSEUDO_LEXICON_CONTROLS_READY':'YES','TARGET_BLIND_PREPARATION':'YES','MULTIPLICITY_PROTOCOL_FROZEN':'YES','REAL_CROSS_DOMAIN_SEARCH_EXECUTED':'NO','CROSS_DOMAIN_SEARCH_READY':'NO','SCIENTIFIC_CLAIM':'NONE'}
 (ROOT/'PREPARATION_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
