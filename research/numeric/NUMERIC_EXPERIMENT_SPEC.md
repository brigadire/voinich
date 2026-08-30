# Numeric experiment specification

Frozen primary corpus: data_work/ZL3b-x7.canonical.txt. Literal byte characters are used; lowercase ASCII
letters a-z are admitted if observed, and every token containing another
symbol is excluded before looking at numeric results. The base is the admitted
inventory size. Digits are assigned by ascending byte order for baseline.

Primary families are SEQUENTIAL (absolute lag-1 Spearman), DIFFERENCE (mean of
local AP closeness and repeated normalized-delta fraction), DOCUMENT (mean of
absolute line-position/value Spearman and folio eta-squared), and EDIT (exact
same-position substitution identity rate). NUMERIC_REGULARITY_SCORE is the
unweighted mean of SEQUENTIAL, DIFFERENCE, and DOCUMENT. EDIT is reported but
excluded from the score because its identity follows algebraically from the
imposed representation.

Mapping search is seeded simulated annealing: 2 restarts, 250 proposed digit
swaps per restart, objective evaluated on every eighth physical line, followed
by full-corpus evaluation. The identical procedure is applied separately to
every matched control. Controls: C1 within-token shuffle; C2 token shuffle
within physical line; C3 first-order glyph Markov generation preserving token
lengths and physical-line layout. Replicates=40; root seed=20260829. Empirical upper
tail p=(1+#null>=observed)/(R+1); registered family/control comparisons use
BH-FDR together. Fixed modular probes: 2,3,4,5,7,8,10,12.

IVTFF alignment preserves folio, section, locus type and physical line. Since
no geometric coordinates are used, document/layout analysis is **2D-LITE**.
