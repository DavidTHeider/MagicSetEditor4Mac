package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die drei Listenwerkzeuge** — Werkstück 3 von Phase 3.7
// (#1168).

// listenprobe ist ein Doppel der Naht.
type listenprobe struct {
	fehler      error
	geschrieben []domain.Spaltenwahl
	geordnet    *domain.Kartenordnung

	// gefragteAnfrage ist, wonach `KartenFiltern` zuletzt gefragt wurde.
	gefragteAnfrage string
}

func (l *listenprobe) KartenFiltern(_ context.Context, _ int64, anfrage string) ([]int64, error) {
	l.gefragteAnfrage = anfrage

	if l.fehler != nil {
		return nil, l.fehler
	}

	return []int64{7, 9}, nil
}

func (l *listenprobe) Listenaufbau(context.Context, int64) (
	domain.Kartenlistenaufbau, error,
) {
	if l.fehler != nil {
		return domain.Kartenlistenaufbau{}, l.fehler
	}

	return domain.Kartenlistenaufbau{
		Spalten: []domain.Listenspalte{
			{Feld: "name", Aufschrift: "Name", Sichtbar: true},
			{Feld: "rarity", Sichtbar: false},
		},
		Zeilen: []domain.Listenzeile{
			{KarteID: 7, Nummer: 0, Werte: map[string]string{"name": "Feuerball"}},
		},
		Ordnung: domain.Kartenordnung{Feld: "name", Aufsteigend: true},
	}, nil
}

func (l *listenprobe) ListenspaltenSetzen(
	_ context.Context, _ int64, wahl []domain.Spaltenwahl,
) error {
	if l.fehler != nil {
		return l.fehler
	}

	l.geschrieben = wahl

	return nil
}

func (l *listenprobe) ListensortierungSetzen(
	_ context.Context, _ int64, ordnung domain.Kartenordnung,
) error {
	if l.fehler != nil {
		return l.fehler
	}

	l.geordnet = &ordnung

	return nil
}

// rufeListenwerkzeug ruft eines der drei über die **echte** Werkzeugliste.
func rufeListenwerkzeug(
	t *testing.T, liste mcp.Fremdliste, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{},
		Liste:     liste,
		Angaben:   neueAngabenprobe(),
		Stile:     neueStilprobe(),
		Bestand:   neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("es gibt kein Werkzeug %q", name)

	return "", nil
}

