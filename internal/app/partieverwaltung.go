package app

// **Eine Partie loswerden, verdoppeln, ausführen, einlesen** — PLAY-03.AK-05
// (#1711).
//
// # Warum die vier zusammenstehen
//
// **Weil sie eine Frage teilen: was geschieht mit einer Partie als Ganzem?**
// Die drei Standwege daneben (`PartiePausieren`, `PartieFortsetzen`,
// `PartieBeenden`) bewegen sie *innerhalb* ihres Lebens; diese vier
// beenden es, verdoppeln es oder tragen es aus dem Programm heraus.
//
// # Wo die Datei bleibt
//
// **Die Datei selbst liegt im Dienst** ([service.Austauschdatei]) —
// hier steht nur, was das Fenster daraus macht. Der Dialog, der nach einem
// Ort fragt, steht auf `Bindungen`; er braucht ein Fenster, und alles
// darunter braucht keines.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Partiedatei legt eine Partie ab und holt sie zurück — die Naht zum Dienst.
type Partiedatei interface {
	Schreibe(pfad string, inhalt []byte) error
	Lies(pfad string) ([]byte, error)
}

// PartieverwaltungAntwort ist der Ausgang eines dieser vier Wege.
type PartieverwaltungAntwort struct {
	// Kennung ist die Partie, die dabei entstanden ist — leer beim Löschen.
	Kennung string `json:"kennung"`

	// Pfad ist der Ort der Datei — leer, wo keine geschrieben wurde.
	Pfad string `json:"pfad"`

	// Inhalt ist die Partie als Text — **nur** dort, wo keine Datei im
	// Spiel ist.
	//
	// # Warum zwei Gestalten und nicht eine
	//
	// **Weil die vier Flächen nicht dieselbe Hand haben** (§16). Fenster und
	// Kommandozeile stehen an einer Platte: dort ist ein **Ort** die
	// Antwort, und der Inhalt wäre ein Umweg durch den Arbeitsspeicher.
	// REST und MCP stehen an keiner: dort ist der **Text** die Antwort, und
	// ein Pfad wäre einer auf einem fremden Rechner.
	Inhalt string `json:"inhalt"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat.
	//
	// **Und das ist kein Fehler** — dieselbe Unterscheidung wie bei
	// `SetSpeichernUnter`: wer den Dialog wegklickt, hat sich entschieden.
	Abgebrochen bool `json:"abgebrochen"`

	// Befunde sind Auffälligkeiten der eingelesenen Datei.
	//
	// **Ein Befund ist kein Fehler** (§8): ein fremdes Feld macht die Partie
	// nicht unbenutzbar, und es wird trotzdem **genannt** statt verschluckt.
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// PartieLoeschen nimmt eine Partie aus dem Bestand.
//
// **Ob sie gelöscht werden darf, entscheidet die Domäne**
// ([domain.Loeschbar]) — hier steht kein `if` über einen Partiestand (§2).
func (a *App) PartieLoeschen(
	ctx context.Context, kennung string,
) PartieverwaltungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	if fehler := a.partienaht().Loeschen(ctx, kennung); fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	return PartieverwaltungAntwort{Befunde: []string{}}
}

// PartieKopieren legt eine Partie als neuen Übungszustand daneben.
func (a *App) PartieKopieren(
	ctx context.Context, kennung string,
) PartieverwaltungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	kopie, fehler := a.partienaht().Kopieren(ctx, kennung, NeuePartiekennung())
	if fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	return PartieverwaltungAntwort{Kennung: kopie.ID, Befunde: []string{}}
}

// PartieAusfuhrSchreiben legt eine Partie als Datei an einen genannten Ort.
func (a *App) PartieAusfuhrSchreiben(
	ctx context.Context, kennung, pfad string,
) PartieverwaltungAntwort {
	ctx = a.imProfil(ctx)

	roh, fehler := a.ausgefuehrtePartie(ctx, kennung)
	if fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	if fehler := a.partiedatei.Schreibe(pfad, roh); fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	return PartieverwaltungAntwort{
		Kennung: kennung, Pfad: pfad, Befunde: []string{},
	}
}

// PartieAusfuhrText gibt eine Partie als Text — für die Flächen ohne Platte.
func (a *App) PartieAusfuhrText(
	ctx context.Context, kennung string,
) PartieverwaltungAntwort {
	ctx = a.imProfil(ctx)

	roh, fehler := a.ausgefuehrtePartie(ctx, kennung)
	if fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	return PartieverwaltungAntwort{
		Kennung: kennung, Inhalt: string(roh), Befunde: []string{},
	}
}

// ausgefuehrtePartie ist der gemeinsame Teil der beiden Ausfuhrwege.
func (a *App) ausgefuehrtePartie(
	ctx context.Context, kennung string,
) ([]byte, error) {
	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return nil, fehler
	}

	return a.partienaht().Ausfuhr(ctx, kennung)
}

// PartieEinfuhrLesen liest eine Partie aus einer genannten Datei.
func (a *App) PartieEinfuhrLesen(
	ctx context.Context, pfad string,
) PartieverwaltungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	roh, fehler := a.partiedatei.Lies(pfad)
	if fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	antwort := a.eingelesenePartie(ctx, roh)
	if antwort.Fehler == nil {
		antwort.Pfad = pfad
	}

	return antwort
}

// PartieEinfuhrText liest eine Partie aus einem Text — für die Flächen ohne
// Platte.
func (a *App) PartieEinfuhrText(
	ctx context.Context, inhalt string,
) PartieverwaltungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	return a.eingelesenePartie(ctx, []byte(inhalt))
}

// eingelesenePartie ist der gemeinsame Teil der beiden Einfuhrwege.
func (a *App) eingelesenePartie(
	ctx context.Context, roh []byte,
) PartieverwaltungAntwort {
	partie, befunde, fehler := a.partienaht().Einfuhr(
		ctx, roh, NeuePartiekennung())
	if fehler != nil {
		return leereVerwaltung(ctx, fehler)
	}

	return PartieverwaltungAntwort{
		Kennung: partie.ID, Befunde: gruendeAus(befunde),
	}
}

// MitPartiedatei setzt die Naht zur Datei.
func (a *App) MitPartiedatei(datei Partiedatei) {
	a.partiedatei = datei
}

// partiedateiDienst gibt den Dienst, den die Composition Root einsetzt.
func partiedateiDienst() Partiedatei {
	return service.NeuePartieaustauschdatei()
}

// leereVerwaltung ist die Antwort, die nichts bewirkt hat.
func leereVerwaltung(
	ctx context.Context, fehler error,
) PartieverwaltungAntwort {
	return PartieverwaltungAntwort{
		Befunde: []string{},
		Fehler:  Uebersetze(ctx, fehler),
	}
}

// gruendeAus macht aus den Befunden ihre Sätze.
//
// **Eine leere Liste und kein `nil`** — „nichts war auffällig" ist eine
// Antwort, und sie sieht anders aus als keine.
func gruendeAus(befunde []domain.Einfuhrbefund) []string {
	gruende := make([]string, 0, len(befunde))

	for _, befund := range befunde {
		gruende = append(gruende, befund.Grund)
	}

	return gruende
}
