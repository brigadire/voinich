# TASK86C_V2_RUNBOOK.md

# Task86C-v2 — G1-v2 Blind Control Computation Runbook

## 1. Purpose

This runbook defines the operator procedure for executing the frozen
Task86C-v2 distributed computation.

The operator's responsibility is limited to:

1. validating the frozen execution environment;
2. deploying/starting compatible workers;
3. staging frozen immutable inputs;
4. submitting the frozen Task86C-v2 run manifest;
5. monitoring infrastructure health;
6. allowing infrastructure-safe retries/requeues;
7. verifying computational completion;
8. freezing the resulting immutable evidence graph;
9. handing the completed result set to Task86C-v2-analysis.

The operator MUST NOT perform scientific interpretation during this
procedure.

The operator MUST NOT:

- change scientific parameters;
- change thresholds;
- change seeds;
- change model definitions;
- change PM/F2 definitions;
- change scientific reachability rules;
- manually select preferred results;
- inspect blind ground-truth mappings;
- unblind confirmatory controls;
- modify the manifest to work around failed scientific jobs;
- rerun scientific failures merely because their result is undesirable.

The run is an execution of an already frozen experiment, not an
interactive scientific workflow.


======================================================================
2. AUTHORITATIVE FREEZE CHAIN
======================================================================

Before execution verify the presence and integrity of the terminal
markers for:

    Task85b
        G1_V2_EXPERIMENT_CONTRACT_FROZEN

    Task86C-v2-prep
        TASK86C_V2_COMPUTE_READY_FROZEN

    Task86C-v2-deploy-audit
        TASK86C_V2_DEPLOYMENT_READY_FROZEN

Do not start Task86C-v2 if any required freeze marker is absent or
its authoritative manifest fails verification.


======================================================================
3. AUTHORITATIVE DOCUMENTS
======================================================================

The execution must follow the frozen contracts under:

    research/phase3/task85b/

    research/phase3/task86c-v2-prep/

    research/phase3/task86c-v2-deploy-audit/

Especially:

    TASK86C_V2_SCIENTIFIC_HANDOFF.md
    TASK86C_V2_CLUSTER_RECOMMENDATION.md
    WORKER_COMPATIBILITY_CONTRACT.md
    WORKER_DEPLOYMENT_CONTRACT.md
    WORKER_OPERATIONS.md

If this runbook conflicts with a frozen authoritative contract,
the frozen contract wins.

STOP rather than improvising.


======================================================================
4. FROZEN EXECUTABLE
======================================================================

Required G1-v2 executable:

    g1v2-executor

Required SHA-256:

    6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718

Required platform:

    Linux/amd64

Compatibility tuple:

    protocol:
        g1v2-execution-v1

    evidence:
        g1v2-evidence-v1

    canonicalization:
        canonical-json-nfc-v1

Do not build the executable independently on workers.

Do not substitute a newer executable even if repository HEAD has
changed.

Archive the exact executable used for the run.


======================================================================
5. FIREWALLS
======================================================================

The following firewalls remain active during execution.

### Voynich firewall

Task86C-v2 is a control-validation experiment.

Do not introduce Voynich target data into this run unless the frozen
Task86C-v2 scientific manifest explicitly contains an already
authorized non-target artifact required by its control design.

No Task86V work is permitted.

### Confirmatory firewall

Do not inspect:

- hidden generator classes;
- escrow mappings;
- blind control identities;
- ground-truth labels intended for later unblinding.

The operator does not need this information to run the computation.

### Scientific firewall

Operational configuration may change:

- worker inventory;
- worker count;
- worker concurrency;
- coordinator address;
- cache paths;
- evidence-storage paths.

Scientific configuration may not change:

- manifest;
- models;
- candidates;
- seeds;
- thresholds;
- metrics;
- dependencies;
- JobIDs;
- scientific status semantics.


======================================================================
6. CLUSTER TOPOLOGY
======================================================================

Architecture:

                 frozen run manifest
                         |
                         v
                    coordinator
                         |
                 mTLS pull / lease
                  /      |      \
                 v       v       v
              worker  worker  worker
                  \      |      /
                         v
                  immutable evidence
                         |
                         v
                    result graph

Workers dial the coordinator.

