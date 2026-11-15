package fenster

import (
	"context"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// probeablage ist eine Kartenablage, die kopiert und sonst nichts kann.
//
// **Die eingebettete Schnittstelle deckt, was nicht gebraucht wird**: das
// Einfügen scheitert an der Zwischenablage, bevor es die Ablage fragt, und
// der Lauf braucht dafür `internal/service` nicht zu kennen.
type probeablage struct{ app.Ablagequelle }

func (probeablage) Kopieren(context.Context, int64, []int64) (string, error) {
	return "abgelegt", nil
}

// probehuelle merkt sich, was die Anwendung vom Fenster wollte.
type probehuelle struct {
	gemeldet  map[string][]any
	leisten   []*application.Menu
	auswahl   []app.Dateiauswahl
	gelesen   int
	geschrieb []string
	geteilt   []string
}

func (p *probehuelle) melden(name string, daten []any) {
	if p.gemeldet == nil {
		p.gemeldet = map[string][]any{}
	}

	p.gemeldet[name] = daten
}

func (p *probehuelle) menueSetzen(leiste *application.Menu) { p.leisten = append(p.leisten, leiste) }

func (p *probehuelle) dateiWaehlen(auswahl app.Dateiauswahl) (string, error) {
	p.auswahl = append(p.auswahl, auswahl)

	return "", nil
}

func (p *probehuelle) ablageLesen() (string, error) {
	p.gelesen++

	return "", errAblageStumm
}

func (p *probehuelle) teilen(pfad, text string) error {
	p.geteilt = append(p.geteilt, pfad+"|"+text)

	return nil
}

func (p *probehuelle) ablageSchreiben(text string) error {
	p.geschrieb = append(p.geschrieb, text)

	return nil
}

// **Jede Naht der Anwendung erreicht das Fenster** — über ihren
// gewöhnlichen Weg und nicht über einen Griff ins Innere.
func TestJedeNahtErreichtDasFenster(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitKartenablage(probeablage{})

	fenster := &probehuelle{}
	mit, _ := probeumsetzer()

	verdrahten(anwendung, fenster, mit)

	// **Melden**: ein Menüeintrag meldet sein Ereignis — ohne Werte, aber als
	// Liste.
	anwendung.Fensterhandlungen().SetSichern()

	if daten, gemeldet := fenster.gemeldet[app.EreignisSetSichern]; !gemeldet || daten == nil {
		t.Errorf("gemeldet: %v", fenster.gemeldet)
	}

	// **Die Leiste**: ein Sprachwechsel setzt sie neu.
	if antwort := anwendung.SpracheSetzen("en"); antwort.Fehler != nil {
		t.Fatalf("Sprache: %+v", antwort.Fehler)
	}

	// **Zwei**: die erste beim Verdrahten, die zweite nach dem Wechsel.
	if len(fenster.leisten) != 2 || fenster.leisten[1].FindByLabel(app.MenuetexteFuer("en").Ablage) == nil {
		t.Errorf("gesetzt wurden %d Leisten, die englische fehlt", len(fenster.leisten))
	}

	// **Der Dialog**: ein Abbruch ist kein Fehler.
	if antwort := anwendung.BestandSichern(context.Background(), "", false); antwort.Fehler != nil {
		t.Errorf("ein Abbruch meldete %+v", antwort.Fehler)
	}

	if len(fenster.auswahl) != 1 || !fenster.auswahl[0].Speichern {
		t.Errorf("gefragt wurde %+v", fenster.auswahl)
	}

	// **Die Zwischenablage**: gelesen wird dort, und ihr Schweigen kommt als
	// Fehler an.
	if antwort := anwendung.AblageEinfuegen(context.Background(), "1"); antwort.Fehler == nil {
		t.Error("eine stumme Zwischenablage kam als Erfolg an")
	}

	if fenster.gelesen != 1 {
		t.Errorf("die Zwischenablage wurde %d-mal gelesen", fenster.gelesen)
	}

	// **Und geschrieben wird dort**, was die Ablage hergibt.
	if antwort := anwendung.AblageKopieren(context.Background(), "1", []string{"2"}); antwort.Fehler != nil {
		t.Errorf("das Kopieren meldete %+v", antwort.Fehler)
	}

	if len(fenster.geschrieb) != 1 || fenster.geschrieb[0] != "abgelegt" {
		t.Errorf("geschrieben wurde %q", fenster.geschrieb)
	}
}

// **Ohne Fenster gibt es keinen Dialog und keine Zwischenablage** — einen
// Fehler mit Namen, keinen Absturz.
func TestOhneFensterGibtEsEinenFehlerMitNamen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	anwendung.Fensterhandlungen().SetSichern()

	if antwort := anwendung.SpracheSetzen("en"); antwort.Fehler != nil {
		t.Errorf("ohne Fenster setzt die Leiste nichts und klagt nicht: %+v", antwort.Fehler)
	}

	if antwort := anwendung.BestandSichern(context.Background(), "", false); antwort.Fehler == nil {
		t.Error("ein Dialog ohne Fenster kam durch")
	}
}

