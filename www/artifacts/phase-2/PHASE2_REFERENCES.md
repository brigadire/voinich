# Phase II references

This bibliography is assembled only from repository source registers and the Phase I bibliography. Repository artifacts are cited in the report and indexed separately. Items whose full text or metadata were not completely verified during the source task are marked accordingly.

## Primary historical sources

- Fontana, Giovanni. *Secretum de thesauro experimentorum ymaginationis hominum*, ca. 1430. Paris, Bibliothèque nationale de France, NAL 635. Facsimile: Gallica `ark:/12148/btv1b10023795x`. Primary authority for images, layout, and manuscript readings; surviving copy is not an autograph.
- Burchard of Worms. *Decretum*, Köln, Erzbischöfliche Diözesan- und Dombibliothek, Cod. 119, books 6/7/11/12/13. Burchards Dekret Digital TEI edition, commit `29f9cb1c34cc9ee3c50e75a6e3e99cfa4a2bc362`, CC BY 4.0. Paired `<abbr>/<expan>` source used in Task82b.

## Fontana editions and historical studies

- Kranz, Horst, ed. and trans. *Methoden des Erinnerns und Vergessens: Johannes Fontanas Secretum…* Boethius 68. Franz Steiner, 2016. ISBN 978-3-515-11583-4. Full paid edition was not wholly inspected in Task74; cited readings were checked through available portions and manuscript/source triangulation.
- Battisti, Eugenio, and Giuseppa Saccaro Battisti. *Le macchine cifrate di Giovanni Fontana…* Milan: Arcadia, 1984. ISBN 88-85684-06-8. Earlier full decipherment; readings require checking against Kranz/manuscript.
- Omont, Henri. “Un traité de physique et d’alchimie du XVe siècle en écriture cryptographique.” *Bibliothèque de l’École des chartes* 58 (1897): 253–258. DOI `10.3406/bec.1897.447898`.
- Cacopardo, Valentina. *Memory and Imagination in the Ars Memorativa in Fifteenth-Century Italy*. PhD dissertation, Warburg Institute, 2021.
- Kranz, Horst, and Walter Oberschelp. *Mechanisches Memorieren und Chiffrieren um 1430*. Boethius 59, 2009. ISBN 978-3-515-09296-8.
- Hogendijk, J. P. Review of Kranz and Oberschelp, *Mechanisches Memorieren und Chiffrieren um 1430*. *Mathematical Reviews* MR2531390, 2010.
- Muccillo, Maria. “Fontana, Giovanni.” *Dizionario Biografico degli Italiani* 48, 1997.
- Long, Pamela O. *Openness, Secrecy, Authorship*. Johns Hopkins University Press, 2001, 110–111.
- Bolzoni, Lina. *The Gallery of Memory*. University of Toronto Press, 2001, 101–103, 142–143.
- Yates, Frances A. *The Art of Memory*. Routledge & Kegan Paul, 1966. Used only for broad context, not specific Fontana instructions.

## Voynich and transcription sources