Workers do not require application ingress.

The architecture supports:

    1 coordinator + N workers.

Physical node count and worker-slot count are distinct.


======================================================================
7. CURRENT KNOWN HOSTS
======================================================================

Known audited compute host:

    cognition
    10.10.24.105

Audited coordinator endpoint during deployment testing:

    https://10.10.24.107:38490

Do not assume these addresses remain correct.

Use the actual production inventory and frozen operational
configuration selected for the Task86C-v2 run.


======================================================================
8. CAPACITY REQUIREMENTS
======================================================================

Task85b/prep estimated approximately:

    ~15,552 candidate fits
    <=5,184 generation batches
    ~36,288 family-scale F2 evaluations

Estimated total computation:

    approximately 346–399 CPU-hours

Expected immutable evidence:

    approximately 60–100 GB

Required authoritative evidence capacity:

    >=120,000,000,000 free bytes

Audited cognition evidence capacity:

    178,870,026,240 bytes

Do not start if authoritative evidence storage is below the frozen
minimum.


======================================================================
9. PRE-RUN RECORD
======================================================================

Before changing cluster state create an operator run record.

Suggested location:

    research/phase3/task86c-v2/run/

Create:

    TASK86C_V2_RUN_RECORD.md

Record:

    UTC start timestamp;
    repository commit;
    Task85b manifest hash;
    Task86C-v2-prep manifest hash;
    deployment-audit manifest hash;
    scientific run-manifest path/hash;
    executable path/hash;
    coordinator host;
    coordinator endpoint;
    evidence-storage path;
    available evidence bytes;
    worker inventory;
    worker slot counts;
    certificate identities;
    input/CAS manifest hash.

Do not record private-key material.


======================================================================
10. REPOSITORY STATE
======================================================================

Record:

    git rev-parse HEAD
    git status --short

Prefer a clean repository.

If unrelated local modifications exist, record them explicitly.

Do not modify scientific implementation after the run starts.


======================================================================
11. EXECUTABLE VERIFICATION
======================================================================

On the controller:

    sha256sum /path/to/g1v2-executor

Require exactly:

    6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718

STOP on mismatch.

Archive a copy or otherwise preserve the exact executable by
content hash.


======================================================================
12. PKI PREFLIGHT
======================================================================

For every worker verify:

    unique worker certificate;
    unique worker private key;
    project CA certificate;
    URI SAN:

        voinich-worker:<identity>

Workers must NOT receive:

    ca.key

Verify the coordinator certificate covers the hostname/IP actually
used by workers.

Never disable TLS verification to resolve a connection problem.


======================================================================
13. ANSIBLE PREFLIGHT
======================================================================

Use the audited:

    ansible/roles/voynich_worker

deployment path.

Before production deployment run:

    ansible-playbook --syntax-check ...

then:

    ansible-playbook --check ...

using the exact intended production variables.

Failure in check mode must be investigated before deployment.


======================================================================
14. WORKER DEPLOYMENT
======================================================================

For each worker configure at minimum:

    coordinator URL;
    concurrency/slot count;
    frozen executable source;
    CA certificate;
    unique worker certificate;
    unique worker key;
    persistent cache;
    evidence configuration where applicable.

Example audited invocation:

    ANSIBLE_CONFIG=/home/brigadire/devops/workdir/ansible/ansible.cfg \
    ansible-playbook \
      -i /home/brigadire/devops/workdir/inventory_dev/voinich/hosts \
      ansible/deploy-workers.yml \
      -e voynich_worker_target_group=cognition \
      -e voynich_worker_coordinator_url=https://10.10.24.107:38490 \
      -e voynich_worker_concurrency=4 \
      -e voynich_g1v2_binary_src=/path/to/frozen/g1v2-executor \
      -e voynich_worker_ca_src=/secure/path/ca.crt \
      -e voynich_worker_cert_src=/secure/path/worker-cognition.crt \
      -e voynich_worker_key_src=/secure/path/worker-cognition.key \
      -e voynich_g1v2_evidence_dir=/usr/local/data/voinich-evidence

This is an example from the audited environment.

Use actual production paths/hosts.


======================================================================
15. WORKER CONCURRENCY
======================================================================

Do not blindly maximize worker processes.

