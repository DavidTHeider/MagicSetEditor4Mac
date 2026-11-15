package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bildquelle ist, was die Fassade vom Bildweg braucht.
// **Sie nimmt die Art als Zeichenkette**, obwohl `scryfall` einen eigenen Typ
// dafür hat. Die Fassade ist die Grenze zu Wails, und dort kommen Wörter an;
// die Übersetzung steht an **einer** Stelle, im Composition Root — und damit
// bleibt `internal/scryfall` aus dieser Datei heraus (§2).
type Bildquelle interface {
	Holen(ctx context.Context, name, art string) (*service.Bilduebernahme, error)
}

// ScryfallBildAntwort ist ein geholtes Bild samt Zuschnittvorschlag.
//
// **Sie erweitert [ZuschnittAntwort] um die Nennung**, statt sie danebenzu-
// stellen: wer das Bild hat, hat damit auch, was neben ihm stehen muss.
type ScryfallBildAntwort struct {
	Pfad    string     `json:"pfad"`
	Auswahl AuswahlDTO `json:"auswahl"`

	Kuenstler string `json:"kuenstler"`

	// Nennung ist der Satz für dieselbe Oberfläche — nachgelesen am
	// 2026-08-23 unter `https://scryfall.com/docs/api`: „list the artist name
	// and copyright elsewhere in the same interface presenting the art crop".
	Nennung string `json:"nennung"`

	// Adresse ist die Kartenseite bei Scryfall — die „source" der Bedingung.
	Adresse string `json:"adresse"`

	// AlterStunden sagt, wie alt der Stand ist (§16, Regel 5).
	AlterStunden float64 `json:"alterStunden"`

	// Befunde ist, was die Kartenauskunft tolerant hingenommen hat.
	//
	// **Sie kommen mit** (§16: „nichts wird still übernommen"). Der Weg
	// dahinter warf sie weg; der Grund steht bei [service.Bilduebernahme].
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// MitScryfallbildern setzt die Bildquelle — die Naht für Tests.
func (a *App) MitScryfallbildern(quelle Bildquelle) { a.scryfallbilder = quelle }

// ScryfallBild holt das Bild einer Karte und bereitet den Zuschnitt vor.
//
// **Es schreibt nichts.** Wie `BildAusDatei` endet es beim Vorschlag; erst
// [App.ScryfallBildEinsetzen] legt ab. Das ist §16, Regel 6: „Nichts wird
// still übernommen."
func (a *App) ScryfallBild(
	ctx context.Context, karteID, feld, name, art string,
) ScryfallBildAntwort {
	ctx = a.imProfil(ctx)

	// **Auch ein Nein trägt die leere Liste.** Der Typ sagt `[]string` zu,
	// und ein `null` im JSON ist für das Fenster keine leere Liste, sondern
	// gar keine — angestrichen vom Wächter über die Bindungen.
	nein := func(fehler *Fehler) ScryfallBildAntwort {
		return ScryfallBildAntwort{Befunde: []string{}, Fehler: fehler}
	}

	if a.scryfallbilder == nil {
		return nein(Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation)))
	}

	_, _, breite, hoehe, befund := a.bildziel(ctx, karteID, feld)
	if befund != nil {
		return nein(befund)
	}

	holung, err := a.scryfallbilder.Holen(ctx, name, art)
	if err != nil {
		return nein(Uebersetze(ctx, err))
	}

	vorbereitet := a.vorbereiten(ctx, holung.Pfad, breite, hoehe)
	if vorbereitet.Fehler != nil {
		return nein(vorbereitet.Fehler)
	}

	return ScryfallBildAntwort{
		Pfad: vorbereitet.Pfad, Auswahl: vorbereitet.Auswahl,
		Kuenstler: holung.Kuenstler, Nennung: holung.Nennung,
		Adresse: holung.Adresse, AlterStunden: holung.Alter.Hours(),
		Befunde: befundtexte(holung.Befunde),
	}
}

// ScryfallBildEinsetzen legt das Bild ab **und schreibt den Künstler**.
//
// # Warum beides in einem Aufruf
//
// Die Bedingung von Scryfall verlangt, dass Künstler und Copyright in
// derselben Oberfläche stehen wie das Bild — und die `.mse-set`-Ausfuhr ist
// auch eine. Wären es zwei Aufrufe, gäbe es einen Zustand dazwischen, in dem
// das Bild an der Karte hängt und der Name nicht: genau den, den die Bedingung
// ausschließt.
//
// **Ein leerer Künstler schreibt nichts.** Ein `illustrator: ""` löschte einen
// vorhandenen Namen — eine Nennungspflicht, die eine Nennung entfernt, wäre
// verkehrt herum.
func (a *App) ScryfallBildEinsetzen(
	ctx context.Context, setID, karteID, feld, pfad, kuenstler, herkunft string,
	auswahl AuswahlDTO,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	// **Die Kennungen kommen aus dem Einsetzen zurück und werden nicht noch
	// einmal gelesen.** Der erste Anlauf las sie hier erneut — und die beiden
	// Fehlerzweige darauf konnte kein Aufruf betreten, weil `BildEinsetzen`
	// dieselbe Prüfung schon gemacht hat. Ein Zweig ohne Weg dorthin ist kein
	// Fehlerweg, sondern eine unbelegte Zeile (§3).
	// **Ein Zug und nicht zwei** (#1163, Befund Z7-02): zwischen zwei Aufrufen
	// trüge die Karte fremde Kunst ohne ihre Nennung — der Stand, den die
	// Bedingung von Scryfall ausschließt und für den es diesen ganzen Weg gibt.
	//
	// **§7, Regel 3 gilt hier nicht.** „Geschrieben ist geschrieben" spricht
	// vom **Nachladen** nach einer geglückten Änderung, nicht davon, eine
	// Änderung in zwei zu zerlegen, die einzeln falsch sind.
	//
	// **Ein leerer Künstler schreibt nichts.** Ein `illustrator: ""` löschte
	// einen vorhandenen Namen — eine Nennungspflicht, die eine Nennung
	// entfernt, wäre verkehrt herum.
	var weitere []domain.Value
	if kuenstler != "" {
		weitere = append(weitere, domain.Value{Field: feldIllustrator, Text: kuenstler})
	}

	return a.bildAblegen(ctx, setID, karteID, feld, pfad, auswahl, herkunft, weitere...)
}

// feldIllustrator ist das Kartenfeld, in dem MSE den Künstler führt.
//
// **Derselbe Name wie in der Zuordnung** (`internal/scryfall/zuordnung.go`):
// jede M15-Vorlage nennt es `illustrator`, und MSE zeichnet es auf die Karte.
const feldIllustrator = "illustrator"
