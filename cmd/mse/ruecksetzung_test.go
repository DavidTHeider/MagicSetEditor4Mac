package main

// **Die abgestufte Rücksetzung auf der Kommandozeile** — Werkstück O4 von
// Phase 5.2 (#1428).
//
// **Der Rückgabewert ist die Antwort**: 1, wenn etwas im Weg steht — ein
// Skript soll den Fall erkennen, ohne den Text zu lesen.

import (
	"strings"
	"testing"
)

func TestDieRuecksetzungsartenKommenInIhrerReihe(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "ruecksetzungsarten")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	zeilen := strings.Fields(aus)
	if len(zeilen) != 6 || zeilen[0] != "layout" || zeilen[5] != "anwendung" {
		t.Errorf("die Reihe ist %v", zeilen)
	}

	if mitArgument, _, _ := fahreAn(t, ort, "ruecksetzungsarten", "mehr"); mitArgument != CodeBedienung {
		t.Errorf("mit Argument: %d", mitArgument)
	}

	code, aus, _ = fahreAn(t, ort, "ruecksetzungsarten", "--json")
	if code != CodeGut || !strings.Contains(aus, `"anwendung"`) {
		t.Errorf("als JSON: %d %s", code, aus)
	}
}

func TestDieLayoutvorschauSagtWasSieAnfasst(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "ruecksetzung-vorschau", "layout")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, erwartet := range []string{
		"art\tlayout", "bereit\ttrue", "nachweisnoetig\tfalse",
		"sicherungempfohlen\tfalse", "betroffen\teinstellungen\t10",
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", erwartet, aus)
		}
	}

	code, aus, _ = fahreAn(t, ort, "ruecksetzung-vorschau", "layout", "--json")
	if code != CodeGut || !strings.Contains(aus, `"bereit": true`) {
		t.Errorf("als JSON: %d %s", code, aus)
	}
}

// **Ein Hindernis gibt 1** — und die Zeile nennt es.
func TestEineGesperrteVorschauGibtEins(t *testing.T) {
	ort, _, _ := mitKarte(t)

	// **Ein fremdes Profil gehört dem Rufenden nicht** — und kommt ohne
	// Zahlen zurück.
	code, aus, _ := fahreAn(t, ort, "ruecksetzung-vorschau", "profil", "4711")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d:\n%s", code, aus)
	}

	if !strings.Contains(aus, "hindernis\tfremdesZiel") {
		t.Errorf("die Ausgabe nennt das Hindernis nicht:\n%s", aus)
	}

	if strings.Contains(aus, "betroffen\t") {
		t.Errorf("ein fremdes Profil wurde gezählt:\n%s", aus)
	}

	// **Und als JSON derselbe Rückgabewert.**
	if alsJSON, _, _ := fahreAn(t, ort,
		"ruecksetzung-vorschau", "profil", "4711", "--json"); alsJSON != CodeFachlich {
		t.Errorf("als JSON: %d", alsJSON)
	}
}

// **Profil und Nutzer lassen sich benennen** — stellenweise, weil `mse` keine
// benannten Argumente kennt.
func TestProfilUndNutzerLassenSichBenennen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	// **Das gezeigte Profil ist 1** — es zu nennen ändert nichts.
	code, aus, fehler := fahreAn(t, ort, "ruecksetzung-vorschau", "profil", "1", "1")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s):\n%s", code, fehler, aus)
	}

	if !strings.Contains(aus, "bereit\ttrue") {
		t.Errorf("das eigene Profil gilt als gesperrt:\n%s", aus)
	}

	// **Eine Kennung, die keine ist**, wird am Feld benannt.
	krumm, _, klage := fahreAn(t, ort, "ruecksetzung-vorschau", "profil", "keine")
	if krumm != CodeFachlich {
		t.Errorf("mit krummer Kennung: %d", krumm)
	}

	if !strings.Contains(klage, "profilID") && !strings.Contains(klage, "Kennung") {
		t.Errorf("die Klage nennt das Feld nicht: %s", klage)
	}

	// **Und beim Ausführen ebenso.**
	if beimAusfuehren, _, _ := fahreAn(t, ort,
		"zuruecksetzen", "profil", "Meins", "", "keine"); beimAusfuehren != CodeFachlich {
		t.Errorf("beim Ausführen: %d", beimAusfuehren)
	}
}

