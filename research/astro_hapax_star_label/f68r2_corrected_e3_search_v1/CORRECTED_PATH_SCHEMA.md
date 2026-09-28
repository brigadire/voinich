# Corrected path schema

Each generated path stores profile_id, LABEL occurrence, EVA token type, canonical identity and lexicon row, source form, transformed source stream, required mappings, forbidden source units, expected EVA output, and scorer trace. A selected profile uses one global mapping table. Replay runs DROP_UNMAPPED on the complete transformed source stream; any extra output invalidates the assignment.
