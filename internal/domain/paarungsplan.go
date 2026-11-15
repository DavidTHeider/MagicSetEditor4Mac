// **Der Paarungsplan einer Runde** — Werkstück 4 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Was nachgelesen ist und was nicht
//
// **Nachgelesen** (MTR §10.4): ohne andere Ankündigung gilt für ein
// Turnier das Schweizer Paarungsverfahren.
//
// **Nicht nachgelesen**: *wie* Swiss paart. Die MTR **nennt** den Algorithmus
// und beschreibt ihn nicht. Alles Folgende ist deshalb die übliche Gestalt und
// damit eine **eigene** Entscheidung — sie steht hier, damit niemand sie für
// nachgelesen hält:
//
//   - nach Stand sortieren (in der ersten Runde ist das die Anmeldung),
//   - benachbart paaren,
//   - eine Wiederholung durch **Herunterpaaren** vermeiden,
//   - und der Bye geht an den Letzten, der noch keinen hatte.
//
// # Warum nicht ausgelost wird
//
// Weil ein Turnier, das sich nach einem Absturz anders fortsetzt, keines ist.
// Derselbe Verlauf gibt denselben Plan.

package domain

// Paarungsplan sagt, wie die nächste Runde zu paaren ist.
//
// **Er schreibt nichts.** Wer ihn nimmt, legt damit eine Runde an — oder
// nicht. Ein Plan, der schon Tatsachen schüfe, ließe sich nicht ansehen,
// bevor man ihn nimmt.
func Paarungsplan(turnier Turnier) []Paarung {
	offen := spielbereit(turnier)
	if len(offen) == 0 {
		return nil
	}

	gespielt, byes := verlaufskenntnis(turnier)

	plan := make([]Paarung, 0, len(offen)/2+1)

	// **Der Bye wird zuerst bestimmt und zuletzt angehängt.** Zuerst, weil
	// er den Übrigen aus der Paarung nimmt; zuletzt, weil ein Turnierleiter
	// die Tische der Reihe nach liest und der Tisch ohne Gegner keiner ist.
	var mitBye int64

	if len(offen)%2 == 1 {
		mitBye = byeempfaenger(offen, byes)
		offen = ohne(offen, mitBye)
	}

	plan = append(plan, paare(offen, gespielt)...)

	if mitBye != 0 {
		plan = append(plan, Paarung{A: mitBye})
	}

	return plan
}

// spielbereit gibt die Teilnehmer in Standreihenfolge — ohne Ausgestiegene.
func spielbereit(turnier Turnier) []int64 {
	var offen []int64

	ausgestiegen := make(map[int64]bool, len(turnier.Teilnehmer))
	for _, teilnehmer := range turnier.Teilnehmer {
		ausgestiegen[teilnehmer.ID] = teilnehmer.Ausgestiegen
	}

	// **Die Rangliste ist die Standreihenfolge** — und in der ersten Runde
	// ist sie die Anmeldung, weil dann alle Werte null sind und der letzte
	// Vergleich die Anmeldestelle ist.
	for _, rang := range Rangliste(turnier) {
		if !ausgestiegen[rang.TeilnehmerID] {
			offen = append(offen, rang.TeilnehmerID)
		}
	}

	return offen
}

// verlaufskenntnis sagt, wer gegen wen gespielt hat und wer einen Bye hatte.
func verlaufskenntnis(turnier Turnier) (
	gespielt map[[2]int64]bool, byes map[int64]bool,
) {
	gespielt = map[[2]int64]bool{}
	byes = map[int64]bool{}

	for _, runde := range turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if paarung.IstBye() {
				byes[paarung.A] = true

				continue
			}

			gespielt[[2]int64{paarung.A, paarung.B}] = true
			gespielt[[2]int64{paarung.B, paarung.A}] = true
		}
	}

	return gespielt, byes
}

// byeempfaenger sagt, wer den Bye bekommt.
//
// **Der Letzte im Stand, der noch keinen hatte** — und wenn alle einen
// hatten, der Letzte. Eine Runde, die wegen eines Byes ausfiele, wäre die
// schlechtere Antwort: sie sagt dem Turnierleiter nichts.
func byeempfaenger(offen []int64, byes map[int64]bool) int64 {
	for stelle := len(offen) - 1; stelle >= 0; stelle-- {
		if !byes[offen[stelle]] {
			return offen[stelle]
		}
	}

	return offen[len(offen)-1]
}

