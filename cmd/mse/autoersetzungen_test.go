package main

// **Die Autoersetzungen an der Kommandozeile** — Werkstück E von Phase 4.11
// (#1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort läuft derselbe
// Ablauf über alle drei Flächen. Hier stehen die **Bedienfehler**: was der
// Befehl abweist, bevor er den Bestand anfasst.

import (
	"strings"
	"testing"
)

// **Ohne Set ist es ein Bedienfehler.**
func TestAutoersetzungOhneSetIstEinBedienfehler(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "autoersetzung"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Eine Kennung, die keine ist, ebenso.**
func TestAutoersetzungMitUnbrauchbarerKennung(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "autoersetzung", "keine Zahl"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Eine Handlung, die es nicht gibt, ebenso — und sie nennt die, die es gibt.**
func TestEineUnbekannteAutoersetzungshandlung(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	code, _, fehler := fahreAn(t, ort, "autoersetzung", setID, "tanzen")
	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}

	if !strings.Contains(fehler, "zuruecksetzen") {
		t.Errorf("die Meldung nennt die Handlungen nicht: %q", fehler)
	}
}

// **Setzen ohne Ersatz ist ein Bedienfehler**, löschen ohne Muster auch.
func TestAutoersetzungMitFehlendenAngaben(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	if code, _, _ := fahreAn(t, ort,
		"autoersetzung", setID, "setzen", "(W)"); code != CodeBedienung {
		t.Error("setzen ohne Ersatz kam durch")
	}

	if code, _, _ := fahreAn(t, ort,
		"autoersetzung", setID, "loeschen"); code != CodeBedienung {
		t.Error("loeschen ohne Muster kam durch")
	}
}

// **Ein leeres Muster ist ein fachliches Nein** — eine Regel ohne Muster
// greift nie.
func TestEineAutoersetzungOhneMusterIstEinFachlichesNein(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	if code, _, _ := fahreAn(t, ort,
		"autoersetzung", setID, "setzen", "   ", "x"); code != CodeFachlich {
		t.Error("das leere Muster kam durch")
	}
}

// **Die Liste eines Sets ohne Spielpaket ist leer und kein Fehler.**
//
// §8: ein Set, dessen Paket nicht daliegt, ist auf einem frischen Rechner der
// Normalfall — und eine leere Liste ist die richtige Antwort.
func TestDieAutoersetzungslisteOhneSpielpaket(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	code, aus, fehler := fahreAn(t, ort, "autoersetzung", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"regeln"`) || !strings.Contains(aus, `"an": true`) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Setzen, Schalten und Zurücksetzen gehen durch — und die Regel steht dann
// da.**
func TestEineEigeneAutoersetzungAnDerKommandozeile(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	// **`teilwort` und `aus` sind Beiworte** — ohne sie gilt „ganzes Wort"
	// und „an", wie bei MSE.
	if code, _, fehler := fahreAn(t, ort, "autoersetzung", setID,
		"setzen", "(W)", "weiss", "teilwort", "aus"); code != CodeGut {
		t.Fatalf("setzen: %d (%s)", code, fehler)
	}

	_, aus, _ := fahreAn(t, ort, "autoersetzung", setID)

	if !strings.Contains(aus, "eigen\taus\t(W)\tweiss") {
		t.Errorf("die Liste lautet %q", aus)
	}

	if code, _, _ := fahreAn(t, ort, "autoersetzung", setID, "aus"); code != CodeGut {
		t.Error("der Schalter ließ sich nicht ausschalten")
	}

	if _, aus, _ := fahreAn(t, ort, "autoersetzung", setID); !strings.HasPrefix(aus, "aus\n") {
		t.Errorf("die Liste sagt nicht, dass der Schalter aus ist: %q", aus)
	}

	if code, _, _ := fahreAn(t, ort, "autoersetzung", setID, "an"); code != CodeGut {
		t.Error("der Schalter ließ sich nicht einschalten")
	}

	if code, _, _ := fahreAn(t, ort,
		"autoersetzung", setID, "loeschen", "(W)"); code != CodeGut {
		t.Error("die eigene Regel ließ sich nicht löschen")
	}

	if code, _, _ := fahreAn(t, ort,
		"autoersetzung", setID, "zuruecksetzen"); code != CodeGut {
		t.Error("das Zurücksetzen scheiterte")
	}
}
