package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein gezeichnetes Symbol geht ins Set** —
// #890,
// Werkstück 4 von Phase X.
//
// **Es liegt im Bildverzeichnis des Sets**, und das ist kein Zufall: `Beiwerk`
// kopiert beim Sichern **alles** von dort ins geschriebene Paket
// (`ausAblageKopieren`). Ein eigenes Verzeichnis daneben käme nie mit, und der
// Fehler fiele erst auf, wenn jemand die Datei woanders öffnet.
func TestEinGezeichnetesSymbolLandetImBildverzeichnis(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueSymbolablage(service.NeueBildablage(wurzel))

	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: "Raute", Kombination: msepaket.KombinationVereinen,
		Punkte: []msepaket.Kontrollpunkt{
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.05}},
			{Stelle: msepaket.Vektor{X: 0.95, Y: 0.5}},
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0.95}},
		},
	}}}

	name, err := ablage.Ablegen(7, zeichnung, "")
	if err != nil {
		t.Fatalf("Ablegen: %v", err)
	}

	if !strings.HasSuffix(name, ".mse-symbol") {
		t.Errorf("die Datei heißt %q — MSE erkennt eine Zeichnung an ihrer Endung", name)
	}

	// #nosec G304 -- `wurzel` ist ein `t.TempDir`, `name` kommt aus der Ablage
	roh, lesefehler := os.ReadFile(filepath.Join(wurzel, "7", name))
	if lesefehler != nil {
		t.Fatalf("die abgelegte Datei: %v", lesefehler)
	}

	// **Und sie ist eine echte `.mse-symbol`** — kein Zwischenformat.
	for _, will := range []string{"mse_version: 0.3.5", "type: shape", "name: Raute"} {
		if !strings.Contains(string(roh), will) {
			t.Errorf("%q fehlt in\n%s", will, roh)
		}
	}
}

// **Und sie kommt zurück, wie sie hineinging.**
func TestEinAbgelegtesSymbolLiestSichZurueck(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueSymbolablage(service.NeueBildablage(wurzel))

	will := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art: msepaket.TeilForm, Name: "Bogen", Kombination: msepaket.KombinationAbziehen,
		Punkte: []msepaket.Kontrollpunkt{
			{
				Stelle: msepaket.Vektor{X: 0.25, Y: 0.5}, GriffDanach: msepaket.Vektor{X: 0.1, Y: 0},
				GriffDavor: msepaket.Vektor{X: -0.1, Y: 0}, KurveDanach: true, KurveDavor: true,
				Sperre: "size",
			},
			{
				Stelle: msepaket.Vektor{X: 0.75, Y: 0.5}, GriffDavor: msepaket.Vektor{X: -0.1, Y: 0},
				GriffDanach: msepaket.Vektor{X: 0.1, Y: 0}, KurveDavor: true, KurveDanach: true,
				Sperre: "size",
			},
		},
	}}}

	name, err := ablage.Ablegen(7, will, "")
	if err != nil {
		t.Fatalf("Ablegen: %v", err)
	}

	hat, lesefehler := ablage.Gestalt(7, name)
	if lesefehler != nil {
		t.Fatalf("Gestalt: %v", lesefehler)
	}

	if len(hat.Teile) != 1 || hat.Teile[0].Name != "Bogen" {
		t.Fatalf("die Zeichnung ist %+v", hat.Teile)
	}

	if len(hat.Teile[0].Punkte) != 2 || hat.Teile[0].Punkte[0].Sperre != "size" {
		t.Errorf("die Punkte sind %+v", hat.Teile[0].Punkte)
	}
}

