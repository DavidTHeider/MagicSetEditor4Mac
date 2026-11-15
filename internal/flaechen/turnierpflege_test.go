// **Auflisten, abmelden, löschen — an der Kommandozeile** — Werkstück 5 von
// Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, weil `turniere_test.go` mit ihr über die Zeilengrenze
// aus §2 ginge. Der Schnitt folgt der Frage: dort steht der **ganze Ablauf**
// eines Turniers, hier stehen die Wege **daneben**.
package flaechen_test

import (
	"strings"
	"testing"
)

// **Auflisten, abmelden, löschen — an der Kommandozeile.**
//
// **Ein eigener Lauf**, weil der große Ablauf sie nicht braucht: er listet
// nicht auf, meldet niemanden ab und löscht nichts. Eine Fähigkeit ohne Lauf
// ist nach §16 keine — und der Wächter hat genau diese drei angestrichen.
func TestAuflistenAbmeldenUndLoeschenAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	turnierID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"turnier", "neu", "Küchentisch", "magic"))

	teilnehmerID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"turnier", "anmelden", turnierID, "Ada"))

	// **Auflisten.**
	if liste := mussKommandozeile(t, umgebung, "turniere"); !strings.Contains(
		liste, "Küchentisch",
	) {
		t.Errorf("die Liste nennt das Turnier nicht: %s", liste)
	}

	// **Abmelden — und er bleibt sichtbar.**
	mussKommandozeile(t, umgebung, "turnier", "aussteigen", turnierID, teilnehmerID)

	gezeigt := mussKommandozeile(t, umgebung, "turnier", turnierID)
	if !strings.Contains(gezeigt, "ausgestiegen") {
		t.Errorf("der Ausgestiegene ist nicht als solcher zu sehen: %s", gezeigt)
	}

	// **Und wieder anmelden.**
	mussKommandozeile(t, umgebung, "turnier", "einsteigen", turnierID, teilnehmerID)

	wieder := mussKommandozeile(t, umgebung, "turnier", turnierID)
	if strings.Contains(wieder, "ausgestiegen") {
		t.Errorf("er ist immer noch ausgestiegen: %s", wieder)
	}

	// **Löschen — samt Verlauf.**
	mussKommandozeile(t, umgebung, "turnier", "loeschen", turnierID)

	if uebrig := mussKommandozeile(t, umgebung, "turniere"); strings.Contains(
		uebrig, "Küchentisch",
	) {
		t.Errorf("das Turnier steht noch da: %s", uebrig)
	}
}
