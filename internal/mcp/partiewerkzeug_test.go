package mcp_test

// **Die vier Partiewerkzeuge** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Sprachmodell nur die Sicht eines Sitzes bekommt** — und dass es
// sie überhaupt bekommt. Der Weg zum Vollzustand fehlt hier nicht aus
// Versehen; er ist ausdrücklich nicht gebaut, und für ein Modell wiegt der
// Grund am schwersten: es vergisst nichts, was es einmal gelesen hat.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// partienprobe ist eine Werkstatt, die antwortet oder klagt.
type partienprobe struct {
	klagt error

	// gesehen hält fest, womit gerufen wurde — die Prüfung gilt dem
	// **Argument** und nicht dem bloßen Ruf (§10).
	gesehen struct {
		spiel, wirt, kennung, fuer, stand string

		// inhalt ist der Text der Einfuhr (PLAY-03.AK-05, #1711).
		inhalt string
		sitze  []mcp.Sitzwunsch
		befehl mcp.Spielbefehl

		// abgewaehlt ist die Abwahl des Anlegens (P1a, #2335).
		abgewaehlt []string
	}
}

func (p *partienprobe) angaben() mcp.Partieangaben {
	return mcp.Partieangaben{
		ID: "p1", Fassung: "1", Spiel: "magic", Stand: "laeuft",
		Fuer: "sitz-1", Wirt: "sitz-1",
		Sitze: []mcp.Partiesitz{
			{Kennung: "sitz-1", Name: "Anna", Karten: 3},
			{Kennung: "sitz-2", Name: "Bert", Karten: 3},
		},
		Objekte: []mcp.Partieobjekt{
			{
				Kennung: "sitz-1/k1/0", Beherrscher: "sitz-1",
				Zone: "bibliothek", Geheim: true,
			},
		},
		Mischpflichten: []mcp.Mischpflicht{},
		Erlaubt:        []string{"objekt.tappen"},
	}
}

func (p *partienprobe) PartieAnlegen(
	_ context.Context, spiel, wirt string, sitze []mcp.Sitzwunsch,
	abgewaehlt []string,
) (mcp.Partieangaben, error) {
	p.gesehen.spiel, p.gesehen.wirt, p.gesehen.sitze = spiel, wirt, sitze
	p.gesehen.abgewaehlt = abgewaehlt

	return p.angaben(), p.klagt
}

// Partien gibt eine Zeile — die Liste ist ein eigener Weg (#1600).
func (p *partienprobe) Partien(context.Context) (
	[]mcp.Partiezeilenangaben, error,
) {
	return []mcp.Partiezeilenangaben{{
		ID: "p1", Spiel: "magic", Stand: "laeuft", Objekte: 12,
		Sitze:       []mcp.Zeilensitzangaben{{Kennung: "anna", Name: "Anna"}},
		EigenerSitz: "anna", Dran: true,
	}}, p.klagt
}

func (p *partienprobe) Partie(_ context.Context, kennung, fuer string) (
	mcp.Partieangaben, error,
) {
	p.gesehen.kennung, p.gesehen.fuer = kennung, fuer

	return p.angaben(), p.klagt
}

func (p *partienprobe) PartieBefehl(_ context.Context, befehl mcp.Spielbefehl) (
	mcp.Befehlsangaben, error,
) {
	p.gesehen.befehl = befehl

	return mcp.Befehlsangaben{
		Fassung: "2",
		Ereignisse: []mcp.Partieereignis{{
			Folge: "1", Art: befehl.Art,
			Oeffentlich: map[string]any{"sitz": befehl.Absender},
			Privat:      map[string]any{"karten": []any{}},
			Kette:       "abcd",
		}},
	}, p.klagt
}

func (p *partienprobe) PartieStand(
	_ context.Context, kennung, fuer, stand string,
) (mcp.Partieangaben, error) {
	p.gesehen.kennung, p.gesehen.fuer, p.gesehen.stand = kennung, fuer, stand

	return p.angaben(), p.klagt
}

