package app

// **Aus welchem Stand dieses Programm gebaut ist** — ENG-06.AK-04
// (#1731).

import (
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func TestDerBauKommtAusDerBauinformation(t *testing.T) {
	t.Parallel()

	mitStand := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: "-trimpath", Value: "true"},
		{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
		{Key: "vcs.time", Value: "2026-09-23T00:00:00Z"},
		{Key: "vcs.modified", Value: "true"},
	}}

	for name, fall := range map[string]struct {
		info     *debug.BuildInfo
		gelesen  bool
		erwartet Bau
	}{
		"aus einem Stand der Versionsverwaltung": {mitStand, true, Bau{
			Stand:  "0123456789abcdef0123456789abcdef01234567",
			Gebaut: "2026-09-23T00:00:00Z", Veraendert: true,
		}},
		"ohne Versionsverwaltung — ein Testlauf etwa": {
			&debug.BuildInfo{}, true, Bau{Stand: unbekannterStand},
		},
		"ohne jede Bauinformation": {nil, false, Bau{Stand: unbekannterStand}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if bau := bauAus(fall.info, fall.gelesen); bau != fall.erwartet {
				t.Errorf("Bau = %+v, erwartet %+v", bau, fall.erwartet)
			}
		})
	}
}

// **Dieser Lauf selbst hat keinen Stand** — `go test` schreibt keinen in das
// Testbinärprogramm. Die Auskunft sagt das, statt ein Datum zu erfinden.
func TestDieserBauNenntSeinenStand(t *testing.T) {
	t.Parallel()

	if bau := DiesenBau(); bau.Stand == "" {
		t.Errorf("der Stand ist leer: %+v", bau)
	}
}

// probegestalt ist ein Verbund mit jeder Art, die eine Gestalt kennen muss.
type probegestalt struct {
	Name     string          `json:"name"`
	Anzahl   int64           `json:"anzahl,omitempty"`
	Liste    []probegestalt  `json:"liste"`
	Tabelle  map[string]bool `json:"tabelle"`
	Zeiger   *probegestalt   `json:"zeiger"`
	Feld     [2]float64      `json:"feld"`
	Beliebig any             `json:"beliebig"`
	Zeit     time.Time       `json:"zeit"`
	Ohne     string          `json:"-"`
	Kahl     bool
	klein    string            //nolint:unused // absichtlich nicht exportiert
	Kanal    map[int][]*string `json:"kanal"`
	eingebettet
}

// eingebettet steht in [probegestalt], wie JSON es flach ausgibt.
type eingebettet struct {
	Innen string `json:"innen"`
}

func TestDieGestaltFolgtDemDraht(t *testing.T) {
	t.Parallel()

	gestalt := gestaltVon(reflect.TypeFor[probegestalt](), map[reflect.Type]bool{})

	for _, teil := range []string{
		"name:string", "anzahl?:int64", "liste:[]probegestalt",
		"tabelle:map[string]bool", "zeiger:*probegestalt", "feld:[2]float64",
		"beliebig:any", "zeit:json(time.Time)", "Kahl:bool",
		"kanal:map[int][]*string", "innen:string",
	} {
		if !strings.Contains(gestalt, teil) {
			t.Errorf("die Gestalt nennt %q nicht: %s", teil, gestalt)
		}
	}

	for _, fehlt := range []string{"Ohne", "klein", "eingebettet"} {
		if strings.Contains(gestalt, fehlt) {
			t.Errorf("die Gestalt nennt %q, der Draht aber nicht: %s", fehlt, gestalt)
		}
	}
}

// **Eine andere Marke ist ein anderer Vertrag** — und genau das soll der
// Abdruck sehen: das Fenster liest die Marke, nicht den Feldnamen.
func TestEineUmbenannteMarkeAendertDenAbdruck(t *testing.T) {
	t.Parallel()

	type vorher struct {
		Wert string `json:"wert"`
	}

	type nachher struct {
		Wert string `json:"werte"`
	}

	if abdruckVon(gestaltVon(reflect.TypeFor[vorher](), map[reflect.Type]bool{})) ==
		abdruckVon(gestaltVon(reflect.TypeFor[nachher](), map[reflect.Type]bool{})) {
		t.Error("eine umbenannte Marke lässt den Abdruck gleich")
	}
}

func TestJedeGebundeneMethodeHatEinenAbdruck(t *testing.T) {
	t.Parallel()

	abdruecke := Bindungsvertrag()

	if len(abdruecke) != reflect.TypeFor[*Bindungen]().NumMethod() {
		t.Errorf("%d Abdrücke für %d Methoden", len(abdruecke),
			reflect.TypeFor[*Bindungen]().NumMethod())
	}

	if len(abdruecke["Sets"]) != abdrucklaenge || len(Vertragsabdruck()) != abdrucklaenge {
		t.Errorf("Sets = %q, gesamt %q", abdruecke["Sets"], Vertragsabdruck())
	}
}

func TestDieBauzeileNenntWasDerBauWeiss(t *testing.T) {
	t.Parallel()

	texte := DialogtexteFuer("de")

	for name, fall := range map[string]struct {
		bau      Bau
		erwartet string
	}{
		"mit Zeit und Vermerk": {
			Bau{Stand: "abc", Gebaut: "2026-09-23T00:00:00Z", Veraendert: true},
			"Bau: abc (2026-09-23T00:00:00Z), Arbeitsbaum verändert",
		},
		"ohne Stand": {Bau{Stand: unbekannterStand}, "Bau: unbekannt"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if zeile := bauzeile(fall.bau, texte); zeile != fall.erwartet {
				t.Errorf("%q, erwartet %q", zeile, fall.erwartet)
			}
		})
	}
}
