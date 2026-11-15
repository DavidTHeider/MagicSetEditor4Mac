package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// paketfuehrer ist die Naht zu `service` — eine Attrappe.
type paketfuehrer struct {
	pakete      []domain.Paket
	befunde     []domain.Befund
	fehler      error
	installiert string

	// angelegt hält fest, was angelegt wurde (#1350).
	angelegt service.Neuanlage
	entfernt string

	// listenfehler scheitert **nur** beim Lesen der Liste.
	//
	// Ein eigener Wert und nicht `fehler`: wer beides zusammenwirft, kommt gar
	// nicht bis zur Liste — das Auffrischen scheiterte vorher, und der Weg
	// dahinter bliebe ungeprüft.
	listenfehler error

	// auffrischfehler scheitert **nur** beim Auffrischen — derselbe Grund
	// (#1291):
	// wer ihn mit `fehler` zusammenwirft, kommt gar nicht bis zum Auffrischen,
	// weil die Übernahme vorher scheitert.
	auffrischfehler error

	// zusatz ist, was die Übersicht an jedes Paket hängt — DOM-01.AK-01
	// (#1696).
	zusatz service.Paketzeile

	// plan ist, was jeder Plan antwortet, und geplant hält fest, wonach
	// gefragt wurde — DOM-01.AK-02 (#1696).
	plan    service.Paketplan
	geplant string

	// bedarf ist, was der Bedarf eines Sets antwortet, und gefragtesSet
	// das Set, zu dem gefragt wurde — DOM-01.AK-04 (#1696).
	bedarf       []service.Abhaengigkeitsbefund
	gefragtesSet int64

	// gefragterKopf ist, wonach die Setvorschau fragt (FMT-03).
	gefragterKopf domain.SetHead

	// vermerkt hält fest, was nach einem Bezug in den Katalog geschrieben wurde.
	vermerkt service.Vermerk

	// uebernommen hält den Ort fest, aus dem übernommen wurde.
	uebernommen string
	uebernahme  *service.Uebernahme

	// abhaengigkeiten sind die Vorbehalte, die der Verwalter meldet.
	abhaengigkeiten []service.Abhaengigkeitsbefund
}

func (p *paketfuehrer) Abhaengigkeiten(context.Context) ([]service.Abhaengigkeitsbefund, error) {
	return p.abhaengigkeiten, p.fehler
}

func (p *paketfuehrer) Uebernehmen(_ context.Context, ort string) (*service.Uebernahme, error) {
	p.uebernommen = ort

	if p.fehler != nil {
		return nil, p.fehler
	}

	if p.uebernahme != nil {
		return p.uebernahme, nil
	}

	return &service.Uebernahme{
		Uebernommen: []domain.Paket{{Art: "mse-game", Name: "magic", Pfad: "/p/magic.mse-game"}},
		Uebergangen: []string{},
		Befunde:     []domain.Befund{},
	}, nil
}

func (p *paketfuehrer) Auffrischen(context.Context) ([]domain.Befund, error) {
	if p.auffrischfehler != nil {
		return nil, p.auffrischfehler
	}

	return p.befunde, p.fehler
}

func (p *paketfuehrer) Vermerken(_ context.Context, vermerke ...service.Vermerk) ([]domain.Befund, error) {
	if len(vermerke) > 0 {
		p.vermerkt = vermerke[0]
	}

	return p.befunde, p.fehler
}

func (p *paketfuehrer) Liste(context.Context) ([]domain.Paket, error) {
	if p.listenfehler != nil {
		return nil, p.listenfehler
	}

	return p.pakete, p.fehler
}

// Herkunftsart rechnet wie der Dienst, mit `/p` als verwaltetem Verzeichnis.
func (p *paketfuehrer) Herkunftsart(paket domain.Paket) domain.Paketherkunftsart {
	return paket.HerkunftsartIn("/p")
}

func (p *paketfuehrer) Uebersicht(context.Context) ([]service.Paketzeile, error) {
	if p.listenfehler != nil {
		return nil, p.listenfehler
	}

	zeilen := make([]service.Paketzeile, 0, len(p.pakete))
	for _, paket := range p.pakete {
		zeile := p.zusatz
		zeile.Paket = paket
		zeilen = append(zeilen, zeile)
	}

	return zeilen, p.fehler
}

func (p *paketfuehrer) Installieren(_ context.Context, quelle string) (domain.Paket, error) {
	p.installiert = quelle

	if p.fehler != nil {
		return domain.Paket{}, p.fehler
	}

	return domain.Paket{Art: "mse-style", Name: "neu", Pfad: "/p/neu.mse-style", Herkunft: quelle}, nil
}

// Kopieren gibt die Kopie unter dem gewünschten Namen — FMT-02 (#1691).
func (p *paketfuehrer) Kopieren(_ context.Context, art, _, neu string) (domain.Paket, error) {
	return domain.Paket{Art: art, Name: neu}, p.fehler
}

func (p *paketfuehrer) Weitergeben(_ context.Context, _ []string, ziel string) (string, error) {
	return ziel, p.fehler
}

func (p *paketfuehrer) Einlesen(context.Context, string) (*service.Uebernahme, error) {
	return &service.Uebernahme{}, p.fehler
}

