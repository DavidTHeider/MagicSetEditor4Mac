package app_test

// **Was eine Zeile sagt, die nicht aufgeht — übersetzbar** — G01
// (#2248).
//
// Am nativen Fenster stand der rohe Fehler einer Zeile da, deutsch in jeder
// Sprache. Jetzt trägt die Zeile ihren Ausfall als [app.Fehler] mit Schlüssel,
// und dass es mehr Drucke gibt als eine Seite, als eigenes Feld.

import "testing"

func TestEineZeileOhneDruckIstKeinAusfall(t *testing.T) {
	deutung, err := einfuhrnaht(t).Deuten(probekontext(t), "1 Gibtesnicht")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	if zeile := deutung.Zeilen[0]; zeile.Fehler != nil || zeile.Drucke != 0 || zeile.Befund != "kein Druck heißt so" {
		t.Errorf("Zeile: Fehler %+v, Drucke %d, Befund %q", zeile.Fehler, zeile.Drucke, zeile.Befund)
	}
}

func TestEinAusfallKommtAlsFehlerMitSchluessel(t *testing.T) {
	deutung, err := einfuhrnaht(t).Deuten(probekontext(t), "1 Ausfall")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	if zeile := deutung.Zeilen[0]; zeile.Fehler == nil || zeile.Fehler.Schluessel == "" {
		t.Errorf("Zeile: Fehler %+v, Befund %q", zeile.Fehler, zeile.Befund)
	}
}

func TestMehrAlsEineSeiteStehtAlsFeldDa(t *testing.T) {
	deutung, err := einfuhrnaht(t).Deuten(probekontext(t), "1 Viele\n1 Erste Probe")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	if !deutung.Zeilen[0].Weitere || deutung.Zeilen[1].Weitere {
		t.Errorf("Weitere: %v, %v", deutung.Zeilen[0].Weitere, deutung.Zeilen[1].Weitere)
	}
}
