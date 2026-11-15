package domain_test

// **Die Schnittmenge, mit allen Kombinationen** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum hier jede Kombination steht und nicht ein Beispiel
//
// **Weil eine Faltung genau dort bricht, wo niemand hinsieht.** Zwei Ebenen
// mit je drei Zuständen — gesetzt streng, gesetzt locker, nichts gesagt —
// sind neun Fälle, und der interessante ist der, in dem eine Ebene schweigt:
// dort entscheidet sich, ob eine Vorgabe die Wahl eines Menschen erstickt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestWasEineRichtlinieIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schluessel string
		ist        bool
	}{
		"nur lokal":                {domain.RichtlinieNurLokal, true},
		"die Kostengrenze":         {domain.RichtlinieKostengrenze, true},
		"ein Dienstschalter":       {domain.DienstEinstellung(domain.DienstScryfall), true},
		"der zweite Faktor":        {domain.NutzerFaktorbeimWechsel, true},
		"die Sprache":              {domain.EinstellungSprache, false},
		"eine Frage an den Dienst": {domain.FrageEinstellung(domain.FrageSuche), false},
		"ein erfundener Schlüssel": {"gibtesnicht", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if domain.IstRichtlinie(fall.schluessel) != fall.ist {
				t.Errorf("%q gilt als Richtlinie: %v — erwartet war %v",
					fall.schluessel, !fall.ist, fall.ist)
			}
		})
	}
}

// **Alle neun Kombinationen eines Schalters.**
func TestDerStrengsteSchalterGewinnt(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	faelle := []struct {
		name, profil, nutzer, erwartet string
	}{
		{"beide schweigen", "", "", ""},
		{"nur das Profil schaltet an", domain.DienstAn, "", domain.DienstAn},
		{"nur das Profil schaltet aus", domain.DienstAus, "", domain.DienstAus},
		{"nur der Mensch schaltet an", "", domain.DienstAn, domain.DienstAn},
		{"nur der Mensch schaltet aus", "", domain.DienstAus, domain.DienstAus},
		{"beide an", domain.DienstAn, domain.DienstAn, domain.DienstAn},
		{"beide aus", domain.DienstAus, domain.DienstAus, domain.DienstAus},

		// **Die zwei Fälle, für die es die Faltung gibt.**
		{"das Profil lockert", domain.DienstAn, domain.DienstAus, domain.DienstAus},
		{"der Mensch lockert", domain.DienstAus, domain.DienstAn, domain.DienstAus},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gilt := domain.Verschaerft(schalter, fall.profil, fall.nutzer)
			if gilt != fall.erwartet {
				t.Errorf("es gilt %q, erwartet war %q", gilt, fall.erwartet)
			}
		})
	}
}

// **Ein Wert, den niemand kennt, entscheidet nichts.**
//
// Er käme durch `PruefeEinstellung` gar nicht herein; wer ihn trotzdem hat,
// hat ihn von außen. Ihn als „locker" zu deuten hieße, eine Schranke über
// einen Tippfehler aufzuheben.
func TestEinUnbekannterWertWirdUebergangen(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	if gilt := domain.Verschaerft(schalter, "vielleicht", domain.DienstAus); gilt != domain.DienstAus {
		t.Errorf("es gilt %q, erwartet war %q", gilt, domain.DienstAus)
	}

	if gilt := domain.Verschaerft(schalter, "vielleicht", ""); gilt != "" {
		t.Errorf("es gilt %q, erwartet war nichts", gilt)
	}
}

// **Die kleinere Obergrenze gewinnt, und leer ist keine.**
func TestDieKleinereGrenzeGewinnt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, profil, nutzer, erwartet string
	}{
		{"beide schweigen", "", "", ""},
		{"nur das Profil setzt eine", "500", "", "500"},
		{"nur der Mensch setzt eine", "", "500", "500"},
		{"der Mensch ist strenger", "500", "200", "200"},
		{"das Profil ist strenger", "200", "500", "200"},
		{"gleich hoch", "300", "300", "300"},
		{"null ist die strengste", "0", "500", "0"},
		{"keine Zahl zählt nicht mit", "viel", "500", "500"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gilt := domain.Verschaerft(
				domain.RichtlinieKostengrenze, fall.profil, fall.nutzer)
			if gilt != fall.erwartet {
				t.Errorf("es gilt %q, erwartet war %q", gilt, fall.erwartet)
			}
		})
	}
}

