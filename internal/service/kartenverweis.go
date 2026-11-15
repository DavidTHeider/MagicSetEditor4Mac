package service

// **Der Verweis einer entworfenen Karte auf einen echten Druck** — Werkstück 2
// von Phase 3.6
// (#1167).
//
// # Was dieser Anwendungsfall entscheidet
//
// **Dass ein Verweis aufgelöst wird, bevor er entsteht.** Wer eine UUID
// ablegte, ohne sie nachzuschlagen, hätte einen Verweis, den niemand lesen
// kann — und der sich beim ersten Abgleich als Tippfehler herausstellt, lange
// nachdem ihn jemand gesetzt hat. Deshalb steht das hier und nicht in
// `internal/app`: es ist eine fachliche Regel, und §16 sagt, eine Fähigkeit
// landet nicht an einer Fläche.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremddrucke ist, was der Verweis von einem fremden Dienst braucht.
//
// **Eine eigene Schnittstelle neben [Fremdkarten]**, obwohl derselbe Klient
// beide erfüllt: sie sagt, was **gebraucht** wird (§2). Wer einen Verweis
// setzt, sucht keinen Namen — er kennt den Druck schon.
type Fremddrucke interface {
	KarteNachKennung(ctx context.Context, kennung string) (
		karte *scryfall.Karte, alter time.Duration,
		befunde []domain.Befund, fehler error)
}

// Verweisablage ist, was der Verweis vom Bestand braucht.
type Verweisablage interface {
	Setzen(ctx context.Context, karteID int64, verweis domain.Kartenverweis) error
	Fuer(ctx context.Context, karteID int64) (*domain.Kartenverweis, error)
	Loesen(ctx context.Context, karteID int64) (bool, error)
	JeSet(ctx context.Context, setID int64) (map[int64]domain.Kartenverweis, error)
}

// Kartenverweise verbindet entworfene Karten mit echten Drucken.
type Kartenverweise struct {
	fremd  Fremddrucke
	ablage Verweisablage
	uhr    func() time.Time

	// sets sagt, ob es ein Set gibt
	// (#1336).
	//
	// **Wahlfrei**: ohne sie antwortet [Kartenverweise.JeSet] wie bisher. Der
	// Grund ist der Aufrufer, der sie nicht hat — und ein Pflichtargument
	// zwänge jeden Test, einen Setleser zu bauen, um einen Verweis zu prüfen.
	sets Setkopfleser
}

// MitSetwache lässt die Verweise eine fehlende Setkennung abweisen.
func (k *Kartenverweise) MitSetwache(sets Setkopfleser) *Kartenverweise {
	k.sets = sets

	return k
}

// NeueKartenverweise baut den Anwendungsfall.
//
// **Die Uhr wird eingesetzt, nicht genommen** (§16): ohne sie bräuchte jeder
// Test die Wanduhr, und der Zeitpunkt eines Verweises wäre nicht prüfbar.
func NeueKartenverweise(
	fremd Fremddrucke, ablage Verweisablage, uhr func() time.Time,
) *Kartenverweise {
	return &Kartenverweise{fremd: fremd, ablage: ablage, uhr: uhr}
}

// Setzen bezieht eine Karte auf einen Druck.
//
// **Der Druck wird geholt, bevor der Verweis entsteht** — und die Anzeige
// kommt aus seiner Antwort, nicht aus der Eingabe. Ein Aufrufer, der Name und
// Setcode selbst mitgäbe, könnte einen Verweis anlegen, der auf `lea` zeigt
// und `2ed` behauptet.
//
// **Die Befunde reisen mit** (§8): ein unbekannter Aufzählungswert soll den
// sehen, der den Verweis setzt.
func (k *Kartenverweise) Setzen(ctx context.Context, karteID int64, printing string) (
	*domain.Kartenverweis, []domain.Befund, error,
) {
	if strings.TrimSpace(printing) == "" {
		return nil, nil, fmt.Errorf(
			"%w: ohne Kennung ist kein Druck gemeint", domain.ErrValidation)
	}

	karte, alter, befunde, fehler := k.fremd.KarteNachKennung(ctx, printing)
	if fehler != nil {
		return nil, nil, fehler
	}

	// **Der Zeitpunkt ist der des Standes, nicht der des Klicks** — Befund 9
	// des Zweitblicks über Phase 3.6.
	//
	// `KarteNachKennung` gibt ausdrücklich das **Alter** der Antwort zurück:
	// ohne Netz kommt der letzte gute Stand aus dem Zwischenspeicher (§16,
	// Regel 5). Die Oberfläche zeigt den Zeitpunkt als „bezogen am", und das
	// ist eine Aussage über das **Alter der Anzeige**.
	verweis := domain.Kartenverweis{
		Printing: karte.Kennung, Name: karte.Name,
		Setcode: karte.Set, Nummer: karte.Nummer,
		Adresse: karte.Adresse, Gesetzt: k.uhr().Add(-alter),
	}

	verweis.Normalisieren()

	// **Der erste Befund geht hinaus, nicht alle** — dieselbe Regel wie bei
	// der Sammlung (`vorbereiten`), und aus demselben Grund: `app.Fehler`
	// trägt ein Feld und einen Satz.
	if mangel := verweis.Validate(); len(mangel) > 0 {
		return nil, befunde, mangel[0]
	}

	if fehler := k.ablage.Setzen(ctx, karteID, verweis); fehler != nil {
		return nil, befunde, fehler
	}

	return &verweis, befunde, nil
}

// Loesen nimmt den Verweis einer Karte zurück.
//
// **Wo keiner war, ist es ein Nein.** Ein stiller Erfolg ließe niemanden
// merken, dass er die falsche Karte erwischt hat.
func (k *Kartenverweise) Loesen(ctx context.Context, karteID int64) error {
	geloest, fehler := k.ablage.Loesen(ctx, karteID)
	if fehler != nil {
		return fehler
	}

	if !geloest {
		return fmt.Errorf("%w: die Karte %d hat keinen Verweis",
			domain.ErrNotFound, karteID)
	}

	return nil
}

// Fuer gibt den Verweis einer Karte — `nil`, wenn sie keinen hat.
func (k *Kartenverweise) Fuer(ctx context.Context, karteID int64) (
	*domain.Kartenverweis, error,
) {
	return k.ablage.Fuer(ctx, karteID)
}

// JeSet gibt die Verweise aller Karten eines Sets.
//
// **Ein Set, das es nicht gibt, ist ein Fehler und keine leere Antwort**
// (#1336). Ohne diese Frage bekämen „dieses Set hat keine Verweise" und
// „dieses Set gibt es nicht" dieselbe Antwort — und wer ein Skript über die
// Setkennungen schreibt, merkte einen Tippfehler nie.
func (k *Kartenverweise) JeSet(ctx context.Context, setID int64) (
	map[int64]domain.Kartenverweis, error,
) {
	if k.sets != nil {
		if fehler := SetDaSein(ctx, k.sets, setID); fehler != nil {
			return nil, fehler
		}
	}

	return k.ablage.JeSet(ctx, setID)
}
