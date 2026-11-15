# Hüllenprobe: e05start

**bestanden** — Commit `f1e733bde`,
2026-10-11 17:23 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-11T17-22-55Z-61621`, Bauabdruck `9e7ac4212fa5`.


**Saat — jedes Mal ein frisches Zuhause**

- ✓ saat 1: ⌘Q beendet
- ✓ saat 2: ⌘Q beendet
- ✓ saat 3: ⌘Q beendet
- ✓ saat 4: ⌘Q beendet
- ✓ saat 5: ⌘Q beendet

**warm — dasselbe Zuhause**

- ✓ erster Start des warmen Zuhauses: ⌘Q beendet
- ✓ warm 1: ⌘Q beendet
- ✓ warm 2: ⌘Q beendet
- ✓ warm 3: ⌘Q beendet
- ✓ warm 4: ⌘Q beendet
- ✓ warm 5: ⌘Q beendet

**älter — ein Bestand von 0.7.0**

- ✓ älter 1: ⌘Q beendet
- ✓ älter 2: ⌘Q beendet
- ✓ älter 3: ⌘Q beendet
- ✓ älter 4: ⌘Q beendet
- ✓ älter 5: ⌘Q beendet

**Die Reihen**

| Reihe | Starts | bedienbar: Median · max | erste Eingabe (⌘K): Median · max | Kernweg (⌘2): Median · max |
| --- | ---: | ---: | ---: | ---: |
| saat | 5 | 507 · 586 ms | 19 · 23 ms | 12 · 13 ms |
| warm | 5 | 504 · 511 ms | 19 · 19 ms | 22 · 25 ms |
| älter | 5 | 492 · 502 ms | 29 · 30 ms | 14 · 14 ms |

Rohwerte — {"saat":[{"bereit":570,"eingabe":23,"weg":11},{"bereit":586,"eingabe":23,"weg":13},{"bereit":494,"eingabe":17,"weg":12},{"bereit":504,"eingabe":19,"weg":12},{"bereit":507,"eingabe":17,"weg":12}],"warm":[{"bereit":511,"eingabe":19,"weg":11},{"bereit":501,"eingabe":18,"weg":22},{"bereit":507,"eingabe":18,"weg":24},{"bereit":504,"eingabe":19,"weg":25},{"bereit":484,"eingabe":19,"weg":14}],"älter":[{"bereit":423,"eingabe":30,"weg":14},{"bereit":492,"eingabe":29,"weg":13},{"bereit":426,"eingabe":29,"weg":14},{"bereit":501,"eingabe":30,"weg":14},{"bereit":502,"eingabe":29,"weg":12}]}
- ✓ bedienbar im Median ≤ 2 s, in jeder Reihe
- ✓ die erste Eingabe wartet auf kein Vorwärmen (≤ 500 ms, jeder Start) — [23,23,17,19,17,19,18,18,19,19,30,29,29,30,29]
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
