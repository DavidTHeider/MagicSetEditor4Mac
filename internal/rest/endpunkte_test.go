package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Vier Testklassen je Endpunkt** — die Zahl steht in
// [0039](../../docs/entscheidungen/0039-kein-dienst-auf-der-rueckschleife.md)
// als das, was ein Ja kostet, und in
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
// als Bedingung 2: wird sie einmal übergangen, fällt die Entscheidung für
// diesen Dienst wieder.
//
// # Warum sie aus **einer** Liste kommen
//
// Vier Tests je Endpunkt von Hand zu schreiben hieße, dass ein siebter
// Endpunkt sie vergessen kann. Hier steht jeder Endpunkt **einmal** in
// `alleEndpunkte`, und ein Wächter hält die Liste gegen den Quelltext — ein
// neuer Weg ohne Zeile macht die Pipeline rot.
func TestJederEndpunktKenntSeineVierKlassen(t *testing.T) {
	t.Parallel()

	for _, endpunkt := range alleEndpunkte {
		t.Run(endpunkt.name, func(t *testing.T) {
			t.Parallel()

			t.Run("gültig", func(t *testing.T) {
				t.Parallel()

				// **Der Anwendungsfall wird gezählt, nicht nur der Status.**
				//
				// Der erste Anlauf prüfte `Code < 400`. Damit wäre auch ein
				// Endpunkt grün, der ein 204 zurückgibt und den
				// Anwendungsfall gar nicht ruft — und die Zusage „geprüft
				// wird der Weg **bis** zum Anwendungsfall" wäre eine
				// Behauptung (Nachlauf zum Zweitblick über Phase U).
				gezaehlt := &zaehlender{}

				antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt, gueltig)
				if antwort.Code != endpunkt.erfolg {
					t.Errorf("ein gültiger Aufruf endete mit %d statt %d: %s",
						antwort.Code, endpunkt.erfolg, antwort.Body.String())
				}

				if gezaehlt.wieOft() != 1 {
					t.Errorf("der Anwendungsfall wurde %dmal gerufen, erwartet war einmal",
						gezaehlt.wieOft())
				}
			})

			t.Run("fehlende Marke", func(t *testing.T) {
				t.Parallel()

				antwort := ruf(t, neuerDienst(t), endpunkt, ohneMarke)
				pruefeAbweisung(t, antwort, http.StatusUnauthorized, "marke")

				if antwort.Header().Get("WWW-Authenticate") == "" {
					t.Error("ein 401 ohne WWW-Authenticate sagt nicht, womit man es versuchen soll")
				}
			})

			t.Run("fremde Herkunft", func(t *testing.T) {
				t.Parallel()

				antwort := ruf(t, neuerDienst(t), endpunkt, fremdeHerkunft)
				pruefeAbweisung(t, antwort, http.StatusForbidden, "herkunft")
			})

			t.Run("Grenze überschritten", func(t *testing.T) {
				t.Parallel()

				dienst := neuerDienst(t)

				// **Die Grenze wird wirklich erreicht, nicht angenommen.** Ein
				// Test, der sie setzte statt sie zu füllen, prüfte die
				// Einstellung und nicht die Wirkung.
				//
				// **Und die sechzig davor müssen durchkommen.** Der erste
				// Anlauf warf ihre Antworten weg: eine Drossel, die schon ab
				// dem zweiten Aufruf abwies, wäre grün gewesen — nachgestellt
				// mit `grenze = 1` (Nachlauf zum Zweitblick über Phase U).
				// **Erlaubt ist, was das Guthaben je Kostenklasse hergibt** (#1429).
				erlaubt := rest.ErlaubteRufe(domain.Vorgabegrenzen(), endpunkt.methode, endpunkt.weg)

				// **Ein Abbruch wird nie gedrosselt** (#2003) — auch nicht
				// nach mehr Rufen, als ein volles Guthaben hergäbe.
				if erlaubt < 0 {
					for range domain.Vorgabegrenzen().Principal.Burst + 1 {
						if antwort := ruf(t, dienst, endpunkt, gueltig); antwort.Code == http.StatusTooManyRequests {
							t.Fatalf("ein Abbruch wurde gedrosselt: %s", antwort.Body.String())
						}
					}

					return
				}

				for versuch := 1; versuch <= erlaubt; versuch++ {
					if antwort := ruf(t, dienst, endpunkt, gueltig); antwort.Code >= http.StatusBadRequest {
						t.Fatalf("Aufruf %d endete mit %d, erlaubt sind %d: %s",
							versuch, antwort.Code, erlaubt, antwort.Body.String())
					}
				}

				antwort := ruf(t, dienst, endpunkt, gueltig)
				pruefeAbweisung(t, antwort, http.StatusTooManyRequests, "drossel")

				if antwort.Header().Get("Retry-After") == "" {
					t.Error("ein 429 ohne Retry-After sagt nicht, wann es wieder geht")
				}
			})
		})
	}
}

// pruefeAbweisung hält Status, Code und die Gestalt der Fehlerantwort.
//
// **An einer Stelle**, weil dieselbe Zusage für alle vier Klassen gilt: ein
// Skript liest `code`, ein Mensch `meldung`, und beide sollen immer da sein.
func pruefeAbweisung(t *testing.T, antwort *httptest.ResponseRecorder, status int, code string) {
	t.Helper()

	if antwort.Code != status {
		t.Errorf("Status %d, erwartet war %d (%s)", antwort.Code, status, antwort.Body.String())
	}

	befund := alsBefund(t, antwort)
	if befund.Code != code {
		t.Errorf("Code %q, erwartet war %q", befund.Code, code)
	}

	if strings.TrimSpace(befund.Meldung) == "" {
		t.Error("die Abweisung sagt nicht, warum")
	}
}

// **Und eine Methode, die es an diesem Weg nicht gibt, wird abgewiesen.**
//
// Nicht eine der vier Klassen, aber dieselbe Familie: `GET /sets` und
// `POST /sets` sind zwei Endpunkte, und wer `DELETE /sets` ruft, soll ein 405
// bekommen und nicht den falschen Behandler.
func TestEineFremdeMethodeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)

	for _, weg := range []string{"/sets", "/sets/1/karten"} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, weg, http.NoBody)
			anfrage.Host = "127.0.0.1:7777"
			anfrage.Header.Set("Authorization", "Bearer "+probemarke)

			antwort := httptest.NewRecorder()
			dienst.Behandler().ServeHTTP(antwort, anfrage)

			if antwort.Code != http.StatusMethodNotAllowed {
				t.Errorf("Status %d, erwartet war %d", antwort.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

// **Ein Weg, den es nicht gibt, ebenso.**
func TestEinUnbekannterWegWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/gibtsnicht", http.NoBody)
	anfrage.Host = "127.0.0.1:7777"
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)

	antwort := httptest.NewRecorder()
	neuerDienst(t).Behandler().ServeHTTP(antwort, anfrage)

	if antwort.Code != http.StatusNotFound {
		t.Errorf("Status %d, erwartet war %d", antwort.Code, http.StatusNotFound)
	}
}
