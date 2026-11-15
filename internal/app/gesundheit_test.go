package app_test

// **Die Gesundheit an der Fassade** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gesundheitsprobe antwortet mit einem festen Zustand.
type gesundheitsprobe struct{ zustand domain.Betriebszustand }

func (g gesundheitsprobe) Zustand(context.Context) domain.Betriebszustand {
	return g.zustand
}

// **Ohne Naht lebt er und ist nicht bereit** — und das ist kein Fehler.
func TestOhneGesundheitsnahtLebtErUndIstNichtBereit(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Gesundheit(t.Context())

	switch {
	case !antwort.Lebt:
		t.Error("er galt als tot")
	case antwort.Bereit:
		t.Error("er galt als bereit")
	case antwort.Grund != "bestandZu":
		t.Errorf("Grund %q", antwort.Grund)
	case antwort.Fremde == nil:
		t.Error("die Liste war null statt leer")
	}
}

// **Ein fremder Ausfall steht in der Antwort und nimmt die Bereitschaft
// nicht** — die Abnahme des Issues, an der Naht nachgeprüft.
func TestEinFremderAusfallKommtDurchOhneDieBereitschaftZuNehmen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitGesundheit(gesundheitsprobe{zustand: domain.Betriebszustand{
		Lebt: true,
		Fremde: []domain.Fremdzustand{
			{Dienst: domain.DienstScryfall, Erreichbar: false, Grund: "kein Netz"},
		},
	}})

	antwort := anwendung.Gesundheit(t.Context())

	switch {
	case !antwort.Bereit:
		t.Error("ein fremder Ausfall nahm die Bereitschaft")
	case len(antwort.Fremde) != 1:
		t.Fatalf("%+v", antwort.Fremde)
	case antwort.Fremde[0].Dienst != "scryfall" || antwort.Fremde[0].Erreichbar:
		t.Errorf("%+v", antwort.Fremde[0])
	}
}

// **Und ein eigener Grund kommt als Grund an.**
func TestEinEigenerGrundKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitGesundheit(gesundheitsprobe{zustand: domain.Betriebszustand{
		Lebt: true, Grund: domain.GrundWanderungLaeuft,
	}})

	antwort := anwendung.Gesundheit(t.Context())

	if antwort.Bereit || antwort.Grund != "wanderungLaeuft" {
		t.Errorf("%+v", antwort)
	}
}

// **Die Bindung reicht durch** — samt ihrer Antwort (§2).
func TestDieBindungGibtDieGesundheit(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitGesundheit(gesundheitsprobe{zustand: domain.Betriebszustand{Lebt: true}})

	if antwort := app.NeueBindungen(anwendung).Gesundheit(t.Context()); !antwort.Bereit {
		t.Errorf("%+v", antwort)
	}
}

// **Die Naht über einem Bestand** — der Weg, den `cmd/` und der MCP-Server
// nehmen (§2: `cmd/` kennt den Dienst nicht).
func TestDieGesundheitsnahtUeberEinemBestand(t *testing.T) {
	t.Parallel()

	bestand := frischerBestand(t)

	if zustand := app.GesundheitsnahtUeber(bestand).Zustand(t.Context()); !zustand.Bereit() {
		t.Errorf("%+v", zustand)
	}

	// **Und ein geschlossener Bestand ist unlesbar** — die Naht fragt ihn
	// wirklich, statt „bereit" zu behaupten.
	if err := bestand.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if grund := app.GesundheitsnahtUeber(bestand).Zustand(t.Context()).Grund; grund !=
		domain.GrundBestandUnlesbar {
		t.Errorf("Grund %q, erwartet %q", grund, domain.GrundBestandUnlesbar)
	}
}
