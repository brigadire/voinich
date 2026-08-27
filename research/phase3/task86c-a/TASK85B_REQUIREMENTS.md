# Task85b / G1-v2 requirements

Task85b is ready to be specified, but not because Task86C-a recovered a latent winner. It is ready because the audit localized an observability failure and retained several independent computational/protocol failure classes.

Required properties:

1. Persist every PM value, baseline, frozen threshold, finite/availability flag, and per-baseline gate outcome for every candidate selected into HELDOUT.
2. Treat PM6 construction availability, PM6 score validity, and PM6 acceptance as separate fields; validate saturated controls before confirmation.
3. Persist each F2 metric/family result, generation-scale result and `NOT_REACHED` reason.
4. Separate model evidence from induction caps, convergence failures, generation failures and protocol vetoes.
5. Require known-correct M0–M5 controls to survive an end-to-end pre-freeze audit, including scale and replicate stability.
6. Hash all per-job intermediate artifacts in the frozen result manifest and prove that decision paths can be regenerated without model execution.
7. Do not tune thresholds from these controls and do not treat CF1/CF2 as scientific classifications.

Concrete defects and validation tests are in `G1_V2_REQUIREMENTS.tsv`.
