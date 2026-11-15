package mcpnaehte

// **Die Recherche am MCP-Server** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Über dieselbe Anwendung wie jede andere Fläche** (§16): was das Fenster
// findet, findet ein Modell auch — mit derselben Herkunftsangabe und
// demselben Grund, wenn nichts kommt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// rechercheauskunft reicht die zwei lesenden Wege an den Server.
type rechercheauskunft struct{ db *sql.DB }

func (r rechercheauskunft) Recherche(
	ctx context.Context, fall, frage string, grenze int,
	sprache, korpusstand, fortsetzung string, freigegeben []string,
) (mcp.Rechercheergebnis, error) {
	return alsMCPRecherche(app.RechercheAnwendungUeber(r.db).Recherchieren(
		ctx, rechercheFallOder(fall), frage, grenze,
		app.RecherchefilterDTO{
			Sprache: sprache, Korpusstand: korpusstand,
			Freigegeben: freigegeben,
		},
		fortsetzung))
}

func (r rechercheauskunft) AehnlicheKarten(
	ctx context.Context, karteID int64, grenze int, sprache, fortsetzung string,
	freigegeben []string,
) (mcp.Rechercheergebnis, error) {
	// **Ohne Stand**: der Stand schützt eine Anzeige, die ein Modell nicht
	// vor sich hat — es fragt nach einer Karte, nicht nach einem Bildschirm.
	return alsMCPRecherche(app.RechercheAnwendungUeber(r.db).AehnlicheKarten(
		ctx, karteID, "", grenze,
		app.RecherchefilterDTO{Sprache: sprache, Freigegeben: freigegeben},
		fortsetzung))
}

// rechercheFallOder gibt den Kartenfall, wo keiner steht.
//
// **Und nicht den Ähnlichkeitsfall**: der braucht eine Ausgangskarte, und ein
// Modell, das keinen Fall nennt, hat keine.
func rechercheFallOder(fall string) string {
	if fall == "" {
		return "karten"
	}

	return fall
}

// alsMCPRecherche übersetzt die Antwort.
func alsMCPRecherche(antwort app.RechercheAntwort) (
	mcp.Rechercheergebnis, error,
) {
	if antwort.Fehler != nil {
		return mcp.Rechercheergebnis{}, antwort.Fehler.AlsFehler()
	}

	treffer := make([]mcp.Recherchetreffer, 0, len(antwort.Treffer))
	for _, einzelner := range antwort.Treffer {
		treffer = append(treffer, mcp.Recherchetreffer{
			Rang: einzelner.Rang, Kennung: einzelner.Kennung,
			Quelle:         einzelner.Quelle,
			Kennungssystem: einzelner.Kennungssystem,
			Fachkennung:    einzelner.Fachkennung,
			Semantik:       einzelner.Semantik,
			EigeneKarteID:  einzelner.EigeneKarteID,
			Titel:          einzelner.Titel, Ausschnitt: einzelner.Ausschnitt,
			Text: einzelner.Text, Sprache: einzelner.Sprache,
			Regelnummer: einzelner.Regelnummer, Urheber: einzelner.Urheber,
			Bezugsadresse: einzelner.Bezugsadresse,
			Quellrevision: einzelner.Quellrevision,
			Quelldatum:    einzelner.Quelldatum,
			Abgerufen:     einzelner.Abgerufen,
			Weitere:       einzelner.Weitere,
		})
	}

	staende := make([]mcp.Korpusstand, 0, len(antwort.Korpusstaende))
	for _, stand := range antwort.Korpusstaende {
		staende = append(staende, mcp.Korpusstand{
			Quelle: stand.Quelle, Revision: stand.Revision, Datum: stand.Datum,
		})
	}

	return mcp.Rechercheergebnis{
		Treffer: treffer, Modus: antwort.Modus,
		Leere: antwort.Leere, Grund: antwort.Grund,
		Modell:            antwort.Modell,
		Suchkonfiguration: antwort.Suchkonfiguration,
		Korpusstaende:     staende,
		Einschraenkungen:  antwort.Einschraenkungen,
		Fortsetzung:       antwort.Fortsetzung,
	}, nil
}
