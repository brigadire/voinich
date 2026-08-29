#!/usr/bin/env node
import fs from "node:fs"; import path from "node:path"; import {createRequire} from "node:module";
const require=createRequire(import.meta.url), Ajv=require("/tmp/task85ci-node/node_modules/ajv/dist/2020.js").default;
const root=path.resolve(path.dirname(new URL(import.meta.url).pathname),"../../task85c-j"), out=path.resolve(path.dirname(new URL(import.meta.url).pathname),"..");
const read=p=>JSON.parse(fs.readFileSync(p,"utf8")), reg=read(path.join(root,"G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json")), ev=read(path.join(out,"TASK85C_H_HANDLER_EVIDENCE_FIXTURES.json"));
const ajv=new Ajv({strict:true,strictRequired:false,allErrors:true,validateFormats:false}); for(const e of reg.entries){const s=read(path.join(root,e.schema_path));ajv.addSchema(s,s.$id)}
for(const x of ev){const typ=x.schema_id.slice(5,-7), e=reg.entries.find(y=>y.evidence_type===typ); if(!e||!ajv.validate(e.schema_id,x))throw new Error(`${typ}: ${ajv.errorsText()}`); for(const [k,v] of [["contract_version","G1_V2_EXECUTABLE_CONTRACT_V1_2"],["schema_id","g1v2.unknown.v1_2_1"],["status","SCIENTIFIC_FAILURE"]]){const bad=structuredClone(x);bad[k]=v;if(ajv.validate(e.schema_id,bad))throw new Error(`mutation accepted ${typ}/${k}`)}}
if(new Set(ev.map(x=>x.schema_id)).size!==15)throw new Error("coverage"); console.log(`EVIDENCE_V1_2_1=PASS CASES=${ev.length} TYPES=15`);
