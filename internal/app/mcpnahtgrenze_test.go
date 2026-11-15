// **Die Naht des MCP-Servers führt über die Kartenaufsicht** (#801).
package app_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Karte aus einem fremden Set kommt über MCP nicht durch.**
//
// Der Test geht über `Schreibnaht` und damit über die **wirkliche**
// Verdrahtung, nicht über eine nachgebaute. Genau dort lag der Befund: die
// Aufsicht steht in `service`, und die Naht reichte den Lebenszyklus bis zum
// 2026-08-22 unmittelbar an `storage` durch. Ein Test gegen die Aufsicht
// allein wäre grün geblieben, während der Weg an ihr vorbeiführte
// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
func TestDieMCPNahtHaeltDieSetgrenze(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad, fehler := app.Datenbankpfad()
	if fehler != nil {
		t.Fatalf("Pfad: %v", fehler)
	}

	if ordnerfehler := os.MkdirAll(filepath.Dir(pfad), 0o750); ordnerfehler != nil {
		t.Fatalf("Verzeichnis: %v", ordnerfehler)
	}

	db, fehler := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if migrationsfehler := storage.Migriere(probekontext(t), db, nil); migrationsfehler != nil {
		t.Fatalf("migrieren: %v", migrationsfehler)
	}

	schreiber, fehler := app.Schreibnaht(db, domain.UrheberServer)
	if fehler != nil {
		t.Fatalf("Schreibnaht bauen: %v", fehler)
	}

	eines := setAnlegen(t, db, "Eines")
	anderes := setAnlegen(t, db, "Anderes")

	karte, fehler := schreiber.KarteAnlegen(probekontext(t), eines)
	if fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	zweite, fehler := schreiber.KarteAnlegen(probekontext(t), eines)
	if fehler != nil {
		t.Fatalf("zweite Karte anlegen: %v", fehler)
	}

	faelle := []struct {
		name string
		ruf  func() error
	}{
		{"löschen", func() error {
			return schreiber.KarteLoeschen(probekontext(t), anderes, karte, false)
		}},
		{"kopieren", func() error {
			_, err := schreiber.KarteKopieren(probekontext(t), anderes, karte)

			return err
		}},
		{"tauschen", func() error {
			return schreiber.KartenTauschen(probekontext(t), anderes, karte, zweite)
		}},
		{"Feld setzen", func() error {
			_, _, err := schreiber.FeldSetzenMitVorwert(
				probekontext(t), anderes, karte, "name", "x")

			return err
		}},
		// **Auch die übernommenen Wege** (#1163). Sie standen bis Phase 3.2
		// nicht in dieser Aufzählung — und ihre Abdeckung hing daran, dass
		// `bildAblegen` sie zufällig benutzte. Als es auf den Zug über
		// mehrere Felder umgestellt wurde, fiel sie auf null: die Grenze war
		// für diese beiden Wege nie geprüft worden.
		{"Feld von außen setzen", func() error {
			_, _, err := schreiber.FeldSetzenVonAussen(
				probekontext(t), anderes, karte, "name", "x", "https://beispiel.test/")

			return err
		}},
		{"Felder von außen setzen", func() error {
			_, _, err := schreiber.FelderVonAussen(probekontext(t), anderes, karte,
				[]domain.Value{{Field: "name", Text: "x"}}, "https://beispiel.test/")

			return err
		}},
		{"Feld entfernen", func() error {
			_, _, _, err := schreiber.FeldEntfernen(probekontext(t), anderes, karte, "name")

			return err
		}},
		{"Feld an seine Stelle zurück", func() error {
			_, _, err := schreiber.FeldSetzenAnStelle(
				probekontext(t), anderes, karte, "name", "x", 0)

			return err
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if err := fall.ruf(); !errors.Is(err, domain.ErrNotFound) {
				t.Errorf("Fehler %v, erwartet war ErrNotFound", err)
			}
		})
	}

	// **Und die Gegenprobe: im eigenen Set geht es.** Ohne sie sagte der Test
	// nur, dass etwas fehlschlägt — nicht, dass die Grenze der Grund ist (§10).
	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), eines, karte, "name", "Eisball"); fehler != nil {
		t.Errorf("Feld im eigenen Set: %v", fehler)
	}

	// **Auch die übernommenen Wege schreiben wirklich** (#1163). Die
	// Aufzählung oben sagt nur, dass sie an der Grenze abweisen; ohne diese
	// beiden bliebe offen, ob sie **innerhalb** etwas tun.
	if _, _, fehler := schreiber.FeldSetzenVonAussen(
		probekontext(t), eines, karte, "name", "Eisball", "https://beispiel.test/"); fehler != nil {
		t.Errorf("Feld von außen im eigenen Set: %v", fehler)
	}

	if _, _, fehler := schreiber.FelderVonAussen(probekontext(t), eines, karte,
		[]domain.Value{{Field: "name", Text: "Eisball"}},
		"https://beispiel.test/"); fehler != nil {
		t.Errorf("Felder von außen im eigenen Set: %v", fehler)
	}

	if _, _, _, fehler := schreiber.FeldEntfernen(
		probekontext(t), eines, karte, "name"); fehler != nil {
		t.Errorf("entfernen im eigenen Set: %v", fehler)
	}

	if _, _, fehler := schreiber.FeldSetzenAnStelle(
		probekontext(t), eines, karte, "name", "Eisball", 0); fehler != nil {
		t.Errorf("zurücksetzen im eigenen Set: %v", fehler)
	}

	if fehler := schreiber.KarteLoeschen(probekontext(t), eines, karte, false); fehler != nil {
		t.Errorf("im eigenen Set: %v", fehler)
	}
}

// setAnlegen legt ein Set an und gibt seine Kennung.
func setAnlegen(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	var id int64

	if fehler := db.QueryRowContext(probekontext(t),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES (?, 'probe', ?, 1)
		 RETURNING id`, name, storage.Zeitstempel(time.Now()),
	).Scan(&id); fehler != nil {
		t.Fatalf("Set %q anlegen: %v", name, fehler)
	}

	return id
}

// **Die Naht zum Bestandssicherer schreibt wirklich** (#799).
//
// Sie steht in `mcpnaht.go`, weil nur die Composition Roots `storage`
// importieren dürfen (§2) — und damit wäre sie ohne diesen Test die eine
// Zeile, die niemand betritt.
//
// **Über `Starten` und nicht über eine gesetzte Naht**: geprüft werden soll,
// dass der Composition Root sie verdrahtet, nicht dass ein Test sie setzen
// kann.
func TestDieBestandsnahtSchreibtEineDatei(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("Beenden: %v", fehler)
		}
	})

	ziel := filepath.Join(t.TempDir(), "sicherung.db")

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return ziel, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	// **Und die Datei steht da, mit Inhalt.** Ohne diese Zusicherung sagte der
	// Test nur, dass kein Fehler kam.
	info, statfehler := os.Stat(ziel)
	if statfehler != nil {
		t.Fatalf("die Sicherung liegt nicht da: %v", statfehler)
	}

	if info.Size() == 0 {
		t.Error("die Sicherung ist leer")
	}
}
