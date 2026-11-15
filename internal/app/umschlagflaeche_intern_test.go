package app

// **Die kopflose Anwendung für den Umschlag** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Was hier geprüft wird
//
// **Dass sie steht, und dass sie genau die Nähte hat, die der Umschlag
// braucht** — den Setschreiber und den Importeur. Ohne sie fiele jeder Ruf
// von der Kommandozeile, vom MCP-Server und vom Dienst in einen `nil`-Zeiger,
// und das sähe aus wie ein Absturz und nicht wie eine fehlende Naht.
//
// **Der Flächenlauf misst, was daraus wird** (`internal/flaechen`); hier steht
// nur, dass der Bauweg selbst hält.

import (
	"testing"
)

// **Die Fläche entsteht und trägt ihre beiden Nähte.**
func TestDieUmschlagflaecheStehtMitIhrenNaehten(t *testing.T) {
	t.Parallel()

	anwendung, fehler := UmschlagflaecheUeber(bestandFuerDenRiegel(t))
	if fehler != nil {
		t.Fatalf("bauen: %v", fehler)
	}

	if anwendung.speicherer == nil {
		t.Error("ohne Setschreiber kann sie nichts schnüren")
	}

	if anwendung.importeur == nil {
		t.Error("ohne Importeur kann sie nichts hereinnehmen")
	}
}

// **Ohne Bilderverzeichnis entsteht sie nicht — und sagt es.**
//
// # Warum das einen Lauf wert ist
//
// **Weil der einzige Fehlerweg dieses Bauwegs von einem Ort auf der Platte
// abhängt** und kein fachlicher Ruf ihn betritt. [Setschreibnaht] braucht das
// Bilderverzeichnis; steht es nicht, gäbe eine Fläche ohne diesen Zweig eine
// Anwendung ohne Bildablage heraus — und die schriebe ein Set, dessen Karten
// leer sind (§3, „eine Vorrichtung, die großzügiger ist als der Betrieb").
func TestOhneBilderverzeichnisEntstehtKeineUmschlagflaeche(t *testing.T) {
	// **`HOME` leer heißt: es gibt kein Benutzerverzeichnis.** Dasselbe
	// Mittel wie in den übrigen Läufen über Orte auf der Platte.
	t.Setenv("HOME", "")

	if _, fehler := UmschlagflaecheUeber(bestandFuerDenRiegel(t)); fehler == nil {
		t.Error("ohne Bilderverzeichnis entstand eine Fläche")
	}
}
