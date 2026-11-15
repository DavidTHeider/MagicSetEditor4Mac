package app

// **Wem ein Abbruchweg gehört** — Befund 3 des Zweitblicks über Phase 5.3.
//
// # Der Befund
//
// `bildlaeufe.anmelden` schrieb den Abbruch **je Karte**, und sein Rückgeber
// löschte den Eintrag, sobald dort **irgendetwas** stand. Der Kommentar
// daneben sagte das Gegenteil: „Nur den eigenen Eintrag entfernen. Hat ein
// zweiter Auftrag ihn überschrieben, gehört er ihm."
//
// **Ein Kopf, der mehr verspricht als sein Rumpf hält** (§10) — und der
// Schaden ist nicht der überschriebene erste Auftrag, den die Datei
// eingestand, sondern der **zweite**: er verlor seinen Abbruchweg, sobald der
// erste zu Ende lief.
//
// # Und was der Nachlauf daran noch fand
//
// **Die Identität schützte nur den jüngeren Lauf.** Der ältere blieb
// überschrieben und damit unabberechenbar — der Fall war behoben, die Klasse
// nicht. Ein zweiter Lauf an derselben Karte wird jetzt **abgewiesen**, wie
// §7, Regel 4 es verlangt: ein Vorgang je Datensatz.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinZweiterLaufAnDerselbenKarteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	var laeufe bildlaeufe

	_, ersterFertig, fehler := laeufe.anmelden(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("der erste Lauf: %v", fehler)
	}

	// **Der zweite kommt nicht durch** (§7, Regel 4). Vorher übernahm er den
	// Abbruchweg der Karte, und der erste lief ohne einen weiter — bezahlt
	// und nicht anzuhalten.
	_, zweiterFertig, zweiterFehler := laeufe.anmelden(t.Context(), 7)
	if !errors.Is(zweiterFehler, domain.ErrConflict) {
		t.Fatalf("der zweite Lauf kam durch: %v", zweiterFehler)
	}

	if zweiterFertig != nil {
		t.Error("ein abgewiesener Lauf gab einen Rückgeber")
	}

	// **Und der erste bleibt abbrechbar.**
	if !laeufe.abbrechen(7) {
		t.Fatal("der erste Auftrag hat seinen Abbruchweg verloren")
	}

	ersterFertig()
}

// **Nach dem Ende des ersten geht ein zweiter wieder.**
//
// **Die Gegenprobe** (§10): eine Sperre, die nie aufginge, machte aus einem
// Auftrag je Karte einen Auftrag je Karte und Programmlauf.
func TestNachDemEndeGehtEinZweiterLaufWieder(t *testing.T) {
	t.Parallel()

	var laeufe bildlaeufe

	_, ersterFertig, fehler := laeufe.anmelden(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("der erste Lauf: %v", fehler)
	}

	ersterFertig()

	_, zweiterFertig, zweiterFehler := laeufe.anmelden(t.Context(), 7)
	if zweiterFehler != nil {
		t.Fatalf("nach dem Ende kam kein zweiter Lauf durch: %v", zweiterFehler)
	}

	zweiterFertig()
}

// **Und wenn der zweite zuerst endet, ist danach nichts mehr offen.**
//
// **Die Gegenprobe**: ein Eintrag, der nur noch dem eigenen Lauf gehörte,
// dürfte nicht liegen bleiben — sonst zeigte `abbrechen` auf einen Auftrag,
// den es nicht mehr gibt.
func TestNachDemLetztenLaufIstNichtsMehrOffen(t *testing.T) {
	t.Parallel()

	var laeufe bildlaeufe

	_, fertig, fehler := laeufe.anmelden(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("der Lauf: %v", fehler)
	}

	fertig()

	if laeufe.abbrechen(7) {
		t.Fatal("nach dem Lauf stand noch ein Abbruchweg offen")
	}
}

// **Der abgebrochene Kontext ist der des eigenen Laufs.**
//
// **Sonst wäre die Identität am Eintrag richtig und am Kontext falsch**: ein
// `fertig`, das den fremden Abbruch riefe, beendete den laufenden Auftrag.
func TestJederLaufBrichtSeinenEigenenKontextAb(t *testing.T) {
	t.Parallel()

	var laeufe bildlaeufe

	ersterCtx, ersterFertig, fehler := laeufe.anmelden(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("der erste Lauf: %v", fehler)
	}

	// **Eine andere Karte** — sie hat ihren eigenen Eintrag.
	zweiterCtx, zweiterFertig, zweiterFehler := laeufe.anmelden(t.Context(), 8)
	if zweiterFehler != nil {
		t.Fatalf("der zweite Lauf: %v", zweiterFehler)
	}

	ersterFertig()

	if ersterCtx.Err() == nil {
		t.Error("der erste Kontext lebt nach seinem Ende weiter")
	}

	if zweiterCtx.Err() != nil {
		t.Error("das Ende des ersten Laufs brach den zweiten ab")
	}

	zweiterFertig()

	if zweiterCtx.Err() == nil {
		t.Error("der zweite Kontext lebt nach seinem Ende weiter")
	}
}