- Zandbergen–Landini ZL3b IVTFF transcription, repository source `data/ZL3b-n.txt`; provenance and SHA-256 are bound by the Task83b deterministic manifest.
- Takahashi IT2a transliteration via Jorge Stolfi’s 1999 interlinear file, IVTFF 2.0, repository source `data/IT2a-n.txt`; provenance and SHA-256 are bound by the Task83b deterministic manifest.
- Currier, Prescott H. “Some important new statistical findings.” In *New Research on the Voynich Manuscript: Proceedings of a Seminar*, 1976.
- Landini, Gabriel. “Evidence of linguistic structure in the Voynich manuscript using spectral analysis.” *Cryptologia* 25(4) (2001): 275–295. DOI `10.1080/0161-110191889932`.
- Reddy, Sravana, and Kevin Knight. “What We Know About The Voynich Manuscript.” Proceedings of the 5th ACL-HLT Workshop on Language Technology for Cultural Heritage, 2011.
- Montemurro, Marcelo A., and Damián H. Zanette. “Keywords and Co-Occurrence Patterns in the Voynich Manuscript.” *PLOS ONE* 8(6) (2013): e66344. DOI `10.1371/journal.pone.0066344`.
- Bowern, Claire L., and Luke Lindemann. “The Linguistics of the Voynich Manuscript.” *Annual Review of Linguistics* 7 (2021): 285–308. DOI `10.1146/annurev-linguistics-011619-030613`.
- Lindemann, Luke, and Claire Bowern. “Character Entropy in Modern and Historical Texts: Comparison Metrics for an Undeciphered Manuscript.” arXiv:2010.14697, 2021.
- Bowern, Claire, and Daniel E. Gaskell. “Enciphered after all? Word-level text metrics are compatible with some types of encipherment.” Proceedings of the 2022 International Voynich Conference.
- Matlach, V., B. A. Janečková, and D. Dostál. “The Voynich manuscript: Symbol roles revisited.” *PLOS ONE* 17 (2022): e0260948. DOI `10.1371/journal.pone.0260948`.

## Historical shorthand and selective-extraction sources

- Cappelli, Adriano. *Lexicon Abbreviaturarum*. 2nd ed., 1912. Reference lexicon, not running-text frequency evidence.
- Scho, Michael. *Abbreviationes*. GitHub/Zenodo DOI `10.5281/zenodo.16628612`. Repository candidate-control provenance; underlying terms require verification before reuse beyond the frozen BDD extraction.
- Rugg, Gordon. “An Elegant Hoax? A Possible Solution to the Voynich Manuscript.” *Cryptologia* 28(1) (2004): 31–46. DOI `10.1080/0161-110491892753`.
- Zandbergen, René. “The Cardan grille approach to the Voynich MS taken to the next level.” arXiv:2104.12548, 2021. Preprint; no released code recorded.
- Clérice, Thibault, et al. “CATMuS Medieval.” ICDAR 2024. Distributional medieval control; no systematic expansion alignment.
- Clérice, T., et al. “CoMMA.” LREC 2026. Publication metadata and final pagination require verification; no aligned expansions.
- Parisel, C. “Evidence of Layered Positional and Directional Constraints in the Voynich Manuscript.” arXiv:2604.19762, 2026. Preprint; no independent replication recorded.

## Relevant mechanism research

- Timm, Torsten, and Andreas Schinner. “A possible generating algorithm of the Voynich manuscript.” *Cryptologia* (2019). DOI `10.1080/01611194.2019.1596999`.
- Greshko, Michael A. “The Naibbe cipher: a substitution cipher that encrypts Latin and Italian as Voynich Manuscript-like ciphertext.” *Cryptologia* (2025). DOI `10.1080/01611194.2025.2566408`.
- Rozanova, Liudmila, and Alexander Temerev. “A Glyph Is Not a Letter, a Token Is Not a Word, a Space Is Not a Space.” arXiv:2608.17096, 2026.

## Methodological/statistical literature

Phase II’s concrete statistical procedures—empirical permutation tests, Monte Carlo nulls, Benjamini–Hochberg correction, bootstrap stability, family-balanced distance, Pareto comparison, matched nulls, and deterministic seeded reconstruction—are defined and implemented by the frozen repository protocols. No external methodological citation with verified repository provenance was necessary for a numerical claim in this synthesis. A future publication bibliography should add canonical method citations only after bibliographic verification; none is invented here.

Source provenance: [Fontana sources](../../../research/phase2/fontana/SOURCES.md), [notation-audit sources](../../../research/phase2/notation-audit/SOURCES.md), [BDD provenance](../../../research/phase2/task82b/SHORTHAND_CORPUS_PROVENANCE.tsv), and [Phase I bibliography](../../literature/bibliography.bib).
