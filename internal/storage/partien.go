package storage

// **Der Partiespeicher** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier in einer Transaktion steht und warum
//
// **Zustand, Ereignisse und Befehl entstehen zusammen oder gar nicht.** Das
// ist die Zusage aus dem Issue — „Spielereignisse und Snapshots in SQLite
// **atomar** persistieren" —, und sie ist eine Transaktionsgrenze, kein
// Zustandsfeld.
//
// **Was ohne sie geschähe, ist nicht „ein Fehler"**, sondern etwas
// Schlimmeres: ein Zustand ohne seine Ereignisse ließe sich nicht mehr
// erklären, und Ereignisse ohne ihren Zustand wären ein Strom, der ins Leere
// führt. Beides sähe beim Lesen gültig aus.
//
// # Und warum das Schreiben auf die Fassung klemmt
//
// **Damit zwei gleichzeitige Schreiber nicht beide gewinnen.** Die Leitung
// prüft die Fassung fachlich und gibt einen Satz zurück; hier verhindert
// `WHERE fassung = ?` das Rennen. Zwei Prüfungen, zwei Fragen — keine
// Doppelung.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partieablage liest und schreibt Partien samt ihrem Strom.
type Partieablage struct{ db *sql.DB }

// NeuePartieablage gibt die Ablage über diesem Bestand.
func NeuePartieablage(db *sql.DB) *Partieablage {
	return &Partieablage{db: db}
}

// Anlegen legt eine Partie an.
func (a *Partieablage) Anlegen(
	ctx context.Context, partie domain.Partie,
) (domain.Partie, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Partie{}, profilfehler
	}

	partie.Fassung = 1

	if _, fehler := a.db.ExecContext(ctx,
		`INSERT INTO partien (id, profil_id, spiel, stand, fassung, zustand)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		partie.ID, profilID, partie.Spiel, string(partie.Stand),
		partie.Fassung, alsPartiejson(partie)); fehler != nil {
		return domain.Partie{}, uebersetze(fehler, "die Partie anlegen")
	}

	return partie, nil
}

// alsPartiejson kodiert eine Partie für die Spalte.
//
// **Ein Fehler ist hier keiner, der eintreten kann** — dieselbe Form und
// derselbe Grund wie bei [alsMerkmalsjson]: was hineingeht, sind
// Zeichenketten, Zahlen, Wahrheitswerte und Abbildungen davon;
// `encoding/json` kann daran nicht scheitern. Ein Zweig dafür stünde
// ungeprüft da und sähe aus wie eine geprüfte Zusage (§10).
//
// **Bei einem [domain.Ereignis] ist das anders**, und deshalb hat es dort
// einen echten Fehlerweg: sein öffentlicher Teil ist ein `map[string]any`,
// und `any` darf alles sein.
func alsPartiejson(partie domain.Partie) string {
	roh, _ := json.Marshal(partie) //nolint:errcheck,errchkjson // siehe oben

	return string(roh)
}

// Partie gibt eine Partie samt Zustand.
func (a *Partieablage) Partie(
	ctx context.Context, kennung string,
) (domain.Partie, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Partie{}, profilfehler
	}

	var roh string

	fehler := a.db.QueryRowContext(ctx,
		`SELECT zustand FROM partien WHERE id = ? AND profil_id = ?`,
		kennung, profilID).Scan(&roh)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Partie{}, fmt.Errorf(
			"%w: die Partie %q gibt es in diesem Profil nicht",
			domain.ErrNotFound, kennung)
	}

	if fehler != nil {
		return domain.Partie{}, uebersetze(fehler, "die Partie lesen")
	}

	var partie domain.Partie

	if fehler := json.Unmarshal([]byte(roh), &partie); fehler != nil {
		return domain.Partie{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Partie lesen")
	}

	return partie, nil
}

// Partien gibt die Partien dieses Profils, zuletzt geänderte zuerst.
func (a *Partieablage) Partien(ctx context.Context) ([]domain.Partie, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT zustand FROM partien WHERE profil_id = ?
		 ORDER BY geaendert DESC, id`, profilID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Partien lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	gefunden := []domain.Partie{}

	for zeilen.Next() {
		var roh string

		if fehler := zeilen.Scan(&roh); fehler != nil {
			return nil, uebersetze(fehler, "eine Partie lesen")
		}

		var partie domain.Partie

		if fehler := json.Unmarshal([]byte(roh), &partie); fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"eine Partie lesen")
		}

		gefunden = append(gefunden, partie)
	}

	return gefunden, uebersetze(zeilen.Err(), "die Partien lesen")
}

