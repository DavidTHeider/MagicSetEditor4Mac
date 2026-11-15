package domain

// **Einen Zauber wirken** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Der Vorgang hat Schritte, und erst am Ende ist etwas gewirkt
//
// **CR 601.2 nennt neun** (a bis i). Dieses Programm geht die, die es ohne
// Kartentext gehen kann:
//
//   - **601.2a** — die Karte wandert von dort, wo sie liegt, auf den Stapel,
//     und wer sie ankündigt, beherrscht sie.
//   - **601.2c** — die Ziele werden genannt (CR 115).
//   - **601.2f** — die Gesamtkosten stehen fest; hier sind sie die
//     Manakosten der Karte, denn Zusatzkosten stehen in ihrem Text.
//   - **601.2h** — bezahlt wird **ganz**, aus dem Manavorrat.
//   - **601.2i** — danach ist er gewirkt.
//
// **Die übrigen brauchen einen Kartentext**: Modi (601.2b), Wertverteilung
// (601.2d), `{X}` (601.2e/g) und Kostenveränderungen — und einen Text
// auszuführen ist `M8`
// (#1610).
//
// # Und was schiefgeht, geht ganz schief
//
// **CR 733 verlangt das Zurückspulen einer unzulässigen Handlung**, und hier
// ist es geschenkt: eine [Partie] ist ein **Wert**. Wer mitten im Vorgang
// scheitert, bekommt die Partie zurück, die er hineingegeben hat — kein
// halber Stand, kein Mana weg, keine Karte auf dem Stapel. Das ist kein
// Zufall, sondern der Grund, warum der Kern mit Werten arbeitet; deshalb
// steht es hier und nicht nur im Lauf.

import "fmt"

// Wirken kündigt einen Zauber an, bezahlt ihn und legt ihn auf den Stapel —
// CR 601.2.
//
// **Ziele dürfen leer sein**: die meisten Zauber haben keines, und CR 115.1
// gilt nur, wo einer verlangt wird. **Was ein Zauber verlangt, steht in
// seinem Text** — ohne `M8` weiß das niemand; geprüft wird deshalb nur, ob
// die **genannten** Ziele zulässig sind.
func (p Partie) Wirken(wer, objekt string, ziele []string) (Partie, error) {
	if p.Zug.Nummer > 0 && p.Zug.Prioritaet != wer {
		return p, fmt.Errorf("%w: %q hat keine Priorität (CR 117.1)",
			ErrValidation, wer)
	}

	stelle, fehler := p.wirkstelle(wer, objekt)
	if fehler != nil {
		return p, fehler
	}

	gefasst, fehler := p.zielePruefen(ziele)
	if fehler != nil {
		return p, fehler
	}

	kosten, fehler := p.wirkkosten(p.Objekte[stelle])
	if fehler != nil {
		return p, fehler
	}

	// **Und die Commandersteuer, wo sie gilt** — CR 903.8 (M19,
	// #1620).
	// **Zusätzliche Kosten sonst gibt es nicht** (CR 601.2f, M8); diese
	// eine steht im Regelwerk statt auf einer Karte, und darum steht sie
	// hier.
	kosten = p.mitCommandersteuer(kosten, wer, p.Objekte[stelle])

	// **Erst bezahlen, dann hinaufstellen** — die Reihenfolge ist hier
	// sichtbar: scheitert die Zahlung, liegt nichts auf dem Stapel.
	bezahlt, fehler := p.manaAbziehen(wer, kosten)
	if fehler != nil {
		return p, fehler
	}

	objekte := make([]Spielobjekt, len(bezahlt.Objekte))
	copy(objekte, bezahlt.Objekte)

	// **Der neue Zauber kommt obenauf** — CR 405.2: der Stapel wird von oben
	// abgeräumt, und wer antwortet, wird zuerst verrechnet. **Das Rücken
	// erledigt [Partie.zonenwechsel]** seit `M18` (#1619) — es ist dieselbe
	// Regel wie für den Friedhof (CR 404.1).

	// **Und auf den Stapel geht ein neues Objekt** — CR 400.7 (M18,
	// #1619):
	// was in der Hand daran hing oder darauf lag, bleibt dort zurück.
	bezahlt.Objekte = objekte
	bezahlt = bezahlt.zonenwechsel(stelle, ZoneStapel)
	objekte = bezahlt.Objekte

	// **Wer ankündigt, beherrscht** — CR 601.2a. Der Eigner bleibt, wer er
	// war (CR 108.3); daran hängt, wohin die Karte später abgelegt wird.
	objekte[stelle].Beherrscher = wer

	// **Die Ziele stehen nach dem Zonenwechsel** — CR 601.2c setzt sie beim
	// Ankündigen, und CR 400.7 räumt die des **vorigen** Daseins weg. Wer
	// sie davor setzte, setzte sie an ein Objekt, das es nicht mehr gibt.
	objekte[stelle].Ziele = gefasst

	bezahlt.Objekte = objekte

	// **Und der Zauberzähler steigt** — CR 731.2a/b (M20, #1621) fragt, wie
	// viele Zauber der aktive Spieler in diesem Zug gewirkt hat.
	bezahlt = bezahlt.zauberZaehlen(wer)

	return bezahlt.steuerZaehlen(wer, p.Objekte[stelle]), nil
}

