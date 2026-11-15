package service_test

// **Die Karten eines Kartenskripts rechnen ihre Werte** — #2516. MSE rechnet
// jede Karte einer `AddCardAction`, auch die eines Kartenskripts
// (`SetScriptManager::onAction`, `src/script/script_manager.cpp`, Zeilen
// 168–178, Fassung f9d9356 vom 2021-07-06): was das Skript gibt, ist der
// Eingang des Wertskripts.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieKartenEinesSkriptsRechnenIhreWerte(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}
	spiel := &spielprobe{
		skripte: []service.Kartenskriptangabe{{Name: "Glut", Skript: `[ new_card([eins: "Glut"]) ]`}},
		felder:  nachzugfelder(),
	}
	stile := &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true}

	if _, fehler := service.NeuerKartenskriptdienst(anleger, spiel).MitRechnung(stile, nil, nil).
		Ausfuehren(t.Context(), 1, "Glut"); fehler != nil {
		t.Fatalf("Ausfuehren: %v", fehler)
	}

	if len(anleger.karten) != 1 {
		t.Fatalf("%d Karten angelegt", len(anleger.karten))
	}

	gelesen := map[string]string{}
	for _, feld := range anleger.karten[0] {
		gelesen[feld.Feld] = feld.Wert
	}

	if gelesen["eins"] != "[Glut]" || gelesen["leser"] != "[Glut]!" || gelesen["fremd"] != "<>" {
		t.Errorf("angelegt %v — erwartet eins=[Glut], leser=[Glut]!, fremd=<>", gelesen)
	}
}
