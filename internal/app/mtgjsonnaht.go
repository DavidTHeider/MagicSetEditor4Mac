// **MTG JSON ohne Fenster** —
// #1291.
//
// # Warum es diese Naht gibt, obwohl `App` dasselbe kann
//
// [App.MTGJSONOeffnen] und [App.MTGJSONSchreiben] sind **Bindungen**: sie
// nehmen Zeichenketten entgegen, geben DTOs mit übersetzten Fehlern zurück und
// kennen ein `Abgebrochen`, das nur ein Dialog auslösen kann. Eine Fläche ohne
// Fenster braucht das Gegenteil — eine Kennung, einen Pfad und einen
// Go-Fehler, den `errors.Is` prüfen kann (§8).
//
// **Eine eigene Datei**, weil `mcpnaht.go` an der Zeilengrenze steht (§2).

package app

import (
	"context"
	"database/sql"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// MTGJSONEinfahrt ist, was ein Einlesen hergibt.
//
// **Ohne `service`-Typen**, damit eine Fläche ohne Fenster sie lesen kann,
// ohne `internal/service` zu kennen (§2) — wie [Bildschreibung].
type MTGJSONEinfahrt struct {
	// Angelegt sind die Kennungen der Sets, die entstanden sind.
	Angelegt []int64

	// SchonOffen sagt, dass aus derselben Quelle schon ein Set im Bestand
	// liegt — „eine Quelle, ein Set". Dann steht in `Angelegt` dessen
	// Kennung, und nichts ist neu entstanden.
	SchonOffen bool

	Befunde []string
}

// MTGJSONLesen liest eine MTG-JSON-Datei oder ein Verzeichnis in den Bestand.
//
// **Ein Abbruch gibt das Ergebnis mit** (#155). Eine AllSets-Datei führt
// Hunderte Sets, jedes in einer eigenen Transaktion; was vor dem Fehler
// durchkam, steht im Bestand. Wer nur den Fehler weiterreicht, lässt Sets
// zurück, von denen niemand weiß.
func MTGJSONLesen(ctx context.Context, db *sql.DB, pfad, spiel string) (
	*MTGJSONEinfahrt, error,
) {
	ergebnis, err := Einfuhrnaht(db).MTGJSON(ctx, pfad, spiel)
	if ergebnis == nil {
		return nil, err
	}

	eingefahren := &MTGJSONEinfahrt{
		Angelegt: ergebnis.Angelegt, SchonOffen: ergebnis.SchonOffen,
		Befunde: alsMeldungen(ergebnis.Befunde),
	}

	if eingefahren.Angelegt == nil {
		eingefahren.Angelegt = []int64{}
	}

	return eingefahren, err
}

// MTGJSONAusfuhr ist, was ein Schreiben hergibt.
type MTGJSONAusfuhr struct {
	Pfad   string
	Karten int

	Befunde []string
}

// MTGJSONSchreiben schreibt ein Set als MTG JSON an einen Pfad.
func MTGJSONSchreiben(ctx context.Context, db *sql.DB, setID int64, spiel, ziel string) (
	*MTGJSONAusfuhr, error,
) {
	ausfuhr, err := service.NeuerExporteur(storage.NeuerSetLeser(db)).
		MitPruefer(AutorisierungAus(db)).
		MTGJSON(ctx, setID, spiel, ziel)
	if err != nil {
		return nil, err
	}

	return &MTGJSONAusfuhr{
		Pfad: ausfuhr.Pfad, Karten: ausfuhr.Karten, Befunde: alsMeldungen(ausfuhr.Befunde),
	}, nil
}

// MTGJSONSpiele nennt die Spiele, für die es eine Zuordnung gibt.
//
// **Ohne Bestand**, weil die Zuordnungen im Programm stehen und nicht in der
// Datenbank — dieselbe Auskunft wie [App.MTGJSONSpiele], nur ohne Fenster.
func MTGJSONSpiele() []string {
	spiele := mtgjson.Spiele()
	slices.Sort(spiele)

	return spiele
}
