// **Die Partie am Dienst** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Dieselbe kopflose Anwendung wie an der Kommandozeile und am MCP-Server**
// ([app.PartieflaecheUeber]) und nicht ein dritter Weg daneben: was über
// HTTP zu sehen ist, soll dasselbe sein, was ein Mensch sieht (§16).
//
// **Und die Sicht ist die einzige Auskunft.** Es gibt hier keinen Weg zum
// Vollzustand — der Grund steht in `internal/app/partie.go`. Über einen Port
// wiegt er noch etwas schwerer als im Fenster: wer hier fragt, sitzt
// womöglich gar nicht an diesem Tisch.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// partien erfüllt `rest.Partienwerkstatt` über der Naht von `app`.
type partien struct{ db *sql.DB }

func (p partien) PartieAnlegen(
	ctx context.Context, spiel, wirt string, sitze []rest.Sitzwunsch,
	abgewaehlt []string,
) (rest.Partieangaben, error) {
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
	rest.Partieangaben, error,
) {
	return angabenAus(app.PartieflaecheUeber(p.db).Partie(ctx, kennung, fuer))
}

// Partien nennt die Partien dieses Profils —
// #1600.
func (p partien) Partien(ctx context.Context) (
	[]rest.Partiezeilenangaben, error,
) {
	antwort := app.PartieflaecheUeber(p.db).Partien(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Partiezeilenangaben, 0, len(antwort.Partien))

	for _, zeile := range antwort.Partien {
		sitze := make([]rest.Zeilensitzangaben, 0, len(zeile.Sitze))
		for _, sitz := range zeile.Sitze {
			sitze = append(sitze, rest.Zeilensitzangaben{
				Kennung: sitz.Kennung, Name: sitz.Name,
			})
		}

		zeilen = append(zeilen, rest.Partiezeilenangaben{
			ID: zeile.ID, Fassung: zeile.Fassung, Spiel: zeile.Spiel,
			Stand: zeile.Stand, Sitze: sitze,
			Objekte: zeile.Objekte, Zug: zeile.Zug,
			EigenerSitz: zeile.EigenerSitz, Dran: zeile.Dran,
		})
	}

	return zeilen, nil
}

func (p partien) PartieBefehl(ctx context.Context, befehl rest.Spielbefehl) (
	rest.Befehlsangaben, error,
) {
	antwort := app.PartieflaecheUeber(p.db).PartieBefehl(ctx, domain.Spielbefehl{
		PartieID: befehl.PartieID,
		Absender: befehl.Absender,
		Kennung:  befehl.Kennung,
		Art:      befehl.Art,
		Nutzlast: befehl.Nutzlast,
	})

	if antwort.Fehler != nil {
		return rest.Befehlsangaben{}, antwort.Fehler.AlsFehler()
	}

	return rest.Befehlsangaben{
		Fassung:    antwort.Fassung,
		Ereignisse: ereignisseAus(antwort.Ereignisse),
	}, nil
}

func (p partien) PartieStand(ctx context.Context, kennung, fuer, stand string) (
	rest.Partieangaben, error,
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

	return rest.Partieangaben{}, domain.NewFieldError("stand", "partiestand",
		"»%s« ist kein Stand — es gibt pausieren, fortsetzen und beenden.",
		stand)
}