// **Zwei Symbole überschreiben sich nicht.**
//
// Derselbe Grund wie bei den Bildern: der Zähler läuft je Verzeichnis, und ein
// zweites Symbol mit demselben Namen nähme dem ersten seinen Platz.
func TestZweiSymboleBekommenZweiNamen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))

	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{Art: msepaket.TeilForm}}}

	erster, err := ablage.Ablegen(7, zeichnung, "")
	if err != nil {
		t.Fatalf("Ablegen: %v", err)
	}

	zweiter, err := ablage.Ablegen(7, zeichnung, "")
	if err != nil {
		t.Fatalf("Ablegen: %v", err)
	}

	if erster == zweiter {
		t.Errorf("beide heißen %q", erster)
	}
}

// **Ohne Verzeichnis wird nichts abgelegt, und es steht dabei, warum.**
func TestOhneVerzeichnisWirdKeinSymbolAbgelegt(t *testing.T) {
	t.Parallel()

	ablage := service.NeueSymbolablage(service.NeueBildablage(""))

	if _, err := ablage.Ablegen(7, msepaket.Symbolzeichnung{}, ""); err == nil {
		t.Error("das Ablegen ohne Verzeichnis ging")
	}

	if _, err := ablage.Gestalt(7, "irgendwas.mse-symbol"); err == nil {
		t.Error("das Lesen ohne Verzeichnis ging")
	}
}

// **Und ein Name, der aus dem Verzeichnis führt, wird abgewiesen.**
//
// Der Name kommt über eine Naht von außen — aus einem Feldwert, den eine
// fremde Datei trägt. `../../../etc/passwd` ist kein Symbol.
func TestEinNameDerHinausfuehrtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))

	for _, name := range []string{"../anderes.mse-symbol", "a/b.mse-symbol", ""} {
		if _, err := ablage.Gestalt(7, name); err == nil {
			t.Errorf("%q kam durch", name)
		}
	}
}

// **Die Fehlerwege der Ablage sind betreten** (§3, §10).
//
// **Sie sind keine Zierde.** Ein Verzeichnis, das eine Datei ist; ein Name, der
// schon zehntausendmal vergeben ist; eine Datei, die keine Zeichnung ist — das
// sind Zustände, die eine Platte wirklich annimmt, und ein Zweig, den kein Lauf
// betritt, ist eine Behauptung.
func TestDieAblageMeldetIhreFehlerwege(t *testing.T) {
	t.Parallel()

	t.Run("das Verzeichnis lässt sich nicht anlegen", func(t *testing.T) {
		t.Parallel()

		wurzel := t.TempDir()

		// **Eine Datei, wo das Verzeichnis hin soll.**
		if err := os.WriteFile(filepath.Join(wurzel, "7"), []byte("keins"), 0o600); err != nil {
			t.Fatalf("Sperre anlegen: %v", err)
		}

		ablage := service.NeueSymbolablage(service.NeueBildablage(wurzel))

		if _, err := ablage.Ablegen(7, msepaket.Symbolzeichnung{}, ""); err == nil {
			t.Error("das Ablegen in eine Datei ging")
		}
	})

	t.Run("die Datei lässt sich nicht schreiben", func(t *testing.T) {
		t.Parallel()

		wurzel := t.TempDir()
		ort := filepath.Join(wurzel, "7")

		if err := os.MkdirAll(ort, 0o500); err != nil {
			t.Fatalf("Verzeichnis anlegen: %v", err)
		}

		ablage := service.NeueSymbolablage(service.NeueBildablage(wurzel))

		if _, err := ablage.Ablegen(7, msepaket.Symbolzeichnung{}, ""); err == nil {
			t.Error("das Schreiben in ein schreibgeschütztes Verzeichnis ging")
		}
	})

	// **Da, aber nicht lesbar** — und damit nicht „nicht gefunden": hier
	// stand bis C01 eine Datei, die es nicht gibt, und seit sie „nicht
	// gefunden" heißt (#2226), prüfte der Fall nicht mehr, was sein Name sagt.
	t.Run("die Datei ist keine Zeichnung", func(t *testing.T) {
		t.Parallel()

		wurzel := t.TempDir()
		bilder := service.NeueBildablage(wurzel)
		ablage := service.NeueSymbolablage(bilder)

		if err := os.MkdirAll(filepath.Join(bilder.Verzeichnis(7), "ordner.mse-symbol"), 0o700); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		_, err := ablage.Gestalt(7, "ordner.mse-symbol")
		if err == nil || errors.Is(err, domain.ErrNotFound) {
			t.Errorf("ein Verzeichnis statt einer Zeichnung gab %v — erwartet ein Lesefehler, der nicht „nicht gefunden“ ist", err)
		}
	})

	t.Run("ohne Bildablage", func(t *testing.T) {
		t.Parallel()

		ablage := service.NeueSymbolablage(nil)

		if _, err := ablage.Ablegen(7, msepaket.Symbolzeichnung{}, ""); err == nil {
			t.Error("das Ablegen ohne Ablage ging")
		}
	})
}

