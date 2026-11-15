package domain_test

// **Was eine Nachricht über den Stand drüben bewirkt** — Werkstück `T12` von
// Phase 5.6
// (#1477).
//
// # Der Fund, für den es das gibt
//
// **Eine Zustimmung konnte nicht reisen.** [Handelsangebot.Nimmt] verlangt
// die **nächste** Revision; eine Stimme ändert die Revision aber nicht — sie
// tritt zu derselben hinzu. Zwei Prozesse kamen deshalb nie über `angeboten`
// hinaus: jeder hatte seine eigene Stimme und keiner die des anderen.
//
// **Gemessen am 2026-09-19** mit zwei `mse`-Prozessen über einen echten
// Draht (`internal/flaechen/handelsabnahme_test.go`).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einStand baut einen Vorgang in einem Zustand.
func einHandelsstand(zustand domain.Handelszustand) domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "v-1",
		Revision:        1,
		Art:             domain.HandelTausch,
		Initiator:       "banna",
		Parteien:        []string{"banna", "bbert"},
		Zustand:         zustand,
		Positionen: []domain.Handelsposition{{
			Von: "banna", Printing: "a1b2", Menge: 1,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	}
}

// **Eine fremde Stimme tritt hinzu.**
func TestEineFremdeStimmeTrittHinzu(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	drueben := einHandelsstand(domain.HandelAngeboten)
	drueben.Zustimmungen = []string{"bbert"}

	danach, fehler := hier.NimmtStand(drueben, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if len(danach.Zustimmungen) != 1 || danach.Zustimmungen[0] != "bbert" {
		t.Errorf("die Stimmen sind %v", danach.Zustimmungen)
	}

	if danach.Zustand != domain.HandelAngeboten {
		t.Errorf("der Zustand ist %q — eine Stimme von zweien", danach.Zustand)
	}
}

// **Und mit der zweiten wechselt der Zustand.**
func TestMitDerZweitenStimmeWechseltDerZustand(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)
	hier.Zustimmungen = []string{"banna"}

	drueben := einHandelsstand(domain.HandelAngeboten)
	drueben.Zustimmungen = []string{"bbert"}

	danach, fehler := hier.NimmtStand(drueben, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if danach.Zustand != domain.HandelAngenommen {
		t.Errorf("der Zustand ist %q", danach.Zustand)
	}
}

// **Ein Schritt, den der Absender gehen darf, wird nachvollzogen.**
//
// **`vorbereiten` darf nur der Initiator** — und ohne diesen Weg bliebe die
// Gegenseite auf `angenommen` stehen und könnte nie übergeben.
func TestEinSchrittDesAbsendersWirdNachvollzogen(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngenommen)
	drueben := einHandelsstand(domain.HandelVorbereitet)

	danach, fehler := hier.NimmtStand(drueben, "banna")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if danach.Zustand != domain.HandelVorbereitet {
		t.Errorf("der Zustand ist %q", danach.Zustand)
	}
}

// **Ein Abschlussschritt wird nicht nachvollzogen.**
//
// **„Bei mir ist gebucht" ist eine Auskunft und kein Befehl.** Wer sie
// nachvollzöge, stünde auf `lokalVerbucht`, ohne dass eine Karte den
// Besitzer gewechselt hätte — und der Nachweis aus `T10` wäre umgangen.
func TestEinAbschlussschrittWirdNichtNachvollzogen(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelUebergabeOffen)
	drueben := einHandelsstand(domain.HandelLokalVerbucht)

	danach, fehler := hier.NimmtStand(drueben, "banna")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if danach.Zustand != domain.HandelUebergabeOffen {
		t.Errorf("der Zustand ist %q", danach.Zustand)
	}
}

// **Ein fremder Vorgang und eine andere Revision gehen nicht.**
func TestEinFremderStandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	fremd := einHandelsstand(domain.HandelAngeboten)
	fremd.Vorgang = "v-anderer"

	if _, fehler := hier.NimmtStand(fremd, "bbert"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("der fremde Vorgang gab %v", fehler)
	}

	spaeter := einHandelsstand(domain.HandelAngeboten)
	spaeter.Revision = 2

	if _, fehler := hier.NimmtStand(spaeter, "bbert"); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("die andere Revision gab %v", fehler)
	}
}

// **Und ein veränderter Kopf ist ein Konflikt.**
//
// **Sonst käme eine Stimme zu einem Angebot an, das jemand nebenher
// umgeschrieben hat** — und sie gälte dem neuen.
func TestEinVeraenderterKopfIstEinKonflikt(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	anders := einHandelsstand(domain.HandelAngeboten)
	anders.Positionen[0].Menge = 9
	anders.Zustimmungen = []string{"bbert"}

	if _, fehler := hier.NimmtStand(anders, "bbert"); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("der veränderte Kopf gab %v", fehler)
	}
}

// **Eine Stimme, die hier nichts bewirkt, ändert nichts.**
//
// **Der Fall der Wiederholung**: dieselbe Nachricht kommt noch einmal, und
// der Stand ist derselbe.
func TestEineBekannteStimmeAendertNichts(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)
	hier.Zustimmungen = []string{"bbert"}

	danach, fehler := hier.NimmtStand(hier, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if danach.Abdruck() != hier.Abdruck() ||
		danach.Zustand != hier.Zustand ||
		len(danach.Zustimmungen) != 1 {
		t.Errorf("der Stand wurde zu %+v", danach)
	}
}

