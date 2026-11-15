// **Pakete einbinden, entfernen und aus MSE2 übernehmen** —
// #1291.
//
// **Eine eigene Datei neben `pakete.go`**, weil sie eine andere Frage
// beantwortet: dort steht, **was da ist** und wie man es holt; hier, wie man
// etwas einbindet, das schon daliegt.
//
// # Warum es ohne Fenster geht
//
// Der Verzeichnisdialog steckt allein in den Bindungen.
// `App.PaketInstallieren` und `App.VorlagenUebernehmen` nehmen einen Pfad
// entgegen — und wer einen Rechner aus einem Skript einrichtet, hat keinen
// Dialog, aber sehr wohl einen Pfad.

package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einbindungAntwort sagt, was eingebunden wurde.
type einbindungAntwort struct {
	Paket paketDTO `json:"paket"`
}

// orteAntwort nennt die Vorschläge für eine MSE2-Installation.
type orteAntwort struct {
	Orte []string `json:"orte"`
}

// uebernahmeAntwort sagt, was eine Übernahme ergab.
//
// **Alle drei Listen ohne `omitempty`** (§3): „nichts übergangen" ist eine
// Aussage, `null` ist keine.
type uebernahmeAntwort struct {
	Uebernommen []paketDTO `json:"uebernommen"`
	Uebergangen []string   `json:"uebergangen"`
	Befunde     []string   `json:"befunde"`
}

// paketInstallieren bindet ein Paket von der Platte ein.
func paketInstallieren(ctx context.Context, db *sql.DB, an *schreiber, quelle string) int {
	weg, err := app.PaketnahtUeber(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	paket, fehler := weg.Installieren(ctx, quelle)
	if fehler != nil {
		an.klageUeber(fehler)

		return paketCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, einbindungAntwort{Paket: alsPaketDTO(paket)})
	}

	an.sag("%s\t%s\t%s\t%s\n", paket.Art, paket.Name, paket.Version, paket.Pfad)

	return an.abschluss()
}

// paketAnlegen schreibt ein neues Paket aus einem Gerüst (#1350).
//
// **Dieselbe Ausgabe wie `paket-installieren`** — eine Zeile aus Art, Name,
// Fassung und Pfad. Ein Paket, das anders angelegt wurde, ist danach kein
// anderes Paket.
func paketAnlegen(
	ctx context.Context, db *sql.DB, an *schreiber, art, name, vollname, spiel string,
) int {
	weg, err := app.PaketnahtUeber(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	paket, fehler := weg.Anlegen(ctx, art, name, vollname, spiel)
	if fehler != nil {
		an.klageUeber(fehler)

		return paketCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, einbindungAntwort{Paket: alsPaketDTO(paket)})
	}

	an.sag("%s\t%s\t%s\t%s\n", paket.Art, paket.Name, paket.Version, paket.Pfad)

	return an.abschluss()
}

// paketEntfernen nimmt ein Paket wieder heraus.
func paketEntfernen(ctx context.Context, db *sql.DB, an *schreiber, art, name string) int {
	weg, err := app.PaketnahtUeber(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if fehler := weg.Entfernen(ctx, art, name, ""); fehler != nil {
		an.klageUeber(fehler)

		return paketCode(fehler)
	}

	an.sag("%s %s entfernt.\n", art, name)

	return an.abschluss()
}

// vorlagenUebernehmen holt die Vorlagen aus einem Verzeichnis.
func vorlagenUebernehmen(ctx context.Context, db *sql.DB, an *schreiber, ort string) int {
	weg, err := app.PaketnahtUeber(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	uebernahme, fehler := weg.Uebernehmen(ctx, ort)

	// **Erst das Ergebnis, dann der Fehler.** Ein Auffrischen, das scheitert,
	// nimmt die Übernahme nicht zurück — geschrieben ist geschrieben (§7).
	if uebernahme != nil {
		uebernahmeZeigen(an, uebernahme)
	}

	if fehler != nil {
		an.klageUeber(fehler)

		return paketCode(fehler)
	}

	return an.abschluss()
}

// uebernahmeZeigen schreibt, was übernommen wurde.
func uebernahmeZeigen(an *schreiber, uebernahme *app.Paketuebernahme) {
	antwort := uebernahmeAntwort{
		Uebernommen: make([]paketDTO, 0, len(uebernahme.Uebernommen)),
		Uebergangen: make([]string, 0, len(uebernahme.Uebergangen)),
		Befunde:     make([]string, 0, len(uebernahme.Befunde)),
	}

	for _, paket := range uebernahme.Uebernommen {
		antwort.Uebernommen = append(antwort.Uebernommen, alsPaketDTO(paket))
	}

	antwort.Uebergangen = append(antwort.Uebergangen, uebernahme.Uebergangen...)

	antwort.Befunde = append(antwort.Befunde, uebernahme.Befunde...)

	if an.alsJSON {
		alsJSONSchreiben(an, antwort)

		return
	}

	for _, paket := range antwort.Uebernommen {
		an.sag("%s\t%s\t%s\n", paket.Art, paket.Name, paket.Version)
	}

	for _, uebergangen := range antwort.Uebergangen {
		an.sag("  übergangen: %s\n", uebergangen)
	}

	for _, befund := range antwort.Befunde {
		an.sag("  %s\n", befund)
	}
}

// mse2Orte nennt die üblichen Orte einer MSE2-Installation.
func mse2Orte(an *schreiber) int {
	orte := app.MSE2Orte()
	if orte == nil {
		orte = []string{}
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, orteAntwort{Orte: orte})
	}

	for _, ort := range orte {
		an.sag("%s\n", ort)
	}

	return an.abschluss()
}

// paketCode ordnet einen Fehler einem Rückgabewert zu.
func paketCode(fehler error) int {
	if errors.Is(fehler, domain.ErrValidation) {
		return CodeBedienung
	}

	return CodeFachlich
}
