#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
const out=path.resolve(path.dirname(new URL(import.meta.url).pathname),"..");
const f=JSON.parse(fs.readFileSync(path.join(out,"fixtures/G1V2_E2_JOBID_FIXTURE.json"),"utf8"));
const canonical=x=>Buffer.from(JSON.stringify(Object.fromEntries(Object.keys(x).sort().map(k=>[k,x[k]])))+"\n","utf8");
const jobid=x=>"j-"+crypto.createHash("sha256").update(Buffer.concat([Buffer.from("G1V2-JOB\0"),canonical(x)])).digest("hex").slice(0,40);
for(const version of ["v1_1","v1_2"]){
  if(canonical(f[version].payload).toString("hex")!==f[version].canonical_payload_hex) throw new Error(`canonical ${version}`);
  if(jobid(f[version].payload)!==f[version].jobid) throw new Error(`jobid ${version}`);
}
if(f.v1_1.jobid===f.v1_2.jobid) throw new Error("cross-version collision");
console.log("JOBID_E2_JAVASCRIPT=PASS");
console.log("CROSS_IMPLEMENTATION_BYTE_IDENTITY=PASS");
