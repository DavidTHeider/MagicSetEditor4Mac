package storage

// **Der Abschluss in einem Zug** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum alles in **einer** Transaktion steht
//
// **Weil jeder Zwischenstand falsch ist.** Ein abgezogener Abgang ohne Zugang
// ist ein Verlust, ein Zugang ohne Abgang ein Geschenk, ein Journal ohne
// beides eine Behauptung, und eine gelöste Zusage ohne Buchung gibt Karten
// wieder frei, die über den Tisch gegangen sind. Das Issue verlangt: entweder
// der ganze Vorzustand oder der ganze Nachzustand.
//
// # Warum die Unterschrift **vorher** gerechnet wird
//
// **Weil in einer offenen Transaktion nichts warten darf.** Eine Unterschrift
// kommt aus dem Schlüsselbund, und der fragt zurück; ein Beleg geht ins Netz,
// und das antwortet nicht. Beides hielte eine SQLite-Transaktion offen, und
// eine offene Transaktion hält die Datenbank. Deshalb nimmt [Handelsabschluss]
// die fertige Unterschrift entgegen und rechnet sie nicht.
//
// # Und warum der Vorgang die Idempotenz trägt, nicht die Nachricht
//
// **Weil eine neue Nachrichtenkennung denselben Abschluss meint.** Wer ihn
// über die Kennung sicherte, buchte beim zweiten Anlauf mit neuer Kennung ein
// zweites Mal. Der Schlüssel ist der **Vorgang**: er steht im Beleg, der Beleg
// hat ihn als Primärschlüssel, und ein zweiter Abschluss scheitert daran.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abschlussablage verbucht Handel, führt das Journal und legt Belege ab.
type Abschlussablage struct{ db *sql.DB }

// NeueAbschlussablage baut sie über einer Datenbank.
func NeueAbschlussablage(db *sql.DB) *Abschlussablage {
	return &Abschlussablage{db: db}
}

// Abschliessen verbucht einen Handel — ganz oder gar nicht.
//
// **Die Reihenfolge ist Absicht.** Zuerst der Bestand, denn dort liegen die
// Prüfungen, die scheitern dürfen; dann das Journal, das nachzeichnet, was
// geschehen ist; dann die Zusagen, die eingelöst sind; zuletzt Revision,
// Beleg, Vermerk und Ausgang. Wer die billigen Schreibvorgänge nach vorn zöge,
// rollte sie öfter zurück.
func (a *Abschlussablage) Abschliessen(
	ctx context.Context, abschluss domain.Handelsabschluss,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if beanstandet := abschluss.Pruefe(); beanstandet != nil {
		return beanstandet
	}

	zug, fehler := a.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "einen Abschluss beginnen")
	}

	defer func() { _ = zug.Rollback() }() //nolint:errcheck // nach dem Commit ohne Wirkung

	if fehler := nochNichtVerbucht(ctx, zug, profilID, abschluss.Beleg.Vorgang); fehler != nil {
		return fehler
	}

	if fehler := bestandBewegen(ctx, zug, profilID, abschluss.Beleg); fehler != nil {
		return fehler
	}

	if fehler := journalSchreiben(
		ctx, zug, profilID, abschluss.Beleg.Zeilen); fehler != nil {
		return fehler
	}

	if fehler := zusagenEinloesen(
		ctx, zug, profilID, abschluss.Beleg.Vorgang); fehler != nil {
		return fehler
	}

	// **Gegen den Stand, von dem aus verbucht wird** (Befund Z1 des
	// Zweitblicks über Phase 6.8): zwei Verbuchungen, die beide
	// `uebergabeOffen` gelesen hatten, buchten sonst beide — der Bestand
	// bewegt sich oben im selben Zug und rollt mit zurück.
	if fehler := revisionFortschreiben(
		ctx, zug, profilID, abschluss.Vorher, abschluss.Angebot); fehler != nil {
		return fehler
	}

	if fehler := nachschreiben(ctx, zug, profilID, abschluss); fehler != nil {
		return fehler
	}

	return uebersetze(zug.Commit(), "einen Abschluss abschließen")
}

