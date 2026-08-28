#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { pathToFileURL } from "node:url";

const require=createRequire(import.meta.url);
const Ajv2020=require("/tmp/task85cc-node/node_modules/ajv/dist/2020.js").default;
const hyper=await import("/tmp/task85cc-node/node_modules/@hyperjump/json-schema/draft-2020-12/index.js");
const root=path.resolve(path.dirname(new URL(import.meta.url).pathname),"..");
const read=p=>JSON.parse(fs.readFileSync(path.join(root,p),"utf8"));
const schemaDir=path.join(root,"schemas");
const schemaFiles=fs.readdirSync(schemaDir).filter(x=>x.endsWith(".schema.json")).sort();
if(schemaFiles.length!==15) throw new Error(`schema count ${schemaFiles.length}`);

const ajv=new Ajv2020({strict:true,strictRequired:false,allErrors:true,validateFormats:false});
const schemas=new Map();
for(const file of schemaFiles){
  const schema=read(`schemas/${file}`);
  if(schema.$schema!=="https://json-schema.org/draft/2020-12/schema") throw new Error(`dialect ${file}`);
  const encoded=JSON.stringify(schema);
  if(encoded.includes('"x-')) throw new Error(`custom normative keyword ${file}`);
  ajv.addSchema(schema,schema.$id);
  hyper.registerSchema(schema,schema.$id);
  schemas.set(file.replace(".schema.json",""),schema);
}

const groups=["golden/schema-positive/cases.json","golden/schema-negative/cases.json","golden/field-mutations/cases.json","golden/regression/cases.json"];
let total=0;
for(const group of groups){
  const cases=read(group);
  for(const tc of cases){
    total++;
    const id=schemas.get(tc.schema).$id;
    const a=ajv.validate(id,tc.instance);
    const h=(await hyper.validate(id,tc.instance)).valid;
    if(a!==tc.expected) throw new Error(`Ajv ${tc.id}: got ${a} expected ${tc.expected}: ${ajv.errorsText()}`);
    if(h!==tc.expected) throw new Error(`Hyperjump ${tc.id}: got ${h} expected ${tc.expected}`);
    if(a!==h) throw new Error(`validator differential ${tc.id}`);
  }
}

// Matrix completeness and schema status branches.
const matrix=fs.readFileSync(path.join(root,"G1V2_EVIDENCE_STATUS_MATRIX_V1_1.tsv"),"utf8").trimEnd().split("\n").slice(1);
if(matrix.length!==15*13) throw new Error(`matrix ${matrix.length}`);
const statuses=new Set(matrix.map(x=>x.split("\t")[1]));
if(statuses.size!==13||statuses.has("SCIENTIFIC_FAILURE")) throw new Error("status domain");

// Exhaustive job producer × status closure, independently from fixture labels.
const stageRows=fs.readFileSync(path.join(root,"registries/G1V2_STAGE_STATUS_CONTRACT.tsv"),"utf8").trimEnd().split("\n").slice(1).map(x=>x.split("\t"));
if(stageRows.length!==7*13) throw new Error(`stage matrix ${stageRows.length}`);
const represented=new Set();
for(const schema of schemas.values()) for(const b of schema.oneOf){
  const scope=b.properties.producer_scope.const;
  if(scope==="JOB"||scope==="DAG_MATERIALIZER") represented.add(`${b.properties.stage.const}\t${b.properties.status.const}`);
}
for(const r of stageRows){
  const key=`${r[0]}\t${r[1]}`;
  if((r[2]==="YES")!==represented.has(key)) throw new Error(`producer/status mismatch ${key} expected ${r[2]}`);
}

console.log("PRIMARY_JSON_SCHEMA_VALIDATOR=Ajv/8.17.1");
console.log("SECONDARY_JSON_SCHEMA_VALIDATOR=Hyperjump/1.17.1");
console.log(`CROSS_VALIDATOR_CONFORMANCE=PASS CASES=${total}`);
