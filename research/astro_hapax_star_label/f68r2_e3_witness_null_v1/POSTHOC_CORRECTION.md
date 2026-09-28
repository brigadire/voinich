# Replay correction

The initial apparent witness was withdrawn. Its path-local rule signatures were individually valid, but applying their union as one global table failed exact scorer replay. The corrected runner now applies the complete union mapping to every transformed source before accepting a witness. The corrected 100-replicate result contains zero valid witnesses.