// Suchschritte ist, wie viele Züge die wiederholungsfreie Suche höchstens tut.
//
// **Eine Grenze, weil die Suche im schlimmsten Fall exponentiell ist.** In
// jedem wirklichen Turnier greift der erste Kandidat fast immer, und die
// Suche kostet dann so viel wie die alte kurzsichtige. Der Ausweg gilt einem
// Feld, das jeder gegen jeden gespielt hat — dort gibt es keine Lösung, und
// ohne Grenze suchte die Rekursion sie bis zum Ende.
//
// **Was gemessen ist** (`suchschritte_intern_test.go`, 2026-09-01):
//
//	64 Teilnehmer, nichts gespielt   →  32 Schritte (einer je Paarung)
//	20 Teilnehmer, jeder gegen jeden →  19 Schritte, dann der Ausweg
//
// **Und was nicht gemessen ist**: der schlimmste Fall. Er liegt weder bei
// „alles offen" noch bei „alles gespielt", sondern dazwischen — dort, wo es
// viele halbe Lösungen gibt und keine ganze. Genau deshalb steht die Grenze
// hier: nicht weil sie greifen wird, sondern weil niemand weiß, wann.
const Suchschritte = 10000

// paare paart die offene Liste — **ohne vermeidbare Wiederholung**.
//
// # Warum es dafür eine Suche braucht
//
// Der erste Anlauf nahm den Ersten und für ihn den nächstbesten Gegner, gegen
// den er noch nicht gespielt hatte. Das sieht nicht, was die Wahl für die
// **Übrigen** bedeutet: bei der Rangfolge 1, 2, 3, 4, wenn nur 3–4 schon
// gespielt hat, wählt es 1–2 und muss danach 3–4 wiederholen — obwohl 1–3
// und 2–4 ganz ohne Wiederholung gegangen wären. Befund 5 des Zweitblicks
// über Phase 4.7.
//
// **Die Suche nimmt die erste vollständige Lösung**, und sie probiert die
// Kandidaten in Standreihenfolge. Damit bleibt der Plan derselbe, wo die
// kurzsichtige Wahl schon richtig war — und derselbe Verlauf gibt weiterhin
// denselben Plan.
//
// **Geht es gar nicht, wird wiederholt.** Eine Runde ohne Tische sagt dem
// Turnierleiter nichts.
func paare(offen []int64, gespielt map[[2]int64]bool) []Paarung {
	schritte := 0

	if gefunden, gut := ohneWiederholung(offen, gespielt, &schritte); gut {
		return gefunden
	}

	// **Geht es nicht ohne, dann mit so wenigen wie möglich.**
	//
	// Hier stand nur `kurzsichtig`, und das greift der Reihe nach zum ersten
	// freien Gegner — auch wenn es damit den einzigen freien Gegner eines
	// anderen wegnimmt (Befund 6 des Nachlaufs über Phase 4.7).
	//
	// **Gemessen, nicht vermutet.** Über alle 32.768 möglichen Verläufe bei
	// **sechs** Teilnehmern gab das Kurzsichtige in **1.483** Fällen mehr
	// Wiederholungen als nötig — 4,5 %; der krasseste ist drei statt zwei.
	// Mit dieser Suche sind es **null**.
	//
	// **Die Messung reicht so weit wie sie reicht** (§10): sie sagt etwas
	// über sechs Teilnehmer und nichts über sechzig. Was zusätzlich gemessen
	// ist, ist die **Laufzeit** — acht bis vierundsechzig Teilnehmer, alles
	// gegen alles gespielt, zusammen unter einer halben Sekunde. Das ist der
	// teuerste Fall für die Suche, und die Schranke greift dort sofort: die
	// erste Lösung kostet so viele Wiederholungen wie es Tische gibt, und
	// besser geht es nicht.
	schritte = 0

	return wenigsteWiederholungen(offen, gespielt, &schritte)
}

