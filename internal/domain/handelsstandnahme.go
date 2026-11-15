package domain

// **Was eine Nachricht über den Stand drüben bewirkt** — Werkstück `T12` von
// Phase 5.6
// (#1477).
//
// # Der Fund, für den es das gibt
//
// **Eine Zustimmung konnte nicht reisen.** [Handelsangebot.Nimmt] verlangt
// die **nächste** Revision — so trägt ein Vorschlag sich fort. Eine Stimme
// ändert die Revision aber nicht: sie tritt zu **derselben** hinzu. Zwei
// Prozesse kamen deshalb nie über `angeboten` hinaus; jeder hatte seine
// eigene Stimme und keiner die des anderen. Gemessen am 2026-09-19 über
// einen echten Draht.
//
// # Warum die Schritte nachgespielt werden und nichts kopiert wird
//
// **Weil der Zustand drüben eine Behauptung ist.** Wer ihn übernähme, nähme
// jeden Zustand, den ein Gegenüber hinschreibt — und mit ihm die Frage, wer
// hier eigentlich welchen Schritt gehen darf. Nachgespielt wird über
// [Handelsangebot.Schreite]: dieselbe Matrix, dieselben Rollen, dieselben
// Beschlüsse.
//
// # Und warum die Abschlussschritte **nicht** nachgespielt werden
//
// **Weil „bei mir ist gebucht" eine Auskunft ist und kein Befehl.** Wer sie
// nachvollzöge, stünde auf `lokalVerbucht`, ohne dass eine Karte den
// Besitzer gewechselt hätte — und der Nachweis aus `T10` wäre umgangen.
// Dieselbe Linie wie in `service.Handelsleitung.Schreiten`.

import (
	"fmt"
	"slices"
)

// NimmtStand nimmt auf, was eine Nachricht über denselben Stand sagt.
//
// **Dieselbe Revision und kein Vorschlag.** Wer eine neue Revision bringt,
// geht über [Handelsangebot.Nimmt]; hier geht es um das, was **zu** einer
// Revision hinzukommt: Stimmen und die Schritte, die keine neue Revision
// erzeugen.
//
// **`absender` ist, wer wirklich spricht** — die Leitung hat ihn geprüft.
// Was drüben im Zustand steht, wird an ihm gemessen: ein Schritt, den er
// nicht gehen darf, wird nicht nachgespielt.
func (a Handelsangebot) NimmtStand(
	fremd Handelsangebot, absender string,
) (Handelsangebot, error) {
	if fehler := a.passtZumStand(fremd); fehler != nil {
		return a, fehler
	}

	// **Eine Stimme gehört zu ihrem Beschluss** (TRADE-03, #1717): eine
	// Zustimmung zum Angebot, die erst ankommt, wenn hier schon vorbereitet
	// ist, zählte sonst als Stimme zur Übergabe. Steht drüben ein anderer
	// Zustand, zählt allein der Schritt dorthin.
	danach := a

	if fremd.Zustand == a.Zustand {
		var fehler error
		if danach, fehler = a.mitFremdenStimmen(fremd, absender); fehler != nil {
			return a, fehler
		}
	}

	return danach.mitFremdemSchritt(fremd, absender)
}

// passtZumStand hält fest, dass beide dieselbe Sache meinen.
func (a Handelsangebot) passtZumStand(fremd Handelsangebot) error {
	switch {
	case fremd.Vorgang != a.Vorgang:
		return fmt.Errorf(
			"%w: der Stand gehört zum Vorgang %q, hier läuft %q",
			ErrValidation, fremd.Vorgang, a.Vorgang)

	case fremd.Revision != a.Revision:
		return fmt.Errorf(
			"%w: der Stand ist Revision %d, hier gilt %d",
			ErrConflict, fremd.Revision, a.Revision)

	case fremd.Abdruck() != a.Abdruck():
		return fmt.Errorf(
			"%w: der Stand setzt auf %q auf, hier gilt %q — inzwischen hat "+
				"jemand geschrieben", ErrConflict, fremd.Abdruck(), a.Abdruck())
	}

	return nil
}

