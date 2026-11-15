package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kartenschreiber ändert einzelne Feldwerte.
//
// Getrennt vom `SetSpeicher`, der ganze Sets anlegt: das eine geschieht beim
// Öffnen einmal, das andere bei jedem Tastendruck. Sie in eine Struktur zu
// legen hieße, die teure Transaktion des einen für das andere mitzuschleppen.
type Kartenschreiber struct {
	db       *sql.DB
	urheber  domain.Urheber
	herkunft string

	// gruppe fasst die Einträge eines Aufrufs zu **einem** Schritt im Verlauf
	// zusammen ([Kartenschreiber.inEinemSchritt]); leer heißt: je Eintrag
	// einer.
	gruppe string
}

// NeuerKartenschreiber baut den Schreiber über einer Datenbank.
//
// **Voreingestellt ist das Fenster**, weil es der Weg ist, den jeder
// Tastendruck nimmt. Wer ihn eigens ansagen müsste, vergäße es einmal — und
// dann stünde eine Tipparbeit unter fremdem Namen.
func NeuerKartenschreiber(db *sql.DB) *Kartenschreiber {
	return &Kartenschreiber{db: db, urheber: domain.UrheberFenster}
}

// Als gibt einen Schreiber, der unter anderem Namen schreibt.
//
// **Einen zweiten, keinen umgestellten.** Ein Aufruf, der den Empfänger
// änderte, stellte jeden anderen Halter mit um — und im Composition Root hält
// `app` denselben Schreiber zweimal, einmal als `schreiber` und einmal als
// `massenschreiber`.
func (k *Kartenschreiber) Als(wer domain.Urheber) *Kartenschreiber {
	zweiter := *k
	zweiter.urheber = wer

	return &zweiter
}

// Von gibt einen Schreiber, dessen Änderungen ihre Herkunft tragen.
//
// **Dieselbe Gestalt wie [Kartenschreiber.Als] und aus demselben Grund**: ein
// zweiter Schreiber, kein umgestellter.
//
// **`woher` ist eine Adresse**, keine Bezeichnung — die Kartenseite bei
// Scryfall, die Datei, das Paket. Ein Name allein ließe später niemanden
// nachsehen.
//
// **Leer setzt zurück.** Wer `Von("")` ruft, bekommt einen Schreiber ohne
// Herkunft; damit lässt sich eine Kette wieder öffnen, ohne den ursprünglichen
// Schreiber zu kennen.
func (k *Kartenschreiber) Von(woher string) *Kartenschreiber {
	zweiter := *k
	zweiter.herkunft = woher

	return &zweiter
}

// inEinemSchritt gibt einen Schreiber, dessen Einträge **ein** Schritt im
// Verlauf sind — eine neue Gruppe je Aufruf (#2396).
//
// **Ein Aufruf, ein Schritt.** Seit der Schreibweg nachzieht, wer ein Feld
// liest, schreibt schon eine einzige Bearbeitung mehrere Felder; MSE nimmt sie
// als **eine** Aktion zurück (`SetScriptManager::onAction`,
// `src/script/script_manager.cpp`, Zeilen 141–147, Fassung f9d9356 vom
// 2021-07-06). Je Feld ein Schritt hieße ⌘Z je nachgezogenem Feld, und
// dazwischen stünde ein halber Stand.
func (k *Kartenschreiber) inEinemSchritt() *Kartenschreiber {
	zweiter := *k
	zweiter.gruppe = neueGruppe()

	return &zweiter
}

