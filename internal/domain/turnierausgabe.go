package domain

// **Vom Turnier zur Datei und zurück** — DOM-04.AK-03
// (#1699),
// Vermerk [0140](../../docs/entscheidungen/0140-ein-turnier-als-datei.md).
//
// # Warum die Prüfung strenger ist als die Tabellen
//
// **Weil ein eingelesener Verlauf keinen Weg nimmt, den die Leitung kennt.**
// Wer am Fenster eine zweite Runde legt, hat die erste zu Ende gespielt; wer
// eine Datei schreibt, kann eine zweite ohne erste hinschreiben. Die Prüfung
// hier verlangt deshalb, was die Leitung beim Spielen verlangt: Runden ab 1
// ohne Lücke, jede außer der letzten ohne offenen Tisch, niemand zweimal in
// einer Runde, ein Abschlussdatum nur ohne offenen Tisch.

import (
	"strings"
	"time"
)

// HoechstensSpiele sagt, über wie viele Spiele ein Match hier geht.
//
// **Null heißt: keine Grenze.** Bei einem Einzelspielturnier ist es eines;
// sonst sagt dieses Programm nichts darüber, und eine erfundene Zahl wäre
// schlimmer als keine.
func (t Turnier) HoechstensSpiele() int {
	if t.Einzelspiel {
		return 1
	}

	return 0
}

// AusgabeVon macht aus einem Turnier des Bestands eines für die Datei — das
// Format steht mit seinem Namen da, weil seine Kennung nur hier gilt.
func AusgabeVon(turnier Turnier, format string) Turnierausgabe {
	stelle := make(map[int64]int, len(turnier.Teilnehmer))
	ausgabe := Turnierausgabe{
		Name: turnier.Name, Spiel: turnier.Spiel, Format: format,
		Einzelspiel: turnier.Einzelspiel, Notiz: turnier.Notiz, Abgeschlossen: turnier.Abgeschlossen,
		Teilnehmer: make([]Teilnehmerausgabe, 0, len(turnier.Teilnehmer)),
		Runden:     make([]Rundenausgabe, 0, len(turnier.Runden)),
	}

	for i, teilnehmer := range turnier.Teilnehmer {
		stelle[teilnehmer.ID] = i + 1

		eines := Teilnehmerausgabe{Name: teilnehmer.Name, Ausgestiegen: teilnehmer.Ausgestiegen}
		for _, platz := range teilnehmer.Deckabzug {
			eines.Deckabzug = append(eines.Deckabzug, Abzugsausgabe{
				Bezeichnung: platz.Bezeichnung, Printing: platz.Printing,
				Menge: platz.Menge, Bereich: platz.Bereich,
			})
		}

		ausgabe.Teilnehmer = append(ausgabe.Teilnehmer, eines)
	}

	for _, runde := range turnier.Runden {
		eine := Rundenausgabe{Nummer: runde.Nummer, Paarungen: make([]Paarungsausgabe, 0, len(runde.Paarungen))}

		for _, paarung := range runde.Paarungen {
			// **Ein Bye bleibt ohne B** — `stelle[0]` gibt es nicht und ist null.
			zeile := Paarungsausgabe{A: stelle[paarung.A], B: stelle[paarung.B]}
			if paarung.Ergebnis != nil {
				zeile.Ergebnis = &Ergebnisausgabe{
					SpieleA: paarung.Ergebnis.SpieleA, SpieleB: paarung.Ergebnis.SpieleB,
					Unentschieden: paarung.Ergebnis.Unentschieden,
				}
			}

			eine.Paarungen = append(eine.Paarungen, zeile)
		}

		ausgabe.Runden = append(ausgabe.Runden, eine)
	}

	return ausgabe
}

// AlsTurnier macht aus der Datei ein Turnier mit **vorläufigen** Kennungen:
// jede teilnehmende Person trägt ihre Stelle in der Datei, und die Paarungen
// verweisen darüber. Der Speicher vergibt die echten.
func (a Turnierausgabe) AlsTurnier() Turnier {
	turnier := Turnier{
		Name: a.Name, Spiel: a.Spiel, Einzelspiel: a.Einzelspiel, Notiz: a.Notiz,
		Abgeschlossen: a.Abgeschlossen,
	}

	for i, eines := range a.Teilnehmer {
		teilnehmer := Teilnehmer{ID: int64(i + 1), Name: eines.Name, Ausgestiegen: eines.Ausgestiegen}
		for _, platz := range eines.Deckabzug {
			teilnehmer.Deckabzug = append(teilnehmer.Deckabzug, Abzugsplatz{
				Bezeichnung: platz.Bezeichnung, Printing: platz.Printing,
				Menge: platz.Menge, Bereich: platz.Bereich,
			})
		}

		turnier.Teilnehmer = append(turnier.Teilnehmer, teilnehmer)
	}

	for _, runde := range a.Runden {
		eine := Runde{Nummer: runde.Nummer}

		for _, zeile := range runde.Paarungen {
			paarung := Paarung{A: int64(zeile.A), B: int64(zeile.B)}
			if zeile.Ergebnis != nil {
				paarung.Ergebnis = &Ergebnis{
					SpieleA: zeile.Ergebnis.SpieleA, SpieleB: zeile.Ergebnis.SpieleB,
					Unentschieden: zeile.Ergebnis.Unentschieden,
				}
			}

			eine.Paarungen = append(eine.Paarungen, paarung)
		}

		turnier.Runden = append(turnier.Runden, eine)
	}

	return turnier
}