// **Wer keine Richtlinie ist, bekommt keine erfundene Ordnung.**
func TestOhneRichtlinieGiltDerErsteWert(t *testing.T) {
	t.Parallel()

	if gilt := domain.Verschaerft(domain.EinstellungSprache, "", "en"); gilt != "en" {
		t.Errorf("es gilt %q, erwartet war %q", gilt, "en")
	}

	if gilt := domain.Verschaerft(domain.EinstellungSprache, "de", "en"); gilt != "de" {
		t.Errorf("es gilt %q, erwartet war %q", gilt, "de")
	}

	if gilt := domain.Verschaerft(domain.EinstellungSprache); gilt != "" {
		t.Errorf("es gilt %q, erwartet war nichts", gilt)
	}
}

// **Was ein Import nicht darf.**
func TestWasEineSchrankeLockert(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	faelle := map[string]struct {
		schluessel, neu, schranke string
		lockert                   bool
	}{
		"ein Dienst gegen ein Verbot": {schalter, domain.DienstAn, domain.DienstAus, true},
		"ein Dienst gegen eine Erlaubnis": {
			schalter, domain.DienstAn, domain.DienstAn, false,
		},
		"ein Verbot gegen eine Erlaubnis": {
			schalter, domain.DienstAus, domain.DienstAn, false,
		},
		"ohne Schranke": {schalter, domain.DienstAn, "", false},
		"eine höhere Grenze": {
			domain.RichtlinieKostengrenze, "900", "200", true,
		},
		"eine niedrigere Grenze": {
			domain.RichtlinieKostengrenze, "100", "200", false,
		},
		"der zweite Faktor": {
			domain.NutzerFaktorbeimWechsel,
			domain.FaktorNurAnmelden, domain.FaktorVerlangen, true,
		},
		"keine Richtlinie": {domain.EinstellungSprache, "en", "de", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lockert := domain.Lockert(fall.schluessel, fall.neu, fall.schranke)
			if lockert != fall.lockert {
				t.Errorf("lockert: %v — erwartet war %v", lockert, fall.lockert)
			}
		})
	}
}

// **Die Kostengrenze ist eine Zahl in Cent oder gar nichts.**
func TestDieKostengrenzeWirdGeprueft(t *testing.T) {
	t.Parallel()

	gut := []string{"", "0", "1", "500", "10000000"}
	for _, wert := range gut {
		t.Run("gut "+wert, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeEinstellung(
				domain.RichtlinieKostengrenze, wert); fehler != nil {
				t.Errorf("%q wird abgewiesen: %v", wert, fehler)
			}
		})
	}

	schlecht := []string{"viel", "-1", "10000001", "5,50", "500 "}
	for _, wert := range schlecht {
		t.Run("schlecht "+wert, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeNutzereinstellung(
				domain.RichtlinieKostengrenze, wert); fehler == nil {
				t.Errorf("%q kam durch", wert)
			}
		})
	}
}

// **Eine Spaltenbreite ist eine Zahl in Punkten — im Profil auch »erben«.**
func TestDieSpaltenbreiteWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		wert                   string
		imProfil, beimMenschen bool
	}{
		{"280", true, true},
		{"120", true, true},
		{"800", true, true},
		{domain.Erben, true, false},
		{"119", false, false},
		{"801", false, false},
		{"breit", false, false},
		{"", false, false},
	}

	for _, fall := range faelle {
		t.Run("wert "+fall.wert, func(t *testing.T) {
			t.Parallel()

			imProfil := domain.PruefeEinstellung(
				domain.ArbeitsplatzNavigatorbreite, fall.wert) == nil
			if imProfil != fall.imProfil {
				t.Errorf("im Profil kommt %q durch: %v — erwartet war %v",
					fall.wert, imProfil, fall.imProfil)
			}

			beimMenschen := domain.PruefeNutzereinstellung(
				domain.ArbeitsplatzInspektorbreite, fall.wert) == nil
			if beimMenschen != fall.beimMenschen {
				t.Errorf("beim Menschen kommt %q durch: %v — erwartet war %v",
					fall.wert, beimMenschen, fall.beimMenschen)
			}
		})
	}
}
