package storage

// **Was ein Handel im Besitz bewegt** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum es diese Datei gibt
//
// **Weil sie der zweite Weg in den Besitz ist, und der gehört benannt.**
// `internal/storage/sammlung.go` ist der erste: Anlegen, Ändern, Löschen —
// und die beiden letzten fragen vorher die offenen Zusagen. Hier steht der
// andere: ein Abschluss bewegt Bestand, **weil** eine bestätigte Übergabe ihn
// deckt.
//
// **Die Prüfung ist deshalb eine andere, nicht eine schwächere.** Dort wird
// gefragt, ob eine Änderung zugesagte Stücke wegnähme; hier, ob die Bewegung
// von einer bestätigten Übergabe gedeckt ist und ob deren Fassung noch gegen
// den heutigen Stand rechnet. Beide Wege stehen in
// `internal/werkzeuge/bestandsschreibwege_test.go`, und ein dritter fiele auf.
//
// # Und warum sie nicht in `sammlung.go` steht
//
// **Weil die Datei bei 399 Zeilen steht** (§2). Das ist der Anlass; der Grund
// ist, dass hier eine andere Frage beantwortet wird — ein Handel und nicht
// eine Bearbeitung.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bestandAbgang nimmt Stücke aus einer Besitzzeile.
//
// **Bei null verschwindet die Zeile, und das Journal bleibt.** Das ist der
// Satz, für den es das Journal gibt: wer das letzte Stück weggibt, hat keine
// Zeile mehr, an der ein Verlauf hinge — und gerade dann will jemand wissen,
// wohin es ging.
func bestandAbgang(
	ctx context.Context, an abfrager, profilID int64, vorgang string,
	zeile domain.Journalzeile,
) error {
	// **Derselbe Ableser wie beim Festhalten einer Zusage.** Zwei Fassungen
	// liefen bei der ersten Änderung auseinander, und die eine, die niemand
	// ändert, ist die im fremden Weg — deshalb gibt es nur diese.
	besitz, fehler := besitzInTx(ctx, an, profilID, zeile.ExemplarID)
	if fehler != nil {
		return fehler
	}

	// **Ein Wunsch ist kein Besitz** und geht deshalb auch nicht ab. Ohne
	// diesen Satz ließe sich aus einer Wunschzeile herausbuchen, und der
	// Bestand meldete einen Abgang von etwas, das nie da war.
	if besitz.Wunsch {
		return domain.MitGrund(domain.ErrConflict, nil,
			"aus einem Wunsch geht nichts ab")
	}

	vorrat, fassung := besitz.Menge, besitz.Fassung

	deckung, fehler := gedeckt(
		ctx, an, profilID, zeile.ExemplarID, vorgang, fassung)
	if fehler != nil {
		return fehler
	}

	// **Ohne bestätigte Übergabe wird nichts abgebucht.** Eine offene Zusage
	// hält fest; abgegeben ist erst, was jemand bestätigt hat.
	if deckung < zeile.Menge {
		return domain.MitGrund(domain.ErrConflict, nil,
			"es sind nur %d Stück als übergeben bestätigt, gebucht werden "+
				"sollen %d", deckung, zeile.Menge)
	}

	if vorrat < zeile.Menge {
		return domain.MitGrund(domain.ErrConflict, nil,
			"es liegen nur %d Stück da, abgehen sollen %d",
			vorrat, zeile.Menge)
	}

	if vorrat == zeile.Menge {
		return besitzzeileLoeschen(ctx, an, profilID, zeile.ExemplarID, fassung)
	}

	return besitzzeileMindern(
		ctx, an, profilID, zeile.ExemplarID, fassung, zeile.Menge)
}

// besitzzeileMindern zieht Stücke ab und hebt die Fassung.
func besitzzeileMindern(
	ctx context.Context, an abfrager, profilID, exemplarID, fassung int64,
	menge int,
) error {
	ergebnis, fehler := an.ExecContext(ctx,
		`UPDATE sammlung SET menge = menge - ?, fassung = fassung + 1
		  WHERE id = ? AND profil_id = ? AND fassung = ?`,
		menge, exemplarID, profilID, fassung)
	if fehler != nil {
		return uebersetze(fehler, "die Besitzzeile %d mindern", exemplarID)
	}

	return genauEine(ergebnis, exemplarID)
}

