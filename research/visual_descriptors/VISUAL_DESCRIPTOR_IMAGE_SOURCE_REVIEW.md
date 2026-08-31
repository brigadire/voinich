# Image-source review

`IMAGE_SOURCE_FROZEN=true` for source selection; source bytes are not redistributed.

## Primary source: Yale IIIF

The authoritative source is Yale University Library / Beinecke Rare Book and Manuscript Library, *Cipher manuscript* (Beinecke MS 408), object OID `2002046`.

- Catalog: `https://collections.library.yale.edu/catalog/2002046`
- IIIF Presentation 3 manifest: `https://collections.library.yale.edu/manifests/2002046`
- Retrieved: 2026-08-31
- Retrieved manifest SHA-256: `317d58fd9ea90392a83d9858a91eada3d0b41416a3c835857dc0154bd123a309`
- Completeness: manifest metadata says “Completely digitized”; 213 canvases were present. They include 206 manuscript page/spread/foldout views plus covers, pastedowns and physical-edge views.
- Resolution: canvas dimensions range from ordinary single pages around 2,700–3,000 by 3,700–3,800 pixels to foldouts up to 9,078 pixels wide and 7,268 pixels high. Full-resolution JPEG locators are recorded per annotation unit.
- Recto/verso/panels: ordinary recto/verso sides normally have one canvas. Yale sometimes supplies a combined spread or foldout and sometimes several “part” canvases. Exact IVTFF panel IDs therefore require an explicit crop mapping; the manifest alone does not assert that mapping.
- Versioning: Yale supplies stable object/canvas OIDs but no release tag. The retrieved manifest hash is the source snapshot identifier.
- Rights: the manifest's current Rights metadata warns that use may be subject to US copyright, site-license, or other terms and places responsibility on the user. This project links to, but does not redistribute, image bytes. Yale's catalog and rights statement remain authoritative.

The annotation viewing cache used 900-pixel-wide IIIF derivatives in `/tmp`; it is not a deliverable. Scientific provenance always points to the full-resolution IIIF URI.

## Alternatives audited

| Source | Mapping/completeness | Resolution/version | Reuse decision |
|---|---|---|---|
| Yale catalog viewer/download | Same OID and content as primary IIIF; full work | High resolution; catalog state changes in place | Suitable human interface, but IIIF is more reproducible |
| Beinecke pre-1600 MS 408 description | Stable scholarly shelfmark and physical description; not a page API | No systematic per-unit image contract | Metadata corroboration only |
| Internet Archive/Wikimedia PDF derivatives | Page-oriented derivative, commonly 214 PDF pages | Resampled derivative; edition/version may differ | Rejected for primary coding to avoid edition mixing |
| voynich.nu gallery/folio browser | Useful scholarly page/panel nomenclature | Third-party derivatives and separate rights conditions | Mapping aid only; no image bytes used |

No images from different editions were mixed.

## Audit outcome

The image source is frozen to Yale OID 2002046. The exact unit is frozen to the IVTFF page/panel ID, but panel crops remain an open mapping/adjudication item. Consequently `ANNOTATION_UNIT_FROZEN=true` as a definition while full panel annotation is not yet ready.
