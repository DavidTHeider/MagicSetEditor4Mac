package domain

// **Die zustandsbasierten Aktionen, die einen Sitz kosten** — Werkstück M17
// der Regelmaschine
// (#1618).
//
// # Drei von fünfundzwanzig, und warum gerade diese drei
//
// **Weil `M17` ihre Wirkung einführt.** Es gibt vier Ausscheidungsgründe
// ([Ausscheidungsgrund]); einer davon — das Aufgeben — ist eine Handlung und
// hat seinen Befehl. Die anderen drei sind **Zustände**, und einen Zustand
// stellt niemand fest, solange ihn niemand prüft. Ein Grund, den nichts
// erzeugt, ist ein Versprechen, das keiner einlöst (§10).
//
// **Die übrigen zweiundzwanzig baut `M5`**
// (#1607) —
// sie brauchen Eigenschaften am Objekt (`M2`), den Stapel (`M1`) und
// Anhängsel (`M12`), und keines davon gibt es hier. Die Naht dafür ist
// **diese Funktion**: `M5` erweitert sie, statt eine zweite daneben zu legen.
//
// # Wann sie läuft
//
// **CR 704.3** (Fassung 2026-08-19): immer dann, wenn ein Spieler Priorität
// bekäme — also in [Partie.abschnittBetreten], bevor sie vergeben wird —, und
// nach jedem Befehl, der einen der drei Zustände ändert.
//
// **Gleichzeitig und nicht nacheinander** — CR 704.4: stehen zwei Sitze bei
// null Leben, verlieren sie zusammen, und daraus wird ein Unentschieden
// (CR 104.4a). Zwei Aufrufe hintereinander machten aus demselben Zustand
// einen Sieger.

// Vergiftungsgrenze ist die Zahl, ab der ein Spieler verliert — CR 704.5c
// (Fassung 2026-08-19).
//
// **Zehn, und die Zahl steht hier und nicht im Profil**: anders als das
// Startleben (CR 119.1) kennt die CR dafür keine Variante. Wer sie einmal
// verstellbar braucht, zieht sie an dieselbe Stelle wie das Startleben.
const Vergiftungsgrenze = 10

// Zustandspruefung wendet die zustandsbasierten Aktionen an, die einen Sitz
// ausscheiden lassen — CR 704.5a, CR 704.5b und CR 704.5c (Fassung
// 2026-08-19).
//
// **Sie gibt dieselbe Partie zurück, wenn nichts anliegt** — und ist damit
// gefahrlos so oft aufzurufen, wie CR 704.3 es verlangt.
func (p Partie) Zustandspruefung() Partie {
	// **Sie wiederholen sich, bis keine mehr anliegt** — CR 704.3 (M5,
	// #1607).
	//
	// **Und die Schleife endet ohne Obergrenze.** Jede Runde, die etwas tut,
	// **nimmt** etwas weg: Marken, ein Objekt, ein Anhängsel, einen Sitz.
	// Nichts davon kommt zurück, und keine dieser Aktionen legt etwas an.
	// Eine Zahl daneben wäre ein Zweig, den kein Lauf betreten kann (§10).
	for {
		danach, etwas := p.zustandsrunde()

		p = danach

		if !etwas {
			return p
		}
	}
}

// zustandsrunde wendet **eine** Runde an — und sagt, ob etwas geschah.
//
// # Die Reihenfolge ist keine Rangfolge
//
// **CR 704.4 lässt alle Aktionen gleichzeitig laufen.** Sie steht so, weil
// jede Stufe die nächste lesen kann: das Verrechnen der Marken (CR 704.5q)
// ändert die Widerstandskraft, die CR 704.5f liest; ein Objekt, das in den
// Friedhof geht, lässt eine Aura an nichts hängen (CR 704.5m). Wer sie
// anders ordnete, bekäme dasselbe Ergebnis — nur eine Runde später.
func (p Partie) zustandsrunde() (Partie, bool) {
	p, marken := p.markenVerrechnen()
	p, objekte := p.objektaktionen()
	p, anhaengsel := p.anhaengselaktionen()
	p, legenden := p.legendenpruefen()

	gegangen := p.sitzeAusscheiden()

	return gegangen, marken || objekte || anhaengsel || legenden ||
		len(gegangen.Ausgang.Ausgeschieden) != len(p.Ausgang.Ausgeschieden)
}

// sitzeAusscheiden wendet an, was einen **Sitz** kostet — CR 704.5a–c.
func (p Partie) sitzeAusscheiden() Partie {
	var gehen []Ausgeschieden

	for _, sitz := range p.Sitze {
		if p.istDraussen(sitz.Kennung) {
			continue
		}

		if grund, trifft := grundFuer(sitz, p.Profil); trifft {
			gehen = append(gehen, Ausgeschieden{Sitz: sitz.Kennung, Grund: grund})
		}
	}

	if len(gehen) == 0 {
		return p
	}

	return p.Scheidenaus(gehen)
}

// grundFuer sagt, ob dieser Sitz die Partie verlässt, und warum.
//
// # Die Reihenfolge ist keine Rangfolge
//
// **CR 704.4 lässt alle Aktionen gleichzeitig laufen**, und ein Sitz kann
// aus mehreren Gründen zugleich gehen — null Leben **und** zehn Marken. Er
// geht aber nur einmal, und dann trägt der Ausgang **einen** Grund. Welcher
// das ist, sagt die CR nicht; hier steht die Reihenfolge der Regelnummern,
// damit die Antwort nicht von der Sitzreihenfolge abhängt.
//
// # Warum das Profil mitkommt
//
// **Weil ein Spiel ohne Lebenspunkte CR 704.5a nicht kennt.**
// [Regelprofil.Startleben] sagt es ausdrücklich: „0 heißt: das Spiel kennt
// keine." Ohne diese Bedingung verlöre in einem solchen Spiel jeder sofort —
// die Regel griffe, wo es die Größe gar nicht gibt, auf die sie sich
// bezieht.
func grundFuer(sitz Sitz, profil Regelprofil) (Ausscheidungsgrund, bool) {
	switch {
	// CR 704.5a — null oder weniger Leben.
	case profil.Startleben > 0 && sitz.Leben <= 0:
		return GrundLeben, true

	// CR 704.5b — wer aus einer leeren Bibliothek ziehen wollte.
	case sitz.Leergezogen:
		return GrundBibliothek, true

	// CR 704.5c — zehn oder mehr Vergiftungsmarken.
	case sitz.Vergiftung >= Vergiftungsgrenze:
		return GrundGift, true

	// **CR 903.10a** — einundzwanzig Kampfschaden desselben Commanders
	// (M19, #1620).
	//
	// **Sie steht hier und nicht in CR 704.5**: die Variante fügt sie
	// hinzu, und CR 903.10a nennt sie ausdrücklich eine zustandsbasierte
	// Aktion. Dass sie nur in einer Commanderpartie greift, entscheidet
	// das **Profil** — dieselbe Bedingung wie beim Startleben darüber.
	case commanderschlag(sitz, profil):
		return GrundCommander, true
	}

	return "", false
}
