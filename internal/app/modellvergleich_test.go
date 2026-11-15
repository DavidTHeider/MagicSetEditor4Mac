//go:build fmlauf

// **Die Textmodelle an denselben Fragen** — AI-02.AK-03
// (#1743):
// „Texttests prüfen Formattreue und Kartenfeldvorschläge". Verglichen werden
// das Gerätemodell dieses Macs und die kleinen Ollama-Modelle aus AI-02.AK-01.
//
// # Warum hier
//
// **Weil hier beide Nähte stehen** — `applefm` und `ollama` — und die
// Anweisung, die die Anwendung wirklich schickt ([app.KIKartenanweisung]).
// Eine Kopie davon im Lauf wäre ein Vergleich mit etwas anderem. Gefragt
// wird über dieselben Klienten, die `geraetequelle` und `ollamaquelle`
// einhüllen: System, Auftrag, Schema — nichts davor und nichts danach.
//
// # Was er misst und was nicht
//
// **Er zählt, was messbar ist**: ob gültiges JSON mit dem verlangten Feld
// kam (Formattreue), ob die Sprache gehalten wurde, ob die Antwort leer,
// abgelehnt oder abgeschnitten war, wie lange sie brauchte und ob zwei
// Runden dasselbe gaben. **Ob ein Vorschlag passt, entscheidet er nicht** —
// die Spalte dafür füllt, wer bewertet, und der Beleg nennt ihn.
//
// **Und keine Ressourcen**: Ladezeit und Speicher misst
// `scripts/modellmessung.mjs` an der Laufzeit, Modell für Modell allein.
//
// # Warum hinter `fmlauf`
//
// **Weil nur dieser Tag den Riegel vor `fm` öffnet** (`echtlauf_an.go`) —
// ein Testlauf startet das Gerätemodell sonst nicht. Gelintet wird er damit
// in jeder Pipeline; gefahren nie: `make ci` fährt unter `fmlauf` nur
// `./internal/applefm/`.
//
// # Wie er gefahren wird
//
//	ollama serve
//	go test -count=1 -tags fmlauf -run TestDieTextmodelleImVergleich -v ./internal/app/
//
// Die Modelle stehen in [textmodelle]; sie werden **von Hand** gezogen,
// nicht von hier. Fehlt das Gerätemodell
// oder eine Laufzeit, ist der Lauf rot mit BLOCKED und nicht übersprungen.
package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// testsatz ist `testdaten/ki/testsatz.json` — dieselbe Datei, die
// `internal/applefm/fmguete_test.go` liest.
type testsatz struct {
	Fassung string `json:"fassung"`
	Karten  []struct {
		Typ  string `json:"typ"`
		Text string `json:"text"`
	} `json:"karten"`
	Auftraege   map[string]map[string]string `json:"auftraege"`
	Kartebisher map[string]string            `json:"kartebisher"`
}

// frage ist eine Zeile des Vergleichs.
type frage struct {
	karte         int
	sprache, feld string
	auftrag       string
	schema        map[string]any
}

// ergebnis ist, was ein Modell auf eine Frage gab.
type ergebnis struct {
	wert  string
	art   string
	dauer time.Duration
}

// Die Ausgänge einer Frage.
const (
	artFertig        = "fertig"
	artSprache       = "Sprache verfehlt"
	artLeer          = "leer"
	artKaputt        = "kein gültiges JSON"
	artAbgelehnt     = "abgelehnt"
	artAbgeschnitten = "abgeschnitten"
	artFehler        = "Fehler"
)

// befrager ist ein Modell samt der Hülle um seinen Klienten.
type befrager struct {
	name  string
	frage func(ctx context.Context, auftrag string, schema map[string]any) (inhalt, art string, fehler error)
}

// runden ist, wie oft der ganze Satz je Modell läuft — **zwei**, damit
// Stabilität und Gleichlauf zu sehen sind.
const runden = 2

