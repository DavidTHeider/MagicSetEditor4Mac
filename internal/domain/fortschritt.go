package domain

// **Wie weit ein Auftrag ist** — FMT-03.AK-02
// (#1692).
//
// Ein Druckbogen über dreihundert Karten, eine MTG-JSON-Datei mit vierzig
// Sets, ein Installer mit zwölf Paketen: was länger läuft, sagt unterwegs, wo
// es steht. **Eine Gestalt für alle Flächen** — MCP schickt daraus
// `notifications/progress`, der Dienst trägt sie am Auftrag, die
// Kommandozeile schreibt sie in eine Zeile und das Fenster ins Auftragsbuch.
// Wer sie hört, steht im Kontext (`service.MitFortschritt`).
//
// **Eine Einheit und kein Satz**: der Satz entsteht an der Fläche, in ihrer
// Sprache. [Fortschritt.Satz] ist der deutsche für die drei Flächen ohne
// Fenster, die deutsch sprechen (§9).

import "fmt"

// Fortschritt ist ein Zwischenstand.
type Fortschritt struct {
	// Erledigt zählt, was fertig ist — nie weniger als beim vorigen Stand.
	Erledigt int
	// Gesamt ist, wie viel es wird; `0` heißt: nicht bekannt.
	Gesamt int
	// Einheit nennt, was gezählt wird — [EinheitKarte] und die übrigen.
	Einheit string
}

// Fortschrittsmelder hört einen Zwischenstand.
type Fortschrittsmelder func(Fortschritt)

// Die Einheiten, in denen gezählt wird — technische Schlüssel, keine Wörter.
const (
	EinheitKarte   = "karte"
	EinheitSet     = "set"
	EinheitPaket   = "paket"
	EinheitBild    = "bild"
	EinheitSchritt = "schritt"
)

// einheitsnamen sind die deutschen Mehrzahlformen — nach „von" im Dativ.
var einheitsnamen = map[string]string{
	EinheitKarte:   "Karten",
	EinheitSet:     "Sets",
	EinheitPaket:   "Paketen",
	EinheitBild:    "Bildern",
	EinheitSchritt: "Schritten",
}

// einheitsnamenOhneGesamt stehen, wo kein „von" davor steht.
var einheitsnamenOhneGesamt = map[string]string{
	EinheitKarte:   "Karten",
	EinheitSet:     "Sets",
	EinheitPaket:   "Pakete",
	EinheitBild:    "Bilder",
	EinheitSchritt: "Schritte",
}

// Satz gibt den Stand als deutschen Satz — „12 von 300 Karten".
func (f Fortschritt) Satz() string {
	if f.Gesamt <= 0 {
		return mitEinheit(fmt.Sprintf("%d", f.Erledigt), einheitsnamenOhneGesamt[f.Einheit])
	}

	return mitEinheit(fmt.Sprintf("%d von %d", f.Erledigt, f.Gesamt), einheitsnamen[f.Einheit])
}

func mitEinheit(zahl, einheit string) string {
	if einheit == "" {
		return zahl
	}

	return zahl + " " + einheit
}
