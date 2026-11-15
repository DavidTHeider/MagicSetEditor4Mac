package main

// **Der Profilwechsel an der Kommandozeile** — Werkstück 3 von Phase 4.12
// (#1351).

import (
	"encoding/json"
	"strings"
	"testing"
)

// **`mse profil` nennt das Profil, in dem gearbeitet wird.**
func TestProfilNenntDasGezeigteProfil(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "profil")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Mein Profil") {
		t.Errorf("die Ausgabe nennt das Profil nicht: %q", aus)
	}
}

// **`mse profil <kennung>` wechselt — und das Set ist danach weg.**
//
// **Am Set und nicht am Zeiger** (§10): eine Abnahme ist beobachtbares
// Verhalten. „Der Zeiger steht" sagt nichts darüber, ob die nächste Abfrage
// ihn benutzt.
func TestProfilWechseltUndDieSetsSindAndere(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, aus, _ := fahreAn(t, ort, "sets"); code != CodeGut ||
		!strings.Contains(aus, "Probeset") {
		t.Fatalf("vor dem Wechsel: %d %q", code, aus)
	}

	// **Ein zweites Profil DESSELBEN Nutzers** — Werkstück 11 (#1351): unter
	// den eigenen zu wechseln braucht keine Freigabe, in ein fremdes schon.
	//
	// **Hier stand bis zum 2026-09-06 das Profil eines zweiten Nutzers**, und
	// der Lauf wurde rot, sobald es Rechte gab. Das war kein Fehler des
	// Werkstücks, sondern die Zusage, die es einlöst: dieser Fall meint den
	// **Wechsel**, nicht die Erlaubnis — und er soll ihn allein prüfen.
	zweites := zweitesEigenesProfil(t, ort)

	if code, _, fehler := fahreAn(t, ort, "profil", zweites); code != CodeGut {
		t.Fatalf("wechseln: %d (%s)", code, fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "sets")
	if code != CodeGut {
		t.Fatalf("nach dem Wechsel: %d (%s)", code, fehler)
	}

	if strings.Contains(aus, "Probeset") {
		t.Errorf("das Set des anderen Profils steht da: %q", aus)
	}
}

// **`--profil` gilt für einen einzigen Aufruf und ändert den Zeiger nicht.**
//
// Das Issue verlangt beides nebeneinander: der gewählte Stand überlebt den
// Aufruf, **und** ein Skript darf sich davon unabhängig machen.
func TestProfilFuerEinenAufrufAendertDenZeigerNicht(t *testing.T) {
	ort, _, _ := mitKarte(t)

	zweites := zweitesProfilAnDerZeile(t, ort)

	code, aus, fehler := fahreAn(t, ort, "--profil="+zweites, "sets")
	if code != CodeGut {
		t.Fatalf("mit --profil: %d (%s)", code, fehler)
	}

	if strings.Contains(aus, "Probeset") {
		t.Errorf("--profil wirkte nicht: %q", aus)
	}

	// **Und danach ist alles wie vorher.**
	if code, aus, _ := fahreAn(t, ort, "sets"); code != CodeGut ||
		!strings.Contains(aus, "Probeset") {
		t.Errorf("--profil hat den Zeiger verstellt: %d %q", code, aus)
	}
}

// **Ein Profil, das es nicht gibt, ist ein fachliches Nein.**
func TestEinWechselAufEinFehlendesProfilAnDerZeile(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "profil", "4711"); code != CodeFachlich {
		t.Errorf("Rückgabewert %d (%s), erwartet war CodeFachlich", code, fehler)
	}
}

// **Und `--profil` mit einem fehlenden Profil ebenso.**
func TestEinFehlendesProfilFuerEinenAufruf(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "--profil=4711", "sets"); code == CodeGut {
		t.Errorf("Rückgabewert %d (%s), erwartet war ein Fehler", code, fehler)
	}
}

// zweitesProfilAnDerZeile legt einen zweiten Nutzer an und gibt seine
// Profilkennung.
func zweitesProfilAnDerZeile(t *testing.T, ort string) string {
	t.Helper()

	code, aus, fehler := fahreAn(t, ort, "nutzer-neu", "Zweiter", "--json")
	if code != CodeGut {
		t.Fatalf("den zweiten Nutzer anlegen: %d (%s)", code, fehler)
	}

	var antwort struct {
		Nutzer struct {
			Profile []struct {
				ID string `json:"id"`
			} `json:"profile"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("die Antwort lesen: %v — %s", err, aus)
	}

	if len(antwort.Nutzer.Profile) == 0 {
		t.Fatalf("der zweite Nutzer hat kein Profil: %s", aus)
	}

	return antwort.Nutzer.Profile[0].ID
}

// zweitesEigenesProfil legt dem **ersten** Nutzer ein zweites Profil an.
func zweitesEigenesProfil(t *testing.T, ort string) string {
	t.Helper()

	code, aus, fehler := fahreAn(t, ort, "profil-neu", "1", "Turnier", "--json")
	if code != CodeGut {
		t.Fatalf("das zweite Profil anlegen: %d (%s)", code, fehler)
	}

	var antwort struct {
		Profil struct {
			ID string `json:"id"`
		} `json:"profil"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("die Antwort lesen: %v — %s", err, aus)
	}

	return antwort.Profil.ID
}
