package app

// **Die Fehlerwege der Sitzung am Fenster** — ID-02 (#1701).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sitzungsablagedoppel kennt einen Nutzer, bis man ihn wegnimmt.
type sitzungsablagedoppel struct {
	nutzer     []domain.Nutzer
	lesefehler error

	// anmeldefehler trifft das Geschäft aus Zeigern und Zustand.
	anmeldefehler error
}

func (d *sitzungsablagedoppel) Nutzer(context.Context) ([]domain.Nutzer, error) {
	return d.nutzer, nil
}

func (d *sitzungsablagedoppel) KennwortSetzen(context.Context, int64, string) error { return nil }

func (d *sitzungsablagedoppel) SitzungLesen(context.Context) (domain.Sitzungszustand, error) {
	return domain.SitzungAngemeldet, d.lesefehler
}

func (d *sitzungsablagedoppel) SitzungFesthalten(context.Context, domain.Sitzungszustand) error {
	return nil
}

func (d *sitzungsablagedoppel) KennwortMitCodeSetzen(context.Context, int64, int64, string, domain.Sitzungszustand) error {
	return nil
}

func (d *sitzungsablagedoppel) AnmeldungFesthalten(context.Context, int64, int64) error {
	return d.anmeldefehler
}

// werdoppel sagt, wer arbeitet — oder scheitert daran.
type werdoppel struct {
	Profilwechsler
	fehler error
}

func (w werdoppel) WerArbeitet(context.Context) (int64, error) { return 1, w.fehler }

// **Ein Fehler beim Aufnehmen schließt die Sitzung** (Zweitblick über ID-02,
// Befund 1) — abgemeldet und nicht offen.
func TestEineSitzungDieNichtBeginntBleibtAbgemeldet(t *testing.T) {
	stolpert := errors.New("der Zeiger klemmt")

	for name, fall := range map[string]struct {
		wer    error
		ablage *sitzungsablagedoppel
	}{
		"wer arbeitet":         {wer: stolpert, ablage: &sitzungsablagedoppel{}},
		"festgehaltener Stand": {ablage: &sitzungsablagedoppel{lesefehler: stolpert}},
	} {
		anwendung := NeueApp("")
		anwendung.profilwechsel = werdoppel{fehler: fall.wer}
		anwendung.sitzungBeginnen(t.Context(), fall.ablage)

		sitzung := anwendung.sitzung.Load()
		if sitzung == nil {
			t.Fatalf("%s: keine Sitzung", name)
		}

		if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAbgemeldet {
			t.Errorf("%s: %s", name, zustand)
		}

		if !anwendung.sitzungVerschlossen() {
			t.Errorf("%s: offen", name)
		}
	}
}

// **Ein Nutzer, den es nicht mehr gibt**, gibt keine Auskunft — der Stand
// sagt es mit.
func TestDerStandMeldetEineFehlendeAuskunft(t *testing.T) {
	ablage := &sitzungsablagedoppel{nutzer: []domain.Nutzer{{ID: 1, Name: "Ich"}}}
	anwendung := NeueApp("")
	anwendung.profilwechsel = werdoppel{}
	anwendung.faktor = service.NeuerZweiterfaktor(nil, nil, nil)
	anwendung.sitzungBeginnen(t.Context(), ablage)

	ablage.nutzer = nil

	if stand := anwendung.Sitzungsstand(t.Context()); stand.Fehler == nil || stand.Zustand != "angemeldet" {
		t.Errorf("%+v", stand)
	}
}

// **Ein unlesbarer festgehaltener Stand** hält auch die Kommandozeile an.
func TestDieFesteSitzungBrauchtIhrenStand(t *testing.T) {
	anwendung := NeueApp("")
	anwendung.profilwechsel = werdoppel{}
	anwendung.sitzungsablage = &sitzungsablagedoppel{lesefehler: errors.New("der Stand klemmt")}

	if _, fehler := anwendung.festeSitzung(t.Context()); fehler == nil {
		t.Error("ohne Stand begonnen")
	}
}

// **Sperren und Abmelden beenden, was am Profil hängt** (Zweitblick über
// ID-02, Befund 6) — dieselben Schritte wie ein Profilwechsel; die Vorschauen
// stehen hier für alle, und der Profilwechsel prüft die übrigen.
func TestSperrenUndAbmeldenRaeumenWieEinWechsel(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, nil)

	if antwort := anwendung.KennwortWechseln(t.Context(), "", "ein-langes-kennwort"); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	// **Der Reihe nach**: gesperrt lässt sich abmelden, abgemeldet nicht
	// sperren.
	for _, fall := range []struct {
		name    string
		schritt func() SchreibAntwort
	}{
		{"sperren", func() SchreibAntwort { return anwendung.Sperren(t.Context()) }},
		{"abmelden", func() SchreibAntwort { return anwendung.Abmelden(t.Context()) }},
	} {
		name, schritt := fall.name, fall.schritt
		anwendung.bildfach.Hinterlege([]byte("vorschau"))

		if antwort := schritt(); antwort.Fehler != nil {
			t.Fatalf("%s: %v", name, antwort.Fehler)
		}

		if zahl := anwendung.bildfach.Zahlen().Eintraege; zahl != 0 {
			t.Errorf("%s: %d Vorschauen stehen noch", name, zahl)
		}
	}
}

// faktorstub lässt jeden Faktor gelten — hier geht es um die Umstellung.
type faktorstub struct{}

func (faktorstub) SensorMoeglich(context.Context, int64) (bool, error) { return false, nil }
func (faktorstub) SensorPruefen(context.Context, int64) error          { return nil }
func (faktorstub) CodeFinden(context.Context, int64, string) (int64, error) {
	return 0, nil
}
func (faktorstub) Pruefe(context.Context, int64, string) error { return nil }

// umstelldoppel scheitert daran, das Profil der Anmeldung zu nennen.
type umstelldoppel struct {
	Profilwechsler
	umstellfehler error
}

func (u umstelldoppel) Anmeldeprofil(context.Context, int64) (int64, error) {
	return 1, u.umstellfehler
}

// **Scheitert die Umstellung nach dem Anmelden, bleibt abgemeldet** — die
// Sitzung geht erst auf, wenn Zeiger und Zustand im Bestand stehen.
func TestEineGescheiterteUmstellungLaesstAbgemeldet(t *testing.T) {
	stolpert := errors.New("der Zeiger klemmt")

	for name, fall := range map[string]struct {
		profil, festhalten error
	}{
		"das Profil":     {profil: stolpert},
		"das Festhalten": {festhalten: stolpert},
	} {
		anwendung := NeueApp("")
		anwendung.profilwechsel = umstelldoppel{umstellfehler: fall.profil}
		sitzung := service.NeueSitzung(&sitzungsablagedoppel{
			nutzer: []domain.Nutzer{{ID: 1, Name: "Ich"}}, anmeldefehler: fall.festhalten,
		}, faktorstub{})
		anwendung.sitzung.Store(sitzung)

		if antwort := anwendung.LokalAnmelden(t.Context(), "1", "", ""); antwort.Fehler == nil {
			t.Errorf("%s: angemeldet trotz Fehler", name)
		}

		if zustand, _ := sitzung.Stand(); zustand != domain.SitzungAbgemeldet {
			t.Errorf("%s: %s", name, zustand)
		}
	}
}
