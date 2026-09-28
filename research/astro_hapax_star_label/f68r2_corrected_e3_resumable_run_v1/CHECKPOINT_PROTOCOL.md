# Transactional checkpoint protocol

Each profile is generated in `profiles/.Sxxx.tmp.<run_id>/`. The profile is eligible for atomic rename only after solver completion, independent replay, PROFILE_STATUS write, and SHA256SUMS creation. A committed profile is immutable. Resume skips only profiles whose status is COMMITTED and whose every checksum verifies. Invalid or incomplete temporary directories are never evidence.
