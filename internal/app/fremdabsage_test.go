package app_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede Bindung über einen Menschen sagt ab, bevor sie wirkt** — ID-03
// (#1702).
//
// An einer Nutzerverwaltung, die **immer** Nein sagt: jede Bindung muss mit
// der Absage zurückkommen und nicht mit einem Fehler dahinter — einem
// fehlenden Faktordienst, einer Datei, die es nicht gibt, einem Dialog. Die
// Wirkung selbst halten die Läufe gegen den echten Bestand
// (`fremdernutzer_test.go`); hier steht, dass **jede** Bindung vorher fragt.
func TestJedeBindungUeberEinenMenschenSagtVorDerWirkungAb(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitNutzern(&nutzerfuehrer{ueber: fmt.Errorf("%w: Probe", domain.ErrFremderNutzer)})

	ctx := t.Context()

	faelle := map[string]*app.Fehler{
		"Konto lösen":           anwendung.KontoLoesen(ctx, "2", "google").Fehler,
		"Profil anlegen":        anwendung.ProfilAnlegen(ctx, "2", "Neu").Fehler,
		"Nutzerbild setzen":     anwendung.NutzerbildSetzen(ctx, "2", "1", "/gibt/es/nicht.png").Fehler,
		"Nutzerbild wählen":     anwendung.NutzerbildWaehlen(ctx, "2", "1").Fehler,
		"Nutzerbild entfernen":  anwendung.NutzerbildEntfernen(ctx, "2", "1").Fehler,
		"Einstellung setzen":    anwendung.NutzereinstellungSetzen(ctx, "2", "sprache", "de").Fehler,
		"Personendaten":         anwendung.Personendaten(ctx, "2").Fehler,
		"Personendaten sichern": anwendung.PersonendatenSichern(ctx, "2").Fehler,
		"Tischbild wählen":      anwendung.TischbildWaehlen(ctx, "2").Fehler,
		"Tischbild entfernen":   anwendung.TischbildEntfernen(ctx, "2").Fehler,
		"Fensterbild wählen":    anwendung.FensterbildWaehlen(ctx, "2").Fehler,
		"Fensterbild entfernen": anwendung.FensterbildEntfernen(ctx, "2").Fehler,
		"Faktor bestätigen":     anwendung.FaktorBestaetigen(ctx, "2", "123456").Fehler,
		"Biometrie einrichten":  anwendung.BiometrieEinrichten(ctx, "2").Fehler,
		"Abzeichen wählen":      anwendung.ProfilbadgeWaehlen(ctx, "2", "1").Fehler,
		"Abzeichen setzen":      anwendung.ProfilbadgeSetzen(ctx, "2", "1", "/gibt/es/nicht.png").Fehler,
		"Abzeichen entfernen":   anwendung.ProfilbadgeEntfernen(ctx, "2", "1").Fehler,
	}

	for name, fehler := range faelle {
		if fehler == nil || fehler.Schluessel != domain.SchluesselFremderNutzer {
			t.Errorf("%s: %+v, erwartet die Absage über einen fremden Menschen", name, fehler)
		}
	}
}

// **Eine Kennung, die keine ist, wird vor der Frage abgewiesen** — auch an
// den Bindungen, die vor einem Dialog fragen.
func TestVorDemDialogWirdDieKennungGelesen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitNutzern(&nutzerfuehrer{})

	ctx := t.Context()

	for name, fehler := range map[string]*app.Fehler{
		"Nutzerbild": anwendung.NutzerbildWaehlen(ctx, "keine", "1").Fehler,
		"Abzeichen":  anwendung.ProfilbadgeWaehlen(ctx, "keine", "1").Fehler,
	} {
		if fehler == nil || fehler.Code != app.CodeValidierung {
			t.Errorf("%s: %+v, erwartet eine Validierung", name, fehler)
		}
	}
}

// **Fehlt das Profil zwischen Frage und Lesen, heißt es „nicht gefunden"** —
// die Frage sagt Ja, und das Lesen danach findet es nicht mehr.
func TestEinProfilDasNachDerFrageFehltIstNichtGefunden(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitNutzern(&nutzerfuehrer{})

	ctx := t.Context()

	for name, fehler := range map[string]*app.Fehler{
		"Profil ändern": anwendung.ProfilAendern(ctx, app.ProfilangabenDTO{
			ID: "9", Name: "Weg", Fassung: "1",
		}).Fehler,
		"Abzeichen entfernen": anwendung.ProfilbadgeEntfernen(ctx, "9", "1").Fehler,
	} {
		if fehler == nil || fehler.Code != app.CodeNichtGefunden {
			t.Errorf("%s: %+v, erwartet nicht gefunden", name, fehler)
		}
	}
}
