package service

// **Der Zeitpunkt einer Sicherung steht erst nach der geprüften Datei** —
// OPS01 (#2347).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// vermerkbuch ist ein Sicherungsbuch im Arbeitsspeicher.
type vermerkbuch struct {
	vermerke []domain.Sicherungsvermerk
	fehler   error
}

func (v *vermerkbuch) Vermerken(_ context.Context, vermerk domain.Sicherungsvermerk) error {
	if v.fehler != nil {
		return v.fehler
	}

	if fehler := vermerk.Pruefe(); fehler != nil {
		return fehler
	}

	v.vermerke = append(v.vermerke, vermerk)

	return nil
}

func (v *vermerkbuch) Letzter(context.Context) (domain.Sicherungsvermerk, bool, error) {
	if len(v.vermerke) == 0 || v.fehler != nil {
		return domain.Sicherungsvermerk{}, false, v.fehler
	}

	return v.vermerke[len(v.vermerke)-1], true, nil
}

func (v *vermerkbuch) Vermerke(_ context.Context, anlass domain.Sicherungsanlass) ([]domain.Sicherungsvermerk, error) {
	var gewaehlt []domain.Sicherungsvermerk

	for stelle := len(v.vermerke) - 1; stelle >= 0; stelle-- {
		if v.vermerke[stelle].Anlass == anlass {
			gewaehlt = append(gewaehlt, v.vermerke[stelle])
		}
	}

	return gewaehlt, v.fehler
}

// vermerkzeitpunkt ist der Augenblick, den die Uhr dieser Läufe nennt.
var vermerkzeitpunkt = time.Date(2026, 10, 15, 9, 30, 0, 0, time.UTC)

// mitBuch hängt ein Buch und eine feste Uhr an.
func mitBuch(bau *wegbau) *vermerkbuch {
	buch := &vermerkbuch{}
	bau.weg.MitSicherungsbuch(buch)
	bau.weg.MitUhr(func() time.Time { return vermerkzeitpunkt })

	return buch
}

// summeVon rechnet Größe und SHA-256 der Datei auf der Platte.
func summeVon(t *testing.T, pfad string) (groesse int64, summe string) {
	t.Helper()

	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	roh256 := sha256.Sum256(roh)

	return int64(len(roh)), hex.EncodeToString(roh256[:])
}

func TestEineGeprueftSicherungWirdVermerkt(t *testing.T) {
	defer sicherung.ArbeitsfaktorSetzen(10)()

	for _, kennwort := range []string{"", "ein langes Kennwort"} {
		bau := neuerWegbau(t)
		buch := mitBuch(bau)
		ziel := filepath.Join(t.TempDir(), "bestand.msesicherung")

		ctx := MitSicherungswunsch(t.Context(), Sicherungswunsch{Kennwort: kennwort, Flaeche: domain.FlaecheCLI})
		if fehler := bau.weg.Sichern(ctx, ziel); fehler != nil {
			t.Fatalf("sichern: %v", fehler)
		}

		groesse, summe := summeVon(t, ziel)
		erwartet := domain.Sicherungsvermerk{
			Gesichert: vermerkzeitpunkt, Anlass: domain.SicherungVonHand, Flaeche: domain.FlaecheCLI,
			Datei: "bestand.msesicherung", Groesse: groesse, Pruefsumme: summe, Verschluesselt: kennwort != "",
		}

		if len(buch.vermerke) != 1 || buch.vermerke[0] != erwartet {
			t.Errorf("vermerkt %+v, erwartet %+v", buch.vermerke, erwartet)
		}
	}
}

// **Die Probe vor einer Aktualisierung vermerkt nichts** — ihre Datei geht
// gleich danach weg, und ein Vermerk nennte eine Sicherung, die es nicht
// gibt. Sie nennt auch keine Fläche: mit [Bestandsweg.Sichern] scheiterte
// sie daran.
func TestDieProbeVorDerAktualisierungVermerktNichts(t *testing.T) {
	bau := neuerWegbau(t)
	buch := mitBuch(bau)

	if fehler := SicherungsprobeUeber(bau.weg.Probesichern).SichereUndPruefe(t.Context()); fehler != nil {
		t.Fatalf("die Probe: %v", fehler)
	}

	if len(buch.vermerke) != 0 {
		t.Errorf("die Probe wurde vermerkt: %+v", buch.vermerke)
	}
}