// zauberZaehlen merkt sich einen gewirkten Zauber — CR 731.2a.
//
// **Nur die des aktiven Spielers**, und zwar, weil CR 731.2a nach ihm fragt.
// Ein Zauber des Gegenübers im selben Zug zählt für den Wechsel nicht.
func (p Partie) zauberZaehlen(wer string) Partie {
	if p.Zug.Nummer > 0 && p.Zug.Aktiv == wer {
		p.Zug.Zauber++
	}

	return p
}

// wirkstelle findet die Karte und prüft, ob sie gewirkt werden darf.
func (p Partie) wirkstelle(wer, objekt string) (int, error) {
	for stelle := range p.Objekte {
		if p.Objekte[stelle].Kennung != objekt {
			continue
		}

		switch {
		case p.Objekte[stelle].Beherrscher != wer:
			return 0, fmt.Errorf(
				"%w: %q beherrscht das Objekt %q nicht (CR 601.2a)",
				ErrValidation, wer, objekt)
		case p.Objekte[stelle].Zone == ZoneStapel:
			return 0, fmt.Errorf(
				"%w: das Objekt %q liegt schon auf dem Stapel",
				ErrValidation, objekt)
		}

		return stelle, nil
	}

	return 0, fmt.Errorf("%w: das Objekt %q gibt es in dieser Partie nicht",
		ErrNotFound, objekt)
}

// wirkkosten gibt die Manakosten der Karte — CR 601.2f.
//
// **Die gedruckten und keine anderen.** Zusätzliche, alternative und
// veränderte Kosten stehen im Text der Karte oder in einem Effekt; beides
// braucht `M8` beziehungsweise `M10`, und das steht als Grenze in der
// Regelmatrix.
func (p Partie) wirkkosten(objekt Spielobjekt) (Manakosten, error) {
	eigenschaften, gefunden := p.Eigenschaften(objekt)
	if !gefunden {
		return Manakosten{}, fmt.Errorf(
			"%w: zu %q steht keine Karte im eingefrorenen Abzug",
			ErrNotFound, objekt.Kennung)
	}

	if befunde := Unbezahlbar(eigenschaften.Manakosten); len(befunde) > 0 {
		return Manakosten{}, fmt.Errorf(
			"%w: die Kosten von %q lassen sich hier nicht zahlen: %s",
			ErrValidation, eigenschaften.Name, befunde[0])
	}

	return eigenschaften.Manakosten, nil
}

// manaAbziehen bezahlt die Kosten aus dem Vorrat — CR 601.2h.
//
// **Ganz oder gar nicht**: eine Teilzahlung gibt es nicht, und was nicht
// aufgeht, ändert am Vorrat nichts.
func (p Partie) manaAbziehen(wer string, kosten Manakosten) (Partie, error) {
	for stelle := range p.Sitze {
		if p.Sitze[stelle].Kennung != wer {
			continue
		}

		danach, ging := p.Sitze[stelle].Manavorrat.Bezahle(kosten)
		if !ging {
			return p, fmt.Errorf(
				"%w: der Vorrat %q deckt die Kosten %q nicht (CR 601.2h)",
				ErrValidation, vorratswort(p.Sitze[stelle].Manavorrat),
				kosten.Roh)
		}

		sitze := make([]Sitz, len(p.Sitze))
		copy(sitze, p.Sitze)
		sitze[stelle].Manavorrat = danach
		p.Sitze = sitze

		return p, nil
	}

	return p, fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
		ErrNotFound, wer)
}

// vorratswort gibt einen Vorrat als Buchstabenfolge — für die Meldung.
//
// **Es steht hier und nicht in `service`**, obwohl es dort schon einmal
// steht: `domain` kennt `service` nicht (§2), und eine Meldung ohne den
// tatsächlichen Vorrat ließe jemanden raten, woran es lag.
func vorratswort(vorrat Manavorrat) string {
	var text string

	for _, art := range Manaarten() {
		for range vorrat.Zahl(art) {
			text += string(art)
		}
	}

	if text == "" {
		return "(leer)"
	}

	return text
}
