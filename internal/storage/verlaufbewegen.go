package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zuruecknehmen macht den jüngsten noch geltenden Schritt rückgängig.
//
// Der zweite Rückgabewert sagt, ob es etwas zurückzunehmen gab — **nicht** ob es
// geklappt hat. „Nichts mehr da" ist kein Fehler: ⌘Z am Anfang des Verlaufs ist
// eine gewöhnliche Lage und keine Meldung wert.
func (v *Verlaufsleser) Zuruecknehmen(
	ctx context.Context, setID int64,
) (domain.Verlaufsschritt, bool, error) {
	return v.bewege(ctx, setID, "", true)
}

// Wiederherstellen führt den jüngsten zurückgenommenen Schritt wieder aus.
func (v *Verlaufsleser) Wiederherstellen(
	ctx context.Context, setID int64,
) (domain.Verlaufsschritt, bool, error) {
	return v.bewege(ctx, setID, "", false)
}

// VorgangZuruecknehmen nimmt **einen benannten Vorgang** zurück (#1393).
//
// # Warum es das neben ⌘Z gibt
//
// [Verlaufsleser.Zuruecknehmen] trifft den jüngsten noch geltenden Schritt des
// Sets, und am Fenster ist das genau richtig: ⌘Z heißt „das da eben". Über
// MCP, Kommandozeile und REST schreibt womöglich ein zweiter Prozess in
// dasselbe Set — dann trifft „der jüngste" **dessen** Änderung.
//
// Die Kennung kommt aus [Kartenschreiber.FelderSetzen]; sie stand längst im
// Verlauf und wurde nur nie zurückgegeben.
//
// # Und sie heilt jede Änderung, nicht nur die Masse
//
// Ein Eintrag ohne Gruppe bekommt über [schrittkennung] die Kennung
// `einzeln:<id>` — wer sie hat, nimmt auch eine einzelne Änderung gezielt
// zurück. Das ist der Grund, warum von den drei Wegen im Vermerk dieser
// gebaut wurde: er behebt die Klasse und nicht den Fall (§10).
func (v *Verlaufsleser) VorgangZuruecknehmen(
	ctx context.Context, setID int64, vorgang string,
) (domain.Verlaufsschritt, bool, error) {
	return v.bewege(ctx, setID, vorgang, true)
}

// VorgangWiederherstellen führt einen benannten Vorgang wieder aus.
//
// **Ohne sie wäre die gezielte Rücknahme eine Einbahnstraße** — und ⌘⇧Z an
// einer Fläche gäbe etwas anderes zurück als das, was eben wegging.
func (v *Verlaufsleser) VorgangWiederherstellen(
	ctx context.Context, setID int64, vorgang string,
) (domain.Verlaufsschritt, bool, error) {
	return v.bewege(ctx, setID, vorgang, false)
}

