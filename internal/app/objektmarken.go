package app

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk [0135](../../docs/entscheidungen/0135-marken-an-jedem-fachobjekt.md).
//
// Was eine Marke ist, wem ein Objekt gehört und was ein Filter heißt, steht
// in `service.Objektmarken`. **Hier stehen die Naht für die Flächen ohne
// Fenster und die Bindungen des Fensters** — beide über denselben
// Anwendungsfall, damit „tausch" überall dasselbe meint (§16).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Objektmarkennaht sind die Marken jeder Art, wie eine Fläche sie ruft.
//
// **Die Namen tragen ihren Gegenstand** — `ObjektmarkeSetzen` und nicht
// `Setzen` —, aus demselben Grund wie bei [Fremdmarken].
type Objektmarkennaht struct{ marken *service.Objektmarken }

// ObjektmarkeSetzen hängt eine Marke an ein Objekt.
func (n *Objektmarkennaht) ObjektmarkeSetzen(ctx context.Context, art, kennung, marke string) error {
	return n.marken.Setzen(ctx, art, kennung, marke)
}

// ObjektmarkeEntfernen nimmt sie wieder ab.
func (n *Objektmarkennaht) ObjektmarkeEntfernen(ctx context.Context, art, kennung, marke string) error {
	return n.marken.Entfernen(ctx, art, kennung, marke)
}

// ObjektmarkenAn nennt die Marken eines Objekts.
func (n *Objektmarkennaht) ObjektmarkenAn(ctx context.Context, art, kennung string) ([]string, error) {
	return n.marken.An(ctx, art, kennung)
}

// MarkenJeObjekt nennt die Marken jedes Objekts einer Art.
func (n *Objektmarkennaht) MarkenJeObjekt(ctx context.Context, art string) (map[string][]string, error) {
	return n.marken.JeObjekt(ctx, art)
}

// MarkeVergeben hängt eine Marke an mehrere Objekte — ganz oder gar nicht.
func (n *Objektmarkennaht) MarkeVergeben(ctx context.Context, art string, kennungen []string, marke string) error {
	return n.marken.Vergeben(ctx, art, kennungen, marke)
}

// MarkenFiltern gibt die Objekte einer Art, die durch den Filter gehen.
func (n *Objektmarkennaht) MarkenFiltern(ctx context.Context, art string, marken []string, verknuepfung string) ([]string, error) {
	return n.marken.Filtern(ctx, art, marken, verknuepfung)
}

// Markenkatalog zählt die Marken des Profils.
func (n *Objektmarkennaht) Markenkatalog(ctx context.Context) ([]domain.Markenposten, error) {
	return n.marken.Katalog(ctx)
}

// MarkeUmbenennen gibt einer Marke einen neuen Namen.
func (n *Objektmarkennaht) MarkeUmbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) (int, error) {
	return n.marken.Umbenennen(ctx, alt, neu, zusammenfuehren)
}

// MarkeUeberallEntfernen nimmt eine Marke von jedem Objekt des Profils.
func (n *Objektmarkennaht) MarkeUeberallEntfernen(ctx context.Context, marke string) (int, error) {
	return n.marken.UeberallEntfernen(ctx, marke)
}

// MarkenpostenDTO ist eine Marke mit der Zahl der Objekte, die sie tragen.
type MarkenpostenDTO struct {
	Marke  string `json:"marke"`
	Anzahl int    `json:"anzahl"`
}

// MarkenJeObjektAntwort trägt die Marken je Kennung.
type MarkenJeObjektAntwort struct {
	Marken map[string][]string `json:"marken"`
	Fehler *Fehler             `json:"fehler"`
}

// MarkenfilterAntwort trägt die Kennungen, die durch den Filter gingen.
type MarkenfilterAntwort struct {
	Kennungen []string `json:"kennungen"`
	Fehler    *Fehler  `json:"fehler"`
}

