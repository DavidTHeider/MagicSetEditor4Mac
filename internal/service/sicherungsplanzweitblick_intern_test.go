package service

// **Der Plan nach dem Zweitblick über Phase 4** — OPS01
// (#2347),
// Befunde 2 bis 4.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine veränderte jüngere Sicherung belegt keinen der sieben Plätze**
// (Befund 2): erkannt wurde nach Name und Größe, die Prüfsumme erst vor dem
// Löschen. Bei sieben gültigen, einer jüngeren gleich groß veränderten und
// einer neuen gingen dann zwei gültige — sechs blieben.
func TestEineVeraenderteSicherungBelegtKeinenPlatz(t *testing.T) {
	bau := neuerPlanbau(t)

	var gueltige []string
	for woche := 8; woche >= 2; woche-- {
		gueltige = append(gueltige, bau.eigene(t, 7*woche))
	}

	// **Jünger als die sieben, und der Plan ist trotzdem fällig.**
	veraendert := bau.eigene(t, 10)
	pfad := filepath.Join(bau.ordner, veraendert)

	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	roh[0] ^= 0xff
	if schreibfehler := os.WriteFile(pfad, roh, 0o600); schreibfehler != nil { // #nosec G703 -- aus t.TempDir()
		t.Fatal(schreibfehler)
	}

	lauf, fehler := bau.plan.Laufen(t.Context())
	if fehler != nil || lauf.Gesichert == "" {
		t.Fatalf("Lauf %+v (%v)", lauf, fehler)
	}

	// **Acht gültige nach dem Lauf — die älteste geht, sonst keine.**
	if !slices.Equal(lauf.Entfernt, gueltige[:1]) {
		t.Errorf("entfernt %v, erwartet %v", lauf.Entfernt, gueltige[:1])
	}

	erwartet := append(append([]string{}, gueltige[1:]...), lauf.Gesichert, veraendert)
	slices.Sort(erwartet)

	if liegt := bau.liegt(t); !slices.Equal(liegt, erwartet) {
		t.Errorf("im Ordner %v,\nerwartet %v", liegt, erwartet)
	}
}

// **Ein neuer Ordner ist sofort fällig** (Befund 3): der Vermerk kannte nur
// den Dateinamen, und der aus dem alten Ordner ließ den neuen eine Woche lang
// leer — samt einem „zuletzt", das zum alten gehörte.
func TestEinNeuerPlanordnerIstSofortFaellig(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.eigene(t, 2)

	if lauf, fehler := bau.plan.Laufen(t.Context()); fehler != nil || lauf.Gesichert != "" {
		t.Fatalf("im alten Ordner nach zwei Tagen: %+v (%v)", lauf, fehler)
	}

	alter := bau.ordner
	bau.ordner = t.TempDir()

	if vermerk, bekannt, fehler := bau.weg.Planvermerk(t.Context(), bau.ordner); fehler != nil || bekannt {
		t.Errorf("im neuen Ordner gilt %+v als letzte (%v)", vermerk, fehler)
	}

	lauf, fehler := bau.plan.Laufen(t.Context())
	if fehler != nil || lauf.Gesichert == "" {
		t.Fatalf("im neuen Ordner: %+v (%v)", lauf, fehler)
	}

	if liegt := bau.liegt(t); !slices.Equal(liegt, []string{lauf.Gesichert}) {
		t.Errorf("im neuen Ordner %v", liegt)
	}

	// **Und ein Ordner, der sich nicht lesen lässt**, nennt den jüngsten
	// Vermerk überhaupt — ob fällig, sagt dann die Zeit.
	if vermerk, bekannt, fehler := bau.weg.Planvermerk(t.Context(), filepath.Join(alter, "fehlt")); fehler != nil ||
		!bekannt || vermerk.Datei != lauf.Gesichert {
		t.Errorf("ohne lesbaren Ordner: %+v, %t (%v)", vermerk, bekannt, fehler)
	}
}

// verschluesselterBestand meldet einen verschlüsselten Bestand — die übrigen
// Wege der Chiffre ruft dieser Lauf nicht.
type verschluesselterBestand struct{ Bestandschiffre }

func (verschluesselterBestand) IstVerschluesselt(string) (bool, error) { return true, nil }

// **Ein verschlüsselter Bestand wird nicht nach Plan gesichert — und das
// sagt der Plan selbst** (Befund 4): er hat kein Kennwort und kein
// ausdrückliches „im Klartext", und sein Rat war der der Sicherung von Hand,
// die beides anbietet.
func TestEinVerschluesselterBestandWirdNichtNachPlanGesichert(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.weg.MitBestandschiffre(verschluesselterBestand{})

	lauf, fehler := bau.plan.Laufen(t.Context())

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Key != domain.SchluesselPlanVerschluesselt || lauf.Gesichert != "" {
		t.Fatalf("Lauf %+v (%v)", lauf, fehler)
	}

	if liegt := bau.liegt(t); len(liegt) != 0 || len(bau.buch.vermerke) != 0 {
		t.Errorf("im Ordner %v, %d Vermerke", liegt, len(bau.buch.vermerke))
	}
}

