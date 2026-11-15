package app

// **Die Partie für die drei Flächen ohne Fenster** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Warum das hier steht und nicht in jeder Fläche
//
// **Damit alle vier dieselbe Anwendung benutzen** (§16) — dieselbe Form wie
// `turnierflaeche.go` und `umschlagflaeche.go`. Eine Fläche, die sich ihre
// Leitung selbst zusammensetzte, hätte ihre eigene Fassung davon, und drei
// Wurzeln liefen auseinander.
//
// # Welcher Regelausleger hier gilt
//
// **Der Sandkasten**, und das ist eine Entscheidung mit Grund: G1 baut den
// Kern, und **welche** Regeln ein Format wirklich trägt, ist Werkstück G4
// (#1461).
// Bis dahin gilt für jede Partie, was auch am Küchentisch gilt: der Tisch
// gehorcht dem Menschen.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Partieflaeche ist die Anwendung selbst, für eine Fläche ohne Fenster.
type Partieflaeche struct{ app *App }

// PartieflaecheUeber baut sie über einem Bestand.
func PartieflaecheUeber(db *sql.DB) *Partieflaeche {
	anwendung := MitPartienleitung(NeueApp(""), db)

	// **Wer die Profile kennt, gehört hierher und nicht in
	// [MitPartienleitung]** (UXR-F21,
	// #2041):
	// das Fenster hat seinen Kenner aus [App.leuteStarten] schon.
	//
	// **Ohne ihn sagt keine dieser Flächen etwas zu einer Bindung ins Leere**
	// ([App.profilbefunde] gibt dann keine): eine Partie mit einem Profil, das
	// es nicht gibt, sähe an Kommandozeile, MCP und Dienst aus wie eine ohne
	// Bindung, und am Fenster steht ein Befund (§16, §8).
	anwendung.MitProfilkenner(storage.NeueNutzerablage(db))

	return &Partieflaeche{app: anwendung}
}

// MitPartienleitung hängt die Partieleitung in eine Anwendung.
//
// **Ein eigener Ruf, damit das Fenster ihn auch hat.** Der Composition Root
// des Fensters baut seine Anwendung selbst; ohne diese Zeile hätte er eine
// ohne Partien, und der Unterschied fiele erst beim ersten Klick auf.
func MitPartienleitung(anwendung *App, db *sql.DB) *App {
	// **Der Bestand muss dran sein, bevor die Leihstelle entsteht** (G6,
	// #1463).
	//
	// **Sie baut ihre Ablagen aus `a.db`** — wie [App.peerwerk], dessen
	// Ausweis sie braucht. Der erste Anlauf ließ das Feld leer, und
	// [Partieflaeche] fiel beim ersten Vorlegen über einen Nullzeiger; ein
	// Lauf über einen Sitz mit Leihe hat es gefunden.
	anwendung.db = db

	// **Die Partiedatei gehört dazu** (PLAY-03.AK-05, #1711): ohne sie
	// fiele die Ausfuhr einer Fläche ohne Fenster über einen Nullzeiger —
	// derselbe Grund wie beim Abzug des Tisches.
	anwendung.partiedatei = partiedateiDienst()

	// **Die Blattquelle macht den Abzug selbsttragend** (G6,
	// #1463):
	// ein verliehenes Deck zeichnet auf einem fremden Rechner nur, wenn
	// Feldwerte und Stil eingefroren mitgehen.
	anwendung.MitPartieaufbau(
		service.NeuerPartieaufbau(storage.NeueDeckablage(db)).
			MitBlaettern(storage.NeuerSetLeser(db)).
			MitLeihen(anwendung.leihstelle()))

	return anwendung.MitPartienaht(service.NeuePartieleitung(
		storage.NeuePartieablage(db),
		service.NeuerSandkasten(service.NeuerZufall())))
}

// Sitzwunsch ist ein Platz am Tisch samt Deck.
//
// **Eine Gestalt dieser Schicht und nicht die von `service`.** Eine Fläche
// ohne Fenster darf `internal/service` nicht kennen (§2), und `deckID` kommt
// von dort als Text herein — aus einer Kommandozeile, aus JSON, aus einem
// Formular.
type Sitzwunsch struct {
	Kennung string
	Name    string
	DeckID  string

	// Commander ist die Kartenkennung des Commanders — CR 903.3 (M19,
	// #1620).
	// **Leer heißt: keiner.**
	Commander string

	// LeiheID ist eine **geliehene** Leihe statt eines eigenen Decks (G6,
	// #1463).
	//
	// **Sie schlägt `DeckID`**, und die bleibt dann leer: wer eine Leihe
	// vorlegt, bringt einen eingefrorenen Abzug mit, und ein Deck daneben
	// wäre eine zweite Antwort auf dieselbe Frage.
	LeiheID string

	// ProfilID bindet diesen Sitz an ein Profil **desselben Nutzers** (G8,
	// #1465)
	// — **leer heißt: keines**, und das ist der Normalfall.
	ProfilID string

	// Eingeladen ist die Peerkennung eines Gastes, der sein Deck mitbringt
	// (P9b, #2377).
	Eingeladen string
}

