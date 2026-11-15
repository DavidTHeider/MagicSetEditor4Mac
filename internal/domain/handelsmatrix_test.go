package domain_test

// **Die Zustandsmatrix eines Handels** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was hier gemessen wird
//
// **Dass jeder Schritt einen Akteur hat.** „Angenommen" ist kein Zustand, in
// den ein Angebot von selbst gerät; jemand hat zugestimmt, und es ist nicht
// egal wer. Ein Handel, in dem der Initiator für die Gegenseite zustimmen
// kann, ist keiner.
//
// # Und die Wiederholung
//
// **Zweimal dasselbe ist einmal.** Über eine Leitung geht eine Nachricht
// verloren und kommt noch einmal; ein zweites „ich stimme zu" darf keine
// zweite Zustimmung werden und keinen Fehler geben — es bekommt **dieselbe
// Antwort** (§7, Regel 4).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Übergang der Matrix führt von seinem Zustand in seinen Zustand.**
func TestJederUebergangFuehrtWohinErSagt(t *testing.T) {
	t.Parallel()

	for _, uebergang := range domain.Handelsmatrix() {
		t.Run(string(uebergang.Schritt)+" aus "+string(uebergang.Von),
			func(t *testing.T) {
				t.Parallel()

				danach := einAngebot()
				danach.Zustand = uebergang.Von

				for _, wer := range werDarf(uebergang) {
					var fehler error

					danach, fehler = danach.Schreite(uebergang.Schritt, wer)
					if fehler != nil {
						t.Fatalf("%q von %q: %v",
							uebergang.Schritt, wer, fehler)
					}
				}

				if danach.Zustand != uebergang.Nach {
					t.Errorf("der Zustand ist %q statt %q",
						danach.Zustand, uebergang.Nach)
				}
			})
	}
}

// werDarf gibt die Parteien, die diesen Übergang gehen — **alle**, wo der
// Übergang es verlangt.
//
// **Ein Schritt, den alle tun müssen, ist nicht derselbe wie einer, den jede
// tun darf** ([domain.Handelsuebergang.Alle]): eine einzelne Zustimmung
// bewegt nichts, und ein einzelner Abbruch bewegt alles.
func werDarf(uebergang domain.Handelsuebergang) []string {
	switch {
	case uebergang.Alle:
		return []string{"bert", "anna"}

	case uebergang.Wer == domain.HandelnderGegenseite:
		return []string{"bert"}
	}

	return []string{"anna"}
}

// **Jeder Zustand steht in der Matrix — als Ausgang oder als Ende.**
//
// **Sonst gäbe es einen, in den ein Handel gerät und aus dem er nicht
// herauskommt**, ohne dass irgendwo stünde, dass er ein Ende ist.
func TestJederZustandStehtInDerMatrix(t *testing.T) {
	t.Parallel()

	erreichbar := map[domain.Handelszustand]bool{domain.HandelEntwurf: true}
	verlassbar := map[domain.Handelszustand]bool{}

	for _, uebergang := range domain.Handelsmatrix() {
		erreichbar[uebergang.Nach] = true
		verlassbar[uebergang.Von] = true
	}

	for _, zustand := range domain.Handelszustaende() {
		if !erreichbar[zustand] {
			t.Errorf("in %q kommt ein Handel nie", zustand)
		}

		if !verlassbar[zustand] && !domain.HandelEnde(zustand) {
			t.Errorf("aus %q kommt ein Handel nie heraus, und es ist kein Ende",
				zustand)
		}
	}
}

// **Ein Schritt aus dem falschen Zustand wird abgewiesen.**
func TestEinSchrittAusDemFalschenZustandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	angebot := einAngebot()
	angebot.Zustand = domain.HandelBeidseitigVerbucht

	if _, fehler := angebot.Schreite(
		domain.SchrittZustimmen, "bert",
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ein Schritt vom falschen Akteur wird abgewiesen.**
//
// **Das ist der Lauf, an dem am meisten hängt**: wer für die Gegenseite
// zustimmen kann, hat keinen Handel, sondern eine Buchung.
func TestEinSchrittVomFalschenAkteurWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	angebot := einAngebot()
	angebot.Zustand = domain.HandelEntwurf

	// **Anbieten darf der Initiator** — nicht die Gegenseite.
	if _, fehler := angebot.Schreite(
		domain.SchrittAnbieten, "bert",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Gegenseite bietet an: %v", fehler)
	}

	// **Und niemand, der gar nicht dabei ist.**
	if _, fehler := angebot.Schreite(
		domain.SchrittAnbieten, "cara",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Fremder bietet an: %v", fehler)
	}
}

