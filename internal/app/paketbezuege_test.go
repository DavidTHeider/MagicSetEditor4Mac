package app_test

// **Die Stände der bezogenen Bäume, übersetzt** — G04
// (#2251).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type aktualisiererProbe struct {
	staende    []service.Bezugsstand
	vorschau   service.Aktualisierungsvorschau
	ergebnis   *service.Bezugsergebnis
	rueckweg   *service.Rueckweg
	fehler     error
	bestaetigt bool
}

func (p *aktualisiererProbe) Staende(context.Context, bool) ([]service.Bezugsstand, error) {
	return p.staende, p.fehler
}

func (p *aktualisiererProbe) Vorschau(context.Context, string, string) (service.Aktualisierungsvorschau, error) {
	return p.vorschau, p.fehler
}

func (p *aktualisiererProbe) Aktualisieren(_ context.Context, _, _ string, bestaetigt bool) (*service.Bezugsergebnis, error) {
	p.bestaetigt = bestaetigt

	return p.ergebnis, p.fehler
}

func (p *aktualisiererProbe) Zurueck(context.Context, string) (*service.Rueckweg, error) {
	return p.rueckweg, p.fehler
}

func mitAktualisierer(probe *aktualisiererProbe, katalog *paketfuehrer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitPaketaktualisierer(probe)
	anwendung.MitPaketen(katalog)

	return anwendung
}

func TestDieBezuegeKommenMitLageUndGrund(t *testing.T) {
	zeit := time.Date(2026, 10, 7, 2, 2, 4, 0, time.UTC)
	anwendung := mitAktualisierer(&aktualisiererProbe{staende: []service.Bezugsstand{
		{
			Name: "M15", Adresse: "https://github.com/x/y", Stand: "a", Geprueft: "a", Pakete: 3,
			Lage: service.BezugNeuer, Spitze: bezug.Spitzenstand{Commit: "b", Zeit: zeit}, Vorsprung: 63,
		},
		{Name: "Fern", Lage: service.BezugUnbekannt, Fehler: domain.ErrNichtErreichbar},
	}}, &paketfuehrer{})

	antwort := anwendung.PaketBezuege(probekontext(t), true)
	if antwort.Fehler != nil || len(antwort.Bezuege) != 2 {
		t.Fatalf("Antwort %+v", antwort)
	}

	erster := antwort.Bezuege[0]
	if erster.Lage != "neuer" || erster.Spitze != "b" || erster.SpitzeZeit != "2026-10-07T02:02:04Z" ||
		erster.Vorsprung != 63 || erster.Geprueft != "a" || erster.Pakete != 3 || erster.Fehler != nil {
		t.Errorf("erster %+v", erster)
	}

	if zweiter := antwort.Bezuege[1]; zweiter.Lage != "unbekannt" || zweiter.Fehler == nil || zweiter.SpitzeZeit != "" {
		t.Errorf("zweiter %+v", zweiter)
	}
}

func TestEinScheiternderStandIstEineLeereListeMitFehler(t *testing.T) {
	anwendung := mitAktualisierer(&aktualisiererProbe{fehler: errors.New("weg")}, &paketfuehrer{})

	if antwort := anwendung.PaketBezuege(probekontext(t), false); antwort.Fehler == nil || antwort.Bezuege == nil {
		t.Errorf("Antwort %+v", antwort)
	}

	if antwort := anwendung.PaketBezugVorschau(probekontext(t), "M15", "b"); antwort.Fehler == nil ||
		antwort.Pakete == nil || antwort.Sets == nil || antwort.Gruende == nil {
		t.Errorf("Vorschau %+v", antwort)
	}
}

func TestDieVorschauKommtGanz(t *testing.T) {
	anwendung := mitAktualisierer(&aktualisiererProbe{vorschau: service.Aktualisierungsvorschau{
		Name: "M15", Von: "a", Zu: "b", Lage: "ahead", Vorsprung: 2,
		Pakete:  []service.Paketaenderung{{Paket: "magic.mse-game", Entfernt: 1}},
		Uebrige: 4, Sets: []string{"Mein Set"}, BestaetigungNoetig: true,
		Gruende: []string{service.GrundSetVerliert}, VorigeErsetzt: true,
	}}, &paketfuehrer{})

	antwort := anwendung.PaketBezugVorschau(probekontext(t), "M15", "b")
	if antwort.Fehler != nil || len(antwort.Pakete) != 1 || antwort.Pakete[0].Entfernt != 1 ||
		antwort.Uebrige != 4 || antwort.Sets[0] != "Mein Set" || !antwort.BestaetigungNoetig ||
		antwort.Gruende[0] != "setVerliert" || !antwort.VorigeErsetzt || antwort.Zu != "b" {
		t.Errorf("Vorschau %+v", antwort)
	}
}

