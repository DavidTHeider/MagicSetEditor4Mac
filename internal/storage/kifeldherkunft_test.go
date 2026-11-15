package storage_test

// **Werte und Herkunft in einem Zug** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Die Zusage, um die es hier geht
//
// Das Issue: „**Scheitert Provenienz/Verlauf in derselben Transaktion, stehen
// keine halben Inhaltsänderungen.**"
//
// Sie lässt sich nur **hier** prüfen. `service` bekommt beides in einem Ruf
// und sieht einen Fehler; ob darunter eine Transaktion liegt, sagt allein die
// Datenbank — und der Weg, das zu messen, ist eine Herkunft, die sich nicht
// schreiben lässt.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Beides steht danach da.**
func TestWerteUndHerkunftStehenBeideDa(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t),
		karteID, []domain.Value{
			{Field: "name", Text: "Blitzschlag"},
			{Field: "text", Text: "3 Schaden"},
		},
		"ki", []domain.KIFeldherkunft{
			eineFeldherkunft("name"), eineFeldherkunft("text"),
		}, nil, ""); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	gelesen, fehler := storage.NeueFeldherkunft(db).Lesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Zwei Zeilen, nach Feldname geordnet** — eine Karte hat keine
	// Reihenfolge, und eine Herkunft, die bei jedem Lauf anders steht, ist
	// nicht zu vergleichen.
	if len(gelesen) != 2 || gelesen[0].Feld != "name" ||
		gelesen[1].Feld != "text" {
		t.Fatalf("gelesen wurde %+v", gelesen)
	}

	if gelesen[0].Modell != "gemma4:8b-q4" ||
		gelesen[0].Aufgabe != "Schreibe den Regeltext." {
		t.Errorf("die Herkunft ist %+v", gelesen[0])
	}

	// **Der Zeitpunkt kommt als Zeitpunkt zurück** — nicht als Zeichenkette,
	// die irgendwo anders noch einmal gedeutet wird.
	if !gelesen[0].Zeitpunkt.Equal(derHerkunftstag) {
		t.Errorf("der Zeitpunkt ist %v", gelesen[0].Zeitpunkt)
	}
}

// **Scheitert die Herkunft, steht auch der Wert nicht da.**
//
// # Wie hier gemessen wird
//
// **Mit einer Herkunft, die die Datenbank abweist**: ein Feldname, der
// zweimal vorkommt, verletzt den Primärschlüssel. Die Werte davor sind zu
// diesem Zeitpunkt schon geschrieben — läge keine Transaktion darunter,
// stünde der neue Name da und die Herkunft nicht.
func TestEineScheiterndeHerkunftNimmtDenWertMit(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	_, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t), karteID,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{
			eineFeldherkunft("name"), eineFeldherkunft("name"),
		}, nil, "")
	if fehler == nil {
		t.Fatal("die doppelte Herkunft ging durch")
	}

	// **Der Wert steht noch, wie er war.**
	karten, lesefehler := storage.NeuerSetLeser(db).KarteLesen(
		probekontext(t), karteID)
	if lesefehler != nil {
		t.Fatalf("die Karte lesen: %v", lesefehler)
	}

	if wert := wertVon(*karten, "name"); wert != "Flammenball" {
		t.Errorf("das Feld trägt %q — die Änderung wurde nicht zurückgenommen",
			wert)
	}

	// **Und es steht keine Herkunft da.**
	gelesen, _ := storage.NeueFeldherkunft(db).Lesen(probekontext(t), karteID) //nolint:errcheck // unten geprüft
	if len(gelesen) != 0 {
		t.Errorf("es steht Herkunft da: %+v", gelesen)
	}
}

// **Eine zweite Übernahme ersetzt die erste.**
//
// **Weil am Feld ein Wert steht und nicht zwei.** Zwei Herkünfte nebeneinander
// ließen offen, welche gilt.
func TestEineZweiteUebernahmeErsetztDieHerkunft(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	for _, modell := range []string{"gemma4:8b-q4", "grok-4"} {
		herkunft := eineFeldherkunft("name")
		herkunft.Modell = modell

		if _, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t),
			karteID, []domain.Value{{Field: "name", Text: "Blitzschlag"}},
			"ki", []domain.KIFeldherkunft{herkunft}, nil, ""); fehler != nil {
			t.Fatalf("schreiben mit %s: %v", modell, fehler)
		}
	}

	gelesen, fehler := storage.NeueFeldherkunft(db).Lesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 || gelesen[0].Modell != "grok-4" {
		t.Fatalf("gelesen wurde %+v", gelesen)
	}
}

// **Und eine Übernahme ohne Herkunft ist keine.**
//
// **Das Issue sagt »verpflichtend«** — und eine Zusage, die man weglassen
// kann, ist keine. Wer ohne schreiben will, nimmt den gewöhnlichen Weg.
func TestEineUebernahmeOhneHerkunftWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	if _, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitHerkunft(
		probekontext(t), karteID,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", nil, nil, ""); fehler == nil {
		t.Fatal("die Übernahme ohne Herkunft ging durch")
	}
}

// derHerkunftstag ist der Zeitpunkt, den diese Läufe schreiben.
var derHerkunftstag = time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)

// eineFeldherkunft baut eine vollständige Herkunft für ein Feld.
func eineFeldherkunft(feld string) domain.KIFeldherkunft {
	return domain.KIFeldherkunft{
		Feld: feld, Anbieter: "ollama", Modell: "gemma4:8b-q4",
		Aufgabe: "Schreibe den Regeltext.", Auftrag: "ki-7-1",
		Zeitpunkt: derHerkunftstag,
		Bezug:     domain.KIInhaltsbezug([]string{"Blitzschlag"}),
	}
}

