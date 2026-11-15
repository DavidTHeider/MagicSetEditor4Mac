package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verweis für die Flächen ohne Fenster** — Werkstück 2 von Phase 3.6
// (#1167).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.

// **Die Naht übersetzt in dieselbe Gestalt wie das Fenster.**
func TestDieVerweisnahtGibtDieselbeGestalt(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()

	naht := (&app.Fremdverweise{}).MitVerweisquelle(quelle)

	verweis, fehler := naht.Setzen(probekontext(t), 7, "516e137a")
	if fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if verweis.KarteID != "7" || verweis.Name != "Lightning Bolt" ||
		verweis.Gesetzt != "2026-08-27T12:00:00Z" {
		t.Errorf("Verweis: %+v", verweis)
	}
}

// **Ohne Verweis kommt keiner, und das ist kein Fehler.**
func TestDieVerweisnahtGibtOhneVerweisKeinen(t *testing.T) {
	t.Parallel()

	naht := (&app.Fremdverweise{}).MitVerweisquelle(neueVerweisquelle())

	verweis, fehler := naht.Fuer(probekontext(t), 7)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if verweis != nil {
		t.Errorf("es kam %+v zurück", *verweis)
	}
}

// **Und mit Verweis kommt er in derselben Gestalt.**
//
// Der Zweig steht neben dem darüber: `Fuer` gibt drei Dinge zurück — einen
// Fehler, keinen Verweis, oder einen. Ohne diesen Lauf bliebe der dritte
// ungeprüft, und das ist der, den es gibt.
func TestDieVerweisnahtGibtDenVorhandenenVerweis(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	naht := (&app.Fremdverweise{}).MitVerweisquelle(quelle)

	if _, fehler := naht.Setzen(probekontext(t), 7, "516e137a"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	verweis, fehler := naht.Fuer(probekontext(t), 7)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if verweis == nil || verweis.Name != "Lightning Bolt" {
		t.Errorf("gelesen wurde %v", verweis)
	}
}

// **Lösen und die Setliste gehen denselben Weg.**
func TestDieVerweisnahtLoestUndZaehltAuf(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	naht := (&app.Fremdverweise{}).MitVerweisquelle(quelle)

	if _, fehler := naht.Setzen(probekontext(t), 7, "516e137a"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	verweise, fehler := naht.JeSet(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("je Set: %v", fehler)
	}

	if len(verweise) != 1 || verweise[0].KarteID != "7" {
		t.Errorf("es kamen %+v", verweise)
	}

	if fehler := naht.Loesen(probekontext(t), 7); fehler != nil {
		t.Fatalf("lösen: %v", fehler)
	}

	if len(quelle.verweise) != 0 {
		t.Errorf("es blieb %v zurück", quelle.verweise)
	}
}

// **Ein Fehler des Dienstes kommt an jeder der vier Methoden durch.**
//
// **Alle vier und nicht eine**: vier gleichlautende Zweige sind vier, von
// denen nur einer je einen Lauf sähe (§3) — und eine Fläche ohne Fenster
// hätte dann einen stillen Erfolg, wo der Bestand belegt war.
func TestDieVerweisnahtReichtIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	quelle := neueVerweisquelle()
	quelle.fehler = domain.ErrBelegt

	naht := (&app.Fremdverweise{}).MitVerweisquelle(quelle)

	if _, fehler := naht.Setzen(probekontext(t), 7, "abc"); fehler == nil {
		t.Error("setzen kam durch")
	}

	if _, fehler := naht.Fuer(probekontext(t), 7); fehler == nil {
		t.Error("lesen kam durch")
	}

	if fehler := naht.Loesen(probekontext(t), 7); fehler == nil {
		t.Error("lösen kam durch")
	}

	if _, fehler := naht.JeSet(probekontext(t), 1); fehler == nil {
		t.Error("die Setliste kam durch")
	}
}

// **Und die Naht kommt zusammen, wenn es ein Zuhause gibt — sonst nicht.**
//
// Dieselbe Zusage wie bei der Scryfallnaht daneben: der Zwischenspeicher
// braucht ein Verzeichnis, und ohne Zuhause gibt es keines.
func TestDieVerweisflaechennahtBrauchtEinZuhause(t *testing.T) {
	t.Run("mit Zuhause", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		naht, fehler := app.Verweisflaechennaht(nil)
		if fehler != nil {
			t.Fatalf("Naht: %v", fehler)
		}

		if naht == nil {
			t.Error("die Naht ist leer")
		}
	})

	t.Run("ohne Zuhause", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("MSE_BESTAND", "")

		if _, fehler := app.Verweisflaechennaht(nil); fehler == nil {
			t.Error("ohne Zuhause kam eine Naht")
		}
	})
}
