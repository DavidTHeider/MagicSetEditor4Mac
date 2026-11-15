// Dass jeder Endpunkt den Fehler seines Anwendungsfalls durchreicht
// (#997).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2) — und
// weil `fehlerwege_test.go` sonst über die Zeilengrenze liefe.
package rest_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Jeder Endpunkt reicht den Fehler seines Anwendungsfalls durch.**
//
// # Warum das je Endpunkt geprüft wird und nicht einmal
//
// Weil jeder seinen eigenen `if fehler != nil` hat. Ein vergessener macht aus
// einem Fehlschlag ein 200 mit halb gefüllter Antwort — und das ist genau die
// Klasse „ein halber Stand bleibt zurück" aus §10.
func TestJederEndpunktReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	for _, ziel := range alleEndpunkte {
		t.Run(ziel.name, func(t *testing.T) {
			t.Parallel()

			jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
			boese := unwilliger{fehler: fmt.Errorf("die Karte: %w", domain.ErrNotFound)}

			dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
				rest.Naehte{
					Leser:     boese,
					Schreiber: boese,
					Ausfuhr:   boese,
					Abgleich:  boese,
					Bilder:    boese,
					Auskunft:  boese,
					Suche:     boese,
					Katalog:   boese, Peer: boese, Kontakte: boese,
					Sichtfreigaben:   boese,
					Tausch:           boese,
					Post:             boese,
					Verbuchung:       boese,
					Leitung:          boese,
					Links:            boese,
					Konto:            boese,
					Nachbarn:         boese,
					Grenzen:          boese,
					Schnittstellen:   boese,
					Schalter:         boese,
					Gesundheit:       boese,
					Metriken:         boese,
					Meta:             boese,
					Vertraeglichkeit: boese,
					Ruecksprung:      boese,
					Pflege:           boese,
					Ruecksetzung:     boese,
					Index:            boese,
					Recherche:        boese,
					Kosten:           boese,
					Verweise:         boese,
					Marken:           boese,
					Ablage:           boese,
					Ersetzung:        boese,
					Liste:            boese,
					Sammeldaten:      boese,
					Anlage:           boese,
					Saetze:           boese,
					Konsole:          boese,
					Werkstatt:        boese,
					Sammlung:         boese,
					Listeneinfuhr:    boese,
					Deckeinfuhr:      boese,
					Sammlungsausfuhr: boese,
					Sammlungssichten: boese,
					Kartenskripte:    boese,
					Auszaehlung:      boese,
					Parameterhilfen:  boese,
					Autoersetzungen:  boese,
					Decks:            boese,
					Turniere:         boese,
					Symbolbibliothek: boese,
					Sprachpakete:     boese,
					Symbol:           boese,
					Masse:            boese,
					Preise:           boese,
					Wortwahl:         boese,
					Woerter:          boese,
					Pack:             boese,
					Angaben:          boese,
					Probleme:         boese,
					Stile:            boese,
					Zusatz:           boese,
					Pakete:           boese,
					Nutzer:           boese,
					Token:            boese,
					Vorgaben:         boese,
					Freigaben:        boese,
					Platz:            boese,
					Deckuebernahme:   boese,
					Zahlen:           boese,
					Verlauf:          boese,
					Einstellungen:    boese,
					Austausch:        austauschprobe(boese),
					Fassungen:        boese,
					Anleger:          boese,
					Auftraege:        boese,
					Setvorschau:      boese,
					Dateiablage:      boese,
					Loescher:         boese,
					Sicherung:        boese,
					Dateien:          boese,
					Objektmarken:     boese,
					Umschlag:         boese,
					Partien:          boese,
					Partienverlauf:   boese,
					Netzweg:          boese,
					Regelfreigabe:    boese,
					Kartenbilder:     boese,
					Vorlagen:         boese,
					Probelauf:        boese,
				})

			antwort := ruf(t, dienst, ziel, gueltig)

			// **Ein Endpunkt ohne Fehlerweg antwortet trotzdem** — die
			// umgedrehte Zusage, der Grund steht am Feld.
			if ziel.ohneFehlerweg {
				if antwort.Code != ziel.erfolg {
					t.Errorf("Status %d, erwartet war %d (%s)",
						antwort.Code, ziel.erfolg, antwort.Body.String())
				}

				return
			}

			pruefeAbweisung(t, antwort, http.StatusNotFound, "nichtGefunden")
		})
	}
}

