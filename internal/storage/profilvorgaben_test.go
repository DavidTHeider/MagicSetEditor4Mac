package storage_test

// **Vorgaben, die auf Zeilen zeigen — im Bestand** — Werkstück 15 von
// Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: sie teilen eine
// Datenquelle im Arbeitsspeicher, und nebenläufig meldet SQLite „database
// table is locked".

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was gesetzt wird, kommt zurück — und in seiner Reihenfolge.**
func TestEineVorgabeGehtInDenBestandUndZurueck(t *testing.T) {
	ablage := storage.NeueVorgabenablage(frischeDatenbank(t))
	ctx := t.Context()

	if fehler := ablage.VorgabeSetzen(ctx, 1, domain.Profilvorgabe{
		Art: domain.VorgabeDeck, Werte: []string{"7", "3", "9"},
	}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	gelesen, fehler := ablage.Vorgaben(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	werte := gelesen[domain.VorgabeDeck]

	// **Die Reihenfolge ist die Aussage** — bei Präferenzdecks heißt sie
	// genau das. Eine Menge zurückzugeben wäre eine andere Antwort.
	if !slices.Equal(werte, []string{"7", "3", "9"}) {
		t.Errorf("gelesen = %v, erwartet [7 3 9]", werte)
	}
}

// **Setzen ersetzt und verschmilzt nicht.**
//
// **Dieselbe Regel wie bei den Angaben eines Profils**: was nicht mitkommt,
// steht danach nicht mehr da. Eine Vorgabe, die sich nur ergänzen ließe, wäre
// eine, die man nicht wegnehmen kann.
func TestVorgabenSetzenErsetzt(t *testing.T) {
	ablage := storage.NeueVorgabenablage(frischeDatenbank(t))
	ctx := t.Context()

	setze := func(werte ...string) {
		t.Helper()

		if fehler := ablage.VorgabeSetzen(ctx, 1, domain.Profilvorgabe{
			Art: domain.VorgabeDeck, Werte: werte,
		}); fehler != nil {
			t.Fatalf("setzen: %v", fehler)
		}
	}

	setze("7", "3")
	setze("9")

	gelesen, fehler := ablage.Vorgaben(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if !slices.Equal(gelesen[domain.VorgabeDeck], []string{"9"}) {
		t.Errorf("gelesen = %v, erwartet [9]", gelesen[domain.VorgabeDeck])
	}

	t.Run("und eine leere Liste räumt ab", func(t *testing.T) {
		setze()

		nachher, fehler := ablage.Vorgaben(ctx, 1)
		if fehler != nil {
			t.Fatalf("lesen: %v", fehler)
		}

		if len(nachher[domain.VorgabeDeck]) != 0 {
			t.Errorf("nachher = %v", nachher[domain.VorgabeDeck])
		}
	})
}

// **Die Arten stören einander nicht.**
func TestJedeArtStehtFuerSich(t *testing.T) {
	ablage := storage.NeueVorgabenablage(frischeDatenbank(t))
	ctx := t.Context()

	for art, werte := range map[domain.Vorgabeart][]string{
		domain.VorgabeFormat:   {"5"},
		domain.VorgabeSpiel:    {"magic"},
		domain.VorgabeVorlage:  {"magic-m15"},
		domain.VorgabeDeck:     {"1", "2"},
		domain.VorgabeLagerort: {"Schrank"},
	} {
		if fehler := ablage.VorgabeSetzen(ctx, 1, domain.Profilvorgabe{
			Art: art, Werte: werte,
		}); fehler != nil {
			t.Fatalf("%s setzen: %v", art, fehler)
		}
	}

	gelesen, fehler := ablage.Vorgaben(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 5 {
		t.Fatalf("gelesen sind %d Arten: %v", len(gelesen), gelesen)
	}

	// **Das Abräumen einer Art lässt die anderen stehen** — sonst wäre jedes
	// Setzen ein Zurücksetzen aller.
	if raeumfehler := ablage.VorgabeSetzen(ctx, 1, domain.Profilvorgabe{
		Art: domain.VorgabeDeck,
	}); raeumfehler != nil {
		t.Fatalf("abräumen: %v", raeumfehler)
	}

	nachher, fehler := ablage.Vorgaben(ctx, 1)
	if fehler != nil {
		t.Fatalf("neu lesen: %v", fehler)
	}

	if len(nachher) != 4 || len(nachher[domain.VorgabeFormat]) != 1 {
		t.Errorf("nachher = %v", nachher)
	}
}

// **Und die Profile stören einander nicht.**
func TestJedesProfilHatSeineEigenen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueVorgabenablage(db)
	ctx := t.Context()

	zweites, fehler := storage.NeueNutzerablage(db).ProfilAnlegen(ctx,
		domain.Profil{NutzerID: 1, Name: "Turnier"})
	if fehler != nil {
		t.Fatalf("Profil anlegen: %v", fehler)
	}

	if setzfehler := ablage.VorgabeSetzen(ctx, 1, domain.Profilvorgabe{
		Art: domain.VorgabeFormat, Werte: []string{"5"},
	}); setzfehler != nil {
		t.Fatalf("setzen: %v", setzfehler)
	}

	fremde, fehler := ablage.Vorgaben(ctx, zweites.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(fremde) != 0 {
		t.Errorf("das zweite Profil sieht %v", fremde)
	}
}

// **Eine Vorgabe, die die Domäne abweist, kommt gar nicht erst hinein.**
func TestEineKrummeVorgabeWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueVorgabenablage(frischeDatenbank(t))

	faelle := map[string]domain.Profilvorgabe{
		"unbekannte Art": {Art: "kartenrücken", Werte: []string{"1"}},
		"zwei bei einer": {Art: domain.VorgabeFormat, Werte: []string{"1", "2"}},
		"leerer Wert":    {Art: domain.VorgabeDeck, Werte: []string{""}},
	}

	for name, vorgabe := range faelle {
		t.Run(name, func(t *testing.T) {
			if fehler := ablage.VorgabeSetzen(
				t.Context(), 1, vorgabe); fehler == nil {
				t.Error("sie kam durch")
			}
		})
	}
}

// **Ein Profil, das es nicht gibt, ist kein leeres Profil** (§8).
//
// **Ohne diese Unterscheidung bekäme ein Tippfehler in der Kennung fünf leere
// Arten zurück** — und die Seite sähe aus, als wäre nichts gesetzt.
func TestEinFehlendesProfilIstKeinLeeres(t *testing.T) {
	ablage := storage.NeueVorgabenablage(frischeDatenbank(t))

	if _, fehler := ablage.Vorgaben(t.Context(), 4711); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
	}

	// **Der Beleg, dass das vorhandene Profil durchkommt** (§10).
	if _, fehler := ablage.Vorgaben(t.Context(), 1); fehler != nil {
		t.Errorf("das vorhandene Profil wurde abgewiesen: %v", fehler)
	}
}
