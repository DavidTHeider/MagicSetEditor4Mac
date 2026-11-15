package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

// **Die Bindungen für den portablen Umschlag** — Werkstück N4 von Phase 5.4
// (#1457).

// UmschlagSchnueren fragt nach einem Ort und legt den Gegenstand dorthin.
//
// **Fünf Arten seit `T7`** (#1472): `set`, `karte`, `vorlage`, `deck` und
// `sammlungssicht` — eine leere Art heißt `set`. Was `kennung` bedeutet,
// entscheidet die Art: eine Set-, Karten- oder Deckkennung, ein Paketname
// oder `<empfänger>/<zweck>` für eine Sicht.
//
// **Schreibend** (`docs/SCHREIBWEGE.md`) — es entsteht eine Datei am
// gewählten Ort.
//
// **Mit Dialog und nicht mit einem Pfad aus dem Fenster**, wie
// [Bindungen.SetSpeichernUnter] daneben: der Ort ist eine Entscheidung eines
// Menschen, und ein Pfad, den die Oberfläche zusammensetzte, wäre eine
// zweite Fassung derselben Frage.
func (b *Bindungen) UmschlagSchnueren(ctx context.Context, art, kennung string) UmschlagAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	leer := UmschlagAntwort{Objekte: []UmschlagobjektDTO{}}

	pfad, fehler := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.UmschlagSchnueren,
		Filter:    texte.UmschlagFilter,
		Muster:    "*" + austausch.Endung,
		Vorschlag: texte.UmschlagVorschlag,
		Speichern: true,
	})
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	// **Ein leerer Pfad ist ein Abbruch** — derselbe Grund wie bei
	// [Bindungen.SetSpeichernUnter].
	if pfad == "" {
		leer.Abgebrochen = true

		return leer
	}

	return b.app(ctx).UmschlagSchnueren(ctx, art, kennung, pfad)
}

// UmschlagWaehlen fragt nach einem Umschlag und sagt, was darin steht.
//
// **Rein lesend**, wie [Bindungen.UmschlagAnsehen] darunter — und der Dialog
// ändert daran nichts.
func (b *Bindungen) UmschlagWaehlen(ctx context.Context) UmschlagAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	leer := UmschlagAntwort{Objekte: []UmschlagobjektDTO{}}

	pfad, fehler := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:  texte.UmschlagOeffnen,
		Filter: texte.UmschlagFilter,
		Muster: "*" + austausch.Endung,
	})
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	if pfad == "" {
		leer.Abgebrochen = true

		return leer
	}

	angesehen := b.app(ctx).UmschlagAnsehen(ctx, pfad)
	angesehen.Pfad = pfad

	return angesehen
}

// UmschlagUebernehmen schreibt einen Umschlag in den Bestand.
//
// **Schreibend, und die Trennung von [Bindungen.UmschlagAnsehen] ist die
// fachliche Übernahmefreigabe**, die #1457 verlangt.
func (b *Bindungen) UmschlagUebernehmen(ctx context.Context, pfad string) OeffnenAntwort {
	return b.app(ctx).UmschlagUebernehmen(b.ruf(ctx), pfad)
}