func TestAktualisierenUndZurueckVermerkenDieHerkunft(t *testing.T) {
	katalog := &paketfuehrer{}
	probe := &aktualisiererProbe{
		ergebnis: &service.Bezugsergebnis{Adresse: "https://github.com/x/y", Ort: "/pakete/M15", Stand: "b", Pruefsumme: "sha256:neu"},
		rueckweg: &service.Rueckweg{Ort: "/pakete/M15", Bezugsvermerk: service.Bezugsvermerk{
			Adresse: "https://github.com/x/y", Stand: "a", Pruefsumme: "sha256:alt",
		}},
	}
	anwendung := mitAktualisierer(probe, katalog)

	if antwort := anwendung.PaketBezugAktualisieren(probekontext(t), "M15", "b", true); antwort.Fehler != nil || antwort.Stand != "b" {
		t.Fatalf("Aktualisieren %+v", antwort)
	}

	if katalog.vermerkt.Herkunft != "https://github.com/x/y@b" || katalog.vermerkt.Pruefsumme != "sha256:neu" || !probe.bestaetigt {
		t.Errorf("vermerkt %+v, bestätigt %v", katalog.vermerkt, probe.bestaetigt)
	}

	if antwort := anwendung.PaketBezugZurueck(probekontext(t), "M15"); antwort.Fehler != nil || antwort.Stand != "a" {
		t.Fatalf("Zurück %+v", antwort)
	}

	if katalog.vermerkt.Herkunft != "https://github.com/x/y@a" || katalog.vermerkt.Pruefsumme != "sha256:alt" {
		t.Errorf("vermerkt %+v", katalog.vermerkt)
	}
}

func TestAktualisierenUndZurueckMeldenIhreFehler(t *testing.T) {
	anwendung := mitAktualisierer(&aktualisiererProbe{fehler: domain.ErrNotFound}, &paketfuehrer{})

	if antwort := anwendung.PaketBezugAktualisieren(probekontext(t), "M15", "b", false); antwort.Fehler == nil || antwort.Uebergangen == nil {
		t.Errorf("Aktualisieren %+v", antwort)
	}

	if antwort := anwendung.PaketBezugZurueck(probekontext(t), "M15"); antwort.Fehler == nil {
		t.Errorf("Zurück %+v", antwort)
	}

	// **Ein Fehler beim Vermerken nimmt den Tausch nicht zurück** (§7,
	// Regel 3): der Baum liegt, und gemeldet wird trotzdem.
	vermerkfehler := mitAktualisierer(&aktualisiererProbe{ergebnis: &service.Bezugsergebnis{Ort: "/pakete/M15", Stand: "b"}},
		&paketfuehrer{fehler: errors.New("der Katalog schweigt")})

	if antwort := vermerkfehler.PaketBezugAktualisieren(probekontext(t), "M15", "b", true); antwort.Fehler == nil || antwort.Ort != "/pakete/M15" {
		t.Errorf("Antwort %+v", antwort)
	}
}

func TestDieBezugsbindungenGehenDurch(t *testing.T) {
	anwendung := mitAktualisierer(&aktualisiererProbe{
		staende:  []service.Bezugsstand{{Name: "M15"}},
		vorschau: service.Aktualisierungsvorschau{Name: "M15"},
		ergebnis: &service.Bezugsergebnis{Ort: "/x", Stand: "b"},
		rueckweg: &service.Rueckweg{Ort: "/x", Bezugsvermerk: service.Bezugsvermerk{Stand: "a"}},
	}, &paketfuehrer{})
	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.PaketBezuege(probekontext(t), false); len(antwort.Bezuege) != 1 {
		t.Errorf("PaketBezuege %+v", antwort)
	}

	if antwort := bindungen.PaketBezugVorschau(probekontext(t), "M15", "b"); antwort.Name != "M15" {
		t.Errorf("PaketBezugVorschau %+v", antwort)
	}

	if antwort := bindungen.PaketBezugAktualisieren(probekontext(t), "M15", "b", false); antwort.Stand != "b" {
		t.Errorf("PaketBezugAktualisieren %+v", antwort)
	}

	if antwort := bindungen.PaketBezugZurueck(probekontext(t), "M15"); antwort.Stand != "a" {
		t.Errorf("PaketBezugZurueck %+v", antwort)
	}
}

