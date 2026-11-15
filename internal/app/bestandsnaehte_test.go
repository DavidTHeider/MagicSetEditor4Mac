// **Kennzahlen, Rücknahme und Einstellungen für die Flächen ohne Fenster**
// (#1107).
package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Zahlennaht zählt aus.**
func TestDieZahlennahtZaehltAus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, _ := saatsetImBestand(t, db)

	weg := app.ZahlennahtUeber(db)

	statistik, fehler := weg.Statistik(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Statistik: %v", fehler)
	}

	if statistik.Karten == 0 {
		t.Errorf("keine Karte gezählt: %+v", statistik)
	}

	verteilung, verteilungsfehler := weg.Verteilung(probekontext(t), setID, "name")
	if verteilungsfehler != nil {
		t.Fatalf("Verteilung: %v", verteilungsfehler)
	}

	if verteilung == nil {
		t.Error("die Verteilung kam als nil — eine leere Liste wäre richtig")
	}
}

// **Und reicht ihre Lesefehler durch.**
func TestDieZahlennahtReichtIhreFehlerDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	weg := app.ZahlennahtUeber(db)

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE card_values`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := weg.Statistik(probekontext(t), 1); err == nil {
		t.Error("die Statistik ohne Tabelle kam durch")
	}

	if _, err := weg.Verteilung(probekontext(t), 1, "name"); err == nil {
		t.Error("die Verteilung ohne Tabelle kam durch")
	}
}

// **Die Verlaufsnaht bewegt in beide Richtungen.**
func TestDieVerlaufsnahtBewegtInBeideRichtungen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, karteID := saatsetImBestand(t, db)

	naht, nahtfehler := app.Schreibnaht(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Schreibnaht: %v", nahtfehler)
	}

	if _, _, err := naht.FeldSetzenMitVorwert(
		probekontext(t), setID, karteID, "name", "Blitz"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	weg := app.VerlaufsnahtUeber(db)

	zurueck, fehler := weg.Zuruecknehmen(probekontext(t), setID, "")
	if fehler != nil {
		t.Fatalf("Zuruecknehmen: %v", fehler)
	}

	if !zurueck.Bewegt {
		t.Fatalf("nichts bewegt, obwohl geschrieben wurde: %+v", zurueck)
	}

	vor, vorfehler := weg.Wiederherstellen(probekontext(t), setID, "")
	if vorfehler != nil {
		t.Fatalf("Wiederherstellen: %v", vorfehler)
	}

	if !vor.Bewegt {
		t.Errorf("nichts wiederhergestellt: %+v", vor)
	}
}

// **Und ein Fehler kommt durch.**
func TestDieVerlaufsnahtReichtIhreFehlerDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	weg := app.VerlaufsnahtUeber(db)

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE history`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := weg.Zuruecknehmen(probekontext(t), 1, ""); err == nil {
		t.Error("das Zurücknehmen ohne Tabelle kam durch")
	}

	if _, err := weg.Wiederherstellen(probekontext(t), 1, ""); err == nil {
		t.Error("das Wiederherstellen ohne Tabelle kam durch")
	}
}

// **Die Einstellungsnaht liest und schreibt.**
func TestDieEinstellungsnahtLiestUndSchreibt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg := app.EinstellungsnahtUeber(frischerBestand(t))

	alle, fehler := weg.Alle(probekontext(t))
	if fehler != nil {
		t.Fatalf("Alle: %v", fehler)
	}

	if len(alle) == 0 {
		t.Fatal("keine Einstellung — die Reihe kommt aus der Domäne")
	}

	// **Ein Wert, der nicht der aktuelle ist** — sonst prüfte der Vergleich
	// danach nichts.
	erste := alle[0]

	neu := erste.Werte[0]
	if neu == erste.Wert && len(erste.Werte) > 1 {
		neu = erste.Werte[1]
	}

	if err := weg.Setzen(probekontext(t), erste.Schluessel, neu); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	nachher, lesefehler := weg.Alle(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Alle: %v", lesefehler)
	}

	if nachher[0].Wert != neu {
		t.Errorf("die Einstellung steht bei %q statt bei %q", nachher[0].Wert, neu)
	}
}

// **Und ein Lesefehler kommt durch.**
func TestDieEinstellungsnahtReichtIhreFehlerDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	weg := app.EinstellungsnahtUeber(db)

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE settings`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := weg.Alle(probekontext(t)); err == nil {
		t.Error("das Lesen ohne Tabelle kam durch")
	}
}

// zaehlprobe scheitert wahlweise bei den Kennzahlen oder bei der Belegung.
//
// **Zwei Werte und nicht einer**: wer beides zusammenwirft, kommt gar nicht
// bis zur Belegung — die Kennzahlen scheiterten vorher, und der Weg dahinter
// bliebe ungeprüft. Im Betrieb hängen beide an denselben Tabellen, und über
// die Datenbank ist der zweite Weg nicht zu erreichen.
type zaehlprobe struct {
	kennzahlfehler  error
	belegungsfehler error
	werte           []domain.ValueCount
}

func (z zaehlprobe) Kennzahlen(context.Context, int64) (domain.SetStats, error) {
	return domain.SetStats{Cards: 3}, z.kennzahlfehler
}

func (z zaehlprobe) Belegung(context.Context, int64) ([]domain.FieldCount, error) {
	return []domain.FieldCount{{Field: "name", Cards: 3}}, z.belegungsfehler
}

func (z zaehlprobe) Verteilung(context.Context, int64, string) ([]domain.ValueCount, error) {
	return z.werte, nil
}

// **Ein Fehler beim Lesen der Belegung kommt durch.**
func TestEinBelegungsfehlerKommtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg := app.ZahlennahtUeber(frischerBestand(t)).
		MitZaehler(zaehlprobe{belegungsfehler: errors.New("die Belegung ist zu")})

	if _, err := weg.Statistik(probekontext(t), 1); err == nil {
		t.Error("der Belegungsfehler kam nicht durch")
	}
}

// **Und eine nichtleere Verteilung wird umgesetzt.**
//
// Ohne einen Wert bliebe die Schleife unbetreten, und der Test sagte nur, dass
// die Naht nicht abstürzt.
func TestEineNichtleereVerteilungWirdUmgesetzt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg := app.ZahlennahtUeber(frischerBestand(t)).MitZaehler(zaehlprobe{
		werte: []domain.ValueCount{{Value: "Feuerball", Cards: 2, CardIDs: []int64{3, 8}}},
	})

	verteilung, fehler := weg.Verteilung(probekontext(t), 1, "name")
	if fehler != nil {
		t.Fatalf("Verteilung: %v", fehler)
	}

	if len(verteilung) != 1 || verteilung[0].Name != "Feuerball" || verteilung[0].Karten != 2 {
		t.Errorf("die Verteilung kam verstümmelt an: %+v", verteilung)
	}

	// **Auch die Flächen ohne Fenster bekommen die Karten des Werts**
	// (UXR-F6, #2013) — dieselbe Zusage an jeder Fläche.
	if !slices.Equal(verteilung[0].KartenIDs, []string{"3", "8"}) {
		t.Errorf("die Karten lauten %v", verteilung[0].KartenIDs)
	}
}
