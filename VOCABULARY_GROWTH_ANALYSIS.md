# Vocabulary growth analysis (Task49)

`vocabulary-growth-analyze` is a language-agnostic corpus-level stage. It
uses the repository's canonical `strings.Fields` token loader through
`internal/sequenceanalyze.ReadTokens`; it does not read IVTFF metadata,
Currier labels, hand IDs, folio labels, or document-language information.

For the ordered token stream, the stage records `V(n)`, the number of types
seen through adaptive checkpoints (`100`, `200`, `500`, then powers of two
from `1000`) and always includes the corpus length. No Voynich-specific corpus
length is embedded in the implementation.

The descriptive Heaps fit is:

```text
log(V(n)) = log(K) + beta * log(n)
```

Outputs include `K`, `beta`, R², SSE, maximum absolute log residual, and the
fitting range. A good fit is not evidence for a language class. Frequency of
frequencies is updated incrementally during the O(N), O(V)-memory observed
pass. Windowed new-type rates use corpus-independent defaults of 500, 1000,
and 2000 tokens. Positional segments use 4 and 8 contiguous segments when
possible.

The shuffled-token null preserves the complete token multiset and destroys
order. Each replicate uses a deterministic seed derived from the base seed
and logical replicate index; null results are reduced in index order. The
output reports null mean, SD, standardized effect, and empirical p-value at
each checkpoint.

The observed pass and default bounded null ensemble are currently profiled as
single-machine work. The CLI accepts `-executor goroutine` and explicitly
rejects process/remote until a measured study justifies distribution; no new
RPC implementation is introduced.

Vocabulary size, TTR, hapax and dis-legomena counts depend on corpus length,
tokenization and order. They do not establish natural language, cipher text,
generated text, or semantic class. Cross-corpus comparisons must use common
available checkpoints; unavailable checkpoints are not numeric zero.
