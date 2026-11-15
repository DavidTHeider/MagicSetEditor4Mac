// **Was die Sammlung wert ist — mit allem, was dieser Satz nicht sagt**
// (#1165).
//
// # Die Zahl trägt ihre Vorbehalte mit
//
// Die Quelle sagt über ihre Preise ausdrücklich: „daily estimates and/or
// market values provided by our affiliates. **Absolutely no guarantee** is
// made for any price information" (nachgelesen am 2026-08-27). Eine Summe
// darüber ist deshalb keine Bewertung, sondern eine Rechnung über Schätzwerte.
//
// **Drei Dinge stehen deshalb neben jeder Summe** und nicht in einer Fußnote:
//
//   - **wie viele Exemplare gezählt wurden** und wie viele **ohne Preis**
//     blieben,
//   - **wie alt** der älteste Preis ist, der eingegangen ist,
//   - **welche Quelle** ihn genannt hat.
//
// Wer die Summe ohne diese drei zeigt, zeigt eine Zahl, für die niemand
// einsteht.

package service

import (
	"context"
	"sort"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Portfoliozeile ist die Summe einer Währung.
type Portfoliozeile struct {
	Waehrung string

	// Wert ist die Summe in der kleinsten Einheit der Währung.
	Wert int64

	// Exemplare ist die Zahl der **Stücke**, die eingegangen sind — nicht die
	// der Zeilen. Wer drei gleiche besitzt, hat drei Exemplare.
	Exemplare int

	// AeltesterPreis ist der Zeitpunkt des ältesten Preises in dieser Summe.
	AeltesterPreis time.Time

	Quelle domain.Preisquelle
}

// Portfolio ist die Übersicht über die eigene Sammlung.
type Portfolio struct {
	// Zeilen sind die Summen je Währung — nach Währung sortiert.
	Zeilen []Portfoliozeile

	// OhnePreis ist die Zahl der Exemplare, für die **kein** Preis vorliegt.
	//
	// **Sie steht eigens da und nicht in einer Summe versteckt.** Eine Sammlung
	// von 300 Karten, von denen 280 keinen Preis haben, hat eine Summe, die
	// nichts aussagt — und ohne diese Zahl sähe sie aus wie eine über alle.
	OhnePreis int
}

// Portfolio rechnet die Sammlung zusammen.
//
// # Warum je Währung und nicht in einer
//
// Weil eine Umrechnung ein **Kurs** wäre, und den hat diese Anwendung nicht.
// Euro und Dollar zu addieren ergäbe eine Zahl ohne Einheit; sie
// umzurechnen hieße, einen Kurs zu erfinden oder eine zweite Quelle zu
// beziehen, deren Bedingungen niemand gelesen hat (§16, Satz 1).
//
// # Und warum der neueste Preis und nicht ein Mittel
//
// Ein Mittelwert über den Verlauf wäre eine andere Frage („was war es wert?").
// Hier steht die, die eine Übersicht beantwortet: **was ist es jetzt wert, so
// weit wir wissen.**
func (p *Preispflege) Portfolio(ctx context.Context) (Portfolio, error) {
	// **Die Sammlung holt sie sich selbst.**
	//
	// Der erste Anlauf ließ sie hereinreichen, und jede Fläche las sie vorher
	// — dreimal derselbe Aufruf, dreimal derselbe Fehlerweg. Einer davon war
	// **nicht zu betreten**: mit einem Bestand, der die Liste hergibt, gibt er
	// auch die Preise her. Ein Zweig ohne Fall gehört weg (§3), und ein
	// Anwendungsfall gehört in **eine** Schicht.
	exemplare, err := p.sammlung.Liste(ctx)
	if err != nil {
		return Portfolio{}, err
	}

	// **Ein Ruf und nicht einer je Exemplar** (#1292).
	//
	// Hier stand `p.ablage.Neueste(ctx, exemplar.Printing)` **in** der
	// Schleife. Gemessen am 2026-08-30 an 5.000 Exemplaren ohne einen einzigen
	// Preis: 89 ms gegen 5 ms für dieselben Zeilen. Der Aufwand wuchs mit dem
	// Bestand, und was er bezahlte, waren 5.000 leere Abfragen.
	//
	// Die Zusage wird als **Zahl der Abfragen** gehalten und nicht als Zeit:
	// `TestPortfolioFragtNichtJeExemplar` gilt auf jeder Maschine, eine
	// Millisekundenschwelle nur auf dieser (§10).
	alle, err := p.ablage.NeuesteAlle(ctx)
	if err != nil {
		return Portfolio{}, err
	}

	jeDruck := map[string][]domain.Preisstand{}
	for _, stand := range alle {
		jeDruck[stand.Printing] = append(jeDruck[stand.Printing], stand)
	}

	summen := map[string]*Portfoliozeile{}
	uebersicht := Portfolio{}

	for _, exemplar := range exemplare {
		// **Ein Wunsch ist kein Besitz** — und was man nicht hat, ist nichts
		// wert ([0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
		//
		// **Auch nicht in `OhnePreis`.** Wer sich zwanzig Karten wünscht,
		// hätte sonst zwanzig Exemplare „ohne Preis" in einer Übersicht über
		// seinen Besitz.
		if !exemplar.Besitzt() {
			continue
		}

		passende := jeWaehrungEiner(jeDruck[exemplar.Printing], exemplar)
		if len(passende) == 0 {
			uebersicht.OhnePreis += exemplar.Menge

			continue
		}

		for _, stand := range passende {
			eintragen(summen, stand, exemplar.Menge)
		}
	}

	uebersicht.Zeilen = geordnet(summen)

	return uebersicht, nil
}

// jeWaehrungEiner sucht **einen** Stand je Währung, der zum Exemplar passt.
//
// # Warum genau einer und nicht alle passenden
//
// Der erste Anlauf nahm jeden passenden — und ein Foil passte auf `foil` **und**
// `etched`. Hat ein Druck beide Preise, wurde er doppelt gezählt: Wert und
// Stückzahl. Das ist kein erfundener Fall; „Abundant Harvest" führt bei
// Scryfall beides. Angestrichen vom Zweitblick über Phase 3.4.
//
// **Ein Foil wird nicht mit dem Nonfoil-Preis gerechnet.** Der Unterschied ist
// oft das Zehnfache; wer ihn übergeht, rechnet eine Sammlung systematisch
// falsch.
//
// **`etched` ist der Rückfall eines Foils, nicht sein Gleichstand.** Scryfall
// führt es als eigene Ausführung; das Exemplar kennt nur ja oder nein. Wer
// beides führen will, braucht am Exemplar eine Ausführung statt eines
// Schalters — das ist eine Erweiterung und keine Berichtigung.
func jeWaehrungEiner(
	staende []domain.Preisstand, exemplar *domain.Exemplar,
) []domain.Preisstand {
	beste := map[string]domain.Preisstand{}

	for _, stand := range staende {
		if !zaehltFuer(stand, exemplar) {
			continue
		}

		vorher, schonDa := beste[stand.Waehrung]
		if !schonDa || rang(stand.Finish, exemplar.Ausfuehrung()) <
			rang(vorher.Finish, exemplar.Ausfuehrung()) {
			beste[stand.Waehrung] = stand
		}
	}

	gewaehlt := make([]domain.Preisstand, 0, len(beste))
	for _, stand := range beste {
		gewaehlt = append(gewaehlt, stand)
	}

	return gewaehlt
}

// ZaehltFuerPortfolio sagt, ob ein Stand für ein Exemplar zählt.
//
// **Ausgeführt für den Lauf über den Wegfall** (#1201): der Zweig ist von
// außen sonst nur über ein ganzes Portfolio zu erreichen, und dann prüfte der
// Lauf zehn Dinge auf einmal.
func ZaehltFuerPortfolio(stand domain.Preisstand, exemplar *domain.Exemplar) bool {
	return zaehltFuer(stand, exemplar)
}

// zaehltFuer sagt, ob ein Stand für dieses Exemplar überhaupt zählt.
//
// **`tix` zählt nicht.** Es ist der Preis einer Karte auf **Magic Online** —
// eine Eintrittskarte für einen digitalen Markt. Ein Exemplar ist ein Stück
// Papier ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md));
// wer den einen mit dem anderen aufrechnet, führt eine Sammlung, die es nicht
// gibt. Der Preis wird trotzdem **abgelegt und gezeigt** — nur nicht summiert.
func zaehltFuer(stand domain.Preisstand, exemplar *domain.Exemplar) bool {
	// **Nur ein genannter Stand zählt** (#1201, A3/#1442). Weggefallen sagt
	// „die Quelle nennt hier nichts mehr", ungeprüft „niemand hat nachgesehen";
	// in beiden Fällen ist der Wert null und bedeutet nichts. Ihn mitzurechnen
	// hieße, eine Karte für wertlos zu erklären, deren Preis niemand kennt.
	if stand.Lage != domain.PreisGenannt {
		return false
	}

	// **Und nur ein Verkaufspreis für Papier** (A3, #1442). Ein Ankaufspreis
	// ist, was ein Händler **zahlt** — er läge deutlich darunter, und eine
	// Sammlung, die beides mischte, hätte zwei Werte in einer Summe.
	if stand.Markt != domain.MarktPapier || stand.Art != domain.PreisartVerkauf {
		return false
	}

	if stand.Waehrung == waehrungTix {
		return false
	}

	// **Veredelt zu veredelt, gewöhnlich zu gewöhnlich.**
	//
	// **`foil` und `etched` sind zwei Veredelungen desselben Drucks**, und
	// nicht jeder Druck führt beide Preise. Wer hier strikt vergliche, ließe
	// eine Karte ohne Preis dastehen, obwohl einer vorliegt — welcher der
	// beiden genommen wird, entscheidet [rang].
	//
	// **Eine gewöhnliche Karte nimmt dagegen keinen veredelten Preis.** Der
	// Unterschied ist bei „Abundant Harvest" ein Vielfaches.
	return veredelt(stand.Finish) == veredelt(exemplar.Ausfuehrung())
}

// veredelt sagt, ob eine Ausführung eine Veredelung ist.
//
// **Seit `T2` sind es zwei** (#1467): `foil` und `etched`. Vorher gab es einen
// Wahrheitswert, und diese Frage stellte sich nicht.
func veredelt(finish domain.Finish) bool {
	return finish != domain.FinishNonfoil
}

// waehrungTix ist der digitale Markt, den eine Papiersammlung nicht summiert.
const waehrungTix = "TIX"

// rang sagt, wie gut ein Preis zu einem Exemplar passt — kleiner ist besser.
//
// **Die eigene Ausführung zuerst.** Ein Foil nimmt den `foil`-Preis und den
// `etched` nur, wenn es keinen `foil` gibt; für eine geätzte Karte gilt
// dasselbe umgekehrt.
//
// **Das Exemplar stand hier einmal nicht.** Solange es zwei Ausführungen gab,
// genügte „foil ist besser": für ein Nonfoil kam ohnehin nur `nonfoil` an. Mit
// der dritten trägt dieser Satz nicht mehr — er gäbe einer geätzten Karte den
// Foilpreis, obwohl ihr eigener danebenliegt.
func rang(preis, ausfuehrung domain.Finish) int {
	if preis == ausfuehrung {
		return 0
	}

	return 1
}

// eintragen addiert einen Stand mit seiner Stückzahl in die Summen.
func eintragen(summen map[string]*Portfoliozeile, stand domain.Preisstand, menge int) {
	zeile, schonDa := summen[stand.Waehrung]
	if !schonDa {
		zeile = &Portfoliozeile{
			Waehrung:       stand.Waehrung,
			AeltesterPreis: stand.Erhoben,
			Quelle:         stand.Quelle,
		}
		summen[stand.Waehrung] = zeile
	}

	zeile.Wert += stand.Wert * int64(menge)
	zeile.Exemplare += menge

	// **Der älteste gewinnt.** Eine Summe ist so frisch wie ihr ältester
	// Bestandteil; wer den jüngsten nennte, gäbe sie frischer aus, als sie ist.
	if stand.Erhoben.Before(zeile.AeltesterPreis) {
		zeile.AeltesterPreis = stand.Erhoben
	}
}

// geordnet macht aus den Summen eine Liste in fester Reihenfolge.
//
// **Nach Währung sortiert.** Eine Map in Go hat keine Reihenfolge, und eine
// Übersicht, die bei jedem Aufschlagen anders sortiert ist, lässt niemanden
// zweimal dasselbe lesen.
func geordnet(summen map[string]*Portfoliozeile) []Portfoliozeile {
	zeilen := make([]Portfoliozeile, 0, len(summen))
	for _, zeile := range summen {
		zeilen = append(zeilen, *zeile)
	}

	sort.Slice(zeilen, func(a, b int) bool {
		return zeilen[a].Waehrung < zeilen[b].Waehrung
	})

	return zeilen
}
