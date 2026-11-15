package app

// **Die Partieverwaltung an der gebundenen Oberfläche** — PLAY-03.AK-05
// (#1711).
//
// **Ein innerer Lauf**, weil er beides braucht: `NeueBindungen` — dieselbe
// Naht, die das Fenster benutzt, samt Dateidialog — **und** die
// schweigsame Leitung, ohne die kein Partieweg eine Antwort hätte.

import (
	"context"
	"errors"
	"testing"
)

// partieprobe merkt sich, was an der Datei ankam.
type partiedateiprobe struct {
	pfad   string
	inhalt []byte
	fehler error
}

func (p *partiedateiprobe) Schreibe(pfad string, inhalt []byte) error {
	p.pfad, p.inhalt = pfad, inhalt

	return p.fehler
}

func (p *partiedateiprobe) Lies(pfad string) ([]byte, error) {
	p.pfad = pfad

	return []byte("{}"), p.fehler
}

// mitPartiedatei baut eine Anwendung mit Dateidialog und Datei.
func mitPartiedatei(
	gewaehlt string, wahlfehler error, probe *partiedateiprobe,
) *App {
	anwendung := NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return gewaehlt, wahlfehler
	})
	anwendung.MitPartiedatei(probe)

	// **Und eine Leitung dazu**: ohne sie hätte kein Partieweg eine
	// Gegenseite, und was hier geprüft wird, ist der Weg dorthin.
	return anwendung.MitPartienaht(&schweigsameLeitung{})
}

// **Der gewählte Ort geht an die Datei** — und ohne Bestand kommt danach
// ein Fehler, denn eine Partie gibt es dort nicht.
func TestPartieAusfuhrSpeichernWaehltDenOrt(t *testing.T) {
	t.Parallel()

	probe := &partiedateiprobe{}
	antwort := NeueBindungen(
		mitPartiedatei("/tmp/partie.msepartie", nil, probe)).
		PartieAusfuhrSpeichern(t.Context(), "p1")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Abgebrochen:
		t.Error("der Dialog galt als weggeklickt")
	case probe.pfad != "/tmp/partie.msepartie":
		t.Errorf("geschrieben wurde nach %q", probe.pfad)
	case string(probe.inhalt) != "{}":
		t.Errorf("in der Datei steht %q", probe.inhalt)
	}
}

// **Und was eingelesen wird, kommt aus der gewählten Datei.**
func TestPartieEinfuhrOeffnenLiestDieGewaehlteDatei(t *testing.T) {
	t.Parallel()

	probe := &partiedateiprobe{}
	antwort := NeueBindungen(
		mitPartiedatei("/tmp/partie.msepartie", nil, probe)).
		PartieEinfuhrOeffnen(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Kennung == "":
		t.Error("die eingelesene Partie hat keine Kennung")
	case probe.pfad != "/tmp/partie.msepartie":
		t.Errorf("gelesen wurde aus %q", probe.pfad)
	}
}

// **Ein weggeklickter Dialog ist ein Abbruch und kein Fehler** — bei beiden
// Wegen, und das ist dieselbe Unterscheidung wie bei `SetSpeichernUnter`.
func TestDiePartiedialogeUnterscheidenAbbruchVonFehler(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		gewaehlt    string
		wahlfehler  error
		abgebrochen bool
		fehler      bool
	}{
		"weggeklickt":  {gewaehlt: "", abgebrochen: true},
		"Dialog klagt": {wahlfehler: errors.New("kein Fenster"), fehler: true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bindungen := NeueBindungen(mitPartiedatei(
				fall.gewaehlt, fall.wahlfehler, &partiedateiprobe{}))

			for weg, antwort := range map[string]PartieverwaltungAntwort{
				"Ausfuhr": bindungen.PartieAusfuhrSpeichern(t.Context(), "p1"),
				"Einfuhr": bindungen.PartieEinfuhrOeffnen(t.Context()),
			} {
				if antwort.Abgebrochen != fall.abgebrochen {
					t.Errorf("%s: abgebrochen=%v", weg, antwort.Abgebrochen)
				}

				if (antwort.Fehler != nil) != fall.fehler {
					t.Errorf("%s: Fehler=%v", weg, antwort.Fehler)
				}
			}
		})
	}
}

// **Und die beiden Wege ohne Dialog gehen durch** — sie stehen auf
// `Bindungen`, weil ein Mensch sie an derselben Stelle sucht.
func TestDieVerwaltungswegeOhneDialogGehenDurch(t *testing.T) {
	t.Parallel()

	bindungen := NeueBindungen(
		mitPartiedatei("", nil, &partiedateiprobe{}))

	if antwort := bindungen.PartieLoeschen(t.Context(), "p1"); antwort.Fehler != nil {
		t.Errorf("das Löschen meldete %v", antwort.Fehler)
	}

	// **Die Kennung der Kopie kommt aus der Naht** — und sie ist eine neue.
	if antwort := bindungen.PartieKopieren(t.Context(), "p1"); antwort.Fehler != nil ||
		antwort.Kennung == "" || antwort.Kennung == "p1" {
		t.Errorf("das Kopieren gab %+v", antwort)
	}
}
