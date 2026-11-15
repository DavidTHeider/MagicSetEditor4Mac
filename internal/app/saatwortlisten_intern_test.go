// Die **Auswahllisten der Saat** — die Naht vom Fenster bis in den Bestand
// (#1074).
//
// # Warum an der Saat und nicht an einer Vorrichtung
//
// Weil die Wortlisten **im Spielpaket** stehen. Ein Test mit einem
// zusammengesetzten `Vorlage`-Verbund prüft den Rechenweg; er prüft nicht,
// dass die Listen aus einem geöffneten Paket bis an die Bindung kommen — und
// genau diese Naht hat bei den Schlüsselwörtern gefehlt (#1064).
package app

import (
	"strings"
	"testing"
)

// TestDieSaatBietetIhreAuswahlAn geht den ganzen Weg: Set öffnen, einen Wert
// mit Marken schreiben, das Menü holen, wählen, nachsehen.
func TestDieSaatBietetIhreAuswahlAn(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	karteID := karten.Karten[0].ID

	// **Die Marken kommen im echten Betrieb aus einem Feldfilter** des
	// Spielpakets; hier werden sie geschrieben, denn geprüft wird die
	// Auswahl und nicht der Filter.
	roh := "<word-list-probe-art>Große Kreatur</word-list-probe-art>"
	if antwort := anwendung.FeldSetzen(
		probekontext(t), geoeffnet.SetID, karteID, "text", roh); antwort.Fehler != nil {
		t.Fatalf("Feld setzen: %+v", antwort.Fehler)
	}

	auswahl := anwendung.Wortlisten(t.Context(), karteID, "text")
	if auswahl.Fehler != nil {
		t.Fatalf("Wortlisten: %+v", auswahl.Fehler)
	}

	if len(auswahl.Bereiche) != 1 {
		t.Fatalf("%d Bereiche: %+v", len(auswahl.Bereiche), auswahl.Bereiche)
	}

	bereich := auswahl.Bereiche[0]
	if bereich.Liste != "probe-art" || !bereich.Kaestchen {
		t.Errorf("der Bereich ist %+v", bereich)
	}

	// **Das Präfix ist angekreuzt und der Typ gewählt** — beides steht im
	// Wert, und beides rechnet `service` aus dem Text.
	if !bereich.Eintraege[0].Aktiv || bereich.Gewaehlt != 1 {
		t.Errorf("der Zustand ist %+v", bereich.Eintraege)
	}

	// **Und die Gruppe bringt ihr Untermenü mit** — im Pack sind es 103.
	if len(bereich.Eintraege[3].Kinder) != 2 {
		t.Errorf("das Untermenü ist %+v", bereich.Eintraege[3])
	}

	// Ein anderer Typ, und das Präfix bleibt stehen.
	if antwort := anwendung.WortWaehlen(
		probekontext(t), geoeffnet.SetID, karteID, "text", 0, []int{2},
	); antwort.Fehler != nil {
		t.Fatalf("wählen: %+v", antwort.Fehler)
	}

	nachher := feldwertVon(t, anwendung, geoeffnet.SetID, karteID, "text")

	erwartet := "<word-list-probe-art>Große Spruch</word-list-probe-art>"
	if nachher != erwartet {
		t.Errorf("im Feld steht\n  %q\nerwartet\n  %q", nachher, erwartet)
	}

	// **Und eine Wahl im Untermenü ersetzt den Bereich ganz** — so tut MSE
	// es: `select` sieht nur die Einträge seines eigenen Menüs
	// (`src/gui/value/text.cpp`, Zeilen 276–297, Fassung f9d9356 vom
	// 2021-07-06). Das Präfix ist danach weg.
	if antwort := anwendung.WortWaehlen(
		probekontext(t), geoeffnet.SetID, karteID, "text", 0, []int{3, 0},
	); antwort.Fehler != nil {
		t.Fatalf("wählen: %+v", antwort.Fehler)
	}

	tiefer := feldwertVon(t, anwendung, geoeffnet.SetID, karteID, "text")
	if !strings.Contains(tiefer, ">Wunder<") {
		t.Errorf("im Feld steht %q", tiefer)
	}
}

// feldwertVon liest einen Feldwert aus dem Bestand.
func feldwertVon(t *testing.T, anwendung *App, setID, karteID, feld string) string {
	t.Helper()

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	for _, karte := range karten.Karten {
		if karte.ID != karteID {
			continue
		}

		for _, wert := range karte.Werte {
			if wert.Feld == feld {
				return wert.Wert
			}
		}
	}

	t.Fatalf("%s/%s hat kein Feld %s", setID, karteID, feld)

	return ""
}
