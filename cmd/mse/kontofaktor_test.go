package main

// **Konten und der zweite Faktor an der Kommandozeile** — Werkstücke 6 und 7
// von Phase 4.12
// (#1351).
//
// # Was hier geprüft wird und was nebenan
//
// **Hier die Bedienung** — welche Wörter ein Befehl nimmt, was er ohne sie
// sagt, und dass das Kennwort von der Standardeingabe kommt.
//
// **Der ganze Weg steht in `internal/flaechen`**: er braucht einen eigenen
// Schlüsselbund, und den legt der Abnahmelauf an.

import (
	"strings"
	"testing"
)

// **Jeder Kontobefehl weist einen falschen Aufruf ab — mit `CodeBedienung`.**
//
// **Und das ist der Unterschied, der zählt** (§4): ein Skript, das
// `CodeFachlich` von `CodeBedienung` nicht unterscheidet, wiederholt einen
// Aufruf, der nie funktionieren wird.
func TestJederKontobefehlWeistEinenFalschenAufrufAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string][]string{
		"konto-neu ohne Kennung":          {"konto-neu"},
		"konto-neu mit zu vielen Wörtern": {"konto-neu", "a", "b", "c"},
		"konto-anmelden ohne Kennung":     {"konto-anmelden"},
		"konto-anonym mit einem Wort":     {"konto-anonym", "zuviel"},
		"faktoren ohne Nutzer":            {"faktoren"},
		"faktor-neu ohne Art":             {"faktor-neu", "1"},
		"faktor-bestaetigen ohne Code":    {"faktor-bestaetigen", "1"},
		"faktor-loeschen ohne Art":        {"faktor-loeschen", "1"},
		"profil mit zu vielen Wörtern":    {"profil", "1", "2", "3"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			code, _, fehler := fahreAn(t, ort, argumente...)
			if code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet war %d (%s)",
					code, CodeBedienung, fehler)
			}
		})
	}
}

// **`mse konto-anonym` legt ein Konto ohne jede Angabe an.**
//
// **Der einzige Kontobefehl ohne Schlüsselbund** — und deshalb der einzige,
// der hier ganz gefahren wird.
func TestKontoAnonymLegtEinKontoAn(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "konto-anonym")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "anonym-") {
		t.Errorf("die Ausgabe nennt keine Zufallskennung: %q", aus)
	}

	// **Und er steht danach in der Liste.**
	if _, liste, _ := fahreAn(t, ort, "nutzer"); !strings.Contains(liste, "anonym-") {
		t.Errorf("die Liste kennt ihn nicht: %q", liste)
	}
}

// **`mse konto-neu` liest das Kennwort von der Standardeingabe.**
//
// **Ohne sie sagt es das** — und weist ab, statt mit einem leeren Kennwort
// weiterzumachen.
func TestKontoNeuBrauchtDasKennwortVonDerEingabe(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, _, fehler := fahreAnMitEingabe(t, ort, "kurz",
		"konto-neu", "wer@example.org")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	// **Die Meldung nennt das Kennwort und nicht die Kennung** (§6): der
	// Fehler steht am verursachenden Feld.
	if !strings.Contains(fehler, "Kennwort") {
		t.Errorf("die Meldung ist %q", fehler)
	}
}

// **`mse faktoren` antwortet für einen Nutzer ohne Faktoren mit nichts.**
//
// # Was hier **nicht** zugesichert wird
//
// **Ob dieser Rechner Biometrie kann.** Der Befehl schreibt eine Zeile
// `# Biometrie: …`, wenn er sie nicht kann — und auf dem Rechner, auf dem
// dieser Lauf gerade läuft, kann er sie vielleicht. Ein Test, der die Zeile
// verlangte, prüfte die Ausstattung des Prüfenden und nicht das Programm
// (§10: die Reichweite einer Messung ist die Reichweite ihrer Aussage).
//
// **Dass die Auskunft ankommt, prüft `internal/app`** — dort steht ein
// Sensor, den der Lauf setzt.
func TestFaktorenAntwortetOhneFaktorenMitNichts(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "faktoren", "1")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, zeile := range strings.Split(strings.TrimSpace(aus), "\n") {
		if zeile != "" && !strings.HasPrefix(zeile, "#") {
			t.Errorf("es steht ein Faktor da: %q", zeile)
		}
	}
}

// **Ein Faktor, den es nicht gibt, lässt sich nicht löschen.**
func TestEinFaktorDenEsNichtGibtLaesstSichNichtLoeschen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "faktor-loeschen", "1", "totp")
	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d (%s)", code, fehler)
	}
}

// **Und `faktor-bestaetigen` ohne Einrichtung ebenso.**
func TestFaktorBestaetigenOhneEinrichtungScheitert(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "faktor-bestaetigen", "1", "123456")
	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d (%s)", code, fehler)
	}
}
