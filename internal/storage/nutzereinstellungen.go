// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von
// Phase 4.12
// (#1351).
//
// # Ohne Profilklammer
//
// **Sie hängen an einem Nutzer und keinem Profil.** Wer das Profil wechselt,
// bleibt derselbe Mensch — und die Klammer aus Wanderung 0035 säße hier
// falsch: sie ordnet **Inhalte** einem Profil zu, nicht Personen.

package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nutzereinstellungen gibt die gesetzten Einstellungen eines Nutzers.
//
// **Leer und nicht `nil`.** Ein `nil` durchquert die Naht als `null`, und die
// Oberfläche prüfte dann darauf; eine leere Abbildung sagt dasselbe und
// braucht keinen zweiten Zweig (§3).
func (a *Nutzerablage) Nutzereinstellungen(ctx context.Context, nutzerID int64) (
	map[string]string, error,
) {
	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT schluessel, wert FROM nutzereinstellungen WHERE nutzer_id = ?`,
		nutzerID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Nutzereinstellungen lesen")
	}

	defer func() { _ = zeilen.Close() }()

	gelesen := map[string]string{}

	for zeilen.Next() {
		var schluessel, wert string

		if err := zeilen.Scan(&schluessel, &wert); err != nil {
			return nil, uebersetze(err, "eine Einstellungszeile lesen")
		}

		gelesen[schluessel] = wert
	}

	return gelesen, uebersetze(zeilen.Err(), "die Einstellungszeilen lesen")
}

// NutzereinstellungSetzen schreibt ein Paar fort.
//
// **`ON CONFLICT` und nicht Lesen-dann-Schreiben**: zwei gleichzeitige
// Schreibvorgänge auf denselben Schlüssel legten sonst zwei Zeilen an, und der
// Primärschlüssel wiese die zweite ab, statt sie zu setzen.
func (a *Nutzerablage) NutzereinstellungSetzen(
	ctx context.Context, nutzerID int64, schluessel, wert string,
) error {
	if fehler := domain.PruefeNutzereinstellung(schluessel, wert); fehler != nil {
		return fehler
	}

	_, fehler := a.db.ExecContext(ctx,
		`INSERT INTO nutzereinstellungen (nutzer_id, schluessel, wert)
		 VALUES (?, ?, ?)
		 ON CONFLICT (nutzer_id, schluessel) DO UPDATE SET wert = excluded.wert`,
		nutzerID, schluessel, wert)

	return uebersetze(fehler, "die Einstellung %q setzen", schluessel)
}

// NutzerwerteZumProfil gibt die Einstellungen des Menschen, dem das gezeigte
// Profil gehört — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 9.
//
// # Warum eine Abfrage und nicht zwei
//
// **Weil zwei zwei Stände wären.** Wer erst den Eigner holte und dann seine
// Einstellungen, läse zwischen beiden Abfragen einen Profilwechsel mit — und
// bekäme die Einstellungen eines Menschen, dem das Profil nicht mehr gehört
// (§7, Regel 2).
//
// **Ein Profil ohne Eigner gibt es nicht**: der Fremdschlüssel steht seit
// Wanderung 0034. Wo die Unterabfrage nichts findet, gibt es das Profil
// nicht, und dann ist die leere Antwort die richtige — die Vorgaben gelten.
func (a *Nutzerablage) NutzerwerteZumProfil(ctx context.Context) (
	map[string]string, error,
) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, abfragefehler := a.db.QueryContext(ctx,
		`SELECT schluessel, wert FROM nutzereinstellungen
		  WHERE nutzer_id = (SELECT nutzer_id FROM profile WHERE id = ?)`,
		profilID)
	if abfragefehler != nil {
		return nil, uebersetze(abfragefehler, "die Nutzereinstellungen lesen")
	}

	defer func() { _ = zeilen.Close() }()

	gelesen := map[string]string{}

	for zeilen.Next() {
		var schluessel, wert string

		if err := zeilen.Scan(&schluessel, &wert); err != nil {
			return nil, uebersetze(err, "eine Einstellungszeile lesen")
		}

		gelesen[schluessel] = wert
	}

	return gelesen, uebersetze(zeilen.Err(), "die Einstellungszeilen lesen")
}
