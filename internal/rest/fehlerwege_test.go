package rest_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Jeder Fehler des Fachkerns bekommt seinen Status**
// (#997).
//
// # Warum die Zuordnung geprüft wird und nicht nur dasteht
//
// Sie ist die eine Stelle, an der dieser Dienst etwas **entscheidet**. Ein
// belegter Bestand als 409 hieße für den Aufrufer „lies neu" — und das liefe
// in dieselbe Sperre ([0045](../../docs/entscheidungen/0045-zwei-programme-an-einem-bestand.md)).
func TestJederFehlerBekommtSeinenStatus(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fehler error
		status int
		code   string
		wieder string // erwarteter Retry-After
	}{
		"nicht gefunden": {fmt.Errorf("die Karte: %w", domain.ErrNotFound), http.StatusNotFound, "nichtGefunden", ""},
		"Konflikt":       {fmt.Errorf("überholt: %w", domain.ErrConflict), http.StatusConflict, "konflikt", ""},
		"Validierung":    {fmt.Errorf("leer: %w", domain.ErrValidation), http.StatusBadRequest, "validierung", ""},
		"belegt":         {fmt.Errorf("warten: %w", domain.ErrBelegt), http.StatusServiceUnavailable, "belegt", "1"},
		"abgebrochen":    {fmt.Errorf("weg: %w", context.Canceled), 499, "abgebrochen", ""},
		"Frist abgelaufen": {
			fmt.Errorf("zu lange: %w", context.DeadlineExceeded), 499, "abgebrochen", "",
		},
		// **Der Nachweis reicht nicht weit genug** — Werkstück 11 von
		// Phase 4.12 (#1351). Bis dahin erzeugte ihn nur die Tokenwache,
		// die ihn selbst in einen 403 übersetzt; seit der Profilwechsel
		// eine Rolle verlangt, kommt er aus dem Dienst und läuft hier
		// durch. **Ein 500 wäre falsch**: der Aufruf ist nicht gescheitert,
		// er war nicht erlaubt.
		"Erlaubnis": {
			fmt.Errorf("wechseln: %w", domain.ErrErlaubnis),
			http.StatusForbidden, "erlaubnis", "",
		},
		"etwas Unbekanntes": {
			errors.New("SQL logic error: no such table"), http.StatusInternalServerError, "unbekannt", "",
		},
		// **Der volle Datenträger** (#1812): 507 und nicht 500 — der
		// Aufruf war richtig, es fehlt Platz. Aus dem Bestand und aus einer
		// Datei.
		// **Ein gesperrter Ort ist eine Eingabe, die nicht passt** — wie im
		// Fenster und über MCP (Nachlauf über Phase 2 von 0.9.0, Befund 7):
		// bis dahin kam er hier als 500 an, ohne den Rat, den Ort freizugeben.
		"gesperrter Ort": {
			domain.MitGrund(domain.ErrGesperrt, errors.New("is a directory"), "gib den Ort frei"),
			http.StatusBadRequest, "validierung", "",
		},
		"voller Bestand": {
			fmt.Errorf("anlegen: %w", domain.ErrVoll), http.StatusInsufficientStorage, "voll", "",
		},
		"volle Platte": {
			&fs.PathError{Op: "write", Path: "/Users/jemand/x", Err: syscall.ENOSPC},
			http.StatusInsufficientStorage, "voll", "",
		},
		// **Ein doppelter Eintrag bleibt ein 409** (ENG-05.AK-03) —
		// verträglich für jeden, der auf den Status schaut.
		"doppelt": {
			fmt.Errorf("anlegen: %w", domain.NewFieldError("name", domain.SchluesselDoppelterName, "gibt es schon")),
			http.StatusConflict, "konflikt", "",
		},
		// **Ein fremder Dienst ist kein 500** (D02, #2233): der Dienst hier
		// arbeitet, der dahinter nicht — 502, und 504, wo eine Frist griff.
		"ein Name, der sich nicht auflöst": {
			fmt.Errorf("Scryfall fragen: %w", domain.NeuerGegenstellenfehler(
				domain.ErrNameUnaufloesbar, errors.New("no such host"), "api.scryfall.com", 0)),
			http.StatusBadGateway, "netz", "",
		},
		"keine Antwort binnen der Frist": {
			domain.NeuerGegenstellenfehler(domain.ErrKeineAntwort, errors.New("timeout"), "api.scryfall.com", time.Minute),
			http.StatusGatewayTimeout, "netz", "",
		},
		"eine Antwort, die stockt": {
			domain.NeuerGegenstellenfehler(domain.ErrAntwortStockt, errors.New("timeout"), "mtgjson.com", time.Minute),
			http.StatusGatewayTimeout, "netz", "",
		},
		// **Ein Abriss ist keine Frist** (#2487): 502, wie ein Port ohne
		// Annahme.
		"eine Verbindung, die abreißt": {
			domain.NeuerGegenstellenfehler(domain.ErrVerbindungAbgerissen, errors.New("EOF"), "api.scryfall.com", 0),
			http.StatusBadGateway, "netz", "",
		},
		"ein 429 dahinter": {
			fmt.Errorf("%w: https://api.scryfall.com/cards", domain.ErrGedrosselt),
			http.StatusBadGateway, "netz", "",
		},
		// **Das eigene Kontingent ist der eigene 429** (Nachlauf über Phase 1
		// von 0.9.0, Befund 4) — dieselbe Klasse wie das Budget des Dienstes,
		// mit der echten Restzeit; bis dahin ein 500.
		"das eigene Kontingent": {
			fmt.Errorf("fragen: %w", domain.NeueKontingentgrenze(2500*time.Millisecond)),
			http.StatusTooManyRequests, "drossel", "3",
		},
		"das Kontingent ohne Zeit": {
			fmt.Errorf("fragen: %w", domain.ErrKontingent),
			http.StatusTooManyRequests, "drossel", "1",
		},
		// **Ein veralteter Dienst ist belegt, mit eigenem Satz** (M02,
		// #2346) — er endet, und nach dem Warten antwortet die neue Fassung.
		"ein veralteter Dienst": {
			&domain.ProzessVeraltet{Fassung: &domain.NeuereFassung{Unbekannt: []string{"0099_kuenftig.sql"}, Stand: "0098_sicherungsvermerke.sql"}},
			http.StatusServiceUnavailable, "belegt", "1",
		},
		"belegt gewinnt gegen den mitgebrachten Fachfehler": {
			fmt.Errorf("%w: %w", domain.ErrBelegt, domain.ErrNotFound),
			http.StatusServiceUnavailable, "belegt", "1",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienstDerScheitert(t, fall.fehler),
				endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig)

			if antwort.Code != fall.status {
				t.Errorf("Status %d, erwartet war %d (%s)",
					antwort.Code, fall.status, antwort.Body.String())
			}

			if befund := alsBefund(t, antwort); befund.Code != fall.code {
				t.Errorf("Code %q, erwartet war %q", befund.Code, fall.code)
			}

			if wieder := antwort.Header().Get("Retry-After"); wieder != fall.wieder {
				t.Errorf("Retry-After %q, erwartet war %q", wieder, fall.wieder)
			}
		})
	}
}