// **Ein zweites Ablegen unter demselben Namen überschreibt.**
//
// # Der Befund
//
// Bis zum 2026-08-25 gab `Ablegen` **jedes Mal** einen freien Namen. Wer sein
// Symbol zeichnete, ins Set legte, weiterzeichnete und wieder ins Set legte,
// bekam `symbol2.mse-symbol`; `set.symbol` zeigte darauf, und `symbol1`
// blieb liegen. `ausAklageKopieren` kopiert beim Sichern **jede** Datei des
// Ablageverzeichnisses — die verwaiste ging als Beiwerk mit ins Set.
//
// Angestrichen vom Nachlauf des Zweitblicks über Phase X, Befund 1 (#890).
func TestEinZweitesAblegenNimmtDenselbenNamen(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	ablage := service.NeueSymbolablage(service.NeueBildablage(ort))

	eine := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{Art: msepaket.TeilForm}}}

	erster, err := ablage.Ablegen(7, eine, "")
	if err != nil {
		t.Fatalf("Ablegen: %v", err)
	}

	// **Dieselbe Ablage, diesmal unter dem Namen von vorhin.**
	andere := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{
		{Art: msepaket.TeilForm, Name: "Zweite"},
	}}

	zweiter, err := ablage.Ablegen(7, andere, erster)
	if err != nil {
		t.Fatalf("zweites Ablegen: %v", err)
	}

	if zweiter != erster {
		t.Errorf("das zweite Ablegen gab %q statt %q", zweiter, erster)
	}

	// **Und es liegt wirklich nur eine Datei da.** Ein gleicher Rückgabewert
	// bei zwei Dateien wäre dieselbe Verwaisung mit besserer Auskunft.
	gefunden, lesefehler := filepath.Glob(filepath.Join(
		service.NeueBildablage(ort).Verzeichnis(7), "*.mse-symbol"))
	if lesefehler != nil {
		t.Fatalf("Verzeichnis lesen: %v", lesefehler)
	}

	if len(gefunden) != 1 {
		t.Errorf("im Verzeichnis liegen %d Dateien: %v", len(gefunden), gefunden)
	}

	zurueck, gestaltfehler := ablage.Gestalt(7, zweiter)
	if gestaltfehler != nil {
		t.Fatalf("Gestalt: %v", gestaltfehler)
	}

	if len(zurueck.Teile) != 1 || zurueck.Teile[0].Name != "Zweite" {
		t.Errorf("gelesen wurde %+v — die zweite Zeichnung steht nicht da", zurueck.Teile)
	}
}

