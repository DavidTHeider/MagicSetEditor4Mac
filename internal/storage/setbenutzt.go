// **Zuletzt benutzt** —
// #891,
// Werkstück 3 von Phase Y.
//
// # Die Regel kommt aus MSE2
//
// `Settings::addRecentFile` (`src/data/settings.cpp`, Fassung f9d9356 vom
// 2021-07-06, Zeilen 186–200): absoluten Pfad bilden, **Doppel entfernen**,
// vorn einfügen, auf `max_recent_sets` = 9 kürzen.
//
// # Warum hier keine Liste steht
//
// **Der Bestand ist die Liste.** `sets.geoeffnet` trägt, wann ein Set zuletzt
// benutzt wurde, und [SetLeser.Setkoepfe] sortiert danach. Doppel und Kürzung
// fallen damit weg: ein Set steht **einmal** im Bestand, und wer die ersten
// neun nimmt, hat die neun jüngsten.
//
// Eine zweite Liste daneben wäre eine zweite Wahrheit — und die, die niemand
// aufräumt, wenn jemand ein Set entfernt.
//
// **Was fehlte**: die Spalte wurde beim Anlegen geschrieben und nie wieder.
// Sie sagte „zuletzt **eingelesen**", nicht „zuletzt benutzt".

package storage

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SetBenutzt merkt sich, dass ein Set gerade aufgeschlagen wurde.
func (s *SetSpeicher) SetBenutzt(ctx context.Context, setID int64) error {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := s.db.ExecContext(ctx,
		`UPDATE sets SET geoeffnet = ? WHERE id = ? AND profil_id = ?`,
		Zeitstempel(s.jetzt()), setID, profilID)
	if err != nil {
		return uebersetze(err, "das Set als benutzt vermerken")
	}

	// **Die Zahl der betroffenen Zeilen ist die Prüfung.** Ein `UPDATE` auf
	// eine Kennung, die es nicht gibt, glückt und ändert nichts; ohne diese
	// Wache sähe das aus wie ein Erfolg, und die Liste bliebe, wie sie war.
	betroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return uebersetze(zaehlfehler, "die geänderten Zeilen zählen")
	}

	if betroffen == 0 {
		return fmt.Errorf("%w: das Set %d gibt es nicht", domain.ErrNotFound, setID)
	}

	return nil
}
