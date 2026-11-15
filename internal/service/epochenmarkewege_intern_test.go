// **Die Wege der Epochenmarke über den Rückweg hinaus** — Befunde 7 und 8 des
// Nachlaufs zum Zweitblick über Phase 5.2 (#1435).
//
// Aus `epochenmarke_intern_test.go` herausgelöst, als es über 400 Zeilen ging
// (§2). Der Schnitt folgt der Frage: dort steht, **was die Marke ist** — wann
// sie entsteht, was sie trägt, wann sie geht —, hier, **was mit ihr geschieht,
// wenn etwas schiefgeht**: derselbe Start zweimal, ein klemmendes Buch, ein
// Rückweg und eine Marke, die sich nicht wegräumen lässt.

package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Dieselbe Marke gibt eine Epoche und nicht zwei.**
//
// Befund 8 des Nachlaufs zum Zweitblick über Phase 5.2: ließ sich die Marke
// nach dem Vermerken nicht wegräumen, legte der nächste Start eine zweite
// Epoche an — aus **einer** Rückkehr wurden zwei Sperren.
func TestDieselbeMarkeGibtNurEineEpoche(t *testing.T) {
	weg, buch := wegMitMarke(t)

	jetzt := time.Date(2026, time.September, 11, 8, 0, 0, 0, time.UTC)

	schreiben := func() {
		t.Helper()

		if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
			Datenstand: jetzt, HandpruefungNoetig: true,
		}, "probe.zip", jetzt); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	schreiben()

	if rest := weg.EpocheNachholen(t.Context()); rest != "" {
		t.Fatalf("der erste Start: %q", rest)
	}

	// **Die Marke kommt zurück** — so, wie sie nach einem gescheiterten
	// Wegräumen liegen bliebe.
	schreiben()

	if rest := weg.EpocheNachholen(t.Context()); rest != "" {
		t.Fatalf("der zweite Start: %q", rest)
	}

	if len(buch.epochen) != 1 {
		t.Errorf("im Buch stehen %d Epochen", len(buch.epochen))
	}
}

// **Und ein Buch, das beim Nachsehen klemmt, meldet sich.**
func TestEinKlemmendesBuchMeldetSichBeimNachholen(t *testing.T) {
	weg, buch := wegMitMarke(t)

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
		HandpruefungNoetig: true,
	}, "probe.zip", time.Now()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	buch.fehler = errKeinBuch

	if rest := weg.EpocheNachholen(t.Context()); rest == "" {
		t.Error("es kam kein Satz fürs Fenster zurück")
	}
}

// **Ein Rückweg nimmt die Marke mit.**
//
// Befund 7 des Nachlaufs: sie steht **vor** dem Tausch. Wer sie nach einem
// gelungenen Rückweg liegen ließe, erzeugte beim nächsten Start eine Epoche
// über eine Rückkehr, die es nicht gibt — der Bestand von vorher steht wieder
// da, und an ihm ist nichts unklar.
func TestEinRueckwegNimmtDieEpochenmarkeMit(t *testing.T) {
	bau := neuerWegbau(t)

	mitStaenden(bau.weg, bau.buch,
		besitz(4, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 4}),
		besitz(1, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1}),
	)

	sicherung := probesicherung(t, bau)

	// **Das Wiederöffnen scheitert** — der Weg dreht zurück.
	bau.leben.oeffnungsfehler = errKeinBuch

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), sicherung); fehler == nil {
		t.Fatal("es kam durch")
	}

	if _, err := os.Stat(bau.weg.epochenmarkenpfad()); !os.IsNotExist(err) {
		t.Fatalf("die Marke blieb liegen: %v", err)
	}

	// **Und der nächste Start trägt nichts nach.**
	bau.leben.oeffnungsfehler = nil

	if rest := bau.weg.EpocheNachholen(t.Context()); rest != "" {
		t.Errorf("der Start meldet %q", rest)
	}

	if len(bau.buch.epochen) != 0 {
		t.Errorf("im Buch stehen %d Epochen", len(bau.buch.epochen))
	}
}

