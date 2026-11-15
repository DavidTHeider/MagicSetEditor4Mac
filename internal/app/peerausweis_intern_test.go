package app

// **Die Identität über die Fläche, die `mse-rest` benutzt** — Werkstück N6
// von Phase 5.4
// (#1572).
//
// # Warum dieser Lauf dasteht
//
// **Weil [App.Ausweis] seit N6 exportiert ist** und damit zwei Aufrufer hat,
// die kein Fenster sind: `cmd/mse-rest` und `cmd/mse-mcp` bauen daraus ihr
// TLS-Zeugnis. Ein Weg, den zwei Programme gehen, ohne dass ihn ein Lauf
// betritt, wäre eine Naht auf Vertrauensbasis.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Die Fläche gibt dieselbe Identität, die das Profil führt.**
func TestDieAusweisflaecheGibtDieIdentitaetDesProfils(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := riegelkontext(t)

	// **Ohne Identität sagt sie es, statt eine leere zu geben.**
	if _, fehler := anwendung.Ausweis(ctx); fehler == nil {
		t.Fatal("ohne Identität kam eine zurück")
	}

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	// **Und es ist wirklich die des Profils** — der Fingerabdruck entsteht
	// aus der Kennung, und der Dienst sagt genau ihn an.
	if ich.Fingerabdruck() != peer.Fingerabdruck(ich.Kennung()) {
		t.Errorf("der Abdruck passt nicht zur Kennung: %q", ich.Fingerabdruck())
	}
}