// **Und die rohe Meldung eines unbekannten Fehlers bleibt drinnen.**
//
// §8: rohe SQLite-Meldungen erreichen keine Oberfläche. Ein Dienst nach außen
// ist eine.
func TestEinUnbekannterFehlerVerraetNichts(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstDerScheitert(t, errors.New("no such table: cards /Users/david/geheim")),
		endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig)

	if gesagt := antwort.Body.String(); strings.Contains(gesagt, "cards") ||
		strings.Contains(gesagt, "/Users") {
		t.Errorf("die Antwort verrät den inneren Fehler: %q", gesagt)
	}
}

// **Eine Kennung, die keine ist, wird abgewiesen.**
func TestEineUnbrauchbareKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, weg := range []string{"/sets/abc/karten", "/sets/0/karten", "/sets/-1/karten"} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{methode: http.MethodGet, weg: weg}, gueltig)
			pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")
		})
	}
}

// **Ein Rumpf, der keiner ist, ebenso.**
func TestEinUnbrauchbarerRumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"kein JSON":            "das ist kein JSON",
		"leer":                 "",
		"ein unbekanntes Feld": `{"pfad":"/tmp/x","hintertuer":"ja"}`,
		"zu groß":              `{"pfad":"` + strings.Repeat("x", 1<<21) + `"}`,
	}

	for name, rumpf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t),
				endpunkt{methode: http.MethodPost, weg: "/sets", rumpf: rumpf}, gueltig)

			if antwort.Code < http.StatusBadRequest {
				t.Fatalf("der Rumpf kam durch: %d", antwort.Code)
			}

			if befund := alsBefund(t, antwort); befund.Code != "validierung" {
				t.Errorf("Code %q, erwartet war %q", befund.Code, "validierung")
			}
		})
	}
}