Use the audited rule:

    never exceed available vCPU count;

and maintain approximately:

    2–4 GB RAM per heavy slot

unless the frozen cluster recommendation specifies otherwise.

Avoid hardware oversubscription.

Changing worker concurrency is operational and does not alter the
scientific manifest.


======================================================================
16. DEPLOYMENT IDEMPOTENCY
======================================================================

After deploying the intended fleet, run the identical Ansible
deployment again.

Require:

    changed=0

for the second stable run.

If configuration continues changing on every invocation, STOP and
investigate before starting the scientific run.


======================================================================
17. INSTALLED BINARY VERIFICATION
======================================================================

Verify the installed executable hash on every physical worker host.

Every worker must run:

    6b015b2e4078b9b5f109ebf3aa8d73918888e431bde267e0d10c3013b524f718

A single drifted worker blocks the run.


======================================================================
18. SERVICE VERIFICATION
======================================================================

Verify all intended worker slots are:

    installed;
    enabled;
    active.

Supported audited service managers include:

    systemd
    OpenRC

Verify workers survive:

    service restart;
    network interruption;
    coordinator restart.

Do not start the scientific manifest until worker admission is
stable.


======================================================================
19. STORAGE PREFLIGHT
======================================================================

Verify:

### Worker cache

Persistent worker cache, normally under:

    /var/lib/voinich-g1v2/cache

### Worker temp

Persistent job temp tree, normally under:

    /var/lib/voinich-g1v2/tmp

### Authoritative evidence

Coordinator-visible authoritative evidence store.

Audited cognition path:

    /usr/local/data/voinich-evidence

Require:

    >=120,000,000,000 free bytes

or the stricter frozen requirement if the final scientific handoff
specifies one.

Never use `/tmp` as authoritative evidence storage.


======================================================================
20. CAS PRESEED
======================================================================

Preseed all frozen immutable inputs that the run contract permits or
requires to avoid unnecessary cold transfer.

Example:

    voynich_g1v2_preseed:
      - src: /secure/frozen-inputs/object.bin
        sha256: <EXPECTED_SHA256>

Objects must be stored by SHA-256 and verified after copy.

After preseed:

    rerun deployment;
    verify changed=0 where expected;
    verify cache objects remain byte-identical.


======================================================================
21. SCIENTIFIC MANIFEST VERIFICATION
======================================================================

Before submission identify the single authoritative Task86C-v2
scientific run manifest.

Record:

    path;
    SHA-256;
    schema/protocol version;
    expected job count or closure metadata.

Run the authoritative manifest verifier.

Do not submit a manifest that fails validation.

After this point treat the manifest as immutable.


======================================================================
22. MANIFEST ARCHIVE
======================================================================

Copy/archive the exact submitted manifest into the run record area.

Suggested:

    research/phase3/task86c-v2/run/
        FROZEN_RUN_MANIFEST.<ext>

Record its SHA-256.

The archived manifest and submitted manifest must be byte-identical.


======================================================================
23. COORDINATOR START
======================================================================

Start the coordinator using the frozen Task86C-v2-prep
implementation/configuration.

Record:

    coordinator command/config;
    executable hash;
    listen endpoint;
    evidence path;
    start timestamp.

Confirm:

    TLS active;
    evidence storage writable;
    manifest accepted;
    no compatibility errors.


======================================================================
24. WORKER ADMISSION
======================================================================

Before scientific submission confirm every intended worker/slot has
successfully completed:

    mTLS authentication;
    compatibility handshake.

Expected compatibility:

    g1v2-execution-v1
    g1v2-evidence-v1
    canonical-json-nfc-v1

Do not equate "service running" with "worker admitted".


======================================================================
25. OPTIONAL ENGINEERING SMOKE TEST
======================================================================

If the frozen handoff calls for a final smoke test, execute only the
approved engineering fixture.

Do NOT use:

    confirmatory controls;
    blind scientific jobs;
    Voynich data.

Require successful accepted evidence.

Then clear/separate engineering execution state exactly as required
by the frozen execution contract.


======================================================================
26. START OF BLIND SCIENTIFIC RUN
======================================================================

Record:

    TASK86C_V2_BLIND_RUN_START_UTC

