package storage_test

// **Freigaben im Bestand** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: sie teilen eine
// Datenquelle im Arbeitsspeicher.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zweiterNutzer legt einen zweiten Nutzer an und gibt seine Kennung.
func zweiterNutzer(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	angelegt, fehler := storage.NeueNutzerablage(db).NutzerAnlegen(
		t.Context(), domain.Nutzer{Name: "Zweiter"})
	if fehler != nil {
		t.Fatalf("den zweiten Nutzer anlegen: %v", fehler)
	}

	return angelegt.ID
}

// **Eine Freigabe geht hinein und kommt zurück — in beiden Richtungen.**
func TestEineFreigabeGehtInDenBestandUndZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueFreigabenablage(db)
	ctx := t.Context()

	nutzerID := zweiterNutzer(t, db)

	if fehler := ablage.FreigabeSetzen(ctx, domain.Profilfreigabe{
		ProfilID: 1, NutzerID: nutzerID, Rolle: domain.RolleMitarbeiter,
	}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	amProfil, fehler := ablage.Freigaben(ctx, 1)
	if fehler != nil {
		t.Fatalf("am Profil lesen: %v", fehler)
	}

	if len(amProfil) != 1 || amProfil[0].Rolle != domain.RolleMitarbeiter {
		t.Fatalf("am Profil = %+v", amProfil)
	}

	if amProfil[0].Angelegt == "" {
		t.Error("die Freigabe trägt keinen Zeitpunkt")
	}

	// **Dieselbe Tabelle, die andere Frage**: „welche fremden Profile darf
	// ich?"
	beimNutzer, fehler := ablage.FreigabenFuer(ctx, nutzerID)
	if fehler != nil {
		t.Fatalf("beim Nutzer lesen: %v", fehler)
	}

	if len(beimNutzer) != 1 || beimNutzer[0].ProfilID != 1 {
		t.Errorf("beim Nutzer = %+v", beimNutzer)
	}
}

// **Setzen legt an oder schreibt fort — und der Zeitpunkt bleibt stehen.**
func TestEineFreigabeWirdFortgeschrieben(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueFreigabenablage(db)
	ctx := t.Context()

	nutzerID := zweiterNutzer(t, db)

	setze := func(rolle domain.Rolle) {
		t.Helper()

		if fehler := ablage.FreigabeSetzen(ctx, domain.Profilfreigabe{
			ProfilID: 1, NutzerID: nutzerID, Rolle: rolle,
		}); fehler != nil {
			t.Fatalf("setzen: %v", fehler)
		}
	}

	setze(domain.RolleLeser)

	vorher, fehler := ablage.Freigaben(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	setze(domain.RolleVerwalter)

	nachher, lesefehler := ablage.Freigaben(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("neu lesen: %v", lesefehler)
	}

	if len(nachher) != 1 {
		t.Fatalf("es sind %d Freigaben", len(nachher))
	}

	if nachher[0].Rolle != domain.RolleVerwalter {
		t.Errorf("Rolle = %q", nachher[0].Rolle)
	}

	// **Der Zeitpunkt sagt, seit wann dieser Mensch Zugang hat** — eine
	// Rollenänderung fängt das nicht neu an.
	if nachher[0].Angelegt != vorher[0].Angelegt {
		t.Errorf("Angelegt = %q, vorher %q",
			nachher[0].Angelegt, vorher[0].Angelegt)
	}
}

// **Eine Freigabe wird zurückgezogen — und ein zweites Mal ist ein Nein.**
func TestEineFreigabeWirdZurueckgezogen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueFreigabenablage(db)
	ctx := t.Context()

	nutzerID := zweiterNutzer(t, db)

	if fehler := ablage.FreigabeSetzen(ctx, domain.Profilfreigabe{
		ProfilID: 1, NutzerID: nutzerID, Rolle: domain.RolleLeser,
	}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if fehler := ablage.FreigabeLoeschen(ctx, 1, nutzerID); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	// **Null Zeilen sind ein Nein und kein stilles Ja** (§8): wer eine
	// Freigabe zurückzieht, die es nicht gibt, hat sich vertan — und ein
	// „gut" ließe ihn glauben, er habe jemandem den Zugang genommen.
	if fehler := ablage.FreigabeLoeschen(ctx, 1, nutzerID); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Fehler = %v, erwartet ErrNotFound", fehler)
	}
}

// **Eine krumme Freigabe kommt gar nicht erst hinein.**
func TestEineKrummeFreigabeWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueFreigabenablage(frischeDatenbank(t))

	faelle := map[string]domain.Profilfreigabe{
		"ohne Profil":     {NutzerID: 1, Rolle: domain.RolleLeser},
		"ohne Nutzer":     {ProfilID: 1, Rolle: domain.RolleLeser},
		"erfundene Rolle": {ProfilID: 1, NutzerID: 1, Rolle: "gott"},
	}

	for name, freigabe := range faelle {
		t.Run(name, func(t *testing.T) {
			if fehler := ablage.FreigabeSetzen(
				t.Context(), freigabe); fehler == nil {
				t.Error("sie kam durch")
			}
		})
	}
}

// **Und die Sichtbarkeit geht in den Bestand und zurück.**
func TestDieSichtbarkeitGehtInDenBestandUndZurueck(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := t.Context()

	profile, fehler := ablage.Profile(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Ein Bestand aus der Zeit davor steht auf `privat`** — die vorsichtige
	// Antwort, und die Wanderung ändert nichts am Verhalten.
	if profile[0].Sichtbarkeit != string(domain.SichtbarPrivat) {
		t.Errorf("Sichtbarkeit = %q", profile[0].Sichtbarkeit)
	}

	geaendert := profile[0]
	geaendert.Sichtbarkeit = string(domain.SichtbarImBestand)

	if _, err := ablage.ProfilAendern(ctx, geaendert); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	nachher, lesefehler := ablage.Profile(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("neu lesen: %v", lesefehler)
	}

	if nachher[0].Sichtbarkeit != string(domain.SichtbarImBestand) {
		t.Errorf("Sichtbarkeit = %q", nachher[0].Sichtbarkeit)
	}

	t.Run("und eine erfundene wird abgewiesen", func(t *testing.T) {
		krumm := nachher[0]
		krumm.Sichtbarkeit = "halb"

		if _, err := ablage.ProfilAendern(ctx, krumm); err == nil {
			t.Error("die erfundene Sichtbarkeit kam durch")
		}
	})
}
