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

// **Die drei Ablagewerkzeuge** — Werkstück 1 von Phase 3.7
// (#1168).

// ablageprobe ist ein Doppel der Naht.
type ablageprobe struct {
	inhalt  string
	fehler  error
	gefragt []int64
	gelegt  string
}

func (a *ablageprobe) AblageKopieren(
	_ context.Context, _ int64, karteIDs []int64,
) (string, error) {
	a.gefragt = karteIDs

	return a.inhalt, a.fehler
}

func (a *ablageprobe) AblageAusschneiden(
	ctx context.Context, setID int64, karteIDs []int64,
) (string, error) {
	return a.AblageKopieren(ctx, setID, karteIDs)
}

func (a *ablageprobe) AblageEinfuegen(_ context.Context, _ int64, inhalt string) (
	karten, befunde []string, fehler error,
) {
	a.gelegt = inhalt

	return []string{"11"}, []string{"das Feld »power« gibt es nicht"}, a.fehler
}

// rufeAblagewerkzeug ruft eines der drei über die **echte** Werkzeugliste.
//
// **Über die Liste und nicht über die Funktion**: was hier zu prüfen ist, ist
// auch, dass die drei überhaupt darin stehen. Eine Naht ohne Eintrag wäre eine
// Fähigkeit, die niemand erreicht (§2).
func rufeAblagewerkzeug(
	t *testing.T, ablage mcp.Fremdablage, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{},
		Ablage:    ablage,
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

// **Kopieren und Ausschneiden gehen denselben Weg und nennen dieselben
// Karten.**
func TestDieAblagewerkzeugeGebenDenInhalt(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"lesen.ablage.kopieren", "schreiben.ablage.ausschneiden",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &ablageprobe{inhalt: "game_name: probe\ncard:\n"}

			text, fehler := rufeAblagewerkzeug(t, probe, name,
				`{"setID":1,"karteIDs":[11,12]}`)
			if fehler != nil {
				t.Fatalf("Ruf: %v", fehler)
			}

			if !strings.Contains(text, "game_name") {
				t.Errorf("die Antwort trägt den Inhalt nicht: %s", text)
			}

			if len(probe.gefragt) != 2 || probe.gefragt[0] != 11 {
				t.Errorf("gefragt wurde nach %v", probe.gefragt)
			}
		})
	}
}

// **Einfügen nennt die neuen Karten und die Befunde.**
func TestDasEinfuegewerkzeugNenntKartenUndBefunde(t *testing.T) {
	t.Parallel()

	probe := &ablageprobe{}

	text, fehler := rufeAblagewerkzeug(t, probe, "schreiben.ablage.einfuegen",
		`{"setID":1,"inhalt":"game_name: probe\ncard:\n"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{`"11"`, "power"} {
		if !strings.Contains(text, teil) {
			t.Errorf("die Antwort führt %q nicht: %s", teil, text)
		}
	}

	if !strings.Contains(probe.gelegt, "game_name") {
		t.Errorf("hineingereicht wurde %q", probe.gelegt)
	}
}

// **Jede Wache greift, und zwar an jedem der drei Werkzeuge.**
//
// **Alle drei und nicht eines**: ein Sprachmodell rät seine Argumente, und
// eine Wache, die nur an einem Werkzeug steht, lässt am nächsten eine Kennung
// durch, die es nicht gibt.
func TestDieAblagewerkzeugeWeisenUnbrauchbareArgumenteAb(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, werkzeug, argumente string }{
		{"kopieren: kein JSON", "lesen.ablage.kopieren", `{`},
		{"kopieren: Set ist null", "lesen.ablage.kopieren", `{"setID":0,"karteIDs":[1]}`},
		{"kopieren: keine Karte", "lesen.ablage.kopieren", `{"setID":1,"karteIDs":[]}`},
		{"kopieren: Karte ist null", "lesen.ablage.kopieren", `{"setID":1,"karteIDs":[0]}`},
		{"ausschneiden: keine Karte", "schreiben.ablage.ausschneiden", `{"setID":1}`},
		{"einfügen: kein JSON", "schreiben.ablage.einfuegen", `{`},
		{"einfügen: Set ist null", "schreiben.ablage.einfuegen", `{"setID":0,"inhalt":"x"}`},
		{"einfügen: kein Inhalt", "schreiben.ablage.einfuegen", `{"setID":1,"inhalt":"  "}`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeAblagewerkzeug(
				t, &ablageprobe{}, fall.werkzeug, fall.argumente); fehler == nil {
				t.Error("das Argument kam durch")
			}
		})
	}
}

// **Ein Fehler der Naht kommt an allen dreien durch.**
func TestDieAblagewerkzeugeReichenIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := []struct{ werkzeug, argumente string }{
		{"lesen.ablage.kopieren", `{"setID":1,"karteIDs":[11]}`},
		{"schreiben.ablage.ausschneiden", `{"setID":1,"karteIDs":[11]}`},
		{"schreiben.ablage.einfuegen", `{"setID":1,"inhalt":"game_name: probe\n"}`},
	}

	for _, fall := range faelle {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &ablageprobe{fehler: domain.ErrNotFound}

			_, fehler := rufeAblagewerkzeug(t, probe, fall.werkzeug, fall.argumente)
			if !errors.Is(fehler, domain.ErrNotFound) {
				t.Errorf("Fehler: %v", fehler)
			}
		})
	}
}

// **Ohne Naht steht keines der drei in der Liste.**
//
// Ein Werkzeug, das sich anbietet und beim Ruf über eine leere Naht stolpert,
// wäre schlimmer als keines: das Modell hätte es angekündigt bekommen.
func TestOhneAblagenahtGibtEsKeineAblagewerkzeuge(t *testing.T) {
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
		if strings.Contains(werkzeug.Name, "ablage") {
			t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
		}
	}
}
