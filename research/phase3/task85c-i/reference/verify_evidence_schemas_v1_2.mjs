#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";

const require=createRequire(import.meta.url);
const Ajv2020=require("/tmp/task85ci-node/node_modules/ajv/dist/2020.js").default;
const hyper=await import("/tmp/task85ci-node/node_modules/@hyperjump/json-schema/draft-2020-12/index.js");
const out=path.resolve(path.dirname(new URL(import.meta.url).pathname),"..");
const repo=path.resolve(out,"../../..");
const oldRoot=path.join(repo,"research/phase3/task85c-c");
const read=p=>JSON.parse(fs.readFileSync(p,"utf8"));
const files=fs.readdirSync(path.join(out,"evidence-schemas-v1_2")).filter(x=>x.endsWith(".schema.json")).sort();
if(files.length!==15) throw new Error(`schema count ${files.length}`);

const ajv=new Ajv2020({strict:true,strictRequired:false,allErrors:true,validateFormats:false});
const schemas=new Map();
for(const file of files){
  const type=file.replace(".schema.json","");
  const schema=read(path.join(out,"evidence-schemas-v1_2",file));
  if(schema.$schema!=="https://json-schema.org/draft/2020-12/schema") throw new Error(`dialect ${type}`);
  if(JSON.stringify(schema).includes('"x-')) throw new Error(`custom keyword ${type}`);
  ajv.addSchema(schema,schema.$id); hyper.registerSchema(schema,schema.$id); schemas.set(type,schema);
}
const oldSchemas=new Map(files.map(file=>[file.replace(".schema.json",""),read(path.join(oldRoot,"schemas",file))]));
const oldCases=read(path.join(oldRoot,"golden/schema-positive/cases.json"));
const newCases=read(path.join(out,"fixtures/G1V2_V1_2_EVIDENCE_POSITIVE_FIXTURES.json"));
if(oldCases.length!==newCases.length) throw new Error("fixture cardinality");

let positives=0, mutations=0, cross=0;
for(let i=0;i<newCases.length;i++){
  const tc=newCases[i], schema=schemas.get(tc.schema), id=schema.$id;
  const a=ajv.validate(id,tc.instance), h=(await hyper.validate(id,tc.instance)).valid;
  if(!a||!h) throw new Error(`positive ${tc.id}: ${ajv.errorsText()}`);
  const accepted=[];
  for(const [type,s] of schemas) if(ajv.validate(s.$id,tc.instance)) accepted.push(type);
  if(accepted.length!==1||accepted[0]!==tc.schema) throw new Error(`mutual exclusivity ${tc.id} ${accepted}`);
  const old=oldCases[i];
  if(old.schema!==tc.schema) throw new Error("fixture alignment");
  if(!ajv.validate(oldSchemas.get(old.schema),old.instance)) throw new Error(`historical positive ${old.id}`);
  if(ajv.validate(oldSchemas.get(tc.schema),tc.instance)) throw new Error(`V1.2 accepted by V1.1 ${tc.id}`);
  if(ajv.validate(schema,old.instance)) throw new Error(`V1.1 accepted by V1.2 ${old.id}`);
  cross+=2;
  const badVersion=structuredClone(tc.instance); badVersion.contract_version="G1_V2_EXECUTABLE_CONTRACT_UNKNOWN";
  const badStatus=structuredClone(tc.instance); badStatus.status="UNKNOWN_STATUS";
  const badType=structuredClone(tc.instance); badType.schema_id="g1v2.unknown.v1_2";
  const matching=schema.oneOf.find(b=>b.properties.status.const===tc.instance.status&&b.properties.stage.const===tc.instance.stage);
  const badPayload=structuredClone(tc.instance); delete badPayload.payload[matching.properties.payload.required[0]];
  for(const [label,bad] of [["contract_version",badVersion],["status",badStatus],["schema_id",badType],["payload",badPayload]]){
    const x=ajv.validate(id,bad), y=(await hyper.validate(id,bad)).valid;
    if(x||y) throw new Error(`mutation accepted ${tc.id} ${label}`);
    mutations++;
  }
  positives++;
}
console.log("PRIMARY_JSON_SCHEMA_VALIDATOR=Ajv/8.17.1");
console.log("SECONDARY_JSON_SCHEMA_VALIDATOR=Hyperjump/1.17.1");
console.log(`V1_2_POSITIVE_REGRESSION=PASS CASES=${positives}`);
console.log(`CROSS_VERSION_NEGATIVE_REGRESSION=PASS CASES=${cross}`);
console.log(`SCHEMA_MUTATIONS=PASS CASES=${mutations}`);
