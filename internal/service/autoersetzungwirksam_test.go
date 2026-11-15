package service_test

// **Die wirksamen Regeln und die Fehlerwege der Pflege** — Werkstück E von
// Phase 4.11
// (#1341).
//
// **Eine eigene Datei**, weil `autoersetzungspflege_test.go` sonst über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **was**
// eine Regel ist und wie sie gemischt wird, hier, **welche greift**.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Lesefehler an den Regeln hält die Liste an — nicht erst der Schalter.**
//
// Zwei Rufe, zwei Fehlerwege: ohne diesen Fall bliebe der zweite unbetreten,
// und eine halbe Liste käme zurück.
func TestEinRegelfehlerHaeltDieListeAn(t *testing.T) {
	t.Parallel()

	eigener := errors.New("die Regeln sind nicht lesbar")

	ablage, spiele := probeersetzung()
	ablage.regelfehler = eigener

	if _, err := service.NeueAutoersetzungspflege(ablage, spiele).
		Liste(t.Context(), 1); !errors.Is(err, eigener) {
		t.Errorf("der Fehler lautet %v", err)
	}
}

// **Die wirksamen Regeln: der Schalter, dann `aktiv`.**
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
//	if (!gs.use_auto_replace) return false;   // Zeile 1034
//	…
//	if (!ar->enabled) continue;               // Zeile 1038
//
// `src/gui/value/text.cpp`, Fassung f9d9356 vom 2021-07-06. **Zwei Filter,
// und der erste ist der ganze.**
func TestDieWirksamenRegelnFolgenBeidenFiltern(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()
	// **Beide Muster des Pakets werden überlagert**, das eine an, das andere
	// aus — und eine dritte, eigene, kommt hinzu.
	ablage.regeln = []domain.Autoersetzungsregel{
		{Muster: "(W)", Ersatz: "weiss", GanzesWort: true, Aktiv: true},
		{Muster: "(U)", Ersatz: "blau", Aktiv: false},
		{Muster: "!!", Ersatz: "…", Aktiv: true},
	}

	pflege := service.NeueAutoersetzungspflege(ablage, spiele)

	wirksam, err := pflege.Wirksame(t.Context(), 1)
	if err != nil {
		t.Fatalf("Wirksame: %v", err)
	}

	// **Zwei von dreien**: `(U)` ist abgeschaltet.
	if len(wirksam) != 2 {
		t.Fatalf("wirksam sind %+v", wirksam)
	}

	for _, regel := range wirksam {
		if regel.Muster == "(U)" && regel.Ersatz == "blau" {
			t.Error("die abgeschaltete Regel ist wirksam")
		}
	}

	// **Und aus heißt aus** — auch für die des Pakets.
	ablage.an = false

	aus, ausFehler := pflege.Wirksame(t.Context(), 1)
	if ausFehler != nil {
		t.Fatalf("Wirksame: %v", ausFehler)
	}

	if len(aus) != 0 {
		t.Errorf("bei ausgeschaltetem Schalter sind %+v wirksam", aus)
	}
}

// **Und ein Fehler kommt auch dort durch.**
func TestEinFehlerKommtDurchDieWirksamenRegeln(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Set")

	ablage, spiele := probeersetzung()
	spiele.fehler = eigener

	if _, err := service.NeueAutoersetzungspflege(ablage, spiele).
		Wirksame(t.Context(), 1); !errors.Is(err, eigener) {
		t.Errorf("der Fehler lautet %v", err)
	}
}
