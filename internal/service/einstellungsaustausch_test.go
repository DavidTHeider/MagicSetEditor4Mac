package service_test

// **Ausfuhr, Vorschau, Übernahme** — Werkstück O3 von Phase 5.2
// (#1427).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nutzerwerteprobe ist die Naht zum Menschen, im Arbeitsspeicher.
type nutzerwerteprobe struct {
	werte  map[string]string
	fehler error
}

func (n *nutzerwerteprobe) NutzerwerteZumProfil(context.Context) (
	map[string]string, error,
) {
	if n.fehler != nil {
		return nil, n.fehler
	}

	return n.werte, nil
}

// austauschablage merkt sich, was geschrieben wurde.
type austauschablage struct {
	profil, nutzer map[string]string
	rufe           int
	fehler         error

	// erwartet ist die Revision, die der letzte Ruf mitgebracht hat — Befund
	// 11 des Zweitblicks über Phase 5.2.
	erwartet string
}

func (a *austauschablage) EinstellungenUebernehmen(
	_ context.Context, profil, nutzer map[string]string, erwartet string,
) error {
	a.rufe++
	a.erwartet = erwartet

	if a.fehler != nil {
		return a.fehler
	}

	a.profil, a.nutzer = profil, nutzer

	return nil
}

// austauschwache ist die Naht zur Autorisierung.
type austauschwache struct {
	nein      error
	akteurAus error
	akteur    domain.Akteur
	gefragt   []domain.Wirkung
}

func (w *austauschwache) Pruefe(_ context.Context, wirkung domain.Wirkung) error {
	w.gefragt = append(w.gefragt, wirkung)

	return w.nein
}

func (w *austauschwache) Akteur(context.Context) (domain.Akteur, error) {
	if w.akteurAus != nil {
		return domain.Akteur{}, w.akteurAus
	}

	return w.akteur, nil
}

// werkbank hält die vier Nähte samt dem Austausch darüber.
type werkbank struct {
	ablage      *einstellungsablage
	nutzerwerte *nutzerwerteprobe
	schreiber   *austauschablage
	wache       *austauschwache
	austausch   *service.Einstellungsaustausch
}

// werkbankBauen gibt einen Austausch über leeren Tafeln.
func werkbankBauen() *werkbank {
	bank := &werkbank{
		ablage:      &einstellungsablage{werte: map[string]string{}},
		nutzerwerte: &nutzerwerteprobe{werte: map[string]string{}},
		schreiber:   &austauschablage{},
		wache: &austauschwache{
			akteur: domain.Akteur{NutzerID: 3, ProfilID: 7},
		},
	}

	bank.austausch = service.NeuerEinstellungsaustausch(
		bank.ablage, bank.nutzerwerte, bank.schreiber, bank.wache)

	return bank
}

// alsDatei packt einen Austausch, wie ihn eine Datei trüge.
func alsDatei(t *testing.T, bereiche map[string]map[string]string) []byte {
	t.Helper()

	roh, fehler := json.Marshal(domain.Austausch{
		Fassung: domain.AustauschFassung, Bereiche: bereiche,
	})
	if fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	return roh
}

