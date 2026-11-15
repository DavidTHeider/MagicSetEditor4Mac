// Was die sechs neuen Wege **antworten**
// (#1107).
//
// **Neben den vier Klassen und nicht statt ihnen.** Jene prüfen, dass der Weg
// bis zum Anwendungsfall trägt und die drei Abweisungen sitzen; sie sagen
// nichts über die **Gestalt** der Antwort. Ein Endpunkt, der 201 gibt und die
// Kennung verschweigt, wäre dort grün — und hier rot.
package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Die neue Kennung kommt zurück** — sonst weiß niemand, welche Karte er
// gerade angelegt hat.
func TestKarteAnlegenNenntDieKennung(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/karten", erfolg: http.StatusCreated,
	}, gueltig)

	var antwort struct {
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal(rumpf.Body.Bytes(), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, rumpf.Body.String())
	}

	if antwort.KarteID != "7" {
		t.Errorf("die Kennung lautet %q statt \"7\"", antwort.KarteID)
	}
}

// **Und die der Kopie ist eine andere.**
func TestKarteKopierenNenntDieNeueKennung(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/karten/2/kopie", erfolg: http.StatusCreated,
	}, gueltig)

	var antwort struct {
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal(rumpf.Body.Bytes(), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, rumpf.Body.String())
	}

	if antwort.KarteID != "8" {
		t.Errorf("die Kennung lautet %q statt \"8\"", antwort.KarteID)
	}
}

// **Ein Treffer trägt Karte, Set, Nummer, Feld und Wert.**
func TestSucheTraegtDieFundstelle(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/suche?text=Feuerball", erfolg: http.StatusOK,
	}, gueltig)

	var antwort struct {
		Treffer []struct {
			KarteID string `json:"karteID"`
			SetID   string `json:"setID"`
			Nummer  int    `json:"nummer"`
			Feld    string `json:"feld"`
			Wert    string `json:"wert"`
		} `json:"treffer"`
		Gekuerzt bool `json:"gekuerzt"`
	}

	if err := json.Unmarshal(rumpf.Body.Bytes(), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, rumpf.Body.String())
	}

	if len(antwort.Treffer) != 1 {
		t.Fatalf("%d Treffer statt einem: %s", len(antwort.Treffer), rumpf.Body.String())
	}

	treffer := antwort.Treffer[0]
	if treffer.KarteID != "2" || treffer.SetID != "1" ||
		treffer.Nummer != 1 || treffer.Feld != "name" || treffer.Wert != "Feuerball" {
		t.Errorf("die Fundstelle kam verstümmelt an: %+v", treffer)
	}
}

// **Ohne Suchtext wird abgewiesen**, statt den ganzen Bestand zu treffen.
func TestSucheOhneTextWeistAb(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/suche", erfolg: http.StatusBadRequest,
	}, gueltig)

	if rumpf.Code != http.StatusBadRequest {
		t.Errorf("eine Suche ohne Text endete mit %d: %s", rumpf.Code, rumpf.Body.String())
	}
}

