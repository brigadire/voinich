# Decomposed E3 search

All 64 frozen operation profiles were solved independently on the f68r2-only graph. The profile-globality gate passed: a profile is selected once for the whole assignment, and no assignment may mix operation profiles.

Results:

- profiles: 64/64 `OPTIMAL`;
- decision `coverage >= 6`: certified infeasible for all 64 profiles;
- certified global maximum: 5/27;
- best profile: `S043` (`DROP_AL`, `IJ_UV`, `KEEP`, `STRIP_LATIN`);
- null trigger: not reached, so null control was not run;
- f68r1 and group crosswalk: unused.

The best profile uses six active mapping rules, each supported by at least two distinct EVA token types. Its five matches are retained with complete operation traces in `E3_BEST_ASSIGNMENT.tsv`.

Because every profile has an exact maximum and every profile is infeasible at 6, the global E3 maximum is certified for this frozen graph and operation family. This is still not an astronomical correlation claim; the preregistered null was not triggered because the maximum is below 6/27.
