package msepaket_test

import "testing"

func TestSymbolmodellLiestZeichenUndBilder(t *testing.T) {
	schrift, _ := oeffneTestpaket(t, "probe-symbole.mse-symbol-font").Symbolmodell()

	if schrift.Name != "Probesymbole" {
		t.Errorf("Name %q", schrift.Name)
	}
	if len(schrift.Symbole) != 2 {
		t.Fatalf("%d Symbole, erwartet waren 2", len(schrift.Symbole))
	}

	weiss, vorhanden := schrift.Symbol("W")
	if !vorhanden || weiss.Bild != "weiss.png" {
		t.Errorf("W ist %+v", weiss)
	}
	if _, vorhanden := schrift.Symbol("gibtsnicht"); vorhanden {
		t.Error("ein erfundenes Zeichen gilt als vorhanden")
	}
}

// Ein kaputtes Symbol ist ein Befund, kein Abbruch.
//
// Eine Schrift mit neunundneunzig brauchbaren Zeichen und einem kaputten ist
// neunmal nützlicher als gar keine.
func TestEinKaputtesSymbolIstEinBefund(t *testing.T) {
	schrift, befunde := oeffneTestpaket(t, "probe-symbole.mse-symbol-font").Symbolmodell()

	if len(befunde) != 2 {
		t.Fatalf("%d Befunde, erwartet waren 2", len(befunde))
	}
	for _, befund := range befunde {
		if befund.Stelle.Zeile == 0 || befund.Meldung == "" {
			t.Errorf("ein Befund ohne Stelle oder Meldung: %+v", befund)
		}
	}

	// Und die brauchbaren Zeichen sind trotzdem da.
	if len(schrift.Symbole) != 2 {
		t.Errorf("%d brauchbare Symbole", len(schrift.Symbole))
	}
}

// Bilder werden nicht beim Öffnen gelesen.
//
// Eine Symbolschrift kann hundert Bilder tragen, und beim Öffnen eines Sets
// braucht man keines davon. Der Pfad reicht; gelesen wird beim Rendern.
func TestSymbolmodellLiestKeineBilder(t *testing.T) {
	schrift, _ := oeffneTestpaket(t, "probe-symbole.mse-symbol-font").Symbolmodell()

	for _, symbol := range schrift.Symbole {
		if symbol.Bild == "" {
			t.Errorf("das Symbol %q trägt keinen Bildpfad", symbol.Code)
		}
	}
}
