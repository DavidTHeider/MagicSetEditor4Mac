package app_test

// **Der Profilwechsel am Fenster** — Werkstück 3 von Phase 4.12
// (#1351).

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Nach einem Wechsel liest das Fenster die Daten des neuen Profils.**
//
// # Warum das der Test ist und nicht „der Zeiger steht"
//
// **Weil der Zeiger im Bestand und das Profil am Fenster zwei Dinge sind.**
// [app.App] merkt sich seines beim Start und hängt es an jede Bindung; ein
// Wechsel, der nur den Bestand umstellt, ließe das Fenster **weiter im alten
// Profil arbeiten** — bis zum nächsten Start, und dann unerklärlich.
//
// Geprüft wird deshalb an einem Set: eines im ersten Profil angelegt, gewechselt,
// und die Liste ist leer. Das ist beobachtbares Verhalten (§10) und kein Feld.
func TestNachEinemWechselLiestDasFensterDasNeueProfil(t *testing.T) {
	anwendung, _ := geoeffnetesProbeset(t)

	if sets := anwendung.Sets(t.Context()); len(sets.Sets) != 1 {
		t.Fatalf("%d Sets im ersten Profil", len(sets.Sets))
	}

	zweites := zweitesProfil(t, anwendung)

	gewechselt := anwendung.ProfilWechseln(t.Context(), zweites, "")
	if gewechselt.Fehler != nil {
		t.Fatalf("wechseln: %v", gewechselt.Fehler)
	}

	if gewechselt.Profil == nil || gewechselt.Profil.ID != zweites {
		t.Fatalf("die Antwort nennt das Profil nicht: %+v", gewechselt.Profil)
	}

	// **Und das ist der Punkt**: dasselbe Fenster, dieselbe Bindung, ein
	// anderes Ergebnis.
	if sets := anwendung.Sets(t.Context()); len(sets.Sets) != 0 {
		t.Errorf("%d Sets im zweiten Profil — erwartet war keines", len(sets.Sets))
	}

	// **Und zurück ist es wieder da.** Ein Wechsel versteckt, er löscht nicht.
	erstes := anwendung.ProfilWechseln(t.Context(), gezeigtesErstes(t, anwendung), "")
	if erstes.Fehler != nil {
		t.Fatalf("zurückwechseln: %v", erstes.Fehler)
	}

	if sets := anwendung.Sets(t.Context()); len(sets.Sets) != 1 {
		t.Errorf("%d Sets zurück im ersten Profil", len(sets.Sets))
	}
}

// **Nach einem Wechsel folgt die Leiste** — „Zuletzt benutzt" gehört dem
// Profil (Nachlauf über Phase 6.3, N6). Bis dahin zeigte die Leiste die Sets
// des vorigen Menschen, bis jemand ein Set öffnete.
func TestNachEinemWechselFolgtDieLeiste(t *testing.T) {
	anwendung, set := geoeffnetesProbeset(t)

	if antwort := anwendung.SetBenutzt(t.Context(), set); antwort.Fehler != nil {
		t.Fatalf("SetBenutzt: %+v", antwort.Fehler)
	}

	var leiste *app.Menue

	anwendung.MitMenue(func(_ context.Context, gesetzt *app.Menue) { leiste = gesetzt })

	gewechselt := anwendung.ProfilWechseln(t.Context(), zweitesProfil(t, anwendung), "")
	if gewechselt.Fehler != nil {
		t.Fatalf("wechseln: %v", gewechselt.Fehler)
	}

	if leiste == nil {
		t.Fatal("nach dem Wechsel wurde die Leiste nicht neu gesetzt")
	}

	ablage := untermenueMitNamen(leiste, app.MenuetexteFuer("de").Ablage)
	if ablage == nil {
		t.Fatal("die Leiste führt keine Ablage")
	}

	if untermenueMitNamen(&app.Menue{Eintraege: ablage.Untermenue.Eintraege},
		app.MenuetexteFuer("de").Zuletzt) != nil {
		t.Error("»Zuletzt benutzt« zeigt die Sets des vorigen Profils")
	}
}

// **Das gezeigte Profil steht in der Übersicht.**
//
// Ohne diese Auskunft könnte die Seite nicht sagen, in welchem Profil jemand
// arbeitet — und ein Wechsel wäre eine Handlung ohne sichtbaren Zustand.
func TestDieUebersichtNenntDasGezeigteProfil(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.GezeigtesProfil(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("das gezeigte Profil: %v", antwort.Fehler)
	}

	if antwort.Profil == nil || antwort.Profil.ID == "" {
		t.Fatalf("kein Profil genannt: %+v", antwort.Profil)
	}
}

// **Ein Wechsel auf ein Profil, das es nicht gibt, ist ein Nein.**
func TestEinWechselAufEinFehlendesProfilWirdAmFensterAbgewiesen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.ProfilWechseln(t.Context(), "4711", "")
	if antwort.Fehler == nil {
		t.Fatal("das fehlende Profil kam durch")
	}

	if antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q, erwartet war nichtGefunden", antwort.Fehler.Code)
	}
}

// **Eine Kennung, die keine Zahl ist, ist ein Bedienfehler.**
func TestEineUnbrauchbareProfilkennungWirdAbgewiesen(t *testing.T) {
	if antwort := angelaufeneAnwendung(t).ProfilWechseln(
		t.Context(), "keine Zahl", ""); antwort.Fehler == nil {
		t.Error("die unbrauchbare Kennung kam durch")
	}
}

// zweitesProfil legt einen zweiten Nutzer an und gibt dessen Profilkennung.
// **Ein zweites Profil DESSELBEN Nutzers** — Werkstück 11 (#1351).
//
// **Hier stand bis zum 2026-09-06 das Profil eines zweiten Nutzers**, und die
// Läufe wurden rot, sobald es Rechte gab. Das war kein Fehler des Werkstücks,
// sondern die Zusage, die es einlöst: diese Fälle meinen den **Wechsel**, nicht
// die Erlaubnis — und sie sollen ihn allein prüfen. Wer die Erlaubnis prüfen
// will, findet sie in `internal/service/wechselerlaubnis_test.go` und im
// Flächenlauf.
func zweitesProfil(t *testing.T, anwendung *app.App) string {
	t.Helper()

	eigner := gezeigtesProfilnutzer(t, anwendung)

	angelegt := anwendung.ProfilAnlegen(t.Context(), eigner, "Turnier")
	if angelegt.Fehler != nil {
		t.Fatalf("das zweite Profil anlegen: %v", angelegt.Fehler)
	}

	return angelegt.Profil.ID
}

// gezeigtesProfilnutzer gibt die Kennung des Nutzers, der gerade arbeitet.
func gezeigtesProfilnutzer(t *testing.T, anwendung *app.App) string {
	t.Helper()

	gezeigt := anwendung.GezeigtesProfil(t.Context())
	if gezeigt.Fehler != nil {
		t.Fatalf("das gezeigte Profil lesen: %v", gezeigt.Fehler)
	}

	return gezeigt.Profil.NutzerID
}

// gezeigtesErstes gibt die Kennung des ersten Profils des ersten Nutzers.
func gezeigtesErstes(t *testing.T, anwendung *app.App) string {
	t.Helper()

	antwort := anwendung.Nutzer(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("die Nutzer lesen: %v", antwort.Fehler)
	}

	return antwort.Nutzer[0].Profile[0].ID
}

// angelaufeneAnwendung baut eine Anwendung auf einer frischen Datenbank.
func angelaufeneAnwendung(t *testing.T) *app.App {
	t.Helper()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	return anwendung
}
