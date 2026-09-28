# Writing System Model Specification (v3 Design)

## 1. Formal Foundations

This specification defines the family of global writing system transformations $\mathcal{M}$ evaluated in the Restricted Dictionary Brute-Force v3 design. Every candidate model $M \in \mathcal{M}$ represents a global orthographic and cipher hypothesis mapping historical star names into the observed EVA script.

## 2. Mathematical Definition of the Model Class

A model $M$ is defined as a tuple:
$$M = \langle T, \text{mode}, \text{deletion}, \text{abbreviation} \rangle$$

### 2.1. Source and Target Alphabets
- **Source Alphabet ($\Sigma_{\text{src}}$)**: Contains 29 graphemes:
  $$\Sigma_{\text{src}} = \{a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, t, u, v, w, x, y, z\} \cup \{\text{kh}, \text{gh}, \text{sh}, \text{th}, \text{dh}\}$$
- **Target Alphabet ($\Sigma_{\text{EVA}}$)**: The closed 16-symbol inventory observed across the 57 star labels:
  $$\Sigma_{\text{EVA}} = \{a, c, d, e, f, h, i, k, l, n, o, p, r, s, t, y\}$$

### 2.2. Substitution Table ($T$)
The table $T$ is a partial function:
$$T: \text{dom}(T) \subseteq \Sigma_{\text{src}} \to \Sigma_{\text{EVA}}$$
where $|T| = k \in \{4, 6, 8, 10, 12\}$.

Mapping modes:
1. **`INJECTIVE`**: A strict one-to-one function:
   $$\forall s_1, s_2 \in \text{dom}(T), \quad s_1 \neq s_2 \implies T(s_1) \neq T(s_2)$$
2. **`MERGE_1`**: A bounded many-to-one function with exactly one collision pair:
   $$|\text{dom}(T)| = k, \quad |\text{range}(T)| = k - 1$$
   Exactly two distinct source graphemes $s_1 \neq s_2$ share the same target symbol $T(s_1) = T(s_2)$, while all other $k - 2$ mappings are injective.

### 2.3. Deletion Rules
Defines the treatment of characters $c \notin \text{dom}(T)$:
1. **`NONE` (KEEP)**: Unmapped characters remain unchanged in the output string:
   $$\text{enc}(c) = c$$
   *(Note: Excluded from primary production search due to inability to produce EVA-only tokens).*
2. **`DROP_UNMAPPED`**: All unmapped characters are deleted:
   $$\text{enc}(c) = \epsilon$$
   Subject to mandatory dictionary collision penalty and complexity cost ($+2$).
3. **`SELECTIVE_VOWEL_DROP`**: Unmapped root consonants remain unmapped, but unmapped vowels ($\{a, e, i, o, u, y\}$) are dropped:
   $$\text{enc}(v) = \epsilon \quad \text{for } v \in \{a, e, i, o, u, y\} \setminus \text{dom}(T)$$
   Complexity cost ($+1$).

### 2.4. Abbreviation Rules
Applied post-substitution to model medieval scribal suspension and truncation:
1. **`NONE`**: Identity; no abbreviation applied ($\text{cost} = 0$).
2. **`SUSPENSION_1`**: Truncates the final grapheme if length exceeds 3 ($\text{cost} = 1$).
3. **`SUSPENSION_2`**: Truncates the final two graphemes if length exceeds 4 ($\text{cost} = 2$).
4. **`PREFIX_4`**: Retains exactly the first 4 characters if longer ($\text{cost} = 2$).

## 3. Strict Invariance Constraints

To maintain scientific falsifiability, the following constraints are strictly enforced:

1. **Global Table Consistency**: One identical table $T$ governs every star name across the entire dictionary. No per-term exceptions are permitted.
2. **No Per-Label Exceptions**: Every observed label is evaluated against the candidate set under identical transformation rules.
3. **Cross-Page Identity**: When validating across folios f68r1 and f68r2, the concrete mapping table $T$ must be **identical**. Family-level equivalence (e.g. "both are injective") is not sufficient.
4. **Order Invariance**: The search objective and selected optimum must be mathematically invariant to arbitrary permutations of the input label list.

## 4. Objective Function and Complexity Regularization

The global fitness function balances matching coverage against transformation complexity:

$$\text{Fitness}(M, \mathcal{L}) = 100 \times \text{Matched}(M, \mathcal{L}) - \text{Complexity}(M)$$

where $\text{Matched}(M, \mathcal{L})$ is the cardinality of the maximum bipartite matching between the transformed lexicon and the target labels $\mathcal{L}$ under the `PER_PAGE_CAPACITY_1` constraint, and:

$$\text{Complexity}(M) = |T| + \text{Cost}(\text{mode}) + \text{Cost}(\text{deletion}) + \text{Cost}(\text{abbreviation})$$

Cost schedule:
- $|T| \in \{4, 6, 8, 10, 12\}$: Table size directly penalizes model degrees of freedom.
- $\text{Cost}(\text{INJECTIVE}) = 0$, $\text{Cost}(\text{MERGE_1}) = 1$.
- $\text{Cost}(\text{NONE}) = 0$, $\text{Cost}(\text{SELECTIVE_VOWEL_DROP}) = 1$, $\text{Cost}(\text{DROP_UNMAPPED}) = 2$.
- $\text{Cost}(\text{NONE}) = 0$, $\text{Cost}(\text{SUSPENSION_1}) = 1$, $\text{Cost}(\text{SUSPENSION_2}) = 2$, $\text{Cost}(\text{PREFIX_4}) = 2$.