// rufePartiewerkzeug ruft eines der vier.
func rufePartiewerkzeug(
	t *testing.T, partien mcp.Partienwerkstatt, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Partien: partien,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// **Ohne Naht gibt es die Werkzeuge gar nicht.**
//
// **Ein Werkzeug, das ohne Bestand dastünde, wäre eine Einladung**, die beim
// ersten Ruf abgelehnt würde — dieselbe Gestalt wie bei jeder anderen Naht.
func TestOhnePartiennahtGibtEsKeinePartiewerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
	}) {
		if strings.Contains(werkzeug.Name, "partie") {
			t.Errorf("%q steht ohne Naht da", werkzeug.Name)
		}
	}
}

// **Die Sicht eines Sitzes kommt — und die fremde Hand nicht mit.**
func TestDasWerkzeugPartieGibtDieSichtEinesSitzes(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{}

	antwort, err := rufePartiewerkzeug(t, probe, "lesen.partie",
		`{"partieID":"p1","fuer":"sitz-1"}`)
	if err != nil {
		t.Fatalf("lesen.partie: %v", err)
	}

	if probe.gesehen.kennung != "p1" || probe.gesehen.fuer != "sitz-1" {
		t.Errorf("gerufen mit %+v", probe.gesehen)
	}

	for _, teil := range []string{
		`"id": "p1"`, `"fuer": "sitz-1"`, `"wirt": "sitz-1"`,
		`"geheim": true`, `"erlaubt"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}

	// **Und `karte` steht nicht darin** — das Feld ist `omitempty`, und wo
	// nichts hineingesehen werden darf, steht nichts.
	if strings.Contains(antwort, `"karte"`) {
		t.Errorf("die Antwort nennt eine verdeckte Karte:\n%s", antwort)
	}
}

// **Was das Anlegen weitergibt, ist der Entwurf.**
func TestDasWerkzeugPartieAnlegenReichtDieSitzeDurch(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{}

	if _, err := rufePartiewerkzeug(t, probe, "schreiben.partie.anlegen",
		`{"spiel":"magic","wirt":"sitz-2","sitze":[`+
			`{"kennung":"sitz-1","name":"Anna","deckID":"1"},`+
			`{"kennung":"sitz-2","name":"Bert","deckID":"2"}]}`); err != nil {
		t.Fatalf("schreiben.partie.anlegen: %v", err)
	}

	if probe.gesehen.spiel != "magic" || probe.gesehen.wirt != "sitz-2" ||
		len(probe.gesehen.sitze) != 2 ||
		probe.gesehen.sitze[1].DeckID != "2" {
		t.Errorf("gerufen mit %+v", probe.gesehen)
	}

	// **Ohne Abwahl bleibt sie leer** — der Normalfall (P1a, #2335).
	if len(probe.gesehen.abgewaehlt) != 0 {
		t.Errorf("ohne Angabe kam die Abwahl %v an", probe.gesehen.abgewaehlt)
	}
}

// **Und die Abwahl reist mit** — P1a
// (#2335).
// Ob es den Bereich gibt, entscheidet der Kern; hier steht, dass er ihn
// zu sehen bekommt.
func TestDasWerkzeugPartieAnlegenReichtDieAbwahlDurch(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{}

	if _, err := rufePartiewerkzeug(t, probe, "schreiben.partie.anlegen",
		`{"spiel":"magic","abgewaehlt":["zugstruktur"],"sitze":[`+
			`{"kennung":"sitz-1","name":"Anna","deckID":"1"},`+
			`{"kennung":"sitz-2","name":"Bert","deckID":"2"}]}`); err != nil {
		t.Fatalf("schreiben.partie.anlegen: %v", err)
	}

	if len(probe.gesehen.abgewaehlt) != 1 ||
		probe.gesehen.abgewaehlt[0] != "zugstruktur" {
		t.Errorf("angekommen: %v", probe.gesehen.abgewaehlt)
	}
}

// **Ein Befehl kommt mit Kennung, Absender, Art und Nutzlast an.**
//
// **Und seine Ereignisse tragen ihren Inhalt zurück** — ohne `oeffentlich`
// und `privat` hätte ein Modell, das seine Bibliothek durchsucht, nichts
// gefunden (§16).
func TestDasWerkzeugPartieBefehlGibtDenInhaltZurueck(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{}

	antwort, err := rufePartiewerkzeug(t, probe, "schreiben.partie.befehl",
		`{"partieID":"p1","absender":"sitz-1","kennung":"b1",`+
			`"art":"zone.blick","nutzlast":{"zone":"bibliothek"}}`)
	if err != nil {
		t.Fatalf("schreiben.partie.befehl: %v", err)
	}

	if probe.gesehen.befehl.Kennung != "b1" ||
		probe.gesehen.befehl.Art != "zone.blick" ||
		probe.gesehen.befehl.Nutzlast["zone"] != "bibliothek" {
		t.Errorf("gerufen mit %+v", probe.gesehen.befehl)
	}

	for _, teil := range []string{
		`"fassung": "2"`, `"art": "zone.blick"`, `"oeffentlich"`, `"privat"`,
		`"kette": "abcd"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Ein Standwechsel reicht sein Wort durch.**
func TestDasWerkzeugPartieStandReichtDenStandDurch(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{}

	if _, err := rufePartiewerkzeug(t, probe, "schreiben.partie.stand",
		`{"partieID":"p1","fuer":"sitz-1","stand":"beenden"}`); err != nil {
		t.Fatalf("schreiben.partie.stand: %v", err)
	}

	if probe.gesehen.stand != "beenden" {
		t.Errorf("gerufen mit %q", probe.gesehen.stand)
	}
}

// **Was ein Werkzeug abweist, bevor es die Naht erreicht.**
//
// **Der Erfolgsweg steht oben** — ohne ihn sagte „wird abgewiesen" nur, dass
// irgendetwas fehlt (§10).
func TestWasDiePartiewerkzeugeAbweisen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		werkzeug, argumente string
	}{
		"lesen ohne Partie":      {"lesen.partie", `{"fuer":"sitz-1"}`},
		"lesen ohne Sitz":        {"lesen.partie", `{"partieID":"p1"}`},
		"lesen mit Unsinn":       {"lesen.partie", `{"partieID":1}`},
		"anlegen ohne Spiel":     {"schreiben.partie.anlegen", `{"sitze":[{},{}]}`},
		"anlegen mit einem Sitz": {"schreiben.partie.anlegen", `{"spiel":"magic","sitze":[{}]}`},
		"anlegen mit Unsinn":     {"schreiben.partie.anlegen", `{"spiel":1}`},
		"befehl ohne Absender": {
			"schreiben.partie.befehl",
			`{"partieID":"p1","kennung":"b1","art":"objekt.tappen"}`,
		},
		"befehl ohne Art": {
			"schreiben.partie.befehl",
			`{"partieID":"p1","absender":"sitz-1","kennung":"b1"}`,
		},
		"befehl mit Unsinn": {"schreiben.partie.befehl", `{"partieID":1}`},
		"stand ohne Sitz": {
			"schreiben.partie.stand", `{"partieID":"p1","stand":"beenden"}`,
		},
		"stand mit Unsinn": {"schreiben.partie.stand", `{"partieID":1}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufePartiewerkzeug(t, &partienprobe{},
				fall.werkzeug, fall.argumente); err == nil {
				t.Errorf("%s kam durch", fall.argumente)
			}
		})
	}
}

// **Und was die Naht klagt, klagt das Werkzeug weiter.**
func TestWasDiePartiennahtKlagtKommtDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	for name, fall := range map[string]struct {
		werkzeug, argumente string
	}{
		"lesen":   {"lesen.partie", `{"partieID":"p1","fuer":"sitz-1"}`},
		"anlegen": {"schreiben.partie.anlegen", `{"spiel":"magic","sitze":[{},{}]}`},
		"befehl": {
			"schreiben.partie.befehl",
			`{"partieID":"p1","absender":"sitz-1","kennung":"b1","art":"x"}`,
		},
		"stand": {
			"schreiben.partie.stand",
			`{"partieID":"p1","fuer":"sitz-1","stand":"beenden"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := rufePartiewerkzeug(t, &partienprobe{klagt: kaputt},
				fall.werkzeug, fall.argumente)

			if !errors.Is(err, kaputt) {
				t.Errorf("Fehler %v, erwartet die Klage der Naht", err)
			}
		})
	}
}
