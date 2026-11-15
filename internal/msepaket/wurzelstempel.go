package msepaket

// ortsstempel ist, woran die [Wurzelsuche] erkennt, dass ein Ort derselbe
// geblieben ist — Befund 2 des Zweitblicks über Phase 0.7.0-I zu PERF-02
// (#2169).
//
// **Nicht die Änderungszeit allein.** Sie lässt sich zurücksetzen (`os.Chtimes`,
// `touch -r`, ein Auspacker, der Zeiten bewahrt), und ein entzogenes Leserecht
// ändert sie gar nicht. Deshalb gehören dazu:
//
//   - **die Statuszeit** (`ctime`) — sie setzt nur das System, und sie wechselt
//     bei jedem Umbenennen darin, bei jedem `chmod` und bei jedem Zurücksetzen
//     der Änderungszeit selbst,
//   - **die Rechte** — ein Ordner, der nicht mehr lesbar ist, gibt nichts mehr,
//   - **der Knoten** — ein Ordner, der durch einen anderen gleichen Namens
//     ersetzt wurde, ist ein anderer.
//
// **Wo es keine Statuszeit gibt** — außerhalb von darwin und linux, also nur
// in den Übersetzungen für Windows aus §12 —, bleiben Änderungszeit und Rechte.
// Gebaut und ausgeliefert wird dort nichts.
type ortsstempel struct {
	aenderung, status int64
	modus             uint32
	knoten            uint64
	da                bool
}
