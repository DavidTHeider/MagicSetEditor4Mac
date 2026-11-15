// **Die Vorschaubilder der Kartenliste** — Werkstück 2 von Phase 3.5
// (#1166).
//
// # Warum ein Stapel und kein Bild je Ruf
//
// Die Liste holt sie für ihr **Sichtfenster** — ein- bis zweiunddreißig
// Zeilen —, und ein Ruf je Zeile wären zweiunddreißig Wege über die
// Wails-Naht für eine Bewegung des Rollbalkens. MSE2 hat dafür einen
// Hintergrundstrang (`gui/thumbnail_thread.cpp`); hier ist es ein Aufruf.
//
// # Und warum die Kennungen und nicht die Dateinamen
//
// Ein Dateiname aus dem Fenster wäre eine **Eingabe** (§8), und diese Naht
// gäbe damit jede lesbare Datei dieses Rechners heraus. Was hereinkommt, sind
// Kartenkennungen; welchen Namen eine Karte trägt, entscheidet der Bestand.

package app

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kartenbildantwort trägt die Bilder als Daten-URI, je Kartenkennung.
//
// **Eine Tabelle und keine Liste.** Eine Karte ohne Bild steht gar nicht
// darin — der Normalfall, und eine Liste voller leerer Zeichenketten wäre die
// teurere Art, dasselbe zu sagen.
type Kartenbildantwort struct {
	Bilder map[string]string `json:"bilder"`
	Fehler *Fehler           `json:"fehler"`
}

// Kartenbilder gibt zu jeder genannten Karte ein kleines Bild.
//
// **Ein Bild, das nicht zu holen ist, fehlt einfach** (§8). Eine Liste, die
// wegen einer kaputten Datei gar nichts zeigte, wäre der schlechtere Tausch —
// und der Zeichenweg meldet denselben Fehler dort, wo er etwas bedeutet.
func (a *App) Kartenbilder(
	ctx context.Context, setID string, karteIDs []string,
) Kartenbildantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return Kartenbildantwort{Bilder: map[string]string{}, Fehler: befund}
	}

	if a.vorschaubilder == nil || len(karteIDs) == 0 {
		return Kartenbildantwort{Bilder: map[string]string{}}
	}

	feld := erstesBildfeld(a.bildfelderDesSets(ctx, kennung))
	if feld == "" {
		return Kartenbildantwort{Bilder: map[string]string{}}
	}

	karten, err := a.leser().Karten(ctx, kennung)
	if err != nil {
		return Kartenbildantwort{Bilder: map[string]string{}, Fehler: Uebersetze(ctx, err)}
	}

	return Kartenbildantwort{
		Bilder: a.bilderZu(karten, gefragte(karteIDs), feld, a.setort(ctx, kennung),
			a.ablagenort(kennung)),
	}
}

// erstesBildfeld gibt das Feld, dessen Bild die Liste zeigt.
//
// **Das erste und nicht irgendeines** — MSE2 nimmt ebenfalls das erste:
//
//	FOR_EACH(f, set->game->card_fields) {
//	  ImageFieldP imgf = dynamic_pointer_cast<ImageField>(f);
//	  if (imgf) return imgf;
//	}
//
// `ImageCardList::findImageField`, `src/gui/control/image_card_list.cpp`,
// Zeilen 42–48, Fassung `5deaeee` vom 2020-05-16.
//
// **Und `bildfelderDesSets` gibt sie in der Reihenfolge der Vorlage**, also
// in derselben, in der MSE2 sie durchläuft.
func erstesBildfeld(felder []string) string {
	if len(felder) == 0 {
		return ""
	}

	return felder[0]
}

// gefragte macht aus der Kennungsliste eine Menge.
func gefragte(karteIDs []string) map[string]bool {
	menge := make(map[string]bool, len(karteIDs))
	for _, id := range karteIDs {
		menge[id] = true
	}

	return menge
}

// bilderZu holt die Bilder der gefragten Karten.
func (a *App) bilderZu(
	karten []*domain.CardRow, gefragt map[string]bool, feld, setpaket, ablage string,
) map[string]string {
	bilder := make(map[string]string, len(gefragt))

	for _, karte := range karten {
		id := strconv.FormatInt(karte.ID, 10)
		if !gefragt[id] {
			continue
		}

		bild, err := a.vorschaubilder.Kartenbild(wertVon(karte, feld), setpaket, ablage)
		if err != nil || bild == nil {
			continue
		}

		uri, kodierfehler := a.alsDatenURI(bild)
		if kodierfehler != nil {
			continue
		}

		bilder[id] = uri
	}

	return bilder
}

// wertVon liest einen Feldwert einer Karte.
func wertVon(karte *domain.CardRow, feld string) string {
	for _, wert := range karte.Values {
		if wert.Field == feld {
			return strings.TrimSpace(wert.Text)
		}
	}

	return ""
}

// ablagenort gibt den Ort der Bilder, die jemand in dieser Anwendung
// eingesetzt hat — leer, wenn es keinen gibt.
//
// **Ohne Ablage ist es kein Fehler.** Dann liegt jedes Bild im Herkunftspaket,
// und `vorschauverweis` fragt dort.
//
// **Kein Wächter auf `a.bilderort == nil`**, obwohl er naheläge: [NeueApp]
// setzt die Naht, und `App.Bilderort` ruft sie seit jeher ungeschützt. Ein
// Zweig, den kein Aufruf betritt, ist eine Behauptung darüber, was geschieht,
// und niemand prüft sie nach (§3).
func (a *App) ablagenort(setID int64) string {
	wurzel, err := a.bilderort()
	if err != nil {
		return ""
	}

	return service.NeueBildablage(wurzel).Verzeichnis(setID)
}

// MitVorschaubildern setzt das Werk für die Listenbilder — für Tests und für
// den Composition Root.
func (a *App) MitVorschaubildern(werk *service.Vorschaubildwerk) {
	a.vorschaubilder = werk
}
