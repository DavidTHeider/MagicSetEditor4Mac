package service

// **Der freiwillige Plan: wöchentlich in einen Ordner, sieben eigene** —
// OPS01 (#2347).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// planbau ist ein Plan über einem frischen Bestandsweg.
type planbau struct {
	*wegbau
	buch   *vermerkbuch
	plan   *Sicherungsplan
	ordner string

	// belegt sagt, ob die Sperre gerade jemand anders hält; frei zählt,
	// wie oft sie zurückgegeben wurde.
	belegt bool
	frei   int
}

func neuerPlanbau(t *testing.T) *planbau {
	t.Helper()

	bau := &planbau{wegbau: neuerWegbau(t), ordner: t.TempDir()}
	bau.buch = mitBuch(bau.wegbau)
	bau.plan = NeuerSicherungsplan(bau.weg,
		func(context.Context) (string, error) { return bau.ordner, nil },
		func() (func(), bool, error) {
			if bau.belegt {
				return nil, true, nil
			}

			return func() { bau.frei++ }, false, nil
		},
		domain.FlaecheFenster)

	return bau
}

// eigene legt eine Sicherung des Plans an, `tage` vor der Uhr — Datei und
// Vermerk, wie ein früherer Lauf sie hinterlassen hätte.
func (b *planbau) eigene(t *testing.T, tage int) string {
	t.Helper()

	zeit := vermerkzeitpunkt.AddDate(0, 0, -tage)
	name := Sicherungsplanname(zeit)
	inhalt := []byte("sicherung von " + name)

	if fehler := os.WriteFile(filepath.Join(b.ordner, name), inhalt, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	summe := sha256.Sum256(inhalt)
	b.buch.vermerke = append(b.buch.vermerke, domain.Sicherungsvermerk{
		Gesichert: zeit, Anlass: domain.SicherungNachPlan, Flaeche: domain.FlaecheFenster,
		Datei: name, Groesse: int64(len(inhalt)), Pruefsumme: hex.EncodeToString(summe[:]),
	})

	return name
}

// liegt gibt die Namen im Ordner.
func (b *planbau) liegt(t *testing.T) []string {
	t.Helper()

	eintraege, fehler := os.ReadDir(b.ordner)
	if fehler != nil {
		t.Fatal(fehler)
	}

	namen := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		namen = append(namen, eintrag.Name())
	}

	return namen
}

func TestOhneOrdnerGibtEsKeinenPlan(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.plan.ordner = func(context.Context) (string, error) { return "", nil }

	lauf, fehler := bau.plan.Laufen(t.Context())
	if fehler != nil || lauf.Gesichert != "" || lauf.Belegt || len(bau.buch.vermerke) != 0 || bau.frei != 0 {
		t.Errorf("ohne Ordner: %+v (%v), %d Vermerke", lauf, fehler, len(bau.buch.vermerke))
	}
}

func TestDerPlanSichertWennErFaelligIst(t *testing.T) {
	bau := neuerPlanbau(t)

	lauf, fehler := bau.plan.Laufen(t.Context())
	if fehler != nil || lauf.Gesichert != Sicherungsplanname(vermerkzeitpunkt) {
		t.Fatalf("Lauf %+v (%v)", lauf, fehler)
	}

	if len(bau.buch.vermerke) != 1 || bau.buch.vermerke[0].Anlass != domain.SicherungNachPlan ||
		bau.buch.vermerke[0].Flaeche != domain.FlaecheFenster {
		t.Errorf("vermerkt %+v", bau.buch.vermerke)
	}

	if !slices.Equal(bau.liegt(t), []string{lauf.Gesichert}) || bau.frei != 1 {
		t.Errorf("im Ordner %v, Sperre %d-mal zurück", bau.liegt(t), bau.frei)
	}

	// **Und gleich danach nicht noch einmal** — der Vermerk von eben gilt.
	if zweiter, fehler := bau.plan.Laufen(t.Context()); fehler != nil || zweiter.Gesichert != "" {
		t.Errorf("zweiter Lauf %+v (%v)", zweiter, fehler)
	}
}

