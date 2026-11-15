package storage

// **Die Folge einer Handelsnachricht** —
// #1637.
//
// # Warum sie im Bestand steht und nicht im Arbeitsspeicher
//
// **Weil sie einen Neustart überleben muss.** Die Gegenseite vergleicht sie
// mit dem, was sie schon hat ([domain.Nachzuliefern]); ein Zähler, der bei
// jedem Start wieder bei eins anfinge, ließe sie auf einen Verlauf
// nachliefern, den es nicht gibt.
//
// # Und warum hochzählen und lesen ein Ruf sind
//
// **Weil zwei Rufe ein Fenster dazwischen hätten.** Zwei Nachrichten, die
// beide „drei" lesen und beide „vier" schreiben, trügen dieselbe Folge — und
// drüben wäre eine davon unsichtbar. `INSERT … ON CONFLICT DO UPDATE …
// RETURNING` tut beides in einer Anweisung.

import (
	"context"
	"database/sql"
)

// Folgeablage vergibt die Folge je Vorgang.
type Folgeablage struct{ db *sql.DB }

// NeueFolgeablage baut sie über einer Datenbank.
func NeueFolgeablage(db *sql.DB) *Folgeablage { return &Folgeablage{db: db} }

// NaechsteFolge zählt die Folge dieses Vorgangs hoch und gibt sie.
func (a *Folgeablage) NaechsteFolge(
	ctx context.Context, vorgang string,
) (int64, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return 0, fehler
	}

	var folge int64

	if fehler := a.db.QueryRowContext(ctx,
		`INSERT INTO handelsfolge (profil_id, vorgang, folge)
		 VALUES (?, ?, 1)
		 ON CONFLICT (profil_id, vorgang)
		   DO UPDATE SET folge = folge + 1
		 RETURNING folge`,
		profilID, vorgang).Scan(&folge); fehler != nil {
		return 0, uebersetze(fehler, "die Folge einer Nachricht vergeben")
	}

	return folge, nil
}