// **Ein Set, das es nicht gibt, ist kein Set ohne Karten.**
//
// # Warum das ein eigener Fall ist
//
// Der Leser gibt für beides eine leere Liste. Ohne die Unterscheidung bekäme
// ein Aufrufer auf eine erfundene Kennung ein fröhliches 200 mit `[]` — und
// hielte das Set für leer statt für nicht vorhanden.
func TestEinSetDasEsNichtGibtIstKeinLeeresSet(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMitLeeremSet(t),
		endpunkt{methode: http.MethodGet, weg: "/sets/4711/karten"}, gueltig)
	pruefeAbweisung(t, antwort, http.StatusNotFound, "nichtGefunden")
}

// **Ein Set, das es gibt und das leer ist, gibt eine leere Liste.**
//
// Die Gegenprobe zum Fall darüber: ohne sie wäre auch ein Dienst grün, der
// jedes leere Set für nicht vorhanden hielte.
func TestEinLeeresSetGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMitLeeremSet(t),
		endpunkt{methode: http.MethodGet, weg: "/sets/1/karten"}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d (%s)", antwort.Code, antwort.Body.String())
	}

	// **`[]` und nicht `null`** — sonst scheitert ein Aufrufer, der schleift.
	if gesagt := antwort.Body.String(); !strings.Contains(gesagt, `"karten":[]`) {
		t.Errorf("die leere Liste kommt nicht als leere Liste an: %q", gesagt)
	}
}

// **Eine Marke lässt sich nicht würfeln, wenn die Quelle versiegt.**
func TestEineVersiegteQuelleGibtKeineMarke(t *testing.T) {
	t.Parallel()

	if _, fehler := rest.MarkeAus(strings.NewReader("zu kurz")); fehler == nil {
		t.Error("aus sieben Bytes wurde eine Marke")
	}
}

// dienstDerScheitert baut einen Dienst, dessen Leser den gegebenen Fehler gibt.
func dienstDerScheitert(t *testing.T, fehler error) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:         scheiternder{fehler: fehler},
			Schreiber:     williger{},
			Ausfuhr:       williger{},
			Abgleich:      williger{},
			Bilder:        nil,
			Auskunft:      nil,
			Sammeldaten:   williger{},
			Anlage:        williger{},
			Saetze:        williger{},
			Konsole:       williger{},
			Werkstatt:     williger{},
			Wortwahl:      nil,
			Woerter:       nil,
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Kartenbilder:  williger{},
			Vorlagen:      williger{},
			Probelauf:     williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       nil,
		})
}

// dienstMitLeeremSet baut einen Dienst, dessen Set 1 keine Karten hat.
func dienstMitLeeremSet(t *testing.T) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:         leeres{},
			Schreiber:     williger{},
			Ausfuhr:       williger{},
			Abgleich:      williger{},
			Bilder:        nil,
			Auskunft:      nil,
			Sammeldaten:   williger{},
			Anlage:        williger{},
			Saetze:        williger{},
			Konsole:       williger{},
			Werkstatt:     williger{},
			Wortwahl:      nil,
			Woerter:       nil,
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Kartenbilder:  williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       nil,
		})
}

// scheiternder gibt beim Lesen der Setköpfe einen Fehler.
type scheiternder struct {
	williger

	fehler error
}

func (s scheiternder) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, s.fehler
}

// leeres kennt genau ein Set, und das hat keine Karten.
type leeres struct{ williger }

func (leeres) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

// **Und ein veralteter Dienst sagt, warum** — der allgemeine Satz für einen
// belegten Bestand riete zum Warten auf denselben Prozess (M02, #2346).
func TestEinVeralteterDienstSagtWarum(t *testing.T) {
	t.Parallel()

	fehler := &domain.ProzessVeraltet{Fassung: &domain.NeuereFassung{Unbekannt: []string{"0099_kuenftig.sql"}, Stand: "0098_sicherungsvermerke.sql"}}
	antwort := ruf(t, dienstDerScheitert(t, fehler), endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig)

	if befund := alsBefund(t, antwort); !strings.Contains(befund.Meldung, "älter als der Bestand") || !strings.Contains(befund.Meldung, "0099_kuenftig.sql") {
		t.Errorf("die Meldung: %q", befund.Meldung)
	}
}