// **Die Stimme zu einem Abschluss reist nicht hier.**
//
// **`bestaetigen` ist ein Beschluss und ein Abschlussschritt**: wer ihn
// nachspielte, brächte den Vorgang auf `beidseitigVerbucht`, ohne dass die
// Verbuchungsstelle je gefragt worden wäre — der Riegel aus `T10` wäre
// umgangen, und zwar über eine Nachricht von drüben.
func TestDieStimmeZuEinemAbschlussReistNichtHier(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAbgleichOffen)

	drueben := einHandelsstand(domain.HandelAbgleichOffen)
	drueben.Zustimmungen = []string{"bbert"}

	danach, fehler := hier.NimmtStand(drueben, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	switch {
	case danach.Zustand != domain.HandelAbgleichOffen:
		t.Errorf("der Zustand ist %q", danach.Zustand)
	case len(danach.Zustimmungen) != 0:
		t.Errorf("die Stimmen sind %v", danach.Zustimmungen)
	}
}

// **Aus einem Zustand ohne Beschluss trägt keine Stimme.**
//
// **`entwurf` hat die Gegenseite nie gesehen** — eine Stimme dazu gehört zu
// einem anderen Stand, und sie stillschweigend zu behalten hieße, sie später
// gegen einen Beschluss zu zählen, den sie nicht meinte.
func TestAusEinemZustandOhneBeschlussTraegtKeineStimme(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelEntwurf)

	drueben := einHandelsstand(domain.HandelEntwurf)
	drueben.Zustimmungen = []string{"bbert"}

	if _, fehler := hier.NimmtStand(drueben, "bbert"); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und eine Stimme, die der Absender gar nicht geben darf, wird abgewiesen.**
//
// **Die Gegenprobe zur Zeile darüber**: übergangen wird, was **andere**
// betrifft; was der Absender über **sich** behauptet, geht durch
// [domain.Handelsangebot.Schreite] — und der weist ihn ab, wenn er nicht zu
// den Parteien gehört.
func TestEineStimmeEinesNichtbeteiligtenAbsendersWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	drueben := einHandelsstand(domain.HandelAngeboten)
	drueben.Zustimmungen = []string{"bfremder"}

	if _, fehler := hier.NimmtStand(drueben, "bfremder"); fehler == nil {
		t.Error("der Fremde stimmte mit")
	}
}

// **Eine Stimme von jemandem, der nicht mitspielt, zählt nicht.**
//
// **Übergangen und nicht abgewiesen**: eine Nachricht trägt den **ganzen**
// Stand, und darin steht regelmäßig auch die eigene Stimme des Empfängers.
// Wer jede fremde Stimme als Fehler nähme, wiese den Normalfall ab. Was
// zählt, ist allein die des Absenders.
func TestEineStimmeVonDraussenZaehltNicht(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	drueben := einHandelsstand(domain.HandelAngeboten)
	drueben.Zustimmungen = []string{"bfremder"}

	danach, fehler := hier.NimmtStand(drueben, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if len(danach.Zustimmungen) != 0 {
		t.Errorf("die Stimmen sind %v", danach.Zustimmungen)
	}
}

// **Einen Schritt, den der Absender nicht gehen darf, spielt niemand nach.**
//
// **`vorbereiten` darf nur der Initiator**: wer die Gegenseite behaupten
// ließe, sie habe vorbereitet, ließe sie den Vorgang dorthin schieben, wo
// nur der Initiator ihn hinbringt.
func TestEinUnerlaubterSchrittWirdNichtNachvollzogen(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngenommen)
	drueben := einHandelsstand(domain.HandelVorbereitet)

	if _, fehler := hier.NimmtStand(drueben, "bbert"); fehler == nil {
		t.Error("die Gegenseite hat vorbereitet")
	}
}

// **Und ein Zustand, zu dem kein einzelner Schritt führt, bleibt liegen.**
//
// **Kein Fehler**: drüben kann jemand weiter sein, als ein Schritt trägt —
// was fehlt, kommt mit der nächsten Nachricht.
func TestEinWeiterZustandBleibtLiegen(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)
	drueben := einHandelsstand(domain.HandelUebergabeOffen)

	danach, fehler := hier.NimmtStand(drueben, "banna")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	if danach.Zustand != domain.HandelAngeboten {
		t.Errorf("der Zustand ist %q", danach.Zustand)
	}
}

// **Wer spricht, stimmt für sich und für niemanden sonst.**
//
// # Der Fund, für den es diesen Lauf gibt
//
// **Ein Absender konnte die Stimme der Gegenseite behaupten** — Befund 2 des
// Zweitblicks über Phase 5.6. Die Leitung prüft, **wer** spricht; was er über
// die Stimmen **anderer** sagt, prüfte niemand. Bert konnte eine Nachricht
// schicken, deren `Zustimmungen` Anna und Bert nennen, und der Empfänger
// spielte beide nach — der Vorgang stand auf `angenommen`, ohne dass Anna je
// zugestimmt hätte.
//
// **Zwei Parteien, eine Stimme je Nachricht**: was Anna zustimmt, sagt Anna,
// und sie hat ihre eigene Leitung.
func TestWerSprichtStimmtNurFuerSich(t *testing.T) {
	t.Parallel()

	hier := einHandelsstand(domain.HandelAngeboten)

	// **Bert behauptet beide Stimmen.**
	drueben := einHandelsstand(domain.HandelAngeboten)
	drueben.Zustimmungen = []string{"banna", "bbert"}

	danach, fehler := hier.NimmtStand(drueben, "bbert")
	if fehler != nil {
		t.Fatalf("NimmtStand: %v", fehler)
	}

	switch {
	case danach.Zustand != domain.HandelAngeboten:
		t.Errorf("der Zustand ist %q — eine behauptete Stimme zählte mit",
			danach.Zustand)
	case len(danach.Zustimmungen) != 1 || danach.Zustimmungen[0] != "bbert":
		t.Errorf("die Stimmen sind %v", danach.Zustimmungen)
	}
}
