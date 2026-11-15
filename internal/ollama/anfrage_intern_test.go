package ollama

// **Eine Anfrage, die sich nicht schreiben lässt** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum der Lauf im Paket steht
//
// **Weil kein öffentlicher Ruf dorthin kommt.** Jeder Aufrufer von [hole]
// reicht einen Verbund oder eine Karte aus Zeichenketten, und beides lässt
// sich immer schreiben. Der Zweig fängt den **nächsten** Aufrufer ab — und
// ein Zweig, den niemand betreten hat, ist eine Behauptung (§10).
//
// **Was er belegt**, ist eine Aussage über die Gestalt des Fehlers: ein
// unschreibbarer Rumpf endet als Fehler dieses Programms und nicht als
// Absturz und nicht als leere Anfrage an die Laufzeit.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestEinUnschreibbarerRumpfWirdZumFehler(t *testing.T) {
	t.Parallel()

	klient := NeuerKlient("http://127.0.0.1:1", netz.NeuerKlient(netz.Fristen{}))

	// **Ein Kanal lässt sich nicht als JSON schreiben** — `json.Marshal`
	// nennt ihn ausdrücklich als nicht unterstützt.
	_, fehler := klient.frage(t.Context(), "http://127.0.0.1:1/api/chat",
		make(chan int))

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
