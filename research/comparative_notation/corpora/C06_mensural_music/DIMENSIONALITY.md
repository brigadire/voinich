# C06 dimensionality and token justification

MUSIC-R1 treats one source event as a token because it is the smallest encoded
ordered musical occurrence; its symbols are frozen notational components.
MUSIC-R2 treats the interval between consecutive pitched events in the same
voice as a token; rests break interval continuity. MUSIC-R3 makes the joint
pitch-duration state explicit. Systems are physical lines only when the source
encodes layout. Simultaneous events share `simultaneity_group`; voice/staff are
attributes and may support a separately registered D view, but are never
discarded during linearization.
