# Task85c-h design

The implementation gate begins by resolving every scientific identity that a
handler must place in its JobID and evidence. Only after those identities are
jointly satisfiable may production packages or handlers be changed. This order
prevents code from silently selecting between contradictory scientific
authorities.

The gate reproduced two contradictions before implementation work. Per §9 of
Task85c-h, they require `STOP SCIENTIFIC_CONTRACT_DEFECT`; no code, production
material, thresholds, secrets, JobIDs, DAG, or candidate executable was created.
