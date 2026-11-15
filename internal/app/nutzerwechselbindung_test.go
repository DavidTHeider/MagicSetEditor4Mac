//go:build darwin

// **Wer arbeitet, und der Wechsel dorthin — an der Naht** — Werkstück 17 von
// Phase 4.12 (#1400).
//
// # Warum über die **gestartete** Anwendung
//
// **Weil `Starten` die Nutzerverwaltung baut** und die Frage seit diesem
// Werkstück an einem Zeiger im Bestand hängt. Ein eingesetztes Doppel prüfte
// die Übersetzung und ließe offen, ob die Verdrahtung steht (§10).

package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der ganze Weg: ansehen, wechseln, wieder ansehen.**
func TestDerNutzerwechselGehtUeberDieBindungHinUndZurueck(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	erste := anwendung.WerArbeitet(t.Context())
	if erste.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", erste.Fehler)
	}

	if erste.NutzerID == "" || erste.NutzerID == "0" {
		t.Fatalf("es arbeitet niemand: %+v", erste)
	}

	// **Ein zweiter Mensch, in denselben Bestand.**
	angelegt := anwendung.NutzerAnlegen(t.Context(), "Zweiter", "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	gewechselt := anwendung.NutzerWechseln(t.Context(), angelegt.Nutzer.ID, "")
	if gewechselt.Fehler != nil {
		t.Fatalf("wechseln: %+v", gewechselt.Fehler)
	}

	// **Der Zeiger zieht mit** — das Profil gehört jetzt dem neuen Menschen.
	if gewechselt.Profil.NutzerID != angelegt.Nutzer.ID {
		t.Errorf("das Profil gehört %q, erwartet war %q",
			gewechselt.Profil.NutzerID, angelegt.Nutzer.ID)
	}

	zweite := anwendung.WerArbeitet(t.Context())
	if zweite.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", zweite.Fehler)
	}

	if zweite.NutzerID != angelegt.Nutzer.ID {
		t.Errorf("es arbeitet %q, erwartet war %q",
			zweite.NutzerID, angelegt.Nutzer.ID)
	}

	// **Und zurück.**
	if zurueck := anwendung.NutzerWechseln(
		t.Context(), erste.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}
}

// **Eine krumme Kennung wird abgewiesen, bevor irgendetwas geschieht.**
func TestEineKrummeKennungAmNutzerwechselWirdAbgewiesen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	antwort := anwendung.NutzerWechseln(t.Context(), "keine Zahl", "")

	if antwort.Fehler == nil {
		t.Fatal("die krumme Kennung ging durch")
	}

	if antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("Code %q, erwartet war %q", antwort.Fehler.Code, app.CodeValidierung)
	}
}

// **Und einen Nutzer, den es nicht gibt, wird niemand.**
func TestEinNutzerwechselAufNichtsKommtAlsFehlerAn(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	antwort := anwendung.NutzerWechseln(t.Context(), "9999", "")

	if antwort.Fehler == nil {
		t.Fatal("der Wechsel auf einen Unbekannten glückte")
	}

	if antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q, erwartet war %q",
			antwort.Fehler.Code, app.CodeNichtGefunden)
	}
}

// **Und dieselbe Fähigkeit über die Naht ohne Fenster.**
//
// **Nicht dieselbe Zeile zweimal**: `Nutzerflaeche` ist der Weg, den
// Kommandozeile, MCP-Server und Dienst nehmen (§16). Ein Lauf, der nur die
// Bindung prüft, ließe drei Flächen ungeprüft — und der Flächenlauf zählt
// nicht zur Abdeckung dieses Pakets.
func TestDerNutzerwechselGehtAuchOhneFenster(t *testing.T) {
	eigenerBund(t)

	db, _ := offenerBestand(t)

	flaeche := app.NutzerflaecheUeber(db)

	erste := flaeche.WerArbeitet(t.Context())
	if erste.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", erste.Fehler)
	}

	angelegt := flaeche.NutzerAnlegen(t.Context(), "Zweiter", "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	gewechselt := flaeche.NutzerWechseln(t.Context(), angelegt.Nutzer.ID, "")
	if gewechselt.Fehler != nil {
		t.Fatalf("wechseln: %+v", gewechselt.Fehler)
	}

	zweite := flaeche.WerArbeitet(t.Context())
	if zweite.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", zweite.Fehler)
	}

	if zweite.NutzerID != angelegt.Nutzer.ID {
		t.Errorf("es arbeitet %q, erwartet war %q",
			zweite.NutzerID, angelegt.Nutzer.ID)
	}
}

// **Und jeder Lesefehler kommt als Fehler an, nicht als leere Antwort.**
//
// **Eine leere Auskunft sähe aus wie „niemand arbeitet"** (§8) — und die
// Oberfläche zeigte dann jeden Knopf, als wäre man niemand.
func TestJederLesefehlerAmNutzerwechselKommtAn(t *testing.T) {
	t.Parallel()

	t.Run("wer arbeitet, ist nicht zu lesen", func(t *testing.T) {
		t.Parallel()

		anwendung := app.NeueApp("")
		anwendung.MitProfilwechsel(&wechseldoppel{fehler: errors.New("klemmt")})

		antwort := anwendung.WerArbeitet(t.Context())

		if antwort.Fehler == nil {
			t.Fatalf("kein Fehler, gemeldet wurde %q", antwort.NutzerID)
		}

		if antwort.NutzerID != "" {
			t.Errorf("bei einem Fehler soll niemand genannt werden: %q",
				antwort.NutzerID)
		}
	})

	// **Der Wechsel glückt, und danach klemmt das Nachsehen.**
	//
	// **Der Zweig steht da, weil die Antwort das neue Profil trägt** (§7,
	// Regel 2): sie wird nach dem Wechsel gelesen, nicht behauptet — und
	// dieses Lesen kann scheitern.
	t.Run("das neue Profil ist nicht zu lesen", func(t *testing.T) {
		t.Parallel()

		anwendung := app.NeueApp("")
		anwendung.MitProfilwechsel(&wechseldoppel{
			gezeigt:     domain.Profil{ID: 1, NutzerID: 1, Name: "Meins"},
			nachWechsel: errors.New("klemmt"),
		})

		antwort := anwendung.NutzerWechseln(t.Context(), "2", "")

		if antwort.Fehler == nil {
			t.Fatal("der Wechsel meldete Erfolg ohne Profil")
		}

		if antwort.Profil != nil {
			t.Errorf("bei einem Fehler soll kein Profil genannt werden: %+v",
				antwort.Profil)
		}
	})
}