// **Der Vorsatz der Anbindung ist der volle Name der Fassade** — so, wie
// Wails 3 ihn bildet: `<Paketpfad>.<Typ>.<Methode>` (`bindings.go`,
// v3.0.0-beta.23). Ein Umzug der Fassade in ein anderes Paket machte jede
// Bindung im Fenster zu „unbekannt", und kein Lauf im Frontend sähe es.
func TestDerVorsatzDerAnbindungIstDerDerFassade(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[app.Bindungen]()
	erwartet := typ.PkgPath() + "." + typ.Name() + "."

	quelle, err := os.ReadFile("../../frontend/src/anbindung.ts")
	if err != nil {
		t.Fatalf("anbindung.ts lesen: %v", err)
	}

	treffer := regexp.MustCompile(`export const vorsatz =\s*'([^']+)'`).FindSubmatch(quelle)
	if treffer == nil {
		t.Fatal("anbindung.ts nennt keinen Vorsatz")
	}

	if string(treffer[1]) != erwartet {
		t.Errorf("die Anbindung ruft %q, Wails bindet %q", treffer[1], erwartet)
	}
}

// **Und die Marke einer Abweisung aus dem Tor ist dieselbe** — OPS-10
// (#1908): ohne sie käme eine abgeschaltete Bindung im Fenster als Satz an
// statt als Bindungsfehler.
func TestDieFehlermarkeDerAnbindungIstDieDerAnwendung(t *testing.T) {
	t.Parallel()

	quelle, err := os.ReadFile("../../frontend/src/anbindung.ts")
	if err != nil {
		t.Fatalf("anbindung.ts lesen: %v", err)
	}

	treffer := regexp.MustCompile(`export const fehlermarke = '([^']+)'`).FindSubmatch(quelle)
	if len(treffer) < 2 || string(treffer[1]) != app.Fehlermarke {
		t.Errorf("die Anbindung sucht %q, das Tor setzt %q", treffer, app.Fehlermarke)
	}
}

// probeteilung legt eine Teilung ab, ohne `internal/service` zu kennen — die
// Hülle darf nur `app` sehen (§2).
type probeteilung struct{ ordner string }

func (p *probeteilung) Ordner() (kennung, ordner string, err error) {
	return "0123456789abcdef0123456789abcdef", p.ordner, nil
}

func (p *probeteilung) Ablegen(_, dateiname string, inhalt []byte) (pfad string, groesse int64, err error) {
	return p.ordner + "/" + dateiname, int64(len(inhalt)), nil
}

func (*probeteilung) Verwerfen(string) error { return nil }

// **Und das Teilen erreicht den Dialog des Fensters** — NET-04 (#1708).
func TestDasTeilenErreichtDenDialog(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("").MitTeilablage(&probeteilung{ordner: t.TempDir()})

	fenster := &probehuelle{}
	mit, _ := probeumsetzer()

	verdrahten(anwendung, fenster, mit)

	gut := anwendung.TeilenVorbereiten(context.Background(), "einladung", "code").Teilgut
	if gut == nil {
		t.Fatal("die Teilung kam nicht zustande")
	}

	if fehler := anwendung.TeilenAnbieten(context.Background(), gut.Kennung).Fehler; fehler != nil {
		t.Fatalf("anbieten: %+v", fehler)
	}

	if len(fenster.geteilt) != 1 || !strings.HasSuffix(fenster.geteilt[0], "/Einladung.txt|code") {
		t.Errorf("geteilt: %v", fenster.geteilt)
	}
}
