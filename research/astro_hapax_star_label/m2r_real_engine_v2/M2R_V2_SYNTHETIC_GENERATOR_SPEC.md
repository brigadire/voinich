# Independent synthetic generator and evaluation

`generator.py` does not import inference. `engine.py` does not import generator,
evaluator, seed registry or family metadata. `worker.py` accepts only neutral JSON
train/heldout partitions, each containing {terms,labels}; each row has {id,surface}.
IDs and row orders are independently randomized. Filenames are hashes of surface JSON.

The generator samples three-position morphology with five units per role, variable
unit lengths 2/3, and independent source/target symbol inventories. Three inventories
serve in-family datasets and a fourth is reserved for out-of-family composition.
Train and heldout use disjoint combinations. All baselines contain affixes and internal
recurrent units; named affix/internal/mixed conditions vary sample size and alphabet,
not three independent linguistic generative mechanisms. Composition adds another
medial unit; partial coverage inserts unexplained fragments; distractors add unmatched
surfaces. Sizes 29, 57, 68 refer to core positive examples; the distractor condition
adds extras and reports actual counts. Support distributions vary through sampling.

Truth records full rule triples, surface-pair assignments and ordered latent paths.
It lives in `latent_truth/`, excluded by the worker read guard. The evaluator opens
truth only after a committed inference output exists. Development/calibration/hidden/
null/hard-negative/out-of-family seed families are disjoint and registered. Calibration
uses the sole fixed profile; no threshold or profile selection occurs.

Nulls include independent random bags, length matched random labels, global unigram
redistribution, per-token Euler trails preserving bigrams, shuffled latent assignments,
repeated fragments, independent role-frequency recombination and size matched random
labels. Shuffled assignments are an identifiability control: their surface distributions
can equal a positive, so separation is not assumed. Euler trails may be unique and
therefore leave tokens unchanged; this is reported rather than concealed.

HN1/HN2/HN10 use nonrecurrent large-inventory labels (local/pair-specific/complexity
stress); HN3 rotates label roles; HN4 changes only the heldout encoding; HN5 adds chance
recurrent fragments; HN6 reverses component order; HN7 mixes incompatible alphabets;
HN8 repeats a single pair with different IDs; HN9 has symmetric Cartesian bags.
These are minimum adversarial fixtures, not a broad benchmark of those families.

Evaluation uses exact role-aware rule intersections, true surface pairs, path and
boundary agreement, coverage, all six MDL components, margins, alternative-optimum
assignment recall and predictive acceptance. Top-k assignment accuracy is candidate
edge recall across retained alternatives, not probability of the complete correct
matching. Ambiguity fields and margins are diagnostic; no probabilistic calibration
claim is made. One replicate per condition is a limited synthetic check, not a power
analysis. Hidden disclosure seals v2 against further implementation changes.
