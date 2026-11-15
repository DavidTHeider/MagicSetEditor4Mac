package app

// **Der Umzug der Vorgaben nach dem Zweitblick über Phase 4** — Befunde 5
// und 6 (#2424).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// huelle macht eine [aufzeichnung] für zwei Goroutinen sicher und lässt
// nach jedem Befehl etwas geschehen.
type huelle struct {
	schranke sync.Mutex
	buch     *aufzeichnung
	nachher  func(befehl []string)

	// abweisen lässt einen Befehl scheitern, den ein Vorsatz nicht trifft.
	abweisen func(befehl []string) bool
}

func (h *huelle) ausfuehren(ctx context.Context, befehl []string) error {
	h.schranke.Lock()
	defer h.schranke.Unlock()

	if h.abweisen != nil && h.abweisen(befehl) {
		h.buch.befehle = append(h.buch.befehle, befehl)

		return errors.New("abgelehnt")
	}

	fehler := h.buch.ausfuehren(ctx, befehl)
	if h.nachher != nil {
		h.nachher(befehl)
	}

	return fehler
}

func (h *huelle) unter(tun func()) {
	h.schranke.Lock()
	defer h.schranke.Unlock()

	tun()
}

// mitAlterDomaene ist ein Umzug über einer alten Domäne mit `a` und `fenster`.
func mitAlterDomaene(t *testing.T) (kennungsumzug, *aufzeichnung, *huelle) {
	t.Helper()

	umzug, buch := umzugIn(t)
	buch.domaenen = map[string]bool{altkennungen[0]: true}
	buch.inhalte = map[string]string{altkennungen[0]: plistMit("alt", "alt")}
	h := &huelle{buch: buch}
	umzug.ausfuehren = h.ausfuehren

	return umzug, buch, h
}

func importiert(buch *aufzeichnung) bool {
	for _, befehl := range buch.nur("defaults") {
		if befehl[1] == "import" {
			return true
		}
	}

	return false
}

// **Ein Lesefehler der neuen Domäne gilt nicht als „leer"** (Befund 6 und
// sein Nachlauf): ihr `defaults read` scheitert, ohne zu sagen, dass es sie
// nicht gibt — dann wird nicht importiert, und ihre Werte bleiben.
func TestEinLesefehlerDerNeuenDomaeneGiltNichtAlsLeer(t *testing.T) {
	t.Parallel()

	umzug, buch, h := mitAlterDomaene(t)
	buch.inhalte[Kennung] = plistMit("", "neu")
	h.nachher = nil
	buch.ausgaenge = map[string]int{"defaults read " + Kennung: 4}

	if err := umzug.vorgabenZiehen(t.Context()); err == nil || importiert(buch) {
		t.Errorf("Fehler %v, importiert %t", err, importiert(buch))
	}

	if neu := buch.inhalte[Kennung]; !strings.Contains(neu, "<string>neu</string>") {
		t.Errorf("die Werte der neuen Fassung sind fort:\n%s", neu)
	}
}

// **Leer ist sie nur, wenn `defaults` es sagt** — dann zieht die alte ganz
// um.
func TestEineFehlendeNeueDomaeneIstLeer(t *testing.T) {
	t.Parallel()

	umzug, buch, _ := mitAlterDomaene(t)

	if err := umzug.vorgabenZiehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	if neu := buch.inhalte[Kennung]; !strings.Contains(neu, "<string>alt</string>") ||
		!strings.Contains(neu, "<key>"+umzugsmarke+"</key>") {
		t.Errorf("die alte Domäne zog nicht um:\n%s", neu)
	}
}

// **Gibt es sie, und scheitert ihr Export oder seine Umwandlung, wird nicht
// importiert** (Nachlauf über Phase 4, Befund 5) — auch ohne Datei unter
// `~/Library/Preferences`, denn cfprefsd führt Werte schon vor ihr.
func TestEinScheiternderExportDerNeuenBrichtAb(t *testing.T) {
	t.Parallel()

	neu := func(befehl []string) bool { return strings.HasSuffix(befehl[len(befehl)-1], "vorgaben-neu.plist") }

	for name, fall := range map[string]struct {
		abweisen func([]string) bool
		inhalt   string
	}{
		"der Export":     {func(b []string) bool { return b[0] == "defaults" && neu(b) }, plistMit("", "neu")},
		"die Umwandlung": {func(b []string) bool { return b[0] == "plutil" && neu(b) }, plistMit("", "neu")},
		"das Lesen":      {nil, "<plist><dict><key>a"},
	} {
		umzug, buch, h := mitAlterDomaene(t)
		buch.domaenen[Kennung] = true
		buch.inhalte[Kennung] = fall.inhalt
		h.abweisen = fall.abweisen

		if err := umzug.vorgabenZiehen(t.Context()); err == nil || importiert(buch) {
			t.Errorf("%s: Fehler %v, importiert %t", name, err, importiert(buch))
		}
	}
}