From this point:

    do not modify scientific code;
    do not modify scientific manifest;
    do not modify seeds;
    do not modify thresholds;
    do not inspect ground truth;
    do not unblind controls.

Submit the frozen scientific manifest exactly once according to the
executor interface.


======================================================================
27. MONITORING DURING EXECUTION
======================================================================

Monitor operational state only.

Allowed monitoring includes:

    pending jobs;
    leased jobs;
    completed jobs;
    worker connectivity;
    lease expiry;
    retries;
    CPU/RAM;
    disk space;
    network;
    coordinator health;
    evidence-store growth.

Avoid inspecting scientific metric values unless required to
diagnose evidence corruption or an implementation failure.


======================================================================
28. SAFE OPERATIONAL ACTIONS
======================================================================

The following are normally allowed without changing the experiment:

    restart failed worker service;
    restart coordinator;
    restore network connectivity;
    add a compatible frozen worker;
    remove a failing worker;
    allow lease expiry/requeue;
    increase/decrease operational worker concurrency within
        validated resource constraints;
    repair infrastructure storage/network failures.

All such events must be logged.


======================================================================
29. ADDING WORKERS DURING RUN
======================================================================

A compatible worker may be added without changing the scientific
run.

Procedure:

    provision Linux/amd64 host;
    issue unique worker certificate;
    add inventory/host vars;
    deploy exact frozen executable;
    verify hash;
    verify resource/storage constraints;
    verify service;
    verify mTLS/compatibility handshake;
    join lease pool.

Do NOT regenerate the scientific manifest.

Do NOT regenerate JobIDs.


======================================================================
30. REMOVING A WORKER
======================================================================

For a failing worker:

    stop worker cleanly where possible;
    allow existing lease to expire;
    verify same JobID is requeued;
    optionally revoke/deny certificate;
    remove worker from operational inventory.

Do not mark affected scientific jobs as failed merely because the
machine disappeared.


======================================================================
31. INFRASTRUCTURE RETRIES
======================================================================

Infrastructure failures may be retried automatically under the
frozen executor semantics.

Examples:

    worker loss;
    coordinator interruption;
    network interruption;
    result-transfer interruption;
    transient storage failure.

The retried job must retain the same immutable JobID/bundle.


======================================================================
32. SCIENTIFIC FAILURES
======================================================================

Do NOT manually retry scientific outcomes to obtain a different
answer.

Examples include:

    TRAINING_FAILED
    INDUCTION_LIMIT_REACHED
    CONVERGENCE_FAILED
    NUMERICALLY_UNSTABLE
    GENERATION_FAILED
    METRIC_UNAVAILABLE

unless the frozen protocol explicitly classifies the event as
retryable infrastructure failure.

These statuses are evidence.


======================================================================
33. DUPLICATE RESULTS
======================================================================

Equal verified duplicate results are handled according to the frozen
executor contract.

Preserve duplicate provenance.

Do not manually delete or select copies.


======================================================================
34. CONFLICTING RESULTS
======================================================================

If identical JobID executions produce conflicting verified content:

    STOP scientific progression as required by the executor;
    preserve all copies;
    preserve provenance;
    quarantine conflict;
    do not choose a preferred result.

A conflict is an integrity event, not an opportunity for majority
voting.


======================================================================
35. DISK PRESSURE
======================================================================

Monitor authoritative evidence storage throughout the run.

If free space approaches an unsafe level:

    do not delete verified evidence;
    do not change scientific workload;
    do not truncate evidence.

Pause/stop new leasing if supported and expand/migrate storage using
an integrity-preserving operational procedure.

Record all actions.


======================================================================
36. COORDINATOR RESTART
======================================================================

Coordinator restart is permitted under the validated resume model.

After restart verify:

    completed JobIDs remain complete;
    incomplete leases recover/requeue;
    dependency graph is restored;
    evidence index is intact;
    no completed scientific work is silently recomputed.

Record the restart event.


======================================================================
37. OPERATOR INCIDENT LOG
======================================================================

Maintain:

    TASK86C_V2_OPERATOR_EVENTS.tsv

Suggested columns:

    timestamp_utc
    event_type
    host
    worker_identity
    job_id_if_known
    action
    reason
    infrastructure_or_scientific
    outcome
    notes

Do not place secrets in this file.