// wertVon liest ein Feld aus einer Kartenzeile.
func wertVon(zeile domain.CardRow, feld string) string {
	for _, wert := range zeile.Values {
		if wert.Field == feld {
			return wert.Text
		}
	}

	return ""
}

// **Ohne Profil im Kontext wird nichts geschrieben.**
//
// **Weil die Herkunft je Profil steht.** Ein Ruf ohne Profil wüsste nicht,
// wessen Herkunft er anlegt — und eine Zeile unter einem geratenen Profil
// wäre schlimmer als keine.
func TestOhneProfilWirdKeineHerkunftGeschrieben(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	if _, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitHerkunft(
		t.Context(), karteID,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{eineFeldherkunft("name")}, nil, ""); fehler == nil {
		t.Fatal("ohne Profil wurde geschrieben")
	}
}

// **Und ohne Profil wird auch nichts gelesen.**
func TestOhneProfilWirdKeineHerkunftGelesen(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := storage.NeueFeldherkunft(db).Lesen(t.Context(), 4); fehler == nil {
		t.Fatal("ohne Profil wurde gelesen")
	}
}

// **Eine unvollständige Herkunft wird abgewiesen — vor dem Schreiben.**
//
// **Weil sie eine Zusage ist.** Eine Zeile ohne Modell sagt „aus einem
// Vorschlag" und lässt offen, aus welchem; das ist weniger als die Marke am
// Feld und sieht nach mehr aus.
func TestEineUnvollstaendigeHerkunftWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	halbe := eineFeldherkunft("name")
	halbe.Modell = ""

	if _, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitHerkunft(
		probekontext(t), karteID,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{halbe}, nil, ""); fehler == nil {
		t.Fatal("die halbe Herkunft ging durch")
	}

	// **Und der Wert steht noch, wie er war** — alles oder nichts.
	karten, _ := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID) //nolint:errcheck // unten geprüft
	if wert := wertVon(*karten, "name"); wert != "Flammenball" {
		t.Errorf("das Feld trägt %q", wert)
	}
}

// **Ein unlesbarer Zeitpunkt im Bestand wird gemeldet.**
//
// **Weil er von Hand dort stehen kann.** Der Bestand ist eine Datei, die
// jemand weitergibt; was darin steht, ist nicht immer, was dieses Programm
// geschrieben hat. Eine Herkunft ohne deutbare Zeit ist keine — und „von
// heute" zu antworten, wäre eine Erfindung.
func TestEinUnlesbarerZeitpunktInDerHerkunftWirdGemeldet(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	if _, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitHerkunft(
		probekontext(t), karteID,
		[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{eineFeldherkunft("name")}, nil, ""); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`UPDATE ki_feldherkunft SET zeitpunkt = 'gestern'`); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if _, fehler := storage.NeueFeldherkunft(db).Lesen(
		probekontext(t), karteID); fehler == nil {
		t.Fatal("der unlesbare Zeitpunkt kam als Herkunft zurück")
	}
}

// **Ein verbogener Bestand wird gemeldet und nicht gedeutet.**
//
// # Warum diese Läufe an der Datei ansetzen
//
// **Weil der Bestand eine Datei ist, die jemand weitergibt.** Was darin
// steht, ist nicht immer, was dieses Programm geschrieben hat — das ist
// dieselbe Überlegung wie bei fremden Paketen (§8). Jeder dieser Fälle endet
// mit einer Meldung; keiner endet mit einer erfundenen Herkunft.
func TestEinVerbogenerHerkunftsbestandWirdGemeldet(t *testing.T) {
	faelle := map[string]struct {
		verbiegen string
		lesen     bool
	}{
		"die Tabelle fehlt beim Schreiben": {
			`DROP TABLE ki_feldherkunft`, false,
		},
		"die Tabelle fehlt beim Lesen": {
			`DROP TABLE ki_feldherkunft`, true,
		},
		"ein Feld steht leer da": {
			`DROP TABLE ki_feldherkunft;
			 CREATE TABLE ki_feldherkunft (
			   profil_id INTEGER, karte_id INTEGER, feld TEXT,
			   anbieter TEXT, modell TEXT, aufgabe TEXT, auftrag TEXT,
			   zeitpunkt TEXT, bezug TEXT);
			 INSERT INTO ki_feldherkunft VALUES
			   (1, 1, 'name', 'ollama', NULL, 'x', 'ki-1', '2026-09-12T09:00:00Z', 'sha256:x')`,
			true,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db := frischeDatenbank(t)
			karteID := eineKarte(t, db,
				domain.Value{Field: "name", Text: "Flammenball"})

			if _, fehler := db.ExecContext(probekontext(t),
				fall.verbiegen); fehler != nil {
				t.Fatalf("verbiegen: %v", fehler)
			}

			if fall.lesen {
				if _, fehler := storage.NeueFeldherkunft(db).Lesen(
					probekontext(t), karteID); fehler == nil {
					t.Fatal("der verbogene Bestand kam als Herkunft zurück")
				}

				return
			}

			if _, _, fehler := storage.NeuerKartenschreiber(db).
				FelderSetzenMitHerkunft(probekontext(t), karteID,
					[]domain.Value{{Field: "name", Text: "Blitzschlag"}},
					"ki", []domain.KIFeldherkunft{
						eineFeldherkunft("name"),
					}, nil, ""); fehler == nil {
				t.Fatal("in den verbogenen Bestand wurde geschrieben")
			}
		})
	}
}
