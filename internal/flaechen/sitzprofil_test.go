package flaechen_test

// **Einen Sitz an ein Profil binden — an jeder Fläche** — UXR-F21
// (#2041),
// Folgebefund aus UXR-F16
// (#2031).
//
// # Was hier gemessen wird
//
// **Dass dieselbe Partie über Kommandozeile, MCP und Dienst angelegt dieselbe
// Auskunft gibt** (§16): welcher Sitz dem Profil gehört und ob die Partie auf
// ihn wartet. Bis dahin ging die Bindung nur am Fenster, und
// `partiendran_test.go` legte seine Partie deshalb über `internal/app` an —
// eine Partie, die an den drei anderen Flächen nie einen eigenen Sitz hatte.
//
// # Die Matrix
//
// **Drei Anleger und drei Leser**: jede Fläche legt eine Partie mit gebundenem
// Sitz und eine ohne Bindung an, und jede Fläche liest beide zurück. Wer nur
// den Anleger mit dem Leser derselben Fläche verglich, sähe nicht, dass die
// Bindung unterwegs verloren ging — sie käme dann in beiden Hälften gleich
// nicht an.
//
// # Und ein Profil, das es nicht gibt
//
// **Ist ein Befund und kein Fehler** — wie am Fenster (§8, G8,
// #1465):
// die Partie entsteht, niemand sitzt dort als „ich", und die Antwort nennt die
// Bindung, die ins Leere zeigt. **Ein Text, der keine Profilkennung ist**, ist
// dagegen ein fachlicher Fehler, und an jeder Fläche derselbe.

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// profilDesBestands gibt die Kennung des Profils, das der Bestand zeigt — das
// Profil, in dem Kommandozeile, MCP und Dienst ohne Token arbeiten.
func profilDesBestands(t *testing.T, umgebung []string) string {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		t.Fatalf("das Profil nennen: %v", fehler)
	}

	return strconv.FormatInt(profil, 10)
}

// anlegerSitzprofil legt je Fläche eine Partie an, deren zweiter Sitz an
// `profil` gebunden ist — leer heißt: ohne Bindung. Es gibt die rohe Antwort.
type anlegerSitzprofil func(profil string) string

// anlegerDerDreiFlaechen baut die drei Anleger über denselben zwei Decks.
func anlegerDerDreiFlaechen(
	t *testing.T, umgebung []string, an dienst, ersterDeck, zweiterDeck string,
) map[string]anlegerSitzprofil {
	t.Helper()

	return map[string]anlegerSitzprofil{
		"Kommandozeile": func(profil string) string {
			argumente := []string{
				"partie-anlegen", "magic",
				"anna:Anna:" + ersterDeck, "bert:Bert:" + zweiterDeck, "--json",
			}

			if profil != "" {
				argumente = append(argumente, "--sitzprofil=bert:"+profil)
			}

			return mussKommandozeile(t, umgebung, argumente...)
		},

		"MCP": func(profil string) string {
			zweiter := map[string]string{
				"kennung": "bert", "name": "Bert", "deckID": zweiterDeck,
			}

			if profil != "" {
				zweiter["profilID"] = profil
			}

			return mussMCP(t, umgebung, "schreiben.partie.anlegen",
				map[string]any{
					"spiel": "magic",
					"sitze": []map[string]string{
						{"kennung": "anna", "name": "Anna", "deckID": ersterDeck},
						zweiter,
					},
				})
		},

		"REST": func(profil string) string {
			zweiter := `{"kennung":"bert","name":"Bert","deckID":"` + zweiterDeck + `"`

			if profil != "" {
				zweiter += `,"profilID":"` + profil + `"`
			}

			return mussREST(t, an, http.MethodPost, "/partien",
				`{"spiel":"magic","sitze":[{"kennung":"anna","name":"Anna",`+
					`"deckID":"`+ersterDeck+`"},`+zweiter+`}]}`)
		},
	}
}

// partiekennungAus liest die Kennung aus der Antwort einer beliebigen Fläche.
func partiekennungAus(t *testing.T, roh string) string {
	t.Helper()

	return bildAusFlachOderCLI(t, roh).ID
}

// bildAusFlachOderCLI deutet die Antwort, gleich von welcher Fläche.
//
// **Die Kommandozeile umhüllt die Partie** (`{"partie":…}`), MCP und Dienst
// geben sie flach — [partiebild] kennt beide Gestalten einzeln.
func bildAusFlachOderCLI(t *testing.T, roh string) partiebild {
	t.Helper()

	var huelle struct {
		Partie *json.RawMessage `json:"partie"`
	}

	if err := json.Unmarshal([]byte(roh), &huelle); err == nil && huelle.Partie != nil {
		return bildAusCLI(t, roh)
	}

	return bildAusFlach(t, roh)
}

