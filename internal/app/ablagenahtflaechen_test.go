package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Zwischenablage für die Flächen ohne Fenster** — Werkstück 1 von
// Phase 3.7
// (#1168).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.

// **Die Naht kommt über einem echten Bestand zusammen und trägt hin und
// zurück.**
//
// **Über den Bestand und nicht über ein Doppel**: was hier zu prüfen ist, ist
// die **Wurzel** — dass `Schreibnaht` die drei Fähigkeiten wirklich mitbringt,
// die die Ablage von ihr verlangt. Ein Doppel bestätigte nur die Zusicherung,
// die es selbst erfüllt.
func TestDieAblagenahtTraegtUeberEinenEchtenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, karteID := saatsetImBestand(t, db)

	naht, nahtfehler := app.Ablageflaechennaht(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	inhalt, kopierfehler := naht.Kopieren(probekontext(t), setID, []int64{karteID})
	if kopierfehler != nil {
		t.Fatalf("kopieren: %v", kopierfehler)
	}

	if !strings.Contains(inhalt, "game_name: probe") {
		t.Fatalf("der Inhalt nennt das Spiel nicht:\n%s", inhalt)
	}

	karten, befunde, einfuegefehler := naht.Einfuegen(probekontext(t), setID, inhalt)
	if einfuegefehler != nil {
		t.Fatalf("einfügen: %v", einfuegefehler)
	}

	if len(karten) != 1 || len(befunde) != 0 {
		t.Errorf("Karten %v, Befunde %v", karten, befunde)
	}

	// **Und Ausschneiden nimmt sie mit** — dieselbe Naht, der andere Weg.
	if _, schnittfehler := naht.Ausschneiden(
		probekontext(t), setID, []int64{karteID}); schnittfehler != nil {
		t.Fatalf("ausschneiden: %v", schnittfehler)
	}
}

// **Ohne Benutzerverzeichnis wird gemeldet statt geraten.**
//
// Dieselbe Zusage wie bei der MCP-Schreibnaht darunter, und aus demselben
// Grund: die Ablage braucht den Stilwähler, der Stilwähler die Paketwurzeln,
// und die stehen im Zuhause.
func TestDieAblageflaechennahtMeldetEinenFehlendenOrt(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.Ablageflaechennaht(nil, domain.UrheberKommandozeile); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Ein Fehler des Anwendungsfalls kommt an allen drei Wegen durch.**
//
// **Alle drei und nicht einer**: drei gleichlautende Zweige sind drei, von
// denen nur einer je einen Lauf sähe (§3) — und eine Fläche ohne Fenster hätte
// dann einen stillen Erfolg, wo der Bestand belegt war.
func TestDieAblagenahtReichtIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	naht := (&app.Fremdablage{}).MitAblagequelle(
		&ablagequelle{fehler: domain.ErrBelegt})

	if _, fehler := naht.Kopieren(probekontext(t), 1, []int64{7}); fehler == nil {
		t.Error("kopieren kam durch")
	}

	if _, fehler := naht.Ausschneiden(probekontext(t), 1, []int64{7}); fehler == nil {
		t.Error("ausschneiden kam durch")
	}

	// **Und Einfügen gibt dabei leere Listen und nicht `nil`.** Eine Fläche,
	// die die Antwort als JSON schreibt, machte aus `nil` ein `null` — und
	// jeder Leser bräuchte einen Zweig dafür.
	karten, befunde, fehler := naht.Einfuegen(probekontext(t), 1, "game_name: probe\n")
	if fehler == nil {
		t.Error("einfügen kam durch")
	}

	if karten == nil || befunde == nil {
		t.Errorf("Karten %v, Befunde %v — erwartet leere Listen", karten, befunde)
	}
}

// **Und ein halbes Ergebnis kommt mit seinem Fehler zurück.**
//
// Scheitert das Füllen einer angelegten Karte, ist die Karte trotzdem da; wer
// sie nicht nennt, lässt sie im Bestand zurück, ohne dass jemand davon weiß.
func TestDieAblagenahtNenntDieKartenAuchBeiEinemFehler(t *testing.T) {
	t.Parallel()

	naht := (&app.Fremdablage{}).MitAblagequelle(&ablagequelle{
		ergebnis: &service.Einfuegung{
			Karten:  []int64{11},
			Befunde: []domain.Befund{{Meldung: "das Feld gibt es nicht"}},
		},
		fehler: domain.ErrValidation,
	})

	karten, befunde, fehler := naht.Einfuegen(probekontext(t), 1, "game_name: probe\n")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(karten) != 1 || karten[0] != "11" || len(befunde) != 1 {
		t.Errorf("Karten %v, Befunde %v", karten, befunde)
	}
}
