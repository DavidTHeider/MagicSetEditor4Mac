package domain

// **Die Einwilligung, dass später jemand mitliest** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Wofür es sie gibt
//
// Das Issue verlangt „Consent für gegebenenfalls spätere Offenlegung". Das ist
// die Gegenrichtung zu [RueckschauFuer]: dort wird **abgeschnitten**, was
// einem Leser nicht gehört — hier sagt der, dem es gehört, dass er es
// freigibt.
//
// # Warum sie den laufenden Tisch nicht anfasst
//
// **Weil eine Einwilligung von vorhin keine von jetzt ist.** Am Tisch rechnet
// [Partie.Sicht] je Sitz, und was dort nicht ankommt, kann auch nicht
// versehentlich verraten werden. Eine Freigabe, die dort schon wirkte, machte
// die laufende Partie zu einer anderen: wer sein Blatt für die Nachlese
// freigibt, hat nicht zugesagt, es **während** des Spiels zu zeigen.
//
// **Deshalb steht sie in der Rückschau und nirgends sonst.** Der Wächter
// dafür ist, dass [Ereignis.Fuer] sie nicht kennt.
//
// # Und warum sie widerruflich ist, aber nichts zurückholt
//
// **Wer sie zurücknimmt, schließt die Tür — er sammelt nicht ein, was
// hindurchging.** Das Issue sagt an anderer Stelle denselben Satz für die
// Deckleihe: keine „beweisbare Löschung beim Empfänger". Eine Rücknahme, die
// mehr verspräche, verspräche etwas, das dieses Programm nicht halten kann
// (§10).

import "fmt"

// OffenlegungAnAlle steht für „jeder darf".
//
// **Ein Stern und keine leere Liste.** Leer heißt hier „niemand"; wer beides
// auf denselben Wert legte, machte aus einem vergessenen Feld eine
// Veröffentlichung.
const OffenlegungAnAlle = "*"

// AufbewahrungPartien ist die Zahl der Tage, nach denen ein **beendetes**
// Protokoll zur Entscheidung ansteht.
//
// # Warum eine Zahl und keine Einstellung
//
// **Weil sie nichts löscht.** Die Datenpflege **bietet an**, und die Klasse
// dahinter ist `entscheidung`: nach neunzig Tagen steht ein beendetes
// Protokoll in der Liste, und ein Mensch sagt ja oder nein. Eine Zahl, über
// die niemand entscheidet, wäre eine automatische Löschung — und die verlangt
// dieses Issue ausdrücklich nicht.
//
// **Neunzig und nicht dreißig**: ein Turnier läuft über ein Wochenende, eine
// Klärung darüber über Wochen. Wer den Verlauf einer strittigen Partie sucht,
// sucht ihn später, als man denkt.
//
// # Und warum sie hier steht und nicht in der Abfrage
//
// **Weil sie eine fachliche Aussage ist.** Eine Zahl in einem SQL-Text stünde
// an der einen Stelle, an der niemand nach ihr sucht — und der Lauf daneben
// müsste sie abschreiben. Zwei Zahlen über dieselbe Frist laufen auseinander
// (§10).
const AufbewahrungPartien = 90

// Offenlegung ist die Einwilligung **eines** Sitzes.
type Offenlegung struct {
	// Von ist der Sitz, dessen private Ergänzungen es sind.
	//
	// **Der Empfänger einer [Ergaenzung] ist ihr Eigner.** Wer zieht,
	// bekommt die Ergänzung, die sagt, was er gezogen hat — sie gehört ihm,
	// und nur er kann sie freigeben.
	Von string

	// An sind die Sitze, die mitlesen dürfen — oder [OffenlegungAnAlle].
	An []string
}

// Pruefe sagt, ob eine Einwilligung eine ist.
func (o Offenlegung) Pruefe() error {
	if o.Von == "" {
		return fmt.Errorf(
			"%w: eine Offenlegung braucht den Sitz, der sie gibt", ErrValidation)
	}

	if len(o.An) == 0 {
		return fmt.Errorf(
			"%w: eine Offenlegung ohne Empfänger gibt nichts frei — dann gibt "+
				"es sie nicht", ErrValidation)
	}

	for _, an := range o.An {
		if an == "" {
			return fmt.Errorf(
				"%w: ein leerer Empfänger ist keiner", ErrValidation)
		}

		// **Sich selbst freizugeben ist keine Freigabe.** Wer es hinnähme,
		// hätte eine Einwilligung, die nichts bewirkt — und jemanden, der
		// glaubt, er habe etwas geteilt.
		if an == o.Von {
			return fmt.Errorf(
				"%w: %q sieht seine eigenen Ergänzungen ohnehin",
				ErrValidation, o.Von)
		}
	}

	return nil
}

// Mitleser sagt, wessen Ergänzungen ein Leser sehen darf.
//
// **Er selbst steht immer darin** — außer er ist niemand: ein leerer Leser ist
// das Protokoll ohne Angabe, und der sieht nur, was [OffenlegungAnAlle]
// freigegeben hat.
func Mitleser(freigaben []Offenlegung, fuer string) map[string]bool {
	darf := map[string]bool{}

	if fuer != "" {
		darf[fuer] = true
	}

	for _, freigabe := range freigaben {
		for _, an := range freigabe.An {
			// **`an == fuer` erst, wenn es einen Leser gibt.** Ohne diese
			// Bedingung träfe ein leeres `an` auf einen leeren Leser, und
			// eine vergessene Zeile gäbe frei, was niemand freigeben wollte.
			if an == OffenlegungAnAlle || (fuer != "" && an == fuer) {
				darf[freigabe.Von] = true
			}
		}
	}

	return darf
}