======================================================================
38. PROGRESS SNAPSHOTS
======================================================================

Periodically record aggregate operational progress.

Suggested:

    timestamp
    total_jobs
    pending
    leased
    completed
    infrastructure_retries
    conflicts
    active_workers
    evidence_bytes
    free_storage_bytes

These snapshots are operational telemetry only.


======================================================================
39. DO NOT STOP ON SCIENTIFIC NONE/FAIL
======================================================================

Do not terminate the bulk computation because intermediate
scientific results appear poor.

The operator must not decide:

    "M3 is failing, so skip the rest"

or:

    "natural language looks NONE, so stop."

Only frozen DAG/reachability semantics may suppress downstream jobs.


======================================================================
40. COMPLETION CONDITION
======================================================================

"Workers are idle" is NOT sufficient.

"Queue is empty" is NOT sufficient.

The computation is complete only when the authoritative executor /
closure verifier establishes that the frozen run manifest has a
complete terminal accounting.

Every planned scientific cell must be represented by:

    verified result evidence

or:

    valid frozen NOT_REACHED evidence.

There must be no unexplained missing jobs.


======================================================================
41. REQUIRED COMPLETION CHECKS
======================================================================

Before declaring the calculation complete verify:

    manifest closure;
    expected JobID accounting;
    dependency closure;
    evidence hashes;
    code/config hashes;
    seed closure;
    schema compatibility;
    duplicate resolution;
    zero unresolved conflicts;
    zero unexplained missing artifacts;
    valid NOT_REACHED records;
    evidence-store integrity.


======================================================================
42. EVIDENCE-ONLY VERIFICATION
======================================================================

Run the frozen evidence-only verifier over the complete result graph.

This verifier must NOT fit models or regenerate corpora.

Require successful reconstruction of the evidence decision graph.

At this stage the purpose is integrity/closure verification.

Do not perform scientific interpretation manually.


======================================================================
43. REPRODUCIBILITY RECORD
======================================================================

Record at completion:

    scientific manifest SHA-256;
    executable SHA-256;
    result/evidence manifest SHA-256;
    result graph root/hash if produced;
    evidence-store size;
    total jobs;
    completed results;
    NOT_REACHED count;
    infrastructure retry count;
    duplicate count;
    conflict count;
    final worker inventory;
    UTC completion timestamp.


======================================================================
44. RESULT FREEZE
======================================================================

Once closure verification succeeds:

    stop accepting changes to the result graph;

    generate/finalize the authoritative result manifest;

    hash all handoff artifacts;

    preserve the evidence store;

    archive the operator event log;

    archive the exact run manifest;

    archive the exact executable or immutable reference to it.

Do not clean verified evidence after freeze.


======================================================================
45. COMPUTATION TERMINAL STATE
======================================================================

The operator may declare:

    TASK86C_V2_COMPUTATION_COMPLETE

only if:

    authoritative manifest verified;
    all planned jobs accounted for;
    evidence closure passes;
    evidence-only verifier passes;
    no unresolved result conflict exists;
    result manifest is frozen.

This is an OPERATIONAL state.

It is NOT a scientific verdict.


======================================================================
46. FAILURE TERMINAL STATE
======================================================================

If the run cannot reach valid closure because of:

    unresolved deterministic conflict;
    evidence corruption;
    incompatible executable;
    manifest integrity failure;
    irrecoverable evidence loss;
    scientific-contract drift;
    firewall violation;

do NOT declare computation complete.

Record:

    TASK86C_V2_COMPUTATION_INTEGRITY_FAILED

and preserve all available evidence for audit.


======================================================================
47. NON-FATAL INFRASTRUCTURE INTERRUPTIONS
======================================================================

Do not use the failure terminal state merely because:

    a worker died;
    coordinator restarted;
    network temporarily failed;
    a lease expired;
    an infrastructure job was safely retried.

These conditions were explicitly designed to be recoverable.


======================================================================
48. HANDOFF TO ANALYSIS
======================================================================

Task86C-v2-analysis must receive at minimum:

    frozen scientific run manifest;
    scientific run-manifest SHA-256;
    frozen result/evidence manifest;
    complete authoritative evidence graph/store;
    exact executable hash;
    repository commit;
    operator run record;
    operator event log;
    completion/closure verifier output;
    duplicate/conflict summary;
    worker/runtime telemetry;
    blind/escrow material required by the frozen unblinding protocol.

