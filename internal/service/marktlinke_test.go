package service_test

// **Die Wege nach draußen** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
// **Dass eine Karte ohne kanonischen Druck keine geratenen Produktlinks
// bekommt** — das Issue sagt es wörtlich, und es ist der Unterschied zwischen
// „ich weiß, wo das liegt" und „das könnte es sein".
//
// **Dass der Öffner das Argument bekommt, das er bekommen soll** — und keine
// Adresse, die jemand anderes geschickt hat. Der Lauf prüft das Argument und
// nicht den Aufruf.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// linkdruckprobe gibt einen Druck samt seinen Kennungen.
type linkdruckprobe struct {
	druck *domain.Druck
	klagt error
}

func (d linkdruckprobe) NachScryfall(
	context.Context, string,
) (*domain.Druck, error) {
	return d.druck, d.klagt
}

// artVon sucht die Art des Links eines Anbieters.
func artVon(links []domain.Marktlink, anbieter domain.Linkanbieter) domain.Linkart {
	for _, link := range links {
		if link.Anbieter == anbieter {
			return link.Art
		}
	}

	return ""
}

func TestEinBelegterDruckGibtProduktlinks(t *testing.T) {
	t.Parallel()

	werk := service.NeuesLinkwerk(linkdruckprobe{druck: &domain.Druck{
		Name: "Lightning Bolt", Setcode: "m10", Nummer: "146",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemTCGplayer, Semantik: domain.SemantikProdukt,
			Wert: "33632",
		}},
	}})

	links, fehler := werk.Linksfuer(t.Context(), "bolt-m10", "")
	if fehler != nil {
		t.Fatalf("Linksfuer: %v", fehler)
	}

	if art := artVon(links, domain.LinkTCGplayer); art != domain.LinkProdukt {
		t.Errorf("TCGplayer gab %q", art)
	}

	// **Der Name kommt aus dem Druck**, wenn der Aufrufer keinen mitbringt —
	// sonst stünde dort eine Suche ohne Suchwort.
	if art := artVon(links, domain.LinkCardmarket); art != domain.LinkSuche {
		t.Errorf("Cardmarket gab %q", art)
	}
}

// **Eine eigene Karte bekommt keine geratenen Produktlinks.**
func TestEineEigeneKarteBekommtKeineGeratenenProduktlinks(t *testing.T) {
	t.Parallel()

	werk := service.NeuesLinkwerk(linkdruckprobe{})

	links, fehler := werk.Linksfuer(t.Context(), "", "Mein Drache")
	if fehler != nil {
		t.Fatalf("Linksfuer: %v", fehler)
	}

	for _, link := range links {
		if link.Art == domain.LinkProdukt {
			t.Errorf("%s gab einen Produktlink: %q",
				link.Anbieter, link.Adresse)
		}
	}

	// **Und Gatherer bleibt bei der Startseite** — ohne Multiverse-ID ist es
	// nicht verfügbar.
	if art := artVon(links, domain.LinkGatherer); art != domain.LinkStartseite {
		t.Errorf("Gatherer gab %q", art)
	}
}

// **Ein Druck, den es nicht gibt, ist kein Fehler.**
func TestEinUnbekannterDruckGibtTrotzdemWege(t *testing.T) {
	t.Parallel()

	werk := service.NeuesLinkwerk(linkdruckprobe{})

	links, fehler := werk.Linksfuer(t.Context(), "gibtesnicht", "Bolt")
	if fehler != nil {
		t.Fatalf("Linksfuer: %v", fehler)
	}

	if len(links) != len(domain.Linkanbieterliste()) {
		t.Errorf("es kamen %d Links", len(links))
	}
}

// **Ein klagender Bestand klagt weiter.**
func TestEinKlagenderBestandKlagtBeimLinkbauWeiter(t *testing.T) {
	t.Parallel()

	klage := errors.New("der Bestand klagt")
	werk := service.NeuesLinkwerk(linkdruckprobe{klagt: klage})

	if _, fehler := werk.Linksfuer(
		t.Context(), "bolt-m10", ""); !errors.Is(fehler, klage) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Der Öffner bekommt genau die geprüfte Adresse.**
func TestDerOeffnerBekommtNurEineGeprueftAdresse(t *testing.T) {
	t.Parallel()

	werk := service.NeuesLinkwerk(linkdruckprobe{})

	gesehen := ""

	// **Der Öffner darf scheitern**, und dann muss es ankommen: `open` gibt
	// es nicht überall, und eine Adresse ohne eingetragenes Programm endet
	// mit einem Fehler. Ein Doppel, das immer glückt, ließe diesen Weg
	// unbetreten (§10).
	var klage error

	oeffner := func(_ context.Context, adresse string) error {
		gesehen = adresse

		return klage
	}

	const gut = "https://www.cardmarket.com/en/Magic/Products/Search?searchString=Bolt"

	if fehler := werk.Oeffne(
		t.Context(), oeffner, "cardmarket", gut); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	if gesehen != gut {
		t.Errorf("der Öffner bekam %q", gesehen)
	}

	// **Und seine Klage kommt zurück** — sie zu verschlucken hieße, einen
	// Browser zu melden, der nie aufging.
	klage = errors.New("kein Browser eingetragen")

	if fehler := werk.Oeffne(
		t.Context(), oeffner, "cardmarket", gut); !errors.Is(fehler, klage) {
		t.Errorf("die Klage kam als %v zurück", fehler)
	}

	klage = nil

	// **Was ein Gegenüber schickt, erreicht ihn nicht** — das ist die
	// Zusage: „Peerpayload darf niemals selbst öffnen."
	gesehen = ""

	for name, fall := range map[string]struct{ anbieter, adresse string }{
		"ein fremder Wirt":      {"cardmarket", "https://boese.example/"},
		"kein https":            {"cardmarket", "http://www.cardmarket.com/"},
		"ein fremdes Protokoll": {"cardmarket", "file:///etc/passwd"},
		"ein fremder Anbieter":  {"kartenkiosk", "https://www.cardmarket.com/"},
		"ein Geheimnis dabei": {
			"cardmarket", "https://www.cardmarket.com/?token=abc",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := werk.Oeffne(t.Context(), oeffner,
				fall.anbieter, fall.adresse); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("die Adresse kam mit %v durch", fehler)
			}

			if gesehen != "" {
				t.Errorf("der Öffner bekam %q", gesehen)
			}
		})
	}
}

// **Ohne Öffner sagt es das.**
//
// **Eine Fläche ohne Fenster hat keinen Browser** — und ein Ruf, der still
// nichts täte, sähe aus wie einer, der etwas getan hat.
func TestOhneOeffnerSagtEsDieFlaeche(t *testing.T) {
	t.Parallel()

	werk := service.NeuesLinkwerk(linkdruckprobe{})

	if fehler := werk.Oeffne(t.Context(), nil, "cardmarket",
		"https://www.cardmarket.com/"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und ein Öffner, der klagt, klagt weiter.**
func TestEinKlagenderOeffnerKlagtWeiter(t *testing.T) {
	t.Parallel()

	klage := errors.New("kein Browser")
	werk := service.NeuesLinkwerk(linkdruckprobe{})

	fehler := werk.Oeffne(t.Context(),
		func(context.Context, string) error { return klage },
		"cardmarket", "https://www.cardmarket.com/")

	if !errors.Is(fehler, klage) {
		t.Errorf("es kam %v", fehler)
	}
}
