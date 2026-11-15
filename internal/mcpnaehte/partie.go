// **Die Partie am MCP-Server** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Dieselbe kopflose Anwendung wie an der Kommandozeile**
// ([app.PartieflaecheUeber]) und nicht ein zweiter Weg daneben: was ein
// Modell sieht, soll dasselbe sein, was ein Mensch sieht (§16).
//
// **Und die Sicht ist die einzige Auskunft.** Es gibt hier keinen Weg zum
// Vollzustand — der Grund steht in `internal/app/partie.go`, und für ein
// Sprachmodell wiegt er am schwersten: es vergisst nichts, was es einmal
// gelesen hat.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// partien erfüllt `mcp.Partienwerkstatt` über der Naht von `app`.
type partien struct{ db *sql.DB }

func (p partien) PartieAnlegen(
	ctx context.Context, spiel, wirt string, sitze []mcp.Sitzwunsch,
	abgewaehlt []string,
) (mcp.Partieangaben, error) {
	// **Die Wünsche gehen durch, die Kennung bleibt Text.** Ob sie eine
	// Zahl ist, entscheidet `internal/app` — und der Bauweg steht dort an
	// **einer** Stelle für alle vier Flächen (§2). Der erste Anlauf baute
	// ihn hier selbst; der Schichtenwächter hat es angestrichen.
	wuensche := make([]app.Sitzwunsch, 0, len(sitze))

	for _, sitz := range sitze {
		wuensche = append(wuensche, app.Sitzwunsch{
			Kennung: sitz.Kennung, Name: sitz.Name, DeckID: sitz.DeckID,
			Commander: sitz.Commander, ProfilID: sitz.ProfilID,
		})
	}

	return angabenAus(app.PartieflaecheUeber(p.db).PartieUnterRegeln(
		ctx, spiel, wirt, wuensche, app.RegelwerkDTO{Abgewaehlt: abgewaehlt}))
}

func (p partien) Partie(ctx context.Context, kennung, fuer string) (
	mcp.Partieangaben, error,
) {
	return angabenAus(app.PartieflaecheUeber(p.db).Partie(ctx, kennung, fuer))
}