// sperreHalten nimmt die Sperre der Vorgaben, wie ein zweites Programm.
func sperreHalten(t *testing.T, heim string) *dateisperre.Griff {
	t.Helper()

	ablage := filepath.Join(heim, "Library", "Caches", Kennung)
	if err := os.MkdirAll(ablage, 0o700); err != nil {
		t.Fatal(err)
	}

	griff, err := dateisperre.Oeffnen(filepath.Join(ablage, "vorgaben.sperre"))
	if err != nil {
		t.Fatal(err)
	}

	if err := griff.AlleinBis(t.Context(), 5*time.Second, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	return griff
}

// **Ein Umzug wartet auf den anderen** (Befund 5) — und gibt er auf,
// importiert er nichts.
func TestEinUmzugWartetAufDenAnderen(t *testing.T) {
	t.Parallel()

	umzug, buch, _ := mitAlterDomaene(t)
	griff := sperreHalten(t, umzug.heim)

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // nur zum Aufräumen

	ctx, abbrechen := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer abbrechen()

	if err := umzug.vorgabenZiehen(ctx); err == nil || !strings.Contains(err.Error(), "warten") {
		t.Errorf("Fehler %v", err)
	}

	if importiert(buch) {
		t.Error("importiert, obwohl ein anderer Umzug lief")
	}
}

// **Unter der Sperre wird die Marke noch einmal gefragt** (Befund 5): wer
// eben umzog, hat sie gesetzt, und ein zweiter Import ersetzte, was die
// neue Fassung seither schrieb.
func TestUnterDerSperreWirdDieMarkeNochEinmalGefragt(t *testing.T) {
	t.Parallel()

	umzug, buch, h := mitAlterDomaene(t)
	griff := sperreHalten(t, umzug.heim)

	fertig := make(chan error, 1)

	go func() { fertig <- umzug.vorgabenZiehen(t.Context()) }()

	// **Der andere wird fertig, während dieser wartet** — erst die Marke,
	// dann die Sperre frei.
	time.Sleep(100 * time.Millisecond)
	h.unter(func() { buch.domaenen[Kennung+" "+umzugsmarke] = true })

	if err := griff.Schliessen(); err != nil {
		t.Fatal(err)
	}

	if err := <-fertig; err != nil {
		t.Fatal(err)
	}

	if importiert(buch) {
		t.Error("ein zweites Mal importiert")
	}
}

// **Jeder Lauf hat seinen Ordner, und danach ist er fort** (Befund 5).
func TestJederUmzugHatSeinenOrdner(t *testing.T) {
	t.Parallel()

	umzug, buch, _ := mitAlterDomaene(t)

	if err := umzug.vorgabenZiehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	ablage := filepath.Join(umzug.heim, "Library", "Caches", Kennung)

	for _, befehl := range buch.nur("defaults") {
		if befehl[1] == "import" && filepath.Dir(befehl[3]) == ablage {
			t.Errorf("importiert aus dem gemeinsamen Ordner: %v", befehl)
		}

		if befehl[1] == "import" && (!strings.HasPrefix(befehl[3], filepath.Join(ablage, "umzug-")) || filepath.Dir(filepath.Dir(befehl[3])) != ablage) {
			t.Errorf("nicht aus einem eigenen Ordner: %v", befehl)
		}
	}

	eintraege, err := os.ReadDir(ablage)
	if err != nil || len(eintraege) != 1 || eintraege[0].Name() != "vorgaben.sperre" {
		t.Errorf("im Zwischenspeicher %v (%v)", eintraege, err)
	}
}

// **Ein Arbeitsordner, der nicht entsteht, und eine vereinte Datei, die
// sich nicht schreiben lässt, melden sich** — importiert wird nichts.
func TestEinUnbeschreibbarerArbeitsordnerWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, buch, _ := mitAlterDomaene(t)
	griff := sperreHalten(t, umzug.heim)

	if err := griff.Schliessen(); err != nil {
		t.Fatal(err)
	}

	ablage := filepath.Join(umzug.heim, "Library", "Caches", Kennung)
	if err := os.Chmod(ablage, 0o500); err != nil { //nolint:gosec // nur lesbar, für diesen Lauf
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(ablage, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	if err := umzug.vorgabenZiehen(t.Context()); err == nil || importiert(buch) {
		t.Errorf("Fehler %v, importiert %t", err, importiert(buch))
	}

	vereint, buch2, h := mitAlterDomaene(t)
	h.nachher = func(befehl []string) {
		if len(befehl) == 4 && befehl[1] == "export" {
			_ = os.MkdirAll(filepath.Join(filepath.Dir(befehl[3]), "vorgaben.plist", "x"), 0o700) //nolint:errcheck // ein Ordner an ihrer Stelle
		}
	}

	if err := vereint.vorgabenZiehen(t.Context()); err == nil || importiert(buch2) {
		t.Errorf("Fehler %v, importiert %t", err, importiert(buch2))
	}
}

// **Die Abwesenheit steht in der Ausgabe, die im Grund reist** — so baut
// [befehlAusfuehren] den Fehler: ein Satz, darunter die Ausgabe von
// `defaults`. Ohne Grund ist nichts belegt.
func TestDieAbwesenheitStehtImGrund(t *testing.T) {
	t.Parallel()

	mitAusgabe := domain.MitGrund(domain.ErrConflict,
		fmt.Errorf("%w: %s", errors.New("exit status 1"), "Domain 'x' not found."), "defaults read scheiterte")
	if !domaeneFehlt(mitAusgabe) {
		t.Error("die Ausgabe im Grund wurde nicht gelesen")
	}

	if domaeneFehlt(domain.MitGrund(domain.ErrConflict, nil, "defaults read scheiterte")) {
		t.Error("ohne Grund galt die Domäne als abwesend")
	}

	if domaeneFehlt(errors.Join(errors.New("abgelehnt"), errors.New("Zeitüberschreitung"))) {
		t.Error("ein anderer Fehler galt als Abwesenheit")
	}
}
