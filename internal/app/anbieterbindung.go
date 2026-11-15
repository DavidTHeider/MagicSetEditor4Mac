package app

// **Anbieterstand und Fremdpreisverlauf am Fenster** (A5, #1444).
//
// # Zwei Bindungen, und beide sagen, wie alt ihre Auskunft ist
//
// Der **Kontostand** kommt vom Anbieter und kann eine halbe Stunde alt sein —
// „Usage info is updated every 20-30 minutes", nachgelesen am 2026-09-11
// ([0096](../../docs/entscheidungen/0096-welche-optionalen-datenquellen-zugaenglich-sind.md)).
// Eine Zahl ohne diesen Satz sähe aus wie eine Sperre, und sie ist keine.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// AnbieterstandDTO ist der Kontostand bei einem Anbieter.
type AnbieterstandDTO struct {
	Dienst string `json:"dienst"`

	// Gesamt, Rest und Verbraucht sind Kredite.
	Gesamt     int64 `json:"gesamt"`
	Rest       int64 `json:"rest"`
	Verbraucht int64 `json:"verbraucht"`

	// Ueberzugssatz ist als **Text** da — es ist Geld je Kredit, und eine
	// Gleitkommazahl im JSON verlöre die letzten Stellen (0093).
	Ueberzugssatz string `json:"ueberzugssatz"`

	// HoechstensAltMinuten sagt, wie alt diese Zahl beim Anbieter sein kann.
	//
	// **Sie steht in jeder Antwort und nicht in der Dokumentation**: wer den
	// Rest liest und die Frist nicht, hält ihn für eine Sperre.
	HoechstensAltMinuten int `json:"hoechstensAltMinuten"`
}

// Anbieterstandantwort trägt ihn samt Fehler.
type Anbieterstandantwort struct {
	AnbieterstandDTO

	Fehler *Fehler `json:"fehler"`
}

// FremdpreisbilanzDTO ist, was ein Verlaufsbezug ergeben hat.
type FremdpreisbilanzDTO struct {
	Quelle string `json:"quelle"`

	// Fassung ist die Kennung der Karte beim Anbieter.
	Fassung string `json:"fassung"`

	// Zugeordnet sagt, ob sie gerade erst gefunden wurde.
	Zugeordnet bool `json:"zugeordnet"`

	// Staende ist die Zahl der abgelegten Preisstände.
	Staende int `json:"staende"`

	Befunde []string `json:"befunde"`
}

// Fremdpreisantwort trägt sie samt Fehler.
type Fremdpreisantwort struct {
	FremdpreisbilanzDTO

	Fehler *Fehler `json:"fehler"`
}

// Anbieterstand fragt, wie viel Kontingent an einem Anbieterkonto noch hängt.
//
// **Heute kann das genau einer** (0096): Scrydex rechnet in Krediten und nennt
// sie unter `/account/v1/usage`. Der Dienst steht trotzdem im Aufruf — eine
// Vorgabe wäre eine Wahl, die aussieht wie keine.
func (a *App) Anbieterstand(ctx context.Context, dienst string) Anbieterstandantwort {
	ctx = a.imProfil(ctx)

	if genannt := dienst; genannt != string(domain.DienstScrydex) {
		return Anbieterstandantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Dienst %q nennt keinen Kontostand — es gibt %q",
			domain.ErrValidation, genannt, string(domain.DienstScrydex)))}
	}

	verbindung := a.laufendeVerbindung()
	if verbindung == nil {
		return Anbieterstandantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es ist kein Bestand offen", domain.ErrValidation))}
	}

	wahl := NeueEinstellungswahl(storage.NeueEinstellungen(verbindung))

	quelle, fehler := scrydexpreiseUeber(ctx, wahl, a.bundFuerDenMarkt())
	if fehler != nil {
		return Anbieterstandantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	stand, standfehler := quelle.klient.Kontostand(ctx)
	if standfehler != nil {
		return Anbieterstandantwort{Fehler: Uebersetze(ctx, standfehler)}
	}

	return Anbieterstandantwort{AnbieterstandDTO: AnbieterstandDTO{
		Dienst: dienst, Gesamt: stand.Gesamt, Rest: stand.Rest,
		Verbraucht: stand.Verbraucht, Ueberzugssatz: stand.Ueberzugssatz,
		HoechstensAltMinuten: int(scrydex.Aktualisierungsfrist.Minutes()),
	}}
}

