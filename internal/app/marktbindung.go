package app

// **Die Marktangebote am Fenster** (A4, #1443).
//
// # Eine Bindung, vier Flächen
//
// Dieselbe Fähigkeit steht an Kommandozeile, REST und MCP (§16). Was hier
// entsteht, ist die DTO-Gestalt und die Übersetzung der Fehler — kein `if`
// über einen Domänenzustand (§2).

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// MarktangebotDTO ist ein Posten, wie das Fenster ihn sieht.
type MarktangebotDTO struct {
	Quelle string `json:"quelle"`

	// Kennung ist die Nummer dieses Postens beim Anbieter — als Zeichenkette
	// (0086: keine int64 ins JavaScript).
	Kennung string `json:"kennung"`

	// Wert ist der Preis in der Form, in der man ihn liest — `"2.17"`.
	Wert     string `json:"wert"`
	Waehrung string `json:"waehrung"`

	Menge   int `json:"menge"`
	Buendel int `json:"buendel"`

	Finish  string `json:"finish"`
	Zustand string `json:"zustand"`
	Sprache string `json:"sprache"`

	Signiert   bool `json:"signiert"`
	Veraendert bool `json:"veraendert"`
	Gradiert   bool `json:"gradiert"`
	ImUrlaub   bool `json:"imUrlaub"`

	Verkaeufer string `json:"verkaeufer"`
	Land       string `json:"land"`
}

// MarktantwortDTO ist, was ein Abruf ergeben hat.
type MarktantwortDTO struct {
	Quelle string `json:"quelle"`

	// Fassung ist die Nummer der Kartenfassung beim Anbieter.
	Fassung string `json:"fassung"`

	// Zugeordnet sagt, ob die Fassung gerade erst gefunden wurde.
	Zugeordnet bool `json:"zugeordnet"`

	Angebote []MarktangebotDTO `json:"angebote"`
	Befunde  []string          `json:"befunde"`
}

// Marktantwort trägt sie samt Fehler.
type Marktantwort struct {
	MarktantwortDTO

	Fehler *Fehler `json:"fehler"`
}

// Marktangebote gibt die Posten eines Marktplatzes zu einem Druck.
//
// **Die Quelle wird genannt und nicht erraten** (A3, A4): es gibt heute eine,
// und trotzdem steht sie im Aufruf.
func (a *App) Marktangebote(
	ctx context.Context, quelle, printing, sprache, finish string, zustaende []string,
) Marktantwort {
	ctx = a.imProfil(ctx)

	// **Der Name des Marktplatzes gehört in die Meldung**, und er ist kein
	// Pfad: er kommt aus dem Aufruf und steht in `FAEHIGKEITEN.md`. Ohne ihn
	// hieße es „den Marktplatz kennt diese Anwendung nicht", und wer
	// `cardmarket` getippt hat, erführe nicht, dass es an seinem Wort lag.
	if genannt := quelle; genannt != string(cardtrader.Familie) {
		return Marktantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: den Marktplatz %q kennt diese Anwendung nicht — es gibt %q",
			domain.ErrValidation, genannt, string(cardtrader.Familie)))}
	}

	verbindung := a.laufendeVerbindung()
	if verbindung == nil {
		return Marktantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es ist kein Bestand offen", domain.ErrValidation))}
	}

	wunsch := domain.Angebotswunsch{
		Sprache: sprache, Finish: domain.Finish(finish), Zustaende: zustaende,
	}

	if wunsch.Finish != "" && !domain.IstFinish(wunsch.Finish) {
		return Marktantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Ausführung %q gibt es nicht", domain.ErrValidation, finish))}
	}

	wahl := NeueEinstellungswahl(storage.NeueEinstellungen(verbindung))

	markt, fehler := cardtradermarktUeber(ctx, wahl, a.bundFuerDenMarkt())
	if fehler != nil {
		return Marktantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	ergebnis, abruffehler := service.NeuerMarktabruf(
		storage.NeueDrucke(verbindung), markt).Angebote(ctx, printing, wunsch)
	if abruffehler != nil {
		return Marktantwort{Fehler: Uebersetze(ctx, abruffehler)}
	}

	return Marktantwort{MarktantwortDTO: zuMarktantwort(quelle, ergebnis)}
}

// bundFuerDenMarkt gibt den Schlüsselbund — den eingesetzten oder den echten.
func (a *App) bundFuerDenMarkt() Geheimnisleser {
	if a.marktbund != nil {
		return a.marktbund
	}

	return schluesselbund.Anmeldebund()
}

// MitMarktbund setzt den Schlüsselbund — für Läufe ohne echten (§3).
func (a *App) MitMarktbund(bund Geheimnisleser) *App {
	a.marktbund = bund

	return a
}

// zuMarktantwort macht aus dem Ergebnis sein DTO.
func zuMarktantwort(quelle string, ergebnis service.Marktergebnis) MarktantwortDTO {
	angebote := make([]MarktangebotDTO, 0, len(ergebnis.Angebote))
	for _, angebot := range ergebnis.Angebote {
		angebote = append(angebote, MarktangebotDTO{
			Quelle:  string(angebot.Quelle),
			Kennung: angebot.Fremdkennung,
			// **Der Preis geht als Text hinaus** — dieselbe Form wie am
			// Preisstand (0093): eine Gleitkommazahl im JSON verlöre Cent.
			//
			// **Mit den Stellen seiner Währung** (LOC-03.AK-02, #1689):
			// ein Angebot in Yen stand als »1.50« da, wo 150 Yen gemeint
			// waren.
			Wert:       preistext(angebot.Preis, angebot.Waehrung),
			Waehrung:   angebot.Waehrung,
			Menge:      angebot.Menge,
			Buendel:    angebot.Buendel,
			Finish:     string(angebot.Finish),
			Zustand:    angebot.Zustand,
			Sprache:    angebot.Sprache,
			Signiert:   angebot.Signiert,
			Veraendert: angebot.Veraendert,
			Gradiert:   angebot.Gradiert,
			ImUrlaub:   angebot.ImUrlaub,
			Verkaeufer: angebot.Verkaeufer,
			Land:       angebot.Land,
		})
	}

	befunde := make([]string, 0, len(ergebnis.Befunde))
	for _, befund := range ergebnis.Befunde {
		befunde = append(befunde, befund.Meldung)
	}

	return MarktantwortDTO{
		Quelle: quelle, Fassung: ergebnis.Fassung, Zugeordnet: ergebnis.Zugeordnet,
		Angebote: angebote, Befunde: befunde,
	}
}

// MarktAnwendungUeber baut eine Anwendung, die den Marktplatz kann.
func MarktAnwendungUeber(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app
	anwendung.db = db

	return anwendung
}
