package service

// **Der zweite Schritt der Listeneinfuhr — schreiben, was entschieden ist**
// (#1292).
//
// **Eine eigene Datei**, weil `listeneinfuhr.go` sonst über die Zeilengrenze
// aus §2 ginge. Der Schnitt folgt der Frage: dort steht, *was es wäre*, hier,
// *was geschieht*.

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Einlesen schreibt die entschiedenen Zeilen in die Sammlung.
//
// # Alle oder so viele wie möglich?
//
// **So viele wie möglich, und jede Absage steht im Bericht** (§8). Eine
// Einfuhr von 300 Zeilen, die an der 217. abbräche, ließe 216 geschriebene
// Zeilen und keine Auskunft darüber zurück — der zweite Anlauf legte sie dann
// doppelt an.
//
// **Und der Bericht kommt auch mit einem Fehler heraus** (§7, Regel 3 und der
// Nachlauf über Phase 4.4): was geschrieben ist, ist geschrieben.
func (l *Listeneinfuhr) Einlesen(
	ctx context.Context, einzulesende []Listenzeile,
) (Einfuhrbericht, error) {
	return l.EinlesenMit(ctx, einzulesende, ZusammenAddieren)
}

// EinlesenMit schreibt die entschiedenen Zeilen — und führt Vorhandenes so
// zusammen, wie gewählt (DOM-03.AK-01, #1698).
func (l *Listeneinfuhr) EinlesenMit(
	ctx context.Context, einzulesende []Listenzeile, modus Zusammenfuehrung,
) (Einfuhrbericht, error) {
	if len(einzulesende) == 0 {
		// **Ein leerer Bericht ist leer, nicht `nil`.** Über die Naht käme
		// sonst ein `null` heraus, wo das Fenster eine Liste zählt — und seit
		// Befund 3 geht der Bericht **auch im Fehlerfall** hinaus, also
		// betrifft es jetzt auch diesen Zweig (§3, „zwischen den Prüfebenen").
		return leererBericht(), fmt.Errorf(
			"%w: ohne eine einzige Zeile gibt es nichts zu übernehmen",
			domain.ErrValidation)
	}

	geprueft, modusfehler := modus.geprueft()
	if modusfehler != nil {
		return leererBericht(), modusfehler
	}

	bericht := Einfuhrbericht{Angelegt: []int64{}, Befunde: []domain.Befund{}}

	for stelle, zeile := range einzulesende {
		fehler := l.einExemplar(ctx, zeile.alsExemplar(), &bericht, geprueft)
		if fehler == nil {
			continue
		}

		// **Ein Systemfehler bricht ab, ein Eingabefehler nicht** (§8).
		//
		// Der erste Anlauf machte aus **jedem** Fehler einen Befund je Zeile.
		// Bei einem klemmenden Bestand hieß das dreihundert gleichlautende
		// Befunde — und der Aufrufer bekam **kein** `error`, obwohl gar nichts
		// geschrieben war. Ein Plattenfehler ist keine Aussage über eine
		// Zeile.
		//
		// **Der Bericht geht trotzdem hinaus** (§7, Regel 3): was vor dem
		// Abbruch geschrieben wurde, ist geschrieben.
		if !errors.Is(fehler, domain.ErrValidation) &&
			!errors.Is(fehler, domain.ErrConflict) {
			return bericht, fehler
		}

		bericht.Befunde = append(bericht.Befunde, domain.Befund{
			Ort:     fmt.Sprintf("Zeile %d", stelle+1),
			Meldung: fehler.Error(),
			Art:     domain.ArtEingabe,
		})
	}

	return bericht, nil
}

// alsExemplar macht aus einer Listenzeile das Exemplar, das sie anlegte —
// für das Einlesen und für die Vorschau, damit es **eine** Abbildung ist.
func (zeile Listenzeile) alsExemplar() *domain.Exemplar {
	return &domain.Exemplar{
		Printing: zeile.Printing,
		Menge:    zeile.Menge,
		Zustand:  zeile.Zustand,
		Sprache:  zeile.Sprache,
		Finish:   zeile.Finish,
		Name:     zeile.Name,
		Setcode:  zeile.Setcode,
		Nummer:   zeile.Nummer,
	}
}

