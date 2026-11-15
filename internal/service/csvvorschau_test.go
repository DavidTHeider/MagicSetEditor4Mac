package service_test

// **Das CSV zeigt vorher, was es zusammenführt** — DOM-03.AK-01
// (#1698).
//
// Der Anlass war dieser Weg: die **eigene Ausfuhr ein zweites Mal
// eingelesen** verdoppelte jede Menge, still. Die Vorschau sagt es vorher,
// und *Vorhandene auslassen* liest dieselbe Datei, ohne etwas zu verdoppeln.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieCSVVorschauSagtWasSieZusammenfuehrt(t *testing.T) {
	t.Parallel()

	ablage := vorhandenerBestand(t)

	vorschau, fehler := service.NeueSammlungsausfuhr(ablage).Vorschauen(t.Context(),
		"printing,menge,zustand,sprache\n"+
			probeprinting+",3,nearMint,en\n"+
			zweitprinting+",1,nearMint,en\n"+
			"kaputt,keineZahl,nearMint,en\n")
	if fehler != nil {
		t.Fatalf("vorschauen: %v", fehler)
	}

	if len(vorschau.Posten) != 2 {
		t.Fatalf("%d Posten, erwartet 2: %+v", len(vorschau.Posten), vorschau)
	}

	if erster := vorschau.Posten[0]; erster.Zeile != 2 || erster.Art != service.EinfuhrZusammen ||
		erster.Vorher != 2 || erster.Nachher != 5 {
		t.Errorf("erster Posten %+v", erster)
	}

	if zweiter := vorschau.Posten[1]; zweiter.Zeile != 3 || zweiter.Art != service.EinfuhrNeu ||
		zweiter.Nachher != 1 {
		t.Errorf("zweiter Posten %+v", zweiter)
	}

	if len(vorschau.Befunde) != 1 || vorschau.Befunde[0].Ort != "Zeile 4" {
		t.Errorf("Befunde %+v", vorschau.Befunde)
	}

	if zeilen := bestandsliste(t, ablage); len(zeilen) != 1 || zeilen[0].Menge != 2 {
		t.Errorf("im Bestand steht %+v", zeilen)
	}
}

// **Die Vorschau legt auch keinen Druck an** — eine Referenz, die der
// Bestand nicht kennt, wäre beim Einlesen ein neuer vorläufiger Druck; hier
// ist sie ein neuer Posten, und das Register bleibt leer.
func TestDieCSVVorschauLegtKeinenDruckAn(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	zeile := ",2,Lightning Bolt,m11,149,en,1,mtgjson,aaaaaaaa-0000-4000-8000-000000000001\n"

	vorschau, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).MitDrucken(register).
		Vorschauen(t.Context(), csvkopf+zeile+zeile)
	if fehler != nil {
		t.Fatalf("vorschauen: %v", fehler)
	}

	if len(register.drucke) != 0 {
		t.Errorf("die Vorschau hat %d Drucke angelegt", len(register.drucke))
	}

	// **Zweimal dieselbe unbekannte Referenz sind zwei neue Posten** — ob es
	// derselbe Druck wird, entscheidet erst das Anlegen; die Vorschau
	// behauptet es nicht.
	if len(vorschau.Posten) != 2 || vorschau.Posten[0].Art != service.EinfuhrNeu ||
		vorschau.Posten[1].Art != service.EinfuhrNeu {
		t.Errorf("Vorschau %+v", vorschau)
	}
}

// **Ein klemmendes Register hält die Vorschau an.**
func TestEinKlemmendesRegisterHaeltDieCSVVorschauAn(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{fehlerBei: "Kandidaten"}
	zeile := ",2,Lightning Bolt,m11,149,en,1,mtgjson,aaaaaaaa-0000-4000-8000-000000000001\n"

	if _, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).MitDrucken(register).
		Vorschauen(t.Context(), csvkopf+zeile); !errors.Is(fehler, errRegister) {
		t.Errorf("Fehler = %v, erwartet errRegister", fehler)
	}
}

