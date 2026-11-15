package service_test

// **Ein Preis, der nicht zu lesen ist, wird genannt** — LOC-03.AK-03
// (#1689).
//
// # Wogegen hier geprüft wird
//
//   - **Die Zeile geht durch** — der alte Grund bleibt: eine Sammlung
//     scheitert nicht an einer Spalte, die niemand braucht.
//   - **Und der Verlust wird gesagt** — bis zum 2026-09-22 wurde aus »1,23«
//     eine Null, ohne ein Wort.
//   - **Der Rundlauf hält die Währung** — 150 Yen kommen als »150« zurück
//     und nicht als »1.50«.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einCSV baut eine Zeile mit dem Kopf, den die Ausfuhr schreibt.
//
// **Abgeschrieben und nicht erfunden**: der erste Anlauf ordnete die Spalten
// nach Gefühl, und jede Zeile fiel an einem Zustand »NM« aus, den es nicht
// gibt. Eine Vorrichtung, die eine andere Gestalt hat als der Betrieb, prüft
// sich selbst (§3).
func einCSV(preis, waehrung string) string {
	return "printing,menge,zustand,sprache,finish,foil,name,setcode," +
		"nummer,erwerbsdatum,erwerbspreis,erwerbswaehrung,notiz,lagerort," +
		"wunsch,wunschrang,wunschart,verweisfassung,kennungssystem,kennung\n" +
		"11111111-1111-4111-8111-111111111111,1,nearMint,de,nonfoil,," +
		"Wald,zen,246,2026-01-01," + zelle(preis) + "," + waehrung +
		",,,0,,,,,\n"
}

// zelle setzt einen Wert in Anführungszeichen, wenn er ein Komma trägt.
//
// **Weil das Komma hier zwei Bedeutungen hat**: es trennt die Spalten, und
// in »1,23« trennt es die Nachkommastellen. Ein Programm, das aus Excel
// kommt, schreibt deshalb `"1,23"` — und genau diese Datei ist der Fall,
// den LOC-03.AK-03 meint.
func zelle(wert string) string {
	if !strings.Contains(wert, ",") {
		return wert
	}

	return `"` + wert + `"`
}

