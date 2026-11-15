package app_test

// **Ein Set vorab ansehen** — FMT-03.AK-01 (#1692): was das Einfahren täte,
// samt dem, was an Paketen fehlt, und ohne zu schreiben.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func (i *importprobe) Vorschau(_ context.Context, pfad string) (*service.Setvorschau, error) {
	i.pfad = pfad

	return i.vorschau, i.fehler
}

func (p *paketfuehrer) Kopfbedarf(_ context.Context, kopf domain.SetHead) ([]service.Abhaengigkeitsbefund, error) {
	p.gefragterKopf = kopf

	return p.bedarf, p.fehler
}

func vorschauanwendung(probe *importprobe, pakete *paketfuehrer) *app.App {
	anwendung := mitImporteur(probe)
	anwendung.MitPaketen(pakete)

	return anwendung
}

func TestDieSetvorschauNenntInhaltUndBedarf(t *testing.T) {
	probe := &importprobe{vorschau: &service.Setvorschau{
		Name: "Mein Probeset", Spiel: "probe", Stil: "probe-stil", Karten: 3, SchonOffen: 7,
		Befunde: []domain.Befund{{Datei: "set", Ort: "12", Meldung: "unlesbar"}},
	}}
	pakete := &paketfuehrer{bedarf: []service.Abhaengigkeitsbefund{
		{Paket: "Mein Probeset", Braucht: "probe.mse-game", Art: service.AbhaengigkeitFehlt},
	}}

	antwort := vorschauanwendung(probe, pakete).SetVorschau(probekontext(t), "/tmp/probe.mse-set")
	if antwort.Fehler != nil || antwort.Vorschau == nil {
		t.Fatalf("Antwort: %+v", antwort)
	}

	vorschau := antwort.Vorschau
	if vorschau.Pfad != "/tmp/probe.mse-set" || vorschau.Name != "Mein Probeset" || vorschau.Karten != 3 ||
		vorschau.SchonOffen != "7" || len(vorschau.Befunde) != 1 || len(vorschau.Bedarf) != 1 ||
		vorschau.Bedarf[0].Braucht != "probe.mse-game" {
		t.Errorf("Vorschau: %+v", vorschau)
	}

	// **Der Bedarf wird am Kopf aus der Datei gefragt** — Spiel und Stil.
	if pakete.gefragterKopf.Game != "probe" || pakete.gefragterKopf.Style != "probe-stil" ||
		pakete.gefragterKopf.Name != "Mein Probeset" {
		t.Errorf("gefragt: %+v", pakete.gefragterKopf)
	}
}

func TestEineNeueSetvorschauNenntKeinSchonOffenes(t *testing.T) {
	probe := &importprobe{vorschau: &service.Setvorschau{Name: "Neu", Spiel: "probe"}}

	antwort := vorschauanwendung(probe, &paketfuehrer{}).SetVorschau(probekontext(t), "/tmp/neu.mse-set")
	if antwort.Vorschau == nil || antwort.Vorschau.SchonOffen != "" ||
		antwort.Vorschau.Befunde == nil || antwort.Vorschau.Bedarf == nil {
		t.Errorf("Vorschau: %+v", antwort.Vorschau)
	}
}

func TestEineSetvorschauOhnePfadIstAbgebrochen(t *testing.T) {
	if antwort := vorschauanwendung(&importprobe{}, &paketfuehrer{}).
		SetVorschau(probekontext(t), ""); !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("Antwort: %+v", antwort)
	}
}

func TestEineSetvorschauMeldetIhreFehler(t *testing.T) {
	klemmt := errors.New("klemmt")

	for _, fall := range []struct {
		name   string
		probe  *importprobe
		pakete *paketfuehrer
	}{
		{"die Datei", &importprobe{fehler: klemmt}, &paketfuehrer{}},
		{"der Bedarf", &importprobe{vorschau: &service.Setvorschau{Name: "X"}}, &paketfuehrer{fehler: klemmt}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			antwort := vorschauanwendung(fall.probe, fall.pakete).SetVorschau(probekontext(t), "/tmp/x.mse-set")
			if antwort.Fehler == nil || antwort.Vorschau != nil {
				t.Errorf("Antwort: %+v", antwort)
			}
		})
	}
}

// **Die Flächen ohne Fenster lesen mit dem echten Importeur** — über dem
// Bestand, in den das Einfahren schriebe, und ohne ihn zu ändern.
func TestDerPaketwegZeigtDieSetvorschau(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	weg, nahtfehler := app.PaketnahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	// **Im gezeigten Profil**, wie jede Wurzel ohne Fenster ihren Ruf
	// ausstattet — „eine Quelle, ein Set" gilt je Profil.
	imProfil, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatal(profilfehler)
	}

	vorschau, err := weg.Setvorschau(imProfil, db, "../../testdaten/probe.mse-set")
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	// **Ohne eingebundene Pakete fehlen Spiel und Stil** — dieselbe Auskunft,
	// die `Paketbedarf` nach dem Einfahren gäbe.
	if vorschau.Name != "Mein Probeset" || vorschau.Karten != 3 || vorschau.SchonOffen != "" ||
		len(vorschau.Bedarf) != 2 || vorschau.Bedarf[0].Braucht != "probe.mse-game" {
		t.Errorf("Vorschau: %+v", vorschau)
	}

	var sets int
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM sets").Scan(&sets); err != nil || sets != 0 {
		t.Errorf("die Vorschau hat %d Sets angelegt (%v)", sets, err)
	}
}

// **Am Fenster fragt der Dialog nach der Datei** — derselbe wie bei „Set
// öffnen …", und ein abgebrochener Dialog ist keine Vorschau.
func TestDieSetvorschauAmFensterFragtNachDerDatei(t *testing.T) {
	stand := bindungsprobe(t)
	anwendung, bindungen := stand.Anwendung, stand.Bindungen

	probe := &importprobe{vorschau: &service.Setvorschau{Name: "Gewählt", Karten: 2}}
	anwendung.MitImporteur(probe)
	anwendung.MitPaketen(&paketfuehrer{})

	var gefragt app.Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gefragt = auswahl

		return "/tmp/gewaehlt.mse-set", nil
	})

	antwort := bindungen.SetVorschauWaehlen(t.Context())
	if antwort.Vorschau == nil || antwort.Vorschau.Name != "Gewählt" || probe.pfad != "/tmp/gewaehlt.mse-set" ||
		gefragt.Muster != "*.mse-set" {
		t.Errorf("Antwort %+v, gefragt %+v", antwort, gefragt)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return "", nil })

	if antwort := bindungen.SetVorschauWaehlen(t.Context()); !antwort.Abgebrochen {
		t.Errorf("abgebrochen: %+v", antwort)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("der Dialog ging nicht auf")
	})

	if antwort := bindungen.SetVorschauWaehlen(t.Context()); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}