// **Die eigene Ausfuhr ein zweites Mal** — beim Auslassen bleibt jede Menge,
// wie sie war.
func TestDieEigeneAusfuhrZweimalVerdoppeltBeimAuslassenNichts(t *testing.T) {
	t.Parallel()

	ablage := vorhandenerBestand(t)
	ausfuhr := service.NeueSammlungsausfuhr(ablage)

	text, fehler := ausfuhr.Ausgeben(t.Context(), service.FormCSV)
	if fehler != nil {
		t.Fatalf("ausgeben: %v", fehler)
	}

	bericht, fehler := ausfuhr.EinfuegenMit(t.Context(), text, "", service.ZusammenAuslassen)
	if fehler != nil {
		t.Fatalf("einlesen: %v", fehler)
	}

	if bericht.Ausgelassen != 1 || len(bericht.Angelegt) != 0 || bericht.Erhoeht != 0 {
		t.Errorf("Bericht %+v", bericht)
	}

	if zeilen := bestandsliste(t, ablage); len(zeilen) != 1 || zeilen[0].Menge != 2 {
		t.Errorf("im Bestand steht %+v", zeilen)
	}

	// **Und ein unbekannter Modus schreibt nichts.**
	if _, fehler := ausfuhr.EinfuegenMit(t.Context(), text, "", "verdoppeln"); fehler == nil {
		t.Error("ein unbekannter Modus ging durch")
	}
}

// **Was das Einlesen abwiese, ist ein Befund der Vorschau** — ein Zustand,
// den es nicht gibt, und eine Referenz, die mehrere Drucke trifft; beide mit
// ihrer Zeile, und die übrigen Zeilen stehen trotzdem da.
func TestDieCSVVorschauNenntAbgewieseneZeilen(t *testing.T) {
	t.Parallel()

	vorschau, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).Vorschauen(t.Context(),
		"printing,menge,zustand\n"+probeprinting+",1,erfunden\n"+zweitprinting+",1,nearMint\n")
	if fehler != nil {
		t.Fatalf("vorschauen: %v", fehler)
	}

	if len(vorschau.Posten) != 1 || vorschau.Posten[0].Zeile != 3 ||
		len(vorschau.Befunde) != 1 || vorschau.Befunde[0].Ort != "Zeile 2" {
		t.Errorf("Vorschau %+v", vorschau)
	}
}

// **Eine mehrdeutige Referenz ist ein Befund und kein Posten** — die
// Vorschau entscheidet so wenig wie das Einlesen, welcher Druck gemeint ist.
func TestDieCSVVorschauNenntEineMehrdeutigeReferenz(t *testing.T) {
	t.Parallel()

	register := &druckregisterdoppel{}
	scryfallDruck(t, register, "11111111-1111-4111-8111-111111111111", "Lightning Bolt", "m11", "149")
	scryfallDruck(t, register, "22222222-2222-4222-8222-222222222222", "Lightning Bolt", "m11", "149")

	vorschau, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).MitDrucken(register).
		Vorschauen(t.Context(), csvkopf+",1,Lightning Bolt,m11,149,en,,,\n")
	if fehler != nil || len(vorschau.Posten) != 0 || len(vorschau.Befunde) != 1 ||
		vorschau.Befunde[0].Ort != "Zeile 2" {
		t.Errorf("Vorschau %+v (%v)", vorschau, fehler)
	}
}

// **Eine schweigende Sammlung hält auch die CSV-Vorschau an.**
func TestEineSchweigendeSammlungHaeltDieCSVVorschauAn(t *testing.T) {
	t.Parallel()

	if _, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{fehler: errKaputt}).Vorschauen(
		t.Context(), "printing,menge\n"+probeprinting+",1\n"); !errors.Is(fehler, errKaputt) {
		t.Errorf("Fehler = %v, erwartet errKaputt", fehler)
	}
}
