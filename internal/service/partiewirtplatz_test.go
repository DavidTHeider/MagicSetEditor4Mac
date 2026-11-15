package service_test

// **Der Wirt setzt einen eingeladenen Gast mit seinem Deck** — P9b
// (#2377).
// Der erste Rückkehrwunsch trägt, was der Gast mitbringt. **Wartet sein Sitz
// und bringt er nichts mit, wird er vor dem Anfang mit Grund abgewiesen** —
// ein Gast mit älterer Fassung sitzt sonst an einem Sitz ohne Karten.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wartenderNehmer gibt eine Sicht, in der der Sitz des Gastes wartet.
type wartenderNehmer struct {
	nehmerdoppel
	wartet bool
}

// einWartenderNehmer steht bei Folge 5 — ein neuer Gast bekommt den Abzug.
func einWartenderNehmer(wartet bool) *wartenderNehmer {
	nehmer := &wartenderNehmer{wartet: wartet}
	nehmer.stand = domain.Wirtsstand{Epoche: domain.Partieepoche, Folge: 5, Kette: "abc", AelteteFolge: 1}

	return nehmer
}

func (n *wartenderNehmer) Partie(
	ctx context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	sicht, fehler := n.nehmerdoppel.Partie(ctx, kennung, fuer)
	sicht.Spiel = "magic"
	sicht.Sitze = []domain.Sichtsitz{{Kennung: "wirt"}, {Kennung: "gast", Wartet: n.wartet}}

	return sicht, fehler
}

// sitzgeberdoppel merkt sich, wer sich womit setzen wollte.
type sitzgeberdoppel struct {
	gesetzt []domain.Partiedeck
	peers   []string
	klagt   error
}

func (p *sitzgeberdoppel) Platznehmen(
	_ context.Context, _, _, peer string, deck domain.Partiedeck,
	verkuende service.Verkuendung,
) (service.Befehlsausgang, error) {
	if p.klagt != nil {
		return service.Befehlsausgang{}, p.klagt
	}

	p.gesetzt = append(p.gesetzt, deck)
	p.peers = append(p.peers, peer)

	ausgang := service.Befehlsausgang{Fassung: 2, Ereignisse: []domain.Ereignis{{
		PartieID: "p1", Folge: 1, Art: service.EreignisPlatznahme,
	}}}
	verkuende(ausgang)

	return ausgang, nil
}

// leihdoppel gibt den Abzug einer Leihe, die der Wirt ausgestellt hat.
type leihdoppel struct {
	gefragt []string
	klagt   error
}

func (l *leihdoppel) VerliehenesDeck(
	_ context.Context, kennung, entleiher, _ string,
) (domain.Partiedeck, error) {
	l.gefragt = append(l.gefragt, kennung+"@"+entleiher)

	if l.klagt != nil {
		return domain.Partiedeck{}, l.klagt
	}

	return domain.Partiedeck{Abdruck: "geliehen", Karten: []domain.Abzugskarte{{Kennung: "k1", Name: "Wald", Menge: 1}}}, nil
}

func einMitbringwunsch(t *testing.T, mit *domain.Mitgebracht) domain.Partienachricht {
	t.Helper()

	return eineKontrollnachricht(t, domain.KontrolleRueckkehr, domain.Wiederaufnahmewunsch{
		Partie: "p1", Epoche: domain.Partieepoche, Mitgebracht: mit,
	})
}

func einGastdeck() *domain.Partiedeck {
	return &domain.Partiedeck{
		Quelle: "peer-jemand-anders", Abdruck: "x",
		Karten: []domain.Abzugskarte{{Kennung: "k9", Name: "Moorgeist", Menge: 3}},
	}
}

// einPlatzwirt baut einen Wirt, an dessen Tisch der Sitz des Gastes wartet.
func einPlatzwirt(wartet bool) (*service.Partiewirt, *nahtdoppel, *sitzgeberdoppel, *leihdoppel, *mitschrift) {
	platz, leihen, fenster := &sitzgeberdoppel{}, &leihdoppel{}, &mitschrift{}
	wirt, naht := einWirt(einWartenderNehmer(wartet))

	return wirt.MitGegenueber("peer-gast").MitPlatzgeber(platz).MitLeihgeber(leihen).
		MitFenster("wirt", fenster), naht, platz, leihen, fenster
}

