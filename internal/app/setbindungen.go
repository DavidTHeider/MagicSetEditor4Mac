// **Die Bindungen rund um ein Set als Ganzes** — anlegen, aus einer Datei
// lesen, als benutzt vermerken
// (#891).
//
// **Eine eigene Datei**, weil `bindungen.go` über die Zeilengrenze wuchs (§2).
// Der Schnitt folgt der Frage: dort **die Fassade als Ganzes**, hier **das
// Set als Gegenstand**.

package app

import "context"

// SetAnlegen legt ein leeres Set an und gibt seine Kennung (#891).
func (b *Bindungen) SetAnlegen(ctx context.Context, name, spiel, stil string) SetAnlegeAntwort {
	return b.app(ctx).SetAnlegen(b.ruf(ctx), name, spiel, stil)
}

// SetLoeschen nimmt ein Set aus dem Bestand — **endgültig**
// (#1218).
//
// **Der Name ist die Bestätigung, und die Rückfrage gehört ins Fenster** —
// derselbe Schnitt wie bei [Bindungen.KarteLoeschen]: hier wird ausgeführt,
// was schon entschieden ist. Was diese Stelle anders macht, ist der Name: er
// reist mit und wird im Bestand gegen den heutigen geprüft, damit ein
// umbenanntes Set nicht dem Klick von vorhin zum Opfer fällt.
func (b *Bindungen) SetLoeschen(ctx context.Context, setID, name string) SetLoeschAntwort {
	return b.app(ctx).SetLoeschen(b.ruf(ctx), setID, name)
}

// SetAusDatei liest ein Setpaket an einem **genannten** Pfad ein (#891).
//
// # Warum sie neben [Bindungen.SetOeffnen] steht
//
// Jene fragt im Dateidialog; hier kommt der Pfad aus einem **Ablegen**. Beides
// ist dieselbe Handlung eines Menschen — eine Datei zeigen —, nur der Weg ist
// ein anderer. Ein zweiter Dialog wäre die falsche Antwort auf eine Geste, die
// die Datei schon genannt hat.
//
// **Anders als [Bindungen.BeispielsetOeffnen] ohne Wache auf eine Liste.** Die
// dort steht, weil ein Beispielset aus dem Bündel kommt und ein beliebiger
// Pfad dort nichts zu suchen hat. Hier ist der Pfad die Geste selbst.
func (b *Bindungen) SetAusDatei(ctx context.Context, pfad string) OeffnenAntwort {
	return imDateilauf(b.app(ctx), b.ruf(ctx), "set", pfad, func(lauf context.Context) OeffnenAntwort {
		return b.app(ctx).SetOeffnen(lauf, pfad)
	}, func(antwort OeffnenAntwort) *Fehler { return antwort.Fehler })
}

// SetBenutzt merkt sich, dass ein Set aufgeschlagen wurde (#891).
func (b *Bindungen) SetBenutzt(ctx context.Context, setID string) SchreibAntwort {
	return b.app(ctx).SetBenutzt(b.ruf(ctx), setID)
}

// SetVorschauWaehlen fragt nach einer Setdatei und sagt, was ihr Einfahren
// täte — FMT-03.AK-01 (#1692). **Derselbe Dialog wie „Set öffnen …"**:
// eingefahren wird danach mit [Bindungen.SetAusDatei], wenn der Mensch es
// will.
func (b *Bindungen) SetVorschauWaehlen(ctx context.Context) SetvorschauAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel: texte.SetOeffnen, Filter: texte.SetFilter, Muster: "*.mse-set",
	})
	if err != nil {
		return SetvorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return b.app(ctx).SetVorschau(ctx, pfad)
}
