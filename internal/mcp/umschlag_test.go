// **Die drei Umschlagwerkzeuge** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Was hier geprüft wird
//
//   - Alle drei kommen durch — mit ihren Argumenten, in ihrer Gestalt.
//   - Alle drei weisen eine leere Angabe ab, **bevor** sie die Naht rufen.
//   - Ein Fehler der Naht kommt an und wird nicht verschluckt.
//   - **Und es gibt sie nicht ohne Naht**: `nil` heißt kein Werkzeug.
//   - **Das lesende Werkzeug schreibt nicht** — es ist als `Lesend` geführt
//     und idempotent, und das ist die Zusage, an der ein Gegenüber seine
//     Liste kürzt.
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// umschlagprobe ist eine Naht, die mitschreibt.
type umschlagprobe struct {
	geschnuert  [2]string
	angesehen   string
	uebernommen string

	fehler error

	// art ist, was der Aufrufer schnüren wollte (T7, #1472).
	art string
}

func (u *umschlagprobe) UmschlagSchnueren(_ context.Context, art, kennung, ziel string) (
	mcp.Umschlagangaben, error,
) {
	u.geschnuert = [2]string{kennung, ziel}
	u.art = art

	if u.fehler != nil {
		return mcp.Umschlagangaben{}, u.fehler
	}

	return mcp.Umschlagangaben{Objekte: []mcp.Umschlagobjekt{
		{Art: "set", Bezeichnung: "set", Groesse: 258, Braucht: []string{}},
	}}, nil
}

func (u *umschlagprobe) UmschlagAnsehen(_ context.Context, pfad string) (
	mcp.Umschlagangaben, error,
) {
	u.angesehen = pfad

	if u.fehler != nil {
		return mcp.Umschlagangaben{}, u.fehler
	}

	return mcp.Umschlagangaben{
		Urheber: "bafzaaja", Fingerabdruck: "06f8-79a5",
		Objekte: []mcp.Umschlagobjekt{
			{Art: "set", Bezeichnung: "set", Groesse: 258, Braucht: []string{}},
		},
	}, nil
}

func (u *umschlagprobe) UmschlagUebernehmen(_ context.Context, pfad string) (
	mcp.Umschlagannahme, error,
) {
	u.uebernommen = pfad

	if u.fehler != nil {
		return mcp.Umschlagannahme{}, u.fehler
	}

	return mcp.Umschlagannahme{SetID: "9", Karten: 3, Befunde: []string{}}, nil
}

// rufeUmschlag sucht ein Werkzeug und ruft es.
func rufeUmschlag(
	t *testing.T, naht mcp.Umschlagwerkstatt, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Umschlag: naht}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q steht nicht in der Liste", name)

	return "", nil
}

func TestDieUmschlagwerkzeugeKommenDurch(t *testing.T) {
	naht := &umschlagprobe{}

	aus, err := rufeUmschlag(t, naht, "schreiben.umschlag.schnueren",
		`{"setID":"7","ziel":"/tmp/x.msepackage"}`)
	if err != nil {
		t.Fatalf("schnüren: %v", err)
	}

	if naht.geschnuert != [2]string{"7", "/tmp/x.msepackage"} {
		t.Errorf("geschnürt wurde %+v", naht.geschnuert)
	}

	if !strings.Contains(aus, `"art": "set"`) {
		t.Errorf("die Antwort nennt kein Objekt: %s", aus)
	}

	aus, err = rufeUmschlag(t, naht, "lesen.umschlag", `{"pfad":"/tmp/x.msepackage"}`)
	if err != nil {
		t.Fatalf("ansehen: %v", err)
	}

	if naht.angesehen != "/tmp/x.msepackage" {
		t.Errorf("angesehen wurde %q", naht.angesehen)
	}

	// **Der Urheber steht dabei** — ohne ihn wäre die Auskunft die halbe.
	for _, erwartet := range []string{`"urheber": "bafzaaja"`, `"fingerabdruck": "06f8-79a5"`} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("%s fehlt in %s", erwartet, aus)
		}
	}

	// **Und das Ansehen hat nichts übernommen** — das ist die Zusage aus
	// #1457, und sie steht hier als Messung und nicht als Kommentar.
	if naht.uebernommen != "" {
		t.Errorf("das Ansehen hat %q übernommen", naht.uebernommen)
	}

	aus, err = rufeUmschlag(t, naht, "schreiben.umschlag.uebernehmen",
		`{"pfad":"/tmp/x.msepackage"}`)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if naht.uebernommen != "/tmp/x.msepackage" {
		t.Errorf("übernommen wurde %q", naht.uebernommen)
	}

	if !strings.Contains(aus, `"setID": "9"`) {
		t.Errorf("die Antwort nennt die neue Kennung nicht: %s", aus)
	}
}