// bewege nimmt zurück oder stellt wieder her.
//
// # Warum beides eine Funktion ist
//
// Es ist derselbe Vorgang mit vertauschten Wegen: Zurücknehmen schreibt den
// `rueckweg` des jüngsten **geltenden** Schritts, Wiederherstellen den `hinweg`
// des ältesten **zurückgenommenen**. Zwei Funktionen daraus zu machen hieße,
// dieselbe Transaktionsführung zweimal zu schreiben.
func (v *Verlaufsleser) bewege(
	ctx context.Context, setID int64, vorgang string, zurueck bool,
) (domain.Verlaufsschritt, bool, error) {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Verlaufsschritt{}, false, uebersetze(err, "den Verlauf des Sets %d bewegen", setID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	// **Nur im Verlauf eines Sets des gezeigten Profils** (#1815).
	if fehler := setImProfil(ctx, tx, setID); fehler != nil {
		return domain.Verlaufsschritt{}, false, fehler
	}

	schritt, err := naechsterSchritt(ctx, tx, setID, vorgang, zurueck)
	if err != nil || len(schritt) == 0 {
		return domain.Verlaufsschritt{}, false, err
	}

	if err := schrittGehen(ctx, tx, schritt, setID, zurueck); err != nil {
		return domain.Verlaufsschritt{}, false, err
	}

	if err := zeigerSetzen(ctx, tx, schritt, zurueck); err != nil {
		return domain.Verlaufsschritt{}, false, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Verlaufsschritt{}, false, uebersetze(err, "den Verlauf bewegen")
	}

	// **Der Schritt, wie die Liste ihn zeigt** ([domain.Verlaufsschritte]):
	// der erste Eintrag vertritt ihn, und gezählt werden Karten (#2396).
	//
	// **Und er trägt seinen neuen Stand.** Wer ihn so zurückgibt, wie er
	// gelesen wurde, meldete eine Rücknahme als „gilt noch" — und die Oberfläche
	// zeigte den Zeiger falsch herum.
	bewegt := domain.Verlaufsschritte(schritt)[0]
	bewegt.Zurueckgenommen = zurueck

	return bewegt, true, nil
}

// schrittGehen schreibt jeden Weg eines Schritts in den Bestand.
//
// **Rückwärts beim Zurücknehmen.** Innerhalb eines Schritts trifft heute keine
// Karte zweimal zu; die Reihenfolge steht trotzdem fest, statt der Abfrage
// überlassen zu bleiben — eine Kette, die nur zufällig stimmt, ist keine.
func schrittGehen(
	ctx context.Context, tx *sql.Tx, schritt []domain.Verlaufseintrag,
	setID int64, zurueck bool,
) error {
	for i := range schritt {
		eintrag := schritt[i]
		if zurueck {
			eintrag = schritt[len(schritt)-1-i]
		}

		// **Ein Schritt der Art `karte` geht einen anderen Weg** — er legt
		// eine ganze Karte an oder entfernt sie, samt ihren Feldwerten.
		// **Ein Schritt der Art `reihenfolge` setzt zwei Stellen** (#628) —
		// in *einem* Eintrag, weil `UNIQUE (set_id, reihenfolge)` einen
		// halben Tausch nicht zulässt.
		if eintrag.Handlung == domain.HandlungReihenfolge {
			stand := eintrag.NachherStellen
			if zurueck {
				stand = eintrag.VorherStellen
			}

			if err := stellenGehen(ctx, tx, stand); err != nil {
				return err
			}

			continue
		}

		// **Und einer der Art `schluesselwort` einen ganzen Wortstand**
		// (#1077).
		if eintrag.Handlung == domain.HandlungSchluesselwort {
			stand := eintrag.NachherWort
			if zurueck {
				stand = eintrag.VorherWort
			}

			if err := schluesselwortGehen(ctx, tx, setID, stand); err != nil {
				return err
			}

			continue
		}

		if eintrag.Handlung == domain.HandlungKarte {
			stand := eintrag.NachherKarte
			if zurueck {
				stand = eintrag.VorherKarte
			}

			if err := ganzeKarteGehen(ctx, tx, eintrag.KarteID, setID, stand); err != nil {
				return err
			}

			continue
		}

		weg := feldweg{
			Feld: eintrag.Feld, Wert: eintrag.Nachher,
			GabEs: eintrag.GabNachher, Reihenfolge: eintrag.StelleNachher,
			Herkunft: eintrag.HerkunftNachher,
		}
		if zurueck {
			weg = feldweg{
				Feld: eintrag.Feld, Wert: eintrag.Vorher,
				GabEs: eintrag.GabVorher, Reihenfolge: eintrag.StelleVorher,
				Herkunft: eintrag.HerkunftVorher,
			}
		}

		if err := wegGehen(ctx, tx, eintrag, setID, weg); err != nil {
			return err
		}
	}

	return nil
}

// zeigerSetzen schreibt den Zeiger für **alle** Zeilen des Schritts.
//
// In einer Anweisung und nicht je Zeile: eine Massenänderung, deren Zeiger zur
// Hälfte gesetzt ist, wäre ein halb zurückgenommener Schritt — und den kann
// niemand zu Ende bringen.
func zeigerSetzen(
	ctx context.Context, tx *sql.Tx, schritt []domain.Verlaufseintrag, zurueck bool,
) error {
	genommen := 0
	if zurueck {
		genommen = 1
	}

	// **Über den Index**, nicht über den Wert: ein Verlaufseintrag wiegt 128
	// Byte, und die Schleife braucht davon eine Zahl.
	kennungen := make([]int64, 0, len(schritt))
	for i := range schritt {
		kennungen = append(kennungen, schritt[i].ID)
	}

	_, err := tx.ExecContext(ctx,
		`UPDATE history SET zurueckgenommen = ?
		  WHERE id IN (SELECT value FROM json_each(?))`,
		genommen, alsJSONFeld(kennungen))
	if err != nil {
		return uebersetze(err, "den Verlauf fortschreiben")
	}

	return nil
}

// wegGehen schreibt einen Weg in den Bestand.
//
// **Wohin, sagt die Handlung.** Bis
// #567 ging
// jeder Weg nach `card_values`; eine Setangabe gehört keiner Karte, und ihre
// Kartenkennung ist null. Eine Rücknahme darüber legte eine Zeile an, die auf
// eine Karte zeigt, die es nicht gibt — SQLite weist sie mit
// `FOREIGN KEY constraint failed` ab, und das ist die freundlichste Form
// dieses Fehlers.
func wegGehen(
	ctx context.Context, tx *sql.Tx, eintrag domain.Verlaufseintrag,
	setID int64, weg feldweg,
) error {
	if eintrag.Handlung == Setwerthandlung {
		return setwegGehen(ctx, tx, setID, weg)
	}

	return kartenwegGehen(ctx, tx, eintrag.KarteID, weg)
}

// setwegGehen schreibt einen Weg nach `set_values`.
//
// **Dieselben drei Fälle wie bei einer Karte**, und aus denselben Gründen: es
// gab die Zeile nicht (dann weg damit), sie hatte eine Stelle (dann dorthin),
// oder sie hatte keine (dann ans Ende).
func setwegGehen(ctx context.Context, tx *sql.Tx, setID int64, weg feldweg) error {
	if !weg.GabEs {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM set_values WHERE set_id = ? AND feld = ?`, setID, weg.Feld)
		if err != nil {
			return uebersetze(err, "die Setangabe %q entfernen", weg.Feld)
		}

		return nil
	}

	if weg.Reihenfolge == nil {
		if _, err := tx.ExecContext(ctx, setwertSetzen,
			setID, weg.Feld, weg.Wert, setID); err != nil {
			return uebersetze(err, "die Setangabe %q setzen", weg.Feld)
		}

		return nil
	}

	if _, err := tx.ExecContext(ctx, setwertSetzenAnStelle,
		setID, weg.Feld, weg.Wert, *weg.Reihenfolge); err != nil {
		return uebersetze(err, "die Setangabe %q an ihre Stelle setzen", weg.Feld)
	}

	return nil
}

// kartenwegGehen schreibt einen Weg nach `card_values` — **samt seiner
// Herkunft** (#1036).
//
// # Warum die Herkunft hierhergehört
//
// Migration 0014 sagt zu, `card_values` und `wert_herkunft` könnten nicht
// auseinanderlaufen. Der **direkte** Schreibweg hielt das; dieser hier ging
// nie durch ihn und schrieb unmittelbar in die Tabelle. Nach der Rücknahme
// eines Scryfall-Bildes blieb die Nennung an einem wiederhergestellten eigenen
// Bild hängen — genau die falsche Nennung, die #1025 ausschließen sollte.
//
// Angestrichen vom Nachlauf über Phase V.
func kartenwegGehen(ctx context.Context, tx *sql.Tx, karteID int64, weg feldweg) error {
	if err := kartenwertGehen(ctx, tx, karteID, weg); err != nil {
		return err
	}

	// **Nach dem Wert, nicht davor.** Beide liegen in derselben Transaktion,
	// die Reihenfolge ändert an der Zusage nichts — aber sie hält die
	// Fehlerwege des Wertes erreichbar: ein Stolpern an der Herkunft davor
	// ließe sie nie betreten (§3).
	return herkunftGehen(ctx, tx, karteID, weg)
}

// kartenwertGehen schreibt den Wert selbst.
func kartenwertGehen(ctx context.Context, tx *sql.Tx, karteID int64, weg feldweg) error {
	if !weg.GabEs {
		// **Es gab das Feld nicht.** Ein leerer Wert wäre nicht dasselbe: eine
		// leere Zeile ist nicht keine Zeile (0010).
		_, err := tx.ExecContext(ctx,
			`DELETE FROM card_values WHERE card_id = ? AND feld = ?`, karteID, weg.Feld)
		if err != nil {
			return uebersetze(err, "das Feld %q entfernen", weg.Feld)
		}

		return nil
	}

	// **An seine Stelle, wenn es eine hatte** (#160).
	//
	// `feldSetzen` hängt hinten an — richtig für ein Feld, das neu entsteht,
	// falsch für eines, das zurückkommt. Der Bestand liest `card_values` nach
	// der Reihenfolge, und die Quelldatei wird danach geschrieben: eine
	// zurückgenommene Entfernung, die das Feld ans Ende setzt, macht aus einer
	// zurückgenommenen Änderung eine bleibende.
	//
	// Ohne Stelle bleibt es beim Anhängen. Das ist kein Nachlassen: ein Feld,
	// das es vorher nicht gab, **hat** keine, und ein Eintrag aus einer
	// Sitzung vor #160 führt sie nicht.
	if weg.Reihenfolge == nil {
		if _, err := tx.ExecContext(ctx, feldSetzen,
			karteID, weg.Feld, weg.Wert, karteID); err != nil {
			return uebersetze(err, "das Feld %q setzen", weg.Feld)
		}

		return nil
	}

	if _, err := tx.ExecContext(ctx, feldSetzenAnStelle,
		karteID, weg.Feld, weg.Wert, *weg.Reihenfolge); err != nil {
		return uebersetze(err, "das Feld %q an seine Stelle setzen", weg.Feld)
	}

	return nil
}

// herkunftGehen führt die Herkunft eines zurückgenommenen Werts mit.
//
// **Leer heißt: von hier** — dann fällt die Zeile. Bliebe sie stehen,
// behauptete die Anwendung über ein wiederhergestelltes eigenes Bild, es sei
// ein Ausschnitt von Scryfall; eine falsche Nennung ist schlimmer als keine.
//
// **Und ein Feld, das es vorher nicht gab, hat auch keine Herkunft.** Der
// Löschzweig darüber räumt den Wert; hier fällt die Zeile mit.
func herkunftGehen(ctx context.Context, tx *sql.Tx, karteID int64, weg feldweg) error {
	if weg.Herkunft == "" || !weg.GabEs {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM wert_herkunft WHERE card_id = ? AND feld = ?`, karteID, weg.Feld)
		if err != nil {
			return uebersetze(err, "die Herkunft von %q löschen", weg.Feld)
		}

		return nil
	}

	_, err := tx.ExecContext(ctx,
		`INSERT INTO wert_herkunft (card_id, feld, herkunft) VALUES (?, ?, ?)
		 ON CONFLICT (card_id, feld) DO UPDATE SET herkunft = excluded.herkunft`,
		karteID, weg.Feld, weg.Herkunft)
	if err != nil {
		return uebersetze(err, "die Herkunft von %q festhalten", weg.Feld)
	}

	return nil
}