// pruefe sagt, ob sich dieser Verlauf so gespielt haben kann.
func (a Turnierausgabe) pruefe() error {
	turnier := a.AlsTurnier()
	if fehler := turnier.Pruefe(); fehler != nil {
		return fehler
	}

	for i, eines := range a.Teilnehmer {
		if strings.TrimSpace(eines.Name) == "" {
			return NewFieldError("teilnehmer", SchluesselValidierung,
				"Die %d. teilnehmende Person hat keinen Namen.", i+1)
		}

		for j, platz := range eines.Deckabzug {
			// **Bezeichnung oder Druck** — die Leitung zieht einen echten
			// Druck mit Kennung und ohne Bezeichnung ab (`abziehen`), eine
			// eigene Karte mit Bezeichnung und ohne Kennung. Ohne beides
			// sagt die Zeile nicht, was darin lag.
			ohneAngabe := strings.TrimSpace(platz.Bezeichnung) == "" && strings.TrimSpace(platz.Printing) == ""
			if ohneAngabe || platz.Menge < 1 || !IstDeckbereich(platz.Bereich) {
				return NewFieldError("teilnehmer", SchluesselValidierung,
					"Die Liste von »%s« ist in Zeile %d unvollständig.", eines.Name, j+1)
			}
		}
	}

	for i, runde := range a.Runden {
		if fehler := a.rundePruefen(turnier, i, runde); fehler != nil {
			return fehler
		}
	}

	return a.abschlussPruefen(turnier)
}

// rundePruefen prüft eine Runde gegen ihre Stelle im Verlauf.
func (a Turnierausgabe) rundePruefen(turnier Turnier, i int, runde Rundenausgabe) error {
	if runde.Nummer != i+1 {
		return NewFieldError("runden", SchluesselValidierung,
			"An %d. Stelle steht Runde %d — die Runden zählen ab 1 ohne Lücke.", i+1, runde.Nummer)
	}

	// **Was die Leitung nie legt**: eine Runde ohne Paarung, und mehr als
	// einen Bye — der Paarungsplan gibt ihn nur dem einen, der bei ungerader
	// Zahl übrig bleibt (`paarungsplan.go`).
	byes := 0
	for _, zeile := range runde.Paarungen {
		if zeile.B == 0 {
			byes++
		}
	}

	if len(runde.Paarungen) == 0 || byes > 1 {
		return NewFieldError("runden", SchluesselValidierung,
			"Runde %d hat %d Paarungen und %d Byes — eine Runde hat mindestens eine Paarung und höchstens einen Bye.",
			runde.Nummer, len(runde.Paarungen), byes)
	}

	gesehen := map[int]bool{}
	letzte := i == len(a.Runden)-1

	for j, zeile := range runde.Paarungen {
		seiten := []int{zeile.A}
		if zeile.B != 0 {
			seiten = append(seiten, zeile.B)
		}

		for _, seite := range seiten {
			if seite < 1 || seite > len(a.Teilnehmer) || gesehen[seite] {
				return NewFieldError("runden", SchluesselValidierung,
					"Runde %d, Paarung %d: die %d. teilnehmende Person gibt es nicht, "+
						"oder sie spielt in dieser Runde zweimal.", runde.Nummer, j+1, seite)
			}

			gesehen[seite] = true
		}

		if fehler := ergebnisPruefen(turnier, runde.Nummer, j, zeile, letzte); fehler != nil {
			return fehler
		}
	}

	return nil
}

// ergebnisPruefen prüft den Stand einer Paarung.
func ergebnisPruefen(turnier Turnier, nummer, j int, zeile Paarungsausgabe, letzte bool) error {
	switch {
	case zeile.B == 0 && zeile.Ergebnis != nil:
		return NewFieldError("runden", SchluesselValidierung,
			"Runde %d, Paarung %d: ein Bye trägt kein Ergebnis — er ist eines.", nummer, j+1)

	case zeile.B != 0 && zeile.Ergebnis == nil && !letzte:
		return NewFieldError("runden", SchluesselValidierung,
			"Runde %d, Paarung %d: ohne Ergebnis, obwohl danach noch eine Runde folgt.", nummer, j+1)

	case zeile.Ergebnis != nil:
		stand := Ergebnis{
			SpieleA: zeile.Ergebnis.SpieleA, SpieleB: zeile.Ergebnis.SpieleB,
			Unentschieden: zeile.Ergebnis.Unentschieden,
		}
		// **Ein eigener Satz und nicht der von [Ergebnis.Pruefe]**: der trägt
		// den Vorsatz von [ErrValidation], und eine Meldung am Feld sagt,
		// **welcher** Stand es ist.
		if stand.Pruefe(turnier.HoechstensSpiele()) != nil {
			return NewFieldError("runden", SchluesselValidierung,
				"Runde %d, Paarung %d: %d–%d–%d ist hier kein möglicher Spielstand.",
				nummer, j+1, stand.SpieleA, stand.SpieleB, stand.Unentschieden)
		}
	}

	return nil
}

// abschlussPruefen hält, dass ein Abschlussdatum ein Datum ist und nur an
// einem Turnier ohne offenen Tisch steht (Vermerk 0139).
func (a Turnierausgabe) abschlussPruefen(turnier Turnier) error {
	if a.Abgeschlossen == "" {
		return nil
	}

	if _, fehler := time.Parse(time.RFC3339, a.Abgeschlossen); fehler != nil {
		return NewFieldError("turnier", SchluesselValidierung,
			"»%s« ist kein Abschlussdatum.", a.Abgeschlossen)
	}

	return turnier.Abschliessbar()
}
