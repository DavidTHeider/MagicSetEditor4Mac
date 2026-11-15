package mcpnaehte

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Grenzen und Guthaben des MCP-Servers** — Werkstück O5 von Phase 5.2 (#1429).
//
// Eine MCP-Sitzung ist **ein** Principal: ein Prozess, ein Modell, ein Guthaben.
// Die Plätze je Ressourcenklasse bucht nicht der Server, sondern der Dienst,
// der die Arbeit tut — derselbe wie am Fenster und im REST-Dienst.

// grenzenauskunft reicht die Grenzen an das Werkzeug.
//
// **Zwei Herkünfte, und die Wurzel wählt** (QA-08.AK-03, #1739): der
// Werkzeugserver nennt, was er **beim Start** gelesen hat — dort entstehen
// seine Guthaben ([MitGrenzenVomStart]) —, die Kommandozentrale des Fensters
// nennt, was der Bestand **jetzt** sagt, wie die Hilfe daneben.
type grenzenauskunft struct {
	db      *sql.DB
	gelesen *app.GrenzenDTO
}

func (g grenzenauskunft) Grenzen(ctx context.Context) (mcp.Grenzen, error) {
	if g.gelesen == nil {
		haushalt, fehler := app.GrenzenUeber(ctx, g.db)
		if fehler != nil {
			return mcp.Grenzen{}, fehler //nolint:wrapcheck // der Fehler des Bestands, wie er ist
		}

		jetzt := app.GrenzenauskunftVon(haushalt)
		g.gelesen = &jetzt
	}

	grenzen := *g.gelesen

	return mcp.Grenzen{
		Parallel:     grenzen.Parallel,
		AntwortBytes: grenzen.AntwortBytes,
		Principal:    mcp.Kontingent(grenzen.Principal),
		Global:       mcp.Kontingent(grenzen.Global),
		Reserve:      grenzen.Reserve,
		Kosten:       grenzen.Kosten,
		Kontingente:  grenzen.Kontingente,
		Verbindungen: grenzen.Verbindungen,
		Auftraege:    mcp.Auftragsgrenzen(grenzen.Auftraege),
		Kostentafel:  grenzen.Kostentafel,
	}, nil
}

// MitGrenzenVomStart setzt die Grenzen, die der Werkzeugserver beim Start
// gelesen hat und nach denen er bucht.
func MitGrenzenVomStart(naehte mcp.Naehte, grenzen domain.Haushaltsgrenzen) mcp.Naehte {
	gelesen := app.GrenzenauskunftVon(grenzen)
	naehte.Grenzen = grenzenauskunft{gelesen: &gelesen}

	return naehte
}

// principalMCP ist der Principal einer MCP-Sitzung **ohne Token**.
//
// **Der Altbetrieb und nichts weiter** — Befund 3 des Nachlaufs zum Zweitblick
// über Phase 5.2. Die Naht kennt den Ruf und zählt nach der Maschine — wie am
// Dienst (0080) —, und dieser Name bleibt für den Weg **ohne** Token.
//
// **Was das Guthaben nicht kann, gehört trotzdem gesagt**: es lebt im
// Arbeitsspeicher **dieses** Prozesses. Ein neu gestarteter MCP-Server fängt
// mit vollem Topf an, gleich welche Maschine er vorzeigt — und der Dienst
// nach einem Neustart ebenso. Ein Guthaben über Prozessgrenzen hinweg wäre
// ein Zähler im Bestand und keine Naht; wer ihn will, entscheidet das in
// einem Vermerk und nicht hier.
const principalMCP = "mcp"

// kontingentbuch ist, was das Guthaben können muss — die Gestalt von
// `service.Kontingente`, das dieses Programm nicht kennt (§2): es holt es
// über `app.NeueKontingente`.
type kontingentbuch interface {
	Abbuchen(principal string, klasse domain.Kostenklasse, fenster bool) (bool, time.Duration)
}

// KontingentFuer baut die Naht zum Guthaben — gezählt nach der Maschine, die
// der Ruf vorzeigt, und global gegen alle anderen Zugänge derselben Maschine.
func KontingentFuer(
	kontingente kontingentbuch,
) func(ruf context.Context, art mcp.Art, name string) (bool, time.Duration) {
	return func(ruf context.Context, art mcp.Art, name string) (bool, time.Duration) {
		return kontingente.Abbuchen(
			principalAus(ruf), kostenklasseFuer(art, name), false)
	}
}

// principalAus gibt die Maschine des Rufs — oder den Namen des Altbetriebs.
//
// **Die Identität kommt von der Tokenwache** und hängt am Kontext (`app`,
// Befund 5 des Zweitblicks über Phase 5.2). Ohne Token hängt dort nichts, und
// dann ist der Prozess selbst der Principal.
func principalAus(ruf context.Context) string {
	if kennung := service.TokenidentitaetAus(ruf); kennung != "" {
		return kennung
	}

	return principalMCP
}

// kostenklasseFuer sagt, was ein Werkzeug kostet — nach der Kostentafel,
// sonst nach seiner Art (#2003).
//
// **Hier stand bis zum 2026-10-02 eine Suche nach Wörtern im Namen**:
// „vorschau" kostete 5, „bild" 3, „einlesen" 8. Eine Vorschau der
// Einstellungen rastert nichts, und ein Profilbild wegzunehmen verarbeitet
// kein Bild — und dieselbe Fähigkeit kostete über REST etwas anderes. Jetzt
// fragen beide Flächen dieselbe Tafel.
func kostenklasseFuer(art mcp.Art, name string) domain.Kostenklasse {
	return domain.KostenklasseFuer(domain.FlaecheMCP, name, art == mcp.Schreibend)
}
