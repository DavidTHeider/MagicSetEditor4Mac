// **Preise und Sammlungsauskunft über jede Fläche ohne Fenster**
// (#1169).
//
// # Warum die sechs zusammen stehen
//
// Sie hängen an **einem** Druck und bauen aufeinander: Preise holen legt sie
// ab, „die abgelegten Preise" liest sie, der Verlauf liest dieselbe Reihe über
// die Zeit, das Portfolio rechnet sie gegen die eigene Sammlung, und
// „nachsehen, was von einem Druck da ist" fragt dieselbe Sammlung von der
// anderen Seite. Sechs Läufe daneben holten sechsmal dieselben Preise.
//
// **Kein Aufruf geht hinaus** — die Gegenstelle steht in
// `scryfallgegenstelle_test.go`, und ihre `/cards/:id` trägt seit #1169 einen
// `prices`-Block.
package flaechen_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// **Preise holen, ablegen und wiederlesen — über jede Fläche.**
//
// # Die Zusage
//
// Jede Fläche holt für **ihren eigenen** Druck und liest ihn zurück. Der Preis
// steht in der Antwort, und er ist der, den die Gegenstelle nennt — nicht
// irgendeiner.
//
// **Eigene Drucke und kein gemeinsamer**: drei Flächen auf demselben Druck
// träfen sich in der abgelegten Reihe, und der Lauf prüfte am Ende, wer
// zuletzt geholt hat.
func TestPreiseHolenUndLesenUeberJedeFlaeche(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	an := starteDienst(t, umgebung)

	// **1,50 EUR steht in der Gegenstelle** — dieselbe Zahl muss aus jeder
	// Fläche zurückkommen.
	const preis = "1.50"

	for name, fall := range map[string]struct {
		druck string
		holen func(string)
		lesen func(string) string
	}{
		"Kommandozeile": {
			druck: "druck-fuer-die-zeile",
			holen: func(druck string) { mussKommandozeile(t, umgebung, "preise-holen", druck) },
			lesen: func(druck string) string {
				return mussKommandozeile(t, umgebung, "preise", druck)
			},
		},
		"MCP": {
			druck: "druck-fuer-das-modell",
			holen: func(druck string) {
				if text, istFehler := mcpRuf(t, umgebung, "schreiben.preiseholen",
					map[string]any{"printing": druck}); istFehler {
					t.Fatalf("schreiben.preiseholen: %s", text)
				}
			},
			lesen: func(druck string) string {
				aus, istFehler := mcpRuf(t, umgebung, "lesen.preise",
					map[string]any{"printing": druck})
				if istFehler {
					t.Fatalf("lesen.preise: %s", aus)
				}

				return aus
			},
		},
		"REST": {
			druck: "druck-fuer-die-schleife",
			holen: func(druck string) {
				if status, rumpf := an.ruf(t, http.MethodPost,
					"/drucke/"+druck+"/preise", ""); status != http.StatusOK {
					t.Fatalf("POST preise: %d (%s)", status, rumpf)
				}
			},
			lesen: func(druck string) string {
				status, rumpf := an.ruf(t, http.MethodGet, "/drucke/"+druck+"/preise", "")
				if status != http.StatusOK {
					t.Fatalf("GET preise: %d (%s)", status, rumpf)
				}

				return rumpf
			},
		},
	} {
		fall.holen(fall.druck)

		if antwort := fall.lesen(fall.druck); !strings.Contains(antwort, preis) {
			t.Errorf("%s liest den Preis %s nicht zurück:\n%s", name, preis, antwort)
		}
	}
}

