package storage

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Fehlerwege des Paketkatalogs.
//
// Sie sind mit einer echten SQLite-Datenbank nicht herzustellen: was die
// Domänenprüfung durchlässt, schreibt sich auch. Im Betrieb entstehen sie an
// einer vollen Platte oder einer abgebrochenen Verbindung — und ein Fehlerpfad,
// den kein Test betritt, ist eine Behauptung darüber, was im Ernstfall
// geschieht (§3).

// einPaket gibt eine Katalogliste mit einem Paket ohne Abhängigkeit.
func einPaket() []domain.Paket {
	return []domain.Paket{{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"}}
}

// mitAbhaengigkeit gibt eine Katalogliste mit einer Abhängigkeit.
func mitAbhaengigkeit() []domain.Paket {
	return []domain.Paket{{
		Art: "mse-style", Name: "probe-schlicht", Pfad: "/p/probe-schlicht.mse-style",
		Braucht: []domain.Bedarf{{Paket: "probe.mse-game"}},
	}}
}

func TestAbgleichenMeldetDieFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
		pakete  []domain.Paket
	}{
		{"die Katalogzeile scheitert", "stolper-fragen", einPaket()},

		// **Mit einer Abhängigkeit**, seit `bedarfSchreiben` nur noch bei
		// einer Abweichung räumt (#1325): ohne eine gibt es nichts zu räumen,
		// und der erste Schreibvorgang wäre das Entfernen der verwaisten
		// Zeilen — der Fall darunter.
		{"das Räumen der Abhängigkeiten scheitert", "stolper-schreiben", mitAbhaengigkeit()},
		{"eine Abhängigkeit scheitert", "stolper-zweites-schreiben", mitAbhaengigkeit()},

		// **Der erste Schreibvorgang, wenn keine Abhängigkeit dasteht.**
		{"das Räumen der verwaisten Zeilen scheitert", "stolper-schreiben", einPaket()},
		{"das Commit scheitert", "stolper-commit", einPaket()},

		// **Die vier Wege, die #1325 hinzugefügt hat.** Ohne sie wäre jeder
		// eine Behauptung darüber, was im Ernstfall geschieht (§3).
		{"der Schreibzähler scheitert", "stolper-schreibzaehler", einPaket()},
		{"das Lesen der Abhängigkeiten scheitert", "stolper-bedarf-lesen", mitAbhaengigkeit()},
		{"das Auslesen einer Abhängigkeit scheitert", "stolper-bedarf-scan", mitAbhaengigkeit()},
		{"der zweite Schreibzähler scheitert", "stolper-zweiter-zaehler", einPaket()},
		{"das Nachlesen der Kennung scheitert", "stolper-kennung-lesen", einPaket()},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0
			zaehlerstand = 0
			zaehleraufrufe = 0

			fehler := NeuerPaketkatalog(db).Abgleichen(probekontext(t), fall.pakete)
			if !errors.Is(fehler, errGestolpert) {
				t.Errorf("Fehler %v, erwartet war der Stolperfehler", fehler)
			}
		})
	}
}

// **Eine geschlossene Datenbank hat keine Transaktion.**
//
// Der einzige Fehlerweg, für den es keinen Stolpertreiber braucht — und der im
// Betrieb wirklich vorkommt: nach `Beenden` läuft noch ein Aufruf.
func TestAbgleichenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db, err := Oeffnen(probekontext(t), ImArbeitsspeicher)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := NeuerPaketkatalog(db).Abgleichen(probekontext(t), einPaket()); fehler == nil {
		t.Error("kein Fehler")
	}
}

func TestPaketeMeldetDieFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
	}{
		{"die Abfrage scheitert", "stolper-fragen"},
		{"eine Zeile ist unlesbar", "stolper-umwandeln"},
		{"ein Lesefehler nach der Schleife", "stolper-lesen"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeuerPaketkatalog(db).Pakete(probekontext(t)); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

func TestPaketEntfernenMeldetEinenSchreibfehler(t *testing.T) {
	db, err := sql.Open("stolper-schreiben", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	fehler := NeuerPaketkatalog(db).PaketEntfernen(probekontext(t), "mse-style", "weg")
	if !errors.Is(fehler, errGestolpert) {
		t.Errorf("Fehler %v, erwartet war der Stolperfehler", fehler)
	}
}

// **Die Fehlerwege des Ortswechsels** (#226).
//
// Er zählt die betroffenen Zeilen, um „das Set gibt es nicht" von „geschrieben"
// zu trennen; beide Wege dahinter sind mit einer echten Datenbank nicht
// herzustellen.
func TestHerkunftSetzenMeldetDieFehlerpfade(t *testing.T) {
	faelle := map[string]string{
		"das Schreiben scheitert": "stolper-schreiben",
		"die Zählung scheitert":   "stolper-zaehlen",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			fehler := NeuerSetLeser(db).HerkunftSetzen(probekontext(t), 1, "/neu")
			if !errors.Is(fehler, errGestolpert) {
				t.Errorf("Fehler %v, erwartet war der Stolperfehler", fehler)
			}
		})
	}
}

// **Ein Wert, den `Scan` nicht umwandeln kann** (#1325).
//
// **Ein eigener Lauf und keine Zeile in der Tabelle darüber**: der Fehler kommt
// von `database/sql` und nicht vom Stolpertreiber — er ist deshalb **nicht**
// `errGestolpert`, und die Zusicherung dort träfe nicht zu. Was hier zählt, ist
// dass der Weg überhaupt einen Fehler meldet, statt eine leere Liste für bare
// Münze zu nehmen.
func TestAbgleichenMeldetEinenUnlesbarenBedarf(t *testing.T) {
	db, err := sql.Open("stolper-bedarf-typ", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	zweiterSchreibzaehler = 0
	zaehlerstand = 0
	zaehleraufrufe = 0

	if fehler := NeuerPaketkatalog(db).Abgleichen(probekontext(t), mitAbhaengigkeit()); fehler == nil {
		t.Error("kein Fehler, obwohl eine Abhängigkeit nicht zu lesen war")
	}
}
