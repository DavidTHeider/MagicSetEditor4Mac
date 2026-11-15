package app

import (
	"context"
	"errors"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Speicherer schreibt ein Set zurück — die Naht zu `internal/service`.
type Speicherer interface {
	Schreibe(ctx context.Context, setID int64, ziel string) (*service.Speicherung, error)
}

// SpeicherAntwort ist der Ausgang eines Speicherns.
type SpeicherAntwort struct {
	Pfad string `json:"pfad"`

	// Befunde sind die Stellen, an denen etwas **nicht mitgekommen** ist.
	//
	// Die Hauptdatei steht im Bestand und kommt heraus
	// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md));
	// was daneben lag — Bilder, Symbole —, liegt am Herkunftsort und kann
	// fehlen. Dass es fehlt, gehört gesagt und nicht verschwiegen (§8).
	Befunde []string `json:"befunde"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`

	// ohneOrt sagt, dass das Set aus keiner Datei kam — für das Fenster, das
	// dann fragt, wohin (FIX-F02). Nicht in der Antwort: die übrigen Flächen
	// nennen ihr Ziel selbst, und für sie ist es ein Fehler wie bisher.
	ohneOrt bool
}

// SetSpeichern schreibt ein Set als `.mse-set` an einen Ort.
//
// **Ein leeres Ziel ist kein Abbruch, sondern ⌘S.** Anders als bei der Ausfuhr
// gibt es hier einen Ort, den niemand nennen muss: den, aus dem das Set kam.
// Wo er fehlt, sagt der Dienst das — und die Oberfläche fragt dann.
func (a *App) SetSpeichern(ctx context.Context, setID, ziel string) SpeicherAntwort {
	ctx = a.imProfil(ctx)

	leer := SpeicherAntwort{Befunde: []string{}}

	kennung, err := strconv.ParseInt(setID, 10, 64)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	speicherung, err := a.speicherer.Schreibe(ctx, kennung, ziel)

	// **Ein Ergebnis samt Fehler geht samt Fehler hinaus** — Befund 1 des
	// Nachlaufs über Phase 4.4 (#1291). `Setschreiber.Schreibe` gibt die
	// `Speicherung` **und** den Fehler zurück, wenn nur das Setzen des
	// Arbeitsorts scheitert: die Datei liegt dann schon am Ziel, und der Pfad
	// ist das, was der Aufrufer braucht (§7 Regel 3).
	if speicherung == nil {
		leer.Fehler = Uebersetze(ctx, err)
		leer.ohneOrt = errors.Is(err, service.ErrOhneOrt)

		return leer
	}

	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)
	}

	for _, befund := range speicherung.Befunde {
		leer.Befunde = append(leer.Befunde, befund.String())
	}

	leer.Pfad = speicherung.Pfad
	a.arbeitVermerken(ctx, domain.ArbeitsartSet, setID, domain.VorgangAusgegeben)

	return leer
}

// MitSpeicherer setzt die Naht zum Zurückschreiben.
func (a *App) MitSpeicherer(speicherer Speicherer) { a.speicherer = speicherer }
