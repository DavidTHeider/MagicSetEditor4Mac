package storage

// **Was der zweite Faktor sich merkt** — SEC-08, SEC-09, SEC-10
// (#2149,
// #2150,
// #2151,
// Vermerk 0178, Wanderung 0097).
//
// **Der Bestand entscheidet, nicht der Prozess.** Fenster, `mse`, `mse-rest`
// und `mse-mcp` sind vier Prozesse auf einer Datei; was einer im Speicher
// zählte, sähe der nächste nicht. Jede Entscheidung hier fällt in einer
// Transaktion, die mit `_txlock=immediate` die Schreibschranke gleich am
// Anfang nimmt ([mitPragmas]) — zwei Prozesse kommen also nacheinander dran.

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SchrittVerbrauchen nimmt den TOTP-Schritt `schritt` an — wenn er später
// ist als der zuletzt angenommene.
//
// **Die Bedingung steht im `WHERE`**, wie bei [Nutzerablage.CodeVerbrauchen]:
// zwei gleichzeitige Anmeldungen mit demselben Code treffen beim zweiten null
// Zeilen. Falsch heißt: schon verbraucht, ein älterer Schritt — oder kein
// TOTP-Faktor.
func (a *Nutzerablage) SchrittVerbrauchen(ctx context.Context, nutzerID, schritt int64) (bool, error) {
	ergebnis, fehler := a.db.ExecContext(ctx,
		`UPDATE nutzerfaktoren SET letzter_schritt = ?
		 WHERE nutzer_id = ? AND art = 'totp' AND letzter_schritt < ?`,
		schritt, nutzerID, schritt)
	if fehler != nil {
		return false, uebersetze(fehler, "den Schritt verbrauchen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)

	return zeilen == 1, zaehlfehler
}

// FaktorversuchBeginnen meldet einen Codeversuch an — oder sagt, bis wann
// keiner geht.
//
// # Vorab gezählt
//
// **Jeder Versuch zählt als Fehlversuch, bis er gelingt**
// ([Nutzerablage.FaktorversuchGelungen]). Erst zu prüfen und dann zu zählen
// ließe gleichzeitige Versuche am vierten Fehlversuch alle durch; so bekommt
// jeder seine Nummer, und ab der Grenze sperrt schon die Anmeldung
// ([domain.Faktorsperre]). Die Null-Zeit heißt: geprüft werden darf.
func (a *Nutzerablage) FaktorversuchBeginnen(
	ctx context.Context, nutzerID int64, jetzt time.Time,
) (gesperrtBis time.Time, fehler error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, uebersetze(err, "einen Codeversuch anmelden")
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	var fehlversuche, bis int64

	lesefehler := tx.QueryRowContext(ctx,
		`SELECT fehlversuche, gesperrt_bis FROM faktorhaushalt WHERE nutzer_id = ?`,
		nutzerID).Scan(&fehlversuche, &bis)
	if lesefehler != nil && !errors.Is(lesefehler, sql.ErrNoRows) {
		return time.Time{}, uebersetze(lesefehler, "den Faktorhaushalt lesen")
	}

	if jetzt.Unix() < bis {
		return time.Unix(bis, 0).UTC(), nil
	}

	fehlversuche++

	neueSperre := int64(0)
	if dauer := domain.Faktorsperre(int(fehlversuche)); dauer > 0 {
		neueSperre = jetzt.Add(dauer).Unix()
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO faktorhaushalt (nutzer_id, fehlversuche, gesperrt_bis) VALUES (?, ?, ?)
		 ON CONFLICT (nutzer_id) DO UPDATE
		 SET fehlversuche = excluded.fehlversuche, gesperrt_bis = excluded.gesperrt_bis`,
		nutzerID, fehlversuche, neueSperre); err != nil {
		return time.Time{}, uebersetze(err, "einen Codeversuch zählen")
	}

	return time.Time{}, bestaetige(tx, "einen Codeversuch anmelden")
}

// FaktorversuchGelungen setzt den Haushalt zurück — ein Code hat gestimmt.
func (a *Nutzerablage) FaktorversuchGelungen(ctx context.Context, nutzerID int64) error {
	_, fehler := a.db.ExecContext(ctx,
		`DELETE FROM faktorhaushalt WHERE nutzer_id = ?`, nutzerID)

	return uebersetze(fehler, "den Faktorhaushalt leeren")
}