// **Eine Marke, die sich nicht wegräumen lässt, hält nichts auf.**
//
// Die beiden Wege, auf denen `epochenmarkeEntfernen` **scheitern** kann, sind
// bis zum 2026-09-11 von keinem Lauf betreten worden — sie kamen mit Befund 7
// des Nachlaufs dazu, und der Beleg dafür fehlte (§3).
//
// **Zwei Wege, zwei Aufbauten**, und beide brauchen, dass nichts unklar ist:
// sonst scheiterte schon `epochenmarkeSchreiben`, und der Lauf endete, bevor
// das Entfernen je liefe — ein Lauf, der aus dem falschen Grund grün ist (§10).
func TestEineKlemmendeMarkeHaeltDenRueckwegNichtAuf(t *testing.T) {
	for name, aufbau := range map[string]func(*testing.T, *wegbau){
		// **Der Tausch scheitert**, weil die ausgepackte Datei zwischen dem
		// Schließen und dem Einsetzen verschwindet — und an der Stelle der
		// Marke steht ein Verzeichnis, das sich nicht wegräumen lässt.
		//
		// **Über den Haken beim Schließen**, weil es keinen anderen Ort gibt:
		// was das Verzeichnis unbrauchbar machte, träfe schon das Auspacken.
		"der Tausch scheitert": func(t *testing.T, bau *wegbau) {
			t.Helper()

			markeUnentfernbar(t, bau.weg)

			bau.leben.beimSchliessen = func() {
				zwischenplatzLeeren(t, filepath.Dir(bau.weg.orte.Datenbank))
			}
		},

		// **Das Wiederöffnen scheitert** — dann dreht der Weg zurück, und
		// dabei geht die Marke mit. Hier steht an ihrer Stelle ein
		// **Verzeichnis**: `os.Remove` gibt darauf `ENOTEMPTY`, und das ist
		// kein `ErrNotExist`.
		"das Wiederöffnen scheitert": func(t *testing.T, bau *wegbau) {
			t.Helper()

			bau.leben.oeffnungsfehler = errKeinBuch
			markeUnentfernbar(t, bau.weg)
		},
	} {
		t.Run(name, func(t *testing.T) {
			bau := neuerWegbau(t)

			gleich := domain.Sammlungsmenge{
				Profil: 1, Schluessel: "scryfall:abc", Menge: 4,
			}

			mitStaenden(bau.weg, bau.buch, besitz(4, gleich), besitz(4, gleich))

			sicherung := probesicherung(t, bau)

			aufbau(t, bau)

			if _, fehler := bau.weg.Zurueckspielen(t.Context(), sicherung); fehler == nil {
				t.Fatal("es kam durch")
			}

			// **Die Marke steht noch da** — und der Weg hat es gemeldet,
			// statt daran zu scheitern.
			if _, err := os.Stat(bau.weg.epochenmarkenpfad()); err != nil {
				t.Errorf("die Marke ist fort: %v", err)
			}
		})
	}
}

// markeUnentfernbar setzt ein **nicht leeres Verzeichnis** an ihre Stelle.
func markeUnentfernbar(t *testing.T, weg *Bestandsweg) {
	t.Helper()

	pfad := weg.epochenmarkenpfad()

	if fehler := os.MkdirAll(pfad, 0o750); fehler != nil {
		t.Fatalf("das Verzeichnis anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(
		filepath.Join(pfad, "darin.txt"), []byte("x"), 0o600); fehler != nil {
		t.Fatalf("die Datei darin anlegen: %v", fehler)
	}
}

// zwischenplatzLeeren nimmt die ausgepackte Datenbank weg.
//
// **Zwischen Schließen und Tausch**: genau dann liegt sie da, und genau dann
// bringt ihr Verschwinden `bestandErsetzen` zum Scheitern.
func zwischenplatzLeeren(t *testing.T, ort string) {
	t.Helper()

	eintraege, fehler := os.ReadDir(ort)
	if fehler != nil {
		t.Fatalf("den Ort lesen: %v", fehler)
	}

	for _, eintrag := range eintraege {
		if !strings.HasPrefix(eintrag.Name(), ".rueckspiel-") {
			continue
		}

		if err := os.RemoveAll(filepath.Join(ort, eintrag.Name())); err != nil {
			t.Fatalf("den Zwischenplatz wegnehmen: %v", err)
		}

		return
	}

	t.Fatalf("kein Zwischenplatz unter %s", ort)
}
