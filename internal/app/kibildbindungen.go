package app

import "context"

// **Die fünf Bindungen des Bildvorschlags** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Eine eigene Datei
//
// `sammlungsbindungen.go` stand mit ihnen über der Zeilengrenze aus §2, und
// der Schnitt folgt der Frage: dort die Sammlung und der Textvorschlag, hier
// **alles, was mit einem erzeugten Bild zu tun hat**.
//
// # Und warum es fünf sind und nicht zwei
//
// **Weil ein Bild vier Zustände durchläuft und jeder einen Ruf hat**:
// erzeugen, ansehen, zuschneiden, einsetzen — oder verwerfen. Ein Ruf, der
// alles täte, schriebe die Karte, bevor jemand das Bild gesehen hat; genau
// das verbietet das Issue („bleiben Vorschläge, bis Nutzende sie mit
// vorhandener Bildverarbeitung übernehmen").

// KIBildVorschlagen lässt ein Bild erzeugen und legt es in den Vorrat.
//
// **Es ändert die Karte nicht.** Was zurückkommt, ist ein Vorschlag mit
// Frist; wer ihn will, ruft [Bindungen.KIBildZuschnitt] und danach
// [Bindungen.KIBildEinsetzen].
func (b *Bindungen) KIBildVorschlagen(
	ctx context.Context,
	karteID, feld, anbieter, alias, aufgabe, groesse, guete string,
	bearbeiten bool, nachweis string,
) KIBildvorschlagAntwort {
	b.tor(ctx)

	kennung, befund := benannteKennungAus("karteId", karteID)
	if befund != nil {
		return KIBildvorschlagAntwort{KIBildvorschlagDTO: KIBildvorschlagDTO{
			Angaben: []KIBildangabeDTO{},
		}, Fehler: befund}
	}

	return b.app(ctx).KIBildVorschlagen(b.ruf(ctx), kennung, feld, anbieter,
		alias, aufgabe, groesse, guete, bearbeiten, nachweis)
}

// KIBildVerwerfen nimmt einen wartenden Vorschlag weg.
func (b *Bindungen) KIBildVerwerfen(ctx context.Context, kennung string) SchreibAntwort {
	return b.app(ctx).KIBildVerwerfen(b.ruf(ctx), kennung)
}

// KIBildZuschnitt bereitet den Zuschnitt eines wartenden Vorschlags vor.
//
// **Derselbe Rückgabewert wie bei einer Datei von der Platte**
// ([Bindungen.BildAusDatei]): das Fenster zeigt denselben Rahmen, und was
// dahinterliegt, ist dieselbe Rechnung.
func (b *Bindungen) KIBildZuschnitt(ctx context.Context, karteID, feld, kennung string) ZuschnittAntwort {
	return b.app(ctx).KIBildZuschnitt(b.ruf(ctx), karteID, feld, kennung)
}

// KIBildEinsetzen schneidet zu, legt ab und schreibt Feld **und Herkunft**.
func (b *Bindungen) KIBildEinsetzen(
	ctx context.Context,
	setID, karteID, feld, kennung string, auswahl AuswahlDTO,
) SchreibAntwort {
	return b.app(ctx).KIBildEinsetzen(b.ruf(ctx), setID, karteID, feld,
		kennung, auswahl)
}

// KIBildHerkunft gibt die Angaben zu einem übernommenen Bild.
//
// **Eine Auskunft und keine Frage nach draußen** — wie
// [Bindungen.KIFeldherkunft]: hier geht kein Aufruf an einen Anbieter.
func (b *Bindungen) KIBildHerkunft(ctx context.Context, karteID, feld string) KIBildherkunftAntwort {
	b.tor(ctx)

	kennung, befund := benannteKennungAus("karteId", karteID)
	if befund != nil {
		// **Eine Liste und kein `null`**: was fehlt, fehlt im Fenster als
		// `angaben.length`.
		return KIBildherkunftAntwort{Angaben: []KIBildangabeDTO{}, Fehler: befund}
	}

	return b.app(ctx).KIBildHerkunft(b.ruf(ctx), kennung, feld)
}

// KIBildAbbrechen hält den laufenden Bildauftrag einer Karte an.
//
// **Ohne Kontext**, wie [Bindungen.IndexAbbrechen]: der Abbruch bricht einen
// **anderen** Lauf ab und hat selbst nichts zu tun, worauf man warten müsste.
func (b *Bindungen) KIBildAbbrechen(ctx context.Context, karteID int64) bool {
	return b.app(ctx).KIBildAbbrechen(karteID)
}
