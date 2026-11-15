package storage

// **Was aus einem reservierten Auftrag wird** — Werkstück O12 von Phase 5.2
// (#1439).
//
// **Eine eigene Datei**, weil `kostenjournal.go` mit dem Abschluss über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, wie ein
// Auftrag **entsteht** und was ihn trägt — hier, wie er **endet**.
//
// # Drei Enden, und eines davon endet nicht
//
//   - **abgerechnet**: der Anbieter hat geantwortet, der Betrag steht.
//   - **freigegeben**: es hat nichts gekostet — der Ruf scheiterte **vor**
//     jeder möglichen Annahme.
//   - **unklar**: die Frist lief danach ab. Der Posten bleibt gebunden, bis
//     ihn ein Mensch klärt; das ist kein Ende, sondern ein Vermerk.

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abschliessen trägt den wirklichen Betrag ein.
//
// **Er darf den reservierten nicht übersteigen.** Sonst wäre die Reservierung
// keine Grenze, sondern eine Vermutung — und die Abnahme verlangt „maximale
// Ausgabe/Batchkosten vor Freigabe".
func (j *Kostenjournal) Abschliessen(
	ctx context.Context, profilID int64, kennung string, betrag domain.Betrag, jetzt time.Time,
) (domain.Kostenposten, error) {
	if fehler := betrag.Pruefe(); fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	return j.abschluss(ctx, profilID, kennung, domain.KostenAbgerechnet, betrag, "", jetzt)
}

// Freigeben sagt, dass nichts angefallen ist.
func (j *Kostenjournal) Freigeben(
	ctx context.Context, profilID int64, kennung string, jetzt time.Time,
) (domain.Kostenposten, error) {
	return j.abschluss(ctx, profilID, kennung, domain.KostenFreigegeben, domain.Betrag{}, "", jetzt)
}

// AlsUnklarMerken hält die Reservierung und sagt, warum.
func (j *Kostenjournal) AlsUnklarMerken(
	ctx context.Context, profilID int64, kennung, grund string, jetzt time.Time,
) (domain.Kostenposten, error) {
	return j.abschluss(ctx, profilID, kennung, domain.KostenUnklar, domain.Betrag{}, grund, jetzt)
}

// abschluss ist der gemeinsame Weg der drei.
//
// # Ein Posten wird genau einmal abgeschlossen
//
// **Ein zweiter Versuch ist ein Konflikt**, und das ist die Zusage gegen die
// „doppelte Antwort" aus der Abnahme: wer zweimal abrechnet, bucht zweimal.
// Die Ausnahme ist der **unklare** Posten — er ist ausdrücklich noch offen,
// und ihn zu klären ist der belegwahrende Weg heraus.
//
// **Im Profil des Postens** (Nachlauf über Phase 6.1, Befund 4): die Kennung
// gibt der Aufrufer vor, und in zwei Profilen kann es dieselbe sein.
func (j *Kostenjournal) abschluss(
	ctx context.Context, profilID int64, kennung string, stand domain.Kostenstand,
	betrag domain.Betrag, grund string, jetzt time.Time,
) (domain.Kostenposten, error) {
	var vorhanden domain.Kostenposten

	fehler := imGeschaeft(ctx, j.db, "den Abschluss", func(geschaeft *sql.Tx) error {
		gelesen, gefunden, lesefehler := postenLesen(ctx, geschaeft, profilID, kennung)
		if lesefehler != nil {
			return lesefehler
		}

		if !gefunden {
			return fmt.Errorf("%w: den Auftrag %q gibt es nicht",
				domain.ErrNotFound, kennung)
		}

		vorhanden = gelesen

		if pruefehler := abschlusspruefung(vorhanden, stand, betrag); pruefehler != nil {
			return pruefehler
		}

		_, schreibfehler := geschaeft.ExecContext(ctx, `
			UPDATE kostenposten
			   SET stand = ?, betrag = ?, grund = ?, beendet = ?
			 WHERE profil_id = ? AND kennung = ?`,
			string(stand), betrag.Wert, grund, Zeitstempel(jetzt), profilID, kennung)

		return uebersetze(schreibfehler, "den Auftrag %q abschließen", kennung)
	})
	if fehler != nil {
		return vorhanden, fehler
	}

	vorhanden.Stand = stand
	vorhanden.Betrag = domain.Betrag{Wert: betrag.Wert, Waehrung: vorhanden.Hoechstbetrag.Waehrung}
	vorhanden.Grund = grund
	vorhanden.Beendet = jetzt

	return vorhanden, nil
}

