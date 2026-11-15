package app

// **Eine Fläche ohne Menschen davor fragt den Sensor nicht** — SEC-07
// (#2148).
//
// Der Sensor dieses Rechners ist hier ein zählendes Doppel an der Stelle, an
// der die Wurzel ihn sonst holt. Die Wurzel mit Sensor fragt ihn — das ist
// der Vorher-Stand, an dem REST und MCP bis SEC-07 hingen; die kopflose fragt
// ihn nie, auch nicht nach `Vorhanden`.

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

const kopflosKennwort = "ein-langes-kennwort"

type zaehlsensor struct{ gefragt, bestaetigt int }

func (s *zaehlsensor) Vorhanden(context.Context) bool {
	s.gefragt++

	return true
}

func (s *zaehlsensor) Bestaetige(context.Context, string) error {
	s.bestaetigt++

	return nil
}

// biometrischerNutzer legt ein Konto an und gibt ihm Touch ID als zweiten
// Faktor — direkt in der Ablage, denn das Einrichten selbst fragt den Sensor.
func biometrischerNutzer(t *testing.T) (db *sql.DB, nutzerID, profilID string) {
	t.Helper()

	ctx := storage.MitProfil(t.Context(), 1)

	offen, err := storage.Oeffnen(ctx, filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = offen.Close() })

	if migrierfehler := storage.Migriere(ctx, offen, nil); migrierfehler != nil {
		t.Fatalf("migrieren: %v", migrierfehler)
	}

	angelegt := NutzerflaecheUeber(offen).KontoEintragen(ctx, "kopflos@example.org", kopflosKennwort, "Kopflos")
	if angelegt.Fehler != nil || angelegt.Nutzer == nil || len(angelegt.Nutzer.Profile) == 0 {
		t.Fatalf("anlegen: %+v", angelegt)
	}

	kennung, err := strconv.ParseInt(angelegt.Nutzer.ID, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := storage.NeueNutzerablage(offen).FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: kennung, Art: domain.FaktorBiometrie,
	}); err != nil {
		t.Fatalf("Faktor: %v", err)
	}

	return offen, angelegt.Nutzer.ID, angelegt.Nutzer.Profile[0].ID
}

func TestEineKopfloseFlaecheFragtDenSensorNie(t *testing.T) {
	zaehler := &zaehlsensor{}
	vorher := rechnersensor
	rechnersensor = zaehler

	t.Cleanup(func() { rechnersensor = vorher })

	db, nutzerID, profilID := biometrischerNutzer(t)
	ctx := storage.MitProfil(t.Context(), 1)

	// **Vorher: die Wurzel mit Sensor fragt ihn** — so hingen REST und MCP
	// bis SEC-07 daran, und so bleibt es für Fenster und Kommandozeile.
	if antwort := NutzerflaecheUeber(db).KontoAnmelden(ctx, "kopflos@example.org", kopflosKennwort, ""); antwort.Fehler != nil {
		t.Fatalf("mit Sensor: %+v", antwort.Fehler)
	}

	if zaehler.bestaetigt != 1 {
		t.Fatalf("mit Sensor wurde %d-mal bestätigt", zaehler.bestaetigt)
	}

	zaehler.gefragt, zaehler.bestaetigt = 0, 0
	kopflos := NutzerflaecheKopflos(db)

	wege := map[string]func() *Fehler{
		"Anmeldung": func() *Fehler {
			return kopflos.KontoAnmelden(ctx, "kopflos@example.org", kopflosKennwort, "").Fehler
		},
		"Profilwechsel": func() *Fehler {
			return kopflos.ProfilWechseln(ctx, profilID, "").Fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			fehler := weg()
			if fehler == nil || fehler.Feld != "faktor" || fehler.Schluessel != domain.SchluesselZweiterFaktor {
				t.Errorf("gekommen ist %+v statt des fehlenden Faktors", fehler)
			}
		})
	}

	t.Run("Einrichtung", func(t *testing.T) {
		// **Am ersten Nutzer, in dessen Profil gearbeitet wird** — ein fremder
		// scheiterte schon an der Erlaubnis, und der Sensor bliebe aus einem
		// anderen Grund still.
		if fehler := kopflos.BiometrieEinrichten(ctx, "1").Fehler; fehler == nil || fehler.Feld != "art" {
			t.Errorf("gekommen ist %+v statt „kein Sensor“", fehler)
		}
	})

	t.Run("Auskunft", func(t *testing.T) {
		if antwort := kopflos.Faktoren(ctx, nutzerID); antwort.Fehler != nil || antwort.Biometrie || antwort.Biometriegrund == "" {
			t.Errorf("gekommen ist %+v", antwort)
		}
	})

	if zaehler.gefragt != 0 || zaehler.bestaetigt != 0 {
		t.Errorf("kopflos gefragt %d, bestätigt %d — erwartet null", zaehler.gefragt, zaehler.bestaetigt)
	}
}

// **Wer den kopflosen Sensor doch fragt, bekommt den fehlenden Faktor** —
// über `Vorhanden` wird er nie erreicht, aber ein Weg, der ihn direkt riefe,
// soll dieselbe Antwort sehen.
func TestDerKopfloseSensorSagtNieJa(t *testing.T) {
	var befund *domain.FieldError
	if fehler := (kopflos{}).Bestaetige(t.Context(), "grund"); !errors.As(fehler, &befund) ||
		befund.Key != domain.SchluesselZweiterFaktor {
		t.Errorf("gekommen ist %v", fehler)
	}
}