// **Jeder Endpunkt mit Kennung prüft sie, und jeder mit Rumpf prüft ihn.**
//
// # Warum aus einer Tabelle und nicht je Endpunkt von Hand
//
// Dieselbe Überlegung wie bei den vier Klassen: was von Hand steht, kann bei
// einem siebten Endpunkt fehlen. Hier steht je Endpunkt **eine** Zeile, und
// **Eine Rumpfmeldung verrät nichts über Go** (Zweitblick über Phase U).
//
// # Der Befund
//
// Der Fehler des Entzifferers ging unverändert in die Antwort. Bei einem
// Typfehler heißt der: `json: cannot unmarshal number into Go value of type
// struct { Pfad string "json:\"pfad\"" }` — das ist die Struktur dieses
// Programms in einer Antwort nach außen.
//
// **Was bleiben soll**, ist die Hilfe: welches Feld nicht stimmt. Die steht
// im Fehler des Entzifferers und im Rumpf des Aufrufers; was nicht bleiben
// soll, ist alles, was mit „Go" anfängt.
func TestEineRumpfmeldungVerraetNichtsUeberGo(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"eine Zahl statt einer Zeichenkette":   `{"pfad":42}`,
		"ein Verbund statt einer Zeichenkette": `{"pfad":{"a":1}}`,
		"eine Liste":                           `[1,2,3]`,
	}

	for name, rumpf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t),
				endpunkt{methode: http.MethodPost, weg: "/sets", rumpf: rumpf}, gueltig)

			gesagt := antwort.Body.String()
			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("Status %d (%s)", antwort.Code, gesagt)
			}

			for _, verboten := range []string{"Go value", "Go struct", "struct {", "json:"} {
				if strings.Contains(gesagt, verboten) {
					t.Errorf("die Antwort enthält %q: %s", verboten, gesagt)
				}
			}
		})
	}
}

// **Und sie sagt trotzdem, welches Feld nicht stimmt.**
//
// Ohne diesen Fall wäre auch eine Antwort grün, die jeden Rumpffehler auf
// „geht nicht" verkürzte — und der Aufrufer hätte nichts, woran er sucht.
func TestEineRumpfmeldungNenntDasFeld(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t),
		endpunkt{methode: http.MethodPost, weg: "/sets", rumpf: `{"pfad":42}`}, gueltig)

	if !strings.Contains(antwort.Body.String(), "pfad") {
		t.Errorf("die Antwort nennt das Feld nicht: %s", antwort.Body.String())
	}
}

// **Der Scryfall-Vorschlag braucht einen Namen.**
//
// Er steht in der Abfrage und nicht im Weg: ein Kartenname kann Leerzeichen
// und Schrägstriche enthalten („Fire // Ice"), und beides zerschnitte einen
// Pfad.
func TestDerScryfallVorschlagBrauchtEinenNamen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t),
		endpunkt{methode: http.MethodGet, weg: "/sets/1/karten/2/scryfall"}, gueltig)
	pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")
}

// **Und beide Scryfall-Endpunkte prüfen ihre Kennungen.**
func TestDieScryfallEndpunktePruefenIhreKennungen(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, methode, weg, rumpf string }{
		{"Vorschlag, unbrauchbares Set", http.MethodGet, "/sets/x/karten/2/scryfall?name=A", ""},
		{"Vorschlag, unbrauchbare Karte", http.MethodGet, "/sets/1/karten/x/scryfall?name=A", ""},
		{"Übernahme, unbrauchbares Set", http.MethodPost, "/sets/x/karten/2/scryfall", `{"name":"A"}`},
		{"Übernahme, unbrauchbare Karte", http.MethodPost, "/sets/1/karten/x/scryfall", `{"name":"A"}`},
		{"Übernahme, kein JSON", http.MethodPost, "/sets/1/karten/2/scryfall", `{`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t),
				endpunkt{methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf}, gueltig)
			pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")
		})
	}
}