func TestDieTextmodelleImVergleich(t *testing.T) {
	ctx := t.Context()
	satz := testsatzLesen(t)
	fragen := fragenAus(satz)

	alle := append([]befrager{geraetemodell(t)}, ollamamodelle(t)...)
	messungen := make([]messung, 0, len(alle))

	for _, einer := range alle {
		messungen = append(messungen, befrage(ctx, t, einer, fragen))
	}

	schreibeVergleich(t, satz, fragen, messungen)
}

// testsatzLesen liest die gemeinsame Datei.
func testsatzLesen(t *testing.T) testsatz {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "testdaten", "ki", "testsatz.json"))
	if fehler != nil {
		t.Fatalf("der Testsatz: %v", fehler)
	}

	var satz testsatz
	if fehler := json.Unmarshal(roh, &satz); fehler != nil {
		t.Fatalf("der Testsatz: %v", fehler)
	}

	return satz
}

// fragenAus baut je Karte, Sprache und Feld eine Frage — Auftrag und Schema
// in der Gestalt des Formulars.
func fragenAus(satz testsatz) []frage {
	fragen := make([]frage, 0, len(satz.Karten)*4)

	for nummer, karte := range satz.Karten {
		for _, sprache := range []string{"de", "en"} {
			for _, feld := range []string{"name", "flavor text"} {
				fragen = append(fragen, frage{
					karte: nummer + 1, sprache: sprache, feld: feld,
					auftrag: satz.Auftraege[sprache][feld] + "\n\n" + satz.Kartebisher[sprache] +
						"\ntype: " + karte.Typ + "\nrule text: " + karte.Text,
					schema: map[string]any{
						"type": "object", "required": []string{feld},
						"properties": map[string]any{feld: map[string]any{"type": "string"}},
					},
				})
			}
		}
	}

	return fragen
}

// geraetemodell startet `fm serve` und fragt darüber.
func geraetemodell(t *testing.T) befrager {
	t.Helper()

	ctx := t.Context()
	if einrichtung := applefm.Erkenne(ctx, applefm.Vorhanden, applefm.Exitcode); einrichtung != applefm.EinrichtungBereit {
		t.Fatalf("BLOCKED: das Gerätemodell ist auf diesem Rechner nicht bereit (%s)", einrichtung)
	}

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	dienst := applefm.NeuerDienst(ordner, applefm.Prozessstarter)
	t.Cleanup(func() { _ = dienst.Beende() }) //nolint:errcheck // der Lauf ist vorbei

	klient, fehler := dienst.Klient(ctx)
	if fehler != nil {
		t.Fatalf("BLOCKED: fm serve startet nicht: %v", fehler)
	}

	return befrager{name: "Gerätemodell (fm)", frage: func(
		ctx context.Context, auftrag string, schema map[string]any,
	) (string, string, error) {
		antwort, fehler := klient.Rede(ctx, applefm.Frage{
			System: app.KIKartenanweisung, Auftrag: auftrag, Schema: schema,
		})
		if fehler != nil {
			return "", artFehler, fehler
		}

		if antwort.Ausgang == applefm.AusgangAbgelehnt {
			return antwort.Naeheres, artAbgelehnt, nil
		}

		return antwort.Inhalt, "", nil
	}}
}

// textmodelle sind die zwei aus AI-02.AK-01.
//
// **Fest und nicht über die Umgebung**: `internal/werkzeuge` hält jede
// Umgebungsvariable, die eine Testdatei in `internal/app` liest, gegen den
// Plan der Fremdläufe — und ein Eintrag dort änderte dessen Beleg, ohne dass
// ein Fremdlauf etwas davon hätte.
var textmodelle = []string{"qwen3:0.6b", "gemma3:1b"}