// karteAnlegen hängt eine Karte hinten an.
//
// **Die Reihenfolge kommt aus der Datenbank.** Sie trägt `UNIQUE (set_id,
// reihenfolge)`; eine im Programm gezählte Zahl gäbe zwei Fenstern auf
// demselben Bestand dieselbe, und das zweite Einfügen scheiterte an einem
// Schlüssel, obwohl nichts falsch war.
//
// **Und der Fremdschlüssel weist ein fremdes Set ab** — vorausgesetzt, die
// Verbindung hat `foreign_keys` an; `datenbank.go` schaltet es ein.
//
// **`RETURNING` und nicht `LastInsertId`.** Der Rückgabewert von
// `LastInsertId` trägt einen Fehler, den der Treiber nur meldet, wenn er die
// Abfrage gar nicht unterstützt — ein Zweig, den kein Aufruf betreten kann
// (§3). `RETURNING` gibt die Kennung auf demselben Weg wie jeden anderen
// Fehler.
//
// **Die Stelle kommt mit zurück**, statt hinterher noch einmal gefragt zu
// werden: der Verlaufsschritt braucht sie, und zwei Wege zu derselben Zahl sind
// zwei Zeitfenster, in denen sie auseinanderlaufen kann.
const karteAnlegen = `
INSERT INTO cards (set_id, reihenfolge)
VALUES (?, (SELECT coalesce(max(reihenfolge) + 1, 0) FROM cards WHERE set_id = ?))
RETURNING id, reihenfolge`

// KarteAnlegen legt eine leere Karte am Ende eines Sets an.
//
// # Ohne Feldwerte
//
// Was eine Karte trägt, entscheidet das Spielpaket über seine Vorgabewerte
// (`mitVorgaben` in `internal/service`). Der Speicher erfindet keine: eine
// Karte mit erfundenen Feldern sähe aus wie eine gefüllte und wäre es nicht
// (#562).
//
// # Ohne Verlaufseintrag
//
// Der Verlauf trägt heute nur Feldänderungen (`verlauf.go`, [feldweg]). Ein
// Schritt „Karte angelegt" bräuchte eine eigene Art — und die Rücknahme müsste
// die Karte wieder entfernen. Das steht in #562 und ist hier **nicht** getan:
// eine neue leere Karte ist kein Verlust, eine gelöschte wäre einer.
func (k *Kartenschreiber) KarteAnlegen(ctx context.Context, setID int64) (int64, error) {
	// **Ohne Felder, aber mit Verlaufsschritt** (#626) — derselbe Weg wie mit
	// gerechneten Werten (#2516).
	return k.KarteAnlegenMitWerten(ctx, setID, nil)
}

// feldSetzen legt einen Feldwert an oder überschreibt ihn.
//
// **`ON CONFLICT` statt „erst lesen, dann schreiben".** Zwischen einem Lesen
// und einem Schreiben liegt Zeit, und in ihr kann ein anderer Vorgang dasselbe
// Feld anlegen; dann scheitert das Einfügen an einem Primärschlüssel, obwohl
// nichts falsch war. Die Datenbank kann beides in einem Schritt.
//
// Die Reihenfolge bleibt, was sie war — sie zählt für den Rundlauf. Ein neuer
// Wert kommt hinten an.
const feldSetzen = `
INSERT INTO card_values (card_id, feld, wert, reihenfolge)
VALUES (?, ?, ?, (SELECT coalesce(max(reihenfolge) + 1, 0)
                    FROM card_values WHERE card_id = ?))
ON CONFLICT (card_id, feld) DO UPDATE SET wert = excluded.wert`

// feldSetzenAnStelle schreibt einen Feldwert **an eine gegebene Stelle**.
//
// Der Weg zurück aus dem Verlauf (#160): ein Feld, das entfernt und dessen
// Entfernung zurückgenommen wurde, gehört dorthin, wo es stand — nicht ans
// Ende. `feldSetzen` kann das nicht, weil es die Stelle selbst vergibt.
//
// **Die Reihenfolge wird auch beim Konflikt gesetzt.** Steht das Feld schon da,
// ist das ein Wiederherstellen über einen vorhandenen Wert; seine Stelle ist
// dann die aus dem Verlauf und nicht die, die zufällig dasteht.
const feldSetzenAnStelle = `
INSERT INTO card_values (card_id, feld, wert, reihenfolge)
VALUES (?, ?, ?, ?)
ON CONFLICT (card_id, feld) DO UPDATE
   SET wert = excluded.wert, reihenfolge = excluded.reihenfolge`

