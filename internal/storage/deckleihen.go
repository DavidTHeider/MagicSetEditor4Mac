package storage

// **Die digitale Deckleihe im Bestand** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Ein Dokument und ein Abzug, beide als JSON** — der Grund steht in der
// Wanderung 0066. Was daneben in Spalten steht, ist der Stand **dieses**
// Rechners und geht nie über die Grenze.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Leihablage hält die Leihen eines Profils.
type Leihablage struct{ db *sql.DB }

// NeueLeihablage baut sie.
func NeueLeihablage(db *sql.DB) *Leihablage { return &Leihablage{db: db} }

// LeiheSchreiben legt eine Leihe an oder schreibt ihren Stand fort.
//
// **Ein Aufruf und kein Anlegen-und-Ändern.** Was sich ändert, ist immer der
// **Stand** — verbraucht, widerrufen; das Dokument bleibt, weil es
// unterschrieben ist. Zwei Wege dafür wären zwei Stellen, an denen jemand
// vergisst, welche gilt.
//
// **Der Konflikt ist der des Profils** (Wanderung 0081): dieselbe Leihe
// liegt beim Verleiher und beim Entleiher, und führt jemand beide Profile in
// einem Bestand, traf `ON CONFLICT (kennung)` die Zeile der anderen Seite.
func (a *Leihablage) LeiheSchreiben(
	ctx context.Context, stand domain.Leihstand,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	dokument, abzug := kodiereLeihe(stand)

	if _, fehler := a.db.ExecContext(ctx,
		`INSERT INTO deckleihen
		   (kennung, profil_id, dokument, abzug, verliehen, verbraucht,
		    widerrufen)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, kennung) DO UPDATE SET
		   verbraucht = excluded.verbraucht,
		   widerrufen = excluded.widerrufen,
		   geaendert  = datetime('now')`,
		stand.Leihe.Kennung, profilID, dokument, abzug,
		stand.Verliehen, stand.Verbraucht, stand.Widerrufen); fehler != nil {
		return uebersetze(fehler, "die Leihe schreiben")
	}

	return nil
}

// Leihe gibt eine Leihe samt dem, was dieses Profil dazu weiß.
func (a *Leihablage) Leihe(
	ctx context.Context, kennung string,
) (domain.Leihstand, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Leihstand{}, profilfehler
	}

	var dokument, abzug, verbraucht string

	var verliehen, widerrufen bool

	fehler := a.db.QueryRowContext(ctx,
		`SELECT dokument, abzug, verliehen, verbraucht, widerrufen
		   FROM deckleihen WHERE kennung = ? AND profil_id = ?`,
		kennung, profilID).Scan(
		&dokument, &abzug, &verliehen, &verbraucht, &widerrufen)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Leihstand{}, fmt.Errorf(
			"%w: die Leihe %q gibt es in diesem Profil nicht",
			domain.ErrNotFound, kennung)
	}

	if fehler != nil {
		return domain.Leihstand{}, uebersetze(fehler, "die Leihe lesen")
	}

	return leiheAus(dokument, abzug, verliehen, verbraucht, widerrufen)
}

// Leihen gibt alle Leihen dieses Profils — die neueste zuerst.
func (a *Leihablage) Leihen(
	ctx context.Context,
) ([]domain.Leihstand, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT dokument, abzug, verliehen, verbraucht, widerrufen
		   FROM deckleihen WHERE profil_id = ?
		  ORDER BY angelegt DESC, kennung`, profilID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Leihen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	gesammelt := []domain.Leihstand{}

	for zeilen.Next() {
		var dokument, abzug, verbraucht string

		var verliehen, widerrufen bool

		if lesefehler := zeilen.Scan(&dokument, &abzug, &verliehen,
			&verbraucht, &widerrufen); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Leihe lesen")
		}

		stand, umwandlungsfehler := leiheAus(
			dokument, abzug, verliehen, verbraucht, widerrufen)
		if umwandlungsfehler != nil {
			return nil, umwandlungsfehler
		}

		gesammelt = append(gesammelt, stand)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Leihen lesen")
	}

	return gesammelt, nil
}

// kodiereLeihe macht aus einem Stand die beiden JSON-Spalten.
//
// # Warum hier kein Fehlerweg steht
//
// **Weil es keinen gibt.** Eine [domain.Deckleihe] trägt Zeichenketten,
// Zeitpunkte und Aufzählungswerte; ein [domain.Partiedeck] dazu Zahlen und
// `map[string]string`. `encoding/json` kann daran nicht scheitern — es
// scheitert an Kanälen, Funktionen und `any`, und keines davon steht darin.
//
// **Der erste Anlauf hatte zwei Fehlerzweige**, und beide waren unbetreten.
// Ein Zweig, den kein Lauf erreicht, sieht aus wie eine geprüfte Zusage (§10)
// — derselbe Grund und dieselbe Form wie bei [alsPartiejson].
func kodiereLeihe(stand domain.Leihstand) (dokument, abzug string) {
	rohesDokument, _ := json.Marshal(stand.Leihe) //nolint:errcheck,errchkjson // siehe oben
	roherAbzug, _ := json.Marshal(stand.Abzug)    //nolint:errcheck,errchkjson // siehe oben

	return string(rohesDokument), string(roherAbzug)
}

// leiheAus baut den Stand aus den Spalten.
func leiheAus(
	dokument, abzug string, verliehen bool, verbraucht string, widerrufen bool,
) (domain.Leihstand, error) {
	stand := domain.Leihstand{
		Verliehen: verliehen, Verbraucht: verbraucht, Widerrufen: widerrufen,
	}

	if fehler := json.Unmarshal([]byte(dokument), &stand.Leihe); fehler != nil {
		return domain.Leihstand{}, domain.MitGrund(domain.ErrValidation,
			fehler, "die gespeicherte Leihe ließ sich nicht lesen")
	}

	if fehler := json.Unmarshal([]byte(abzug), &stand.Abzug); fehler != nil {
		return domain.Leihstand{}, domain.MitGrund(domain.ErrValidation,
			fehler, "der gespeicherte Abzug ließ sich nicht lesen")
	}

	return stand, nil
}