// **Ein halbes Archiv des Plans geht mit dem nächsten Lauf** (Befund 12):
// bricht ein Lauf ab, bleibt `.sicherung-<zufall>.unfertig` im Ordner des
// Plans, und keine Pflege kannte den Namen. Ein frisches bleibt — ein
// anderer Lauf kann gerade daran schreiben.
func TestEinHalbesArchivDesPlansGeht(t *testing.T) {
	bau := neuerPlanbau(t)

	alt := filepath.Join(bau.ordner, ".sicherung-424242.unfertig")
	frisch := filepath.Join(bau.ordner, ".sicherung-77.unfertig")

	// **Ein Ordner mit demselben Namen ist kein halbes Archiv** (Nachlauf
	// über Phase 4, Befund 1) — der Archiver schreibt nur Dateien, und
	// `RemoveAll` nähme ihn samt Inhalt.
	ordnerGleichenNamens := filepath.Join(bau.ordner, ".sicherung-9.unfertig")
	if fehler := os.MkdirAll(filepath.Join(ordnerGleichenNamens, "darin"), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	for _, pfad := range []string{alt, frisch} {
		if fehler := os.WriteFile(pfad, []byte("halb"), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	damals := bau.weg.uhr().Add(-2 * Restfrist)
	for _, pfad := range []string{alt, ordnerGleichenNamens} {
		if fehler := os.Chtimes(pfad, damals, damals); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if lauf, fehler := bau.plan.Laufen(t.Context()); fehler != nil || lauf.Gesichert == "" {
		t.Fatalf("Lauf %+v (%v)", lauf, fehler)
	}

	if _, fehler := os.Stat(alt); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("das alte halbe Archiv steht noch: %v", fehler)
	}

	if _, fehler := os.Stat(frisch); fehler != nil {
		t.Errorf("das frische ist fort: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ordnerGleichenNamens, "darin")); fehler != nil {
		t.Errorf("der Ordner gleichen Namens ist fort: %v", fehler)
	}
}

// **Vor dem Löschen wird noch einmal geprüft** — zwischen Erkennen und
// Löschen kann sich eine Sicherung ändern, und dann gehört sie nicht mehr
// diesem Plan.
func TestVorDemLoeschenWirdNochEinmalGeprueft(t *testing.T) {
	bau := neuerPlanbau(t)
	name := bau.eigene(t, 70)
	vermerk := bau.buch.vermerke[len(bau.buch.vermerke)-1]
	vermerk.Pruefsumme = "0000"

	if weg, fehler := wegnehmen(filepath.Join(bau.ordner, name), vermerk); weg || fehler != nil {
		t.Errorf("weg %t (%v) — eine veränderte ging", weg, fehler)
	}

	if liegt := bau.liegt(t); !slices.Equal(liegt, []string{name}) {
		t.Errorf("im Ordner %v", liegt)
	}
}

// **Ändert sich eine aufzubewahrende Sicherung nach dem Erkennen, wird nichts
// entfernt** (Nachlauf über Phase 4, Befund 2): die Liste der Löschungen
// beruhte auf ihr, und eine gültige ältere ginge für sie.
func TestEineSpaeterVeraenderteBehalteneHaeltDasAusduennenAn(t *testing.T) {
	bau := neuerPlanbau(t)

	var gueltige []string
	for woche := 9; woche >= 2; woche-- {
		gueltige = append(gueltige, bau.eigene(t, 7*woche))
	}

	juengste := filepath.Join(bau.ordner, gueltige[len(gueltige)-1])
	nachDemErkennen = func() {
		if fehler := os.WriteFile(juengste, []byte("anders, und gleich lang? nein"), 0o600); fehler != nil {
			t.Error(fehler)
		}
	}

	t.Cleanup(func() { nachDemErkennen = nil })

	entfernt, fehler := bau.plan.ausduennen(t.Context(), bau.ordner)
	if !errors.Is(fehler, domain.ErrConflict) || len(entfernt) != 0 {
		t.Errorf("entfernt %v (%v)", entfernt, fehler)
	}

	if liegt := bau.liegt(t); len(liegt) != len(gueltige) {
		t.Errorf("im Ordner %v", liegt)
	}
}

// **Ein Ordner, der nicht erreichbar ist, sagt es — auch, wenn der Plan nicht
// fällig wäre** (Nachlauf über Phase 4, Befund 3): der Vermerk aus einem
// früheren Ordner ließ den Lauf vor der Prüfung still enden.
func TestEinUnerreichbarerOrdnerSagtEsVorDerFaelligkeit(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.eigene(t, 2)

	bau.ordner = filepath.Join(bau.ordner, "abgezogen")

	if lauf, fehler := bau.plan.Laufen(t.Context()); !errors.Is(fehler, domain.ErrNotFound) || lauf.Gesichert != "" {
		t.Errorf("Lauf %+v (%v)", lauf, fehler)
	}
}

// **Eine gleich groß beschädigte letzte Sicherung ist nicht frisch**
// (Nachlauf über Phase 4, Befund 3): Name und Größe hielten sie für eine,
// und der Plan wartete eine Woche über einem kaputten Stand.
func TestEineBeschaedigteLetzteSicherungMachtDenPlanFaellig(t *testing.T) {
	bau := neuerPlanbau(t)
	name := bau.eigene(t, 2)
	pfad := filepath.Join(bau.ordner, name)

	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	roh[0] ^= 0xff
	if schreibfehler := os.WriteFile(pfad, roh, 0o600); schreibfehler != nil { // #nosec G703 -- aus t.TempDir()
		t.Fatal(schreibfehler)
	}

	if lauf, laufehler := bau.plan.Laufen(t.Context()); laufehler != nil || lauf.Gesichert == "" {
		t.Errorf("Lauf %+v (%v) — der Plan hielt die beschädigte für frisch", lauf, laufehler)
	}
}