// besitzzeileLoeschen nimmt eine leergewordene Zeile weg.
func besitzzeileLoeschen(
	ctx context.Context, an abfrager, profilID, exemplarID, fassung int64,
) error {
	ergebnis, fehler := an.ExecContext(ctx,
		`DELETE FROM sammlung
		  WHERE id = ? AND profil_id = ? AND fassung = ?`,
		exemplarID, profilID, fassung)
	if fehler != nil {
		return uebersetze(fehler, "die Besitzzeile %d löschen", exemplarID)
	}

	return genauEine(ergebnis, exemplarID)
}

// genauEine weist ab, was keine Zeile getroffen hat.
//
// **Innerhalb einer Transaktion kann das eigentlich nicht geschehen** — die
// Fassung wurde eben gelesen. Der Zweig steht trotzdem da, weil `sql.Result`
// ihn hergibt und weil ein zweiter Schreiber auf derselben Verbindung
// denkbar ist; er ist die Stelle, an der ein solcher Fall laut würde statt
// still zu buchen.
func genauEine(ergebnis interface{ RowsAffected() (int64, error) }, exemplarID int64) error {
	getroffen, fehler := ergebnis.RowsAffected()
	if fehler != nil {
		return uebersetze(fehler, "die Besitzzeile %d ändern", exemplarID)
	}

	if getroffen != 1 {
		return fmt.Errorf(
			"%w — die Besitzzeile %d hat sich unter der Hand geändert",
			domain.ErrUeberholt, exemplarID)
	}

	return nil
}

// bestandZugang legt ein empfangenes Stück an oder erhöht seine Zeile.
//
// **Ohne Erwerbspreis und ohne Währung.** Was der Absender bezahlt hat, ist
// nicht, was es mich gekostet hat; ein Marktwert an dieser Stelle wäre
// erfunden, und beide sähen später aus wie erfasste Zahlen
// ([domain.Journalzeile.AlsExemplar]).
func bestandZugang(
	ctx context.Context, an abfrager, profilID int64,
	zeile domain.Journalzeile,
) error {
	exemplar := zeile.AlsExemplar()

	// **Der erste Befund und nicht alle** — dieselbe Form wie in
	// `arbeitsverlauf.go`: ein Abschluss bricht ab, und wer ihn wiederholen
	// will, braucht einen Satz und keine Liste.
	if befunde := exemplar.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	druck, fehler := druckZeile(ctx, an, druckbezug{printing: exemplar.Printing})
	if fehler != nil {
		return fehler
	}

	// **`ON CONFLICT` und kein Nachschlagen davor.** Zwischen Frage und
	// Antwort läge ein Fenster, und die Tabelle beantwortet dieselbe Frage
	// ohne eines: die Eindeutigkeit über Profil, Druck (Alias und Zeile,
	// Migration 0086), Zustand, Sprache, Ausführung und Wunsch ist genau die
	// Zeile, die gemeint ist.
	// **Der Preis steht ausdrücklich auf null und nicht auf dem Vorgabewert
	// der Spalte.** Der ist die leere Zeichenkette, und die liest sich nicht
	// als Zahl — eine Zeile, die so entstünde, wäre beim nächsten Lesen ein
	// Fehler statt eines preislosen Stücks.
	//
	// **Und der Zusammenschluss lässt den Preis der bestehenden Zeile
	// stehen.** Ihn zu löschen nähme die Zahl weg, die jemand für die
	// gekauften Stücke bezahlt hat; ihn auf die neuen zu erstrecken, gäbe
	// ihnen eine Kostenbasis, die es nie gab. Beides wäre falsch, und die
	// Zeile kann nur eine Zahl tragen. **Was gilt, steht im Journal**: dort
	// hat der Zugang keinen Preis, und dort steht er einzeln. Eine
	// Kostenbasis je Stück ist für 0.5.0 ausdrücklich nicht zugesagt.
	if _, fehler := an.ExecContext(ctx,
		`INSERT INTO sammlung
		   (printing, druck_id, menge, zustand, sprache, finish, lagerort,
		    notiz, erwerbspreis, erwerbswaehrung, wunsch, profil_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, '', 0, ?)
		 ON CONFLICT (profil_id, printing, druck_id, zustand, sprache, finish, wunsch)
		 DO UPDATE SET menge   = menge + excluded.menge,
		               fassung = fassung + 1`,
		exemplar.Printing, druck, exemplar.Menge, string(exemplar.Zustand),
		exemplar.Sprache, string(exemplar.Ausfuehrung()), exemplar.Lagerort,
		exemplar.Notiz, profilID,
	); fehler != nil {
		return uebersetze(fehler, "ein empfangenes Stück eintragen")
	}

	return nil
}
