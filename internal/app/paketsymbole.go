// **Die Symbole eines Pakets am Fenster** — Werkstück U7 von Phase 5.1
// (#1418).
//
// Zwei Fragen der Werkstatt, zwei Bindungen: **was definiert dieses Paket**
// (für Vervollständigung, Hover und den Sprung zur Definition) und **wer
// nennt diesen Namen** (die Referenzen). Beide lesen die Dateien des Pakets
// über dieselbe Naht wie der Editor; der Index selbst steht in `service`,
// weil er die beiden Zerteiler braucht und das Fenster keine zweite Lesart
// des Formats bekommen soll.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PaketsymbolDTO ist ein definierter Name mit seiner Stelle.
type PaketsymbolDTO struct {
	Name   string `json:"name"`
	Art    string `json:"art"`
	Datei  string `json:"datei"`
	Zeile  int    `json:"zeile"`
	Spalte int    `json:"spalte"`
}

// PaketsymboleAntwort trägt die Symbole eines Pakets.
type PaketsymboleAntwort struct {
	Symbole []PaketsymbolDTO `json:"symbole"`
	Fehler  *Fehler          `json:"fehler"`
}

// SymbolverwendungDTO ist eine Zeile, die einen Namen nennt.
type SymbolverwendungDTO struct {
	Datei string `json:"datei"`
	Zeile int    `json:"zeile"`
	Text  string `json:"text"`
}

// SymbolverwendungenAntwort trägt die Verwendungen eines Namens.
type SymbolverwendungenAntwort struct {
	Verwendungen []SymbolverwendungDTO `json:"verwendungen"`
	Fehler       *Fehler               `json:"fehler"`
}

// Paketsymbole gibt jede Definition in den Dateien eines Pakets.
func (a *App) Paketsymbole(ctx context.Context, paket string) PaketsymboleAntwort {
	ctx = a.imProfil(ctx)

	texte, fehler := a.pakettexte(ctx, paket)
	if fehler != nil {
		return PaketsymboleAntwort{Symbole: []PaketsymbolDTO{}, Fehler: fehler}
	}

	symbole := service.Paketsymbole(texte)
	dtos := make([]PaketsymbolDTO, 0, len(symbole))

	for _, symbol := range symbole {
		dtos = append(dtos, PaketsymbolDTO{
			Name: symbol.Name, Art: symbol.Art, Datei: symbol.Datei,
			Zeile: symbol.Zeile, Spalte: symbol.Spalte,
		})
	}

	return PaketsymboleAntwort{Symbole: dtos}
}

// Symbolverwendungen gibt jede Zeile eines Pakets, die den Namen nennt.
func (a *App) Symbolverwendungen(ctx context.Context, paket, name string) SymbolverwendungenAntwort {
	ctx = a.imProfil(ctx)

	texte, fehler := a.pakettexte(ctx, paket)
	if fehler != nil {
		return SymbolverwendungenAntwort{Verwendungen: []SymbolverwendungDTO{}, Fehler: fehler}
	}

	verwendungen := service.Symbolverwendungen(texte, name)
	dtos := make([]SymbolverwendungDTO, 0, len(verwendungen))

	for _, verwendung := range verwendungen {
		dtos = append(dtos, SymbolverwendungDTO{
			Datei: verwendung.Datei, Zeile: verwendung.Zeile, Text: verwendung.Text,
		})
	}

	return SymbolverwendungenAntwort{Verwendungen: dtos}
}

// pakettexte holt die Dateien samt Inhalt über die Werkstattnaht — mit dem
// Kontext, den der Rufer schon ins Profil gesetzt hat.
func (a *App) pakettexte(ctx context.Context, paket string) ([]service.Paketdateitext, *Fehler) {
	if a.paketdateien == nil {
		return nil, Uebersetze(ctx, keinePaketwerkstatt())
	}

	texte, err := a.paketdateien.Texte(ctx, paket)
	if err != nil {
		return nil, Uebersetze(ctx, err)
	}

	return texte, nil
}