func TestDerPaketwegNenntDieselbenBezuege(t *testing.T) {
	katalog := &paketfuehrer{}
	probe := &aktualisiererProbe{
		staende:  []service.Bezugsstand{{Name: "M15", Lage: service.BezugAktuell}},
		vorschau: service.Aktualisierungsvorschau{Name: "M15", Gruende: []string{"gekuerzt"}},
		ergebnis: &service.Bezugsergebnis{Adresse: "q", Ort: "/x", Stand: "b"},
		rueckweg: &service.Rueckweg{Ort: "/x", Bezugsvermerk: service.Bezugsvermerk{Adresse: "q", Stand: "a"}},
	}
	weg := (&app.Paketweg{}).MitVerwalter(katalog).MitAktualisierer(probe)

	if bezuege, err := weg.Bezuege(probekontext(t), false); err != nil || bezuege[0].Lage != "aktuell" {
		t.Errorf("Bezuege %+v, %v", bezuege, err)
	}

	if vorschau, err := weg.BezugVorschau(probekontext(t), "M15", "b"); err != nil || vorschau.Gruende[0] != "gekuerzt" {
		t.Errorf("Vorschau %+v, %v", vorschau, err)
	}

	if _, err := weg.BezugAktualisieren(probekontext(t), "M15", "b", true); err != nil || katalog.vermerkt.Herkunft != "q@b" {
		t.Errorf("Aktualisieren %v, vermerkt %+v", err, katalog.vermerkt)
	}

	if _, err := weg.BezugZurueck(probekontext(t), "M15"); err != nil || katalog.vermerkt.Herkunft != "q@a" {
		t.Errorf("Zurück %v, vermerkt %+v", err, katalog.vermerkt)
	}

	probe.fehler = domain.ErrNotFound

	for name, ruf := range map[string]func() error{
		"Bezuege":       func() error { _, err := weg.Bezuege(probekontext(t), true); return err },
		"Vorschau":      func() error { _, err := weg.BezugVorschau(probekontext(t), "M15", "b"); return err },
		"Aktualisieren": func() error { _, err := weg.BezugAktualisieren(probekontext(t), "M15", "b", true); return err },
		"Zurück":        func() error { _, err := weg.BezugZurueck(probekontext(t), "M15"); return err },
	} {
		if err := ruf(); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// **Ein Bezug ohne Fenster vermerkt Herkunft und Prüfsumme wie einer im
// Fenster** — gefunden vom Flächenlauf über die Bezugsstände (G04): nach
// `mse paket-beziehen` stand im Katalog keine Herkunft, und der Baum war für
// jede Frage nach seinem Stand unsichtbar. #56 sagte „vermerkt" — eingelöst
// war es nur im Fenster.
func TestEinBezugOhneFensterVermerktDieHerkunft(t *testing.T) {
	katalog := &paketfuehrer{}
	weg := (&app.Paketweg{}).MitVerwalter(katalog).MitBezieher(&bezugsprobe{ergebnis: &service.Bezugsergebnis{
		Ort: "/pakete/M15", Stand: "3.1.0", Pruefsumme: "sha256:abc",
	}})

	if _, err := weg.Beziehen(probekontext(t), "https://github.com/x/y", "3.1.0", "data", "M15"); err != nil {
		t.Fatalf("Beziehen: %v", err)
	}

	if katalog.vermerkt.Ort != "/pakete/M15" || katalog.vermerkt.Herkunft != "https://github.com/x/y@3.1.0" ||
		katalog.vermerkt.Pruefsumme != "sha256:abc" {
		t.Errorf("vermerkt: %+v", katalog.vermerkt)
	}
}