func TestEineKrummeRuecksetzungWirdAbgewiesen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, argumente := range map[string][]string{
		"ohne Art":            {"ruecksetzung-vorschau"},
		"mit zu vielen":       {"ruecksetzung-vorschau", "a", "b", "c", "d"},
		"erfundene Art":       {"ruecksetzung-vorschau", "alles"},
		"zuruecksetzen leer":  {"zuruecksetzen"},
		"zuruecksetzen viele": {"zuruecksetzen", "a", "b", "c", "d", "e", "f"},
	} {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code == CodeGut {
				t.Error("kam durch")
			}
		})
	}
}

// **Das Layout lässt sich zurücksetzen** — und die ganze Anwendung nicht ohne
// Freigabe.
func TestZurueckgesetztWirdMitAusgabeUndOhneFreigabeNicht(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "zuruecksetzen", "layout", "")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "zurueckgesetzt\t") {
		t.Errorf("die Ausgabe nennt die Zahl nicht:\n%s", aus)
	}

	code, aus, _ = fahreAn(t, ort, "zuruecksetzen", "layout", "", "--json")
	if code != CodeGut || !strings.Contains(aus, `"zurueckgesetzt"`) {
		t.Errorf("als JSON: %d %s", code, aus)
	}

	// **Ohne Revision wird nichts gelöscht** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2.
	code, _, klage := fahreAn(t, ort, "zuruecksetzen", "anwendung", "alles weg")
	if code != CodeFachlich || !strings.Contains(klage, "Revision") {
		t.Errorf("ohne Revision: %d %s", code, klage)
	}

	// **Die ganze Anwendung verlangt einen Nachweis**, und den gibt es hier
	// nicht: er entsteht am Fenster, von einem Menschen.
	//
	// **Mit der Revision aus der Vorschau** — sonst endete der Lauf eine
	// Prüfung früher und sagte nichts über den Nachweis.
	marke := revisionAusVorschau(t, ort, "anwendung")

	code, _, klage = fahreAn(t, ort,
		"zuruecksetzen", "anwendung", "alles weg", "--revision", marke)
	if code != CodeFachlich {
		t.Errorf("ohne Nachweis: %d", code)
	}

	if !strings.Contains(klage, "Freigabe") {
		t.Errorf("die Klage sagt nicht, was fehlt: %s", klage)
	}

	// **Und als JSON ebenso** — mit derselben Aussage im Rückgabewert.
	if alsJSON, _, _ := fahreAn(t, ort, "zuruecksetzen", "anwendung",
		"alles weg", "--revision", marke, "--json"); alsJSON != CodeFachlich {
		t.Errorf("als JSON: %d", alsJSON)
	}
}

// revisionAusVorschau geht den Weg, den eine Bedienung geht: erst nachsehen.
//
// **Über die Zeilenausgabe und nicht über das JSON** — Befund 9 des Nachlaufs:
// die Marke stand nur im JSON, und damit war die Fähigkeit an der einen Fläche
// zweigeteilt (§16).
func revisionAusVorschau(t *testing.T, ort, art string) string {
	t.Helper()

	code, aus, klage := fahreAn(t, ort, "ruecksetzung-vorschau", art)
	// **Auch mit Hindernissen** — der Befund ist ein Rückgabewert und kein
	// Fehlschlag, und die Marke steht dann trotzdem da.
	if code != CodeGut && code != CodeFachlich {
		t.Fatalf("die Vorschau: %d (%s)", code, klage)
	}

	for zeile := range strings.SplitSeq(aus, "\n") {
		if marke, gefunden := strings.CutPrefix(zeile, "revision\t"); gefunden {
			return strings.TrimSpace(marke)
		}
	}

	t.Fatalf("die Vorschau nennt keine Revision:\n%s", aus)

	return ""
}