// Anwenden schreibt Zustand, Ereignisse und Befehl in **einer** Transaktion.
func (a *Partieablage) Anwenden(
	ctx context.Context, erwartet int64, befehlID string,
	zustand domain.Partie, ereignisse []domain.Ereignis,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	transaktion, fehler := a.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "das Anwenden beginnen")
	}

	defer func() { _ = transaktion.Rollback() }() //nolint:errcheck // nach Commit ein Nichts

	// **Eine fremde oder fehlende Partie ist „nicht gefunden"** (#1815) — und
	// kein Konflikt. Das Schreiben darunter klemmte schon auf das Profil,
	// antwortete aber auf beides mit „die Fassung galt nicht mehr".
	if fehler := partieImProfil(ctx, transaktion, zustand.ID); fehler != nil {
		return fehler
	}

	if fehler := a.imGeschaeft(
		ctx, transaktion, profilID, erwartet, befehlID, zustand,
		alsPartiejson(zustand), ereignisse); fehler != nil {
		return fehler
	}

	// **Ein Satz und kein Zweig** — dieselbe Form wie in
	// `autoersetzungen.go`: ein eigener Fehlerzweig fürs Abschließen wäre
	// einer, den kein Lauf betreten kann (§3).
	return uebersetze(transaktion.Commit(), "das Anwenden abschließen")
}

// imGeschaeft schreibt Zustand, Strom und Befehl — innerhalb der Transaktion.
//
// **Getrennt vom Ruf darüber**, damit dort die Transaktionsgrenze steht und
// hier, was in ihr geschieht. Zusammen wären es zwei Fragen in einer Funktion
// und über der Zeilengrenze aus §2.
func (a *Partieablage) imGeschaeft(
	ctx context.Context, tx *sql.Tx, profilID, erwartet int64,
	befehlID string, zustand domain.Partie, roh string,
	ereignisse []domain.Ereignis,
) error {
	{
		ergebnis, fehler := tx.ExecContext(ctx,
			`UPDATE partien
			 SET zustand = ?, stand = ?, fassung = ?,
			     geaendert = datetime('now')
			 WHERE id = ? AND profil_id = ? AND fassung = ?`,
			roh, string(zustand.Stand), zustand.Fassung,
			zustand.ID, profilID, erwartet)
		if fehler != nil {
			return uebersetze(fehler, "die Partie fortschreiben")
		}

		gaengig, fehler := ergebnis.RowsAffected()
		if fehler != nil {
			return uebersetze(fehler, "die Partie fortschreiben")
		}

		// **Null Zeilen heißen: die Fassung galt nicht mehr.** Das ist der
		// Konflikt, und er kommt hier heraus statt als stilles Nichts.
		if gaengig == 0 {
			return fmt.Errorf(
				"%w: die Partie %q hat die Fassung %d nicht mehr",
				domain.ErrConflict, zustand.ID, erwartet)
		}
	}

	return schreibeStrom(ctx, tx, zustand.ID, befehlID, ereignisse)
}

// schreibeStrom hängt die Ereignisse an und merkt den Befehl.
func schreibeStrom(
	ctx context.Context, tx *sql.Tx, partieID, befehlID string,
	ereignisse []domain.Ereignis,
) error {
	if len(ereignisse) == 0 {
		return nil
	}

	for _, ereignis := range ereignisse {
		oeffentlich, fehler := json.Marshal(ereignis.Oeffentlich)
		if fehler != nil {
			return domain.MitGrund(domain.ErrValidation, fehler,
				"ein Ereignis kodieren")
		}

		privat, fehler := json.Marshal(ereignis.Privat)
		if fehler != nil {
			return domain.MitGrund(domain.ErrValidation, fehler,
				"eine Ergänzung kodieren")
		}

		if _, fehler := tx.ExecContext(ctx,
			`INSERT INTO partieereignisse
			   (partie_id, folge, art, oeffentlich, privat, kette)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			partieID, ereignis.Folge, ereignis.Art,
			string(oeffentlich), string(privat), ereignis.Kette); fehler != nil {
			return uebersetze(fehler, "ein Ereignis schreiben")
		}
	}

	_, fehler := tx.ExecContext(ctx,
		`INSERT INTO partiebefehle
		   (partie_id, kennung, absender, art, von_folge, bis_folge)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		partieID, befehlID,
		ereignisse[0].Oeffentlich["sitz"], ereignisse[0].Art,
		ereignisse[0].Folge, ereignisse[len(ereignisse)-1].Folge)

	return uebersetze(fehler, "den Befehl merken")
}

// Standsetzen schreibt Pause, Fortsetzung oder Ende.
func (a *Partieablage) Standsetzen(
	ctx context.Context, partieID string, stand domain.Partiestand,
) (domain.Partie, error) {
	partie, fehler := a.Partie(ctx, partieID)
	if fehler != nil {
		return domain.Partie{}, fehler
	}

	partie.Stand = stand
	partie.Fassung++

	// **Das Profil steht schon fest.** [Partieablage.Partie] darüber hat
	// ohne eines gar nicht geantwortet; ein zweiter Zweig dafür wäre einer,
	// den kein Lauf betreten kann (§3).
	profilID, _ := ProfilAus(ctx) //nolint:errcheck // siehe darüber

	if _, fehler := a.db.ExecContext(ctx,
		`UPDATE partien SET zustand = ?, stand = ?, fassung = ?,
		        geaendert = datetime('now')
		 WHERE id = ? AND profil_id = ?`,
		alsPartiejson(partie), string(stand), partie.Fassung,
		partieID, profilID); fehler != nil {
		return domain.Partie{}, uebersetze(fehler, "den Stand schreiben")
	}

	return partie, nil
}