// PartieAusWuenschen baut eine Partie aus Sitzwünschen und legt sie an.
//
// # Warum dieser Weg hier steht und nicht in jeder Fläche
//
// **Weil er sonst dreimal dastünde.** Der erste Anlauf baute den
// [service.Partieaufbau] in `cmd/mse`, in `internal/mcpnaehte` **und** in
// `cmd/mse-rest` — drei Wurzeln für dieselbe Frage, und §2 warnt genau davor.
// **Angestrichen hat es der Schichtenwächter**: beide Kommandozeilen
// importierten `internal/service`, und das dürfen sie nicht.
//
// **Die Vorgaben stehen an einer Stelle**: der Sandkasten als Regelwerk, der
// erste Sitz als Wirt, eine frische Kennung. Wer sie je Fläche setzte, hätte
// drei Vorgaben für eine Entscheidung.
func (p *Partieflaeche) PartieAusWuenschen(
	ctx context.Context, spiel, wirt string, sitze []Sitzwunsch,
) PartieAntwort {
	return p.PartieAusWuenschenMit(ctx, spiel, wirt, sitze, "")
}

// PartieAusWuenschenMit legt sie unter einem **genannten** Regelwerk an —
// CR 903 (M19,
// #1620).
//
// **Leer heißt: der Sandkasten**, und das ist der Normalfall. Was der
// Kern daraus macht, entscheidet [domain.ProfilZu] — an **einer** Stelle
// für alle vier Flächen (§2).
func (p *Partieflaeche) PartieAusWuenschenMit(
	ctx context.Context, spiel, wirt string, sitze []Sitzwunsch,
	regelwerk string,
) PartieAntwort {
	return p.PartieUnterRegeln(ctx, spiel, wirt, sitze,
		RegelwerkDTO{Kennung: regelwerk})
}

// PartieUnterRegeln legt sie unter einem Regelwerk **samt Abwahl** an — P1a
// (#2335).
//
// **Derselbe Weg wie das Fenster**: [regelwerkAus] macht aus dem DTO das
// Regelwerk, und der Aufbau friert die Abwahl samt wirksamer Matrix ein.
// Bis zum 2026-10-13 kannten die drei Flächen ohne Fenster gar keine
// Abwahl — wer an der Kommandozeile die Zugstruktur selbst übernehmen
// wollte, konnte es nicht sagen. Eine unbekannte Abwahl weist
// [domain.Regelwerk.Pruefe] ab, bevor etwas entsteht.
func (p *Partieflaeche) PartieUnterRegeln(
	ctx context.Context, spiel, wirt string, sitze []Sitzwunsch,
	werk RegelwerkDTO,
) PartieAntwort {
	partie, fehler := p.app.partieAusWuenschenMit(ctx, spiel, wirt, sitze,
		regelwerkAus(werk))
	if fehler != nil {
		return leerePartie(ctx, fehler)
	}

	return p.app.PartieAnlegen(ctx, partie)
}

// PartieAnlegen legt eine fertig gebaute Partie an.
func (p *Partieflaeche) PartieAnlegen(
	ctx context.Context, partie domain.Partie,
) PartieAntwort {
	return p.app.PartieAnlegen(ctx, partie)
}

// Partie gibt eine Partie, wie `fuer` sie sehen darf.
func (p *Partieflaeche) Partie(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	return p.app.Partie(ctx, kennung, fuer)
}

// Partien nennt die Partien dieses Profils —
// #1600.
//
// **Auch hier und nicht nur am Fenster** (§16): wer über die Kommandozeile
// spielt, braucht die Kennung seiner laufenden Partie genauso — und ohne
// diesen Weg müsste er sie sich aufschreiben.
func (p *Partieflaeche) Partien(ctx context.Context) PartienAntwort {
	return p.app.Partien(ctx)
}

