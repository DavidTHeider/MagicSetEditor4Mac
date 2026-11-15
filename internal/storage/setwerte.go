package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setwerte gibt die Angaben eines Sets — Titel, Symbol, Copyright, Setcode.
//
// # Warum es das erst jetzt gibt
//
// `set_values` wurde seit Phase 2 **geschrieben und nie gelesen**. Aufgefallen
// ist das dem Zweitblick über Phase 9, und zwar an einer Wirkung: die Ausfuhr
// nach MTG JSON baute ein Set ohne seine Angaben, und der Setcode ging dabei
// verloren — aus `PRB` wurde der Setname.
//
// Eine Tabelle, in die nur hineingeschrieben wird, ist dieselbe Klasse wie
// `sets.herkunft` in Phase 2: eine Zusage ohne Einlösung.
//
// **In ihrer Reihenfolge**, wie die Kartenwerte: sie ist die der Datei, und ein
// Rundlauf, der sie umstellt, ändert die Datei ungefragt.
func (l *SetLeser) Setwerte(ctx context.Context, setID int64) ([]domain.Value, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx,
		`SELECT feld, wert FROM set_values WHERE set_id = ? ORDER BY reihenfolge`, setID)
	if err != nil {
		return nil, uebersetze(err, "Setangaben lesen")
	}

	defer func() { _ = zeilen.Close() }()

	var werte []domain.Value

	for zeilen.Next() {
		var wert domain.Value
		if err := zeilen.Scan(&wert.Field, &wert.Text); err != nil {
			return nil, uebersetze(err, "eine Setangabe lesen")
		}

		werte = append(werte, wert)
	}

	return werte, uebersetze(zeilen.Err(), "Setangaben lesen")
}
