# Results report

All 64 frozen E3 profiles were processed through the transactional runner. Every profile has an atomic committed checkpoint, independently replay-validated assignment, and verified checkpoint hashes. There are 64 unique canonical profile graphs.

Forty profiles have exact `OPTIMAL` maxima. Twenty-four profiles remain `UNKNOWN` after their bounded solver runs. The best valid incumbent is coverage 3, reached by the profiles listed in `GLOBAL_BEST_PROFILES.tsv`. Because unknown profiles remain and their global upper bounds are not sufficient to close the family, `GLOBAL_E3_MAXIMUM` is not established.

The maximum is below the predeclared threshold 4 among currently validated incumbents, therefore the null requirement remains undetermined until the UNKNOWN profiles are closed; it is not authorized in this run. No scientific interpretation or decryption claim is made.
