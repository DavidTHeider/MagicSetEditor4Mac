// **Suchen, auszählen und einstellen über jede Fläche ohne Fenster**
// (#1169).
//
// # Warum die drei zusammen stehen
//
// Sie fassen den **Bestand als Ganzes** an und nicht ein einzelnes Set: die
// Volltextsuche geht über alle Sets, die Auszählung fasst eines zusammen, und
// die Einstellungen gehören dem Programm. Was sie teilen, ist die Vorrichtung
// — ein Bestand mit der Saat und einem eingefahrenen Set.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Volltextsuche findet an jeder Fläche dieselbe Karte.**
//
// # Die Zusage
//
// Das Probeset führt eine Karte namens „Doppelt". Jede Fläche muss sie auf
// dieselbe Frage nennen — und keine darf auf eine Frage, die niemand
// beantwortet, etwas zurückgeben.
func TestDieVolltextsucheUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	fuerAlleDrei(t, "die Volltextsuche", "Doppelt",
		func() string { return mussKommandozeile(t, umgebung, "suche", "Doppelt") },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.suche", map[string]any{"text": "Doppelt"})
			if istFehler {
				t.Fatalf("lesen.suche: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/suche?text=Doppelt", "")
			if status != http.StatusOK {
				t.Fatalf("GET suche: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Und was es nicht gibt, findet auch niemand.**
	//
	// Ohne diese Hälfte sagte der Lauf darüber nur, dass eine Antwort kommt —
	// eine Fläche, die stets alle Karten zurückgäbe, wäre grün.
	fuerKeineDrei(t, "die Volltextsuche", "Doppelt",
		func() string { return mussKommandozeile(t, umgebung, "suche", "Gibtsnirgends") },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.suche",
				map[string]any{"text": "Gibtsnirgends"})
			if istFehler {
				t.Fatalf("lesen.suche: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/suche?text=Gibtsnirgends", "")
			if status != http.StatusOK {
				t.Fatalf("GET suche: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Die Auszählung eines Sets nennt an jeder Fläche dieselbe Kartenzahl.**
//
// Das Probeset hat drei Karten — und eine Fläche, die vier oder zwei nennt,
// zählt etwas anderes als die anderen.
func TestDieAuszaehlungUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	// **Mit dem Wort und nicht bloß die Ziffer** — `"3"` steht auch in `"13"`
	// (#1253, Befund 4). Alle drei Flächen setzen die Zahl hinter das Wort:
	// die Tabelle mit einem Tabulator, MCP und der Dienst als JSON-Feld.
	fuerAlleDreiMit(t, "die Kennzahlen", [3]string{"Karten\t3", `"karten": 3`, `"karten":3`},
		func() string { return mussKommandozeile(t, umgebung, "statistik", setID) },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.statistik",
				map[string]any{"setID": zahl(t, setID)})
			if istFehler {
				t.Fatalf("lesen.statistik: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/statistik", "")
			if status != http.StatusOK {
				t.Fatalf("GET statistik: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Und die Verteilung eines Feldes.** Sie ist die zweite Hälfte
	// derselben Fähigkeit: `statistik` ohne Feld gibt die Kennzahlen, mit
	// Feld die Verteilung seiner Werte. Das Probeset führt `farbe` mit
	// „gruen".
	fuerAlleDrei(t, "die Verteilung", "gruen",
		func() string { return mussKommandozeile(t, umgebung, "statistik", setID, "farbe") },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.verteilung",
				map[string]any{"setID": zahl(t, setID), "feld": "farbe"})
			if istFehler {
				t.Fatalf("lesen.verteilung: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet,
				"/sets/"+setID+"/statistik/farbe", "")
			if status != http.StatusOK {
				t.Fatalf("GET verteilung: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Eine Einstellung, an einer Fläche gesetzt, gilt an allen.**
//
// # Warum sie hier anders geprüft wird als die Setangaben
//
// Die Einstellungen sind **eine** Ablage für das ganze Programm — es gibt
// keine drei davon. Die Zusage ist deshalb nicht „jede Fläche hat ihre",
// sondern das Gegenteil: **wer an einer setzt, sieht es an allen dreien.** Das
// ist die eigentliche Aussage von §16, und dieser Fall trägt sie am
// deutlichsten.
//
// # Gelesen wird der Wert, nicht ein Wort
//
// Hier stand `fuerAlleDrei(…, "en", …)`: „en" steht in jeder Antwort schon
// unter den möglichen Werten, an der Kommandozeile im Bereich `anwendung` —
// der Lauf war grün, gleich was gesetzt war (Zweitblick über Phase C,
// Befund 3). Jetzt liest jede Fläche den Wert von `sprache`, und gesetzt wird
// an drei Flächen, gelesen an allen vier.
func TestEineEinstellungGiltAnAllenFlaechen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	fenster := fensterAus(t, umgebung)

	lesen := map[string]func() string{
		"Fenster": func() string {
			antwort := fenster.Einstellungen(t.Context())
			if antwort.Fehler != nil {
				t.Fatalf("Einstellungen: %+v", antwort.Fehler)
			}

			return string(mussJSON(t, antwort))
		},
		"Kommandozeile": func() string { return mussKommandozeile(t, umgebung, "einstellungen", "--json") },
		"MCP": func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.einstellungen", map[string]any{})
			if istFehler {
				t.Fatalf("lesen.einstellungen: %s", aus)
			}

			return aus
		},
		"Dienst": func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/einstellungen", "")
			if status != http.StatusOK {
				t.Fatalf("GET einstellungen: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	}

	ueberall := func(nach, will string) {
		t.Helper()

		for flaeche, lies := range lesen {
			if steht := einstellungswert(t, lies(), "sprache"); steht != will {
				t.Errorf("nach %s sieht %s sprache=%q, gesetzt war %q", nach, flaeche, steht, will)
			}
		}
	}

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "sprache", "en")
	ueberall("dem Setzen an der Kommandozeile", "en")

	if status, rumpf := an.ruf(t, http.MethodPut, "/einstellungen/sprache",
		`{"wert":"de"}`); status != http.StatusOK {
		t.Fatalf("PUT einstellungen: %d (%s)", status, rumpf)
	}

	ueberall("dem Setzen über den Dienst", "de")

	if antwort := fenster.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("EinstellungSetzen: %+v", antwort.Fehler)
	}

	ueberall("dem Setzen im Fenster", "en")
}

// einstellungswert liest den Wert einer Einstellung aus der Gestalt, die alle
// vier Flächen teilen (`{"einstellungen":[{"schluessel","wert"}]}`).
func einstellungswert(t *testing.T, roh, schluessel string) string {
	t.Helper()

	var antwort struct {
		Einstellungen []struct {
			Schluessel string `json:"schluessel"`
			Wert       string `json:"wert"`
		} `json:"einstellungen"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("kein JSON: %v\n%s", fehler, roh)
	}

	for _, eine := range antwort.Einstellungen {
		if eine.Schluessel == schluessel {
			return eine.Wert
		}
	}

	t.Fatalf("die Einstellung %q steht nicht in der Antwort:\n%s", schluessel, roh)

	return ""
}

// **Eine Einstellung ohne Schlüssel wird überall gleich abgewiesen.**
func TestEineEinstellungOhneSchluesselWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Einstellung ohne Schlüssel", abweisung{
		kern:   "ist keine Einstellung dieser Anwendung",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"einstellungen", "setzen", "  ", "egal")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.einstellung",
				map[string]any{"schluessel": "  ", "wert": "egal"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPut,
				"/einstellungen/"+url.PathEscape("  "), `{"wert":"egal"}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).EinstellungSetzen(t.Context(), "  ", "egal").Fehler
		},
	})
}

// fuerKeineDrei ist die Verneinung zu [fuerAlleDrei].
//
// **Sie steht daneben, weil eine Zusage ohne Gegenfall keine ist**: ein Lauf,
// der nur prüft, dass etwas gefunden wird, bliebe grün, wenn eine Fläche
// stets alles zurückgäbe.
func fuerKeineDrei(t *testing.T, was, kern string, cli, mcp, rest func() string) {
	t.Helper()

	for name, hole := range map[string]func() string{
		"Kommandozeile": cli, "MCP": mcp, "REST": rest,
	} {
		if antwort := hole(); strings.Contains(antwort, kern) {
			t.Errorf("%s nennt bei %q das %q, das nicht gesucht war:\n%s",
				name, was, kern, antwort)
		}
	}
}
