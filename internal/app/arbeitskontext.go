// **Der Arbeitskontext: wer, wo, als was — in einem Zug** — Werkstück O1 von
// Phase 5.2 (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Warum eine Bindung und nicht drei
//
// **Weil drei Rufe drei Stände sind.** `WerArbeitet`, `GezeigtesProfil` und
// die Rolle dazwischen — wer sie nacheinander holt, kann zwischen dem ersten
// und dem dritten einen Wechsel erwischen und zeigt dann den Menschen von
// vorhin im Profil von jetzt. Der Akteur entsteht im Dienst aus **einem**
// Kontext, und die Revision sagt dem Fenster, ob er noch gilt (Werkstück U4).
//
// **Keine fachliche Entscheidung hier** (§2): übersetzt wird zwischen Wails
// und dem Dienst.

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ArbeitskontextAntwort sagt, wer gerade in welchem Profil als was arbeitet.
type ArbeitskontextAntwort struct {
	NutzerID string `json:"nutzerID"`
	ProfilID string `json:"profilID"`

	// EignerID ist der Mensch, dem das Profil gehört — derselbe wie
	// `NutzerID`, solange niemand in ein fremdes gewechselt ist.
	EignerID string `json:"eignerID"`

	// Rolle ist die wirksame Rolle in diesem Profil: `verwalter` für den
	// Eigner, sonst die Freigabe.
	Rolle string `json:"rolle"`

	// NurLesend heißt: hier geht nichts, was schreibt — weil die Rolle
	// `leser` ist **oder** ein Token nur Lesen erlaubt.
	NurLesend bool `json:"nurLesend"`

	// Modus ist die Fläche, über die dieser Ruf kam: `fenster`,
	// `kommandozeile`, `mcp` oder `dienst`.
	Modus string `json:"modus"`

	// Revision steigt mit jedem Wechsel und jeder Freigabe. Zwei Antworten
	// mit derselben Revision beschreiben denselben Stand.
	Revision int64 `json:"revision"`

	Fehler *Fehler `json:"fehler"`
}

// MitAutorisierung setzt den Dienst, der den Akteur baut.
func (a *App) MitAutorisierung(autorisierung *service.Autorisierung) {
	a.autorisierung = autorisierung
}

// Arbeitskontext sagt, wer gerade in welchem Profil als was arbeitet.
//
// **Vor dem Start gibt es keinen** — die Bindung antwortet mit einem Fehler
// und nicht mit einem leeren Menschen, denn ein Fenster, das „niemand" für
// einen Zustand hielte, zeigte ihn an.
func (a *App) Arbeitskontext(ctx context.Context) ArbeitskontextAntwort {
	ctx = a.imProfil(ctx)

	if a.autorisierung == nil {
		return ArbeitskontextAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Anwendung sagt noch nicht, wer arbeitet", domain.ErrNotFound))}
	}

	akteur, fehler := a.autorisierung.Akteur(ctx)
	if fehler != nil {
		return ArbeitskontextAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return ArbeitskontextAntwort{
		NutzerID:  strconv.FormatInt(akteur.NutzerID, 10),
		ProfilID:  strconv.FormatInt(akteur.ProfilID, 10),
		EignerID:  strconv.FormatInt(akteur.EignerID, 10),
		Rolle:     string(akteur.Rolle),
		NurLesend: akteur.Darf(domain.WirkungSchreiben) != nil,
		Modus:     string(akteur.Modus),
		Revision:  akteur.Revision,
	}
}

// Arbeitskontext sagt es dem Fenster.
func (b *Bindungen) Arbeitskontext(ctx context.Context) ArbeitskontextAntwort {
	return b.app(ctx).Arbeitskontext(b.ruf(ctx))
}

// Arbeitskontext sagt es auch den Flächen ohne Fenster.
func (f *Nutzerflaeche) Arbeitskontext(ctx context.Context) ArbeitskontextAntwort {
	return f.app.Arbeitskontext(ctx)
}

// MitDienstmodus trägt die Fläche in den Kontext ein — für die drei Wurzeln
// ohne Fenster, die `internal/service` nicht kennen dürfen (§2) und deshalb
// hier rufen.
func MitDienstmodus(ctx context.Context, modus domain.Dienstmodus) context.Context {
	return service.MitDienstmodus(ctx, modus)
}

// DienstmodusAus liest die Fläche wieder heraus — für [App.Meta], damit ein
// Ruf sagen kann, über welche er kam.
func DienstmodusAus(ctx context.Context) domain.Dienstmodus {
	return service.DienstmodusAus(ctx)
}

// darf fragt die Autorisierung — für die wenigen Wege, die das Fenster
// selbst schreibt statt über einen Dienst (die Exportvorlage).
//
// **Ohne Autorisierung wird nicht gefragt**: sie fehlt nur einer Anwendung
// vor dem Start, und die schreibt nichts.
func (a *App) darf(ctx context.Context, wirkung domain.Wirkung) error {
	// **Die eingesetzte Wache schlägt die echte** — für Läufe, die keinen
	// Bestand mit Rollen aufbauen wollen (§3). Ohne diese Naht ließe sich
	// nicht prüfen, **wonach** gefragt wird; ein Lauf sähe nur, dass ein
	// Weg scheitert, und das sieht bei jedem anderen Grund gleich aus.
	if a.wache != nil {
		return a.wache.Pruefe(ctx, wirkung)
	}

	if a.autorisierung == nil {
		return nil
	}

	return a.autorisierung.Pruefe(ctx, wirkung)
}

// Wache ist, was vor einer Wirkung gefragt wird.
//
// **Eine Schnittstelle dort, wo sie gebraucht wird** (§2) — `service` stellt
// [service.Autorisierung] auf, und dass sie hier hineinpasst, ist eine Folge
// der strukturellen Typisierung und keine Kante.
type Wache interface {
	Pruefe(ctx context.Context, wirkung domain.Wirkung) error
}

// MitWache setzt die Wache — für Läufe ohne echte Rollen (§3).
func (a *App) MitWache(wache Wache) *App {
	a.wache = wache

	return a
}
