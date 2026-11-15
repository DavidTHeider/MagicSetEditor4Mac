// **Ein Spiel- und ein Stilpaket anlegen — über jede Fläche**
// (#1350).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Die Zusage, um die es geht
//
// Ein `.mse-set` ließ sich schreiben, eine Datei **in** einem Paket ändern —
// ein `.mse-game` oder `.mse-style` **anlegen** nicht. Wer sich eine eigene
// Vorlage bauen wollte, legte die Dateien von Hand an.
//
// # Und die Abnahme verlangt beide
//
// „Ein Set lässt sich darauf anlegen" — und ein Set trägt Spiel **und** Stil.
// Ein Spielpaket allein wäre eine Fähigkeit, die man nicht benutzen kann.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestEinPaketWirdAnJederFlaecheAngelegt(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-anlegen", "mse-game", "clispiel", "Spiel der Kommandozeile"); code != 0 {
			t.Fatalf("anlegen: %d (%s)", code, fehler)
		}

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-anlegen", "mse-style", "schlicht", "clispiel", "Schlicht"); code != 0 {
			t.Fatalf("den Stil anlegen: %d (%s)", code, fehler)
		}

		imKatalog(t, umgebung, "die Kommandozeile", "clispiel")
		imKatalog(t, umgebung, "die Kommandozeile", "clispiel-schlicht")
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.anlegen",
			map[string]any{"art": "mse-game", "name": "mcpspiel"}); istFehler {
			t.Fatalf("anlegen: %s", text)
		}

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.anlegen",
			map[string]any{
				"art": "mse-style", "name": "schlicht", "spiel": "mcpspiel",
			}); istFehler {
			t.Fatalf("den Stil anlegen: %s", text)
		}

		imKatalog(t, umgebung, "MCP", "mcpspiel")
		imKatalog(t, umgebung, "MCP", "mcpspiel-schlicht")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPost, "/pakete/neu",
			`{"art":"mse-game","name":"restspiel"}`); status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		if status, rumpf := an.ruf(t, http.MethodPost, "/pakete/neu",
			`{"art":"mse-style","name":"schlicht","spiel":"restspiel"}`,
		); status != http.StatusCreated {
			t.Fatalf("POST Stil: %d (%s)", status, rumpf)
		}

		imKatalog(t, umgebung, "REST", "restspiel")
		imKatalog(t, umgebung, "REST", "restspiel-schlicht")
	})
}

// **Und dieselbe fachliche Abweisung: ein Stil ohne Spiel** (§16).
//
// Sie ist der Grund, warum dieser Lauf beide Hälften braucht. Ein Stil legt
// die Felder **seines** Spiels aus; ohne es hätte er nichts, wonach er sich
// richten könnte — und ein Paket, das nichts auslegt, malt eine leere Karte.
func TestEinStilOhneSpielWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		// **Ein leeres Spiel und kein fehlendes Wort.** Seit die Wortstellung
		// von der Art abhängt, weist die Kommandozeile ein fehlendes Wort
		// schon beim Zerlegen ab — und dann prüfte dieser Lauf die
		// Bedienung statt der Fachlichkeit. Die drei Flächen sollen aber
		// **denselben fachlichen** Fehler geben (§16), und der kommt aus
		// `service.pruefe`.
		if _, _, code := kommandozeile(t, umgebung,
			"paket-anlegen", "mse-style", "waise", ""); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}
	})

	// **Und ein fehlendes Wort wird ebenso abgewiesen** — an der Bedienung,
	// nicht an der Fachlichkeit. Beides gehört geprüft: der Lauf darüber
	// belegte bis zum Zweitblick nur, dass **irgendetwas** schiefgeht.
	t.Run("Kommandozeile, ohne das Wort", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if _, _, code := kommandozeile(t, umgebung,
			"paket-anlegen", "mse-style", "waise"); code == 0 {
			t.Error("die Kommandozeile hat ein fehlendes Spiel hingenommen")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.anlegen",
			map[string]any{"art": "mse-style", "name": "waise"}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPost, "/pakete/neu",
			`{"art":"mse-style","name":"waise"}`); status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// **Und eine Art, die sich nicht anlegen lässt, weisen alle drei gleich ab** —
// der ausführbare Test der Zellen »Erstellen« für Symbolschrift, Sprachpaket,
// Include und Exportvorlage (FMT-01.AK-02,
// #1690).
// Je eine, die nicht anwendbar ist, und eine, die noch fehlt: die Abweisung
// kommt an allen Flächen aus `service`, mit **demselben** Satz.
func TestEineNichtAnlegbareArtWeistJedeFlaecheAb(t *testing.T) {
	const satz = "Anlegen lässt sich ein Spiel (mse-game), ein Stil (mse-style)"

	// **Alle vier, nicht zwei.** Der erste Anlauf fuhr je eine, die nicht
	// anwendbar ist, und eine, die fehlt — und sein Kopf versprach alle vier
	// (Nachlauf zum Zweitblick über Phase 6.0).
	// **Include und Exportvorlage gehen seit FMT-02** (#1691) —
	// `paketgeruest_test.go`.
	for _, art := range []string{"mse-symbol-font", "mse-locale"} {
		t.Run(art+", Kommandozeile", func(t *testing.T) {
			umgebung := mitEigenerSaat(t)

			if _, fehler, code := kommandozeile(t, umgebung,
				"paket-anlegen", art, "eigenes"); code == 0 || !strings.Contains(fehler, satz) {
				t.Errorf("Code %d, Fehler %q", code, fehler)
			}

			nichtImKatalog(t, umgebung, "die Kommandozeile", "eigenes")
		})

		t.Run(art+", MCP", func(t *testing.T) {
			umgebung := mitEigenerSaat(t)

			if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.anlegen",
				map[string]any{"art": art, "name": "eigenes"}); !istFehler || !strings.Contains(text, satz) {
				t.Errorf("Fehler %v, Text %q", istFehler, text)
			}

			nichtImKatalog(t, umgebung, "MCP", "eigenes")
		})

		t.Run(art+", REST", func(t *testing.T) {
			umgebung := mitEigenerSaat(t)
			an := starteDienst(t, umgebung)

			if status, rumpf := an.ruf(t, http.MethodPost, "/pakete/neu",
				`{"art":"`+art+`","name":"eigenes"}`); status != http.StatusBadRequest ||
				!strings.Contains(rumpf, satz) {
				t.Errorf("Status %d, Rumpf %q", status, rumpf)
			}

			nichtImKatalog(t, umgebung, "REST", "eigenes")
		})
	}
}