// **Ein deutsch formatierter Preis wird genannt — und die Karte kommt an.**
func TestEinUnlesbarerPreisWirdGenanntUndDieKarteKommtAn(t *testing.T) {
	t.Parallel()

	ausfuhr := service.NeueSammlungsausfuhr(&sammlungsdoppel{})

	bericht, fehler := ausfuhr.Einfuegen(
		t.Context(), einCSV("1,23", "EUR"), "")
	if fehler != nil {
		t.Fatalf("einfügen: %v", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Fatalf("angelegt wurden %d Zeilen", len(bericht.Angelegt))
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("es kamen %d Befunde: %+v", len(bericht.Befunde),
			bericht.Befunde)
	}

	befund := bericht.Befunde[0]

	switch {
	case !strings.Contains(befund.Meldung, "1,23"):
		t.Errorf("die Meldung nennt den Wert nicht: %q", befund.Meldung)
	case !strings.Contains(befund.Meldung, "Komma"):
		t.Errorf("die Meldung sagt nicht, was falsch ist: %q", befund.Meldung)
	case befund.Ort != "Zeile 2":
		t.Errorf("der Befund sitzt in %q", befund.Ort)
	case befund.Art != domain.ArtEingabe:
		t.Errorf("der Befund ist von der Art %q", befund.Art)
	}
}

// **Ein lesbarer Preis kommt ohne Befund durch.**
func TestEinLesbarerPreisKommtOhneBefundDurch(t *testing.T) {
	t.Parallel()

	ausfuhr := service.NeueSammlungsausfuhr(&sammlungsdoppel{})

	bericht, fehler := ausfuhr.Einfuegen(
		t.Context(), einCSV("1.23", "EUR"), "")
	if fehler != nil {
		t.Fatalf("einfügen: %v", fehler)
	}

	if len(bericht.Befunde) != 0 {
		t.Errorf("es kamen Befunde: %+v", bericht.Befunde)
	}
}

// **Und der Yen läuft rund** — 150 hinein, »150« heraus.
//
// **Das ist die Zusage aus AK-02**: die Stellen kommen aus der Währung.
// Vorher stand im CSV »1.50«, und beim nächsten Einlesen waren aus 150 Yen
// 150 Hundertstel geworden.
func TestDerYenLaeuftRund(t *testing.T) {
	t.Parallel()

	ausfuhr := service.NeueSammlungsausfuhr(&sammlungsdoppel{})
	ctx := t.Context()

	if _, fehler := ausfuhr.Einfuegen(
		ctx, einCSV("150", "JPY"), ""); fehler != nil {
		t.Fatalf("einfügen: %v", fehler)
	}

	text, fehler := ausfuhr.Ausgeben(ctx, service.FormCSV)
	if fehler != nil {
		t.Fatalf("ausgeben: %v", fehler)
	}

	if !strings.Contains(text, ",150,JPY,") {
		t.Errorf("im CSV steht nicht »150«:\n%s", text)
	}
}

// **Ein Preis, den die Sammlung abwiese, nimmt nur sich selbst mit** — Befund
// G16 des Zweitblicks über Phase 6.3
// (#1666).
//
// # Der Befund
//
// **Lesbar ist nicht gültig.** »-1.00« liest [domain.GeldAus] ohne Klage,
// und erst die Prüfung des Exemplars weist ihn ab — für die **ganze** Zeile:
// die Karte kam nicht an. Vermerk 0127 sagt das Gegenteil: „Beim Einlesen
// einer Sammlung bleibt die Zeile trotzdem", und „Preis **und Währung** der
// Zeile bleiben dann leer, denn eine Währung ohne Betrag sagt nichts".
//
// **Vier Gestalten, eine Regel**: was `Exemplar.Preisbefunde` abweist, fällt
// als Paar weg und wird genannt — und nichts sonst.
func TestEinUngueltigerPreisNimmtNurSichSelbstMit(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, preis, waehrung, nennt string
	}{
		{"negativ", "-1.00", "EUR", "nicht negativ"},
		{"ohne Währung", "1.00", "", "gehört seine Währung"},
		{"kein Währungscode", "1.00", "EURO", "»EURO«"},
		{"Währung ohne Preis", "", "EUR", "gehört ihr Preis"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := &sammlungsdoppel{}

			bericht, fehler := service.NeueSammlungsausfuhr(ablage).Einfuegen(
				t.Context(), einCSV(fall.preis, fall.waehrung), "")
			if fehler != nil {
				t.Fatalf("einfügen: %v", fehler)
			}

			if len(bericht.Angelegt) != 1 || len(ablage.zeilen) != 1 {
				t.Fatalf("angelegt %d, im Bestand %d — die Karte kam nicht an: %+v",
					len(bericht.Angelegt), len(ablage.zeilen), bericht.Befunde)
			}

			if zeile := ablage.zeilen[0]; zeile.Erwerbspreis != 0 || zeile.Erwerbswaehrung != "" {
				t.Errorf("im Bestand stehen %d %q — erwartet war beides leer",
					zeile.Erwerbspreis, zeile.Erwerbswaehrung)
			}

			if len(bericht.Befunde) != 1 {
				t.Fatalf("es kamen %d Befunde: %+v", len(bericht.Befunde), bericht.Befunde)
			}

			befund := bericht.Befunde[0]
			if !strings.Contains(befund.Meldung, fall.nennt) ||
				!strings.Contains(befund.Meldung, "die Karte kommt an") ||
				befund.Ort != "Zeile 2" || befund.Art != domain.ArtEingabe {
				t.Errorf("der Befund lautet %+v", befund)
			}
		})
	}
}

// **Und der Tag des Erwerbs ebenso** — dieselbe Klasse wie der Preis
// darüber. Die Prüfung des Tages kam mit UX-06
// (#1678)
// für die Formulare, und am Einlesen nahm sie seither die ganze Zeile mit:
// ein »gestern« aus einer fremden Liste, und die Karte fehlte. **Preis,
// Währung und Tag sind die Angaben zum Erwerb** (`domain/exemplarerwerb.go`)
// — sie beschreiben den Kauf, nicht die Karte.
func TestEinUngueltigerErwerbstagNimmtNurSichSelbstMit(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	text := strings.Replace(einCSV("1.00", "EUR"), "2026-01-01", "gestern", 1)

	bericht, fehler := service.NeueSammlungsausfuhr(ablage).Einfuegen(t.Context(), text, "")
	if fehler != nil {
		t.Fatalf("einfügen: %v", fehler)
	}

	if len(ablage.zeilen) != 1 {
		t.Fatalf("im Bestand %d — die Karte kam nicht an: %+v", len(ablage.zeilen), bericht.Befunde)
	}

	if zeile := ablage.zeilen[0]; zeile.Erwerbsdatum != "" || zeile.Erwerbspreis != 100 {
		t.Errorf("im Bestand stehen %q und %d — erwartet waren kein Tag und der Preis",
			zeile.Erwerbsdatum, zeile.Erwerbspreis)
	}

	if len(bericht.Befunde) != 1 || !strings.Contains(bericht.Befunde[0].Meldung, "»gestern«") {
		t.Errorf("die Befunde lauten %+v", bericht.Befunde)
	}
}
