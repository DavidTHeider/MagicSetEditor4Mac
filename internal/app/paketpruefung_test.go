// **Die Bindung, die eine Paketdatei prüft** — Werkstück B von Phase 4.9
// (#1324).
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestPaketdateiPruefenNenntStelleUndSchluessel(t *testing.T) {
	antwort := (&app.App{}).PaketdateiPruefen("mse version: 0.3.8\nscript: f(1, 2\n")

	if antwort.Stelle == nil {
		t.Fatalf("keine Stelle: %+v", antwort)
	}

	if antwort.Stelle.Zeile != 2 || antwort.Stelle.Spalte != 10 {
		t.Errorf("Stelle %+v, erwartet Zeile 2, Spalte 10", *antwort.Stelle)
	}

	// **Der Schlüssel geht mit**, damit das Fenster übersetzen kann (§9).
	if antwort.Schluessel != "klammerOffen" {
		t.Errorf("Schlüssel %q, erwartet %q", antwort.Schluessel, "klammerOffen")
	}

	// **Und die Werte füllen die Platzhalter** — ohne sie stünde im Fenster
	// „diese {0} wird nicht geschlossen".
	if len(antwort.Werte) != 2 {
		t.Errorf("Werte %v, erwartet zwei", antwort.Werte)
	}
}

// **Eine heile Datei gibt eine leere Liste und keine `null`.**
//
// Das ist die Falle aus §3: eine nicht gesetzte Liste ist in Go `nil`, und
// über die Wails-Naht kommt daraus `null` — im Fenster steht dann `undefined`
// statt einer leeren Liste.
func TestPaketdateiPruefenGibtOhneBefundEineLeereListe(t *testing.T) {
	antwort := (&app.App{}).PaketdateiPruefen("mse version: 0.3.8\nscript: 1 + 2\n")

	if antwort.Stelle != nil {
		t.Errorf("Stelle %+v, erwartet keine", *antwort.Stelle)
	}

	if antwort.Werte == nil {
		t.Error("Werte sind nil — über die Naht wird daraus null")
	}
}

// **Und eine Meldung ohne Platzhalter trägt trotzdem eine Liste.**
//
// `zuweisungBrauchtNamen` hat keine Werte; ohne den Rückfall käme auch hier
// `null` an. Der Zweig steht deshalb im Code, und dieser Fall betritt ihn.
func TestEineMeldungOhnePlatzhalterTraegtEineLeereListe(t *testing.T) {
	antwort := (&app.App{}).PaketdateiPruefen("mse version: 0.3.8\nscript: 1 := 2\n")

	if antwort.Stelle == nil {
		t.Fatalf("keine Stelle: %+v", antwort)
	}

	if antwort.Werte == nil {
		t.Error("Werte sind nil — über die Naht wird daraus null")
	}
}