// **Ein Name, der aus dem Verzeichnis führt, wird auch beim Ablegen
// abgewiesen.**
//
// `Gestalt` prüft ihn seit jeher; `Ablegen` bekam ihn erst mit dem Befund
// oben. Ein Feldwert aus einer fremden Datei erreicht damit beide Wege (§8).
func TestEinNameDerHinausfuehrtWirdAuchBeimAblegenAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))
	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{Art: msepaket.TeilForm}}}

	for _, name := range []string{"../fremd.mse-symbol", "unter/fremd.mse-symbol", ".."} {
		if _, err := ablage.Ablegen(7, zeichnung, name); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%q gab %v", name, err)
		}
	}

	// **Leerraum ist kein Name, sondern keiner.** Ein Feldwert, der nur aus
	// Leerzeichen besteht, heißt „dieses Set hat noch kein Symbol" — er wird
	// nicht abgewiesen, sondern führt zu einem frischen Namen. `Gestalt` weist
	// ihn ab, und das ist kein Widerspruch: dort ist er eine Frage nach einer
	// Datei, hier eine Aussage über keine.
	frisch, err := ablage.Ablegen(7, zeichnung, "  ")
	if err != nil || frisch == "" {
		t.Errorf("Leerraum gab %q und %v", frisch, err)
	}
}

// **Ein unlesbares Bildverzeichnis wird gemeldet.**
//
// # Warum dieser Lauf hier steht und nicht bei den Bildern
//
// `freierName` sucht mit `os.Stat`, und seit Phase 3.2 benutzt ihn nur noch
// die Symbolablage: die **Bilder** belegen ihren Namen beim Anlegen, weil zwei
// gleichzeitige Einsetzungen sonst denselben bekämen
// (#1163).
//
// Ein Symbol trägt seinen Inhalt im **Namen** — dieselbe Zeichnung ergibt
// dieselbe Datei —, und dort ist eine Momentaufnahme genug.
func TestEinUnlesbaresBildverzeichnisWirdGemeldet(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root greift keine Rechteprüfung")
	}

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	ort := ablage.Verzeichnis(7)
	if err := os.MkdirAll(ort, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Ohne Ausführrecht**: `os.Stat` auf einen Eintrag darin scheitert mit
	// `EACCES` — und das ist etwas anderes als „gibt es nicht".
	if err := os.Chmod(ort, 0o600); err != nil {
		t.Fatalf("zusperren: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck // Aufräumen

	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{Art: msepaket.TeilForm}}}

	if _, fehler := service.NeueSymbolablage(ablage).Ablegen(7, zeichnung, ""); fehler == nil {
		t.Error("das unlesbare Verzeichnis kam durch")
	}
}

// **Auch ein gezeichnetes Symbol hält die Grenzen** — DOM-02.AK-04 (#1697):
// was sich nicht einlesen ließe, lässt sich auch nicht ablegen.
func TestEinUeberkomplexesSymbolWirdNichtAbgelegt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueSymbolablage(service.NeueBildablage(wurzel))

	viel := msepaket.Symbolzeichnung{}
	for range msepaket.GrenzeSymbolteile + 1 {
		viel.Teile = append(viel.Teile, msepaket.Symbolteil{Art: msepaket.TeilForm})
	}

	var feldfehler *domain.FieldError
	if _, err := ablage.Ablegen(7, viel, ""); !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselSymbolZuVieleTeile {
		t.Errorf("zu viele Teile: %v", err)
	}

	if eintraege, err := os.ReadDir(filepath.Join(wurzel, "7")); err == nil && len(eintraege) != 0 {
		t.Errorf("es liegt doch etwas da: %v", eintraege)
	}
}

// **Ein Symbol, das nicht neben dem Set liegt, ist „nicht gefunden"** — C01
// (#2226). Der Vertragsrundgang rief `Symbol` mit einem Namen, den es nicht
// gibt, und bekam „Es ist ein unerwarteter Fehler aufgetreten": der Lesefehler
// der Datei ging roh hinaus.
func TestEinFehlendesSymbolIstNichtGefunden(t *testing.T) {
	t.Parallel()

	ablage := service.NeueSymbolablage(service.NeueBildablage(t.TempDir()))

	if _, fehler := ablage.Gestalt(1, "gibtsnicht.mse-symbol"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Symbol gab %v, erwartet ErrNotFound", fehler)
	}
}
