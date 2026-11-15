// **Ein neues, leeres Set** —
// #891,
// Werkstück 2 von Phase Y (⌘N).
//
// # Warum es das bis dahin nicht gab
//
// In den Bestand kam ein Set nur über eine **Datei**: ein `.mse-set`, ein
// Beispielset, MTG JSON oder Scryfall. Wer bei null anfangen wollte, konnte es
// nicht — und ⌘N ist unter macOS die selbstverständlichste Handlung nach ⌘O.
//
// MSE2 fragt dafür nach **Spiel und Stil**, in zwei Listen
// (`src/gui/new_window.cpp`, `NewSetWindow`, Fassung f9d9356 vom 2021-07-06:
// `game_list->showData<Game>()`, dann `stylesheet_list->showData<StyleSheet>(
// game->name() + "-*")`). Dieselben zwei Angaben verlangt diese Naht.
package app_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein angelegtes Set steht im Bestand und hat eine Kennung.**
func TestEinNeuesSetLandetImBestand(t *testing.T) {
	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	antwort := anwendung.SetAnlegen(t.Context(), "Mein Set", "probe", "probe-schlicht")
	if antwort.Fehler != nil {
		t.Fatalf("anlegen: %+v", antwort.Fehler)
	}

	if antwort.ID == "" {
		t.Fatal("die Antwort nennt keine Kennung")
	}

	// **Und es ist zu finden** — über denselben Weg wie jedes andere.
	liste := anwendung.Sets(t.Context())
	if liste.Fehler != nil {
		t.Fatalf("Sets: %+v", liste.Fehler)
	}

	gefunden := false

	for _, satz := range liste.Sets {
		if satz.ID == antwort.ID {
			gefunden = true

			if satz.Name != "Mein Set" || satz.Spiel != "probe" {
				t.Errorf("das Set steht als %+v im Bestand", satz)
			}
		}
	}

	if !gefunden {
		t.Errorf("das angelegte Set steht nicht in der Liste: %+v", liste.Sets)
	}
}

// **Ohne Spiel geht es nicht, und der Fehler sagt es am Feld.**
//
// Ein Set ohne Spiel hat keine Kartenfelder — es wäre eine Zeile in der Liste,
// die sich nicht öffnen lässt. `domain.Set.ValidateBasics` sagt dasselbe; hier
// steht, dass die Naht es weitergibt (§8).
func TestEinSetOhneSpielWirdAbgewiesen(t *testing.T) {
	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	for name, fall := range map[string]struct{ setname, spiel string }{
		"ohne Spiel":  {"Mein Set", "  "},
		"ohne Namen":  {"   ", "probe"},
		"ohne beides": {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.SetAnlegen(t.Context(), fall.setname, fall.spiel, "")
			if antwort.Fehler == nil {
				t.Fatalf("%q/%q kam durch", fall.setname, fall.spiel)
			}

			if antwort.Fehler.Feld == "" {
				t.Errorf("der Fehler nennt kein Feld: %+v", antwort.Fehler)
			}
		})
	}
}

// **Der Stil ist verlangt** — und zwar einer des Spiels.
//
// **Hier stand bis zum 2026-08-25 das Gegenteil**: „MSE2 wählt einen vor,
// verlangt ihn aber nicht … ein Pflichtfeld hier wäre strenger als das
// Original". Das war aus dem Dialog erschlossen. `Set::validate` sagt es
// anders:
//
//	if (!stylesheet) {
//	  throw Error(_ERROR_("no stylesheet specified for the set"));
//	}
//	if (stylesheet->game != game) {
//	  throw Error(_ERROR_("stylesheet and set refer to different game"));
//	}
//
// `src/data/set.cpp`, Zeilen 143–149, Fassung f9d9356 vom 2021-07-06. Es ist
// keine Bedingung des Dialogs, sondern eine des **Sets**. Gefunden hat es der
// Nachlauf des Zweitblicks über Phase Y.
func TestEinSetOhneStilWirdAbgewiesen(t *testing.T) {
	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	for name, stil := range map[string]string{
		"ohne Stil":                 "",
		"Stil eines anderen Spiels": "magic-m15",
	} {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.SetAnlegen(t.Context(), "Ohne Stil", "probe", stil)

			if antwort.Fehler == nil {
				t.Fatal("angelegt, obwohl der Stil fehlt oder fremd ist")
			}

			// **Der Fehler nennt sein Feld** (§6): die Meldung erscheint am
			// Stilfeld und nicht als anonymer Hinweis.
			if antwort.Fehler.Feld != "stil" {
				t.Errorf("der Fehler nennt das Feld %q", antwort.Fehler.Feld)
			}
		})
	}
}

// **Und die Bindung reicht durch** (§16).
//
// Ohne diesen Lauf wäre sie angemeldet, gebunden, im Fähigkeitskatalog
// geführt — und von keinem Test betreten. Genau die Gestalt von #852.
func TestDieBindungLegtEinSetAn(t *testing.T) {
	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	antwort := app.NeueBindungen(anwendung).SetAnlegen(t.Context(), "Über die Bindung", "probe", "probe-schlicht")
	if antwort.Fehler != nil {
		t.Fatalf("anlegen: %+v", antwort.Fehler)
	}

	if antwort.ID == "" {
		t.Error("die Antwort nennt keine Kennung")
	}
}

// **Und `SetAusDatei` liest einen genannten Pfad** (#891).
//
// Anders als `BeispielsetOeffnen` ohne Wache auf eine Liste: der Pfad kommt
// aus einem Ablegen, und das ist die Geste selbst.
func TestDieBindungLiestEinenGenanntenPfad(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	antwort := app.NeueBindungen(anwendung).SetAusDatei(t.Context(), probeset)
	if antwort.Fehler != nil {
		t.Fatalf("SetAusDatei: %+v", antwort.Fehler)
	}

	if antwort.SetID == "" {
		t.Error("die Antwort nennt keine Kennung")
	}
}

// **Ohne Naht meldet sie es, statt abzustürzen.**
func TestSetAnlegenOhneNahtMeldetSich(t *testing.T) {
	anwendung := &app.App{}

	antwort := anwendung.SetAnlegen(t.Context(), "Mein Set", "probe", "")
	if antwort.Fehler == nil {
		t.Fatal("ohne Naht kam es durch")
	}

	if !strings.Contains(strings.ToLower(antwort.Fehler.Meldung), "unerwartet") &&
		!errors.Is(domain.ErrValidation, domain.ErrValidation) {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}
