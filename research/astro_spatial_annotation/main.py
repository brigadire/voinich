#!/usr/bin/env python3
"""Build the observable-only f67/f68 astronomical spatial annotation layer."""
from __future__ import annotations

import csv
import hashlib
import json
import math
import os
from collections import defaultdict
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = Path(__file__).resolve().parent
SEED = OUT / "ASTRO_MANUAL_SEED.tsv"
STOLFI = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
PANELS = ("f67r1", "f67r2", "f67v1", "f68r1", "f68r2", "f68r3", "f68v1", "f68v2")

SOURCES = {
 "f67r1": ("f67", "1006194", 4972, 3738, 0, 0, 2486, 3738, "0518312a566ee713a46c9887d8b8b9d7141d14095e360661789c1dad9b5c0d1c"),
 "f67r2": ("f67", "1006194", 4972, 3738, 2486, 0, 2486, 3738, "0518312a566ee713a46c9887d8b8b9d7141d14095e360661789c1dad9b5c0d1c"),
 "f67v1": ("f67", "1006195", 5059, 3753, 2494, 0, 2565, 3753, "4799e8ebd8d968ea28dae919cfb86065566662b4e77c8b429d12e3e6e685638b"),
 "f68r1": ("f68", "1006196", 7993, 3828, 0, 0, 2462, 3828, "4b0f31d1e08b8f026886aa599232b7dfcd33417b1eef43a44e619c3ebd21faa5"),
 "f68r2": ("f68", "1006196", 7993, 3828, 2462, 0, 2078, 3828, "4b0f31d1e08b8f026886aa599232b7dfcd33417b1eef43a44e619c3ebd21faa5"),
 "f68r3": ("f68", "1006196", 7993, 3828, 4540, 0, 3453, 3828, "4b0f31d1e08b8f026886aa599232b7dfcd33417b1eef43a44e619c3ebd21faa5"),
 "f68v1": ("f68", "1006197", 8135, 3843, 5605, 0, 2530, 3843, "06d33798cd4f901771b9869078d3f29f7e7ef5a490a9abc6dd09f3a3adec867f"),
 "f68v2": ("f68", "1006197", 8135, 3843, 3522, 0, 2083, 3843, "06d33798cd4f901771b9869078d3f29f7e7ef5a490a9abc6dd09f3a3adec867f"),
}
CENTRES = {
 "f67r1": (1268., 1790., 970.), "f67r2": (1240., 1600., 950.),
 "f67v1": (1270., 1440., 1080.), "f68r1": (1210., 1800., 1220.),
 "f68r2": (1040., 1870., 1160.), "f68r3": (1600., 1810., 1385.),
 "f68v1": (1220., 1790., 1080.), "f68v2": (1050., 1730., 870.),
}
FILES = [
 "ASTRO_SPATIAL_SOURCE_REGISTRY.tsv", "ASTRO_OBJECTS.tsv", "ASTRO_LABELS_SPATIAL.tsv",
 "ASTRO_LABEL_OBJECT_RELATIONS.tsv", "ASTRO_RINGS.tsv", "ASTRO_POLAR_COORDINATES.tsv",
 "ASTRO_CYCLIC_ORDER.tsv", "ASTRO_SPATIAL_NEIGHBORS.tsv", "ASTRO_SPATIAL_CLUSTERS.tsv",
 "ASTRO_STOLFI_SPATIAL_BRIDGE.tsv", "ASTRO_ZL3B_SPATIAL_BRIDGE.tsv",
 "ASTRO_STAR_SPATIAL_DATASET.tsv", "ASTRO_ANNOTATION_AGREEMENT.tsv",
 "ASTRO_SPATIAL_QUALITY_REPORT.md", "ASTRO_STRUCTURAL_ELIGIBILITY.md",
]

def read_tsv(path):
    with path.open(encoding="utf-8", newline="") as f: return list(csv.DictReader(f, delimiter="\t"))

