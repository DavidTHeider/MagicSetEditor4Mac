package app

// **Ein Vorschlag wird geschrieben** — Werkstück A11 von Phase 5.3
// (#1450),
// abgetrennt mit A13.1 (#1452),
// weil `kivorschlag.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **wie ein Vorschlag entsteht** — und dass er nichts schreibt.
// Hier steht, **wie er geschrieben wird**: ausgewählt, gegen den
// eingefrorenen Stand geprüft, mit Herkunft und in einem Zug.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIVorschlagUebernehmen schreibt einen abgelegten Vorschlag.
func (a *App) KIVorschlagUebernehmen(
	ctx context.Context, kennung string, felder []string,
) KIUebernahmeAntwort {
	ctx = a.imProfil(ctx)

	// **Zum Übernehmen braucht es keine Quelle** — der Vorschlag liegt schon
	// da, und das Werk prüft ihn gegen den Stand der Karte. Welcher Anbieter
	// ihn erzeugt hat, steht im Auftrag.
	werk, fehler := a.kiwerkOhneQuelle(ctx)
	if fehler != nil {
		return KIUebernahmeAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return KIUebernahmeAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Der Vorschlag gehört einem Profil** (Befund 1 des Nachlaufs über
	// Phase 5.3): das Buch überlebt einen Wechsel, und ohne diese Grenze
	// schriebe eine Übernahme hier die Karte von dort.
	uebernahme, fehler := werk.Uebernehmen(ctx, kennung, felder, profil)
	if fehler != nil {
		return KIUebernahmeAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return KIUebernahmeAntwort{Felder: uebernahme.Felder}
}

// KIVorschlagInEinemZug ist der Weg für die Flächen ohne Fenster.
//
// # Warum dort ein Zug und am Fenster zwei
//
// **Weil ein offener Vorschlag im Prozess lebt** (A6, und dort steht warum).
// Kommandozeile, MCP und REST bauen die Anwendung je Aufruf; ein zweiter Ruf
// fände den Vorschlag des ersten nicht mehr. Wer ihn dafür ablegte, hätte
// eine Tabelle abgelaufener Vorschläge und eine Aufräumfrage dazu.
//
// **Das Profil steht auch hier in der ersten Zeile**, obwohl beide Rufe
// darunter es selbst setzen: eine Bindung, die es nicht nennt, sähe aus wie
// eine, die es vergessen hat — und ein Wächter zählt sie nach.
//
// **Die Zusage ändert sich dabei nicht** (§16): auch hier schreibt nichts,
// solange niemand es verlangt. Am Fenster sieht ein Mensch den Vorschlag und
// drückt dann; hier sagt er es vorher — `uebernehmen` ist die ausdrückliche
// Entscheidung und kein Nebeneffekt.
func (a *App) KIVorschlagInEinemZug(
	ctx context.Context, bezug KIVorschlagsbezug, uebernehmen bool,
	felder []string,
) KIVorschlagAntwort {
	ctx = a.imProfil(ctx)

	vorschlag := a.KIVorschlagenMit(ctx, bezug)
	if vorschlag.Fehler != nil || !uebernehmen {
		return vorschlag
	}

	// **Leer heißt: alles, was durchkam** — und das ist keine Abkürzung um
	// die Auswahl herum (A11, #1450), sondern sie: wer hier `uebernehmen`
	// setzt, hat den Vorschlag nicht gesehen und kann nicht je Feld
	// entscheiden. Wer enger wählen will, nennt die Felder; dann gilt
	// dieselbe Prüfung wie am Fenster.
	if len(felder) == 0 {
		felder = feldnamenAus(vorschlag.Felder)
	}

	uebernahme := a.KIVorschlagUebernehmen(ctx, vorschlag.Kennung, felder)
	if uebernahme.Fehler != nil {
		return KIVorschlagAntwort{
			KIVorschlagDTO: vorschlag.KIVorschlagDTO, Fehler: uebernahme.Fehler,
		}
	}

	vorschlag.Uebernommen = uebernahme.Felder

	return vorschlag
}
