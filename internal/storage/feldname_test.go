// Ein Feld ohne Namen kommt nicht in den Bestand.
//
// # Warum das hier steht und nicht in `app`
//
// #89 fragt
// nicht „mehr prüfen", sondern **wo**: eine Regel, die nur an einer Fassade
// steht, kennt die Domäne nicht — und der nächste Schreiber geht daran vorbei.
//
// **Genau das war eingetreten.** `app.FeldSetzen` wies einen leeren Feldnamen
// ab, seit Phase 5. Der MCP-Server nimmt einen anderen Weg
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)) und wies
// ihn nicht ab — ein Sprachmodell konnte eine Zeile anlegen, die niemand
// wiederfindet, und beim Zurückschreiben stünde `: wert` in der Paketdatei.
//
// Die Regel steht jetzt in `internal/domain` und wird auf dem **Schreibweg**
// angewandt. Damit gilt sie für jeden Schreiber, auch für den, den es noch
// nicht gibt.
package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinFeldOhneNamenWirdAbgewiesen(t *testing.T) {
	db, _, karteID := bestandMitKarte(t)

	schreiber := storage.NeuerKartenschreiber(db)

	faelle := map[string]func(feld string) error{
		"setzen": func(feld string) error {
			_, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, feld, "x")

			return err
		},
		"entfernen": func(feld string) error {
			_, _, _, err := schreiber.FeldEntfernen(probekontext(t), karteID, feld)

			return err
		},
		"Massenänderung": func(feld string) error {
			_, fehler := schreiber.FelderSetzen(probekontext(t), []int64{karteID}, feld, "x")

			return fehler
		},
	}

	// **Auch Leerraum ist kein Name.** Ein Feld `" "` fände niemand wieder, und
	// in der Paketdatei stünde eine Zeile, die aussieht wie ein Einrückungsfehler.
	for _, feld := range []string{"", " ", "\t"} {
		for name, schreiben := range faelle {
			t.Run(name+" "+visuell(feld), func(t *testing.T) {
				err := schreiben(feld)
				if err == nil {
					t.Fatal("kein Fehler — ein Feld ohne Namen kam in den Bestand")
				}

				if !errors.Is(err, domain.ErrValidation) {
					t.Errorf("%v ist keine Prüfmeldung", err)
				}
			})
		}
	}
}

// visuell macht einen Namen für einen Testfall aus einer leeren Zeichenkette.
func visuell(feld string) string {
	switch feld {
	case "":
		return "leer"

	case " ":
		return "Leerzeichen"

	default:
		return "Tabulator"
	}
}

// **Ein Name mit Leerraum außen bleibt, wie er ist.**
//
// Beschnitten wird nur zum **Prüfen**, nicht zum Schreiben: ein Feld heißt in
// der Paketdatei, wie es dasteht, und wer es beim Schreiben stutzte, änderte
// eine Zeile, die niemand angefasst hat — gegen die Byte-Zusage aus #48.
func TestEinNameMitLeerraumBleibtWieErIst(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, _, err := storage.NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), karteID, " rand ", "x"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if _, gabEs := karten[0].Value(" rand "); !gabEs {
		t.Errorf("das Feld heißt nicht mehr wie geschrieben: %+v", karten[0].Values)
	}
}
