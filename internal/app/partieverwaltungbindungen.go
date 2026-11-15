package app

import "context"

// **Die Partieverwaltung am Fenster** — PLAY-03.AK-05
// (#1711).
//
// **Auf `Bindungen` und nicht auf `App`**, wie `SetSpeichernUnter`: zwei der
// vier Wege fangen mit einem Dialog an, und ein Dialog braucht ein Fenster.
// Die beiden anderen stehen hier, weil ein Mensch sie an derselben Stelle
// sucht.

// PartieLoeschen nimmt eine Partie aus dem Bestand.
func (b *Bindungen) PartieLoeschen(ctx context.Context, kennung string) PartieverwaltungAntwort {
	return b.app(ctx).PartieLoeschen(b.ruf(ctx), kennung)
}

// PartieKopieren legt eine Partie als neuen Übungszustand daneben.
func (b *Bindungen) PartieKopieren(ctx context.Context, kennung string) PartieverwaltungAntwort {
	return b.app(ctx).PartieKopieren(b.ruf(ctx), kennung)
}

// PartieAusfuhrSpeichern fragt nach einem Ort und legt die Partie dorthin.
func (b *Bindungen) PartieAusfuhrSpeichern(
	ctx context.Context,
	kennung string,
) PartieverwaltungAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.PartieSpeichern,
		Filter:    texte.PartieFilter,
		Muster:    "*.msepartie",
		Vorschlag: texte.PartieVorschlag,
		Speichern: true,
	})
	if err != nil {
		return leereVerwaltung(ctx, err)
	}

	if pfad == "" {
		return PartieverwaltungAntwort{
			Abgebrochen: true, Befunde: []string{},
		}
	}

	return b.app(ctx).PartieAusfuhrSchreiben(ctx, kennung, pfad)
}

// PartieEinfuhrOeffnen fragt nach einer Datei und legt die Partie daraus an.
func (b *Bindungen) PartieEinfuhrOeffnen(ctx context.Context) PartieverwaltungAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:  texte.PartieOeffnen,
		Filter: texte.PartieFilter,
		Muster: "*.msepartie",
	})
	if err != nil {
		return leereVerwaltung(ctx, err)
	}

	if pfad == "" {
		return PartieverwaltungAntwort{
			Abgebrochen: true, Befunde: []string{},
		}
	}

	return b.app(ctx).PartieEinfuhrLesen(ctx, pfad)
}