// einExemplar legt ein Exemplar an oder erhöht das vorhandene.
//
// # Warum es die Trennung gibt
//
// **Weil eine [Listenzeile] weniger trägt als ein Exemplar** (#1292). Eine
// eingefügte Kartenliste kennt keinen Erwerbspreis und keine Notiz — ein CSV
// kennt beide, und der Rundlauf über [Sammlungsausfuhr] wäre ohne diese Naht
// verlustbehaftet: der erste Anlauf gab den Preis mit 0 und die Notiz leer
// zurück, und beide Male sah es nach einem leeren Feld aus statt nach einem
// verlorenen.
func (l *Listeneinfuhr) einExemplar(
	ctx context.Context, exemplar *domain.Exemplar, bericht *Einfuhrbericht,
	modus Zusammenfuehrung,
) error {
	// **Über `vorbereiten` und nicht über `Validate` unmittelbar**: es glättet
	// zuerst und gibt den ersten Befund als Fehler. Warum nur den ersten,
	// steht dort.
	if fehler := vorbereiten(exemplar); fehler != nil {
		return fehler
	}

	// **Erst nachsehen, dann anlegen.** Der Bestand hat eine Eindeutigkeit
	// über Druck, Zustand, Sprache und Ausführung; wer stur anlegte, bekäme
	// bei der zweiten Liste einen Konflikt statt eines größeren Stapels.
	//
	// **Warum das kein Wettlauf ist, den jemand herstellt**: eine Einfuhr ist
	// ein Vorgang, und §7 Regel 4 sagt, dass ein zweiter nicht daneben läuft.
	// Käme trotzdem einer, gäbe `Anlegen` den Konflikt — das ist die
	// Absicherung darunter, und sie bleibt.
	vorhanden, fehler := l.vorhandene(ctx, exemplar)
	if fehler != nil {
		return fehler
	}

	for _, schon := range vorhanden {
		if !gleicheZeile(schon, exemplar) {
			continue
		}

		// **Auslassen trifft nur, was vor dieser Einfuhr da war** — eine
		// Zeile, die sie eben angelegt hat, wächst weiter.
		if modus == ZusammenAuslassen && !slices.Contains(bericht.Angelegt, schon.ID) {
			bericht.Ausgelassen++

			return nil
		}

		erhoeht := *schon
		erhoeht.Menge += exemplar.Menge

		if schreibfehler := l.sammlung.Aendern(ctx, &erhoeht); schreibfehler != nil {
			return schreibfehler
		}

		bericht.Erhoeht++
		bericht.Stueck += exemplar.Menge

		return nil
	}

	kennung, fehler := l.sammlung.Anlegen(ctx, exemplar)
	if fehler != nil {
		return fehler
	}

	bericht.Angelegt = append(bericht.Angelegt, kennung)
	bericht.Stueck += exemplar.Menge

	return nil
}

// gleicheZeile sagt, ob zwei Exemplare dieselbe Zeile des Bestands meinen.
//
// **Die fünf Spalten der Eindeutigkeit und nicht mehr.** Wer Notiz oder
// Erwerbspreis mitverglich, legte bei jeder Liste eine neue Zeile an — und die
// Eindeutigkeit im Bestand wiese sie ab.
//
// # Warum fünf und nicht vier
//
// **`wunsch` gehört dazu, seit Migration 0026 sie aufgenommen hat.** Hier
// standen vier, und das war ein Fehler mit Wirkung: ein CSV mit demselben
// Druck einmal als Besitz und einmal als Wunsch legte beim Zurücklesen die
// **zweite Zeile auf die erste** — der Wunsch verschwand und der Besitz wuchs.
// Dieselbe Verwechslung machte eine Listeneinfuhr, die auf einen vorhandenen
// **Wunsch** traf: sie erhöhte ihn, statt Besitz anzulegen.
//
// **Und der Druck ist der Schlüssel, nicht der Alias** (A2, #1441): zwei
// vorläufige Zeilen ohne Alias sind nur dann dieselbe, wenn sie dieselbe
// Druckzeile meinen.
func gleicheZeile(links, rechts *domain.Exemplar) bool {
	return links.Druckschluessel() == rechts.Druckschluessel() &&
		links.Zustand == rechts.Zustand &&
		links.Sprache == rechts.Sprache &&
		links.Ausfuehrung() == rechts.Ausfuehrung() &&
		links.Wunsch == rechts.Wunsch
}

// vorhandene gibt die Zeilen, die denselben Druck meinen — über den Alias,
// oder über die Druckzeile, wenn es keinen Alias gibt (A2, #1441).
func (l *Listeneinfuhr) vorhandene(
	ctx context.Context, exemplar *domain.Exemplar,
) ([]*domain.Exemplar, error) {
	if exemplar.Printing == "" {
		return l.sammlung.NachDruck(ctx, exemplar.DruckID)
	}

	return l.sammlung.Fuer(ctx, exemplar.Printing)
}

// leererBericht gibt einen Bericht, der nichts sagt — aber jede Liste hat.
//
// **Kein `Einfuhrbericht{}`**: sein `Angelegt` wäre `nil`, und über die Naht
// käme ein `null` heraus, wo das Fenster eine Liste zählt (§3).
func leererBericht() Einfuhrbericht {
	return Einfuhrbericht{Angelegt: []int64{}, Befunde: []domain.Befund{}}
}
