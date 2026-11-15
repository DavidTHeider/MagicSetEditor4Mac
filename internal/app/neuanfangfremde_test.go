package app_test

// **Der Neuanfang sieht den ganzen Bestand, den er beiseitelegt** — Befund 1
// des Zweitblicks über Phase G
// (#2282).
//
// # Der Befund
//
// **Vorschau und Sperre zählten nur die Profile dessen, der ruft**, die
// Ausführung legt aber die ganze Datenbank beiseite. Was ein zweiter Mensch
// im Bestand hatte, stand in keiner Vorschau; seine offene Rückkehr hielt den
// Neuanfang nicht auf, und eine Änderung in seinem Profil änderte die
// Revision nicht. G12 hat diese Lücke ausführbar gemacht.
//
// # Die Regel für fremde Daten
//
// **Mitgezählt, nicht benannt.** Die Mengen eines fremden Profils gehen in die
// Summe ein — beim Neuanfang nennt die Vorschau ohnehin kein Profil —, und
// eine offene Rückkehr darin sperrt, **ohne** den Namen des fremden Profils
// zu nennen.

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// fremdesProfil legt einen zweiten Menschen an und gibt sein Profil.
func fremdesProfil(t *testing.T, ablage *storage.Nutzerablage) domain.Profil {
	t.Helper()

	fremder, fehler := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Fremde"})
	if fehler != nil {
		t.Fatalf("den zweiten Menschen anlegen: %v", fehler)
	}

	profile, profilfehler := ablage.Profile(probekontext(t), fremder.ID)
	if profilfehler != nil || len(profile) != 1 {
		t.Fatalf("sein Profil: %v, %v", profile, profilfehler)
	}

	return profile[0]
}

// einstellungenIn liest die Zahl der Einstellungen aus einer Vorschau.
func einstellungenIn(vorschau app.RuecksetzungsvorschauAntwort) int {
	for _, gruppe := range vorschau.Betroffen {
		if gruppe.Was == domain.BetroffenEinstellungen {
			return gruppe.Anzahl
		}
	}

	return 0
}

func TestDerNeuanfangZaehltFremdeProfileMit(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	db := bestandBei(t, ort)
	ctx := probekontext(t)
	fremd := fremdesProfil(t, storage.NeueNutzerablage(db))
	anwendung := app.RuecksetzungsAnwendungUeber(db, ort)

	vorher := anwendung.Ruecksetzungsvorschau(ctx, "anwendung", "", "")
	if vorher.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorher.Fehler)
	}

	// **Eine Wahl im fremden Profil** — sie ginge mit dem Bestand beiseite.
	if err := storage.NeueEinstellungen(db).Setzen(storage.MitProfil(ctx, fremd.ID),
		domain.NetzProxyHTTP, "http://127.0.0.1:3128"); err != nil {
		t.Fatalf("die fremde Wahl setzen: %v", err)
	}

	nachher := anwendung.Ruecksetzungsvorschau(ctx, "anwendung", "", "")
	if nachher.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", nachher.Fehler)
	}

	if einstellungenIn(nachher) != einstellungenIn(vorher)+1 {
		t.Errorf("die Vorschau zählt %d Einstellungen, vorher %d — die fremde fehlt",
			einstellungenIn(nachher), einstellungenIn(vorher))
	}

	// **Und die Revision von vorher gilt nicht mehr**: was zwischen Hinsehen
	// und Klicken im fremden Profil dazukam, ginge sonst ungesehen mit.
	if nachher.Revision == vorher.Revision {
		t.Error("eine Änderung im fremden Profil ließ die Revision stehen")
	}
}

func TestEineFremdeOffeneRueckkehrSperrtDenNeuanfang(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	db := bestandBei(t, ort)
	ctx := probekontext(t)
	fremd := fremdesProfil(t, storage.NeueNutzerablage(db))

	if _, err := storage.NeuesWiederherstellungsbuch(db).Vermerken(ctx,
		domain.Wiederherstellungsepoche{
			Zurueckgespielt: time.Now(), Datenstand: time.Now(),
			Quelle: "probe.msesicherung", Klaerung: domain.KlaerungOffen,
			Mengen: []domain.Sammlungsmenge{
				{Profil: fremd.ID, Schluessel: "scryfall:abc", Menge: 3},
			},
		}); err != nil {
		t.Fatalf("die Epoche vermerken: %v", err)
	}

	vorschau := app.RuecksetzungsAnwendungUeber(db, ort).Ruecksetzungsvorschau(
		ctx, "anwendung", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	if vorschau.Bereit || len(vorschau.Hindernisse) != 1 ||
		vorschau.Hindernisse[0].Art != domain.RuecksetzungOffeneRueckkehr {
		t.Fatalf("die fremde offene Rückkehr sperrt nicht: %+v", vorschau.Hindernisse)
	}

	if vorschau.Hindernisse[0].Stelle != "" {
		t.Errorf("das Hindernis nennt das fremde Profil: %q", vorschau.Hindernisse[0].Stelle)
	}
}

// bestandBei öffnet und migriert einen Bestand an einem Ort, den die
// Rücksetzung kennt.
func bestandBei(t *testing.T, ort string) *sql.DB {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	return db
}

// **Und nennt nichts aus ihm** — neuer Befund des Nachlaufs über Phase G: die
// Vorschau zählte fremde Mengen mit, gab aber auch ihre Namen weiter. Hatte
// das eigene Profil kein Deck, stand der Name eines fremden da.
func TestDerNeuanfangNenntKeineFremdenNamen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	db := bestandBei(t, ort)
	ctx := probekontext(t)
	fremd := fremdesProfil(t, storage.NeueNutzerablage(db))
	imFremden := storage.MitProfil(ctx, fremd.ID)

	if _, err := storage.NeueDeckablage(db).Anlegen(imFremden,
		domain.Deck{Name: "Fremdes Deck", Spiel: "magic"}); err != nil {
		t.Fatalf("das fremde Deck anlegen: %v", err)
	}

	if err := storage.NeueEinstellungen(db).Setzen(imFremden,
		domain.NetzProxyHTTP, "http://127.0.0.1:3128"); err != nil {
		t.Fatalf("die fremde Wahl setzen: %v", err)
	}

	vorschau := app.RuecksetzungsAnwendungUeber(db, ort).Ruecksetzungsvorschau(
		ctx, "anwendung", "", "")
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", vorschau.Fehler)
	}

	gezaehlt := map[string]app.BetroffenesDTO{}
	for _, gruppe := range vorschau.Betroffen {
		gezaehlt[gruppe.Was] = gruppe
	}

	// **Gezählt ist es** — sonst prüfte das Fehlen der Namen nichts.
	if gezaehlt[domain.BetroffenDecks].Anzahl != 1 {
		t.Fatalf("das fremde Deck ist nicht gezählt: %+v", vorschau.Betroffen)
	}

	for _, gruppe := range vorschau.Betroffen {
		if len(gruppe.Namen) > 0 {
			t.Errorf("%s nennt fremde Namen: %v", gruppe.Was, gruppe.Namen)
		}
	}
}
