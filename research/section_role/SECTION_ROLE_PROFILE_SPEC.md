metric_id	definition	aggregation_unit	normalization	missingness_policy
freq_norm	count/section_tokens	token-section	section token total	observed
page_presence	pages_with_token/pages	token-section	section pages	observed
leaf_presence	leaves_with_token/leaves	token-section	section leaves	observed
line_initial_rate	token occurrences at line index 0 / occurrences	token-section	occurrences	observed
line_final_rate	token occurrences at final line index / occurrences	token-section	occurrences	observed
internal_rate	remaining occurrences / occurrences	token-section	occurrences	observed
mean_line_position	mean index/(line_length-1)	token-section	occurrences	observed
first_third_rate	index fraction < 1/3	token-section	occurrences	observed
final_third_rate	index fraction >= 2/3	token-section	occurrences	observed
self_repeat_rate	adjacent same-token pairs / token occurrences	token-section	line transitions	observed
near_repeat_rate	adjacent edit-distance-1 pairs / token occurrences	token-section	line transitions	observed
predecessor_entropy	H(predecessor|token,section)	token-section	line contexts	observed
successor_entropy	H(successor|token,section)	token-section	line contexts	observed
predecessor_neff	exp(predecessor_entropy)	token-section	line contexts	observed
successor_neff	exp(successor_entropy)	token-section	line contexts	observed
position_entropy	entropy of thirds	token-section	occurrences	observed
