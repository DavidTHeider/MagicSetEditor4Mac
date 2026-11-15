package storage

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// feldweg ist, was ein Feldeintrag in `hinweg` und `rueckweg` trägt.
type feldweg struct {
	Feld  string `json:"feld"`
	Wert  string `json:"wert"`
	GabEs bool   `json:"gabEs"`

	// Reihenfolge ist die Stelle, an der das Feld in der Karte stand.
	//
	// **Ein Zeiger, damit „hatte keine" von „hatte die 0" zu unterscheiden
	// ist.** Ein Feld, das es vorher nicht gab, hat keine Stelle, und eine zu
	// erfinden hieße, sich eine Ordnung auszudenken, die nie dastand.
	//
	// **Und deshalb keine Migration**: `hinweg` und `rueckweg` sind JSON, und
	// eine Zeile aus einer älteren Sitzung trägt den Schlüssel schlicht nicht
	// — sie kommt als `nil` an und wird behandelt wie ein Feld ohne Stelle
	// (#160).
	Reihenfolge *int64 `json:"reihenfolge,omitempty"`

	// Herkunft ist die Adresse, aus der dieser Wert stammte — leer: von hier.
	//
	// **Damit eine Rücknahme sie mitführen kann**
	// (#1036).
	// Migration 0014 sagt zu, `card_values` und `wert_herkunft` könnten nicht
	// auseinanderlaufen — der direkte Schreibweg hielt das, die Rücknahme ging
	// nie durch ihn. Ohne diese Angabe im Verlaufseintrag ist die Herkunft des
	// vorigen Werts nach der Änderung nicht mehr zu rekonstruieren.
	//
	// **Und deshalb keine Migration**, wie bei `Reihenfolge` darüber: `hinweg`
	// und `rueckweg` sind JSON, und eine Zeile aus einer älteren Sitzung trägt
	// den Schlüssel schlicht nicht — sie kommt als leer an, und das heißt „von
	// hier". Für den Zeitraum davor ist das eine **Tatsache**: es gab keine
	// Netzübernahme in ein Kartenfeld (0013).
	Herkunft string `json:"herkunft,omitempty"`
}

// verlaufsspalten ist die Zeile, die `eintragAus` erwartet — an **einer**
// Stelle, weil drei Abfragen sie lesen und eine vergessene Spalte erst beim
// Scan auffällt.
const verlaufsspalten = `id, karte_id, handlung, hinweg, rueckweg,
                         zurueckgenommen, urheber, gruppe, herkunft, zeitpunkt`

// Verlaufsleser liest und bewegt den Verlauf eines Sets.
type Verlaufsleser struct{ db *sql.DB }

// NeuerVerlaufsleser baut den Leser über einer Datenbank.
func NeuerVerlaufsleser(db *sql.DB) *Verlaufsleser { return &Verlaufsleser{db: db} }

// Verlauf gibt die Einträge eines Sets, älteste zuerst.
func (v *Verlaufsleser) Verlauf(ctx context.Context, setID int64) ([]domain.Verlaufseintrag, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, v.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := v.db.QueryContext(ctx,
		`SELECT `+verlaufsspalten+` FROM history WHERE set_id = ? ORDER BY id`, setID)
	if err != nil {
		return nil, uebersetze(err, "den Verlauf des Sets %d lesen", setID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	var verlauf []domain.Verlaufseintrag

	for zeilen.Next() {
		eintrag, fehler := eintragAus(zeilen)
		if fehler != nil {
			return nil, fehler
		}

		verlauf = append(verlauf, eintrag)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "den Verlauf des Sets %d lesen", setID)
	}

	return verlauf, nil
}

// eintragAus liest eine Zeile in einen Eintrag.
func eintragAus(zeilen *sql.Rows) (domain.Verlaufseintrag, error) {
	var (
		eintrag         domain.Verlaufseintrag
		hin, zurueck    string
		zurueckgenommen int
		urheber         string
	)

	if err := zeilen.Scan(&eintrag.ID, &eintrag.KarteID, &eintrag.Handlung,
		&hin, &zurueck, &zurueckgenommen, &urheber, &eintrag.Gruppe,
		&eintrag.Herkunft, &eintrag.Zeitpunkt); err != nil {
		return domain.Verlaufseintrag{}, uebersetze(err, "einen domain.Verlaufseintrag lesen")
	}

	// **Ein Schritt der Art `karte` trägt einen ganzen Kartenstand**, keinen
	// Feldweg — und wer ihn als Feldweg läse, bekäme ein leeres Feld und
	// hielte es für eine Änderung an einem Feld ohne Namen
	// (#621).
	if eintrag.Handlung == domain.HandlungKarte {
		return kartenschrittAus(eintrag, hin, zurueck, zurueckgenommen, urheber)
	}

	// **Und einer der Art `reihenfolge` zwei Stellen** (#628).
	if eintrag.Handlung == domain.HandlungReihenfolge {
		return stellenschrittAus(eintrag, hin, zurueck, zurueckgenommen, urheber)
	}

	// **Und einer der Art `schluesselwort` zwei Wortstände** (#1077).
	if eintrag.Handlung == domain.HandlungSchluesselwort {
		return schluesselwortschrittAus(eintrag, hin, zurueck, zurueckgenommen, urheber)
	}

	nachher, err := wegAus(hin)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	vorher, err := wegAus(zurueck)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	eintrag.Feld = vorher.Feld
	eintrag.Vorher, eintrag.GabVorher = vorher.Wert, vorher.GabEs
	eintrag.Nachher, eintrag.GabNachher = nachher.Wert, nachher.GabEs
	eintrag.StelleVorher, eintrag.StelleNachher = vorher.Reihenfolge, nachher.Reihenfolge
	eintrag.HerkunftVorher, eintrag.HerkunftNachher = vorher.Herkunft, nachher.Herkunft
	eintrag.Zurueckgenommen = zurueckgenommen != 0
	eintrag.Urheber = domain.Urheber(urheber)

	return eintrag, nil
}

// wegAus liest einen Weg aus seinem JSON.
func wegAus(roh string) (feldweg, error) {
	var weg feldweg

	if roh == "" {
		return weg, nil
	}

	if err := json.Unmarshal([]byte(roh), &weg); err != nil {
		return weg, uebersetze(err, "einen Verlaufsweg lesen")
	}

	return weg, nil
}

// kartenschrittAus füllt einen Eintrag der Art [domain.HandlungKarte].
func kartenschrittAus(
	eintrag domain.Verlaufseintrag, hin, zurueck string,
	zurueckgenommen int, urheber string,
) (domain.Verlaufseintrag, error) {
	nachher, err := kartenstandAus(hin)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	vorher, err := kartenstandAus(zurueck)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	eintrag.VorherKarte, eintrag.NachherKarte = vorher, nachher
	eintrag.Zurueckgenommen = zurueckgenommen != 0
	eintrag.Urheber = domain.Urheber(urheber)

	return eintrag, nil
}

// stellenschrittAus füllt einen Eintrag der Art `reihenfolge`.
func stellenschrittAus(
	eintrag domain.Verlaufseintrag, hin, zurueck string,
	zurueckgenommen int, urheber string,
) (domain.Verlaufseintrag, error) {
	nachher, err := stellenstandAus(hin)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	vorher, err := stellenstandAus(zurueck)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	eintrag.VorherStellen, eintrag.NachherStellen = vorher, nachher
	eintrag.Zurueckgenommen = zurueckgenommen != 0
	eintrag.Urheber = domain.Urheber(urheber)

	return eintrag, nil
}
