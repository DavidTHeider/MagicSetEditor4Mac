package domain_test

// **Die Partie und ihre Sitze** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wonach hier gefragt wird
//
// Danach, dass eine Partie **einen** benannten Wirt hat, dass ein Zuschauer
// **keinen Sitz belegt**, und dass ein Sitz ohne Deck keiner ist. Das sind die
// drei Sätze, auf denen alles Weitere steht: wer serialisiert, wer sehen darf
// und was überhaupt gespielt wird.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einePartie ist die kleinste, die durchgeht.
func einePartie() domain.Partie {
	return domain.Partie{
		Spiel: "magic",
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung, Strenge: domain.StrengeSandkasten,
		},
		Profil: domain.Grundprofil("magic"),
		Wirt:   "sitz-1",
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "David", Deck: einDeckabzug()},
			{Kennung: "sitz-2", Name: "Gegenüber", Deck: einDeckabzug()},
		},
	}.MitStartleben()
}

func einDeckabzug() domain.Partiedeck {
	return domain.Partiedeck{
		Abdruck: "abcd",
		Karten:  []domain.Abzugskarte{{Kennung: "k1", Name: "Wald", Menge: 40}},
	}
}

// **Die kleinste Partie geht durch.**
//
// **Der Lauf steht zuerst da**, damit jede Abweisung darunter eine Aussage
// ist: ohne ihn sagte „wird abgewiesen" nur, dass **irgendetwas** fehlt.
func TestDieKleinstePartieGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := einePartie().Pruefe(); fehler != nil {
		t.Fatalf("die kleinste Partie: %v", fehler)
	}
}

// **Was eine Partie nicht sein kann.**
func TestWasKeinePartieIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Partie){
		"ohne Spiel": func(p *domain.Partie) { p.Spiel = "" },

		"ohne Regelwerk": func(p *domain.Partie) {
			p.Regelwerk = domain.Regelwerk{}
		},

		// **Ein Wirt, den es nicht gibt.** Er serialisiert die Befehle und
		// hält den Vollzustand; ein Name ohne Sitz wäre eine Partie, die auf
		// niemanden wartet.
		"mit einem Wirt ohne Sitz": func(p *domain.Partie) {
			p.Wirt = "sitz-9"
		},

		"ohne Wirt": func(p *domain.Partie) { p.Wirt = "" },

		"ohne jeden Sitz": func(p *domain.Partie) { p.Sitze = nil },

		"mit zwei gleichen Sitzkennungen": func(p *domain.Partie) {
			p.Sitze[1].Kennung = p.Sitze[0].Kennung
		},

		"mit einem Sitz ohne Deck": func(p *domain.Partie) {
			p.Sitze[1].Deck = domain.Partiedeck{}
		},

		"mit einem Sitz ohne Namen": func(p *domain.Partie) {
			p.Sitze[0].Name = "   "
		},
	}

	for name, verderben := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := einePartie()
			verderben(&partie)

			if fehler := partie.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Zuschauer belegt keinen Sitz.**
//
// **Das ist die Zusage, an der die ganze Sichtbarkeit hängt** (#1458): wer
// keinen Sitz hat, hat keine Hand — und bekommt deshalb auch keine zu sehen.
func TestEinZuschauerBelegtKeinenSitz(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Zuschauer = []domain.Zuschauer{{Kennung: "z1", Name: "Nachbar"}}

	if fehler := partie.Pruefe(); fehler != nil {
		t.Fatalf("mit Zuschauer: %v", fehler)
	}

	if partie.SitzVon("z1") != nil {
		t.Error("ein Zuschauer hat einen Sitz bekommen")
	}

	if sitz := partie.SitzVon("sitz-2"); sitz == nil || sitz.Name != "Gegenüber" {
		t.Errorf("der Sitz fehlt: %+v", sitz)
	}

	// **Und seine Kennung darf keine Sitzkennung sein** — sonst sähe eine
	// Sichtprüfung ihn für einen Spieler an.
	partie.Zuschauer[0].Kennung = "sitz-1"

	if fehler := partie.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Zuschauer auf einer Sitzkennung: %v", fehler)
	}
}

// **Der Wirt ist einer der Sitze, und er ist benannt.**
func TestDerWirtIstEinSitz(t *testing.T) {
	t.Parallel()

	partie := einePartie()

	if !partie.IstWirt("sitz-1") || partie.IstWirt("sitz-2") {
		t.Error("die Wirtsfrage beantwortet sich falsch")
	}

	// **Auch die leere Kennung ist kein Wirt** — sonst wäre jeder Ruf ohne
	// Absender einer.
	if partie.IstWirt("") {
		t.Error("die leere Kennung gilt als Wirt")
	}
}