// **Eine leere Angabe wird abgewiesen, bevor die Naht gerufen wird.**
func TestEineLeereAngabeErreichtDieUmschlagnahtNicht(t *testing.T) {
	for name, fall := range map[string]struct {
		werkzeug, argumente string
	}{
		"schnüren ohne Ziel":   {"schreiben.umschlag.schnueren", `{"setID":"7","ziel":""}`},
		"schnüren ohne Set":    {"schreiben.umschlag.schnueren", `{"setID":"","ziel":"/tmp/x"}`},
		"ansehen ohne Pfad":    {"lesen.umschlag", `{"pfad":""}`},
		"übernehmen ohne Pfad": {"schreiben.umschlag.uebernehmen", `{"pfad":""}`},
	} {
		t.Run(name, func(t *testing.T) {
			naht := &umschlagprobe{}

			if _, err := rufeUmschlag(t, naht, fall.werkzeug, fall.argumente); err == nil {
				t.Fatal("die leere Angabe kam durch")
			} else if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("der Fehler ist keine Validierung: %v", err)
			}

			if naht.angesehen != "" || naht.uebernommen != "" ||
				naht.geschnuert != [2]string{} {
				t.Error("die Naht wurde trotzdem gerufen")
			}
		})
	}
}

// **Fremde Argumente und ein Fehler der Naht kommen an.**
func TestDieUmschlagwerkzeugeReichenIhreFehlerDurch(t *testing.T) {
	naht := &umschlagprobe{fehler: errors.New("die Datei ist weg")}

	for name, fall := range map[string]struct {
		werkzeug, argumente string
	}{
		"schnüren":   {"schreiben.umschlag.schnueren", `{"setID":"7","ziel":"/tmp/x"}`},
		"ansehen":    {"lesen.umschlag", `{"pfad":"/tmp/x"}`},
		"übernehmen": {"schreiben.umschlag.uebernehmen", `{"pfad":"/tmp/x"}`},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := rufeUmschlag(t, naht, fall.werkzeug, fall.argumente); err == nil {
				t.Error("der Fehler wurde verschluckt")
			}
		})
	}

	// **Und ein kaputtes Argument endet vor der Naht.**
	for name, werkzeug := range map[string]string{
		"schnüren":   "schreiben.umschlag.schnueren",
		"ansehen":    "lesen.umschlag",
		"übernehmen": "schreiben.umschlag.uebernehmen",
	} {
		t.Run(name+" mit kaputtem JSON", func(t *testing.T) {
			if _, err := rufeUmschlag(t, naht, werkzeug, `kein JSON`); err == nil {
				t.Error("das kaputte Argument kam durch")
			}
		})
	}
}

// **Ohne Naht gibt es die drei Werkzeuge nicht.**
//
// Die Gegenprobe zu `rufeUmschlag`: es ist dieselbe Liste, und ein `nil`
// lässt sie schrumpfen statt einen Ruf ins Leere anzubieten.
func TestOhneUmschlagnahtGibtEsDieWerkzeugeNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "umschlag") {
			t.Errorf("%s steht da, obwohl keine Naht eingerichtet ist", werkzeug.Name)
		}
	}
}

// **Das lesende Werkzeug ist als lesend geführt und idempotent.**
//
// **Daran kürzt ein Gegenüber seine Liste** — wer es als schreibend führte,
// nähme einem vorsichtigen Aufrufer die Auskunft, und wer ein schreibendes
// als lesend führte, gäbe sie ihm unter falschem Namen.
func TestDasLesendeUmschlagwerkzeugIstLesend(t *testing.T) {
	arten := map[string]mcp.Art{}

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Umschlag: &umschlagprobe{}}) {
		if strings.Contains(werkzeug.Name, "umschlag") {
			arten[werkzeug.Name] = werkzeug.Art

			if werkzeug.Name == "lesen.umschlag" && !werkzeug.Idempotent {
				t.Error("lesen.umschlag ist nicht als idempotent geführt")
			}
		}
	}

	erwartet := map[string]mcp.Art{
		"lesen.umschlag":                 mcp.Lesend,
		"schreiben.umschlag.schnueren":   mcp.Schreibend,
		"schreiben.umschlag.uebernehmen": mcp.Schreibend,
	}

	for name, art := range erwartet {
		if arten[name] != art {
			t.Errorf("%s ist als %v geführt, erwartet %v", name, arten[name], art)
		}
	}
}
