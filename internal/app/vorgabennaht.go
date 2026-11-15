// **Die Verdrahtung der Profilvorgaben** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Eine eigene Datei**, weil sie fünf Quellen zusammenführt und `start.go`
// damit über die Zeilengrenze ginge (§2). Der Schnitt folgt der Frage: dort
// steht, **was die Anwendung hat**, hier, **woher das Angebot kommt**.

package app

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// vorgabenverwaltungUeber baut die Verwaltung samt ihrem Angebot.
//
// **Der Paketkatalog wird beim Fragen aufgefrischt** — dieselbe Überlegung wie
// bei [App.Pakete]: ein Katalog, der die Platte nicht kennt, ist eine Liste von
// Behauptungen, und eine Vorgabe auf ein Paket, das dasteht, würde abgewiesen.
//
// **Die Befunde des Auffrischens fallen dabei weg**, und das ist Absicht: sie
// gehören auf die Paketseite, nicht in eine Auswahlliste. Wer wissen will, was
// mit seinen Paketen ist, sieht dort nach.
func vorgabenverwaltungUeber(db *sql.DB) (*service.Vorgabenverwaltung, error) {
	paketweg, fehler := PaketnahtUeber(db)
	if fehler != nil {
		return nil, fehler
	}

	// **Formate und Decks kommen aus der Ablage und nicht aus ihrem Dienst.**
	//
	// **Weil hier nichts entschieden wird.** `Formatpflege.Formate` und
	// `Deckbau.Decks` reichen die Abfrage unverändert durch; ein Dienst
	// dazwischen wäre eine Naht ohne Aufgabe — und beim Deckbau eine, die
	// `MitKartennamen` und `MitDrucknamen` verlangt, die eine Auswahlliste
	// nicht braucht. Der Wächter über die Verdrahtung
	// (`internal/werkzeuge/nahtverdrahtung_test.go`) hat genau das
	// angestrichen.
	//
	// **Die Lagerorte kommen dagegen aus ihrem Dienst**: sie werden
	// zusammengezählt, und das ist eine Entscheidung.
	formate := storage.NeueFormatablage(db)
	decks := storage.NeueDeckablage(db)
	sichten := service.NeueSammlungssichten(storage.NeuerSammlungsspeicher(db)).
		MitReservierungen(storage.NeueReservierungsablage(db))

	angebot := NeuesVorgabenangebot(Vorgabenquellen{
		Formate:   formate.Formate,
		Decks:     decks.Decks,
		Pakete:    paketweg.Pakete,
		Lagerorte: sichten.Lagerorte,
	})

	return service.NeueVorgabenverwaltung(
		storage.NeueVorgabenablage(db), angebot), nil
}