func TestDerPlanVermerktSeinenAnlass(t *testing.T) {
	bau := neuerWegbau(t)
	buch := mitBuch(bau)

	ctx := MitSicherungswunsch(t.Context(), Sicherungswunsch{
		Flaeche: domain.FlaecheFenster, Anlass: domain.SicherungNachPlan,
	})
	if fehler := bau.weg.Sichern(ctx, filepath.Join(t.TempDir(), "plan.msesicherung")); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	if len(buch.vermerke) != 1 || buch.vermerke[0].Anlass != domain.SicherungNachPlan {
		t.Errorf("vermerkt %+v", buch.vermerke)
	}
}

func TestEineGescheiterteSicherungVermerktNichts(t *testing.T) {
	faelle := []struct {
		name    string
		richten func(*wegbau, string)
	}{
		{"die Datenbank lässt sich nicht kopieren", func(bau *wegbau, _ string) {
			bau.quelle.fehler = errors.New("die Platte ist voll")
		}},
		{"der Name ist schon vergeben", func(_ *wegbau, ziel string) {
			if fehler := os.WriteFile(ziel, []byte("fremd"), 0o600); fehler != nil {
				t.Fatal(fehler)
			}
		}},
	}

	for _, fall := range faelle {
		bau := neuerWegbau(t)
		buch := mitBuch(bau)
		ziel := filepath.Join(t.TempDir(), "bestand.msesicherung")
		fall.richten(bau, ziel)

		ctx := MitSicherungswunsch(t.Context(), Sicherungswunsch{Flaeche: domain.FlaecheREST})
		if fehler := bau.weg.Sichern(ctx, ziel); fehler == nil {
			t.Fatalf("%s: geglückt", fall.name)
		}

		if len(buch.vermerke) != 0 {
			t.Errorf("%s: vermerkt %+v — eine gescheiterte Sicherung ändert die Zeit nicht", fall.name, buch.vermerke)
		}
	}
}

// **Geschrieben ist geschrieben** (§7, Regel 3): scheitert erst der Vermerk,
// bleibt die geprüfte Datei liegen, und der Fehler sagt es.
func TestEinGescheiterterVermerkLaesstDieSicherungStehen(t *testing.T) {
	bau := neuerWegbau(t)
	buch := mitBuch(bau)
	buch.fehler = domain.ErrBelegt
	ziel := filepath.Join(t.TempDir(), "bestand.msesicherung")

	fehler := bau.weg.Sichern(MitSicherungswunsch(t.Context(), Sicherungswunsch{Flaeche: domain.FlaecheMCP}), ziel)
	if !errors.Is(fehler, domain.ErrSicherungUnvermerkt) {
		t.Fatalf("Fehler %v, erwartet ErrSicherungUnvermerkt", fehler)
	}

	if _, lage := os.Stat(ziel); lage != nil {
		t.Errorf("die geprüfte Sicherung ist fort: %v", lage)
	}
}

func TestDasAlterKommtAusDemBuch(t *testing.T) {
	bau := neuerWegbau(t)

	if stand, fehler := bau.weg.Sicherungsstand(t.Context()); fehler != nil || stand.Stufe != domain.SicherungsstufeUnbekannt {
		t.Errorf("ohne Buch: %+v (%v)", stand, fehler)
	}

	buch := mitBuch(bau)

	alt := domain.Sicherungsvermerk{
		Gesichert: vermerkzeitpunkt.AddDate(0, 0, -16), Anlass: domain.SicherungVonHand, Flaeche: domain.FlaecheCLI,
		Datei: "alt.msesicherung", Groesse: 1, Pruefsumme: hex.EncodeToString(make([]byte, 32)),
	}
	buch.vermerke = append(buch.vermerke, alt)

	stand, fehler := bau.weg.Sicherungsstand(t.Context())
	if fehler != nil || stand.Stufe != domain.SicherungsstufeAlt || stand.Tage != 16 || stand.Letzter != alt {
		t.Errorf("Stand %+v (%v), erwartet alt seit 16 Tagen", stand, fehler)
	}

	buch.fehler = domain.ErrBelegt

	if _, fehler := bau.weg.Sicherungsstand(t.Context()); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("ein Lesefehler wurde zu %v — unbekannt hieße hier, es nie gewusst zu haben", fehler)
	}
}

