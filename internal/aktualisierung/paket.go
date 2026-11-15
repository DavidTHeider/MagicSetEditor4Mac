// Package aktualisierung sagt, ob eine neuere Fassung dieser Anwendung
// vorliegt — und lädt sie auf Wunsch herunter.
//
// # Ein Blatt (AGENTS.md §16)
//
// Es kennt `internal/domain` und `internal/netz` und sonst nichts aus diesem
// Projekt. Drei Schichten, und sie werden nicht vermischt:
//
//   - **Transport** — HTTP, Fristen, Drossel, Zwischenspeicher: `internal/netz`
//   - **Fremdes DTO** — die Gestalt, die GitHub schickt, mit seinen Feldnamen
//   - **Zuordnung** — die Übersetzung auf das, was diese Anwendung kennt
//
// # Was gilt, und wann es nachgelesen wurde
//
// **Am 2026-08-25 unter `https://docs.github.com/en/rest`**, nicht erinnert
// (§16, Satz 1):
//
//	unangemeldet         60 Anfragen je Stunde
//	mit Marke         5.000 Anfragen je Stunde
//	x-ratelimit-remaining   was noch übrig ist
//	x-ratelimit-reset       wann das Fenster neu beginnt, in UTC-Sekunden
//
// **`User-Agent` ist Pflicht**, und zwar wörtlich: „All API requests must
// include a valid `User-Agent` header. […] Requests with no `User-Agent`
// header will be rejected." `Accept: application/vnd.github+json` und
// `X-GitHub-Api-Version` sind empfohlen, nicht verlangt.
//
// # Die Vorgabe ist „aus"
//
// Nichts geschieht ohne ausdrückliche Einstellung — §16, Regel 6: nichts wird
// still übernommen, und ein Programm, das beim ersten Start ungefragt ins Netz
// greift, ist eines, das etwas will.
package aktualisierung
