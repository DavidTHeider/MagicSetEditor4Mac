package flaechen_test

// **Ob man dran ist, sagt jede Fläche gleich** — UXR-F16
// (#2031).
//
// # Was hier gemessen wird
//
// **Dass Kommandozeile, Dienst und MCP für dieselbe Partie dasselbe sagen** (§16):
// welcher Sitz dem Profil gehört und ob die Partie auf ihn wartet. Das
// Fenster trägt dasselbe Feld; es zeichnet `internal/app/partiendran_intern_test.go`
// nach — dieser Lauf fährt das Fenster nicht (gemessen in `FLAECHENPARITAET.md`).
//
// # Drei Partien in einem Bestand
//
//   - **Eine mit gebundenem Sitz**: der zweite Sitz gehört dem gezeigten
//     Profil, und der Beginn steht aus (CR 103.5, Fassung 2026-09-25) — die
//     Partie wartet also auf ihn.
//   - **Eine, deren gebundener Sitz einem zweiten Profil gehört**, an der
//     Kommandozeile angelegt: das gezeigte Profil sieht zu, und niemand ist
//     dran. **Ohne jede Bindung** säße es am Wirt (Vermerk 0164, UXR-F21) —
//     das hält `sitzprofil_test.go`.
//   - **Dieselbe erste, angehalten**: der eigene Sitz bleibt, und es wartet
//     nichts mehr.
//
// **Die Bindung kommt am Fenster her.** Kommandozeile, MCP und Dienst kennen sie
// beim Anlegen nicht (G8, #1465) — wie das Fenster öffnet dieser Lauf deshalb
// den Bestand und legt die Partie über `internal/app` an, so wie es
// `handelsleitung_test.go` für die Freigabe tut.

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dranzeile ist, was jede Fläche über „dran" sagt.
type dranzeile struct {
	ID          string `json:"id"`
	Stand       string `json:"stand"`
	EigenerSitz string `json:"eigenerSitz"`
	Dran        bool   `json:"dran"`
}

// einePartieAmProfil legt eine Partie an, deren zweiter Sitz an das gezeigte
// Profil gebunden ist, und gibt ihre Kennung.
func einePartieAmProfil(
	t *testing.T, umgebung []string, ersterDeck, zweiterDeck string,
) string {
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

	angelegt := app.PartieflaecheUeber(db).PartieAusWuenschen(ctx, "magic", "anna",
		[]app.Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: ersterDeck},
			{
				Kennung: "bert", Name: "Bert", DeckID: zweiterDeck,
				ProfilID: strconv.FormatInt(profil, 10),
			},
		})
	if angelegt.Fehler != nil {
		t.Fatalf("die Partie anlegen: %v", angelegt.Fehler.AlsFehler())
	}

	return angelegt.Partie.ID
}

// dranzeilenAus liest die Liste, hüllenlos oder aus der Hülle der Kommandozeile.
func dranzeilenAus(t *testing.T, roh string, mitHuelle bool) map[string]dranzeile {
	t.Helper()

	var zeilen []dranzeile

	if mitHuelle {
		var huelle struct {
			Partien []dranzeile `json:"partien"`
		}

		if fehler := json.Unmarshal([]byte(roh), &huelle); fehler != nil {
			t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", fehler, roh)
		}

		zeilen = huelle.Partien
	} else if fehler := json.Unmarshal([]byte(roh), &zeilen); fehler != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", fehler, roh)
	}

	nach := map[string]dranzeile{}
	for _, zeile := range zeilen {
		nach[zeile.ID] = zeile
	}

	return nach
}

// **Dieselbe Auskunft an jeder Fläche — und mit dem Anhalten ändert sie sich
// überall.**
func TestDieMarkeDranIstAnJederFlaecheDieselbe(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	meine := einePartieAmProfil(t, umgebung, ersterDeck, zweiterDeck)
	zweites := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Zweites", "--json"))
	zugesehen := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck,
		"--sitzprofil=bert:"+zweites, "--json")).ID

	// listen liest die Liste von allen drei Flächen.
	listen := func() map[string]map[string]dranzeile {
		return map[string]map[string]dranzeile{
			"Kommandozeile": dranzeilenAus(t, mussKommandozeile(t, umgebung,
				"partien", "--json"), true),
			"MCP": dranzeilenAus(t, mussMCP(t, umgebung, "lesen.partien",
				map[string]any{}), false),
			"Dienst": dranzeilenAus(t, mussREST(t, an, http.MethodGet,
				"/partien", ""), false),
		}
	}

	pruefe := func(wann string, eigener string, dran bool, stand string) {
		t.Helper()

		for flaeche, zeilen := range listen() {
			if zeile := zeilen[meine]; zeile.EigenerSitz != eigener ||
				zeile.Dran != dran || zeile.Stand != stand {
				t.Errorf("%s, %s: eigener Sitz %q, dran=%v, Stand %q — erwartet %q, %v, %q",
					flaeche, wann, zeile.EigenerSitz, zeile.Dran, zeile.Stand,
					eigener, dran, stand)
			}

			// **Wer nicht am Tisch sitzt, ist nie dran** — an jeder Fläche.
			if zeile := zeilen[zugesehen]; zeile.EigenerSitz != "" || zeile.Dran {
				t.Errorf("%s, %s: bei der Partie eines anderen Profils steht eigener Sitz %q, dran=%v",
					flaeche, wann, zeile.EigenerSitz, zeile.Dran)
			}
		}
	}

	// **Der Beginn steht aus, und der gebundene Sitz hat noch einen Schritt.**
	pruefe("der Beginn steht aus", "bert", true, "laeuft")

	// **Die Textausgabe nennt es in Worten** — und nur, wo das Profil sitzt.
	text := mussKommandozeile(t, umgebung, "partien")

	for _, zeile := range strings.Split(strings.TrimSpace(text), "\n") {
		nennt := strings.Contains(zeile, ", dein Sitz: bert, dran")

		if strings.HasPrefix(zeile, meine+" ") != nennt {
			t.Errorf("die Zeile %q: dein Sitz und dran stehen %v", zeile, nennt)
		}
	}

	if !strings.Contains(text, meine+" — laeuft") {
		t.Errorf("die Textausgabe nennt die Partie nicht: %q", text)
	}

	mussKommandozeile(t, umgebung, "partie-stand", meine, "bert", "pausieren")

	pruefe("angehalten", "bert", false, "pausiert")
}
