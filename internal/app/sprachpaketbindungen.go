package app

// **Eigene Übersetzungen am Fenster** — LOC-02
// (#1688).
//
// Fünf Wege reichen durch; Aus- und Eingeben fangen mit einem Dialog an und
// stehen deshalb hier mit ihm, wie `TurnierAusgabeSpeichern`.

import "context"

// Sprachpakete gibt alle eigenen Übersetzungen.
func (b *Bindungen) Sprachpakete(ctx context.Context) SprachpaketlisteAntwort {
	return b.app(ctx).Sprachpakete(b.ruf(ctx))
}

// Sprachpaket gibt eine Übersetzung mit ihren Sätzen.
func (b *Bindungen) Sprachpaket(ctx context.Context, kennung string) SprachpaketAntwort {
	return b.app(ctx).Sprachpaket(b.ruf(ctx), kennung)
}

// SprachpaketAnlegen legt eine Übersetzung über einer Basis an.
func (b *Bindungen) SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) SprachpaketAntwort {
	return b.app(ctx).SprachpaketAnlegen(b.ruf(ctx), kennung, name, basis)
}

// SprachpaketAendern setzt Angaben und Sätze.
func (b *Bindungen) SprachpaketAendern(
	ctx context.Context,
	kennung string, angaben SprachpaketangabenDTO, texte map[string]string,
) SprachpaketAntwort {
	return b.app(ctx).SprachpaketAendern(b.ruf(ctx), kennung, angaben, texte)
}

// SprachpaketLoeschen entfernt eine Übersetzung.
func (b *Bindungen) SprachpaketLoeschen(ctx context.Context, kennung string) SchreibAntwort {
	return b.app(ctx).SprachpaketLoeschen(b.ruf(ctx), kennung)
}

// SprachpaketAusgabeSpeichern fragt nach einem Ort und legt die Übersetzung
// dorthin — vorgeschlagen unter ihrer Kennung.
func (b *Bindungen) SprachpaketAusgabeSpeichern(ctx context.Context, kennung string) SprachpaketAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachSprachdialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.SprachpaketAusgeben, Filter: texte.SprachpaketFilter, Muster: "*.json",
		Vorschlag: kennung + ".sprachpaket.json", Speichern: true,
	}, func(pfad string) SprachpaketAntwort {
		return b.app(ctx).SprachpaketAusgabeSchreiben(ctx, kennung, pfad)
	})
}

// SprachpaketEinlesenOeffnen fragt nach einer Datei und nimmt die
// Übersetzung auf — **eine vorhandene nur mit `ersetzen`**.
func (b *Bindungen) SprachpaketEinlesenOeffnen(ctx context.Context, ersetzen bool) SprachpaketAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return nachSprachdialog(ctx, b.app(ctx), Dateiauswahl{
		Titel: texte.SprachpaketEinlesen, Filter: texte.SprachpaketFilter, Muster: "*.json",
	}, func(pfad string) SprachpaketAntwort {
		return b.app(ctx).SprachpaketEinlesenLesen(ctx, pfad, ersetzen)
	})
}

// nachSprachdialog fragt nach dem Ort und ruft den Weg — **ein
// weggeklickter Dialog ist ein Abbruch und kein Fehler**, wie bei
// `nachDialog`.
func nachSprachdialog(
	ctx context.Context, anwendung *App, auswahl Dateiauswahl, weg func(pfad string) SprachpaketAntwort,
) SprachpaketAntwort {
	pfad, err := anwendung.waehleDatei(ctx, auswahl)
	if err != nil {
		return leeresSprachpaket(Uebersetze(ctx, err))
	}

	if pfad == "" {
		antwort := leeresSprachpaket(nil)
		antwort.Abgebrochen = true

		return antwort
	}

	return weg(pfad)
}