// **Ausgeführt wird nur, was gesetzt ist — keine Vorgaben.**
//
// Eine Datei voller Vorgabewerte sähe drüben wie eine Wahl aus und **fröre**
// den heutigen Stand ein: wer später eine Vorgabe ändert, erreichte jeden
// nicht, der einmal importiert hat.
func TestDieAusfuhrTraegtNurGesetzteWerte(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.ablage.werte[domain.EinstellungSprache] = "en"
	bank.nutzerwerte.werte[domain.NutzerKontrast] = domain.KontrastHoch

	austausch, fehler := bank.austausch.Ausfuhr(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if austausch.Fassung != domain.AustauschFassung {
		t.Errorf("die Fassung ist %d", austausch.Fassung)
	}

	imProfil := austausch.Bereiche[string(domain.EbeneProfil)]
	if len(imProfil) != 1 || imProfil[domain.EinstellungSprache] != "en" {
		t.Errorf("im Profil steht %v", imProfil)
	}

	beimMenschen := austausch.Bereiche[string(domain.EbeneNutzer)]
	if len(beimMenschen) != 1 ||
		beimMenschen[domain.NutzerKontrast] != domain.KontrastHoch {
		t.Errorf("beim Menschen steht %v", beimMenschen)
	}
}

// **Kein Geheimnis, kein Pfad, keine lokale Kennung.**
//
// Der Wächter dazu steht in `internal/werkzeuge/austauschsperre_test.go` und
// hält die **Klasse**; hier steht der Weg: was in der Tafel liegt, kommt nicht
// heraus.
func TestDieAusfuhrLaesstDenRechnerZurueck(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.ablage.werte[domain.VorgabeHochladepfad] = "/Users/jemand/Bilder"
	bank.ablage.werte[domain.AnbieterKennung(domain.Anbieter()[0])] = "Ov23geheim"
	bank.ablage.werte[domain.EinstellungAktualisierung] = domain.AktualisierungMelden
	bank.ablage.werte[domain.ArbeitsplatzNavigator] = domain.SpalteVerborgen

	austausch, fehler := bank.austausch.Ausfuhr(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	roh, fehler := json.Marshal(austausch)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	for _, verboten := range []string{
		"/Users/jemand", "Ov23geheim", domain.EinstellungAktualisierung,
		domain.ArbeitsplatzNavigator,
	} {
		if strings.Contains(string(roh), verboten) {
			t.Errorf("%q steht in der Ausfuhr: %s", verboten, roh)
		}
	}
}

// **Der Arbeitsplatz reist nur, wer ihn nennt.**
func TestDerArbeitsplatzReistNurAufWunsch(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.ablage.werte[domain.ArbeitsplatzNavigator] = domain.SpalteVerborgen

	gewuenscht, fehler := bank.austausch.Ausfuhr(
		t.Context(), []string{string(domain.EbeneArbeitsplatz)})
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	imArbeitsplatz := gewuenscht.Bereiche[string(domain.EbeneArbeitsplatz)]
	if imArbeitsplatz[domain.ArbeitsplatzNavigator] != domain.SpalteVerborgen {
		t.Errorf("der Arbeitsplatz kam nicht mit: %v", gewuenscht.Bereiche)
	}

	if _, dabei := gewuenscht.Bereiche[string(domain.EbeneProfil)]; dabei {
		t.Errorf("das Profil kam ungefragt mit: %v", gewuenscht.Bereiche)
	}
}

// **Ein Abschnitt, den es nicht gibt, und der Rechner werden abgewiesen.**
func TestWelcheAbschnitteSichAusfuehrenLassen(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"woanders", string(domain.EbeneGeraet)} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := werkbankBauen().austausch.Ausfuhr(
				t.Context(), []string{name})
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%v", fehler)
			}
		})
	}
}

// **Ein Leser führt nicht aus.**
func TestOhneErlaubnisKeineAusfuhr(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	bank.wache.nein = domain.ErrErlaubnis

	if _, fehler := bank.austausch.Ausfuhr(t.Context(), nil); !errors.Is(
		fehler, domain.ErrErlaubnis) {
		t.Errorf("%v", fehler)
	}

	if len(bank.wache.gefragt) != 1 || bank.wache.gefragt[0] != domain.WirkungLesen {
		t.Errorf("gefragt wurde nach %v", bank.wache.gefragt)
	}

	// **Und keine Vorschau** — sie liest denselben Bestand.
	if _, fehler := bank.austausch.Vorschau(
		t.Context(), alsDatei(t, nil)); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("%v", fehler)
	}
}

// **Ein Wert, den der Mensch nie gesetzt hat, kommt aus der Vorgabe.**
//
// **Und die Herkunft sagt es**: wer „system" sieht, weiß sonst nicht, ob er es
// gewählt hat — und damit nicht, was ein Zurücksetzen bewirkte.
func TestDieHerkunftEinesUngesetztenNutzerwerts(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()

	vorschau, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneNutzer): {
				domain.NutzerKontrast: domain.KontrastHoch,
			},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(vorschau.Zeilen) != 1 {
		t.Fatalf("%+v", vorschau.Zeilen)
	}

	zeile := vorschau.Zeilen[0]
	if zeile.Vorher != domain.NachSystem || zeile.Herkunft != domain.GeltungVorgabe {
		t.Errorf("%+v", zeile)
	}

	// **Und die Gegenprobe**: hat er gewählt, steht seine Wahl da und die
	// Herkunft ist er. Ohne sie sagte der Lauf darüber nur, dass eine Herkunft
	// dasteht — nicht, dass sie unterscheidet.
	bank.nutzerwerte.werte[domain.NutzerKontrast] = domain.NachSystem

	gewaehlt, fehler := bank.austausch.Vorschau(t.Context(), alsDatei(t,
		map[string]map[string]string{
			string(domain.EbeneNutzer): {
				domain.NutzerKontrast: domain.KontrastHoch,
			},
		}))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(gewaehlt.Zeilen) != 1 ||
		gewaehlt.Zeilen[0].Herkunft != domain.GeltungNutzer {
		t.Errorf("%+v", gewaehlt.Zeilen)
	}
}