// nochNichtVerbucht weist einen Vorgang ab, der schon einen Beleg hat.
//
// **Vor dem Bestand und nicht erst beim Ablegen des Belegs** (C05, #2230):
// der Schlüssel von `handelsbelege` hätte einen zweiten Abschluss auch
// abgewiesen — aber erst, nachdem `bestandBewegen` gelaufen war. Hatte der
// erste eine Besitzzeile ganz abgegeben, fand der zweite sie nicht mehr und
// meldete „nicht gefunden" statt eines Konflikts; zwei gleichzeitige
// Verbuchungen über Fenster und Kommandozeile bekamen so je nach Zuschnitt
// eine andere Klasse. Gefunden von erzeugten Abschlüssen.
func nochNichtVerbucht(ctx context.Context, an abfrager, profilID int64, vorgang string) error {
	var da int

	fehler := an.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM handelsbelege WHERE profil_id = ? AND vorgang = ?`,
		profilID, vorgang).Scan(&da)
	if fehler != nil {
		return uebersetze(fehler, "den Beleg eines Vorgangs suchen")
	}

	if da > 0 {
		return domain.MitGrund(domain.ErrConflict, nil, "dieser Handel ist schon verbucht")
	}

	return nil
}

// nachschreiben legt Beleg, Vermerk und Ausgang ab.
//
// **Ein eigenes Verfahren, damit [Abschlussablage.Abschliessen] lesbar
// bleibt** — es sind drei Ablagen, die alle dasselbe tun: festhalten, dass es
// geschehen ist.
func nachschreiben(
	ctx context.Context, an abfrager, profilID int64,
	h domain.Handelsabschluss,
) error {
	if _, fehler := an.ExecContext(ctx,
		`INSERT INTO handelsbelege
		   (profil_id, vorgang, abdruck, form, unterschrift, beidseitig)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		profilID, h.Beleg.Vorgang, h.Beleg.Abdruck,
		string(h.Beleg.Unterschriftsform()), h.Unterschrift,
		alsZahl(h.Beleg.Beidseitig),
	); fehler != nil {
		return uebersetze(fehler, "einen Beleg ablegen")
	}

	if h.Vermerk != nil {
		if _, fehler := an.ExecContext(ctx,
			`INSERT INTO handelsbriefkasten
			   (profil_id, schluessel, abdruck, ergebnis, vorgang)
			 VALUES (?, ?, ?, ?, ?)`,
			profilID, h.Vermerk.Schluessel, h.Vermerk.Nutzlastabdruck,
			h.Vermerk.Ergebnis, h.Beleg.Vorgang,
		); fehler != nil {
			return uebersetze(fehler, "einen Abschluss vermerken")
		}
	}

	if h.Ausgang != nil {
		if _, fehler := an.ExecContext(ctx,
			`INSERT INTO handelsausgang
			   (profil_id, nachricht_id, vorgang, empfaenger, nachricht)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (profil_id, nachricht_id) DO NOTHING`,
			profilID, h.Ausgang.NachrichtID, h.Beleg.Vorgang,
			h.Ausgang.Empfaenger, h.Ausgang.Nachricht,
		); fehler != nil {
			return uebersetze(fehler, "einen Beleg einreihen")
		}
	}

	return nil
}

// zusagenEinloesen nimmt die Zusagen eines abgeschlossenen Vorgangs weg.
//
// # Warum das hier geht und in `Freigeben` nicht
//
// **Weil es zwei verschiedene Dinge sind.** [Reservierungsablage.Freigeben]
// weist eine bestätigte Übergabe ab, und zwar mit Absicht: von außen soll
// niemand Karten wieder verfügbar machen, die über den Tisch gegangen sind.
// Hier sind sie **eingelöst** — in derselben Transaktion, in der der Bestand
// sie abgibt. Nicht die Zusage wird aufgehoben, sondern ihr Zweck erfüllt.
func zusagenEinloesen(
	ctx context.Context, an abfrager, profilID int64, vorgang string,
) error {
	if _, fehler := an.ExecContext(ctx,
		`DELETE FROM reservierungen WHERE profil_id = ? AND vorgang = ?`,
		profilID, vorgang); fehler != nil {
		return uebersetze(fehler, "die Zusagen einlösen")
	}

	return nil
}

// bestandBewegen führt jede Zeile des Belegs im Bestand aus.
func bestandBewegen(
	ctx context.Context, an abfrager, profilID int64, beleg domain.Handelsbeleg,
) error {
	for _, zeile := range beleg.Zeilen {
		var fehler error

		if zeile.Art == domain.JournalAbgang {
			fehler = bestandAbgang(ctx, an, profilID, beleg.Vorgang, zeile)
		} else {
			fehler = bestandZugang(ctx, an, profilID, zeile)
		}

		if fehler != nil {
			return fehler
		}
	}

	return nil
}

// gedeckt sagt, wie viel eine bestätigte Übergabe für diesen Vorgang hält.
//
// **Und ob sie noch gegen den heutigen Stand rechnet.** Eine Zusage trägt die
// Fassung, gegen die sie festgehalten wurde (§7, Regel 2); hat jemand die
// Zeile seither geändert, rechnet sie gegen einen Stand, den es nicht mehr
// gibt — und dann wird nicht gebucht, sondern gefragt.
func gedeckt(
	ctx context.Context, an abfrager, profilID, exemplarID int64,
	vorgang string, fassung int64,
) (int, error) {
	zeilen, fehler := an.QueryContext(ctx,
		`SELECT menge, fassung FROM reservierungen
		  WHERE profil_id = ? AND exemplar_id = ? AND vorgang = ?
		    AND uebergeben = 1`, profilID, exemplarID, vorgang)
	if fehler != nil {
		return 0, uebersetze(fehler, "die bestätigten Zusagen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt aus Err()

	summe := 0

	for zeilen.Next() {
		var menge int

		var stand int64

		if fehler := zeilen.Scan(&menge, &stand); fehler != nil {
			return 0, uebersetze(fehler, "eine bestätigte Zusage lesen")
		}

		if stand != fassung {
			return 0, domain.MitGrund(domain.ErrConflict, nil,
				"der Bestand hat sich geändert, seit zugesagt wurde — "+
					"das klärt ein Mensch")
		}

		summe += menge
	}

	return summe, uebersetze(zeilen.Err(), "die bestätigten Zusagen lesen")
}
