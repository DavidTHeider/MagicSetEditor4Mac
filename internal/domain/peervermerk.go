package domain

// **Was über die eigene Peerkennung feststeht** — Werkstück N1 von Phase 5.4
// (#1454).

import "time"

// Peervermerk ist, was der Bestand über die Identität eines Profils hält.
//
// **Die Kennung und nicht der Schlüssel.** Der private Schlüssel liegt im
// Schlüsselbund; hier steht nur, welche Kennung zuletzt galt — und genau das
// macht einen **Verlust** sichtbar: ohne den Vermerk wäre „der Bund hat
// nichts" nicht von „hatte noch nie einen" zu unterscheiden.
type Peervermerk struct {
	// Kennung ist die öffentliche Peerkennung — leer heißt: es gab noch keine.
	Kennung string

	// Seit sagt, wann sie entstanden ist.
	Seit time.Time

	// Vorher ist die Kennung, die vor der letzten Rotation galt — leer, wenn
	// es keine gab.
	//
	// **Ein Kontakt kennt sie.** Wer rotiert, verlangt von jedem bekannten
	// Gegenüber eine neue Prüfung; ohne die alte Kennung könnte ihm niemand
	// sagen, welche abgelöst wurde.
	Vorher string
}

// Gibts sagt, ob überhaupt eine Kennung vermerkt ist.
func (p Peervermerk) Gibts() bool { return p.Kennung != "" }

// Peerstand sagt, wie es um die Identität eines Profils steht.
//
// **Vier Stände, und nur einer heißt „alles in Ordnung".** Drei davon sind
// kein Fehler: es gab noch keine, sie ist verloren, sie widerspricht sich —
// jeder verlangt eine andere Handlung, und keiner darf still behoben werden.
type Peerstand string

const (
	// PeerstandOhne heißt: dieses Profil hatte noch keine Identität.
	PeerstandOhne Peerstand = "ohne"

	// PeerstandBereit heißt: Schlüssel und Vermerk sagen dasselbe.
	PeerstandBereit Peerstand = "bereit"

	// PeerstandVerloren heißt: der Vermerk steht, der Schlüssel fehlt.
	//
	// **Der häufigste Weg dorthin ist eine kopierte Bestandsdatei** — sie
	// bringt die Kennung mit, der fremde Rechner hat den Schlüssel nicht.
	PeerstandVerloren Peerstand = "verloren"

	// PeerstandFremd heißt: beide stehen da und meinen nicht dasselbe.
	//
	// **Anders als ein Verlust ist das kein Fehlen, sondern ein
	// Widerspruch** — und der wird nicht aufgelöst, indem man sich für eine
	// Seite entscheidet.
	PeerstandFremd Peerstand = "fremd"
)
