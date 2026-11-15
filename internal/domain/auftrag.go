package domain

// **Hintergrundaufgaben** — Werkstück U11 von Phase 5.1
// (#1422).
//
// Ein **Auftrag** ist etwas, das länger läuft als ein Klick: Preise für die
// Sammlung holen, die Sammeldaten laden, eine neue Fassung holen. Die
// Übersicht zeigt, was gerade läuft und was zuletzt fertig wurde — aus einer
// gemeinsamen Naht (`service.Auftragsbuch`), nicht aus je einem Zähler.

import "time"

// Auftragsstand ist, wo ein Auftrag steht.
type Auftragsstand string

// Die vier Stände — **abgebrochen** seit FMT-03 (#1692): ein Abbruch ist
// kein Scheitern, und er trägt keinen Grund.
const (
	AuftragLaeuft      Auftragsstand = "laeuft"
	AuftragFertig      Auftragsstand = "fertig"
	AuftragGescheitert Auftragsstand = "gescheitert"
	AuftragAbgebrochen Auftragsstand = "abgebrochen"
)

// Auftrag ist ein Vorgang mit Anfang, Stand und Ende.
type Auftrag struct {
	ID int64
	// Art nennt die Sorte — `preise`, `sammeldaten`, `fassung` und seit
	// FMT-03 die Datei-Läufe des Fensters: `bogen`, `set`, `mtgjson`,
	// `vorlage`, `installer`, `weitergabe`. Ein technischer Schlüssel; das
	// Wort dazu setzt das Fenster in seiner Sprache.
	Art string
	// Was nennt den Gegenstand — die Datei, die Fassung — oder bleibt leer,
	// wenn die Art alles sagt. **Kein Satz**: der stand hier bis FMT-03 auf
	// Deutsch und erschien so auch im englischen Fenster.
	Was string
	// Fortschritt ist der letzte gemeldete Stand — frei formuliert.
	Fortschritt string
	// Gezaehlt ist der letzte gezählte Stand eines Laufs, der seine Stücke
	// kennt — das Fenster setzt ihn in seiner Sprache.
	Gezaehlt Fortschritt
	// Abbrechbar sagt, ob der laufende Auftrag sich anhalten lässt.
	Abbrechbar bool
	Stand      Auftragsstand
	// Grund ist, woran ein gescheiterter Auftrag scheiterte.
	Grund    string
	Begonnen time.Time
	Beendet  time.Time
}

// Dauer sagt, wie lange der Auftrag lief oder läuft.
func (a Auftrag) Dauer(jetzt time.Time) time.Duration {
	if a.Stand == AuftragLaeuft {
		return jetzt.Sub(a.Begonnen)
	}

	return a.Beendet.Sub(a.Begonnen)
}
