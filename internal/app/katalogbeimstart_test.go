package app

import "testing"

// **Die Stilauswahl steht beim ersten Start**
// (#692).
//
// # Was vorher geschah
//
// `Stilwaehler.Stile` liest den Paketkatalog aus dem Bestand. Gefüllt wurde
// der nur von `App.Pakete` — der **Seite** *Pakete*. Auf einem frischen
// Rechner war die Stilauswahl damit leer, bis jemand zufällig dort gewesen
// war; die Stile lagen die ganze Zeit auf der Platte.
//
// Gefunden hat es der Abnahmeweg „ein Set umgestalten"
// (#680),
// nicht das Auge.
func TestDieStilauswahlStehtOhneDenBlickAufDiePaketseite(t *testing.T) {
	t.Parallel()

	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	// **Ohne `Pakete()` dazwischen** — das ist der ganze Fall.
	stile := anwendung.Stile(t.Context(), geoeffnet.SetID)
	if stile.Fehler != nil {
		t.Fatalf("Stile: %+v", stile.Fehler)
	}

	if len(stile.Stile) == 0 {
		t.Fatal("die Stilauswahl ist leer, obwohl die Saat Stile mitbringt")
	}

	// **Und der gewählte ist als gewählt gekennzeichnet** — sonst stünde die
	// Liste da, ohne zu sagen, wo man ist.
	gewaehlte := 0
	for _, stil := range stile.Stile {
		if stil.Gewaehlt {
			gewaehlte++
		}
	}

	if gewaehlte != 1 {
		t.Errorf("%d Stile sind als gewählt gekennzeichnet, erwartet einer", gewaehlte)
	}
}

// **Und ein Rechner ohne jede Vorlage bekommt eine leere Auswahl, keinen
// Fehler.**
//
// Der Unterschied ist die Zusage: „es gibt keine Stile" ist eine Antwort,
// „etwas ist schiefgegangen" wäre eine andere.
func TestOhneVorlagenIstDieStilauswahlLeerUndOhneFehler(t *testing.T) {
	t.Parallel()

	anwendung := ohneSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	stile := anwendung.Stile(t.Context(), geoeffnet.SetID)
	if stile.Fehler != nil {
		t.Fatalf("Stile: %+v", stile.Fehler)
	}

	if len(stile.Stile) != 0 {
		t.Errorf("ohne Vorlagen stehen %d Stile da", len(stile.Stile))
	}
}