// **Die Aufzählungen sind vollständig und werden gebraucht.**
//
// **Eine Liste ohne Leser ist keine Fläche** (§10, erste Dimension) — und
// eine, die ein Wert nicht enthält, den [Partie.Pruefe] durchlässt, wäre
// schlimmer: sie sähe vollständig aus.
func TestDieAufzaehlungenDerPartie(t *testing.T) {
	t.Parallel()

	if len(domain.Partiestaende()) != 3 {
		t.Errorf("es sind %d Stände", len(domain.Partiestaende()))
	}

	for _, stand := range domain.Partiestaende() {
		if strings.TrimSpace(string(stand)) == "" {
			t.Error("ein Stand ohne Namen")
		}
	}

	if len(domain.Strengegrade()) != 3 {
		t.Errorf("es sind %d Strengegrade", len(domain.Strengegrade()))
	}

	for _, grad := range domain.Strengegrade() {
		if !domain.IstStrenge(grad) {
			t.Errorf("%q gilt nicht als Strenge", grad)
		}
	}

	// **Vier seit `M9`** (#1611):
	// eine ausgelöste Fähigkeit auf dem Stapel ist ein **Objekt** und keine
	// Karte (CR 603.3).
	if len(domain.Objektarten()) != 4 {
		t.Errorf("es sind %d Objektarten", len(domain.Objektarten()))
	}
}

// **Ein Regelwerk mit einer Fassung, die es nie gab.**
func TestWasKeinRegelwerkIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Regelwerk{
		"ohne Kennung": {Fassung: 1, Strenge: domain.StrengeSandkasten},
		"ohne Fassung": {Kennung: "x", Strenge: domain.StrengeSandkasten},
		"mit Fassung null": {
			Kennung: "x", Fassung: 0, Strenge: domain.StrengeSandkasten,
		},
		"mit erfundener Strenge": {
			Kennung: "x", Fassung: 1, Strenge: domain.Strenge("locker"),
		},
	}

	for name, werk := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := werk.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Was kein Deckabzug ist — und was er zählt.**
func TestDerDeckabzug(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Partiedeck{
		"ohne Abdruck": {Karten: []domain.Abzugskarte{
			{Kennung: "k1", Name: "Wald", Menge: 1},
		}},
		"ohne Karten": {Abdruck: "ab"},
		"mit namenloser Karte": {Abdruck: "ab", Karten: []domain.Abzugskarte{
			{Kennung: " ", Name: "Wald", Menge: 1},
		}},
		"mit derselben Karte zweimal": {Abdruck: "ab", Karten: []domain.Abzugskarte{
			{Kennung: "k1", Name: "Wald", Menge: 1},
			{Kennung: "k1", Name: "Wald", Menge: 1},
		}},
		"mit null Stück": {Abdruck: "ab", Karten: []domain.Abzugskarte{
			{Kennung: "k1", Name: "Wald", Menge: 0},
		}},
	}

	for name, abzug := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := abzug.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}

	// **Gezählt wird die Menge und nicht die Zeile** — vier Wälder sind
	// vier Karten und nicht eine.
	abzug := domain.Partiedeck{Abdruck: "ab", Karten: []domain.Abzugskarte{
		{Kennung: "k1", Name: "Wald", Menge: 24},
		{Kennung: "k2", Name: "Bär", Menge: 36},
	}}

	if gezaehlt := abzug.Zaehle(); gezaehlt != 60 {
		t.Errorf("gezählt wurden %d Karten", gezaehlt)
	}
}

// **Ein Sitz ohne Kennung und ein Zuschauer ohne Kennung.**
//
// **Beide Wege getrennt**, weil sie zwei verschiedene Listen prüfen — und
// eine Abweisung im einen sagt nichts über die andere.
func TestKennungenSindPflicht(t *testing.T) {
	t.Parallel()

	ohneSitzkennung := einePartie()
	ohneSitzkennung.Sitze[1].Kennung = "  "

	if fehler := ohneSitzkennung.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ein Sitz ohne Kennung: %v", fehler)
	}

	ohneZuschauerkennung := einePartie()
	ohneZuschauerkennung.Zuschauer = []domain.Zuschauer{{Name: "Nachbar"}}

	if fehler := ohneZuschauerkennung.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ein Zuschauer ohne Kennung: %v", fehler)
	}
}

// **Ein einzelner Sitz ist ein Aufbau und kein Spiel — und genau deshalb
// erlaubt** (G3, #1460).
//
// **CR 100.1** (Fassung 2026-08-19): die Regeln gelten für Spiele mit zwei
// oder mehr Spielern. Ein Goldfish hat einen — er ist damit **kein Spiel nach
// den Regeln**, und dieses Programm setzt an ihm auch keine durch.
//
// **Hier stand bis zum 2026-09-17 das Gegenteil**: „eine Partie braucht
// mindestens zwei Sitze — mit einem ist es kein Spiel, sondern ein Aufbau."
// Der Satz war richtig und die Folgerung falsch: #1460 verlangt „Solo
// Goldfish", also genau diesen Aufbau. Wer allein übt, will kein Spiel — er
// will seine Kurve sehen.
func TestEinEinzelnerSitzIstEinAufbauUndErlaubt(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze = partie.Sitze[:1]
	partie.Wirt = partie.Sitze[0].Kennung

	if fehler := partie.Pruefe(); fehler != nil {
		t.Errorf("ein einzelner Sitz wurde abgewiesen: %v", fehler)
	}
}
