package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SetLeser liest den Bestand. Getrennt vom Speicher, weil Lesen und Schreiben
// verschiedene Fragen beantworten und verschieden oft geändert werden.
type SetLeser struct {
	db *sql.DB
}

// NeuerSetLeser baut den Leser über einer Datenbank.
func NeuerSetLeser(db *sql.DB) *SetLeser { return &SetLeser{db: db} }

const setkoepfeLesen = `
SELECT sets.id, sets.name, sets.spiel, sets.stil, sets.herkunft,
       (SELECT count(*) FROM cards WHERE cards.set_id = sets.id)
  FROM sets
 WHERE sets.profil_id = ?
 ORDER BY sets.geoeffnet DESC, sets.id DESC`

// Setkoepfe gibt alle Sets, zuletzt geöffnete zuerst.
func (l *SetLeser) Setkoepfe(ctx context.Context) ([]domain.SetHead, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := l.db.QueryContext(ctx, setkoepfeLesen, profilID)
	if err != nil {
		return nil, uebersetze(err, "Sets lesen")
	}

	defer func() { _ = zeilen.Close() }()

	var koepfe []domain.SetHead

	for zeilen.Next() {
		var kopf domain.SetHead
		if err := zeilen.Scan(&kopf.ID, &kopf.Name, &kopf.Game, &kopf.Style,
			&kopf.Origin, &kopf.Cards); err != nil {
			return nil, uebersetze(err, "ein Set lesen")
		}

		koepfe = append(koepfe, kopf)
	}

	return koepfe, uebersetze(zeilen.Err(), "Sets lesen")
}

// Setstil gibt das Stilpaket eines Sets.
//
// **Eine eigene Abfrage und kein Griff in [SetLeser.Setkoepfe]** (G6,
// #1463):
// jene liest **jedes** Set dieses Profils samt einer Zählung über `cards`,
// und der Aufbau einer Partie fragt je Set einmal danach. Eine Liste zu
// holen, um einen Namen zu lesen, lässt jemanden warten — derselbe Grund
// steht bei [domain.SetHead.Cards].
func (l *SetLeser) Setstil(ctx context.Context, setID int64) (string, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return "", profilfehler
	}

	var stil string

	fehler := l.db.QueryRowContext(ctx,
		`SELECT stil FROM sets WHERE id = ? AND profil_id = ?`,
		setID, profilID).Scan(&stil)

	if errors.Is(fehler, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: das Set %d gibt es in diesem Profil nicht",
			domain.ErrNotFound, setID)
	}

	if fehler != nil {
		return "", uebersetze(fehler, "den Stil eines Sets lesen")
	}

	return stil, nil
}

// Setzeile gibt die Bestandszeile **eines** Sets — Name, Spiel, Stil, Herkunft.
//
// **Eine eigene Abfrage und kein Griff in [SetLeser.Setkoepfe]** — derselbe
// Grund wie bei [SetLeser.Setstil]: jene liest jedes Set dieses Profils samt
// einer Zählung über `cards`, und wer daraus einen Namen liest, lässt jemanden
// warten.
//
// **Ohne die Kartenzahl**, und das steht hier, damit es niemand für ein
// Versehen hält: [domain.SetHead.Cards] bleibt null. Der Zähler ist genau das,
// was diese Abfrage einspart; wer ihn braucht, fragt die Liste. Der heutige
// Verbraucher ist das Zusammensetzen eines Sets ohne Quelldatei
// (#1746),
// und das zählt keine Karten — es schreibt sie.
//
// **Nicht zu verwechseln mit [SetLeser.Setkopf]**: der gibt die fünf
// Formatangaben aus dem Dateikopf, dies hier die Zeile im Bestand.
func (l *SetLeser) Setzeile(ctx context.Context, setID int64) (domain.SetHead, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.SetHead{}, profilfehler
	}

	kopf := domain.SetHead{ID: setID}

	fehler := l.db.QueryRowContext(ctx,
		`SELECT name, spiel, stil, herkunft FROM sets WHERE id = ? AND profil_id = ?`,
		setID, profilID).Scan(&kopf.Name, &kopf.Game, &kopf.Style, &kopf.Origin)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.SetHead{}, fmt.Errorf(
			"%w: das Set %d gibt es in diesem Profil nicht", domain.ErrNotFound, setID)
	}

	if fehler != nil {
		return domain.SetHead{}, uebersetze(fehler, "die Bestandszeile eines Sets lesen")
	}

	return kopf, nil
}

const kartenLesen = `
SELECT cards.id, cards.reihenfolge, card_values.feld, card_values.wert
  FROM cards
  LEFT JOIN card_values ON card_values.card_id = cards.id
 WHERE cards.set_id = ?
 ORDER BY cards.reihenfolge, card_values.reihenfolge`

