package app

// **Die Bindungen für die zwei Dienste** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Drei je Dienst und kein Schalter
//
// **Weil Anfangen und Aufhören zwei Entscheidungen sind** und der Stand eine
// Frage — dieselbe Gestalt wie beim Horchposten aus N2. Ein Wahrheitswert
// nähme die Adresse auch beim Ausschalten entgegen.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// DienstAntwort ist der Ausgang eines Dienstwegs.
type DienstAntwort struct {
	Stand Dienstzustand `json:"stand"`

	Fehler *Fehler `json:"fehler"`
}

// RESTStarten macht den REST-Dienst auf.
//
// **Schreibend** (`docs/SCHREIBWEGE.md`) — es entsteht ein Port, und das ist
// die Wirkung `externEmpfangen` (N7, #1568).
func (b *Bindungen) RESTStarten(ctx context.Context, adresse string, port int) DienstAntwort {
	return b.dienstStarten(ctx, domain.DienstREST, adresse, port)
}

// RESTBeenden schließt ihn wieder.
func (b *Bindungen) RESTBeenden(ctx context.Context) DienstAntwort {
	return b.dienstBeenden(ctx, domain.DienstREST)
}

// RESTStand sagt, ob er läuft und wo.
//
// **Rein lesend.**
func (b *Bindungen) RESTStand(ctx context.Context) DienstAntwort {
	return DienstAntwort{Stand: b.app(ctx).dienststand(domain.DienstREST)}
}

// MCPStarten macht den Werkzeugserver über HTTP auf.
func (b *Bindungen) MCPStarten(ctx context.Context, adresse string, port int) DienstAntwort {
	return b.dienstStarten(ctx, domain.DienstMCP, adresse, port)
}

// MCPBeenden schließt ihn wieder.
func (b *Bindungen) MCPBeenden(ctx context.Context) DienstAntwort {
	return b.dienstBeenden(ctx, domain.DienstMCP)
}

// MCPStand sagt, ob er läuft und wo.
func (b *Bindungen) MCPStand(ctx context.Context) DienstAntwort {
	return DienstAntwort{Stand: b.app(ctx).dienststand(domain.DienstMCP)}
}

// dienstStarten ist der gemeinsame Weg beider Dienste.
func (b *Bindungen) dienstStarten(
	ctx context.Context,
	art domain.Dienstart, adresse string, port int,
) DienstAntwort {
	ctx = b.ruf(ctx)

	// **Leer heißt die Einstellung** `netz.lauschen` und ohne sie die Schleife
	// (G08, #2272); eine Liste mit Komma nennt mehrere Adressen.
	var genannt []string
	if strings.TrimSpace(adresse) != "" {
		genannt = []string{adresse}
	}

	stand, fehler := b.app(ctx).dienstStartenUeberWahl(ctx, art, genannt, port)
	if fehler != nil {
		return DienstAntwort{
			Stand:  b.app(ctx).dienststand(art),
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	return DienstAntwort{Stand: stand}
}

// dienstBeenden ist der gemeinsame Weg zurück.
//
// **Ohne Fehlerweg**, und das ist Absicht: „beenden" ist erreicht, wenn
// nichts mehr läuft — auch dann, wenn nichts lief. Ein Fehler dafür wäre eine
// Meldung über ein erreichtes Ziel.
func (b *Bindungen) dienstBeenden(ctx context.Context, art domain.Dienstart) DienstAntwort {
	// **Der Kontext wird gebunden und nicht gebraucht** — derselbe Grund wie
	// bei [Bindungen.HorchBeenden]: wer ihn ausließe, hielte diese Bindung
	// aus der Umschaltsperre heraus, und ein Beenden mitten im Profilwechsel
	// griffe an einem Stand, den es gerade nicht mehr gibt (§7, Regel 2).
	_ = b.ruf(ctx)

	b.app(ctx).dienstBeenden(art)

	return DienstAntwort{Stand: b.app(ctx).dienststand(art)}
}