// mitFremdenStimmen trägt ein, was der Absender zugestimmt hat.
//
// **Über [Handelsangebot.Schreite] und nicht über die Scheibe**: eine Stimme,
// die hineingeschrieben würde, käme an der Frage vorbei, ob sie hier
// überhaupt zählt — und der Beschluss fiele nicht, wenn sie die letzte ist.
//
// # Warum nur seine eigene
//
// **Weil er über die anderen nichts zu sagen hat** (Befund 2 des Zweitblicks
// über Phase 5.6). Die Leitung prüft, **wer** spricht; was er über fremde
// Stimmen behauptet, prüft niemand. Ein Absender konnte deshalb beide
// Stimmen schicken, und der Vorgang stand auf `angenommen`, ohne dass die
// Gegenseite je zugestimmt hätte.
//
// **Ein Vertrag hat zwei Parteien** ([Handelsangebot.Pruefe]): was die andere
// zustimmt, sagt sie selbst — über ihre eigene Leitung, mit ihrem eigenen
// Handschlag.
func (a Handelsangebot) mitFremdenStimmen(
	fremd Handelsangebot, absender string,
) (Handelsangebot, error) {
	for _, stimme := range fremd.Zustimmungen {
		if stimme != absender || slices.Contains(a.Zustimmungen, stimme) {
			continue
		}

		beschluss, da := a.beschlussschritt()

		// **Die Stimme zu einem Abschluss reist nicht hier.** `bestaetigen`
		// ist ein Beschluss **und** ein Abschlussschritt: er zeichnet einen
		// Beleg gegen und verlangt einen Nachweis (`T10`). Wer ihn hier
		// nachspielte, brächte den Vorgang auf `beidseitigVerbucht`, ohne
		// dass die Verbuchungsstelle je gefragt worden wäre — genau der
		// Riegel, den `service.Handelsleitung.Schreiten` schützt.
		if da && IstAbschlussschritt(beschluss) {
			return a, nil
		}

		if !da {
			return a, fmt.Errorf(
				"%w: aus %q folgt kein Beschluss — die Stimme von %q gehört "+
					"zu einem anderen Stand", ErrConflict, a.Zustand, stimme)
		}

		gegangen, fehler := a.Schreite(beschluss, stimme)
		if fehler != nil {
			return a, fehler
		}

		a = gegangen
	}

	return a, nil
}

// OffenerBeschluss gibt den Schritt, dem die Stimmen dieses Zustands gelten
// — leer, wenn keiner alle Parteien braucht (TRADE-03, #1717).
func (a Handelsangebot) OffenerBeschluss() Handelsschritt {
	beschluss, _ := a.beschlussschritt()

	return beschluss
}

// beschlussschritt gibt den Schritt, der aus diesem Zustand alle Stimmen
// braucht — oder sagt, dass es keinen gibt.
func (a Handelsangebot) beschlussschritt() (Handelsschritt, bool) {
	for _, uebergang := range Handelsmatrix() {
		if uebergang.Von == a.Zustand && uebergang.Alle {
			return uebergang.Schritt, true
		}
	}

	return "", false
}

// mitFremdemSchritt vollzieht nach, was der Absender gegangen ist.
//
// **Höchstens einen, und nur den, der hierher führt.** Eine Kette
// nachzuspielen hieße, Zwischenstände zu erfinden, die niemand geschickt
// hat; was fehlt, kommt mit der nächsten Nachricht.
func (a Handelsangebot) mitFremdemSchritt(
	fremd Handelsangebot, absender string,
) (Handelsangebot, error) {
	if fremd.Zustand == a.Zustand {
		return a, nil
	}

	for _, uebergang := range Handelsmatrix() {
		if uebergang.Von != a.Zustand || uebergang.Nach != fremd.Zustand {
			continue
		}

		// **Die drei Abschlussschritte bleiben liegen** — der Grund steht im
		// Kopf dieser Datei.
		if IstAbschlussschritt(uebergang.Schritt) {
			return a, nil
		}

		gegangen, fehler := a.Schreite(uebergang.Schritt, absender)
		if fehler != nil {
			return a, fehler
		}

		return gegangen, nil
	}

	// **Kein Weg von hier nach dort ist kein Fehler.** Drüben kann jemand
	// weiter sein, als ein einzelner Schritt trägt; was fehlt, kommt mit der
	// nächsten Nachricht.
	return a, nil
}
