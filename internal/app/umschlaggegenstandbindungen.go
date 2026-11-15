package app

// **Die Bindungen der vier weiteren Arten** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// **Vier Bindungen seit dem 2026-09-19.** Ein Deck wird angenommen, eine
// Sammlungssicht wird **angesehen** — und Karte und Vorlage brauchen ein
// Ziel, das ein Mensch wählt: ein Set und einen Paketnamen. Das steht jetzt
// **im Ruf** (#1639);
// bis dahin gingen beide hinaus und wurden drüben nur angesehen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

// UmschlagDeck nimmt ein Deck aus einem Umschlag an.
//
// **Schreibend** (`docs/SCHREIBWEGE.md`) — es entsteht ein **neues** Deck.
func (b *Bindungen) UmschlagDeck(ctx context.Context) DeckAntwort {
	ctx = b.ruf(ctx)

	pfad, abgebrochen, fehler := umschlagdatei(ctx, b.app(ctx))
	if fehler != nil {
		return DeckAntwort{Fehler: fehler}
	}

	if abgebrochen {
		return DeckAntwort{}
	}

	return b.app(ctx).UmschlagDeck(ctx, pfad)
}

// UmschlagSicht sieht eine empfangene Sammlungssicht an.
//
// **Rein lesend** — im Bestand ändert sich nichts, und das ist die Sache: ein
// Empfang ist keine Buchung.
func (b *Bindungen) UmschlagSicht(ctx context.Context) FreigabesichtAntwort {
	ctx = b.ruf(ctx)

	leer := FreigabesichtAntwort{Zeilen: []SichtkarteDTO{}}

	pfad, abgebrochen, fehler := umschlagdatei(ctx, b.app(ctx))
	if fehler != nil {
		leer.Fehler = fehler

		return leer
	}

	if abgebrochen {
		return leer
	}

	return b.app(ctx).UmschlagSicht(ctx, pfad)
}

// umschlagdatei fragt nach einer Umschlagdatei.
//
// **Geteilt, weil beide Bindungen dieselbe Frage stellen.** Zwei Fassungen
// liefen mit dem Dialogtext auseinander.
//
// **Und als freie Funktion und nicht als Verfahren an [Bindungen]**: was dort
// hängt, ist eine **gebundene Methode** — der Wächter zählt sie, und eine
// Hilfe, die wie eine Fähigkeit aussieht, wäre eine Zusage ohne Einlösung.
func umschlagdatei(
	ctx context.Context, anwendung *App,
) (pfad string, abgebrochen bool, fehler *Fehler) {
	texte := DialogtexteFuer(anwendung.sprache)

	gewaehlt, wahlfehler := anwendung.waehleDatei(ctx, Dateiauswahl{
		Titel:  texte.UmschlagOeffnen,
		Filter: texte.UmschlagFilter,
		Muster: "*" + austausch.Endung,
	})
	if wahlfehler != nil {
		return "", false, Uebersetze(ctx, wahlfehler)
	}

	// **Ein leerer Pfad ist ein Abbruch** — derselbe Grund wie bei
	// [Bindungen.UmschlagWaehlen].
	return gewaehlt, gewaehlt == "", nil
}

// UmschlagKarte nimmt eine Karte aus einem Umschlag in ein **gewähltes** Set.
//
// **Schreibend** (`docs/SCHREIBWEGE.md`) — es entsteht eine **neue** Karte.
// Die Kennung aus dem Umschlag ist hier keine: sie stammt aus einem fremden
// Bestand (#1639).
func (b *Bindungen) UmschlagKarte(ctx context.Context, setID string) AnlegeAntwort {
	ctx = b.ruf(ctx)

	pfad, abgebrochen, fehler := umschlagdatei(ctx, b.app(ctx))
	if fehler != nil {
		return AnlegeAntwort{Fehler: fehler}
	}

	if abgebrochen {
		return AnlegeAntwort{}
	}

	return b.app(ctx).UmschlagKarte(ctx, pfad, setID)
}

// UmschlagVorlage nimmt eine Vorlage aus einem Umschlag als **gewähltes**
// Paket an.
//
// **Schreibend** — es entsteht ein **neues** Paket, und ein vorhandenes ist
// ein Konflikt und kein Überschreiben (#1639).
func (b *Bindungen) UmschlagVorlage(ctx context.Context, name string) PaketAntwort {
	ctx = b.ruf(ctx)

	pfad, abgebrochen, fehler := umschlagdatei(ctx, b.app(ctx))
	if fehler != nil {
		return PaketAntwort{Fehler: fehler}
	}

	if abgebrochen {
		return PaketAntwort{}
	}

	return b.app(ctx).UmschlagVorlage(ctx, pfad, name)
}
