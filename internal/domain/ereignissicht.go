package domain

// **Was ein Empfänger von einem Ereignis sieht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Neben `partiesicht.go` und nicht darin**: dort wird ein **Zustand**
// projiziert, hier ein **Ereignis**. Zwei Fragen, und die Datei stand mit
// beiden bei 408 Zeilen (§2).

// Fuer gibt dieses Ereignis, so weit `wer` es bekommen darf.
//
// **Der öffentliche Teil geht an jeden**, die Ergänzung nur an ihren
// Empfänger — und die Kette bleibt dieselbe, damit ein Empfänger prüfen kann,
// was er bekommen hat.
func (e Ereignis) Fuer(wer string) Ereignis {
	darf := map[string]bool{}
	if wer != "" {
		darf[wer] = true
	}

	return e.FuerMitleser(darf)
}

// FuerMitleser kürzt auf das, was eine Menge von Sitzen sehen darf.
//
// **Die Rechnung, und [Ereignis.Fuer] ist ihr einfacher Fall.** Es gibt sie
// zweimal, weil es zwei Fragen gibt: am Tisch sieht jeder nur sich selbst, in
// der Nachlese kommt dazu, was ihm jemand freigegeben hat ([Offenlegung]).
// Zwei **Rechnungen** wären es erst, wenn beide Stellen zählten — und die
// zweite vergäße beim nächsten neuen Feld, was die erste kann (§10).
//
// **Und keine Kennung geht hinaus** (#2480, Vermerk 0201): was das Ereignis
// nennt, steht hier unter der Sichtkennung seines Augenblicks.
func (e Ereignis) FuerMitleser(darf map[string]bool) Ereignis {
	namen := e.zuordnung()

	gekuerzt := e
	gekuerzt.Privat = nil

	for _, ergaenzung := range e.Privat {
		if !darf[ergaenzung.Empfaenger] {
			continue
		}

		// **Nur mit einer Zuordnung umgeschrieben** — ohne sie bliebe ein
		// Wert, der sich nicht kodieren lässt, nicht stehen, und der
		// Sichtabdruck ginge über etwas anderes als das Ereignis.
		if len(namen) > 0 {
			ergaenzung.Inhalt, _ = umbenannt(ergaenzung.Inhalt, namen).(map[string]any) //nolint:errcheck // eine Tabelle bleibt eine
		}

		gekuerzt.Privat = append(gekuerzt.Privat, ergaenzung)
	}

	if len(namen) > 0 {
		gekuerzt.Oeffentlich, _ = umbenannt(e.Oeffentlich, namen).(map[string]any) //nolint:errcheck // dito
	}

	return gekuerzt
}