// **Der Preisverlauf nennt an jeder Fläche dieselbe Reihe.**
//
// # Warum ein Lauf für sich
//
// Er liest **dieselben** abgelegten Preise wie die Zeile darüber, aber über
// die **Zeit** und je Währung und Ausführung. Ohne ihn stünde von den drei
// Preisbindungen eine ungefahren da — und das ist genau die Lücke, die der
// Wächter aus #1169 zählt.
func TestDerPreisverlaufUeberJedeFlaeche(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	const druck = "druck-mit-verlauf"

	mussKommandozeile(t, umgebung, "preise-holen", druck)

	an := starteDienst(t, umgebung)

	fuerAlleDrei(t, "der Preisverlauf", "1.50",
		func() string {
			// **Die ganze Reihe und keine Vorgabe** (A3, #1442): welche
			// Reihe gemeint ist, sagt der Aufrufer — an jeder Fläche.
			return mussKommandozeile(t, umgebung, "preisverlauf", druck,
				"scryfall", "paper", "retail", "EUR", "nonfoil")
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.preisverlauf",
				map[string]any{
					"printing": druck, "quelle": "scryfall", "markt": "paper",
					"art": "retail", "waehrung": "EUR", "finish": "nonfoil",
				})
			if istFehler {
				t.Fatalf("lesen.preisverlauf: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet,
				"/drucke/"+druck+"/preisverlauf?quelle=scryfall&markt=paper"+
					"&art=retail&waehrung=EUR&finish=nonfoil", "")
			if status != http.StatusOK {
				t.Fatalf("GET preisverlauf: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Was von einem Druck in der Sammlung liegt, sagt jede Fläche gleich — und
// das Portfolio rechnet es zusammen.**
//
// # Warum beide in einem Lauf
//
// Sie lesen **dieselbe** Sammlung von zwei Seiten: die eine fragt nach einem
// Druck, die andere nach der Summe. Ein Exemplar, das nur eine der beiden
// sieht, ist der Fehler, den dieser Lauf finden soll.
func TestSammlungFuerEinenDruckUndDasPortfolioUeberJedeFlaeche(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	const druck = "druck-in-der-sammlung"

	// **Zwei Exemplare, damit die Summe etwas zu addieren hat.**
	mussKommandozeile(t, umgebung, "exemplar", druck, "2", "played", "de", "nonfoil")
	mussKommandozeile(t, umgebung, "preise-holen", druck)

	an := starteDienst(t, umgebung)

	fuerAlleDrei(t, "was von einem Druck da ist", druck,
		func() string { return mussKommandozeile(t, umgebung, "sammlung", druck) },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.sammlungfuer",
				map[string]any{"printing": druck})
			if istFehler {
				t.Fatalf("lesen.sammlungfuer: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/sammlung/drucke/"+druck, "")
			if status != http.StatusOK {
				t.Fatalf("GET sammlungfuer: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Und das Portfolio nennt die zwei Karten.**
	// **Mit dem Wort und nicht bloß die Ziffer** (#1253, Befund 4): die
	// Kommandozeile schreibt »2 Exemplare«, die beiden JSON-Flächen ein Feld.
	fuerAlleDreiMit(t, "das Portfolio", [3]string{"2 Exemplare", `"exemplare": 2`, `"exemplare":2`},
		func() string { return mussKommandozeile(t, umgebung, "portfolio") },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.portfolio", map[string]any{})
			if istFehler {
				t.Fatalf("lesen.portfolio: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/portfolio", "")
			if status != http.StatusOK {
				t.Fatalf("GET portfolio: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Ein Exemplar herausnehmen — über jede Fläche, und danach ist es fort.**
//
// # Die Fassung ist die halbe Zusage
//
// MCP und der Dienst verlangen sie: wer löscht, sagt, welchen Stand er
// gesehen hat. Ein Löschen ohne sie träfe einen Stand, den ein anderer
// inzwischen geändert hat — dieselbe Regel wie beim Ändern (#1196).
//
// **Die Kommandozeile nimmt sie nicht entgegen**, und das ist die begründete
// Ausnahme aus §16: wer tippt, hat den vorigen Stand nicht zur Hand. Sie liest
// ihn selbst.
func TestEinExemplarHerausnehmenUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	an := starteDienst(t, umgebung)

	for name, loeschen := range map[string]func(flaechenexemplar){
		"Kommandozeile": func(exemplar flaechenexemplar) {
			mussKommandozeile(t, umgebung, "exemplar-loeschen", fmt.Sprint(exemplar.ID))
		},
		"MCP": func(exemplar flaechenexemplar) {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.exemplarloeschen",
				map[string]any{"id": exemplar.ID, "fassung": exemplar.Fassung}); istFehler {
				t.Fatalf("schreiben.exemplarloeschen: %s", text)
			}
		},
		"REST": func(exemplar flaechenexemplar) {
			weg := fmt.Sprintf("/sammlung/%d?fassung=%d", exemplar.ID, exemplar.Fassung)
			if status, rumpf := an.ruf(t, http.MethodDelete, weg, ""); status != http.StatusOK &&
				status != http.StatusNoContent {
				t.Fatalf("DELETE sammlung: %d (%s)", status, rumpf)
			}
		},
	} {
		druck := "druck-fuer-" + strings.ToLower(name)

		mussKommandozeile(t, umgebung, "exemplar", druck, "1", "played", "de", "nonfoil")

		exemplar := exemplarFuer(t, umgebung, druck)
		loeschen(exemplar)

		if aus := mussKommandozeile(t, umgebung, "sammlung"); strings.Contains(aus, druck) {
			t.Errorf("%s: das Exemplar steht nach dem Herausnehmen noch da:\n%s", name, aus)
		}
	}
}

// **Ein Exemplar mit veralteter Fassung wird abgewiesen.**
//
// Der Fehlerfall neben dem Erfolg (§16) — und die Zusage, für die es die
// Fassung überhaupt gibt. **Nur an den beiden Flächen, die sie führen**: die
// Kommandozeile liest selbst, und das steht begründet in der Zeile darüber.
func TestEineVeralteteFassungBeimHerausnehmenWirdAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	const druck = "druck-mit-alter-fassung"

	mussKommandozeile(t, umgebung, "exemplar", druck, "1", "played", "de", "nonfoil")

	exemplar := exemplarFuer(t, umgebung, druck)
	veraltet := exemplar.Fassung + 99

	an := starteDienst(t, umgebung)

	// **Nur zwei Flächen, und das ist die begründete Ausnahme.** Die
	// Kommandozeile nimmt die Fassung nicht entgegen: wer tippt, hat den
	// vorigen Stand nicht zur Hand — sie liest ihn selbst. Ein dritter Aufruf
	// hier prüfte deshalb eine Zusage, die es nicht gibt.
	if text, istFehler := mcpRuf(t, umgebung, "schreiben.exemplarloeschen",
		map[string]any{"id": exemplar.ID, "fassung": veraltet}); !istFehler {
		t.Errorf("MCP löschte gegen eine veraltete Fassung: %s", text)
	}

	weg := fmt.Sprintf("/sammlung/%d?fassung=%d", exemplar.ID, veraltet)

	status, rumpf := an.ruf(t, http.MethodDelete, weg, "")
	if status != http.StatusConflict {
		t.Errorf("der Status ist %d — erwartet war %d (Konflikt):\n%s",
			status, http.StatusConflict, rumpf)
	}

	// **Der Code im Rumpf und nicht nur der Status** — dieselbe Zusage wie in
	// [abweisungGleich], und aus demselben Grund: nach ihm verzweigt ein
	// Aufrufer.
	if !strings.Contains(rumpf, `"code":"`+klasseKonflikt+`"`) {
		t.Errorf("der Rumpf nennt die Klasse %q nicht:\n%s", klasseKonflikt, rumpf)
	}

	// **Und danach steht das Exemplar noch da.** Ohne diese Zeile sagte der
	// Lauf nur, dass eine Meldung kommt — nicht, dass nichts geschah.
	if aus := mussKommandozeile(t, umgebung, "sammlung"); !strings.Contains(aus, druck) {
		t.Errorf("das Exemplar wurde trotz veralteter Fassung gelöscht:\n%s", aus)
	}
}

// alleAusHuelle liest **alle** Exemplare aus einer Antwort von `mse sammlung`.
//
// **Neben [ersteAusHuelle] und nicht statt ihr**: die dort verlangt genau
// eines und sagt es, wenn es mehr sind — das ist die Zusage jener Läufe. Hier
// liegen mehrere, und gesucht wird eines davon.
func alleAusHuelle(t *testing.T, roh string) []flaechenexemplar {
	t.Helper()

	var huelle struct {
		Exemplare []flaechenexemplar `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &huelle); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	return huelle.Exemplare
}

// exemplarFuer sucht das Exemplar eines Drucks in der Sammlung.
func exemplarFuer(t *testing.T, umgebung []string, druck string) flaechenexemplar {
	t.Helper()

	for _, eines := range alleAusHuelle(t, mussKommandozeile(t, umgebung, "--json", "sammlung")) {
		if eines.Printing == druck {
			return eines
		}
	}

	t.Fatalf("kein Exemplar für den Druck %q in der Sammlung", druck)

	return flaechenexemplar{}
}