// nichtImKatalog sagt, dass eine Abweisung nichts angelegt hat — **gefragt an
// der Kommandozeile**, also an einer anderen Fläche als der abweisenden.
func nichtImKatalog(t *testing.T, umgebung []string, wer, name string) {
	t.Helper()

	if liste := mussKommandozeile(t, umgebung, "pakete"); strings.Contains(liste, name) {
		t.Errorf("nach der Abweisung durch %s steht %q im Katalog: %s", wer, name, liste)
	}
}

// **Und auf dem neuen Paar lässt sich ein Set anlegen** — die Abnahme selbst.
//
// Ohne diesen Lauf wären die beiden darüber ein Beleg dafür, dass zwei
// Verzeichnisse entstehen. Ob daraus eine **Vorlage** wird, sagt erst ein Set,
// das darauf steht.
func TestAufDemNeuenPaarLaesstSichEinSetAnlegen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	for _, argumente := range [][]string{
		{"paket-anlegen", "mse-game", "eigenes", "Eigenes Spiel"},
		{"paket-anlegen", "mse-style", "schlicht", "eigenes", "Schlicht"},
	} {
		if _, fehler, code := kommandozeile(t, umgebung, argumente...); code != 0 {
			t.Fatalf("%v: %d (%s)", argumente, code, fehler)
		}
	}

	if _, fehler, code := kommandozeile(t, umgebung,
		"set-neu", "Auf eigener Vorlage", "eigenes", "eigenes-schlicht"); code != 0 {
		t.Fatalf("das Set anlegen: %d (%s)", code, fehler)
	}

	// **Gelesen und nicht nur geschrieben.** „Angelegt" wäre auch ein Set, das
	// die Liste danach nicht führt — und genau das prüft die Abnahme nicht,
	// wenn man beim Anlegen stehen bleibt.
	liste := mussKommandozeile(t, umgebung, "sets", "--json")

	for _, teil := range []string{"Auf eigener Vorlage", "eigenes", "eigenes-schlicht"} {
		if !strings.Contains(liste, teil) {
			t.Errorf("die Setliste führt %q nicht:\n%s", teil, liste)
		}
	}
}

// **Ein Stil lässt sich ohne Vollnamen anlegen** — Befund 1 des Zweitblicks
// über Phase 4.10.
//
// # Der Befund
//
// Die Hilfe schrieb `paket-anlegen <art> <name> [voll] [spiel]` — beide
// letzten als wahlfrei. Sie sind es nicht: das dritte Wort war **immer** der
// Vollname, und wer es ausließ, gab damit auch das Spiel nicht an. `mse
// paket-anlegen mse-style schlicht eigenes` legte „eigenes" als Vollnamen ab
// und scheiterte am fehlenden Spiel.
//
// **Und der Lauf darüber sah es nicht**, weil er den Vollnamen mitgab — die
// Umgehung, nicht der Weg.
//
// # Warum die Wortstellung von der Art abhängt
//
// Weil die Begriffe es tun: ein **Spiel** hat kein Spiel. Zwei Zeilen in der
// Hilfe, jede eindeutig, sind ehrlicher als eine mit zwei wahlfreien Wörtern,
// von denen man das erste nicht auslassen kann.
func TestEinStilLaesstSichOhneVollnamenAnlegen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	if _, fehler, code := kommandozeile(t, umgebung,
		"paket-anlegen", "mse-game", "eigenes"); code != 0 {
		t.Fatalf("das Spiel anlegen: %d (%s)", code, fehler)
	}

	// **Drei Wörter für einen Stil**: Art, Name, Spiel — kein Vollname.
	if _, fehler, code := kommandozeile(t, umgebung,
		"paket-anlegen", "mse-style", "schlicht", "eigenes"); code != 0 {
		t.Fatalf("den Stil anlegen: %d (%s)", code, fehler)
	}

	imKatalog(t, umgebung, "die Kommandozeile", "eigenes-schlicht")
}

// **Und ein Spiel mit Vollnamen bleibt drei Wörter.**
//
// Die Gegenprobe zur Zeile darüber: hätte ich die Stellen einfach vertauscht,
// bräuchte ein Spiel mit Vollnamen ein leeres Wort dazwischen.
func TestEinSpielMitVollnamenBleibtDreiWoerter(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	if _, fehler, code := kommandozeile(t, umgebung,
		"paket-anlegen", "mse-game", "eigenes", "Eigenes Spiel"); code != 0 {
		t.Fatalf("anlegen: %d (%s)", code, fehler)
	}

	imKatalog(t, umgebung, "die Kommandozeile", "eigenes")
}
