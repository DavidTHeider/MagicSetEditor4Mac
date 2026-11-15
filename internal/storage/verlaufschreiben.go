package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// GrenzeVerlauf ist die Zahl der **Schritte**, die ein Set im Verlauf behält.
//
// **Zweihundert, und die Zahl ist benannt, weil #52 es verlangt.** Ein Verlauf,
// der mit jeder Sitzung wächst, macht aus einem Arbeitsformat ein Protokoll:
// bei dreihundert Karten und einem Nachmittag Tipparbeit sind das schnell
// zehntausend Zeilen, die niemand je zurücknimmt.
//
// **Schritte und nicht Zeilen**, denn eine Massenänderung ist ein Schritt über
// vielen Zeilen. Wer Zeilen zählte, zerschnitte sie: die Hälfte der Karten käme
// zurück, die andere bliebe stehen — ein Stand, den es nie gab. Und eine
// einzige Massenänderung über dreihundert Karten löschte den ganzen übrigen
// Verlauf.
const GrenzeVerlauf = 200

// neueGruppe gibt eine Kennung, die eine Massenänderung zusammenhält.
//
// **Aus `crypto/rand` und nicht aus der Uhr.** Zwei Massenänderungen in
// derselben Nanosekunde sind unwahrscheinlich, aber „unwahrscheinlich" ist
// keine Zusage — und zwei Schritte, die dieselbe Kennung tragen, nähme eine
// Rücknahme als einen.
func neueGruppe() string {
	roh := make([]byte, 8)

	// `rand.Read` aus `crypto/rand` füllt vollständig oder gerät in Panik; ein
	// Fehlerpfad wäre einer, den kein Test beträte.
	_, _ = rand.Read(roh) //nolint:errcheck // siehe darüber

	return hex.EncodeToString(roh)
}

// schreibeVerlauf legt einen Eintrag an — **in der Transaktion des Aufrufers**.
//
// Ein Verlaufseintrag, der nach der Änderung entsteht, fehlt genau dann, wenn
// jemand dazwischen abstürzt: dann ist die Änderung da und die Rücknahme fort.
//
// `gruppe` ist leer, wenn der Eintrag für sich steht.
func schreibeVerlauf(
	ctx context.Context, tx *sql.Tx, karteID int64, feld string,
	vorher, nachher feldweg, wer domain.Urheber, gruppe, herkunft string,
) error {
	hin, _ := json.Marshal(nachher)    //nolint:errcheck // ein Verbund aus Text und Wahrheit
	zurueck, _ := json.Marshal(vorher) //nolint:errcheck // dito

	_, err := tx.ExecContext(ctx,
		`INSERT INTO history
		        (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg,
		         urheber, gruppe, herkunft)
		 SELECT set_id, ?, ?, 'feld', ?, ?, ?, ?, ? FROM cards WHERE id = ?`,
		karteID, time.Now().UTC().Format(time.RFC3339Nano),
		string(hin), string(zurueck), string(wer), gruppe, herkunft, karteID)
	if err != nil {
		return uebersetze(err, "den Verlauf zu Feld %q schreiben", feld)
	}

	return kuerzeVerlauf(ctx, tx, karteID)
}

// schrittkennung ist der Ausdruck, der eine Zeile ihrem Schritt zuordnet.
//
// Eine Zeile mit Gruppe gehört zu ihrer Gruppe; eine ohne steht für sich und
// bekommt einen Schritt aus ihrer eigenen Kennung. Das Präfix trennt die beiden
// Räume — eine Gruppenkennung ist Hex und träfe eine Zahl nie, aber sich darauf
// zu verlassen hieße, eine Eigenschaft der Kennung zur Regel zu machen.
const schrittkennung = `coalesce(nullif(gruppe, ''), 'einzeln:' || id)`

// kuerzeVerlauf hält die Grenze — **beim Schreiben, nicht beim Lesen**.
//
// Wer sie erst beim Lesen anwendete, hätte eine Datenbank, die wächst, und eine
// Anzeige, die es verschweigt.
func kuerzeVerlauf(ctx context.Context, tx *sql.Tx, karteID int64) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM history
		  WHERE set_id = (SELECT set_id FROM cards WHERE id = ?)
		    AND `+schrittkennung+` NOT IN (
		      SELECT schritt FROM (
		        SELECT `+schrittkennung+` AS schritt, max(id) AS juengste
		          FROM history WHERE set_id = (SELECT set_id FROM cards WHERE id = ?)
		         GROUP BY schritt ORDER BY juengste DESC LIMIT ?))`,
		karteID, karteID, GrenzeVerlauf)
	if err != nil {
		return uebersetze(err, "den Verlauf kürzen")
	}

	return nil
}

// kuerzeSetverlauf hält den Verlauf eines Sets auf [GrenzeVerlauf] Schritte.
//
// **Am Set und nicht an der Karte.** Eine Setangabe gehört keiner
// ([Kartenschreiber.SetwertSetzenMitVorwert]), und der Verlauf ist der des
// Sets — er war es immer, nur führte bis hierher jeder Weg über eine Karte.
//
// **Zwei Abfragen, nicht eine mit einer Unterabfrage.** Der Kartenweg holt
// sein Set über `(SELECT set_id FROM cards WHERE id = ?)` und bleibt so, wie
// er war; ihn über diese Fassung zu führen kostete eine Abfrage mehr, und ein
// Test, der Treiberfehler an einer bestimmten Stelle einspeist, zählt sie mit.
func kuerzeSetverlauf(ctx context.Context, tx *sql.Tx, setID int64) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM history
		  WHERE set_id = ?
		    AND `+schrittkennung+` NOT IN (
		      SELECT schritt FROM (
		        SELECT `+schrittkennung+` AS schritt, max(id) AS juengste
		          FROM history WHERE set_id = ?
		         GROUP BY schritt ORDER BY juengste DESC LIMIT ?))`,
		setID, setID, GrenzeVerlauf)
	if err != nil {
		return uebersetze(err, "den Verlauf kürzen")
	}

	return nil
}
