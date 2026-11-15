package app

// **Der Umschlag an den Flächen ohne Fenster** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Wofür es diesen Bauweg gibt
//
// **Weil §16 es verlangt**: „Was das Fenster kann, können Kommandozeile und
// MCP — oder es steht begründet, warum nicht." Beim Umschlag steht kein
// Grund dagegen: der Dateidialog beantwortet *welche Datei*, und an der
// Kommandozeile steht sie im Argument. Denselben Satz trägt
// `bestandssicherung.go` schon.
//
// **Und er ist klein wie [PeerflaecheUeber]**: geschnürt wird über den
// Setschreiber, hereingenommen über den Importeur, unterschrieben über den
// Schlüsselbund. Wer ihn über [KIAnwendungUeber] baute, zöge vier
// KI-Anbieter und einen Suchindex mit herein, die kein Ruf hier anfasst.

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// UmschlagflaecheUeber baut eine Anwendung, die Umschläge schnürt und liest.
func UmschlagflaecheUeber(db *sql.DB) (*App, error) {
	speicherer, fehler := Setschreibnaht(db)
	if fehler != nil {
		return nil, fehler
	}

	anwendung := NeueApp("")
	anwendung.db = db
	anwendung.MitSpeicherer(speicherer)

	// **Dieselbe Einfuhrnaht wie die Kommandozeile** und nicht eine zweite
	// daneben — „eine Quelle, ein Set" ist eine fachliche Regel und gilt für
	// jede Fläche.
	anwendung.MitImporteur(service.NeuerImporteur(storage.NeuerSetSpeicher(db)).
		MitKoepfen(storage.NeuerSetLeser(db)))

	return anwendung, nil
}
