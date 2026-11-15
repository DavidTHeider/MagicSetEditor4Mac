// **Die Spalten der Kartenliste und ihre Ordnung** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Eine eigene Datei neben `kartenliste.go`**, und der Schnitt folgt einer
// Frage: dort steht, **welche Karten** die Liste führt — drei Mengen für drei
// Zwecke (#891) —, hier, **welche Spalten** sie zeigt und wonach sie ordnet.

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Listenquelle ist, was das Fenster von der Kartenliste braucht.
type Listenquelle interface {
	Aufbau(ctx context.Context, setID int64) (domain.Kartenlistenaufbau, error)
	SpaltenSetzen(ctx context.Context, setID int64, wahl []domain.Spaltenwahl) error
	SortierungSetzen(ctx context.Context, setID int64, ordnung domain.Kartenordnung) error

	// Filtern gibt die Kennungen der Karten, die durch den Schnellfilter
	// kommen — Werkstück U5 von Phase 5.1 (#1416).
	Filtern(ctx context.Context, setID int64, anfrage string) ([]int64, error)
}

// MitKartenliste setzt den Anwendungsfall — für Tests und den Composition Root.
func (a *App) MitKartenliste(quelle Listenquelle) { a.listenbau = quelle }

// ListenspalteDTO ist eine Spalte, wie die Liste sie anbietet.
type ListenspalteDTO struct {
	Feld string `json:"feld"`

	// Aufschrift ist `card list name:` — **leer bleibt leer**, und das Fenster
	// entscheidet, was es dann zeigt. Ein Rückfall auf den Feldnamen gehört
	// dorthin: hier wüsste niemand, ob er übersetzt werden soll.
	Aufschrift  string `json:"aufschrift"`
	Sichtbar    bool   `json:"sichtbar"`
	Ausrichtung string `json:"ausrichtung"`
}

// ListenzeileDTO ist eine Karte, wie die Liste sie zeigt.
type ListenzeileDTO struct {
	KarteID string `json:"karteID"`

	// Nummer ist die Stelle im **Set**, eins-basiert — sie bleibt, wenn
	// jemand anders sortiert.
	Nummer int `json:"nummer"`

	// Werte trägt **jede** angebotene Spalte, auch die unsichtbaren: wer eine
	// Spalte einblendet, soll nicht auf eine neue Antwort warten müssen.
	Werte map[string]string `json:"werte"`

	// Marken sind die Marken dieser Karte, alphabetisch
	// (#1208).
	//
	// **Nie `nil`**: ein `nil`-Slice käme über die Bindung als `null` an, und
	// die Liste prüfte auf zwei Gestalten desselben Nichts.
	Marken []string `json:"marken"`
}

// KartenlisteAntwort ist der Aufbau der Liste.
type KartenlisteAntwort struct {
	Spalten []ListenspalteDTO `json:"spalten"`
	Zeilen  []ListenzeileDTO  `json:"zeilen"`

	// Sortierfeld ist die **wirksame** Ordnung — leer heißt: die Reihenfolge
	// des Sets.
	Sortierfeld string `json:"sortierfeld"`
	Aufsteigend bool   `json:"aufsteigend"`

	// OhneVorlage sagt, dass die Kartenfelder des Spiels nicht zu lesen waren
	// — dann kommen die Spalten aus den Karten, ohne Aufschrift und ohne die
	// Reihenfolge des Pakets.
	OhneVorlage bool `json:"ohneVorlage"`

	Fehler *Fehler `json:"fehler"`
}