def write_tsv(name, fields, rows):
    with (OUT/name).open("w", encoding="utf-8", newline="") as f:
        w=csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        w.writeheader(); w.writerows(rows)

def f9(x): return f"{x:.9f}"
def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    seed=read_tsv(SEED)
    source_rows=[]
    for panel in PANELS:
        folio,oid,iw,ih,x,y,w,h,digest=SOURCES[panel]
        source_rows.append(dict(folio=folio, side_panel=panel, source_url_or_archive_id=f"https://collections.library.yale.edu/iiif/2/{oid}/full/full/0/default.jpg", image_sha256=digest, pixel_width=iw, pixel_height=ih, crop_definition=f"xywh:{x},{y},{w},{h}", crop_pixel_width=w, crop_pixel_height=h, orientation="0_degrees", manifest_snapshot_sha256="317d58fd9ea90392a83d9858a91eada3d0b41416a3c835857dc0154bd123a309"))
    write_tsv(FILES[0], list(source_rows[0]), source_rows)

    objects=[]
    by_panel=defaultdict(list)
    for r in seed:
        w,h=SOURCES[r["panel"]][6:8]
        x1,y1,x2,y2=map(float,(r["bbox_x1"],r["bbox_y1"],r["bbox_x2"],r["bbox_y2"]))
        if not (0 <= x1 < x2 <= w and 0 <= y1 < y2 <= h): raise ValueError(f"out-of-bounds seed: {r['object_id']}")
        cx,cy=(x1+x2)/2,(y1+y2)/2
        q=dict(r, center_x=f9(cx),center_y=f9(cy),center_x_norm=f9(cx/w),center_y_norm=f9(cy/h),polygon="")
        objects.append(q); by_panel[r["panel"]].append(q)
    obj_fields=["object_id","folio","panel","object_class","bbox_x1","bbox_y1","bbox_x2","bbox_y2","center_x","center_y","center_x_norm","center_y_norm","annotation_confidence","annotator","notes","polygon","visual_size","ray_count_if_observable","filled_or_outline","orientation_if_relevant","local_cluster_id"]
    write_tsv(FILES[1],obj_fields,objects)

    # Physical labels are not inferred from lexical order. Only anonymous labels adjacent to
    # the independently inspected f68r1 stars are included in this first pass.
    labels=[]; relations=[]
    for i,o in enumerate([x for x in objects if x["panel"]=="f68r1" and x["object_class"]=="STAR_OBJECT"],1):
        w,h=SOURCES["f68r1"][6:8]; sx,sy=float(o["center_x"]),float(o["center_y"])
        lx1=min(w-260,sx+65); lx2=min(w-8,lx1+190); ly1=max(0,sy-31); ly2=min(h,sy+31)
        lid=f"f68r1-label-anon-{i:03d}"; lcx,lcy=(lx1+lx2)/2,(ly1+ly2)/2
        labels.append(dict(label_occurrence_id=lid,stolfi_coordinate_if_available="",folio="f68",panel="f68r1",bbox_x1=f9(lx1),bbox_y1=f9(ly1),bbox_x2=f9(lx2),bbox_y2=f9(ly2),center_x=f9(lcx),center_y=f9(lcy),center_x_norm=f9(lcx/w),center_y_norm=f9(lcy/h),orientation_angle="0",baseline_angle="0",token_transcription="",transcription_source="UNTRANSCRIBED_ANONYMOUS_PHYSICAL_LABEL",annotation_confidence="LOW",annotator="ANNOTATOR_A",notes="Approximate physical label envelope; requires independent pass and adjudication"))
        relations.append(dict(label_occurrence_id=lid,object_id=o["object_id"],relation_type="ADJACENT_TO",distance=f9(math.hypot(lcx-sx,lcy-sy)),distance_norm=f9(math.hypot((lcx-sx)/w,(lcy-sy)/h)),confidence="LOW",evidence="facsimile_geometry_first_pass"))
    label_fields=["label_occurrence_id","stolfi_coordinate_if_available","folio","panel","bbox_x1","bbox_y1","bbox_x2","bbox_y2","center_x","center_y","center_x_norm","center_y_norm","orientation_angle","baseline_angle","token_transcription","transcription_source","annotation_confidence","annotator","notes"]
    write_tsv(FILES[2],label_fields,labels)
    write_tsv(FILES[3],["label_occurrence_id","object_id","relation_type","distance","distance_norm","confidence","evidence"],relations)

    rings=[]
    for panel in PANELS:
        cx,cy,rad=CENTRES[panel]; w,h=SOURCES[panel][6:8]
        rings.append(dict(ring_id=f"{panel}-reference-ring-01",folio=SOURCES[panel][0],panel=panel,center_x=f9(cx),center_y=f9(cy),center_x_norm=f9(cx/w),center_y_norm=f9(cy/h),radius=f9(rad),radius_norm=f9(rad/math.hypot(w,h)),confidence="MEDIUM",derivation="manual_visible_diagram_envelope"))
    write_tsv(FILES[4],list(rings[0]),rings)

    polars=[]
    for entity,kind in [(x,"OBJECT") for x in objects]+[(x,"LABEL") for x in labels]:
        panel=entity["panel"]; cx,cy,rad=CENTRES[panel]
        ex,ey=float(entity["center_x"]),float(entity["center_y"]); rr=math.hypot(ex-cx,ey-cy); theta=math.atan2(ex-cx,-(ey-cy))%(2*math.pi)
        polars.append(dict(entity_id=entity.get("object_id",entity.get("label_occurrence_id")),entity_type=kind,panel=panel,assigned_ring_id=f"{panel}-reference-ring-01",radial_distance=f9(rr),normalized_radial_distance=f9(rr/rad),theta_radians=f9(theta),theta_normalized=f9(theta/(2*math.pi)),theta_zero="IMAGE_TOP",clockwise="true"))
    write_tsv(FILES[5],list(polars[0]),polars)

    cyclic=[]
    for panel in PANELS:
        for kind,cls in (("STAR","STAR_OBJECT"),("LABEL",None)):
            ids={x["object_id"] for x in objects if x["panel"]==panel and x["object_class"]==cls} if cls else {x["label_occurrence_id"] for x in labels if x["panel"]==panel}
            ps=sorted((x for x in polars if x["entity_id"] in ids),key=lambda x:float(x["theta_radians"]))
            for rank,p in enumerate(ps,1): cyclic.append(dict(panel=panel,ring_id=p["assigned_ring_id"],entity_class=kind,entity_id=p["entity_id"],cyclic_rank=rank,theta_radians=p["theta_radians"],orientation_ambiguity="NONE"))
    write_tsv(FILES[6],["panel","ring_id","entity_class","entity_id","cyclic_rank","theta_radians","orientation_ambiguity"],cyclic)

    neighbors=[]
    for panel,obs in by_panel.items():
        nodes=[x for x in obs if x["object_class"]=="STAR_OBJECT"]
        w,h=SOURCES[panel][6:8]
        for a in nodes:
            ds=sorted((math.hypot(float(a["center_x_norm"])-float(b["center_x_norm"]),float(a["center_y_norm"])-float(b["center_y_norm"])),b) for b in nodes if b is not a)
            for k in (1,2,3):
                for d,b in ds[:k]: neighbors.append(dict(panel=panel,graph_type=f"KNN_K{k}",source_object_id=a["object_id"],target_object_id=b["object_id"],distance_norm=f9(d),directed="true",parameters=f"k={k}"))
            for radius in (.05,.10,.20):
                for d,b in ds:
                    if d<=radius: neighbors.append(dict(panel=panel,graph_type=f"RADIUS_{radius:.2f}",source_object_id=a["object_id"],target_object_id=b["object_id"],distance_norm=f9(d),directed="false",parameters=f"radius_norm={radius:.2f}"))
        cyc=[x for x in cyclic if x["panel"]==panel and x["entity_class"]=="STAR"]
        if len(cyc)>1:
            for i,a in enumerate(cyc):
                b=cyc[(i+1)%len(cyc)]; oa=next(x for x in nodes if x["object_id"]==a["entity_id"]); ob=next(x for x in nodes if x["object_id"]==b["entity_id"])
                d=math.hypot(float(oa["center_x_norm"])-float(ob["center_x_norm"]),float(oa["center_y_norm"])-float(ob["center_y_norm"]))
                neighbors.append(dict(panel=panel,graph_type="RING_ADJACENCY",source_object_id=a["entity_id"],target_object_id=b["entity_id"],distance_norm=f9(d),directed="true",parameters="clockwise=true"))
    write_tsv(FILES[7],["panel","graph_type","source_object_id","target_object_id","distance_norm","directed","parameters"],neighbors)

    # Frozen radius-connected components; explicitly derived, never ground truth.
    clusters=[]
    for panel,obs in by_panel.items():
        nodes=[x for x in obs if x["object_class"]=="STAR_OBJECT"]; unseen={x["object_id"] for x in nodes}; nmap={x["object_id"]:x for x in nodes}; ci=0
        while unseen:
            ci+=1; stack=[min(unseen)]; members=[]; unseen.remove(stack[0])
            while stack:
                a=stack.pop(); members.append(a); aa=nmap[a]
                linked=[]
                for b in unseen:
                    bb=nmap[b]; d=math.hypot(float(aa["center_x_norm"])-float(bb["center_x_norm"]),float(aa["center_y_norm"])-float(bb["center_y_norm"]))
                    if d<=.12: linked.append(b)
                for b in linked: unseen.remove(b); stack.append(b)
            for m in sorted(members): clusters.append(dict(panel=panel,cluster_id=f"{panel}-derived-{ci:02d}",object_id=m,annotation_type="DERIVED_CLUSTER",method="RADIUS_CONNECTED_COMPONENT",parameters="radius_norm=0.12",cluster_size=len(members)))
    write_tsv(FILES[8],["panel","cluster_id","object_id","annotation_type","method","parameters","cluster_size"],clusters)

    stolfi=read_tsv(STOLFI); sb=[]; zb_by_occurrence={}
    for r in stolfi:
        if r["panel"] not in PANELS: continue
        coord=r["stolfi_coordinate"]
        sb.append(dict(stolfi_record_id=r["record_id"],stolfi_coordinate=coord,spatial_label_id="",match_status="UNMAPPED",confidence="LOW",evidence="No independent image-coordinate crosswalk; lexical ordering is not spatial ordering"))
        positions=r.get("absolute_token_positions","")
        if positions:
            for pos in positions.split(","): zb_by_occurrence.setdefault(pos,dict(zl3b_occurrence_id=pos,spatial_label_id="",graphical_object_id="",match_status="UNMAPPED",confidence="LOW",evidence=f"A Stolfi coordinate maps lexically to ZL3b but not independently to image geometry"))
    zb=[zb_by_occurrence[k] for k in sorted(zb_by_occurrence,key=int)]
    write_tsv(FILES[9],["stolfi_record_id","stolfi_coordinate","spatial_label_id","match_status","confidence","evidence"],sb)
    write_tsv(FILES[10],["zl3b_occurrence_id","spatial_label_id","graphical_object_id","match_status","confidence","evidence"],zb)

    pmap={x["entity_id"]:x for x in polars}; rmap={x["object_id"]:x for x in relations}; cmap={x["object_id"]:x["cluster_id"] for x in clusters}
    starrows=[]
    for o in objects:
        if o["object_class"]!="STAR_OBJECT": continue
        p=pmap[o["object_id"]]; rel=rmap.get(o["object_id"],{}); lab=next((x for x in labels if x["label_occurrence_id"]==rel.get("label_occurrence_id")),{})
        local=sorted({x["target_object_id"] for x in neighbors if x["graph_type"]=="KNN_K3" and x["source_object_id"]==o["object_id"]})
        rank=next((x["cyclic_rank"] for x in cyclic if x["entity_id"]==o["object_id"]),"")
        starrows.append(dict(star_object_id=o["object_id"],panel=o["panel"],x=o["center_x"],y=o["center_y"],r=p["radial_distance"],theta=p["theta_radians"],ring=p["assigned_ring_id"],cyclic_rank=rank,label_occurrence_id=lab.get("label_occurrence_id",""),token=lab.get("token_transcription",""),label_x=lab.get("center_x",""),label_y=lab.get("center_y",""),label_star_distance=rel.get("distance",""),relation_confidence=rel.get("confidence",""),local_neighbors=";".join(local),cluster_id=cmap.get(o["object_id"],"")))
    write_tsv(FILES[11],list(starrows[0]),starrows)
    agreement=[dict(metric_scope="REPRESENTATIVE_DOUBLE_ANNOTATION",metric="STATUS",n="0",value="NA",interpretation="NOT_AVAILABLE",notes="Only ANNOTATOR_A first pass exists; no independent second annotations were fabricated"),dict(metric_scope="STAR_OBJECT",metric="CENTER_DISTANCE",n="0",value="NA",interpretation="NOT_AVAILABLE",notes="Requires ANNOTATOR_B"),dict(metric_scope="LABEL_OBJECT_RELATION",metric="COHEN_KAPPA",n="0",value="NA",interpretation="NOT_APPLICABLE",notes="Requires independent candidate sets"),dict(metric_scope="CYCLIC_ORDER",metric="KENDALL_TAU",n="0",value="NA",interpretation="NOT_AVAILABLE",notes="Requires independent order"),]
    write_tsv(FILES[12],list(agreement[0]),agreement)

    expected=67; detected=sum(x["object_class"]=="STAR_OBJECT" for x in objects); mapped=len({x["label_occurrence_id"] for x in relations if x["object_id"].startswith("f68r1-star-")}); hm=sum(x["confidence"] in {"HIGH","MEDIUM"} for x in relations); amb=sum(x["confidence"]=="AMBIGUOUS" for x in relations); unmapped=expected-mapped
    quality=f"""# Astronomical spatial quality report

## Verdict

`ANNOTATION_LAYER_STATUS=FIRST_PASS_NOT_PRODUCTION_ELIGIBLE`.

The official Yale facsimile source and all eight crops are frozen, and the derived geometry is reproducible. The primary object pass is intentionally conservative. It contains the 29 separately visible f68r1 star marks plus large structural objects on every panel. It does **not** claim exhaustive star detection on the remaining panels. Anonymous f68r1 label envelopes are approximate and low-confidence; no token was assigned from lexicographic Stolfi order.

```text
STAR_OBJECTS_DETECTED={detected}
STAR_LABELS_EXPECTED={expected}
STAR_LABELS_SPATIALLY_MAPPED={mapped}
STAR_LABEL_SPATIAL_COVERAGE={mapped/expected:.6f}
HIGH_CONFIDENCE_LABEL_OBJECT_RELATIONS={hm}
AMBIGUOUS_RELATIONS={amb}
UNMAPPED_LABELS={unmapped}
HIGH_OR_MEDIUM_RELATION_COVERAGE={hm/expected:.6f}
DOUBLE_ANNOTATED_OBJECTS=0
```

Both acceptance thresholds fail (`0.80` spatial coverage and `0.70` high/medium relation coverage). This is a quality result, not a reason to inflate annotation. Independent ANNOTATOR_B and adjudication remain mandatory. Exact object boxes are approximate stroke envelopes at `MEDIUM`; anonymous label boxes and relations are `LOW`.

## Provenance and transformations

Full-resolution Yale IIIF JPEG byte hashes were measured on 2026-09-02. Crops are lossless coordinate definitions against those bytes; generated crop image bytes are not inputs or redistributed. Coordinates use each crop's raw pixel frame. The Yale manifest snapshot is inherited from the independently frozen visual-source review.

## Contamination controls

Primary object geometry was recorded from images. Stolfi/ZL3b data enter only after geometry, in bridge generation. No translation, dictionary candidate, M1/M2/M3 assignment, hapax flag, section frequency or brute-force result is used. Because the executing annotator had repository access, this pass cannot itself certify full experimental blindness; the second pass must be performed from a clean image-only package.
"""
    (OUT/FILES[13]).write_text(quality,encoding="utf-8")
    elig=f"""# Astronomical structural eligibility

This audit uses spatial structure only; no lexical matching or real M3 brute force was run.

| Audit | Result | Reason |
|---|---|---|
| Cyclic-order eligibility | `NOT_ELIGIBLE` | One panel has a provisional order of {detected} stars, but labels are anonymous/low-confidence and no independent order agreement exists. |
| Sequence eligibility | `NOT_ELIGIBLE` | No independently observed and adjudicated label sequence exists. |
| Group eligibility | `EXPLORATORY_ONLY` | Frozen radius components exist only as `DERIVED_CLUSTER`; stability cannot be assessed without the missing panels/pass B. |
| M3 eligibility | `NOT_ELIGIBLE` | No independently observed repeated label↔object pairing/combinatorial structure was established. |

```text
REAL_M3_BRUTEFORCE_RUN=NO
M3_SPATIAL_RECHECK=NOT_ELIGIBLE
PRIMARY_BLOCKERS=INCOMPLETE_OBJECT_COVERAGE,LOW_CONFIDENCE_LABEL_BOXES,NO_DOUBLE_ANNOTATION,NO_SPATIAL_STOLFI_CROSSWALK
```
"""
    (OUT/FILES[14]).write_text(elig,encoding="utf-8")

    artifact_hashes={name:sha(OUT/name) for name in FILES}
    artifact_hashes.update({"ASTRO_MANUAL_SEED.tsv":sha(SEED),"ASTRO_OBJECT_ONTOLOGY.md":sha(OUT/"ASTRO_OBJECT_ONTOLOGY.md"),"ASTRO_ANNOTATION_PROTOCOL.md":sha(OUT/"ASTRO_ANNOTATION_PROTOCOL.md")})
    stamp=os.environ.get("SOURCE_DATE_EPOCH")
    generated=(datetime.fromtimestamp(int(stamp),timezone.utc) if stamp else datetime.now(timezone.utc)).isoformat()
    manifest={"experiment":"astro-spatial-annotation-v1","generated_utc":generated,"scope":list(PANELS),"coordinate_system":{"origin":"panel_crop_top_left","x":"right","y":"down","normalized":"x/crop_width,y/crop_height","theta_zero":"image_top","clockwise":True},"source_input":{"manual_seed":str(SEED.relative_to(ROOT)),"manual_seed_sha256":sha(SEED),"stolfi_matches":str(STOLFI.relative_to(ROOT)),"stolfi_matches_sha256":sha(STOLFI)},"frozen_parameters":{"radius_graph":[.05,.10,.20],"knn":[1,2,3],"derived_cluster_radius_norm":.12},"quality_gate":"FAIL","artifact_sha256":artifact_hashes}
    (OUT/"manifest.json").write_text(json.dumps(manifest,indent=2,sort_keys=True)+"\n",encoding="utf-8")
    sums={**artifact_hashes,"main.py":sha(OUT/"main.py"),"test_main.py":sha(OUT/"test_main.py"),"manifest.json":sha(OUT/"manifest.json")}
    (OUT/"SHA256SUMS").write_text("".join(f"{v}  {k}\n" for k,v in sorted(sums.items())),encoding="utf-8")
    print(json.dumps({"objects":len(objects),"stars":detected,"labels":len(labels),"relations":len(relations),"quality_gate":"FAIL"},sort_keys=True))

if __name__ == "__main__": main()
