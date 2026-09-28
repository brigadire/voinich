# Reproduction

Run from the repository root:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v2/prepare.py
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s research/astro_hapax_star_label/restricted_dictionary_bruteforce_v2 -p 'test_*.py' -v
PYTHONDONTWRITEBYTECODE=1 python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v2/run.py freeze
PYTHONDONTWRITEBYTECODE=1 python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v2/run.py run
```

Preparation refuses to overwrite frozen v2 files. v1 and all upstream outputs are
read-only inputs. The production runner uses deterministic checkpoints and a lock;
it refuses modified frozen bytes and resumes only the same seed/grid manifest.