// PartieVerlauf gibt den Ereignisstrom, wie `fuer` ihn sehen darf.
//
// **Werkstück G5 von Phase 5.5**
// (#1462):
// dieselbe Fähigkeit an jeder Fläche (§16), und dieselbe Grenze — gefiltert
// hat `service`, bevor es hier ankommt.
func (p *Partieflaeche) PartieVerlauf(
	ctx context.Context, kennung, fuer string,
) PartieverlaufAntwort {
	return p.app.PartieVerlauf(ctx, kennung, fuer)
}

// PartieBefehl stellt einen Spielbefehl.
func (p *Partieflaeche) PartieBefehl(
	ctx context.Context, befehl domain.Spielbefehl,
) BefehlAntwort {
	return p.app.PartieBefehl(ctx, befehl)
}

// PartiePausieren hält eine Partie an.
func (p *Partieflaeche) PartiePausieren(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	return p.app.PartiePausieren(ctx, kennung, fuer)
}

// PartieFortsetzen nimmt sie wieder auf.
func (p *Partieflaeche) PartieFortsetzen(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	return p.app.PartieFortsetzen(ctx, kennung, fuer)
}

// PartieBeenden schließt sie ab.
func (p *Partieflaeche) PartieBeenden(
	ctx context.Context, kennung, fuer string,
) PartieAntwort {
	return p.app.PartieBeenden(ctx, kennung, fuer)
}

// Die vier Wege der Verwaltung — PLAY-03.AK-05
// (#1711).
//
// **Ausfuhr und Einfuhr gehen hier über den Text und nicht über einen Ort**
// — REST und MCP stehen an keiner Platte, und ein Pfad wäre dort einer auf
// einem fremden Rechner (§16). Die Kommandozeile steht an einer und nimmt
// deshalb [Partieflaeche.PartieAusfuhrSchreiben].

// PartieLoeschen nimmt eine Partie aus dem Bestand.
func (p *Partieflaeche) PartieLoeschen(
	ctx context.Context, kennung string,
) PartieverwaltungAntwort {
	return p.app.PartieLoeschen(ctx, kennung)
}

// PartieKopieren legt sie als neuen Übungszustand daneben.
func (p *Partieflaeche) PartieKopieren(
	ctx context.Context, kennung string,
) PartieverwaltungAntwort {
	return p.app.PartieKopieren(ctx, kennung)
}

// PartieAusfuhrText gibt sie als Text.
func (p *Partieflaeche) PartieAusfuhrText(
	ctx context.Context, kennung string,
) PartieverwaltungAntwort {
	return p.app.PartieAusfuhrText(ctx, kennung)
}

// PartieEinfuhrText legt eine aus einem Text an.
func (p *Partieflaeche) PartieEinfuhrText(
	ctx context.Context, inhalt string,
) PartieverwaltungAntwort {
	return p.app.PartieEinfuhrText(ctx, inhalt)
}

// PartieAusfuhrSchreiben legt sie an einen Ort auf dieser Platte.
func (p *Partieflaeche) PartieAusfuhrSchreiben(
	ctx context.Context, kennung, pfad string,
) PartieverwaltungAntwort {
	return p.app.PartieAusfuhrSchreiben(ctx, kennung, pfad)
}

// PartieEinfuhrLesen liest eine von dort.
func (p *Partieflaeche) PartieEinfuhrLesen(
	ctx context.Context, pfad string,
) PartieverwaltungAntwort {
	return p.app.PartieEinfuhrLesen(ctx, pfad)
}

// leihstelle gibt die Leihstelle dieses Profils (G6,
// #1463).
//
// # Warum sie je Ruf entsteht und nicht im Bauplatz
//
// **Weil sie nichts hält.** Eine Leihstelle liest und schreibt; sie hat
// keinen Lauf, keine Warteschlange und keinen Zustand, der einen Ruf
// überlebt — derselbe Schnitt wie bei [App.peerwerk].
//
// # Und warum sie hier steht und nicht in einer eigenen Datei
//
// **Weil sie einen Bestand baut und das nur ein Composition Root darf** (§2).
// Der erste Anlauf legte sie in ein `deckleihenaht.go`, und der
// Schichtenwächter hat es angestrichen — zu Recht: die Ausnahmeliste ist
// lang genug, und eine Leihe gehört ohnehin hierher. Sie trägt eine Partie.
func (a *App) leihstelle() *service.Leihstelle {
	return service.NeueLeihstelle(
		storage.NeueLeihablage(a.db), storage.NeueDeckablage(a.db),
		a.peerwerk(), a.uhr).MitHandelsschranke(einstellungsschranke{a: a})
}