func TestDerPlanWartetSeineWocheAb(t *testing.T) {
	bau := neuerPlanbau(t)
	gestern := bau.eigene(t, 1)

	if lauf, fehler := bau.plan.Laufen(t.Context()); fehler != nil || lauf.Gesichert != "" {
		t.Errorf("Lauf %+v (%v)", lauf, fehler)
	}

	if !slices.Equal(bau.liegt(t), []string{gestern}) {
		t.Errorf("im Ordner %v", bau.liegt(t))
	}
}

// **Zwei Prozesse**: hält ein anderer die Sperre, sichert dieser nicht.
func TestEinBelegterPlanSichertNicht(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.belegt = true

	if lauf, fehler := bau.plan.Laufen(t.Context()); fehler != nil || !lauf.Belegt || lauf.Gesichert != "" {
		t.Errorf("Lauf %+v (%v)", lauf, fehler)
	}

	if len(bau.liegt(t)) != 0 {
		t.Errorf("im Ordner %v", bau.liegt(t))
	}
}

// **Ein Stick, der fehlt, ist kein Grund, einen Ordner anzulegen** — sonst
// läge die Sicherung unter `/Volumes` auf der eigenen Platte.
func TestEinFehlenderOrdnerWirdNichtAngelegt(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.ordner = filepath.Join(bau.ordner, "Stick")

	if _, fehler := bau.plan.Laufen(t.Context()); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", fehler)
	}

	if _, lage := os.Stat(bau.ordner); !errors.Is(lage, os.ErrNotExist) {
		t.Errorf("der Ordner wurde angelegt: %v", lage)
	}
}

