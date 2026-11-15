package app

// **Ein Turnier als Datei und als Bericht** — DOM-04.AK-03
// (#1699),
// Vermerk [0140](../../docs/entscheidungen/0140-ein-turnier-als-datei.md).
//
// **Zwei Gestalten je Weg**, wie bei der Partiedatei: Fenster und
// Kommandozeile stehen an einer Platte und bekommen einen **Ort**, REST und
// MCP an keiner und bekommen den **Text** (§16). Die Datei selbst liegt im
// Dienst ([service.Austauschdatei]); der Dialog steht auf `Bindungen`.

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TurnierdateiAntwort ist der Ausgang von Ausgabe, Einlesen und Bericht.
type TurnierdateiAntwort struct {
	// Kennung ist das Turnier — das ausgegebene oder das eingelesene.
	Kennung string `json:"kennung"`

	// Pfad ist der Ort der Datei — leer, wo keine im Spiel ist.
	Pfad string `json:"pfad"`

	// Inhalt ist die Datei oder der Bericht als Text — **nur** dort, wo
	// keine Datei im Spiel ist.
	Inhalt string `json:"inhalt"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat — kein Fehler.
	Abgebrochen bool `json:"abgebrochen"`

	// Befunde sind Auffälligkeiten der eingelesenen Datei (§8).
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// errOhneTurnierdatei ist das Nein einer Anwendung, der niemand die Datei
// gesetzt hat — ein Fehler des Aufbaus und nicht der Eingabe.
var errOhneTurnierdatei = errors.New("hier ist keine Turnierdatei eingerichtet")

// MitTurnierdatei setzt die Naht zur Datei.
func (a *App) MitTurnierdatei(datei Partiedatei) { a.turnierdatei = datei }

// TurnierAusgabeSchreiben legt ein Turnier als Datei an einen genannten Ort.
func (a *App) TurnierAusgabeSchreiben(ctx context.Context, turnierID, pfad string) TurnierdateiAntwort {
	ctx = a.imProfil(ctx)

	return a.turnierdateiSchreiben(ctx, turnierID, pfad, a.ausgabe)
}

// TurnierAusgabeText gibt ein Turnier als Text.
func (a *App) TurnierAusgabeText(ctx context.Context, turnierID string) TurnierdateiAntwort {
	ctx = a.imProfil(ctx)

	return a.turnierdateiText(ctx, turnierID, a.ausgabe)
}

// TurnierBerichtSchreiben legt den Bericht an einen genannten Ort.
func (a *App) TurnierBerichtSchreiben(ctx context.Context, turnierID, pfad string) TurnierdateiAntwort {
	ctx = a.imProfil(ctx)

	return a.turnierdateiSchreiben(ctx, turnierID, pfad, a.bericht)
}

// TurnierBerichtText gibt den Bericht als Text.
func (a *App) TurnierBerichtText(ctx context.Context, turnierID string) TurnierdateiAntwort {
	ctx = a.imProfil(ctx)

	return a.turnierdateiText(ctx, turnierID, a.bericht)
}

// TurnierEinlesenLesen liest ein Turnier aus einer genannten Datei.
func (a *App) TurnierEinlesenLesen(ctx context.Context, pfad string) TurnierdateiAntwort {
	ctx = a.imProfil(ctx)

	if a.turnierdatei == nil {
		return leereTurnierdatei(ctx, errOhneTurnierdatei)
	}

	roh, err := a.turnierdatei.Lies(pfad)
	if err != nil {
		return leereTurnierdatei(ctx, err)
	}

	antwort := a.eingelesenesTurnier(ctx, roh)
	if antwort.Fehler == nil {
		antwort.Pfad = pfad
	}

	return antwort
}

// TurnierEinlesenText liest ein Turnier aus einem Text.
func (a *App) TurnierEinlesenText(ctx context.Context, inhalt string) TurnierdateiAntwort {
	ctx = a.imProfil(ctx)

	return a.eingelesenesTurnier(ctx, []byte(inhalt))
}

// eingelesenesTurnier ist der gemeinsame Teil beider Einlesewege.
func (a *App) eingelesenesTurnier(ctx context.Context, roh []byte) TurnierdateiAntwort {
	turnier, befunde, err := a.turniere.Einlesen(ctx, roh)
	if err != nil {
		return leereTurnierdatei(ctx, err)
	}

	return TurnierdateiAntwort{Kennung: kennungtext(turnier.ID), Befunde: gruendeAus(befunde)}
}

// turnierdateiweg erzeugt den Inhalt zu einem Turnier — Ausgabe oder Bericht.
type turnierdateiweg func(ctx context.Context, turnierID int64) ([]byte, error)

func (a *App) ausgabe(ctx context.Context, turnierID int64) ([]byte, error) {
	return a.turniere.Ausgeben(ctx, turnierID)
}

// bericht schreibt ihn **in der Sprache dieser Anwendung** — der des
// Fensters; ohne gewählte Deutsch, wie an der Kommandozeile.
func (a *App) bericht(ctx context.Context, turnierID int64) ([]byte, error) {
	text, err := a.turniere.Bericht(ctx, turnierID, BerichttexteFuer(a.sprache))

	return []byte(text), err
}

// turnierdateiText gibt den Inhalt als Text.
func (a *App) turnierdateiText(ctx context.Context, turnierID string, weg turnierdateiweg) TurnierdateiAntwort {
	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return TurnierdateiAntwort{Befunde: []string{}, Fehler: befund}
	}

	inhalt, err := weg(ctx, kennung)
	if err != nil {
		return leereTurnierdatei(ctx, err)
	}

	return TurnierdateiAntwort{Kennung: turnierID, Inhalt: string(inhalt), Befunde: []string{}}
}

// turnierdateiSchreiben legt den Inhalt an den Ort.
func (a *App) turnierdateiSchreiben(
	ctx context.Context, turnierID, pfad string, weg turnierdateiweg,
) TurnierdateiAntwort {
	antwort := a.turnierdateiText(ctx, turnierID, weg)
	if antwort.Fehler != nil {
		return antwort
	}

	if a.turnierdatei == nil {
		return leereTurnierdatei(ctx, errOhneTurnierdatei)
	}

	if err := a.turnierdatei.Schreibe(pfad, []byte(antwort.Inhalt)); err != nil {
		return leereTurnierdatei(ctx, err)
	}

	return TurnierdateiAntwort{Kennung: turnierID, Pfad: pfad, Befunde: []string{}}
}

// leereTurnierdatei ist die Antwort auf einen Fehler.
func leereTurnierdatei(ctx context.Context, err error) TurnierdateiAntwort {
	return TurnierdateiAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, err)}
}

// berichttexte hält die Wörter des Berichts je Sprache — **Deutsch ist die
// Ausgangssprache** (§9), und eine unbekannte fällt auf sie zurück.
var berichttexte = map[string]domain.Berichttexte{
	"de": {
		Spiel: "Spiel", Format: "Format", Stand: "Stand", Abgeschlossen: "abgeschlossen am",
		Laufend: "läuft", Teilnehmende: "Teilnehmende", Ausgestiegen: "ausgestiegen", Runden: "Runden",
		Rangliste: "Rangliste", Platz: "Platz", Name: "Name", Punkte: "Punkte", Bilanz: "Bilanz",
		Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
		Runde: "Runde", Tisch: "Tisch", SeiteA: "Seite A", SeiteB: "Seite B", Ergebnis: "Ergebnis",
		Bye: "Freilos", Offen: "offen",
	},
	"en": {
		Spiel: "Game", Format: "Format", Stand: "Status", Abgeschlossen: "closed on",
		Laufend: "running", Teilnehmende: "Players", Ausgestiegen: "dropped", Runden: "Rounds",
		Rangliste: "Standings", Platz: "Rank", Name: "Name", Punkte: "Points", Bilanz: "Record",
		Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
		Runde: "Round", Tisch: "Table", SeiteA: "Player A", SeiteB: "Player B", Ergebnis: "Result",
		Bye: "Bye", Offen: "open",
	},
	"es": berichttexteEs,
	"fr": berichttexteFr,
	"it": berichttexteIt,
	"ja": berichttexteJa,
	"ru": berichttexteRu,
	"zh": berichttexteZh,
}

// BerichttexteFuer gibt die Wörter des Berichts in einer Sprache.
func BerichttexteFuer(sprache string) domain.Berichttexte {
	if texte, da := berichttexte[sprache]; da {
		return texte
	}

	return berichttexte["de"]
}
