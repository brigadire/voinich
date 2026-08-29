# Expanded contract defect D01-BLIND-ID-CLOSURE

## Exact contradiction

The frozen DAG payload makes `control_instance_id` a required JobID identity field. Sections 36–41 and 79 of Task85c-d require the exact complete set of 1,321,152 JobIDs, its root, and all 2,617,152 resolved edges.

The parent closure supplies counts for 144 blind instances but no 144 opaque blind IDs. Its escrow contract defines each blind ID as the first 20 hexadecimal characters of `HMAC-SHA256(random 32-byte escrow key, G1V2-CJ-1(secret truth record))`. Consequently those identity values do not exist until production blind truth and a new escrow key are used.

Task85c-d sections 64 and 67 prohibit constructing or inspecting the 144 production blind controls and prohibit generating blind escrow. Therefore the exact JobID set required for the complete-DAG bijection cannot be calculated within this task's authorized inputs.

## Why substitutions are invalid

- `BLIND_M0_A/2000/0`-style IDs reveal class/generator/scale/replicate truth and violate the opaque blind identity rule.
- `blind-000`-style placeholders are not the HMAC-derived production identities and produce a different JobID set/root.
- Omitting `control_instance_id` violates the frozen JobID payload.
- Deferring the 144 rows makes the required 1,321,152-job bijection and edge root unprovable.

Resolving this requires changing blind identity construction timing, providing an already frozen blind-safe ID set, or relaxing the complete-DAG proof to a parameterized theorem. Each is outside the exact R2-G01/R2-G02 repair scope.

## Required action

Issue a separate authority that reconciles complete-DAG identity audit with blind escrow timing without exposing truth or changing scientific inputs. Then rerun Task85c-d. No V1.1.1 patch is issued here.