// Kartenlistenaufbau gibt Spalten, Zeilen und die wirksame Ordnung.
func (a *App) Kartenlistenaufbau(ctx context.Context, setID string) KartenlisteAntwort {
	ctx = a.imProfil(ctx)

	leer := KartenlisteAntwort{
		Spalten: []ListenspalteDTO{}, Zeilen: []ListenzeileDTO{}, Aufsteigend: true,
	}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.listenbau == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf(
			"%w: die Kartenliste ist nicht eingerichtet", domain.ErrValidation))

		return leer
	}

	aufbau, err := a.listenbau.Aufbau(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return KartenlisteAntwort{
		Spalten:     spaltenDTO(aufbau.Spalten),
		Zeilen:      zeilenDTO(aufbau.Zeilen),
		Sortierfeld: aufbau.Ordnung.Feld,
		Aufsteigend: aufbau.Ordnung.Aufsteigend,
		OhneVorlage: aufbau.OhneVorlage,
	}
}

// ListenspaltenSetzen schreibt die Wahl fort.
//
// **Die ganze Wahl und nicht eine Spalte**: „welche Spalten sehe ich" ist
// **eine** Frage, und eine Antwort je Spalte ließe zwei Läufe nebeneinander
// gegen einen Stand gehen, den der Bestand noch nicht kennt (§7, Regel 4).
func (a *App) ListenspaltenSetzen(
	ctx context.Context, setID string, wahl []ListenspalteDTO,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.listenbau == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Kartenliste ist nicht eingerichtet", domain.ErrValidation))}
	}

	gewaehlt := make([]domain.Spaltenwahl, 0, len(wahl))

	// **Die Stelle kommt aus der Reihenfolge der Liste** und nicht aus einer
	// Zahl im DTO: die Oberfläche ordnet, indem sie die Liste umstellt, und
	// eine zweite Angabe daneben liefe damit auseinander.
	for stelle, eine := range wahl {
		gewaehlt = append(gewaehlt, domain.Spaltenwahl{
			Feld: eine.Feld, Sichtbar: eine.Sichtbar, Stelle: stelle + 1,
		})
	}

	return SchreibAntwort{
		Fehler: Uebersetze(ctx, a.listenbau.SpaltenSetzen(ctx, kennung, gewaehlt)),
	}
}

// ListensortierungSetzen schreibt sie fort — leeres Feld heißt: gar nicht.
func (a *App) ListensortierungSetzen(
	ctx context.Context, setID, feld string, aufsteigend bool,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.listenbau == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Kartenliste ist nicht eingerichtet", domain.ErrValidation))}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.listenbau.SortierungSetzen(
		ctx, kennung, domain.Kartenordnung{Feld: feld, Aufsteigend: aufsteigend}))}
}

// spaltenDTO übersetzt die Spalten — **immer eine Liste**, nie `nil`.
func spaltenDTO(spalten []domain.Listenspalte) []ListenspalteDTO {
	umgesetzt := make([]ListenspalteDTO, 0, len(spalten))

	for _, spalte := range spalten {
		umgesetzt = append(umgesetzt, ListenspalteDTO{
			Feld:        spalte.Feld,
			Aufschrift:  spalte.Aufschrift,
			Sichtbar:    spalte.Sichtbar,
			Ausrichtung: string(spalte.Ausrichtung),
		})
	}

	return umgesetzt
}

// zeilenDTO übersetzt die Zeilen — ebenso.
func zeilenDTO(zeilen []domain.Listenzeile) []ListenzeileDTO {
	umgesetzt := make([]ListenzeileDTO, 0, len(zeilen))

	for _, zeile := range zeilen {
		// **Eine leere Liste und kein `nil`** (#1208): über die Bindung käme
		// `null` an, und die Liste prüfte auf zwei Gestalten desselben Nichts.
		marken := zeile.Marken
		if marken == nil {
			marken = []string{}
		}

		umgesetzt = append(umgesetzt, ListenzeileDTO{
			KarteID: strconv.FormatInt(zeile.KarteID, 10),
			// **Eins-basiert**, wie die Liste zählt (#666).
			Nummer: zeile.Nummer + 1,
			Werte:  zeile.Werte,
			Marken: marken,
		})
	}

	return umgesetzt
}
