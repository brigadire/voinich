#!/usr/bin/env python3
def run(valid_index):
    draws=0
    for attempt in range(1024):
        # Synthetic productive branch: one channel plus prefix/stem/suffix.
        draws+=4
        if attempt==valid_index: return "OK",attempt,attempt+1,draws
    return "GENERATION_FAILURE",1023,1024,draws
assert run(0)==("OK",0,1,4)
assert run(1023)==("OK",1023,1024,4096)
assert run(None)==("GENERATION_FAILURE",1023,1024,4096)
print("M5_ATTEMPTS_AND_COUNTERS=PASS")

