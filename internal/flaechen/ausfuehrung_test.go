// **Die dritte Ausführung über jede Fläche** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum das hier steht und nicht in `sammlung_test.go`
//
// **Weil nebenan der alte Klient steht.** `vollesExemplar` schickt `foil: true`
// und kein `finish` — genau so, wie ein Gegenüber es tut, das diese Fassung
// nicht kennt. Dass das weiter trägt, ist die eine Zusage; dass `etched`
// ankommt, ist die andere, und zwei Zusagen in einem Lauf verdecken einander.
//
// # Was der Lauf nicht hält
//
// **Das Fenster.** Es hat keine Fläche ohne Fenster (§16, benannte Ausnahme);
// seine Naht steht in `internal/app/exemplarausfuehrung_test.go`.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Eine geätzte Karte kommt über jede Fläche als `etched` zurück.**
//
// **Und der alte Wahrheitswert sagt dazu `false`**: `etched` ist eine eigene
// Ausführung und nicht der Gleichstand eines Foils. Wer sie als Foil führte,
// rechnete sie mit dem Foilpreis ab.
func TestEineGeaetzteKarteUeberJedeFlaeche(t *testing.T) {
	for name, fall := range map[string]struct {
		aufnehmen func(t *testing.T, umgebung []string)
		lesen     func(t *testing.T, umgebung []string) flaechenexemplar
	}{
		"Kommandozeile": {
			aufnehmen: func(t *testing.T, umgebung []string) {
				t.Helper()

				mussKommandozeile(t, umgebung, "exemplar", probedruck, "1",
					"played", "de", "etched")
			},
			lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
				t.Helper()

				return ersteAusHuelle(t,
					mussKommandozeile(t, umgebung, "--json", "sammlung"))
			},
		},
		"MCP": {
			aufnehmen: func(t *testing.T, umgebung []string) {
				t.Helper()

				werte := vollesExemplar()
				delete(werte, "foil")
				werte["finish"] = "etched"

				if text, istFehler := mcpRuf(t, umgebung, "schreiben.exemplar",
					werte); istFehler {
					t.Fatalf("schreiben.exemplar: %s", text)
				}
			},
			lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
				t.Helper()

				text, istFehler := mcpRuf(t, umgebung, "lesen.sammlung",
					map[string]any{})
				if istFehler {
					t.Fatalf("lesen.sammlung: %s", text)
				}

				return ersteAusHuelle(t, text)
			},
		},
		"Dienst": {
			aufnehmen: func(t *testing.T, umgebung []string) {
				t.Helper()

				werte := vollesExemplar()
				delete(werte, "foil")
				werte["finish"] = "etched"

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodPost, "/sammlung",
					alsRumpf(t, werte))
				if status != http.StatusCreated {
					t.Fatalf("anlegen: %d (%s)", status, rumpf)
				}
			},
			lesen: func(t *testing.T, umgebung []string) flaechenexemplar {
				t.Helper()

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodGet, "/sammlung", "")
				if status != http.StatusOK {
					t.Fatalf("lesen: %d (%s)", status, rumpf)
				}

				return ersteAusHuelle(t, rumpf)
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := eigenesZuhause(t)

			fall.aufnehmen(t, umgebung)

			gelesen := fall.lesen(t, umgebung)

			if gelesen.Finish != "etched" {
				t.Errorf("die Fläche gibt %q", gelesen.Finish)
			}

			if gelesen.Foil {
				t.Error("die Fläche hält die geätzte Karte für ein Foil")
			}
		})
	}
}

// **Eine Ausführung, die es nicht gibt, wird an jeder Fläche abgewiesen.**
//
// **Und nicht still zu `nonfoil` gemacht.** Die Kommandozeile tat genau das
// bis T2: alles, was nicht `foil` hieß, war nonfoil — ein Tippfehler wurde zu
// einer Angabe, die niemand mehr wiederfand (§16, Regel 6).
func TestEineErfundeneAusfuehrungWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	for name, fall := range map[string]func(t *testing.T, umgebung []string) string{
		"Kommandozeile": func(t *testing.T, umgebung []string) string {
			t.Helper()

			aus, klage, code := kommandozeile(t, umgebung, "exemplar",
				probedruck, "1", "played", "de", "glitzernd")
			if code == 0 {
				t.Fatalf("die erfundene Ausführung kam durch: %s", aus)
			}

			return klage
		},
		"MCP": func(t *testing.T, umgebung []string) string {
			t.Helper()

			werte := vollesExemplar()
			werte["finish"] = "glitzernd"

			text, istFehler := mcpRuf(t, umgebung, "schreiben.exemplar", werte)
			if !istFehler {
				t.Fatalf("die erfundene Ausführung kam durch: %s", text)
			}

			return text
		},
		"Dienst": func(t *testing.T, umgebung []string) string {
			t.Helper()

			werte := vollesExemplar()
			werte["finish"] = "glitzernd"

			status, rumpf := starteDienst(t, umgebung).ruf(t, http.MethodPost,
				"/sammlung", alsRumpf(t, werte))
			if status == http.StatusCreated {
				t.Fatalf("die erfundene Ausführung kam durch: %s", rumpf)
			}

			return rumpf
		},
	} {
		t.Run(name, func(t *testing.T) {
			meldung := fall(t, eigenesZuhause(t))

			// **Die Meldung nennt das erfundene Wort.** Sonst steht dort „die
			// Angabe ist ungültig", und niemand weiß, welche.
			if !strings.Contains(meldung, "glitzernd") {
				t.Errorf("die Meldung ist %q", meldung)
			}
		})
	}
}