// **Ein Vorschlag ohne Änderungen kommt als leere Liste an, nicht als `null`.**
//
// Die Falle aus §3: ein Aufrufer, der darüber schleift, bekäme sonst einen
// Fehler statt keiner Runde.
func TestEinLeererVorschlagKommtAlsListeAn(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:          williger{},
			Schreiber:      williger{},
			Ausfuhr:        williger{},
			Abgleich:       leererabgleich{},
			Bilder:         williger{},
			Auskunft:       williger{},
			Suche:          williger{},
			Katalog:        williger{},
			Grenzen:        williger{},
			Verweise:       williger{},
			Ablage:         williger{},
			Ersetzung:      williger{},
			Liste:          williger{},
			Sammeldaten:    williger{},
			Anlage:         williger{},
			Saetze:         williger{},
			Konsole:        williger{},
			Werkstatt:      williger{},
			Wortwahl:       nil,
			Woerter:        nil,
			Pack:           williger{},
			Angaben:        williger{},
			Stile:          williger{},
			Zusatz:         williger{},
			Pakete:         williger{},
			Nutzer:         williger{},
			Token:          williger{},
			Vorgaben:       williger{},
			Freigaben:      williger{},
			Deckuebernahme: williger{},
			Zahlen:         williger{},
			Verlauf:        williger{},
			Einstellungen:  williger{},
			Fassungen:      williger{},
			Anleger:        nil,
		})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodGet, weg: "/sets/1/karten/2/scryfall?name=A",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d (%s)", antwort.Code, antwort.Body.String())
	}

	for _, erwartet := range []string{`"aenderungen":[]`, `"befunde":[]`} {
		if !strings.Contains(antwort.Body.String(), erwartet) {
			t.Errorf("die Antwort enthält %s nicht: %s", erwartet, antwort.Body.String())
		}
	}
}

// **Und ein Vorschlag mit Befunden reicht sie durch.**
//
// Die Gegenprobe zum Fall darüber: ohne sie wäre auch eine Antwort grün, die
// **jede** Befundliste durch eine leere ersetzte.
func TestEinVorschlagMitBefundenReichtSieDurch(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:          williger{},
			Schreiber:      williger{},
			Ausfuhr:        williger{},
			Abgleich:       mitbefunden{},
			Bilder:         williger{},
			Auskunft:       williger{},
			Suche:          williger{},
			Katalog:        williger{},
			Verweise:       williger{},
			Ablage:         williger{},
			Ersetzung:      williger{},
			Liste:          williger{},
			Sammeldaten:    williger{},
			Anlage:         williger{},
			Saetze:         williger{},
			Konsole:        williger{},
			Werkstatt:      williger{},
			Wortwahl:       nil,
			Woerter:        nil,
			Pack:           williger{},
			Angaben:        williger{},
			Stile:          williger{},
			Zusatz:         williger{},
			Pakete:         williger{},
			Nutzer:         williger{},
			Token:          williger{},
			Vorgaben:       williger{},
			Freigaben:      williger{},
			Deckuebernahme: williger{},
			Zahlen:         williger{},
			Verlauf:        williger{},
			Einstellungen:  williger{},
			Fassungen:      williger{},
			Anleger:        nil,
		})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodGet, weg: "/sets/1/karten/2/scryfall?name=A",
	}, gueltig)

	if !strings.Contains(antwort.Body.String(), "unbekannter Wert") {
		t.Errorf("der Befund kam nicht durch: %s", antwort.Body.String())
	}
}
