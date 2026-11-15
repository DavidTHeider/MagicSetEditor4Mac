package app

// **Turnierdatei und Bericht am Fenster** — DOM-04.AK-03
// (#1699).
//
// **Auf `Bindungen`**, wie `PartieAusfuhrSpeichern`: jeder der drei Wege
// fängt mit einem Dialog an, und ein Dialog braucht ein Fenster.

import "context"

// TurnierAusgabeSpeichern fragt nach einem Ort und legt das Turnier dorthin.
func (b *Bindungen) TurnierAusgabeSpeichern(ctx context.Context, turnierID string) TurnierdateiAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachDialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.TurnierSpeichern, Filter: texte.TurnierFilter, Muster: "*.mseturnier",
		Vorschlag: texte.TurnierVorschlag, Speichern: true,
	}, func(pfad string) TurnierdateiAntwort {
		return b.app(ctx).TurnierAusgabeSchreiben(ctx, turnierID, pfad)
	})
}

// TurnierEinlesenOeffnen fragt nach einer Datei und legt das Turnier daraus
// an.
func (b *Bindungen) TurnierEinlesenOeffnen(ctx context.Context) TurnierdateiAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachDialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.TurnierOeffnen, Filter: texte.TurnierFilter, Muster: "*.mseturnier",
	}, func(pfad string) TurnierdateiAntwort {
		return b.app(ctx).TurnierEinlesenLesen(ctx, pfad)
	})
}

// TurnierBerichtSpeichern fragt nach einem Ort und legt den Bericht dorthin.
func (b *Bindungen) TurnierBerichtSpeichern(ctx context.Context, turnierID string) TurnierdateiAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachDialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.BerichtSpeichern, Filter: texte.BerichtFilter, Muster: "*.md",
		Vorschlag: texte.BerichtVorschlag, Speichern: true,
	}, func(pfad string) TurnierdateiAntwort {
		return b.app(ctx).TurnierBerichtSchreiben(ctx, turnierID, pfad)
	})
}

// nachDialog fragt nach dem Ort und ruft den Weg — **ein weggeklickter
// Dialog ist ein Abbruch und kein Fehler**, wie bei `SetSpeichernUnter`.
//
// **Eine Funktion und kein Verfahren an `Bindungen`**: dort wäre sie eine
// Bindung, die das Fenster sähe.
func nachDialog(
	ctx context.Context, anwendung *App, auswahl Dateiauswahl, weg func(pfad string) TurnierdateiAntwort,
) TurnierdateiAntwort {
	pfad, err := anwendung.waehleDatei(ctx, auswahl)
	if err != nil {
		return leereTurnierdatei(ctx, err)
	}

	if pfad == "" {
		return TurnierdateiAntwort{Abgebrochen: true, Befunde: []string{}}
	}

	return weg(pfad)
}