func TestDerEingeladeneSetztSichBeimErstenWunsch(t *testing.T) {
	t.Parallel()

	wirt, naht, platz, _, fenster := einPlatzwirt(true)

	if fehler := wirt.Angekommen(context.Background(),
		einMitbringwunsch(t, &domain.Mitgebracht{Deck: einGastdeck(), Spiel: "magic"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	// **Die Quelle ist der Gast — was er behauptet, zählt nicht.**
	if len(platz.gesetzt) != 1 || platz.gesetzt[0].Quelle != "peer-gast" || platz.peers[0] != "peer-gast" {
		t.Fatalf("gesetzt: %+v, %v", platz.gesetzt, platz.peers)
	}

	if antwort := dieAntwort(t, naht); antwort.Art == domain.WegAbbruch {
		t.Errorf("abgewiesen: %+v", antwort)
	}

	// **Und das Fenster des Wirts erfährt es**, ohne selbst etwas zu tun.
	if len(fenster.ereignisse) != 1 || fenster.ereignisse[0].Art != service.EreignisPlatznahme {
		t.Errorf("am Fenster: %+v", fenster.ereignisse)
	}
}

// **Eine Leihe des Wirts bringt der Gast bei ihrer Kennung mit** — gesetzt
// wird der Abzug, den der Wirt selbst hält.
func TestEineLeiheDesWirtsBringtDerGastBeiIhrerKennungMit(t *testing.T) {
	t.Parallel()

	wirt, naht, platz, leihen, _ := einPlatzwirt(true)

	if fehler := wirt.Angekommen(context.Background(),
		einMitbringwunsch(t, &domain.Mitgebracht{Leihe: "l1"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if len(leihen.gefragt) != 1 || leihen.gefragt[0] != "l1@peer-gast" {
		t.Errorf("gefragt: %v", leihen.gefragt)
	}

	if len(platz.gesetzt) != 1 || platz.gesetzt[0].Abdruck != "geliehen" || platz.gesetzt[0].Quelle != "" {
		t.Errorf("gesetzt: %+v", platz.gesetzt)
	}

	if dieAntwort(t, naht).Art == domain.WegAbbruch {
		t.Error("abgewiesen")
	}
}

// **Wer nicht sitzen kann, erfährt es vor dem Anfang** — mit Grund, und die
// Sitzung endet.
func TestWerSichNichtSetzenKannErfaehrtWarum(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		mit    *domain.Mitgebracht
		richte func(*sitzgeberdoppel, *leihdoppel)
		grund  string
	}{
		"ein Gast ohne Deck":            {nil, nil, "keines mitgebracht"},
		"ein ungültiges Mitbringen":     {&domain.Mitgebracht{Deck: einGastdeck(), Leihe: "l1", Spiel: "magic"}, nil, "nicht beides"},
		"ein Deck eines anderen Spiels": {&domain.Mitgebracht{Deck: einGastdeck(), Spiel: "pokemon"}, nil, "pokemon"},
		"eine Leihe, die nicht gilt": {&domain.Mitgebracht{Leihe: "l1"}, func(_ *sitzgeberdoppel, l *leihdoppel) {
			l.klagt = errors.New("diese Leihe ist zurückgenommen")
		}, "zurückgenommen"},
		"der Sitz nimmt das Deck nicht": {&domain.Mitgebracht{Deck: einGastdeck(), Spiel: "magic"}, func(p *sitzgeberdoppel, _ *leihdoppel) {
			p.klagt = errors.New("an diesem Sitz liegt schon ein Deck")
		}, "schon ein Deck"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wirt, naht, platz, leihen, _ := einPlatzwirt(true)
			if fall.richte != nil {
				fall.richte(platz, leihen)
			}

			fehler := wirt.Angekommen(context.Background(), einMitbringwunsch(t, fall.mit))
			if !errors.Is(fehler, domain.ErrConflict) {
				t.Errorf("es kam %v", fehler)
			}

			antwort := dieAntwort(t, naht)
			if antwort.Art != domain.WegAbbruch || !strings.Contains(antwort.Grund, fall.grund) {
				t.Errorf("die Antwort: %+v", antwort)
			}

			if fall.richte == nil && len(platz.gesetzt) != 0 {
				t.Errorf("trotzdem gesetzt: %+v", platz.gesetzt)
			}
		})
	}
}

// **Ein Sitz mit Deck nimmt nichts Mitgebrachtes** — wer nach dem Hinsetzen
// wiederkommt, setzt sich nicht noch einmal.
func TestEinSitzMitDeckNimmtNichtsMitgebrachtes(t *testing.T) {
	t.Parallel()

	wirt, naht, platz, _, _ := einPlatzwirt(false)

	if fehler := wirt.Angekommen(context.Background(),
		einMitbringwunsch(t, &domain.Mitgebracht{Deck: einGastdeck(), Spiel: "magic"})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if len(platz.gesetzt) != 0 || dieAntwort(t, naht).Art == domain.WegAbbruch {
		t.Errorf("gesetzt: %+v", platz.gesetzt)
	}
}

// **Ein Wirt ohne Platzgeber nimmt keine Decks an** — und sagt es.
func TestOhnePlatzgeberSetztSichNiemand(t *testing.T) {
	t.Parallel()

	wirt, naht := einWirt(einWartenderNehmer(true))
	wirt = wirt.MitGegenueber("peer-gast")

	if fehler := wirt.Angekommen(context.Background(),
		einMitbringwunsch(t, &domain.Mitgebracht{Deck: einGastdeck(), Spiel: "magic"})); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if antwort := dieAntwort(t, naht); antwort.Art != domain.WegAbbruch {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

// **Ist der Tisch nicht zu lesen, geht der Fehler zurück** — eine Antwort
// gäbe es zu einem Stand, den niemand kennt.
func TestEinUnlesbarerTischSetztNiemanden(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Partie ließ sich nicht lesen")
	nehmer := einWartenderNehmer(true)
	nehmer.bildklage = kaputt
	wirt, _ := einWirt(nehmer)

	if fehler := wirt.MitPlatzgeber(&sitzgeberdoppel{}).Angekommen(context.Background(),
		einMitbringwunsch(t, &domain.Mitgebracht{Deck: einGastdeck(), Spiel: "magic"})); !errors.Is(fehler, kaputt) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ohne Leihgeber gibt es keine Leihe** — und der Gast erfährt es.
func TestOhneLeihgeberKeineLeihe(t *testing.T) {
	t.Parallel()

	wirt, naht := einWirt(einWartenderNehmer(true))
	wirt = wirt.MitGegenueber("peer-gast").MitPlatzgeber(&sitzgeberdoppel{})

	if fehler := wirt.Angekommen(context.Background(),
		einMitbringwunsch(t, &domain.Mitgebracht{Leihe: "l1"})); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if antwort := dieAntwort(t, naht); !strings.Contains(antwort.Grund, "Leihe") {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

// **Der Grund geht hinaus, bevor die Sitzung endet** — Zweitblick über Phase
// 7, Befund 7: der Abbruch lag in der Schlange, und `Partienetz.Bediene`
// kehrt mit dem Fehler zurück, bevor es hinausgibt. Der Gast las dann keinen
// Grund, sondern eine abgerissene Leitung.
func TestDerGrundEinesAbbruchsGehtHinaus(t *testing.T) {
	t.Parallel()

	naht := &leerendeNaht{}
	sitzung := service.NeuePartiesitzung("p1", "wirt", naht, service.NeueReihenfolge(1))
	wirt := service.NeuerPartiewirt("p1", "gast", einWartenderNehmer(true), sitzung).
		MitGegenueber("peer-gast")

	if fehler := wirt.Angekommen(context.Background(), einMitbringwunsch(t, nil)); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	if naht.geleert == 0 {
		t.Error("der Abbruch liegt noch in der Schlange")
	}
}

// klemmendeNaht gibt nichts hinaus.
type klemmendeNaht struct {
	nahtdoppel
	klage error
}

func (n *klemmendeNaht) Leere(context.Context) error { return n.klage }

// **Klemmt das Hinausgeben, kommt dieser Fehler zurück** — er sagt mehr
// als der Abbruch, der nicht ankam.
func TestKlemmtDasHinausgebenKommtDasZurueck(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("die Leitung klemmt")
	naht := &klemmendeNaht{klage: klemmt}
	sitzung := service.NeuePartiesitzung("p1", "wirt", naht, service.NeueReihenfolge(1))
	wirt := service.NeuerPartiewirt("p1", "gast", einWartenderNehmer(true), sitzung)

	if fehler := wirt.Angekommen(context.Background(), einMitbringwunsch(t, nil)); !errors.Is(fehler, klemmt) {
		t.Fatalf("es kam %v", fehler)
	}
}