// Karten gibt die Karten eines Sets in ihrer Reihenfolge.
//
// **Eine Abfrage, nicht eine je Karte.** Bei 300 Karten wären das 301 Wege zur
// Datenbank; der Verbund kostet einmal und liefert dasselbe.
//
// Der äußere Verbund (`LEFT JOIN`) ist Absicht: eine Karte ohne einen einzigen
// Feldwert ist in einem fremden Set der Normalfall (Phase 2 hat sie eigens
// aufgehoben), und ein innerer Verbund ließe sie verschwinden.
func (l *SetLeser) Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx, kartenLesen, setID)
	if err != nil {
		return nil, uebersetze(err, "Karten lesen")
	}

	defer func() { _ = zeilen.Close() }()

	var (
		karten  []*domain.CardRow
		laufend *domain.CardRow
	)

	for zeilen.Next() {
		var (
			id     int64
			nummer int
			feld   sql.NullString
			wert   sql.NullString
		)

		if err := zeilen.Scan(&id, &nummer, &feld, &wert); err != nil {
			return nil, uebersetze(err, "eine Karte lesen")
		}

		if laufend == nil || laufend.ID != id {
			// **Das Set steht an der Karte** — seit dem 2026-08-26
			// (#1164).
			//
			// Hier stand es nicht, und jeder Aufrufer, der eine dieser
			// Karten weiterreicht, bekam eine mit `SetID: 0`. Der Planbauer
			// fragt danach — `StilFuerKarte` sucht den Stil des Sets —, und
			// bei null findet er keines: gezeichnet wurde der **Behelfsplan**,
			// 375 × 523, ohne dass irgendwo ein Fehler stand.
			laufend = &domain.CardRow{ID: id, SetID: setID, Number: nummer}
			karten = append(karten, laufend)
		}

		if feld.Valid {
			laufend.Values = append(laufend.Values,
				domain.Value{Field: feld.String, Text: wert.String})
		}
	}

	return karten, uebersetze(zeilen.Err(), "Karten lesen")
}

// Rohbloecke gibt die roh mitgeführten Bereiche eines Sets.
//
// Sie sind die Grundlage der Vorbehalte, die #39 sichtbar machen will: was
// beim Import nicht gedeutet wurde, steht hier und kann angezeigt werden,
// statt unbemerkt mitzureisen.
func (l *SetLeser) Rohbloecke(ctx context.Context, setID int64) ([]domain.RawBlock, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx,
		`SELECT datei, zeile, inhalt FROM raw_blocks WHERE set_id = ? ORDER BY reihenfolge`, setID)
	if err != nil {
		return nil, uebersetze(err, "Rohblöcke lesen")
	}

	defer func() { _ = zeilen.Close() }()

	var bloecke []domain.RawBlock

	for zeilen.Next() {
		var block domain.RawBlock
		if err := zeilen.Scan(&block.File, &block.Line, &block.Text); err != nil {
			return nil, uebersetze(err, "einen Rohblock lesen")
		}

		bloecke = append(bloecke, block)
	}

	return bloecke, uebersetze(zeilen.Err(), "Rohblöcke lesen")
}

// Herkuenfte gibt je Feld die Adresse, aus der sein **aktueller** Wert stammt.
//
// **Nur die Felder, die eine haben.** Ein Feld ohne Eintrag ist von hier — aus
// einer Eingabe, aus einer Vorlage, aus einer Einfuhr. Eine Karte mit
// zweihundert Feldern und einem übernommenen Bild gibt damit einen Eintrag
// zurück und nicht zweihundert.
//
// **Die Zusage steht in 0014**: gepflegt wird die Tabelle in derselben
// Transaktion wie der Wert selbst. Was hier zurückkommt, gehört zu dem, was in
// `card_values` steht — und nicht zu etwas, das einmal dastand.
func (l *SetLeser) Herkuenfte(ctx context.Context, karteID int64) (map[string]string, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := karteImProfil(ctx, l.db, karteID); fehler != nil {
		return nil, fehler
	}

	return herkuenfteVon(ctx, l.db, karteID)
}

// Stilwerte gibt die setweiten Stiloptionen eines Sets — für den
// Setschreiber, der sie in den Block `styling:` schreibt (F05,
// #2247).
//
// **Derselbe Weg wie [Stilablage.Stilwerte]** und keine zweite Abfrage: zwei
// Antworten auf dieselbe Frage liefen sonst auseinander.
func (l *SetLeser) Stilwerte(ctx context.Context, setID int64) ([]domain.Value, error) {
	return NeueStilablage(l.db).Stilwerte(ctx, setID)
}
