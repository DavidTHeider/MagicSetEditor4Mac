//go:build darwin

// **Ein Leser lässt kein Modell arbeiten** — ID-03.AK-04
// (#1702).
//
// # Der Abgleich
//
// **Die Matrix führt jede dieser Fähigkeiten unter `lokalSchreiben`,
// Mindestrolle Mitarbeiter** — mit Grund: ein Modellruf ändert nichts an der
// Karte, kostet aber Zeit und Strom (A7, #1446) und bei einem Anbieter Geld.
// Gehalten hat das bis ID-03 nur die nur lesende Verbindung, und die hält einen
// Schritt nicht auf, der nicht schreibt. MCP-Server und Dienst fragen seit
// ID-03 vor jedem schreibenden Ruf den Akteur; ohne denselben Riegel hier gäbe
// dieselbe Fähigkeit am Fenster und an der Kommandozeile eine andere Antwort
// (§16).
//
// **Gefragt wird vor dem Modell**, nicht davor die Laufzeit: in diesem Lauf
// gibt es keine, und eine Absage aus einem anderen Grund sähe genauso aus.
// Deshalb steht der Code in der Zusicherung.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEinLeserLaesstKeinModellArbeiten(t *testing.T) {
	anwendung, geoeffnet := alsLeserVorEinemSet(t)
	ctx := t.Context()

	zeilen := anwendung.Kartenzeilen(ctx, geoeffnet.SetID)
	if zeilen.Fehler != nil || len(zeilen.Zeilen) == 0 {
		t.Fatalf("der Leser sieht die Karten nicht: %+v", zeilen)
	}

	karte := zahlAus(t, zeilen.Zeilen[0].ID)

	faelle := map[string]*app.Fehler{
		"vorschlagen": anwendung.KIVorschlagen(ctx, karte, "ollama", "", "", "").Fehler,
		"vorschlagen mit Belegen": anwendung.KIVorschlagenMit(ctx,
			app.KIVorschlagsbezug{KarteID: karte}).Fehler,
		"in einem Zug": anwendung.KIVorschlagInEinemZug(ctx,
			app.KIVorschlagsbezug{KarteID: karte}, false, nil).Fehler,
		"mehrere Karten":    anwendung.KIBatchVorschlagen(ctx, []int64{karte}, "ollama", "", "", "").Fehler,
		"ein Bild":          anwendung.KIBildVorschlagen(ctx, karte, "bild", "ollama", "", "", "", "", false, "").Fehler,
		"einschätzen":       anwendung.KIEinschaetzen(ctx, app.KIEinschaetzungsbezug{KarteID: karte}).Fehler,
		"erkennen":          anwendung.KIKarteErkennen(ctx, app.KIErkennungsbezug{}).Fehler,
		"ein Modell messen": anwendung.KIModellPruefen(ctx, "ollama", "", "").Fehler,
	}

	for name, fehler := range faelle {
		if fehler == nil {
			t.Errorf("%s: ein Leser ließ ein Modell arbeiten", name)

			continue
		}

		if fehler.Code != app.CodeNurLesend {
			t.Errorf("%s: Code %q, erwartet %q — %s", name, fehler.Code, app.CodeNurLesend, fehler.Meldung)
		}
	}
}
