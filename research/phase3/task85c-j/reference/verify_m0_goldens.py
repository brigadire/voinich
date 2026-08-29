#!/usr/bin/env python3
import hashlib, json, math
from pathlib import Path

out=Path(__file__).resolve().parent.parent
s=json.loads((out/"G1V2_SCIENTIFIC_GOLDEN_SUITE_V1_2_1.json").read_text())
fit=next(x for x in s["cases"] if x["id"]=="M0-FIT")["expected"]
unseen=next(x for x in s["cases"] if x["id"]=="M0-UNSEEN")["expected"]
assert fit["outcomes"]==["a","b","<UNK>","<EOS>"] and fit["denominator"]=="9"
assert math.isclose(float(fit["p_unk"]),1/9,rel_tol=0,abs_tol=1e-17)
assert unseen["alpha1_probability"]==fit["p_unk"]
assert s["historical_negative_regressions"][0]["expected_disposition"]=="REJECT_UNDER_V1_2_1"
diff=(out/"G1V2_M0_GOLDEN_DIFF.tsv").read_text().splitlines()
assert any("M0-FIT" in x and "DEFECTIVE_AND_REPLACED" in x for x in diff)
assert any("M0-UNSEEN" in x and "DEFECTIVE_AND_REPLACED" in x for x in diff)
root=json.loads((out/"G1V2_SCIENTIFIC_GOLDEN_ROOT_V1_2_1.json").read_text())
encoded=(json.dumps(root["items"],ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()
assert hashlib.sha256(encoded).hexdigest()==root["root_sha256"]
for item in root["items"]: assert hashlib.sha256((out/item["path"]).read_bytes()).hexdigest()==item["sha256"]
print("M0_GOLDENS=PASS")
