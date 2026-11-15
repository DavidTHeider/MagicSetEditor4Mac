package storage

// **Ein Altbestand bekommt die Sicherungsvermerke** — M01
// (#2345).
// Im Paket und nicht daneben, weil sich nur hier der Stand von 0.8.0 ohne
// Umweg herstellen lässt: die Grundlinie und nichts danach. Gegen eine Datei,
// die 0.8.0 selbst geschrieben hat, läuft derselbe Weg in `internal/app`.

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinAltbestandBekommtDieSicherungsvermerke(t *testing.T) {
	db, err := Oeffnen(probekontext(t), "file:"+filepath.Join(t.TempDir(), "alt.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := migriereAusMitVorbereitung(probekontext(t), db, nurGrundlinie,
		PruefeBestand, nil, grundlinie); fehler != nil {
		t.Fatalf("der Stand von 0.8.0: %v", fehler)
	}

	fuelleBestand(t, db)

	vorher, _ := AngewendeteMigrationen(probekontext(t), db) //nolint:errcheck // unten verglichen
	if !slices.Equal(vorher, grundlinienumfangOder(t)) {
		t.Fatalf("der Stand von 0.8.0 führt %d Wanderungen", len(vorher))
	}

	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("der Aufstieg: %v", fehler)
	}

	nachher, _ := AngewendeteMigrationen(probekontext(t), db) //nolint:errcheck // unten verglichen
	// **Als erste nach der Grundlinie** — und nicht als letzte: 0099 folgt
	// ihr seit PERF02 (#2352).
	if len(nachher) <= len(vorher) || nachher[len(vorher)] != "0098_sicherungsvermerke.sql" {
		t.Fatalf("nach der Grundlinie verbucht: %v", nachher[len(vorher):])
	}

	// **Was vorher dastand, steht danach da** — die Zusage hinter „aufwärts".
	if zahl := treffer(t, db, `"Feuer"*`); zahl != 1 {
		t.Errorf("die Karte von vorher findet die Suche %d-mal", zahl)
	}

	vermerke := NeueSicherungsvermerke(db)

	// **Unbekannt heißt unbekannt**: ein Bestand aus 0.8.0 weiß nicht, wann er
	// gesichert wurde, und sagt das.
	if _, bekannt, fehler := vermerke.Letzter(probekontext(t)); fehler != nil || bekannt {
		t.Fatalf("ein Altbestand nennt einen letzten Vermerk (%v, %v)", bekannt, fehler)
	}

	gesichert := domain.Sicherungsvermerk{
		Gesichert: time.Date(2026, 10, 14, 17, 0, 0, 123_000_000, time.UTC),
		Anlass:    domain.SicherungVonHand, Flaeche: domain.FlaecheCLI,
		Datei: "bestand.db", Groesse: 8192, Pruefsumme: strings.Repeat("0f", 32),
	}
	if fehler := vermerke.Vermerken(probekontext(t), gesichert); fehler != nil {
		t.Fatalf("vermerken: %v", fehler)
	}

	gelesen, bekannt, fehler := vermerke.Letzter(probekontext(t))
	if fehler != nil || !bekannt || gelesen != gesichert {
		t.Errorf("gelesen %+v (%v, %v), erwartet %+v", gelesen, bekannt, fehler, gesichert)
	}
}

// **Ein Zeitpunkt, den das Schema durchlässt, der aber keiner ist**: der
// `GLOB` prüft Ziffern und nicht den Kalender. Gelesen wird er als Fehler,
// nicht als Nullzeit — sonst hieße ein kaputter Vermerk „nie gesichert".
func TestEinUnlesbarerVermerkIstEinFehler(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sicherungsvermerke (gesichert, anlass, flaeche, datei, groesse, pruefsumme)
		 VALUES ('2026-13-45T10:00:00.000Z', 'hand', 'cli', 'b.db', 1, ?)`,
		strings.Repeat("0f", 32)); fehler != nil {
		t.Fatalf("eintragen: %v", fehler)
	}

	if _, bekannt, fehler := NeueSicherungsvermerke(db).Letzter(probekontext(t)); fehler == nil || bekannt {
		t.Errorf("ein Monat 13 wurde gelesen (%v, %v)", bekannt, fehler)
	}
}

// **Ohne die Tafel ist es ein Fehler und kein „unbekannt"** — ein Bestand,
// der nicht gewandert ist, darf nicht wie einer aussehen, der nie gesichert
// wurde.
func TestOhneDieTafelIstEsEinFehler(t *testing.T) {
	db := leererBestand(t)

	if _, bekannt, fehler := NeueSicherungsvermerke(db).Letzter(probekontext(t)); fehler == nil || bekannt {
		t.Errorf("ohne Tafel: (%v, %v), erwartet ein Fehler", bekannt, fehler)
	}
}

// **Eine künftige Fassung ohne `migrationen/`** ist ein Fehler im Lauf, der
// sie baut — und nicht eine Fassung ohne Wanderungen danach.
func TestEineKuenftigeFassungBrauchtIhreWanderungen(t *testing.T) {
	db := leererBestand(t)

	if fehler := MigriereMit(probekontext(t), db, nurGrundlinie, nil); fehler != nil {
		t.Fatalf("mit Ordner: %v", fehler)
	}

	if fehler := MigriereMit(probekontext(t), db, fstestOhneOrdner(), nil); fehler == nil {
		t.Error("eine künftige Fassung ohne migrationen/ wurde angenommen")
	}
}

// **Die zweite Reihe hält, was die Domäne abweist** — eine Zeile, die an
// `Pruefe` vorbei geschrieben wird, scheitert an der Tafel (docs/INVARIANTEN.md).
func TestDieTafelWeistUngueltigeVermerkeAb(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	gut := map[string]any{
		"gesichert": "2026-10-14T17:00:00.000Z", "anlass": "hand", "flaeche": "cli",
		"datei": "b.db", "groesse": 1, "pruefsumme": strings.Repeat("0f", 32), "verschluesselt": 0,
	}

	for name, anders := range map[string]map[string]any{
		"eine Zeit ohne Millisekunden":        {"gesichert": "2026-10-14T17:00:00Z"},
		"ein unbekannter Anlass":              {"anlass": "nebenbei"},
		"eine unbekannte Fläche":              {"flaeche": "fenster2"},
		"ein Pfad":                            {"datei": "/Users/jemand/b.db"},
		"ein Pfad mit Gegenstrich":            {"datei": `C:\b.db`},
		"leer":                                {"groesse": 0},
		"eine Prüfsumme in Großbuchstaben":    {"pruefsumme": strings.Repeat("0F", 32)},
		"verschlüsselt ist kein Ja oder Nein": {"verschluesselt": 2},
	} {
		t.Run(name, func(t *testing.T) {
			zeile := map[string]any{}
			for k, v := range gut {
				zeile[k] = v
			}

			for k, v := range anders {
				zeile[k] = v
			}

			_, fehler := db.ExecContext(probekontext(t),
				`INSERT INTO sicherungsvermerke (gesichert, anlass, flaeche, datei, groesse, pruefsumme, verschluesselt)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				zeile["gesichert"], zeile["anlass"], zeile["flaeche"], zeile["datei"],
				zeile["groesse"], zeile["pruefsumme"], zeile["verschluesselt"])
			if fehler == nil {
				t.Error("die Tafel nahm die Zeile an")
			}
		})
	}
}