// fremdpreisAbgelehnt ist die Antwort auf eine Verlaufsfrage, die nicht
// durchging — **mit einer leeren Befundliste und nicht `null`** (C01,
// #2226): das Fenster erwartet eine Liste, auch neben einem Fehler.
func fremdpreisAbgelehnt(ctx context.Context, fehler error) Fremdpreisantwort {
	return Fremdpreisantwort{
		FremdpreisbilanzDTO: FremdpreisbilanzDTO{Befunde: []string{}},
		Fehler:              Uebersetze(ctx, fehler),
	}
}

// Fremdpreisverlauf holt den Verlauf eines Drucks bei einer Quelle.
func (a *App) Fremdpreisverlauf(
	ctx context.Context, quelle, printing, finish string, zustaende []string,
) Fremdpreisantwort {
	ctx = a.imProfil(ctx)

	if genannt := quelle; genannt != string(scrydex.Familie) {
		return fremdpreisAbgelehnt(ctx, fmt.Errorf(
			"%w: die Verlaufsquelle %q kennt diese Anwendung nicht — es gibt %q",
			domain.ErrValidation, genannt, string(scrydex.Familie)))
	}

	verbindung := a.laufendeVerbindung()
	if verbindung == nil {
		return fremdpreisAbgelehnt(ctx, fmt.Errorf(
			"%w: es ist kein Bestand offen", domain.ErrValidation))
	}

	wunsch := domain.Angebotswunsch{
		Finish: domain.Finish(finish), Zustaende: zustaende,
	}

	if wunsch.Finish != "" && !domain.IstFinish(wunsch.Finish) {
		return fremdpreisAbgelehnt(ctx, fmt.Errorf(
			"%w: die Ausführung %q gibt es nicht", domain.ErrValidation, finish))
	}

	wahl := NeueEinstellungswahl(storage.NeueEinstellungen(verbindung))

	preisquelle, fehler := scrydexpreiseUeber(ctx, wahl, a.bundFuerDenMarkt())
	if fehler != nil {
		return fremdpreisAbgelehnt(ctx, fehler)
	}

	bilanz, bezugsfehler := service.NeuerFremdpreisbezug(
		storage.NeueDrucke(verbindung), preisquelle,
		storage.NeuerPreisspeicher(verbindung)).Holen(ctx, printing, wunsch)
	if bezugsfehler != nil {
		return fremdpreisAbgelehnt(ctx, bezugsfehler)
	}

	return Fremdpreisantwort{FremdpreisbilanzDTO: zuFremdpreisbilanz(quelle, bilanz)}
}

// zuFremdpreisbilanz macht aus der Bilanz ihr DTO.
func zuFremdpreisbilanz(quelle string, bilanz service.Fremdpreisbilanz) FremdpreisbilanzDTO {
	befunde := make([]string, 0, len(bilanz.Befunde))
	for _, befund := range bilanz.Befunde {
		befunde = append(befunde, befund.Meldung)
	}

	return FremdpreisbilanzDTO{
		Quelle: quelle, Fassung: bilanz.Fassung, Zugeordnet: bilanz.Zugeordnet,
		Staende: bilanz.Staende, Befunde: befunde,
	}
}

// AnbieterAnwendungUeber baut eine Anwendung, die beides kann.
func AnbieterAnwendungUeber(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app
	anwendung.db = db

	return anwendung
}
