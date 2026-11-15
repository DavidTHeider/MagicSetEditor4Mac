package storage_test

// **Der eingereichte Deckabzug überlebt das Schreiben** — Befund 1 des
// Zweitblicks über Phase 4.7
// (#1294).
//
// # Der Befund
//
// `service.Turnierleitung.Anmelden` baute den Abzug und gab ihn zurück; die
// Tabelle `turnierdeckplaetze` hatte **weder Schreiber noch Leser**. Die
// Antwort auf das Anmelden trug ihn, jedes spätere Lesen nicht mehr.
//
// **Das ist derselbe Fehler wie `SideboardOffen` in Phase 4.6** — ein Feld,
// das nur eine Schicht kennt. Dort wurde er vom Nachlauf gefunden, hier vom
// Zweitblick, und beide Male war die Pipeline grün.

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Angemeldet, wieder gelesen — und die Liste ist noch da.**
func TestEinDeckabzugUeberlebtDasSchreiben(t *testing.T) {
	ablage := turnierablage(t)

	turnier, err := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	abzug := []domain.Abzugsplatz{
		{
			Bezeichnung: "Lightning Bolt", Menge: 4,
			Bereich: domain.BereichHauptdeck, Stelle: 1,
		},
		{
			Printing: "a1b2c3d4", Menge: 2,
			Bereich: domain.BereichSideboard, Stelle: 2,
		},
	}

	if _, anmeldefehler := ablage.Anmelden(probekontext(t), turnier.ID,
		domain.Teilnehmer{Name: "Ada", Deckabzug: abzug}); anmeldefehler != nil {
		t.Fatalf("anmelden: %v", anmeldefehler)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Teilnehmer) != 1 {
		t.Fatalf("%d Teilnehmer, erwartet einen", len(gelesen.Teilnehmer))
	}

	gefunden := gelesen.Teilnehmer[0].Deckabzug

	if len(gefunden) != 2 {
		t.Fatalf("%d Zeilen im Abzug, erwartet 2: %+v", len(gefunden), gefunden)
	}

	// **In ihrer Reihenfolge** — eine Deckliste ist eine Liste.
	if gefunden[0].Bezeichnung != "Lightning Bolt" || gefunden[0].Menge != 4 ||
		gefunden[0].Bereich != domain.BereichHauptdeck {
		t.Errorf("die erste Zeile ist %+v", gefunden[0])
	}

	if gefunden[1].Printing != "a1b2c3d4" ||
		gefunden[1].Bereich != domain.BereichSideboard {
		t.Errorf("die zweite Zeile ist %+v", gefunden[1])
	}
}

// **Ein Teilnehmer ohne Deck hat eine leere Liste — und keine `nil`.**
//
// Der Unterschied trägt bis an die Oberfläche: `null` und `[]` sind in JSON
// zwei verschiedene Antworten, und eine Seite, die über `null` iteriert,
// bricht ab.
func TestEinTeilnehmerOhneDeckHatEineLeereListe(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada")

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Teilnehmer[0].Deckabzug == nil {
		t.Error("der Abzug ist nil, erwartet eine leere Liste")
	}
}

// **Und er entsteht in einem Zug mit dem Teilnehmer.**
//
// **Die Gegenprobe zur Transaktion**: ohne sie stünde der Teilnehmer im
// Bestand und seine Liste nicht — ein eingereichtes Dokument, von dem die
// Hälfte fehlt.
func TestEinHalberAbzugEntstehtNicht(t *testing.T) {
	ablage := turnierablage(t)

	turnier, err := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Eine Zeile, die der `CHECK` abweist**: weder Bezeichnung noch Druck.
	_, err = ablage.Anmelden(probekontext(t), turnier.ID, domain.Teilnehmer{
		Name: "Ada",
		Deckabzug: []domain.Abzugsplatz{
			{Bezeichnung: "Gut", Menge: 1, Bereich: domain.BereichHauptdeck},
			{Menge: 1, Bereich: domain.BereichHauptdeck},
		},
	})
	if err == nil {
		t.Fatal("angenommen, erwartet abgewiesen")
	}

	gelesen, lesefehler := ablage.Turnier(probekontext(t), turnier.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen.Teilnehmer) != 0 {
		t.Errorf("%d Teilnehmer im Bestand — erwartet keinen", len(gelesen.Teilnehmer))
	}
}

// **Die Fehlerwege des Abzugs.**
//
// **Sie brauchen die Stolpertreiber**: gegen eine echte Datei scheitert weder
// `LastInsertId` noch ein Scan über `STRICT`-Spalten.
func TestDerAbzugReichtSeineFehlerDurch(t *testing.T) {
	for _, fall := range []struct {
		name    string
		treiber string
	}{
		{"das Lesen der Listen scheitert", "stolper-abzug-abfrage"},
		{"eine Zeile der Liste ist unlesbar", "stolper-abzug-scan"},
		{"die Kennung einer Zeile ist unlesbar", "stolper-abzug-kennung"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			ablage := storage.NeueTurnierablage(db)

			if fall.treiber == "stolper-abzug-kennung" {
				_, fehler := ablage.Anmelden(probekontext(t), 1, domain.Teilnehmer{
					Name: "Ada",
					Deckabzug: []domain.Abzugsplatz{
						{
							Bezeichnung: "Gut", Menge: 1,
							Bereich: domain.BereichHauptdeck,
						},
					},
				})
				stolperfehlerPruefen(t, fehler)

				return
			}

			_, fehler := ablage.Turnier(probekontext(t), 1)
			stolperfehlerPruefen(t, fehler)
		})
	}
}

// stolperfehlerPruefen ist `pruefeStolperfehler` für das äußere Testpaket:
// ein Fehler muss kommen, und **nicht** der an der Spaltenzahl der
// Vorrichtung — sonst prüfte der Lauf etwas anderes als seinen Fall
// (Nachlauf des Zweitblicks über Phase B).
func stolperfehlerPruefen(t *testing.T, fehler error) {
	t.Helper()

	switch {
	case fehler == nil:
		t.Error("der Fehler ging verloren")
	case strings.Contains(fehler.Error(), "destination arguments"):
		t.Errorf("gescheitert an der Spaltenzahl der Vorrichtung, nicht an der Stelle: %v", fehler)
	}
}