// FeldSetzen schreibt einen Feldwert.
//
// Eine Karte, die es nicht gibt, ist ein **`ErrNotFound`** und kein Konflikt:
// wer sie ändern will, hat sie nicht gefunden, nicht überholt. Der Unterschied
// entscheidet, ob das Frontend nachlädt oder eine Meldung zeigt (§8).
func (k *Kartenschreiber) FeldSetzen(ctx context.Context, karteID int64, feld, wert string) error {
	// **Die Karte gehört dem gezeigten Profil — geprüft in derselben
	// Transaktion wie das Schreiben** (#1815).
	//
	// Hier stand „Geprüft wird gar nicht — der Fremdschlüssel entscheidet".
	// Für eine Karte, die es nicht gibt, stimmte das; für die Karte eines
	// anderen Profils nicht: sie gibt es, und der Fremdschlüssel ließ das
	// Schreiben durch. Eine Karte, die es nicht gibt, fällt weiter an
	// derselben Prüfung durch.
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "Feld %q schreiben", feld)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
		return fehler
	}

	if _, fehler := tx.ExecContext(ctx, feldSetzen, karteID, feld, wert, karteID); fehler != nil {
		return uebersetze(fehler, "Feld %q schreiben", feld)
	}

	return uebersetze(tx.Commit(), "Feld %q schreiben", feld)
}

// KarteLesen gibt eine einzelne Karte samt ihren Werten.
//
// Für die Vorschau: sie braucht den Stand **nach** dem Schreiben, und ihn aus
// der Antwort des Schreibers zusammenzusetzen hieße, ihn an zwei Stellen zu
// führen.
func (l *SetLeser) KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error) {
	// **Nur eine Karte des gezeigten Profils, und in derselben Abfrage**
	// (#1815): eine fremde fehlt wie eine, die es nicht gibt.
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := l.db.QueryContext(ctx, `
SELECT cards.id, cards.set_id, cards.reihenfolge, card_values.feld, card_values.wert
  FROM cards
  JOIN sets ON sets.id = cards.set_id
  LEFT JOIN card_values ON card_values.card_id = cards.id
 WHERE cards.id = ? AND sets.profil_id = ?
 ORDER BY card_values.reihenfolge`, karteID, profilID)
	if err != nil {
		return nil, uebersetze(err, "Karte lesen")
	}

	defer func() { _ = zeilen.Close() }()

	var karte *domain.CardRow

	for zeilen.Next() {
		var (
			id     int64
			setID  int64
			nummer int
			feld   sql.NullString
			wert   sql.NullString
		)

		if err := zeilen.Scan(&id, &setID, &nummer, &feld, &wert); err != nil {
			return nil, uebersetze(err, "Karte lesen")
		}

		if karte == nil {
			karte = &domain.CardRow{ID: id, SetID: setID, Number: nummer}
		}

		if feld.Valid {
			karte.Values = append(karte.Values, domain.Value{Field: feld.String, Text: wert.String})
		}
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "Karte lesen")
	}

	if karte == nil {
		return nil, fmt.Errorf("%w: die Karte %d", domain.ErrNotFound, karteID)
	}

	return karte, nil
}

// FeldSetzenVonAussen setzt ein Feld und hält fest, woher der Wert kam.
//
// # Warum das eine eigene Methode ist und nicht ein zweiter Schreiber
//
// Weil die Herkunft zur **Änderung** gehört und nicht zu dem, der schreibt:
// dasselbe Fenster schreibt getippte Werte und übernommene, im Wechsel. Ein
// Schreiber „von Scryfall" müsste umgeschaltet werden, und wer das vergisst,
// hängt eine falsche Adresse an eine Tipparbeit.
//
// **[Kartenschreiber.Von] bleibt trotzdem**: es ist das Mittel, mit dem diese
// Methode arbeitet, und es ist prüfbar für sich.
func (k *Kartenschreiber) FeldSetzenVonAussen(
	ctx context.Context, karteID int64, feld, wert, herkunft string,
) (vorher string, hatte bool, fehler error) {
	return k.Von(herkunft).FeldSetzenMitVorwert(ctx, karteID, feld, wert)
}