// **Beide Zustimmungen machen aus einem Angebot eine Einigung.**
func TestBeideZustimmungenErgebenEineEinigung(t *testing.T) {
	t.Parallel()

	angebot := einAngebot()
	angebot.Zustand = domain.HandelAngeboten

	erste, fehler := angebot.Schreite(domain.SchrittZustimmen, "bert")
	if fehler != nil {
		t.Fatalf("Bert stimmt zu: %v", fehler)
	}

	// **Eine allein genügt nicht** — sonst wäre „angenommen" eine Aussage
	// über einen von zweien.
	if erste.Zustand != domain.HandelAngeboten {
		t.Errorf("nach einer Zustimmung ist der Zustand %q", erste.Zustand)
	}

	zweite, fehler := erste.Schreite(domain.SchrittZustimmen, "anna")
	if fehler != nil {
		t.Fatalf("Anna stimmt zu: %v", fehler)
	}

	if zweite.Zustand != domain.HandelAngenommen {
		t.Errorf("nach beiden Zustimmungen ist der Zustand %q", zweite.Zustand)
	}
}

// **Dieselbe Zustimmung zweimal ist einmal.**
//
// **Ohne Fehler und ohne zweite Zustimmung** — über eine Leitung geht eine
// Nachricht verloren und kommt noch einmal. Ein Fehler zwänge den Absender zu
// entscheiden, ob er sie noch einmal schicken darf; eine zweite Zustimmung
// machte aus einem von zwei Parteien zwei.
func TestDieselbeZustimmungZweimalIstEinmal(t *testing.T) {
	t.Parallel()

	angebot := einAngebot()
	angebot.Zustand = domain.HandelAngeboten

	einmal, fehler := angebot.Schreite(domain.SchrittZustimmen, "bert")
	if fehler != nil {
		t.Fatalf("Bert stimmt zu: %v", fehler)
	}

	zweimal, fehler := einmal.Schreite(domain.SchrittZustimmen, "bert")
	if fehler != nil {
		t.Fatalf("Bert stimmt noch einmal zu: %v", fehler)
	}

	if len(zweimal.Zustimmungen) != 1 {
		t.Errorf("es stehen %d Zustimmungen da: %v",
			len(zweimal.Zustimmungen), zweimal.Zustimmungen)
	}

	if zweimal.Zustand != einmal.Zustand {
		t.Errorf("der Zustand wurde %q statt %q",
			zweimal.Zustand, einmal.Zustand)
	}
}

// **Ein Ende ist ein Ende.**
func TestAusEinemEndeFuehrtKeinSchritt(t *testing.T) {
	t.Parallel()

	enden := []domain.Handelszustand{
		domain.HandelBeidseitigVerbucht, domain.HandelAbgelehnt,
		domain.HandelAbgebrochen, domain.HandelAbgelaufen,
	}

	for _, zustand := range enden {
		t.Run(string(zustand), func(t *testing.T) {
			t.Parallel()

			if !domain.HandelEnde(zustand) {
				t.Fatalf("%q gilt nicht als Ende", zustand)
			}

			angebot := einAngebot()
			angebot.Zustand = zustand

			for _, uebergang := range domain.Handelsmatrix() {
				if _, fehler := angebot.Schreite(
					uebergang.Schritt, "anna",
				); fehler == nil {
					t.Errorf("%q geht aus %q noch",
						uebergang.Schritt, zustand)
				}
			}
		})
	}
}

// **Ein lokaler Abschluss ist kein beidseitiger.**
//
// # Die dritte beobachtbare Abnahme des Issues
//
// **„Ein lokaler Commit wird nicht als bereits beidseitig abgeschlossener
// Handel angezeigt."** Zwei SQLite-Bestände sind keine gemeinsame
// Transaktion; wer auf dieser Seite verbucht hat, weiß nichts über die
// andere. Der Unterschied steht deshalb in **zwei Zuständen** und nicht in
// einem Flag daneben — ein Flag ließe sich vergessen, ein Zustand nicht.
func TestEinLokalerAbschlussIstKeinBeidseitiger(t *testing.T) {
	t.Parallel()

	for _, offen := range []domain.Handelszustand{
		domain.HandelLokalVerbucht, domain.HandelAbgleichOffen,
	} {
		if domain.HandelEnde(offen) {
			t.Errorf("%q gilt als Ende — die Gegenbestätigung fehlt noch",
				offen)
		}
	}

	if !domain.HandelEnde(domain.HandelBeidseitigVerbucht) {
		t.Error("der Zustand „beidseitig verbucht“ gilt nicht als Ende")
	}

	// **Und der Weg dorthin führt über den Abgleich.** Wer vom lokalen
	// Verbuchen unmittelbar fertig würde, hätte die Gegenbestätigung
	// übersprungen.
	angebot := einAngebot()
	angebot.Zustand = domain.HandelLokalVerbucht

	if _, fehler := angebot.Schreite(
		domain.SchrittBestaetigen, "anna"); fehler == nil {
		t.Error("aus dem lokalen Verbuchen wurde unmittelbar fertig")
	}
}