func TestDerLetzteVermerkDesPlans(t *testing.T) {
	bau := neuerWegbau(t)

	if _, bekannt, fehler := bau.weg.Planvermerk(t.Context(), ""); bekannt || fehler != nil {
		t.Errorf("ohne Buch: %v, %v", bekannt, fehler)
	}

	buch := mitBuch(bau)

	if _, bekannt, fehler := bau.weg.Planvermerk(t.Context(), ""); bekannt || fehler != nil {
		t.Errorf("ohne Vermerk: %v, %v", bekannt, fehler)
	}

	plan := domain.Sicherungsvermerk{
		Gesichert: vermerkzeitpunkt, Anlass: domain.SicherungNachPlan, Flaeche: domain.FlaecheFenster,
		Datei: Sicherungsplanname(vermerkzeitpunkt), Groesse: 1, Pruefsumme: hex.EncodeToString(make([]byte, 32)),
	}
	buch.vermerke = append(buch.vermerke, plan)

	if gelesen, bekannt, fehler := bau.weg.Planvermerk(t.Context(), ""); !bekannt || fehler != nil || gelesen != plan {
		t.Errorf("gelesen %+v (%v, %v)", gelesen, bekannt, fehler)
	}

	buch.fehler = domain.ErrBelegt

	if _, _, fehler := bau.weg.Planvermerk(t.Context(), ""); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("ein Lesefehler wurde zu %v", fehler)
	}
}

// ordnerbuch schaut beim Vermerken in den Ordner der Sicherung.
type ordnerbuch struct {
	vermerkbuch
	ordner string
	reste  []string
}

func (o *ordnerbuch) Vermerken(ctx context.Context, vermerk domain.Sicherungsvermerk) error {
	eintraege, _ := os.ReadDir(o.ordner) //nolint:errcheck // ein fehlender Ordner zeigt sich am Vermerk
	for _, eintrag := range eintraege {
		if strings.HasPrefix(eintrag.Name(), Sicherungsvorsatz) {
			o.reste = append(o.reste, eintrag.Name())
		}
	}

	return o.vermerkbuch.Vermerken(ctx, vermerk)
}

// **Vermerkt wird erst, wenn der Zwischenplatz fort ist** — gefunden in der
// Pipeline am 2026-10-16 (`TestDerPlanSichertGleichNachDerOrdnerwahl` unter
// `-race`): der Plan meldete „gesichert“, während neben der Datei noch
// `.sicherung-…` stand; das Aufräumen lief verzögert, nach dem Vermerk. Wer
// in dem Augenblick in den Ordner sah — ein Mensch im Finder, ein Lauf —,
// fand einen Rest, den der Stand schon für erledigt hielt.
func TestVermerktWirdErstNachDemAufraeumen(t *testing.T) {
	bau := neuerWegbau(t)
	ordner := t.TempDir()
	buch := &ordnerbuch{ordner: ordner}
	bau.weg.MitSicherungsbuch(buch)
	bau.weg.MitUhr(func() time.Time { return vermerkzeitpunkt })

	ctx := MitSicherungswunsch(t.Context(), Sicherungswunsch{Flaeche: domain.FlaecheFenster, Anlass: domain.SicherungNachPlan})
	if fehler := bau.weg.Sichern(ctx, filepath.Join(ordner, "plan.msesicherung")); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	if len(buch.vermerke) != 1 || len(buch.reste) != 0 {
		t.Errorf("beim Vermerk lag noch %v im Ordner (%d Vermerke)", buch.reste, len(buch.vermerke))
	}
}
