package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Ohne Marke kommt niemand durch** (#997).
//
// # Warum `127.0.0.1` nicht genügt
//
// Auf einem Mehrbenutzerrechner erreicht **jeder** lokale Prozess die
// Rückschleife, und ein Browser auf demselben Gerät auch. Das ist der
// Unterschied zwischen stdio und einem Port, und er ist der Grund, warum
// [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md) Marken
// überhaupt als Preis eines Dienstes genannt hat.
func TestOhneMarkeKommtNiemandDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*http.Request){
		"gar keine Kopfzeile":     func(*http.Request) {},
		"leere Kopfzeile":         func(r *http.Request) { r.Header.Set("Authorization", "") },
		"nur das Wort Bearer":     func(r *http.Request) { r.Header.Set("Authorization", "Bearer") },
		"eine fremde Marke":       func(r *http.Request) { r.Header.Set("Authorization", "Bearer fremd") },
		"die richtige, halbiert":  func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+probemarke[:8]) },
		"ein anderes Verfahren":   func(r *http.Request) { r.Header.Set("Authorization", "Basic "+probemarke) },
		"die Marke ohne Vorspann": func(r *http.Request) { r.Header.Set("Authorization", probemarke) },
	}

	for name, setzen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
			setzen(anfrage)

			if rest.MarkeStimmt(probemarke, anfrage) {
				t.Error("die Anfrage kam durch")
			}
		})
	}
}

// **Und mit der richtigen kommt man durch.**
//
// Ohne diesen Fall stünde die Prüfung auch dann grün da, wenn sie jede Anfrage
// abwiese (§10).
func TestMitDerRichtigenMarkeKommtManDurch(t *testing.T) {
	t.Parallel()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)

	if !rest.MarkeStimmt(probemarke, anfrage) {
		t.Error("die richtige Marke wurde abgewiesen")
	}
}

// **Eine neue Marke ist lang genug und jedes Mal eine andere.**
//
// # Was „lang genug" heißt
//
// 32 Byte aus `crypto/rand`, hexadezimal geschrieben. Der Dienst drosselt, aber
// eine Marke, die sich erraten lässt, macht die Drossel zur einzigen Hürde —
// und die ist gegen Raten nicht gebaut.
func TestEineNeueMarkeIstLangUndEinmalig(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for range 100 {
		marke, fehler := rest.NeueMarke()
		if fehler != nil {
			t.Fatalf("Marke erzeugen: %v", fehler)
		}

		if len(marke) < 64 {
			t.Fatalf("die Marke ist %d Zeichen lang: %q", len(marke), marke)
		}

		if gesehen[marke] {
			t.Fatalf("dieselbe Marke zweimal: %q", marke)
		}

		gesehen[marke] = true
	}
}

// **Der Vergleich läuft in gleichbleibender Zeit.**
//
// # Was dieser Test prüft und was nicht
//
// Er prüft, dass `MarkeStimmt` `subtle.ConstantTimeCompare` **benutzt** — über
// den Quelltext, nicht über eine Zeitmessung. Eine Messung wäre auf einer
// Maschine unter Last nicht verlässlich, und eine unzuverlässige Messung ist
// als Zusage schlechter als keine.
//
// **Nicht geprüft** ist damit, dass der Vergleich wirklich gleich lange
// braucht. Das ist eine Eigenschaft der Standardbibliothek, und sie zu prüfen
// hieße, sie zu misstrauen — ohne einen Fall, der das begründet.
func TestDerMarkenvergleichIstZeitkonstant(t *testing.T) {
	t.Parallel()

	quelle := lies(t, "marke.go")
	if !strings.Contains(quelle, "subtle.ConstantTimeCompare") {
		t.Error("MarkeStimmt vergleicht nicht zeitkonstant")
	}
}