// **Eine unbrauchbare Kennung im Rumpf klagt wie eine im Weg.**
//
// Zwei Meldungsformen für dieselbe Sache zwängen jeden Aufrufer, beide zu
// kennen.
func TestTauschWeistUnbrauchbareKennungAb(t *testing.T) {
	t.Parallel()

	for name, koerper := range map[string]string{
		"keine Zahl": `{"eine":"x","andere":"3"}`,
		"null":       `{"eine":"2","andere":"0"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rumpf := ruf(t, neuerDienst(t), endpunkt{
				methode: http.MethodPost, weg: "/sets/1/karten/tausch",
				rumpf: koerper, erfolg: http.StatusBadRequest,
			}, gueltig)

			if rumpf.Code != http.StatusBadRequest {
				t.Errorf("%s endete mit %d: %s", name, rumpf.Code, rumpf.Body.String())
			}
		})
	}
}

// **Die Feldnamen kommen als Liste**, nie als `null`.
func TestFelderNenntDieNamen(t *testing.T) {
	t.Parallel()

	rumpf := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/sets/1/felder", erfolg: http.StatusOK,
	}, gueltig)

	var antwort struct {
		Felder []string `json:"felder"`
	}

	if err := json.Unmarshal(rumpf.Body.Bytes(), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, rumpf.Body.String())
	}

	if len(antwort.Felder) != 2 || antwort.Felder[0] != "name" {
		t.Errorf("die Feldnamen kamen als %v", antwort.Felder)
	}
}

// **Und ein Set ohne Felder gibt `[]`, nicht `null`.**
//
// Der Kommentar am Endpunkt sagt das zu; ohne diesen Fall bliebe es eine
// Behauptung — `williger` gibt zwei Namen und betritt den Weg nie, auf dem der
// Speicher `nil` liefert (§10, „Ein Wächter, der die halbe Zusage hält").
func TestFelderGibtLeereListeStattNull(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:         felderlos{},
			Schreiber:     williger{},
			Ausfuhr:       williger{},
			Abgleich:      williger{},
			Bilder:        williger{},
			Auskunft:      williger{},
			Sammeldaten:   williger{},
			Anlage:        williger{},
			Saetze:        williger{},
			Konsole:       williger{},
			Werkstatt:     williger{},
			Wortwahl:      williger{},
			Woerter:       williger{},
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       williger{},
		})

	rumpf := ruf(t, dienst, endpunkt{
		methode: http.MethodGet, weg: "/sets/1/felder", erfolg: http.StatusOK,
	}, gueltig)

	if strings.Contains(rumpf.Body.String(), "null") {
		t.Errorf("aus keinem Feld wurde `null`: %s", rumpf.Body.String())
	}
}

// felderlos kennt ein Set, das keine Felder hat — und gibt dafür `nil`.
type felderlos struct{ williger }

func (felderlos) Felder(context.Context, int64) ([]string, error) { return nil, nil }

// **Eine unbrauchbare Kennung im Weg wird abgewiesen** — an jedem der sechs
// neuen Wege.
//
// **Tabellengetrieben und nicht je Weg einzeln**: es ist dieselbe Zusage, und
// eine Wache, die an fünf Wegen sitzt und am sechsten fehlt, fiele sonst nicht
// auf.
func TestDieNeuenWegeWeisenUnbrauchbareKennungenAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Karte anlegen", methode: http.MethodPost, weg: "/sets/x/karten"},
		{name: "Karte kopieren", methode: http.MethodPost, weg: "/sets/1/karten/x/kopie"},
		{name: "Karte kopieren, Set", methode: http.MethodPost, weg: "/sets/x/karten/2/kopie"},
		{name: "Karte löschen", methode: http.MethodDelete, weg: "/sets/1/karten/x"},
		{name: "Karten tauschen", methode: http.MethodPost, weg: "/sets/x/karten/tausch"},
		{name: "Felder lesen", methode: http.MethodGet, weg: "/sets/x/felder"},
		{name: "Verlauf lesen", methode: http.MethodGet, weg: "/sets/x/verlauf"},
		{name: "Suche mit Set", methode: http.MethodGet, weg: "/suche?text=a&set=x"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Ein unlesbarer Rumpf beim Tausch wird abgewiesen.**
func TestTauschWeistUnlesbarenRumpfAb(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/karten/tausch", rumpf: `{`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
	}
}

// **Mit `set` sucht sie eingeschränkt** — und nicht nur ohne.
func TestSucheNimmtEinSetEntgegen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/suche?text=Feuerball&set=1&feld=name",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Errorf("endete mit %d statt 200: %s", antwort.Code, antwort.Body.String())
	}
}

// **Die drei Setangabenwege weisen eine unbrauchbare Kennung ab** — und der
// Rumpf des Setzens wird gelesen.
func TestDieSetangabenwegeWeisenAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Angaben lesen", methode: http.MethodGet, weg: "/sets/x/angaben"},
		{
			name: "Angabe setzen", methode: http.MethodPut, weg: "/sets/x/angaben/title",
			rumpf: `{"wert":"a"}`,
		},
		{
			name: "Angabe ohne Rumpf", methode: http.MethodPut, weg: "/sets/1/angaben/title",
			rumpf: `{`,
		},
		{name: "Vorbehalte lesen", methode: http.MethodGet, weg: "/sets/x/vorbehalte"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Und leere Angaben kommen als `[]`, nicht als `null`.**
//
// Der Kommentar an `ohneNil` sagt das zu; `williger` gibt zwei Felder und
// betritt den Weg nie, auf dem der Anwendungsfall `nil` liefert (§10).
func TestAngabenGebenLeereListenStattNull(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser: williger{}, Schreiber: williger{}, Ausfuhr: williger{},
			Abgleich: williger{}, Bilder: williger{}, Auskunft: williger{},
			Sammeldaten: williger{}, Anlage: williger{}, Saetze: williger{},
			Konsole: williger{}, Wortwahl: williger{}, Woerter: williger{},
			Pack: williger{}, Anleger: williger{}, Angaben: angabenlos{},
		})

	for name, weg := range map[string]string{
		"Angaben":    "/sets/1/angaben",
		"Vorbehalte": "/sets/1/vorbehalte",
		"Beispiele":  "/beispielsets",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: weg}, gueltig)
			if strings.Contains(antwort.Body.String(), "null") {
				t.Errorf("aus nichts wurde `null`: %s", antwort.Body.String())
			}
		})
	}
}

// angabenlos gibt für alles `nil` und keinen Fehler.
type angabenlos struct{ williger }

func (angabenlos) Angaben(context.Context, int64) (
	[]rest.Feldbeschreibung, []rest.Setwert, error,
) {
	return nil, nil, nil
}

func (angabenlos) Vorbehalte(context.Context, int64) ([]rest.Vorbehalt, error) {
	return nil, nil
}

func (angabenlos) Beispielsets() ([]rest.Beispielset, error) { return nil, nil }