// **Sieben eigene bleiben, und nichts Fremdes geht** — eine Datei mit dem
// Namen des Plans, die kein Vermerk kennt, eine, die sich seither geändert
// hat, ein Verweis und eine Notiz bleiben stehen.
func TestDerPlanDuenntNurEigeneAus(t *testing.T) {
	bau := neuerPlanbau(t)

	var eigene []string
	for woche := 7; woche >= 1; woche-- {
		eigene = append(eigene, bau.eigene(t, 7*woche))
	}

	fremd := Sicherungsplanname(vermerkzeitpunkt.AddDate(0, 0, -100))

	for name, inhalt := range map[string]string{fremd: "fremd", "notizen.txt": "meins"} {
		if fehler := os.WriteFile(filepath.Join(bau.ordner, name), []byte(inhalt), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	// **Gleich groß und doch nicht mehr dieselbe** — erst die Prüfsumme
	// erkennt sie nicht wieder.
	geaendert := bau.eigene(t, 99)

	roh, lesefehler := os.ReadFile(filepath.Join(bau.ordner, geaendert)) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatal(lesefehler)
	}

	roh[0] ^= 0xff
	if err := os.WriteFile(filepath.Join(bau.ordner, geaendert), roh, 0o600); err != nil { // #nosec G703 -- aus t.TempDir()
		t.Fatal(err)
	}

	// **Ein Verweis, den ein Vermerk kennte** — er ist keine Datei, die
	// dieser Plan geschrieben hat.
	verweis := bau.eigene(t, 98)
	if err := os.Remove(filepath.Join(bau.ordner, verweis)); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(eigene[0], filepath.Join(bau.ordner, verweis)); err != nil {
		t.Fatal(err)
	}

	lauf, fehler := bau.plan.Laufen(t.Context())
	if fehler != nil || lauf.Gesichert == "" {
		t.Fatalf("Lauf %+v (%v)", lauf, fehler)
	}

	// Acht eigene nach dem Lauf: die älteste der sieben geht.
	if !slices.Equal(lauf.Entfernt, []string{eigene[0]}) {
		t.Errorf("entfernt %v, erwartet %v", lauf.Entfernt, eigene[:1])
	}

	erwartet := append(append([]string{}, eigene[1:]...), lauf.Gesichert, fremd, geaendert, verweis, "notizen.txt")
	slices.Sort(erwartet)

	if liegt := bau.liegt(t); !slices.Equal(liegt, erwartet) {
		t.Errorf("im Ordner %v,\nerwartet %v", liegt, erwartet)
	}
}

// **Eine gescheiterte Sicherung räumt nichts weg** — der letzte gute Stand
// bleibt, auch wenn er der achte wäre.
func TestEinGescheiterterPlanRaeumtNichts(t *testing.T) {
	bau := neuerPlanbau(t)

	var eigene []string
	for woche := 8; woche >= 1; woche-- {
		eigene = append(eigene, bau.eigene(t, 7*woche))
	}

	bau.quelle.fehler = errors.New("die Platte ist voll")

	if lauf, fehler := bau.plan.Laufen(t.Context()); fehler == nil || lauf.Gesichert != "" || len(lauf.Entfernt) != 0 {
		t.Fatalf("Lauf %+v (%v)", lauf, fehler)
	}

	slices.Sort(eigene)

	if liegt := bau.liegt(t); !slices.Equal(liegt, eigene) {
		t.Errorf("im Ordner %v, erwartet die acht von vorher", liegt)
	}
}

func TestDerNameDesPlans(t *testing.T) {
	t.Parallel()

	zeit := time.Date(2026, 10, 15, 11, 30, 5, 0, time.FixedZone("MESZ", 2*3600))

	if name := Sicherungsplanname(zeit); name != "bestand-plan-20261015-093005.msesicherung" {
		t.Errorf("Name %q — in Weltzeit, damit sie sich nach der Zeit ordnen", name)
	}

	for name, eigen := range map[string]bool{
		"bestand-plan-20261015-093005.msesicherung":  true,
		"bestand-plan-20261015-093005.db":            false,
		"bestand-plan-2026-10-15.msesicherung":       false,
		"bestand-20261015-093005.msesicherung":       false,
		".bestand-plan-20261015-093005.msesicherung": false,
	} {
		if istPlanname(name) != eigen {
			t.Errorf("%q: eigen %v", name, !eigen)
		}
	}
}

// **Jeder Fehler beim Ausdünnen wird gemeldet**, und keiner nimmt dabei mehr
// weg als vorher — der Ordner ist nicht lesbar, eine Datei nicht, sie lässt
// sich nicht entfernen, das Buch antwortet nicht.
func TestDasAusduennenMeldetJedenFehler(t *testing.T) {
	faelle := []struct {
		name    string
		richten func(*planbau)
		nachher func(*planbau)
	}{
		{"das Buch antwortet nicht", func(bau *planbau) { bau.buch.fehler = domain.ErrBelegt }, func(*planbau) {}},
		{
			"der Ordner ist nicht lesbar", func(bau *planbau) { aendern(t, bau.ordner, 0o300) },
			func(bau *planbau) { aendern(t, bau.ordner, 0o700) },
		},
		{"eine Datei ist nicht lesbar", func(bau *planbau) {
			aendern(t, filepath.Join(bau.ordner, Sicherungsplanname(vermerkzeitpunkt.AddDate(0, 0, -56))), 0o000)
		}, func(*planbau) {}},
		{
			"sie lässt sich nicht entfernen", func(bau *planbau) { aendern(t, bau.ordner, 0o500) },
			func(bau *planbau) { aendern(t, bau.ordner, 0o700) },
		},
	}

	for _, fall := range faelle {
		bau := neuerPlanbau(t)
		for woche := 8; woche >= 1; woche-- {
			bau.eigene(t, 7*woche)
		}

		fall.richten(bau)
		entfernt, fehler := bau.plan.ausduennen(t.Context(), bau.ordner)
		fall.nachher(bau)

		if fehler == nil || len(entfernt) != 0 {
			t.Errorf("%s: entfernt %v (%v)", fall.name, entfernt, fehler)
		}

		if liegt := bau.liegt(t); len(liegt) != 8 {
			t.Errorf("%s: im Ordner %v", fall.name, liegt)
		}
	}
}

// aendern setzt die Rechte eines Pfads.
func aendern(t *testing.T, pfad string, rechte os.FileMode) {
	t.Helper()

	if fehler := os.Chmod(pfad, rechte); fehler != nil {
		t.Fatal(fehler)
	}
}

func TestEinPlanOhneAntwortDesBuchsSichertNicht(t *testing.T) {
	bau := neuerPlanbau(t)
	bau.buch.fehler = domain.ErrBelegt

	if lauf, fehler := bau.plan.Laufen(t.Context()); !errors.Is(fehler, domain.ErrBelegt) || lauf.Gesichert != "" {
		t.Errorf("Lauf %+v (%v)", lauf, fehler)
	}
}
