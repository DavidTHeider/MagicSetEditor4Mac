package app

// **Eine Einladung genügt** — NET-01b
// (#1954),
// entschieden in
// [0156](../../docs/entscheidungen/0156-eine-einladung-genuegt.md).
//
// Hier steht nur die Übersetzung: welcher Befund welcher [Einloesemeldung]
// entspricht, und welche Sitzung wohin geht. **Was eingelöst wird und wann,
// entscheidet `service`.**

import (
	"context"
	"errors"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die drei Stände einer Meldung beim Einladenden.
const (
	// MeldungGemeldet heißt: der Einladende kennt dieses Profil jetzt.
	MeldungGemeldet = "gemeldet"

	// MeldungOhne heißt: die Einladung ist nicht einlösbar — versucht wurde
	// nichts.
	MeldungOhne = "ohne"

	// MeldungGescheitert heißt: versucht und nicht geglückt.
	MeldungGescheitert = "gescheitert"
)

// Einloesemeldung sagt, ob sich dieses Profil beim Einladenden melden
// konnte.
//
// **Neben der Liste und nicht an ihrer Stelle**: die Annahme ist
// geschrieben, was auch immer hier steht (§7, Regel 3).
type Einloesemeldung struct {
	Stand string `json:"stand"`

	// Grund ist das Ergebnis der Einlösung — der Schlüssel, unter dem das
	// Fenster seinen Satz findet. Leer, wenn gemeldet.
	Grund string `json:"grund,omitempty"`

	// Meldung ist der deutsche Satz für die Flächen ohne Katalog.
	Meldung string `json:"meldung,omitempty"`
}

// alsMeldung übersetzt den Befund des Meldens.
func alsMeldung(ctx context.Context, fehler error) *Einloesemeldung {
	switch {
	case fehler == nil:
		return &Einloesemeldung{Stand: MeldungGemeldet}
	case errors.Is(fehler, domain.ErrEinladungNichtEinloesbar):
		return &Einloesemeldung{
			Stand: MeldungOhne, Meldung: domain.ErrEinladungNichtEinloesbar.Error(),
		}
	}

	// **Der Grund geht ins Protokoll, der Satz an die Fläche** (§8): was
	// über die Leitung scheiterte, sagt nicht dem Menschen, was er tun kann.
	grund := fehler
	if verdeckt := domain.GrundVon(fehler); verdeckt != nil {
		grund = verdeckt
	}

	slog.InfoContext(ctx, "eine Einlösung blieb ohne Zusage", "grund", grund)

	antwort := domain.AntwortAuf(fehler)

	return &Einloesemeldung{
		Stand: MeldungGescheitert, Grund: string(antwort.Ergebnis),
		Meldung: antwort.Fehler().Error(),
	}
}

// einloesebote öffnet eine Paarungssitzung — die, die jeder öffnen darf,
// der nicht gesperrt ist ([domain.Peerkontakt.Darf]).
func (a *App) einloesebote() service.Einloesebote {
	return func(ctx context.Context, kennung peer.Kennung) (service.Einloeseleitung, error) {
		griff, fehler := a.sitzungsverwaltung().Oeffne(ctx, kennung, domain.ZweckPaarung)
		if fehler != nil {
			return nil, fehler
		}

		return griff, nil
	}
}

// einloesungAufnehmen beantwortet eine Paarungssitzung, die hereinkam.
//
// **Eine Nachricht, eine Antwort, dann zu** — die Paarung ist das
// Kennenlernen und sonst nichts.
//
// **Ohne [App.imProfil]** (Zweitblick über Phase 6.7, Z2): der Kontext kommt
// aus dem Horchposten und trägt das Profil, das ihn geöffnet hat. Die
// Umschaltsperre zu nehmen hieße, auf einen Wechsel zu warten, der selbst auf
// den Horchposten wartet.
func (a *App) einloesungAufnehmen(ctx context.Context, griff service.Angenommene) {
	ctx = MitNeuemVorgang(ctx)

	if fehler := a.einladungswerk().EinloesungBeantworten(ctx, griff); fehler != nil {
		slog.InfoContext(ctx, "eine Einlösung wurde nicht aufgenommen",
			"gegenueber", griff.Stand().Fingerabdruck, "grund", fehler)
	}

	// **Auch nach einem Fehler**: wer erst nach dem Eintragen scheitert —
	// beim Antworten —, hat geschrieben, und das Fenster lädt dann einmal zu
	// viel statt einmal zu wenig ([standnaht]).
	a.nebenherGeschrieben()

	if fehler := griff.Schliesse(); fehler != nil {
		slog.InfoContext(ctx, "eine Einlösesitzung ließ sich nicht schließen",
			"grund", fehler)
	}
}
