# G1-v2 status and reachability contract V2

+Version: `G1_V2_STATUS_REACHABILITY_CONTRACT_V2`. This is a normative correction layer over `G1_V2_EXECUTABLE_CONTRACT_V1` limited to status taxonomy, stage producer legality, reachability, and failure aggregation. Machine JSON and fully expanded TSVs are normative. It does not issue executable contract V1_1 or repair JSON Schemas.

+## Taxonomy and central invariant

+Scientific/procedure failure is not scientific FAIL. PASS and FAIL are exclusively assessable gate/verifier verdicts. NOT_ASSESSABLE is a scientifically specified inability to assess a reached statistic. FIT_SUCCESS, GENERATION_SUCCESS, COMPLEXITY_SUCCESS, and AGGREGATION_SUCCESS record procedure completion without claiming scientific PASS. Concrete failures are FIT_FAILURE, NUMERICAL_FAILURE, INDUCTION_CAP, GENERATION_FAILURE, and PROTOCOL_VETO. NOT_REACHED is only a causal DAG-suppression record. `SCIENTIFIC_FAILURE` is removed and has no wildcard meaning.

+## Stages and evidence

+The job stages are FIT, PREDICTIVE, GENERATION, F2_METRIC, COMPLEXITY, CANDIDATE_AGGREGATION, and CONTROL_AGGREGATION, exactly matching G1V2-DAG-1 templates. Structural family/gate/verdict evaluation is a deterministic suboperation of CANDIDATE_AGGREGATION over F2 evidence, not a separate DAG job. The evidence/status mapping is normative TSV; procedure failures use scientific_failure evidence, suppressed jobs use not_reached evidence, and ordinary evidence types never carry a failure status.

+## Reachability

+FIT success permits predictive, generation, and complexity dependency paths; concrete FIT failures suppress them. Generation additionally requires predictive PASS. Predictive FAIL, NOT_ASSESSABLE, failure, veto, or NOT_REACHED suppresses generation but never candidate aggregation. Generation success alone permits F2; every generation missing/failure/veto state suppresses F2. F2 and complexity records always flow to candidate aggregation, including missing/failure records. Candidate aggregation always flows to control aggregation.

+For multiple dependencies, a job runs iff every direct upstream transition says RUN. Otherwise it emits NOT_REACHED. Causal precedence is PROTOCOL_VETO, procedure failure, NOT_ASSESSABLE, FAIL, NOT_REACHED; ties use lowest bytewise upstream JobID while retaining all dependency IDs. DAG edges and all 1,321,152 jobs remain present.

+## Aggregation

+Candidate ADEQUATE requires the complete frozen positive path. Candidate INADEQUATE requires complete assessable negative evidence with no required missing/failure evidence. Otherwise it is UNRESOLVED; PROTOCOL_VETO produces PROTOCOL_INVALID. No failure contributes negative evidence.

+M3 is ADEQUATE if either route is adequate, INADEQUATE only if both routes are inadequate, otherwise UNRESOLVED. A class is adequate if any candidate is adequate, inadequate only when every required candidate is assessably inadequate, otherwise unresolved. NONE requires every M0-M5 class inadequate. Missing/failure evidence that can change the minimum yields NOT_IDENTIFIABLE. A lowest adequate singleton with all lower classes inadequate yields RECOVERED_Mj; equivalent minima yield NOT_IDENTIFIABLE/EQUIVALENT_SET. PROTOCOL_VETO invalidates the chain and produces PROTOCOL_INVALID, not an ordinary scientific verdict.

+The fully expanded transition table and reason registry contain the exact machine behavior. Infrastructure retry/lease/worker states never enter this scientific machine.