// abschlusspruefung sagt, ob dieser Posten so abgeschlossen werden darf.
//
// **Ein Posten wird genau einmal abgeschlossen** — die Ausnahme ist der
// **unklare**: er ist ausdrücklich noch offen, und ihn zu klären ist der
// belegwahrende Weg heraus.
func abschlusspruefung(
	vorhanden domain.Kostenposten, stand domain.Kostenstand, betrag domain.Betrag,
) error {
	if !vorhanden.Stand.Offen() {
		return fmt.Errorf("%w: der Auftrag %q ist schon %s — ein zweiter "+
			"Abschluss bucht ein zweites Mal",
			domain.ErrConflict, vorhanden.Kennung, vorhanden.Stand)
	}

	if stand != domain.KostenAbgerechnet {
		return nil
	}

	ueber, vergleichsfehler := betrag.Uebersteigt(vorhanden.Hoechstbetrag)
	if vergleichsfehler != nil {
		return vergleichsfehler
	}

	if ueber {
		return fmt.Errorf("%w: %d %s übersteigt den reservierten Höchstbetrag von %d",
			domain.ErrValidation, betrag.Wert, betrag.Waehrung,
			vorhanden.Hoechstbetrag.Wert)
	}

	return nil
}

// Stand zählt zusammen, was an einem Konto steht.
//
// **Die vier Zahlen, die jede Fläche zeigt** — und `verfuegbar` ist gerechnet
// und nicht gespeichert: zwei Stellen, die dieselbe Subtraktion machen, laufen
// auseinander.
func (j *Kostenjournal) Stand(
	ctx context.Context, profilID int64, konto string,
	zeitraum domain.Zeitraum, jetzt time.Time,
) (domain.Haushaltsstand, error) {
	grenze, fehler := j.Budget(ctx, profilID, konto, zeitraum)
	if fehler != nil {
		return domain.Haushaltsstand{}, fehler
	}

	// **Die drei Summen beginnen in der Währung der Grenze.**
	//
	// Sonst stünden sie ohne eine da, solange nichts gebucht ist — und die
	// Seite zeigte „0" neben „50,00 €" in derselben Spalte. In der laufenden
	// Anwendung nachgesehen (2026-09-10).
	//
	// **Und es ist nicht nur eine Anzeigefrage**: eine Buchung in einer
	// anderen Währung als der Grenze fällt damit **sofort** auf, statt erst
	// beim Abziehen.
	leer := domain.Betrag{Waehrung: grenze.Waehrung}

	stand := domain.Haushaltsstand{
		Konto: konto, Zeitraum: zeitraum, Grenze: grenze,
		Reserviert: leer, Verbraucht: leer, Unklar: leer,
	}

	zeilen, lesefehler := j.db.QueryContext(ctx, `
		SELECT hoechstbetrag, betrag, waehrung, stand FROM kostenposten
		 WHERE profil_id = ? AND konto = ? AND begonnen >= ?`,
		profilID, konto, Zeitstempel(zeitraum.Fenster(jetzt)))
	if lesefehler != nil {
		return domain.Haushaltsstand{}, uebersetze(lesefehler,
			"die Posten von %q zählen", konto)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	for zeilen.Next() {
		var (
			posten domain.Kostenposten
			marke  string
		)

		if fehler := zeilen.Scan(&posten.Hoechstbetrag.Wert, &posten.Betrag.Wert,
			&posten.Hoechstbetrag.Waehrung, &marke); fehler != nil {
			return domain.Haushaltsstand{}, uebersetze(fehler, "einen Posten lesen")
		}

		posten.Betrag.Waehrung = posten.Hoechstbetrag.Waehrung
		posten.Stand = domain.Kostenstand(marke)

		if fehler := aufaddieren(&stand, posten); fehler != nil {
			return domain.Haushaltsstand{}, fehler
		}
	}

	if fehler := zeilen.Err(); fehler != nil {
		return domain.Haushaltsstand{}, uebersetze(fehler, "die Posten von %q zählen", konto)
	}

	return verfuegbarRechnen(stand), nil
}

// aufaddieren legt einen Posten auf den passenden Topf.
func aufaddieren(stand *domain.Haushaltsstand, posten domain.Kostenposten) error {
	var (
		ziel   *domain.Betrag
		fehler error
	)

	switch posten.Stand {
	case domain.KostenReserviert:
		ziel = &stand.Reserviert
	case domain.KostenAbgerechnet:
		ziel = &stand.Verbraucht
	case domain.KostenUnklar:
		ziel = &stand.Unklar
	case domain.KostenFreigegeben:
		return nil
	default:
		// **Ein unbekannter Stand ist ein Befund und kein stiller Nullwert**
		// (§8): er käme aus einer neueren Fassung, und ihn wegzulassen machte
		// aus gebundenem Geld verfügbares.
		return fmt.Errorf("%w: der Posten %q trägt den unbekannten Stand %q",
			domain.ErrValidation, posten.Kennung, posten.Stand)
	}

	*ziel, fehler = ziel.Plus(posten.Gebunden())

	return fehler
}

// verfuegbarRechnen zieht ab, was gebunden ist.
//
// # Ohne Fehlerweg, und das ist eine Zusage über die Zeilen darüber
//
// **Alle vier Beträge stehen in der Währung der Grenze**: die drei Summen
// beginnen so ([Kostenjournal.Stand]), und ein Posten in einer anderen fällt
// schon beim Aufaddieren auf. Ein `Plus` mit Fehlerrückgabe hätte hier drei
// Zweige, die kein Aufruf betritt (§3).
//
// **Unklar zählt wie verbraucht** — das ist die konservative Seite, und sie
// steht so in der Abnahme.
func verfuegbarRechnen(stand domain.Haushaltsstand) domain.Haushaltsstand {
	gebunden := stand.Reserviert.Wert + stand.Verbraucht.Wert + stand.Unklar.Wert

	stand.Verfuegbar = domain.Betrag{Waehrung: stand.Grenze.Waehrung}

	// **Nie unter null.** Wer seine Grenze unter das senkt, was schon gebunden
	// ist, hat null verfügbar — und keine negative Zahl, die niemand liest.
	if stand.Grenze.Wert > gebunden {
		stand.Verfuegbar.Wert = stand.Grenze.Wert - gebunden
	}

	return stand
}

// Posten gibt die Aufträge eines Kontos im laufenden Zeitraum — jüngste zuerst.
func (j *Kostenjournal) Posten(
	ctx context.Context, profilID int64, konto string,
	zeitraum domain.Zeitraum, jetzt time.Time, hoechstens int,
) ([]domain.Kostenposten, error) {
	zeilen, fehler := j.db.QueryContext(ctx, `
		SELECT kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
		       waehrung, stand, grund, begonnen, beendet
		  FROM kostenposten
		 WHERE profil_id = ? AND konto = ? AND begonnen >= ?
		 ORDER BY begonnen DESC, kennung DESC
		 LIMIT ?`,
		profilID, konto, Zeitstempel(zeitraum.Fenster(jetzt)), hoechstens)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Posten von %q lesen", konto)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	posten := []domain.Kostenposten{}

	for zeilen.Next() {
		einer, lesefehler := postenAusZeile(zeilen)
		if lesefehler != nil {
			return nil, lesefehler
		}

		posten = append(posten, einer)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Posten von %q lesen", konto)
	}

	return posten, nil
}

// postenAusZeile liest einen Posten aus einer Ergebniszeile.
func postenAusZeile(zeilen *sql.Rows) (domain.Kostenposten, error) {
	var (
		posten            domain.Kostenposten
		marke             string
		begonnen, beendet string
	)

	if fehler := zeilen.Scan(&posten.Kennung, &posten.ProfilID, &posten.Konto,
		&posten.Aufgabe, &posten.Hoechstbetrag.Wert, &posten.Betrag.Wert,
		&posten.Hoechstbetrag.Waehrung, &marke, &posten.Grund,
		&begonnen, &beendet); fehler != nil {
		return domain.Kostenposten{}, uebersetze(fehler, "einen Posten lesen")
	}

	posten.Betrag.Waehrung = posten.Hoechstbetrag.Waehrung
	posten.Stand = domain.Kostenstand(marke)
	posten.Begonnen = zeitpunktAus(begonnen)
	posten.Beendet = zeitpunktAus(beendet)

	return posten, nil
}