// angabenAus übersetzt eine Partieantwort.
//
// **Eine Stelle für alle vier Wege**, die sie geben — vier Kopien liefen bei
// der ersten Änderung auseinander.
func angabenAus(antwort app.PartieAntwort) (rest.Partieangaben, error) {
	if antwort.Fehler != nil {
		return rest.Partieangaben{}, antwort.Fehler.AlsFehler()
	}

	angaben := rest.Partieangaben{
		ID:      antwort.Partie.ID,
		Fassung: antwort.Partie.Fassung,
		Spiel:   antwort.Partie.Spiel,
		Stand:   antwort.Partie.Stand,
		Fuer:    antwort.Partie.Fuer,
		Wirt:    antwort.Partie.Wirt,
		Sitze:   make([]rest.Partiesitz, 0, len(antwort.Partie.Sitze)),
		Objekte: make([]rest.Partieobjekt, 0, len(antwort.Partie.Objekte)),
		Mischpflichten: make([]rest.Mischpflicht, 0,
			len(antwort.Partie.Mischpflichten)),
		Legendenwahlen: make([]rest.Legendenwahl, 0,
			len(antwort.Partie.Legendenwahlen)),
		Vorrat:  gefuelltOderLeer(antwort.Partie.Vorrat),
		Befunde: gefuelltOderLeer(antwort.Befunde),
		Zug:     zugangaben(antwort.Partie.Zug),
		Erlaubt: antwort.Erlaubt,

		Handlungen: make([]rest.Objekthandlung, 0, len(antwort.Handlungen)),

		Abgewaehlt: gefuelltOderLeer(antwort.Partie.Regelwerk.Abgewaehlt),
	}

	for _, eine := range antwort.Handlungen {
		angaben.Handlungen = append(angaben.Handlungen, rest.Objekthandlung{
			Objekt: eine.Objekt, Handlungen: eine.Handlungen,
			Primaer: eine.Primaer, Zweit: eine.Zweit,
		})
	}

	for _, wahl := range antwort.Partie.Legendenwahlen {
		angaben.Legendenwahlen = append(angaben.Legendenwahlen,
			rest.Legendenwahl{Sitz: wahl.Sitz, Name: wahl.Name})
	}

	for _, pflicht := range antwort.Partie.Mischpflichten {
		angaben.Mischpflichten = append(angaben.Mischpflichten,
			rest.Mischpflicht{Sitz: pflicht.Sitz, Zone: pflicht.Zone})
	}

	for _, sitz := range antwort.Partie.Sitze {
		angaben.Sitze = append(angaben.Sitze, rest.Partiesitz{
			Kennung: sitz.Kennung, Name: sitz.Name, Karten: sitz.Karten,
			Leben: sitz.Leben, Vergiftung: sitz.Vergiftung,
			Manavorrat: sitz.Manavorrat,
		})
	}

	for _, objekt := range antwort.Partie.Objekte {
		angaben.Objekte = append(angaben.Objekte, rest.Partieobjekt{
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

	angaben.Nachladefehler = alsNachladebefund(antwort.Nachladefehler)

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
func zugangaben(zug app.ZugDTO) rest.Partiezug {
	return rest.Partiezug{
		Nummer: zug.Nummer, Aktiv: zug.Aktiv, Starter: zug.Starter,
		Phase: zug.Phase, Schritt: zug.Schritt,
		Prioritaet: zug.Prioritaet, Abwurf: zug.Abwurf,
		Angriffe: kampfpaare(zug.Angriffe),
		Bloecke:  kampfpaare(zug.Bloecke),
	}
}

// kampfpaare bildet den Kampfstand ab — §16, seit
// #1623.
func kampfpaare(paare []app.KampfpaarDTO) []rest.Kampfpaar {
	abgebildet := make([]rest.Kampfpaar, 0, len(paare))

	for _, paar := range paare {
		abgebildet = append(abgebildet, rest.Kampfpaar{
			Kreatur: paar.Kreatur, Ziel: paar.Ziel,
			Angreifer: paar.Angreifer,
		})
	}

	return abgebildet
}

// Die vier Wege der Verwaltung — PLAY-03.AK-05
// (#1711).
//
// **Über dieselbe kopflose Anwendung wie die fünf darüber** und nicht über
// einen eigenen Weg: was über HTTP geschieht, soll dasselbe sein, was ein
// Mensch im Fenster tut (§16).

func (p partien) PartieLoeschen(ctx context.Context, kennung string) (
	rest.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(app.PartieflaecheUeber(p.db).PartieLoeschen(ctx, kennung))
}

func (p partien) PartieKopieren(ctx context.Context, kennung string) (
	rest.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(app.PartieflaecheUeber(p.db).PartieKopieren(ctx, kennung))
}

func (p partien) PartieAusfuhr(ctx context.Context, kennung string) (
	rest.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(
		app.PartieflaecheUeber(p.db).PartieAusfuhrText(ctx, kennung))
}

func (p partien) PartieEinfuhr(ctx context.Context, inhalt string) (
	rest.Partieverwaltungsangaben, error,
) {
	return verwaltungAus(
		app.PartieflaecheUeber(p.db).PartieEinfuhrText(ctx, inhalt))
}

// verwaltungAus bildet die Antwort der Anwendung auf die des Dienstes ab.
//
// **Der Pfad geht nicht mit**: er wäre einer auf **diesem** Rechner, und wer
// hier fragt, sitzt an einem anderen.
func verwaltungAus(antwort app.PartieverwaltungAntwort) (
	rest.Partieverwaltungsangaben, error,
) {
	if antwort.Fehler != nil {
		return rest.Partieverwaltungsangaben{}, antwort.Fehler.AlsFehler()
	}

	befunde := antwort.Befunde
	if befunde == nil {
		befunde = []string{}
	}

	return rest.Partieverwaltungsangaben{
		Kennung: antwort.Kennung, Inhalt: antwort.Inhalt, Befunde: befunde,
	}, nil
}