// wenigsteWiederholungen sucht die Paarung mit den wenigsten Wiederholungen.
//
// **Mit Schranke**: ein Zweig, der schon jetzt nicht besser ist als die beste
// gefundene Lösung, wird nicht weiter verfolgt. Ohne sie wäre es die volle
// Aufzählung aller Paarungen, und die wächst mit (n-1)!!.
//
// **Die Kandidaten werden in Standreihenfolge probiert**, wie oben — damit
// bleibt der Plan derselbe, wo die kurzsichtige Wahl schon richtig war, und
// derselbe Verlauf gibt weiterhin denselben Plan.
//
// **Sie fängt beim Kurzsichtigen an** und verbessert es. Das hat zwei
// Gründe, und der zweite ist der wichtigere: die Schranke sitzt vom ersten
// Zweig an eng — und es gibt **keinen Fall ohne Antwort**. Ein Ausweg, der
// nur greift, wenn das Suchbudget vor der ersten vollständigen Lösung
// aufgebraucht ist, wäre bei 10.000 Schritten und einer Lösung in n/2
// Schritten unerreichbar: toter Code, der eine Zusage macht.
func wenigsteWiederholungen(
	offen []int64, gespielt map[[2]int64]bool, schritte *int,
) []Paarung {
	beste := kurzsichtig(offen, gespielt)
	bester := gezaehlteWiederholungen(beste, gespielt)

	var suche func(rest []int64, bisher []Paarung, wiederholt int)

	suche = func(rest []int64, bisher []Paarung, wiederholt int) {
		if wiederholt >= bester {
			return
		}

		if len(rest) == 0 {
			beste = append([]Paarung(nil), bisher...)
			bester = wiederholt

			return
		}

		a := rest[0]

		for stelle, kandidat := range rest[1:] {
			*schritte++
			if *schritte > Suchschritte {
				return
			}

			weiter := make([]int64, 0, len(rest)-2)
			weiter = append(weiter, rest[1:stelle+1]...)
			weiter = append(weiter, rest[stelle+2:]...)

			kosten := 0
			if gespielt[[2]int64{a, kandidat}] {
				kosten = 1
			}

			suche(weiter, append(append([]Paarung(nil), bisher...),
				Paarung{A: a, B: kandidat}), wiederholt+kosten)
		}
	}

	suche(offen, nil, 0)

	return beste
}

// gezaehlteWiederholungen sagt, wie viele Paarungen eines Plans es schon gab.
func gezaehlteWiederholungen(plan []Paarung, gespielt map[[2]int64]bool) int {
	var gezaehlt int

	for _, paarung := range plan {
		if gespielt[[2]int64{paarung.A, paarung.B}] {
			gezaehlt++
		}
	}

	return gezaehlt
}

// ohneWiederholung sucht eine Paarung, in der niemand zweimal gegen denselben
// spielt.
func ohneWiederholung(
	offen []int64, gespielt map[[2]int64]bool, schritte *int,
) ([]Paarung, bool) {
	if len(offen) == 0 {
		return nil, true
	}

	a := offen[0]

	for stelle, kandidat := range offen[1:] {
		*schritte++
		if *schritte > Suchschritte {
			return nil, false
		}

		if gespielt[[2]int64{a, kandidat}] {
			continue
		}

		rest := make([]int64, 0, len(offen)-2)
		rest = append(rest, offen[1:stelle+1]...)
		rest = append(rest, offen[stelle+2:]...)

		weiter, gut := ohneWiederholung(rest, gespielt, schritte)
		if gut {
			return append([]Paarung{{A: a, B: kandidat}}, weiter...), true
		}
	}

	return nil, false
}

// kurzsichtig paart der Reihe nach und nimmt Wiederholungen in Kauf.
//
// **Der Ausweg, wenn es keine wiederholungsfreie Paarung gibt** — und der
// einzige Weg, der immer eine Antwort gibt.
func kurzsichtig(offen []int64, gespielt map[[2]int64]bool) []Paarung {
	plan := make([]Paarung, 0, len(offen)/2)

	for len(offen) > 0 {
		a := offen[0]
		gegen := gegner(a, offen[1:], gespielt)

		plan = append(plan, Paarung{A: a, B: gegen})
		offen = ohne(ohne(offen, a), gegen)
	}

	return plan
}

// gegner sucht den besten Gegner, gegen den noch nicht gespielt wurde.
//
// **Herunterpaaren**: wer als Nächster käme und schon gespielt hat, wird
// übersprungen.
func gegner(a int64, uebrige []int64, gespielt map[[2]int64]bool) int64 {
	for _, kandidat := range uebrige {
		if !gespielt[[2]int64{a, kandidat}] {
			return kandidat
		}
	}

	return uebrige[0]
}

// ohne gibt die Liste ohne einen Eintrag.
func ohne(alle []int64, wer int64) []int64 {
	uebrig := make([]int64, 0, len(alle))

	for _, einer := range alle {
		if einer != wer {
			uebrig = append(uebrig, einer)
		}
	}

	return uebrig
}
