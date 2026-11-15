package app

// **Die Ansage, hin und zurück** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Warum beide Seiten in einem Lauf stehen
//
// **Weil die Ansage ein Vertrag zwischen zwei Prozessen ist.** Ein Lauf, der
// nur den Text prüft, sagt nichts darüber, ob der Leser ihn versteht; einer,
// der nur den Leser prüft, füttert ihn mit einer Zeichenkette, die vielleicht
// niemand schreibt. Geprüft wird deshalb der Weg: schreiben, lesen,
// vergleichen.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was ein Dienst schreibt, liest der Posten zurück.**
func TestDieAnsageKommtAnWieSieGeschriebenWurde(t *testing.T) {
	t.Parallel()

	faelle := map[string]Dienstansage{
		"jenseits der Schleife": {
			MitTLS:        true,
			Adresse:       "0.0.0.0:54591",
			Fingerabdruck: "AB:CD:EF:01",
			Marke:         "geheim",
		},
		"auf der Schleife": {
			Adresse: "127.0.0.1:8123",
			Marke:   "auchgeheim",
		},
		"mit Endpunkt": {
			MitTLS:        true,
			Adresse:       "[::]:41234",
			Pfad:          "/mcp",
			Fingerabdruck: "12:34",
			Marke:         "geheim",
		},

		// **Mit eigenem Zeugnis** (NET-03, #1707) — Aussteller und Ende der
		// Gültigkeit kommen mit, und ein Aussteller darf „ bis " enthalten.
		"mit eigenem Zeugnis": {
			MitTLS:        true,
			Adresse:       "0.0.0.0:8443",
			Fingerabdruck: "AB:CD",
			Aussteller:    "Stelle bis zum Rand",
			GueltigBis:    time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC),
			Marke:         "geheim",
		},
	}

	for name, gesagt := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gelesen, fehler := leseAnsage(context.Background(),
				strings.NewReader(gesagt.Text()), time.Second)
			if fehler != nil {
				t.Fatalf("leseAnsage: %v", fehler)
			}

			schema := "http://"
			if gesagt.MitTLS {
				schema = "https://"
			}

			if erwartet := schema + gesagt.Adresse + gesagt.Pfad; gelesen.adresse != erwartet {
				t.Errorf("die Adresse kam als %q an statt als %q",
					gelesen.adresse, erwartet)
			}

			if gelesen.marke != gesagt.Marke {
				t.Errorf("die Marke kam als %q an", gelesen.marke)
			}

			if gelesen.fingerabdruck != gesagt.Fingerabdruck {
				t.Errorf("der Fingerabdruck kam als %q an",
					gelesen.fingerabdruck)
			}

			if gelesen.aussteller != gesagt.Aussteller || !gelesen.gueltigBis.Equal(gesagt.GueltigBis) {
				t.Errorf("das Zeugnis kam als %q bis %v an", gelesen.aussteller, gelesen.gueltigBis)
			}
		})
	}
}

// **Ohne TLS steht keine Zeile da** — und nicht eine leere.
//
// **Die Gegenprobe zum Lauf darüber**: dort ist zu sehen, dass ein leerer
// Abdruck ankommt; hier, dass er gar nicht erst geschrieben wird. Eine Zeile
// „Fingerabdruck " sähe aus wie ein Zeugnis, dessen Abdruck verlorenging.
func TestOhneZeugnisStehtKeinFingerabdruckDa(t *testing.T) {
	t.Parallel()

	text := Dienstansage{Adresse: "127.0.0.1:8123", Marke: "geheim"}.Text()

	if strings.Contains(text, "Fingerabdruck") {
		t.Errorf("die Ansage nennt einen Abdruck: %q", text)
	}

	// **Und „Marke" steht zuletzt** — darauf stützt sich der Leser.
	if zeilen := strings.Split(strings.TrimRight(text, "\n"), "\n"); !strings.HasPrefix(
		zeilen[len(zeilen)-1], "Marke ") {
		t.Errorf("die letzte Zeile ist %q", zeilen[len(zeilen)-1])
	}
}

// **Eine abgebrochene Ansage endet als Fehler, nicht als Warten.**
func TestEineHalbeAnsageIstKeine(t *testing.T) {
	t.Parallel()

	_, fehler := leseAnsage(context.Background(),
		strings.NewReader("Dienst auf http://127.0.0.1:1\n"), time.Second)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("eine halbe Ansage gab %v", fehler)
	}
}

// **Eine Zeugniszeile ohne „ bis "** nennt ihre Stelle und keine Zeit —
// eine unlesbare Zeit bleibt null, statt eine zu erfinden.
func TestEineZeugniszeileOhneEndeNenntKeineZeit(t *testing.T) {
	t.Parallel()

	for zeile, stelle := range map[string]string{
		"Haus-Stelle":            "Haus-Stelle",
		"Haus-Stelle bis morgen": "Haus-Stelle",
	} {
		aussteller, bis := zeugnisAus(zeile)
		if aussteller != stelle || !bis.IsZero() {
			t.Errorf("%q gab %q bis %v", zeile, aussteller, bis)
		}
	}
}