// ollamamodelle fragt über den Klienten, den auch `ollamaquelle` einhüllt.
func ollamamodelle(t *testing.T) []befrager {
	t.Helper()

	klient := ollama.NeuerKlient(ollama.Vorgabeadresse, netz.NeuerKlient(netz.Fristen{
		Verbindung: 5 * time.Second, Kopf: 5 * time.Minute, Stillstand: 5 * time.Minute,
	}))

	fassung, fehler := klient.Fassung(t.Context())
	if fehler != nil {
		t.Fatalf("BLOCKED: keine Ollama-Laufzeit unter %s: %v", ollama.Vorgabeadresse, fehler)
	}

	t.Logf("Ollama %s", fassung)

	modelle := make([]befrager, 0, len(textmodelle))

	for _, alias := range textmodelle {
		modelle = append(modelle, befrager{name: alias, frage: func(
			ctx context.Context, auftrag string, schema map[string]any,
		) (string, string, error) {
			antwort, fehler := klient.Rede(ctx, ollama.Frage{
				Alias: alias, System: app.KIKartenanweisung, Auftrag: auftrag, Schema: schema,
			})

			switch {
			case fehler != nil:
				return "", artFehler, fehler
			case antwort.Abgeschnitten():
				return antwort.Inhalt, artAbgeschnitten, nil
			}

			return antwort.Inhalt, "", nil
		}})
	}

	return modelle
}

// befrage stellt einem Modell alle Fragen, [runden] Mal — nach einer
// Aufwärmfrage, deren Dauer als Kaltstart zählt.
func befrage(ctx context.Context, t *testing.T, einer befrager, fragen []frage) messung {
	t.Helper()

	anfang := time.Now()
	if _, _, fehler := einer.frage(ctx, fragen[0].auftrag, fragen[0].schema); fehler != nil {
		t.Logf("die Aufwärmfrage an %s: %v", einer.name, fehler)
	}

	gemessen := messung{name: einer.name, kalt: time.Since(anfang)}

	for runde := range runden {
		gemessen.runden = append(gemessen.runden, make([]ergebnis, 0, len(fragen)))

		for _, eine := range fragen {
			gemessen.runden[runde] = append(gemessen.runden[runde], einordnen(ctx, einer, eine))
		}
	}

	return gemessen
}

// einordnen stellt eine Frage und ordnet die Antwort ein.
func einordnen(ctx context.Context, einer befrager, eine frage) ergebnis {
	anfang := time.Now()
	inhalt, art, fehler := einer.frage(ctx, eine.auftrag, eine.schema)
	dauer := time.Since(anfang)

	if fehler != nil {
		return ergebnis{wert: fehler.Error(), art: artFehler, dauer: dauer}
	}

	if art != "" {
		return ergebnis{wert: inhalt, art: art, dauer: dauer}
	}

	var felder map[string]any
	if json.Unmarshal([]byte(inhalt), &felder) != nil {
		return ergebnis{wert: inhalt, art: artKaputt, dauer: dauer}
	}

	wert, _ := felder[eine.feld].(string) //nolint:errcheck // kein Text heißt leer
	if wert = strings.TrimSpace(wert); wert == "" {
		return ergebnis{wert: inhalt, art: artLeer, dauer: dauer}
	}

	if !haeltSprache(wert, eine.sprache) {
		return ergebnis{wert: wert, art: artSprache, dauer: dauer}
	}

	return ergebnis{wert: wert, art: artFertig, dauer: dauer}
}

// haeltSprache ist dieselbe grobe Probe wie in `fmguete_test.go`: ein
// deutscher Text ohne deutsches Merkmal oder ein englischer mit Umlauten
// fällt auf. **Grob** — die Spalte „passt?" ist die eigentliche Prüfung.
func haeltSprache(text, sprache string) bool {
	klein := " " + strings.ToLower(text) + " "
	deutsch := strings.ContainsAny(text, "äöüÄÖÜß") || strings.Contains(klein, " der ") ||
		strings.Contains(klein, " die ") || strings.Contains(klein, " das ") || strings.Contains(klein, " und ")

	if sprache == "en" {
		return !strings.ContainsAny(text, "äöüÄÖÜß")
	}

	return deutsch || utf8.RuneCountInString(text) < 25
}