// MarkenkatalogAntwort trägt die Marken des Profils.
type MarkenkatalogAntwort struct {
	Marken []MarkenpostenDTO `json:"marken"`
	Fehler *Fehler           `json:"fehler"`
}

// MarkenzahlAntwort trägt, wie viele Objekte betroffen waren.
type MarkenzahlAntwort struct {
	Anzahl int     `json:"anzahl"`
	Fehler *Fehler `json:"fehler"`
}

// MitObjektmarken setzt die Naht — für Tests und den Composition Root.
func (a *App) MitObjektmarken(naht *Objektmarkennaht) { a.objektmarken = naht }

// ObjektmarkeSetzen hängt eine Marke an ein Objekt.
func (a *App) ObjektmarkeSetzen(ctx context.Context, art, kennung, marke string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.objektmarken.ObjektmarkeSetzen(ctx, art, kennung, marke))}
}

// ObjektmarkeEntfernen nimmt eine Marke von einem Objekt.
func (a *App) ObjektmarkeEntfernen(ctx context.Context, art, kennung, marke string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.objektmarken.ObjektmarkeEntfernen(ctx, art, kennung, marke))}
}

// MarkeVergeben hängt eine Marke an mehrere Objekte.
func (a *App) MarkeVergeben(ctx context.Context, art string, kennungen []string, marke string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.objektmarken.MarkeVergeben(ctx, art, kennungen, marke))}
}

// MarkenJeObjekt nennt die Marken jedes Objekts einer Art.
func (a *App) MarkenJeObjekt(ctx context.Context, art string) MarkenJeObjektAntwort {
	ctx = a.imProfil(ctx)

	je, fehler := a.objektmarken.MarkenJeObjekt(ctx, art)
	if fehler != nil {
		return MarkenJeObjektAntwort{Marken: map[string][]string{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return MarkenJeObjektAntwort{Marken: je}
}

// MarkenFiltern gibt die Objekte einer Art, die durch den Filter gehen.
func (a *App) MarkenFiltern(ctx context.Context, art string, marken []string, verknuepfung string) MarkenfilterAntwort {
	ctx = a.imProfil(ctx)

	kennungen, fehler := a.objektmarken.MarkenFiltern(ctx, art, marken, verknuepfung)
	if fehler != nil {
		return MarkenfilterAntwort{Kennungen: []string{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return MarkenfilterAntwort{Kennungen: kennungen}
}

// Markenkatalog zählt die Marken des Profils.
func (a *App) Markenkatalog(ctx context.Context) MarkenkatalogAntwort {
	ctx = a.imProfil(ctx)

	katalog, fehler := a.objektmarken.Markenkatalog(ctx)
	if fehler != nil {
		return MarkenkatalogAntwort{Marken: []MarkenpostenDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	posten := make([]MarkenpostenDTO, 0, len(katalog))
	for _, einer := range katalog {
		posten = append(posten, MarkenpostenDTO(einer))
	}

	return MarkenkatalogAntwort{Marken: posten}
}

// MarkeUmbenennen gibt einer Marke einen neuen Namen — auf einen vorhandenen
// nur mit `zusammenfuehren`.
func (a *App) MarkeUmbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) MarkenzahlAntwort {
	ctx = a.imProfil(ctx)

	zahl, fehler := a.objektmarken.MarkeUmbenennen(ctx, alt, neu, zusammenfuehren)

	return MarkenzahlAntwort{Anzahl: zahl, Fehler: Uebersetze(ctx, fehler)}
}

// MarkeUeberallEntfernen nimmt eine Marke von jedem Objekt des Profils.
func (a *App) MarkeUeberallEntfernen(ctx context.Context, marke string) MarkenzahlAntwort {
	ctx = a.imProfil(ctx)

	zahl, fehler := a.objektmarken.MarkeUeberallEntfernen(ctx, marke)

	return MarkenzahlAntwort{Anzahl: zahl, Fehler: Uebersetze(ctx, fehler)}
}
