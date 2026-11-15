package mtgjson

import (
	"encoding/json"
	"testing"
)

// **Wer die Gruppen bildet, entscheidet über die Karten** —
// #1038.
//
// # Warum dieser Lauf innen sitzt und nicht am Leser
//
// **Weil zwei Gegenproben am Leser grün blieben.** Die Bedingung „eine
// Vorderseite nennt **genau eine** andere Seite" wurde zu „mindestens eine"
// aufgeweicht — und das Endergebnis stimmte trotzdem, weil spätere Durchläufe
// `gehoertZu` überschreiben und den Fehler zufällig heilen.
//
// **Zufällig ist kein Zustand, auf den man baut.** Der Unterschied ist in den
// Gruppen sichtbar, nicht im Ergebnis; also wird er dort geprüft (§10).
func TestNurEineVorderseiteBildetEineGruppe(t *testing.T) {
	t.Parallel()

	// **Die gemessene Gestalt** (EMN.json, 2026-08-24): zwei Vorderseiten mit
	// je einer Kennung, eine geteilte Rückseite mit zweien.
	rohe := []map[string]json.RawMessage{
		rohkarte(t, `{"uuid":"b1","side":"b","otherFaceIds":["a1","a2"]}`),
		rohkarte(t, `{"uuid":"a1","side":"a","otherFaceIds":["b1"]}`),
		rohkarte(t, `{"uuid":"a2","side":"a","otherFaceIds":["b1"]}`),
	}

	gruppen, gehoertZu, _ := gruppiereSeiten(rohe)

	// **Die Rückseite ist kein Kopf.** Wäre sie einer, entstünde eine Karte
	// „Das Ganze" mit einer Vorderseite als zweiter Seite.
	if _, kopf := gruppen[0]; kopf {
		t.Errorf("die geteilte Rückseite bildet eine eigene Gruppe: %v", gruppen[0])
	}

	// **Beide Vorderseiten sind Köpfe**, und jede führt die Rückseite mit.
	for _, vorderseite := range []int{1, 2} {
		mitglieder, kopf := gruppen[vorderseite]
		if !kopf {
			t.Errorf("die Vorderseite %d bildet keine Gruppe", vorderseite)

			continue
		}

		if len(mitglieder) != 2 || mitglieder[1] != 0 {
			t.Errorf("Gruppe %d = %v, erwartet [%d 0]", vorderseite, mitglieder, vorderseite)
		}
	}

	// **Und die Rückseite steht nicht für sich** — sie hat einen Kopf.
	if _, da := gehoertZu[0]; !da {
		t.Error("die geteilte Rückseite würde eine eigene Karte")
	}
}

// **Ein gegenseitiges Paar wird einmal gruppiert, nicht zweimal.**
//
// Bei `transform` zeigen beide Seiten aufeinander; wer für jede eine Gruppe
// bildete, bekäme die Karte doppelt.
func TestEinGegenseitigesPaarBildetEineGruppe(t *testing.T) {
	t.Parallel()

	rohe := []map[string]json.RawMessage{
		rohkarte(t, `{"uuid":"a","side":"a","otherFaceIds":["b"]}`),
		rohkarte(t, `{"uuid":"b","side":"b","otherFaceIds":["a"]}`),
	}

	gruppen, _, _ := gruppiereSeiten(rohe)

	if len(gruppen) != 1 {
		t.Errorf("%d Gruppen, erwartet eine: %v", len(gruppen), gruppen)
	}

	if _, kopf := gruppen[0]; !kopf {
		t.Error("die kleinere Stelle führt nicht")
	}
}

// rohkarte macht aus einer Zeile eine rohe Karte.
func rohkarte(t *testing.T, zeile string) map[string]json.RawMessage {
	t.Helper()

	var roh map[string]json.RawMessage
	if err := json.Unmarshal([]byte(zeile), &roh); err != nil {
		t.Fatalf("Probe unlesbar: %v", err)
	}

	return roh
}
