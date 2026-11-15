# Wanderungen nach der Grundlinie

**Bis 0097 gibt es keine Wanderungen mehr** — entschieden am 2026-10-12,
Vermerk `docs/entscheidungen/0189-nur-die-grundlinie.md` (J02, #2266). Ein
leerer Bestand entsteht in einem Zug aus `../grundlinie/`; welche Wanderungen
sie enthält, steht mit Name und Prüfsumme in `../grundlinie/umfang.txt`.

**Eine Änderung am Schema ist eine neue Wanderung hier**, beginnend bei
`0098_…sql` — nicht eine Änderung der Grundlinie. Ein Bestand, der über die
alte Grundlinie entstand, sähe sonst anders aus als einer über die neue, und
beide trügen dieselbe Buchführung. `TestDieGrundlinieIstUnveraendert` hält das.

**Die erste ist `0098_sicherungsvermerke.sql`** (M01,
#2345),
**die zweite `0099_setstand.sql`** (PERF02,
#2352): ein
Änderungszähler je Set, gehoben von Auslösern; **die dritte
`0100_setstand_genauer.sql`** ersetzt fünf davon und legt drei dazu — die
Stiloptionen, der Ort der Bilder, und gehoben wird nur noch, was sich ändert
(Zweitblick über Phase 3, Befunde 1 und 14); **die vierte,
`0101_gegenbelegbefunde.sql`** (#2500), hält einen Gegenbeleg fest, der
nicht passt. **Wer eine Zahl der
Wanderungen nennt**, liest sie aus diesem Ordner
(`scripts/huellenprobe/aufstiegsteile.mjs`) — eine feste Zahl bricht mit der
nächsten.
**Eine Probe in einem Test** bringt eine Wanderung mit der nächsten freien
Nummer mit und legt sie über die heutigen (`storage.MigriereMit`,
`kuenftigeWanderung`) — nie eine feste Nummer, sonst bricht die nächste echte
Wanderung jede Probe.
