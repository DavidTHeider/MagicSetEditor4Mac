package service_test

// **Fremde Profile beim Neuanfang: mitgezählt, nicht benannt** — Befund 1 des
// Zweitblicks über Phase G
// (#2282).
//
// Der Weg durch den echten Speicher steht in
// `internal/app/neuanfangfremde_test.go`; hier steht die Regel für die
// Hindernisse, die ein Doppel bequemer zeigt: das eigene Profil wird genannt,
// ein fremdes nicht, und zwei fremde sperren **einmal**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerNeuanfangNenntNurDasEigeneProfil(t *testing.T) {
	t.Parallel()

	// **Der Akteur ist Nutzer 7** (`wache()`); die Rückkehr ist überall offen.
	// **Jedes Profil hat ein Deck namens „Ein Deck"** — das Doppel gibt allen
	// denselben Umfang. Gezählt werden drei, genannt wird nur das eigene
	// (neuer Befund des Nachlaufs über Phase G).
	doppel := &ruecksetzungsdoppel{
		offen: 1,
		umfang: []domain.Betroffenes{
			{Was: domain.BetroffenDecks, Anzahl: 1, Namen: []string{"Ein Deck"}},
		},
		profile: []domain.Profil{{ID: 3, Name: "Meins", NutzerID: 7}},
		fremde: []domain.Profil{
			{ID: 4, Name: "Ihres", NutzerID: 8},
			{ID: 5, Name: "Seines", NutzerID: 9},
		},
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(t.Context(),
		domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	erwartet := []domain.Hindernis{
		{Art: domain.RuecksetzungOffeneRueckkehr, Stelle: "Meins"},
		{Art: domain.RuecksetzungOffeneRueckkehr},
	}

	if len(vorschau.Hindernisse) != len(erwartet) {
		t.Fatalf("die Hindernisse sind %+v, erwartet %+v", vorschau.Hindernisse, erwartet)
	}

	for stelle, hindernis := range erwartet {
		if vorschau.Hindernisse[stelle] != hindernis {
			t.Errorf("Hindernis %d ist %+v, erwartet %+v", stelle,
				vorschau.Hindernisse[stelle], hindernis)
		}
	}

	if len(vorschau.Betroffen) != 1 || vorschau.Betroffen[0].Anzahl != 3 ||
		len(vorschau.Betroffen[0].Namen) != 1 {
		t.Errorf("betroffen ist %+v — erwartet drei Decks, genannt nur das eigene",
			vorschau.Betroffen)
	}
}

// **Ein Profil sieht nur, was dem Menschen gehört** — die fremden Profile
// kommen beim Löschen eines Profils gar nicht erst in den Blick.
func TestDieProfilvorschauSiehtKeineFremdenProfile(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		offen:   1,
		nutzer:  2,
		profile: []domain.Profil{{ID: 3, Name: "Meins", NutzerID: 7}},
		fremde:  []domain.Profil{{ID: 4, Name: "Ihres", NutzerID: 8}},
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(t.Context(),
		domain.RuecksetzungNutzer, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if len(vorschau.Hindernisse) != 1 || vorschau.Hindernisse[0].Stelle != "Meins" {
		t.Errorf("die Hindernisse sind %+v", vorschau.Hindernisse)
	}
}

// **Die Löschung eines Menschen zählt seine Zugänge an fremden Profilen** —
// #2287: sie gehen per Kaskade mit, und die Vorschau nannte sie nicht.
func TestDieNutzerloeschungZaehltFremdeZugaenge(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		nutzer:         2,
		profile:        []domain.Profil{{ID: 3, Name: "Meins", NutzerID: 7}},
		fremdeZugaenge: 2,
	}

	vorschau, fehler := ruecksetzung(doppel).Vorschau(t.Context(),
		domain.RuecksetzungNutzer, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	for _, gruppe := range vorschau.Betroffen {
		if gruppe.Was == domain.BetroffenZugaenge {
			if gruppe.Anzahl != 2 || len(gruppe.Namen) != 0 {
				t.Errorf("die Zugänge sind %+v — erwartet zwei, ohne Namen", gruppe)
			}

			return
		}
	}

	t.Errorf("die Vorschau nennt keine Zugänge: %+v", vorschau.Betroffen)
}