// **Dieselbe Partie mit gebundenem Sitz sagt an jeder Fläche dasselbe — von
// jeder Fläche angelegt.**
func TestDerGebundeneSitzIstAnJederFlaecheDerEigene(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)
	profil := profilDesBestands(t, umgebung)

	// gebunden und ungebunden: Anleger → Kennung
	gebunden, frei := map[string]string{}, map[string]string{}

	for anleger, anlegen := range anlegerDerDreiFlaechen(t, umgebung, an, ersterDeck, zweiterDeck) {
		gebunden[anleger] = partiekennungAus(t, anlegen(profil))
		frei[anleger] = partiekennungAus(t, anlegen(""))
	}

	leser := map[string]func() map[string]dranzeile{
		"Kommandozeile": func() map[string]dranzeile {
			return dranzeilenAus(t, mussKommandozeile(t, umgebung, "partien", "--json"), true)
		},
		"MCP": func() map[string]dranzeile {
			return dranzeilenAus(t, mussMCP(t, umgebung, "lesen.partien", map[string]any{}), false)
		},
		"REST": func() map[string]dranzeile {
			return dranzeilenAus(t, mussREST(t, an, http.MethodGet, "/partien", ""), false)
		},
	}

	for lesende, lesen := range leser {
		zeilen := lesen()

		for anleger, kennung := range gebunden {
			// **Der Beginn steht aus, und der gebundene Sitz hat noch einen
			// Schritt** — dieselbe Lage wie in `partiendran_test.go`.
			if zeile := zeilen[kennung]; zeile.EigenerSitz != "bert" || !zeile.Dran || zeile.Stand != "laeuft" {
				t.Errorf("von %s angelegt, von %s gelesen: eigener Sitz %q, dran=%v, Stand %q — erwartet bert, true, laeuft",
					anleger, lesende, zeile.EigenerSitz, zeile.Dran, zeile.Stand)
			}
		}

		for anleger, kennung := range frei {
			// **Ohne Bindung sitzt das gezeigte Profil am Wirt** — Vermerk
			// 0164; der Wirt ist der erste Sitz, wenn keiner genannt ist. Bis
			// dahin stand hier „sitzt nirgends", und die Marke „dran" erschien
			// für niemanden mit einem einzigen Profil.
			if zeile := zeilen[kennung]; zeile.ID != kennung || zeile.EigenerSitz != "anna" || !zeile.Dran {
				t.Errorf("ohne Bindung von %s angelegt, von %s gelesen: %+v — erwartet anna, dran",
					anleger, lesende, zeile)
			}
		}
	}
}

// **Ein Profil, das es nicht gibt, ist ein Befund — an jeder Fläche.**
func TestEinProfilOhneEsGibtIstAnJederFlaecheEinBefund(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	const verschwunden = "987654"

	for anleger, anlegen := range anlegerDerDreiFlaechen(t, umgebung, an, ersterDeck, zweiterDeck) {
		roh := anlegen(verschwunden)

		var antwort struct {
			Befunde []string `json:"befunde"`
		}

		if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
			t.Fatalf("%s: die Antwort ist kein gültiges JSON: %v (%s)", anleger, err, roh)
		}

		// **Der Befund nennt Sitz und Profil**, wie am Fenster.
		if len(antwort.Befunde) != 1 ||
			!strings.Contains(antwort.Befunde[0], `"bert"`) ||
			!strings.Contains(antwort.Befunde[0], verschwunden) {
			t.Errorf("%s: Befunde %q — erwartet einer, der bert und %s nennt",
				anleger, antwort.Befunde, verschwunden)
		}

		// **Die Partie ist entstanden und niemand sitzt als „ich" dort.**
		kennung := partiekennungAus(t, roh)

		for lesende, gelesen := range map[string]string{
			"Kommandozeile": mussKommandozeile(t, umgebung, "partien", "--json"),
			"MCP":           mussMCP(t, umgebung, "lesen.partien", map[string]any{}),
			"REST":          mussREST(t, an, http.MethodGet, "/partien", ""),
		} {
			zeile := dranzeilenAus(t, gelesen, lesende == "Kommandozeile")[kennung]
			if zeile.ID != kennung || zeile.EigenerSitz != "" || zeile.Dran {
				t.Errorf("%s angelegt, %s gelesen: %+v", anleger, lesende, zeile)
			}
		}
	}
}

// **Eine Profilkennung, die keine ist, weist jede Fläche gleich ab** — als
// fachlichen Fehler, der das Profil nennt.
func TestEineUngueltigeProfilkennungWeistJedeFlaecheAb(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	// Kommandozeile: nicht abbrechen, sondern den Code und den Satz lesen.
	aus, fehler, code := kommandozeile(t, umgebung, "partie-anlegen", "magic",
		"anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck,
		"--sitzprofil=bert:kein-profil")
	if code == 0 || !strings.Contains(aus+fehler, "Profil") {
		t.Errorf("Kommandozeile: Code %d, %q", code, aus+fehler)
	}

	// MCP: ein Fehlerergebnis, das das Profil nennt.
	text, istFehler := mcpRuf(t, umgebung, "schreiben.partie.anlegen",
		map[string]any{
			"spiel": "magic",
			"sitze": []map[string]string{
				{"kennung": "anna", "name": "Anna", "deckID": ersterDeck},
				{
					"kennung": "bert", "name": "Bert", "deckID": zweiterDeck,
					"profilID": "kein-profil",
				},
			},
		})
	if !istFehler || !strings.Contains(text, "Profil") {
		t.Errorf("MCP: Fehler=%v, %q", istFehler, text)
	}

	// Dienst: 400, und der Satz nennt das Profil.
	status, antwort := an.ruf(t, http.MethodPost, "/partien",
		`{"spiel":"magic","sitze":[{"kennung":"anna","name":"Anna","deckID":"`+
			ersterDeck+`"},{"kennung":"bert","name":"Bert","deckID":"`+
			zweiterDeck+`","profilID":"kein-profil"}]}`)
	if status != http.StatusBadRequest || !strings.Contains(antwort, "Profil") {
		t.Errorf("Dienst: %d, %q", status, antwort)
	}

	// **Und es ist keine Partie entstanden** — ein Fehler hinterlässt keinen
	// halben Stand (§10).
	if zeilen := dranzeilenAus(t, mussKommandozeile(t, umgebung, "partien", "--json"), true); len(zeilen) != 0 {
		t.Errorf("nach dreimal Abweisung stehen %d Partien da", len(zeilen))
	}
}
