package storage

// **Kontakte aus einem Adressbuch im Bestand** — ID-04
// (#1703),
// Migration 0085.
//
// **Nicht [Kontaktablage]**: dort stehen Gegenüber mit Schlüssel, hier
// Menschen aus einem Adressbuch. Zwei Begriffe, zwei Tafeln (Vermerk 0137).

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Adressbuchablage liest und schreibt die Kontakte eines Profils.
type Adressbuchablage struct {
	db *sql.DB
}

// NeueAdressbuchablage baut sie über einer Datenbank.
func NeueAdressbuchablage(db *sql.DB) *Adressbuchablage { return &Adressbuchablage{db: db} }

// kontaktweg ist die Gestalt eines Wegs in der JSON-Liste.
type kontaktweg struct {
	Art  string `json:"art"`
	Wert string `json:"wert"`
}

// Kontakte gibt alle Kontakte des Profils, nach Namen.
func (a *Adressbuchablage) Kontakte(ctx context.Context) ([]domain.Kontakt, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT id, fassung, name, wege, bild, bildart, ursprung, uebernommen
		   FROM kontakte WHERE profil_id = ?
		  ORDER BY name COLLATE NOCASE, id`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Kontakte lesen")
	}

	defer func() { _ = zeilen.Close() }()

	kontakte := []domain.Kontakt{}

	for zeilen.Next() {
		kontakt, lesefehler := kontaktLesen(zeilen)
		if lesefehler != nil {
			return nil, lesefehler
		}

		kontakte = append(kontakte, kontakt)
	}

	return kontakte, uebersetze(zeilen.Err(), "die Kontakte lesen")
}

// kontaktLesen liest eine Zeile.
func kontaktLesen(zeilen *sql.Rows) (domain.Kontakt, error) {
	var (
		kontakt     domain.Kontakt
		wege        string
		uebernommen string
	)

	if err := zeilen.Scan(&kontakt.ID, &kontakt.Fassung, &kontakt.Name, &wege,
		&kontakt.Bild, &kontakt.Bildart, &kontakt.Ursprung, &uebernommen); err != nil {
		return domain.Kontakt{}, uebersetze(err, "einen Kontakt lesen")
	}

	var gelesen []kontaktweg
	if err := json.Unmarshal([]byte(wege), &gelesen); err != nil {
		return domain.Kontakt{}, fmt.Errorf("die Wege des Kontakts %d lesen: %w", kontakt.ID, err)
	}

	kontakt.Wege = make([]domain.Kontaktweg, 0, len(gelesen))
	for _, weg := range gelesen {
		kontakt.Wege = append(kontakt.Wege, domain.Kontaktweg{Art: domain.Kontaktwegart(weg.Art), Wert: weg.Wert})
	}

	// **Ein unlesbarer Zeitpunkt ist kein Fehler**: er ist eine Auskunft
	// über den Kontakt und nicht der Kontakt; der Nullwert heißt „nicht
	// bekannt“ (§8).
	kontakt.Uebernommen, _ = time.Parse(time.RFC3339, uebernommen) //nolint:errcheck // siehe oben

	return kontakt, nil
}

// KontakteUebernehmen legt an und ersetzt — **in einem Geschäft** (ID-04.AK-02):
// scheitert ein Kontakt, bleibt keiner der anderen, und niemand steht vor
// einer halben Übernahme. Ohne Kennung wird angelegt, mit Kennung gegen die
// Fassung ersetzt.
func (a *Adressbuchablage) KontakteUebernehmen(ctx context.Context, kontakte []*domain.Kontakt) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	return imGeschaeft(ctx, a.db, "Kontakte übernehmen", func(geschaeft *sql.Tx) error {
		for _, kontakt := range kontakte {
			if fehler := kontaktSchreiben(ctx, geschaeft, profilID, kontakt); fehler != nil {
				return fehler
			}
		}

		return nil
	})
}

// kontaktSchreiben legt einen Kontakt an oder ersetzt ihn.
func kontaktSchreiben(ctx context.Context, geschaeft *sql.Tx, profilID int64, kontakt *domain.Kontakt) error {
	if kontakt.ID == 0 {
		_, err := geschaeft.ExecContext(ctx,
			`INSERT INTO kontakte (profil_id, name, wege, bild, bildart, ursprung, uebernommen)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			profilID, kontakt.Name, wegeAlsJSON(kontakt.Wege), bildOderNull(kontakt.Bild),
			kontakt.Bildart, kontakt.Ursprung, jetzt())

		return uebersetze(err, "einen Kontakt anlegen")
	}

	ergebnis, err := geschaeft.ExecContext(ctx,
		`UPDATE kontakte
		    SET name = ?, wege = ?, bild = ?, bildart = ?, ursprung = ?, uebernommen = ?,
		        fassung = fassung + 1
		  WHERE id = ? AND fassung = ? AND profil_id = ?`,
		kontakt.Name, wegeAlsJSON(kontakt.Wege), bildOderNull(kontakt.Bild), kontakt.Bildart,
		kontakt.Ursprung, jetzt(), kontakt.ID, kontakt.Fassung, profilID)
	if err != nil {
		return uebersetze(err, "einen Kontakt ersetzen")
	}

	return kontaktGetroffen(ctx, geschaeft, ergebnis, kontakt.ID, profilID)
}

// KontaktLoeschen nimmt einen Kontakt heraus — gegen seine Fassung.
func (a *Adressbuchablage) KontaktLoeschen(ctx context.Context, id, fassung int64) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM kontakte WHERE id = ? AND fassung = ? AND profil_id = ?`, id, fassung, profilID)
	if err != nil {
		return uebersetze(err, "einen Kontakt löschen")
	}

	return kontaktGetroffen(ctx, a.db, ergebnis, id, profilID)
}

// kontaktGetroffen unterscheidet „gibt es nicht" von „hat sich geändert".
func kontaktGetroffen(ctx context.Context, bei frager, ergebnis sql.Result, id, profilID int64) error {
	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil || zeilen > 0 {
		return zaehlfehler
	}

	var vorhanden int64
	if err := bei.QueryRowContext(ctx,
		`SELECT count(*) FROM kontakte WHERE id = ? AND profil_id = ?`, id, profilID).Scan(&vorhanden); err != nil {
		return uebersetze(err, "den Kontakt nachschlagen")
	}

	if vorhanden == 0 {
		return fmt.Errorf("%w: den Kontakt %d gibt es nicht", domain.ErrNotFound, id)
	}

	return fmt.Errorf("%w — der Kontakt %d hat sich seit dem Lesen geändert", domain.ErrUeberholt, id)
}

// wegeAlsJSON schreibt die Wege als Liste — **nie `null`**, die Bedingung
// verlangt eine Liste.
func wegeAlsJSON(wege []domain.Kontaktweg) string {
	liste := make([]kontaktweg, 0, len(wege))
	for _, weg := range wege {
		liste = append(liste, kontaktweg{Art: string(weg.Art), Wert: weg.Wert})
	}

	roh, _ := json.Marshal(liste) //nolint:errcheck // zwei Zeichenketten je Eintrag lassen sich immer schreiben

	return string(roh)
}

// bildOderNull gibt `NULL` ohne Bild — so hält die Bedingung Bild und Art
// zusammen.
func bildOderNull(bild []byte) any {
	if len(bild) == 0 {
		return nil
	}

	return bild
}