func (p partien) PartieBefehl(ctx context.Context, befehl mcp.Spielbefehl) (
	mcp.Befehlsangaben, error,
) {
	antwort := app.PartieflaecheUeber(p.db).PartieBefehl(ctx, domain.Spielbefehl{
		PartieID: befehl.PartieID,
		Absender: befehl.Absender,
		Kennung:  befehl.Kennung,
		Art:      befehl.Art,
		Nutzlast: befehl.Nutzlast,
	})

	if antwort.Fehler != nil {
		return mcp.Befehlsangaben{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Befehlsangaben{
		Fassung:    antwort.Fassung,
		Ereignisse: ereignisseAus(antwort.Ereignisse),
	}, nil
}

func (p partien) PartieStand(ctx context.Context, kennung, fuer, stand string) (
	mcp.Partieangaben, error,
) {
	flaeche := app.PartieflaecheUeber(p.db)

	switch stand {
	case "pausieren":
		return angabenAus(flaeche.PartiePausieren(ctx, kennung, fuer))

	case "fortsetzen":
		return angabenAus(flaeche.PartieFortsetzen(ctx, kennung, fuer))

	case "beenden":
		return angabenAus(flaeche.PartieBeenden(ctx, kennung, fuer))
	}

	return mcp.Partieangaben{}, domain.NewFieldError("stand", "partiestand",
		"»%s« ist kein Stand — es gibt pausieren, fortsetzen und beenden.",
		stand)
}

// angabenAus übersetzt eine Partieantwort.
//
// **Eine Stelle für alle vier Wege**, die sie geben — vier Kopien liefen bei
// der ersten Änderung auseinander.
// Partien nennt die Partien dieses Profils —
// #1600.
func (p partien) Partien(ctx context.Context) (
	[]mcp.Partiezeilenangaben, error,
) {
	antwort := app.PartieflaecheUeber(p.db).Partien(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Partiezeilenangaben, 0, len(antwort.Partien))

	for _, zeile := range antwort.Partien {
		sitze := make([]mcp.Zeilensitzangaben, 0, len(zeile.Sitze))
		for _, sitz := range zeile.Sitze {
			sitze = append(sitze, mcp.Zeilensitzangaben{
				Kennung: sitz.Kennung, Name: sitz.Name,
			})
		}

		zeilen = append(zeilen, mcp.Partiezeilenangaben{
			ID: zeile.ID, Fassung: zeile.Fassung, Spiel: zeile.Spiel,
			Stand: zeile.Stand, Sitze: sitze,
			Objekte: zeile.Objekte, Zug: zeile.Zug,
			EigenerSitz: zeile.EigenerSitz, Dran: zeile.Dran,
		})
	}

	return zeilen, nil
}

func angabenAus(antwort app.PartieAntwort) (mcp.Partieangaben, error) {
	if antwort.Fehler != nil {
		return mcp.Partieangaben{}, antwort.Fehler.AlsFehler()
	}

	angaben := mcp.Partieangaben{
		ID:      antwort.Partie.ID,
		Fassung: antwort.Partie.Fassung,
		Spiel:   antwort.Partie.Spiel,
		Stand:   antwort.Partie.Stand,
		Fuer:    antwort.Partie.Fuer,
		Wirt:    antwort.Partie.Wirt,
		Sitze:   make([]mcp.Partiesitz, 0, len(antwort.Partie.Sitze)),
		Objekte: make([]mcp.Partieobjekt, 0, len(antwort.Partie.Objekte)),
		Mischpflichten: make([]mcp.Mischpflicht, 0,
			len(antwort.Partie.Mischpflichten)),
		Legendenwahlen: make([]mcp.Legendenwahl, 0,
			len(antwort.Partie.Legendenwahlen)),
		Vorrat:  gefuelltOderLeer(antwort.Partie.Vorrat),
		Befunde: gefuelltOderLeer(antwort.Befunde),
		Zug:     zugangaben(antwort.Partie.Zug),
		Erlaubt: antwort.Erlaubt,

		Handlungen: make([]mcp.Objekthandlung, 0, len(antwort.Handlungen)),

		Abgewaehlt: gefuelltOderLeer(antwort.Partie.Regelwerk.Abgewaehlt),
	}

	for _, eine := range antwort.Handlungen {
		angaben.Handlungen = append(angaben.Handlungen, mcp.Objekthandlung{
			Objekt: eine.Objekt, Handlungen: eine.Handlungen,
			Primaer: eine.Primaer, Zweit: eine.Zweit,
		})
	}

	for _, wahl := range antwort.Partie.Legendenwahlen {
		angaben.Legendenwahlen = append(angaben.Legendenwahlen,
			mcp.Legendenwahl{Sitz: wahl.Sitz, Name: wahl.Name})
	}

	for _, pflicht := range antwort.Partie.Mischpflichten {
		angaben.Mischpflichten = append(angaben.Mischpflichten,
			mcp.Mischpflicht{Sitz: pflicht.Sitz, Zone: pflicht.Zone})
	}

	for _, sitz := range antwort.Partie.Sitze {
		angaben.Sitze = append(angaben.Sitze, mcp.Partiesitz{
			Kennung: sitz.Kennung, Name: sitz.Name, Karten: sitz.Karten,
			Leben: sitz.Leben, Vergiftung: sitz.Vergiftung,
			Manavorrat: sitz.Manavorrat,
		})
	}

	for _, objekt := range antwort.Partie.Objekte {
		angaben.Objekte = append(angaben.Objekte, mcp.Partieobjekt{
			Kennung:     objekt.Kennung,
			Beherrscher: objekt.Beherrscher,
			Zone:        objekt.Zone,
			Karte:       objekt.Karte,
			Name:        objekt.Name,
			KarteID:     objekt.KarteID,
			Farben:      objekt.Eigenschaften.Farbtext(),
			Typzeile:    objekt.Eigenschaften.Typtext(),
			Werte:       objekt.Eigenschaften.Wertetext(),
			Manawert:    manawert(objekt.Eigenschaften),
			Getappt:     objekt.Getappt,
			Geheim:      objekt.Geheim,
		})
	}

	angaben.Nachladefehler = alsNachladefehler(antwort.Nachladefehler)

	return angaben, nil
}

// gefuelltOderLeer gibt eine Liste, die nie `nil` ist.
//
// **Dieselbe Zusage wie an jeder anderen Liste dieser Fläche**: im Fenster
// und in jedem fremden Leser stünde sonst `null`, und eine Schleife darüber
// fiele um.
func gefuelltOderLeer(werte []string) []string {
	if werte == nil {
		return []string{}
	}

	return werte
}

// manawert gibt den Manawert — **null, wo es keine Eigenschaften gibt**.
//
// **Eine eigene Zeile, weil ein Zeiger kein Feld hat.** Die drei Kurzformen
// daneben sind Methoden und vertragen `nil`; eine Zahl ist keine.
func manawert(eigen *app.EigenschaftenDTO) int {
	if eigen == nil {
		return 0
	}

	return eigen.Manawert
}

// zugangaben bildet den Zugstand ab — §16, seit `M6`
// (#1608).
func zugangaben(zug app.ZugDTO) mcp.Partiezug {
	return mcp.Partiezug{
		Nummer: zug.Nummer, Aktiv: zug.Aktiv, Starter: zug.Starter,
		Phase: zug.Phase, Schritt: zug.Schritt,
		Prioritaet: zug.Prioritaet, Abwurf: zug.Abwurf,
		Angriffe: kampfpaare(zug.Angriffe),
		Bloecke:  kampfpaare(zug.Bloecke),
	}
}

// kampfpaare bildet den Kampfstand ab — §16, seit
// #1623.
func kampfpaare(paare []app.KampfpaarDTO) []mcp.Kampfpaar {
	abgebildet := make([]mcp.Kampfpaar, 0, len(paare))

	for _, paar := range paare {
		abgebildet = append(abgebildet, mcp.Kampfpaar{
			Kreatur: paar.Kreatur, Ziel: paar.Ziel,
			Angreifer: paar.Angreifer,
		})
	}

	return abgebildet
}