Do not unblind the material before the analysis procedure calls for
it.


======================================================================
49. ANALYSIS FIREWALL
======================================================================

The handoff should allow the analysis agent to proceed in this
order:

    1. verify provenance;
    2. verify hashes;
    3. verify manifest closure;
    4. verify evidence closure;
    5. regenerate decisions from evidence;
    6. establish experiment integrity;
    7. only then unblind;
    8. compute recovery verdicts;
    9. analyze natural-language controls;
    10. determine G1-v2 identifiability;
    11. write scientific report.

Do not provide a manually interpreted summary as a substitute for
the evidence graph.


======================================================================
50. OPERATOR CHECKLIST
======================================================================

Before run:

- [ ] Task85b freeze verified.
- [ ] Task86C-v2-prep freeze verified.
- [ ] Deployment-audit freeze verified.
- [ ] Repository commit recorded.
- [ ] Frozen executable archived.
- [ ] Executable SHA-256 verified.
- [ ] Scientific manifest identified.
- [ ] Scientific manifest hash recorded.
- [ ] Manifest verifier passes.
- [ ] Coordinator certificate verified.
- [ ] Unique worker certificates verified.
- [ ] No worker has ca.key.
- [ ] Ansible syntax-check passes.
- [ ] Ansible check-mode passes.
- [ ] Worker deployment passes.
- [ ] Second deployment has changed=0.
- [ ] Installed hashes verified.
- [ ] Worker services active.
- [ ] mTLS/compatibility admission verified.
- [ ] Worker concurrency reviewed.
- [ ] Immutable inputs preseeded/verified.
- [ ] Authoritative evidence storage >= required minimum.
- [ ] Operator run record created.
- [ ] Operator event log created.
- [ ] Blind material has not been inspected.

During run:

- [ ] Scientific manifest remains unchanged.
- [ ] Executable remains unchanged.
- [ ] Scientific configuration remains unchanged.
- [ ] Only operational state is monitored.
- [ ] Worker failures use normal lease/requeue.
- [ ] Infrastructure retries preserve JobID.
- [ ] Scientific failures are not manually retried.
- [ ] Added workers pass frozen compatibility checks.
- [ ] No verified evidence is deleted.
- [ ] Conflicts, if any, are quarantined.
- [ ] Storage capacity remains safe.
- [ ] Operational interventions are logged.

At completion:

- [ ] All manifest jobs accounted for.
- [ ] Missing cells = 0 except valid NOT_REACHED.
- [ ] Evidence hashes verify.
- [ ] Dependencies verify.
- [ ] Seeds/config/code closure verifies.
- [ ] No unresolved duplicate conflicts.
- [ ] Evidence-only verifier passes.
- [ ] Result graph/manifest frozen.
- [ ] Result manifest hash recorded.
- [ ] Operator logs archived.
- [ ] Exact executable reference archived.
- [ ] TASK86C_V2_COMPUTATION_COMPLETE recorded.
- [ ] Blind ground truth remains unopened.
- [ ] Complete package handed to Task86C-v2-analysis.


======================================================================
51. CRITICAL STOP CONDITIONS
======================================================================

STOP and preserve state if any of the following occurs:

    frozen executable hash mismatch;

    scientific manifest changes after start;

    unexpected scientific configuration drift;

    worker accepted with incompatible protocol/schema;

    evidence hash mismatch;

    unresolved same-JobID result conflict;

    authoritative evidence loss;

    result graph cannot be closed;

    evidence-only verifier fails;

    blind ground truth is accidentally exposed;

    Voynich firewall is violated;

    fixing the problem would require changing the frozen scientific
    contract.

Do not repair these conditions by modifying the experiment in place.


======================================================================
52. PRINCIPLE
======================================================================

The operator is executing a frozen experiment.

The operator controls:

    machines,
    services,
    storage,
    networking,
    worker capacity.

The frozen experiment controls:

    jobs,
    models,
    metrics,
    seeds,
    dependencies,
    evidence semantics,
    scientific decisions.

The two layers must remain separate throughout Task86C-v2.
