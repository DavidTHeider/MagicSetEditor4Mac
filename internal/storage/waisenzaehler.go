package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// **Verwaiste Zeilen zählen** —
// #1220.
//
// # Wozu
//
// `0022_waisen.sql` räumt weg, was ein Bestand vor dem 2026-08-15 angesammelt
// hat: bis dahin war `foreign_keys` aus, und jede Kaskade im Schema ein toter
// Buchstabe.
//
// **Damit ist der Bestand sauber — und niemand merkte, wenn er es nicht mehr
// wäre.** Das Issue nennt es als eigenen Punkt: „ein Wächter, der es merkt.
// `mse diagnose` nennt Schemastand und Paketlage; eine Zeile »verwaiste
// Zeilen: n« gehörte dazu."
//
// # Was gezählt wird
//
// Die drei Stellen, an denen es Waisen geben kann: Kartenwerte ohne Karte,
// Setangaben ohne Set, Suchzeilen ohne Karte. `card_suche` steht dabei, weil
// sie ihren eigenen Auslöser hat — sie kann auseinanderlaufen, ohne dass
// `card_values` es täte.
//
// # Was es nicht ist
//
// **Keine Reparatur.** Der Zähler sagt eine Zahl; wer sie größer als null
// sieht, hat einen Befund und keine erledigte Sache.

// Waisenzahlen sind die verwaisten Zeilen je Tabelle.
type Waisenzahlen struct {
	Kartenwerte int
	Setangaben  int
	Suchzeilen  int
}

// Zusammen gibt die Summe — null heißt: nichts liegt herum.
func (w Waisenzahlen) Zusammen() int {
	return w.Kartenwerte + w.Setangaben + w.Suchzeilen
}

// Waisen zählt die verwaisten Zeilen eines Bestands.
func Waisen(ctx context.Context, db *sql.DB) (Waisenzahlen, error) {
	var zahlen Waisenzahlen

	for _, was := range []struct {
		abfrage string
		ziel    *int
	}{
		{`SELECT count(*) FROM card_values WHERE card_id NOT IN (SELECT id FROM cards)`, &zahlen.Kartenwerte},
		{`SELECT count(*) FROM set_values WHERE set_id NOT IN (SELECT id FROM sets)`, &zahlen.Setangaben},
		{`SELECT count(*) FROM card_suche WHERE card_id NOT IN (SELECT id FROM cards)`, &zahlen.Suchzeilen},
	} {
		if err := db.QueryRowContext(ctx, was.abfrage).Scan(was.ziel); err != nil {
			return Waisenzahlen{}, uebersetze(err, "verwaiste Zeilen zählen")
		}
	}

	return zahlen, nil
}

// WaisenText gibt die Zeile für die Diagnose.
//
// **Im Klartext und nicht als Verbund**: die Diagnose ist ein Text, den jemand
// in einen Bericht klebt.
func WaisenText(zahlen Waisenzahlen) string {
	if zahlen.Zusammen() == 0 {
		return "keine"
	}

	return fmt.Sprintf("%d (Kartenwerte %d, Setangaben %d, Suchzeilen %d)",
		zahlen.Zusammen(), zahlen.Kartenwerte, zahlen.Setangaben, zahlen.Suchzeilen)
}