// Anlegen schreibt ein neues Paket (#1350).
func (p *paketfuehrer) Anlegen(_ context.Context, wunsch service.Neuanlage) (
	domain.Paket, error,
) {
	p.angelegt = wunsch

	if p.fehler != nil {
		return domain.Paket{}, p.fehler
	}

	return domain.Paket{Art: wunsch.Art, Name: wunsch.Name}, nil
}

func (p *paketfuehrer) Entfernen(_ context.Context, art, name, _ string) error {
	p.entfernt = art + "/" + name

	return p.fehler
}

// **Der Katalog kommt an die Oberfläche — mit Herkunft.**
func TestPaketeGibtDenKatalog(t *testing.T) {
	fuehrer := &paketfuehrer{pakete: []domain.Paket{{
		Art: "mse-style", Name: "probe-schlicht", Version: "2026-08-15",
		Pfad: "/p/probe-schlicht.mse-style", Herkunft: "aus der Saat",
		Braucht: []domain.Bedarf{{Paket: "probe.mse-game", Version: "2026-08-15"}},
	}}}

	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Pakete) != 1 {
		t.Fatalf("%d Pakete", len(antwort.Pakete))
	}

	paket := antwort.Pakete[0]
	if paket.Name != "probe-schlicht" || paket.Herkunft != "aus der Saat" {
		t.Errorf("Paket: %+v", paket)
	}

	if len(paket.Braucht) != 1 || paket.Braucht[0] != "probe.mse-game 2026-08-15" {
		t.Errorf("Abhängigkeiten: %v", paket.Braucht)
	}
}

// **Leere Listen bleiben Listen.**
//
// Ein `null` statt `[]` ist die Fehlerklasse, die zwischen den Prüfebenen liegt
// (§3): die Go-Seite sieht die Struct-Marken nicht, die Frontend-Vorrichtung
// nie die Serialisierung — und die Seite bliebe leer.
func TestEinLeererKatalogGibtLeereListen(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{})

	antwort := anwendung.Pakete(t.Context())

	if antwort.Pakete == nil || antwort.Befunde == nil {
		t.Errorf("nil statt leerer Liste: %+v", antwort)
	}
}

// **Befunde des Abgleichs kommen mit, und ein Fehler beendet ihn nicht.**
//
// Ein unlesbares Paketverzeichnis ist ein Befund (§8) — der Katalog daneben
// bleibt brauchbar.
func TestBefundeDesAbgleichsKommenAn(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{
		befunde: []domain.Befund{{Datei: "kaputt.mse-style", Meldung: "nicht lesbar"}},
		pakete:  []domain.Paket{{Art: "mse-game", Name: "probe", Pfad: "/p"}},
	})

	antwort := anwendung.Pakete(t.Context())

	if len(antwort.Befunde) != 1 {
		t.Fatalf("Befunde: %v", antwort.Befunde)
	}

	if len(antwort.Pakete) != 1 {
		t.Errorf("der Katalog fehlt: %+v", antwort.Pakete)
	}
}

// **Ein Fehler aus dem Bestand wird übersetzt.**
func TestEinFehlerImKatalogWirdUebersetzt(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{fehler: errors.New("die Datenbank schweigt")})

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Pakete == nil || antwort.Befunde == nil {
		t.Errorf("nil statt leerer Liste: %+v", antwort)
	}
}

// **Ein leerer Pfad heißt „abgebrochen" und nicht „Fehler".**
//
// Wer den Dateidialog schließt, hat nichts gewählt — eine Fehlermeldung dafür
// wäre eine Meldung über etwas, das niemand versucht hat.
func TestInstallierenOhnePfadIstEinAbbruch(t *testing.T) {
	fuehrer := &paketfuehrer{}
	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.PaketInstallieren(t.Context(), "")

	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("Antwort: %+v", antwort)
	}

	if fuehrer.installiert != "" {
		t.Errorf("trotzdem installiert: %q", fuehrer.installiert)
	}
}

// **Was installiert wurde, steht in der Antwort.**
func TestInstallierenGibtDasPaketZurueck(t *testing.T) {
	fuehrer := &paketfuehrer{}
	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.PaketInstallieren(t.Context(), "/quelle/neu.mse-style")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Paket == nil || antwort.Paket.Name != "neu" {
		t.Fatalf("Paket: %+v", antwort.Paket)
	}

	if fuehrer.installiert != "/quelle/neu.mse-style" {
		t.Errorf("installiert: %q", fuehrer.installiert)
	}
}

// **Entfernen reicht Art und Name durch und übersetzt den Fehler.**
func TestEntfernenReichtDurch(t *testing.T) {
	fuehrer := &paketfuehrer{}
	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	if antwort := anwendung.PaketEntfernen(t.Context(), "mse-style", "weg"); antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if fuehrer.entfernt != "mse-style/weg" {
		t.Errorf("entfernt: %q", fuehrer.entfernt)
	}

	kaputt := &app.App{}
	kaputt.MitPaketen(&paketfuehrer{fehler: domain.ErrNotFound})

	if antwort := kaputt.PaketEntfernen(probekontext(t), "mse-style", "weg"); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}