// **Das lesende gibt Spalten, Zeilen und die Ordnung.**
func TestDasListenwerkzeugGibtDieGanzeGestalt(t *testing.T) {
	t.Parallel()

	text, fehler := rufeListenwerkzeug(t, &listenprobe{}, "lesen.kartenliste",
		`{"setID":1}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{
		`"feld": "name"`, `"aufschrift": "Name"`, `"sichtbar": true`,
		"Feuerball", `"sortierfeld": "name"`, `"aufsteigend": true`,
	} {
		if !strings.Contains(text, teil) {
			t.Errorf("die Antwort führt %q nicht: %s", teil, text)
		}
	}
}

// **Die Reihenfolge der Liste ist die Reihenfolge der Spalten.**
func TestDasSpaltenwerkzeugNimmtDieReihenfolgeDerListe(t *testing.T) {
	t.Parallel()

	probe := &listenprobe{}

	if _, fehler := rufeListenwerkzeug(t, probe, "schreiben.kartenliste.spalten",
		`{"setID":1,"spalten":[{"feld":"rarity","sichtbar":true},
		                       {"feld":"name","sichtbar":false}]}`); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if len(probe.geschrieben) != 2 ||
		probe.geschrieben[0].Feld != "rarity" || probe.geschrieben[0].Stelle != 1 ||
		probe.geschrieben[1].Stelle != 2 {
		t.Errorf("geschrieben wurde %+v", probe.geschrieben)
	}
}

// **Ein leeres Feld hebt die Ordnung auf.**
func TestDasSortierwerkzeugKenntDasLeereFeld(t *testing.T) {
	t.Parallel()

	probe := &listenprobe{}

	if _, fehler := rufeListenwerkzeug(t, probe, "schreiben.kartenliste.sortierung",
		`{"setID":1,"feld":"","aufsteigend":true}`); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.geordnet == nil || probe.geordnet.Feld != "" ||
		!probe.geordnet.Aufsteigend {
		t.Errorf("geordnet wurde %+v", probe.geordnet)
	}
}

// **Jede Wache greift, und zwar an jedem der drei Werkzeuge.**
func TestDieListenwerkzeugeWeisenUnbrauchbareArgumenteAb(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, werkzeug, argumente string }{
		{"lesen: kein JSON", "lesen.kartenliste", `{`},
		{"lesen: Set ist null", "lesen.kartenliste", `{"setID":0}`},
		{"spalten: kein JSON", "schreiben.kartenliste.spalten", `{`},
		{"spalten: Set ist null", "schreiben.kartenliste.spalten", `{"setID":0}`},
		{"sortierung: kein JSON", "schreiben.kartenliste.sortierung", `{`},
		{"sortierung: Set ist null", "schreiben.kartenliste.sortierung", `{"setID":0}`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeListenwerkzeug(
				t, &listenprobe{}, fall.werkzeug, fall.argumente); fehler == nil {
				t.Error("das Argument kam durch")
			}
		})
	}
}

// **Ein Fehler der Naht kommt an allen dreien durch.**
func TestDieListenwerkzeugeReichenIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := []struct{ werkzeug, argumente string }{
		{"lesen.kartenliste", `{"setID":1}`},
		{"schreiben.kartenliste.spalten", `{"setID":1,"spalten":[]}`},
		{"schreiben.kartenliste.sortierung", `{"setID":1}`},
	}

	for _, fall := range faelle {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			_, fehler := rufeListenwerkzeug(t,
				&listenprobe{fehler: domain.ErrNotFound}, fall.werkzeug, fall.argumente)
			if !errors.Is(fehler, domain.ErrNotFound) {
				t.Errorf("Fehler: %v", fehler)
			}
		})
	}
}

// **Eine leere Liste kommt als `[]` und nicht als `null`.**
//
// Ein Sprachmodell, das `null` bekommt, erzählt „es gibt keine Zeilen" — was
// stimmt — oder es stolpert. Eine leere Liste sagt dasselbe ohne Zweig.
func TestEineLeereListeKommtAlsListe(t *testing.T) {
	t.Parallel()

	probe := &leereliste{}

	text, fehler := rufeListenwerkzeug(t, probe, "lesen.kartenliste", `{"setID":1}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{`"spalten": []`, `"zeilen": []`} {
		if !strings.Contains(text, teil) {
			t.Errorf("die Antwort führt %q nicht: %s", teil, text)
		}
	}
}

// leereliste gibt nichts zurück — ein Set ohne Karten und ohne Spalten.
type leereliste struct{ listenprobe }

func (l *leereliste) Listenaufbau(context.Context, int64) (
	domain.Kartenlistenaufbau, error,
) {
	return domain.Kartenlistenaufbau{
		Ordnung: domain.StandardKartenordnung(),
	}, nil
}

// **Ohne Naht steht keines der drei in der Liste.**
func TestOhneListennahtGibtEsKeineListenwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{},
		Angaben:   neueAngabenprobe(),
		Stile:     neueStilprobe(),
		Bestand:   neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "kartenliste") {
			t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
		}
	}
}

// **Der Schnellfilter** — Werkstück U5 von Phase 5.1 (#1416).
func TestDasFilterwerkzeugGibtKennungenAlsZeichenketten(t *testing.T) {
	t.Parallel()

	probe := &listenprobe{}

	text, fehler := rufeListenwerkzeug(t, probe, "lesen.kartenfilter",
		`{"setID":1,"anfrage":"type:creature -goblin"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(text, `"7"`) || !strings.Contains(text, `"9"`) {
		t.Errorf("die Antwort führt die Kennungen nicht als Zeichenketten: %s", text)
	}

	if probe.gefragteAnfrage != "type:creature -goblin" {
		t.Errorf("gefragt wurde nach %q", probe.gefragteAnfrage)
	}
}

func TestDasFilterwerkzeugReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	if _, fehler := rufeListenwerkzeug(t, &listenprobe{fehler: domain.ErrNotFound},
		"lesen.kartenfilter", `{"setID":1,"anfrage":"x"}`); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", fehler)
	}

	if _, fehler := rufeListenwerkzeug(t, &listenprobe{}, "lesen.kartenfilter",
		`{"setID":"eins"}`); fehler == nil {
		t.Error("eine Setkennung, die keine Zahl ist, kam durch")
	}

	if _, fehler := rufeListenwerkzeug(t, &listenprobe{}, "lesen.kartenfilter",
		`{"setID":0,"anfrage":"x"}`); fehler == nil {
		t.Error("die Setkennung null kam durch")
	}
}
