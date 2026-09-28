# M2R-v2 implementation contract — fixed before implementation

Scope: synthetic-only, no real-data authorization. Unicode code points, exact NFC
surface strings; no alphabet-specific transliteration. Inference receives only two
unordered lists of {id, surface}, per train/heldout partition. IDs are opaque,
validated, and never used to rank models. Repeated surfaces are collapsed before
inference, even under different IDs; repeated IDs with inconsistent surfaces are
errors. Train/heldout overlapping surfaces are rejected. Canonical results use
content-derived identities, with original IDs retained only in loader provenance.

## Units and derivations

Enumerate every contiguous substring of length 1–4 plus the entire token.
Roles are WHOLE (both boundaries), INITIAL, MEDIAL, FINAL, determined by span,
not by unit spelling. Stable SHA-256 IDs include side, role, surface. Record
occurrences, unique surface-instance support and provenance spans. Candidate
rules pair same-role units recurring in >=3 distinct surfaces on each side.
A rule is (role, source surface, target surface). No deletions or insertions as
rules; a documented gap consumes one code point on one side and pays literal
plus exception cost. Composition concatenates rules in order, without role
transfer. Unit substitution may change length. Conflicting same-role source to
different targets, or same-role target to different sources, is forbidden.
Overlapping segmentations are competing derivations; transform reports all
outputs and explicit ambiguity, never dictionary overwrite.

For every possible pair retain top 2 order-preserving paths within the pruned
candidate space. Each path records spans, rules, gaps, literal costs, complexity,
coverage and score. Local paths are not assignments. Final pair scores use the
current global model. Inadmissible whole-token singleton rules remain excluded.

## Support and assignment

MIN_INDEPENDENT_SUPPORT=3. Count occurrences, distinct pairs, distinct terms and
distinct labels from the chosen one-to-one assignment and its derivations. All
three distinct counts must be >=3. Same-surface duplicates cannot add support.
Invalid models cannot receive a finite improving score during training. Heldout
application never re-estimates support or changes the trained rule set.

Exact Hungarian maximum-weight matching on an augmented bipartite graph with
one dummy column per term. Unmatched terms and labels cost their literal bits.
Real edges are baseline savings minus alignment and assignment costs. Zero or
negative saving edges remain unmatched. Report exclusion margin for every
selected edge and distinct alternative optimal assignments (up to 3). This
is a bounded list, not exhaustive enumeration of all symmetric permutations.
Canonical tie-breaking sorts surfaces and rule identities; renaming opaque IDs
cannot change the scientific output. True ambiguity prevents predictive acceptance.

## Explicit description lengths (bits)

L(s)=8*len(s.encode('utf-8')). Baseline=sum L over unique terms and labels.
RULE_CODE_LENGTH=sum(10+L(source)+L(target)) over model rules.
ASSIGNMENT_CODE_LENGTH=matches*(log2(n_terms+1)+log2(n_labels+1)).
Each rule occurrence costs 1+ceil(log2(number_rules+1)) bits (EXPLAINED_ALIGNMENT_COST).
UNEXPLAINED_SOURCE_COST and UNEXPLAINED_TARGET_COST are literal costs of gaps
and unmatched strings. EXCEPTION_COST=number of gap code points inside matches.
TOTAL=sum these six terms. NORMALIZED_COMPRESSION_GAIN=(BASELINE-TOTAL)/BASELINE.
No-rule baseline has zero gain. Empty data: EMPTY, gain 0, no acceptance.
<3 unique examples: SMALL, no train rules. Negative gain allowed.
Model complexity is charged in full on heldout; no re-fitting or threshold changes.
Scores are dimensionless, comparable across sizes; fixed model overhead is not
assumed to vanish at 29/57/68. Exact duplication leaves all metrics unchanged.

## Bounded search

GLOBAL_OPTIMUM_CLAIMED=NO
SEARCH_TYPE=BOUNDED_DETERMINISTIC_OPTIMIZATION

State = canonical rules, pair scores, exact assignment, derivations, support,
raw MDL and normalized gain. Search proposals: add, remove, replace rule and
re-evaluate assignment. Merge/split/role changes are not separate operators in v2;
these are reachable through remove/add in the frozen candidate pool.
Prune candidates by descending potential literal saving, frequency difference,
then canonical rule key. Pool 48; proposals per beam state 12 additions, all
removals, up to 4 replacements. Beam width 3; max states 96; max iterations 8;
max rules 12; final models 3; local paths 2; max unique strings per side 80;
max token length 32; max input rows 256; max candidate units 12000.
Wall clock 120 seconds per search including candidate generation, memory 512 MiB
(worker RLIMIT_AS plus runtime RSS checks). Exceeding wall/memory/size budget
returns INCOMPLETE, never NO_MODEL. Reaching the predeclared state/iteration
budget completes the bounded algorithm, explicitly not an exhaustive search.
Check deadline within alignment/matching loops. Canonical state deduplication.

Checkpoint after each complete iteration: atomic replace, fsync, SHA-256 envelope,
engine/config/input fingerprints, beam, visited states, count, iteration. Resume
verifies all fingerprints and reconstructs deterministic derived state. Interrupted
partial iterations are discarded. Cumulative wall budget is checkpointed; timing
is excluded from canonical scientific output. Seed registry rejects duplicates;
seeds/family/shard metadata exist solely in orchestration, never engine input.
Aggregation sorts content dataset keys and rejects duplicate/conflicting results.

## Evaluation and fixed gates

Generator is an independent module; engine cannot import generator/evaluator or
open dataset paths. A synthetic worker runs with a filesystem audit guard denying
reads outside its package and Python runtime and denying truth access. Neutral
surface JSON uses no family/seed/latent IDs. Truth lives separately; evaluator
opens it only after outputs commit. Positive families vary affixes, medial units,
composition lengths, partial fragments, unmatched data, alphabet and support.
Six disjoint seed families: development, calibration, hidden, null, hard-negative,
out-of-family. One fixed profile; calibration may validate but cannot tune it.
Hidden validation runs once following candidate code/config hash seal. Any change
after disclosure invalidates v2; a new attempt requires v3 and new hidden seeds.

Predictive acceptance: complete train/heldout, train and heldout gain >0,
train and heldout fully-derived token coverage >=0.70, transfer retention >=0.70,
all train rules supported, no assignment ambiguity. Thresholds fixed here.
Evaluation compares full rule tuples and assignment surfaces against latent truth,
not an alphabet membership test. Precision empty denominator =0, recall =0.
Boundary and path metrics are exact per latent true pair; no alignment means 0.
Hidden mandatory macro gates: precision>=.80, recall>=.75, assignment>=.75,
heldout coverage>=.70, heldout gain>0, singleton=0, unsupported=0,
mean independent support>=3, positive median gain-null P95>=.05,
zero predictive hard-negative acceptances; all regression and permutation,
neutral-ID, shard and checkpoint invariance tests pass. Any failure: BLOCKED,
no freeze manifest, no production execution. Quantiles use nearest rank.

A shuffled-assignment null preserves the unordered input bags by definition.
It is deliberately retained as an identifiability control: inference cannot
separate it from the same positive surfaces. Do not manufacture score separation
using truth or input order. Report this limitation and gate failure if observed.

SHA256SUMS covers every regular package file except itself; bytecode is disabled.
The ledger cannot hash itself. A successful freeze additionally records a full
source snapshot, configuration, runtime manifest, seed/input/result hashes and
independent audit instructions. On failure produce an integrity ledger only,
explicitly not a frozen production package. Original v1 hashes must be unchanged.
