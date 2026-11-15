package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// BefehlsergebnisAntwort ist, was ein Ruf bringt: den Text des Werkzeugs
// und seine Dauer — oder den übersetzten Fehler.
type BefehlsergebnisAntwort struct {
	Name    string  `json:"name"`
	Text    string  `json:"text"`
	DauerMs int64   `json:"dauerMs"`
	Fehler  *Fehler `json:"fehler"`
}

// BefehlAusfuehren ruft einen Befehl mit seinen Argumenten — als JSON-Objekt,
// so wie das Werkzeug sie über MCP bekäme.
//
// **Die Argumente bleiben Daten.** Es gibt keine Zeichenkette, die irgendwo
// zerteilt würde: was hier ankommt, geht als Rohtext an das Werkzeug, und das
// liest es als JSON. Ein Semikolon ist ein Semikolon.
func (a *App) BefehlAusfuehren(ctx context.Context, name, argumente string, bestaetigt bool) BefehlsergebnisAntwort {
	ctx = a.imProfil(ctx)

	befehl, gefunden := a.befehlNamens(name)
	if !gefunden {
		return BefehlsergebnisAntwort{Name: name, Fehler: Uebersetze(ctx, fmt.Errorf("%w: unbekannter Befehl %q", domain.ErrNotFound, name))}
	}

	// **Die Schalter des Werkzeugs, nicht die des Kernwegs** (Befund 1 des
	// Gesamtreviews über 0.6.0, #1660): das Bindungstor sieht nur
	// `BefehlAusfuehren`, und der Befehl ruft das Werkzeug an seinem
	// MCP-Server vorbei. Gefragt wird **unmittelbar vor dem Ruf** — ein
	// Schalter kann zwischen Auswahl und Klick gefallen sein.
	if fehler := a.befehlsschalter().Pruefe(ctx, domain.FlaecheMCP, name); fehler != nil {
		return BefehlsergebnisAntwort{Name: name, Fehler: Uebersetze(ctx, fehler)}
	}

	roh, err := argumenteAls(argumente)
	if err != nil {
		return BefehlsergebnisAntwort{Name: name, Fehler: Uebersetze(ctx, err)}
	}

	if befehl.Schreibend {
		if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
			return BefehlsergebnisAntwort{Name: name, Fehler: Uebersetze(ctx, verweigert)}
		}
	}

	if gefaehrlich(befehl) && !bestaetigt {
		return BefehlsergebnisAntwort{Name: name, Fehler: Uebersetze(ctx, domain.NewFieldError(
			"bestaetigt", domain.SchluesselBefehlBestaetigen,
			"%s ändert etwas, das sich nicht ohne Weiteres zurücknehmen lässt — erst die Vorschau lesen, dann bestätigen", name))}
	}

	lauf, abbrechen, err := a.befehlBeginnen(ctx)
	if err != nil {
		return BefehlsergebnisAntwort{Name: name, Fehler: Uebersetze(ctx, err)}
	}

	defer a.befehlBeenden(abbrechen)

	begonnen := time.Now()
	text, ruffehler := befehl.Ruf(lauf, roh)
	dauer := time.Since(begonnen).Milliseconds()

	if ruffehler != nil {
		return BefehlsergebnisAntwort{Name: name, DauerMs: dauer, Fehler: Uebersetze(ctx, ruffehler)}
	}

	return BefehlsergebnisAntwort{Name: name, Text: text, DauerMs: dauer}
}

// BefehlAbbrechen bricht den laufenden Befehl ab — ohne laufenden ist es nichts.
func (a *App) BefehlAbbrechen(_ context.Context) SchreibAntwort {
	a.befehlsriegel.Lock()
	defer a.befehlsriegel.Unlock()

	if a.befehlAbbrechen != nil {
		a.befehlAbbrechen()
	}

	return SchreibAntwort{}
}

func (a *App) befehlNamens(name string) (Befehl, bool) {
	for _, befehl := range a.befehle {
		if befehl.Name == name {
			return befehl, true
		}
	}

	return Befehl{}, false
}

// argumenteAls nimmt die Argumente als JSON-Objekt — leer heißt leeres Objekt.
func argumenteAls(argumente string) (json.RawMessage, error) {
	argumente = strings.TrimSpace(argumente)
	if argumente == "" {
		return json.RawMessage(`{}`), nil
	}

	if !json.Valid([]byte(argumente)) || !strings.HasPrefix(argumente, "{") {
		return nil, domain.NewFieldError("argumente", domain.SchluesselBefehlArgumente, "die Argumente sind kein JSON-Objekt")
	}

	return json.RawMessage(argumente), nil
}

// befehlBeginnen belegt den einen Platz — oder sagt, dass einer läuft.
func (a *App) befehlBeginnen(ctx context.Context) (context.Context, context.CancelFunc, error) {
	a.befehlsriegel.Lock()
	defer a.befehlsriegel.Unlock()

	if a.befehlAbbrechen != nil {
		return nil, nil, fmt.Errorf("%w: ein Befehl läuft noch — erst abwarten oder abbrechen", domain.ErrConflict)
	}

	lauf, abbrechen := context.WithCancel(ctx)
	a.befehlAbbrechen = abbrechen

	return lauf, abbrechen, nil
}

// befehlBeenden gibt den Platz frei — und den Abbruch, den niemand mehr braucht.
func (a *App) befehlBeenden(abbrechen context.CancelFunc) {
	a.befehlsriegel.Lock()
	defer a.befehlsriegel.Unlock()

	abbrechen()
	a.befehlAbbrechen = nil
}

// BefehlAusfuehren ist die Fensterbindung.
func (b *Bindungen) BefehlAusfuehren(ctx context.Context, name, argumente string, bestaetigt bool) BefehlsergebnisAntwort {
	return b.app(ctx).BefehlAusfuehren(b.ruf(ctx), name, argumente, bestaetigt)
}

// BefehlAbbrechen ist die Fensterbindung.
func (b *Bindungen) BefehlAbbrechen(ctx context.Context) SchreibAntwort {
	return b.app(ctx).BefehlAbbrechen(b.ruf(ctx))
}

// befehlsschalter ist die Wache der Kommandozentrale — dieselbe wie am
// Bindungstor, und wie dort beendet sie nebenbei, was ein anderer Prozess
// abgeschaltet hat (OPS-10b, #1978).
func (a *App) befehlsschalter() Schalterwache {
	return Schalterwache{einstellungen: a.einstellungen, gelesen: a.laeufeBeenden}
}
