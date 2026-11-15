package app

// **Die Liste der eigenen Partien** —
// #1600.
//
// # Der Befund, aus dem sie entstanden ist
//
// **`storage.Partieablage.Partien` hatte seit G1 keinen Aufrufer** — gemessen
// am 2026-09-17 über den ganzen Baum: eine Fundstelle, und das war die
// Definition. Das ist die erste Dimension aus §10, und sie kostete etwas
// Sichtbares: der **Wirt** kam in eine laufende Partie nicht zurück. Sie lag
// im Bestand, im Fenster lebte sie nur im Zustand von `usePartie`, und den
// wirft jeder Neustart und jeder Profilwechsel weg.
//
// # Warum eine eigene Bindung und kein Feld an einer anderen Antwort
//
// **Weil sie eine eigene Frage ist.** „Welche Partien habe ich?" wird
// gestellt, **bevor** eine offensteht — und eine Antwort, die an `Partie`
// hinge, bräuchte eine Partie, um zu sagen, dass es eine gibt.

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PartiezeileDTO ist eine Partie, wie eine Liste sie zeigt.
//
// **Ohne Objekte und ohne Sitzkennungen** — der Grund steht bei
// [domain.Partiezeile]: eine Liste ist keine Sicht.
type PartiezeileDTO struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Spiel string `json:"spiel"`
	Stand string `json:"stand"`

	// Sitze sind die Plätze am Tisch — Kennung und Name.
	Sitze []ZeilensitzDTO `json:"sitze"`

	// Objekte ist, wie viele Dinge in dieser Partie liegen.
	Objekte int `json:"objekte"`

	// Zug ist der wievielte Zug — null heißt: es geht nicht zugweise zu.
	Zug int `json:"zug"`

	// EigenerSitz ist die Kennung des Sitzes, der an das gezeigte Profil
	// gebunden ist — **leer heißt: keiner**, das Profil sieht hier zu (UXR-F16,
	// #2031).
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// ein Feld mehr, keines weniger.
	EigenerSitz string `json:"eigenerSitz"`

	// Dran sagt, ob die Partie gerade auf den eigenen Sitz wartet — wegen
	// Priorität oder einer offenen Entscheidung. **Falsch ohne eigenen Sitz,
	// im freien Modus, in einer angehaltenen oder beendeten Partie.** Gerechnet
	// hat es `service` (CR 117.1, Fassung 2026-09-25); hier steht nur das
	// Ergebnis.
	Dran bool `json:"dran"`
}

// ZeilensitzDTO ist ein Platz, wie die Liste ihn zeigt.
type ZeilensitzDTO struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
}

// PartienAntwort ist die Liste samt Fehler.
type PartienAntwort struct {
	Partien []PartiezeileDTO `json:"partien"`
	Fehler  *Fehler          `json:"fehler"`
}

// Partien gibt die Partien dieses Profils, zuletzt geänderte zuerst.
func (a *App) Partien(ctx context.Context) PartienAntwort {
	ctx = a.imProfil(ctx)

	leer := PartienAntwort{Partien: []PartiezeileDTO{}}

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	// **Das gezeigte Profil kommt aus dem Kontext** — warum, steht bei
	// [gezeigtesProfil]. Gefragt wird nach der Erlaubnis und vor dem Dienst:
	// ohne Profil lässt sich nicht sagen, welcher Sitz der eigene ist.
	profilID, fehler := gezeigtesProfil(ctx)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	zeilen, fehler := a.partienaht().Partien(ctx, profilID)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	gezeigt := make([]PartiezeileDTO, 0, len(zeilen))
	for _, zeile := range zeilen {
		gezeigt = append(gezeigt, partiezeileAus(zeile))
	}

	return PartienAntwort{Partien: gezeigt}
}

// partiezeileAus bildet eine Zeile auf ihr DTO ab.
//
// **Die Fassung als Zeichenkette** — wie überall an dieser Naht: JavaScript
// rechnet mit `number` und verlöre bei großen Zahlen die letzten Stellen.
func partiezeileAus(zeile domain.Partiezeile) PartiezeileDTO {
	sitze := make([]ZeilensitzDTO, 0, len(zeile.Sitze))
	for _, sitz := range zeile.Sitze {
		sitze = append(sitze, ZeilensitzDTO{
			Kennung: sitz.Kennung, Name: sitz.Name,
		})
	}

	return PartiezeileDTO{
		ID:      zeile.ID,
		Fassung: strconv.FormatInt(zeile.Fassung, 10),
		Spiel:   zeile.Spiel,
		Stand:   string(zeile.Stand),
		Sitze:   sitze,
		Objekte: zeile.Objekte,
		Zug:     zeile.Zug,

		EigenerSitz: zeile.EigenerSitz,
		Dran:        zeile.Dran,
	}
}
