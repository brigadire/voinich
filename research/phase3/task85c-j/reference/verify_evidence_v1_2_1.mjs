#!/usr/bin/env node
import fs from "node:fs"; import path from "node:path"; import {createRequire} from "node:module";
const require=createRequire(import.meta.url), Ajv=require("/tmp/task85ci-node/node_modules/ajv/dist/2020.js").default;
const out=path.resolve(path.dirname(new URL(import.meta.url).pathname),"..");
const read=p=>JSON.parse(fs.readFileSync(p,"utf8"));
const registry=read(path.join(out,"G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json"));
const fixtures=read(path.join(out,"fixtures/G1V2_V1_2_1_EVIDENCE_POSITIVE_FIXTURES.json"));
const ajv=new Ajv({strict:true,strictRequired:false,allErrors:true,validateFormats:false});
for(const e of registry.entries){const s=read(path.join(out,e.schema_path));ajv.addSchema(s,s.$id)}
if(registry.entries.length!==15||new Set(fixtures.map(x=>x.schema)).size!==15)throw new Error("cardinality");
for(const tc of fixtures){const e=registry.entries.find(x=>x.evidence_type===tc.schema);if(!ajv.validate(e.schema_id,tc.instance))throw new Error(`${tc.id}: ${ajv.errorsText()}`);const bad=structuredClone(tc.instance);bad.contract_version="G1_V2_EXECUTABLE_CONTRACT_V1_2";if(ajv.validate(e.schema_id,bad))throw new Error(`mixed accepted ${tc.id}`)}
console.log(`EVIDENCE_V1_2_1=PASS CASES=${fixtures.length} MIXED_FAIL_CLOSED=${fixtures.length}`);
