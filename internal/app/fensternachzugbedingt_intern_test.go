package app

// **Der Entwurf prüft sein Feld, nicht das, was nachgezogen wird** —
// Gesamtreview über 0.9.0, Befund 1
// (#2382).
//
// Das Fenster schickt die Werte, die es von **dem** Feld kennt, das es
// bearbeitet (U02b, #2179). Seit #2396 schreibt derselbe Aufruf auch die
// Felder, die das Wertskript nachzieht — und jedes wurde gegen dieselbe
// Liste geprüft. `echo` hatte das Fenster nie gesehen; sein Wert stand nicht
// darin, und die Bearbeitung von `zeile` scheiterte als fremde Änderung. In
// einem M15-Set ist das jede Änderung der Manakosten, die die Farbe nachzieht.
//
// Die Fenstertests aus #2396 riefen `App.FeldSetzen` und gingen an dieser Naht
// vorbei.

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDasFensterSchreibtBedingtUndZiehtNach(t *testing.T) {
	anwendung, setID, karteID, werte := nachzugsfenster(t)
	karte := strconv.FormatInt(karteID, 10)

	if gesetzt := anwendung.FeldSetzen(probekontext(t), setID, karte, "zeile", "Funke"); gesetzt.Fehler != nil {
		t.Fatalf("FeldSetzen: %+v", gesetzt.Fehler)
	}

	// **Was das Fenster kennt**: den Wert im Feld und den getippten.
	if gesetzt := anwendung.FeldSetzenBedingt(probekontext(t), setID, karte, "zeile", "Glut",
		[]string{"[Funke]", "Glut"}); gesetzt.Fehler != nil {
		t.Fatalf("FeldSetzenBedingt: %+v — der Nachzug von echo galt als fremder Wert", gesetzt.Fehler)
	}

	if gelesen := werte(); gelesen["zeile"] != "[Glut]" || gelesen["echo"] != "[Glut]!" {
		t.Errorf("im Bestand steht zeile=%q, echo=%q — erwartet [Glut] und [Glut]!", gelesen["zeile"], gelesen["echo"])
	}
}

// **Und ein fremder Wert im bearbeiteten Feld bleibt ein Konflikt** — sonst
// wäre die Korrektur nur ein Abschalten der Bedingung.
func TestEinFremderWertBleibtMitNachzugEinKonflikt(t *testing.T) {
	anwendung, setID, karteID, werte := nachzugsfenster(t)
	karte := strconv.FormatInt(karteID, 10)

	for _, wert := range []string{"Funke", "Asche"} {
		if gesetzt := anwendung.FeldSetzen(probekontext(t), setID, karte, "zeile", wert); gesetzt.Fehler != nil {
			t.Fatalf("FeldSetzen %q: %+v", wert, gesetzt.Fehler)
		}
	}

	gesetzt := anwendung.FeldSetzenBedingt(probekontext(t), setID, karte, "zeile", "Glut",
		[]string{"[Funke]", "Glut"})
	if gesetzt.Fehler == nil || gesetzt.Fehler.Schluessel != domain.SchluesselFremderWert {
		t.Fatalf("Entwurf über den fremden Wert: %+v, erwartet %q", gesetzt.Fehler, domain.SchluesselFremderWert)
	}

	if gelesen := werte(); gelesen["zeile"] != "[Asche]" || gelesen["echo"] != "[Asche]!" {
		t.Errorf("im Bestand steht zeile=%q, echo=%q — erwartet [Asche] und [Asche]! unverändert", gelesen["zeile"], gelesen["echo"])
	}
}
