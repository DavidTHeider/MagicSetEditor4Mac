// **Die Bindungen der Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// **Eine eigene Datei**, weil `bindungenkarten.go` an der Zeilengrenze aus §2
// steht — und weil der Schnitt der Frage folgt: dort die **Karten**, hier der
// **Besitz**. Das sind zwei Dinge, und der Vermerk
// [0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)
// sagt, warum.

package app

import "context"

// Sammlung gibt die ganze Sammlung.
func (b *Bindungen) Sammlung(ctx context.Context) Sammlungsantwort {
	return b.app(ctx).Sammlung(b.ruf(ctx))
}

// **`SammlungFuer` steht hier nicht**, und das ist kein Versehen.
//
// Das Fenster zeigt die **ganze** Sammlung; eine Abfrage je Druck hätte dort
// keinen Aufrufer, und eine Bindung ohne Aufrufer ist Code mit voller
// Abdeckung, den niemand benutzt (§10, erste Dimension). Über MCP, an der
// Kommandozeile und im Dienst ist sie der Normalfall: dort fragt jemand nach
// **einem** Druck, statt eine Liste durchzugehen.
//
// **Sie kommt ans Fenster, wenn eine Karte danach fragt** — mit den
// Kartenreferenzen aus Phase 3.6
// (#1167):
// „besitze ich diesen Druck?“ ist die Frage, die sie dort beantwortet.

// ExemplarAnlegen nimmt ein Exemplar in die Sammlung auf.
func (b *Bindungen) ExemplarAnlegen(ctx context.Context, exemplar ExemplarDTO) Exemplarantwort {
	return b.app(ctx).ExemplarAnlegen(b.ruf(ctx), exemplar)
}

// ExemplarAendern schreibt ein Exemplar zurück.
func (b *Bindungen) ExemplarAendern(ctx context.Context, exemplar ExemplarDTO) *Fehler {
	return b.app(ctx).ExemplarAendern(b.ruf(ctx), exemplar)
}

// ExemplarLoeschen nimmt ein Exemplar aus der Sammlung.
func (b *Bindungen) ExemplarLoeschen(ctx context.Context, id, fassung int64) *Fehler {
	return b.app(ctx).ExemplarLoeschen(b.ruf(ctx), id, fassung)
}

// PreiseHolen fragt Scryfall nach den Preisen eines Drucks.
//
// **Auf Anweisung und nicht von allein** — der einzige dieser Wege, der ins
// Netz geht.
func (b *Bindungen) PreiseHolen(ctx context.Context, printing string) Preisantwort {
	return b.app(ctx).PreiseHolen(b.ruf(ctx), printing)
}

// Preise gibt die abgelegten Stände eines Drucks.
func (b *Bindungen) Preise(ctx context.Context, printing string) Preisantwort {
	return b.app(ctx).Preise(b.ruf(ctx), printing)
}

// Preisverlauf gibt eine Preisreihe in der Zeit.
func (b *Bindungen) Preisverlauf(ctx context.Context, reihe PreisreiheDTO) Preisantwort {
	return b.app(ctx).Preisverlauf(b.ruf(ctx), reihe)
}

// Preisbulk holt eine ganze Preisdatei und legt ab, was darin steht.
func (b *Bindungen) Preisbulk(ctx context.Context, quelle, umfang string) Bulkbezugsantwort {
	return b.app(ctx).Preisbulk(b.ruf(ctx), quelle, umfang)
}

// Marktangebote gibt die Posten eines Marktplatzes zu einem Druck.
//
// **Lesend für den Bestand, mit einer Ausnahme** (A4, #1443): gemerkt wird die
// gefundene Zuordnung Druck → Kartenfassung, damit der nächste Abruf sie nicht
// wieder sucht. Gekauft wird nichts.
func (b *Bindungen) Marktangebote(
	ctx context.Context,
	quelle, printing, sprache, finish string, zustaende []string,
) Marktantwort {
	return b.app(ctx).Marktangebote(b.ruf(ctx),
		quelle, printing, sprache, finish, zustaende)
}

// Anbieterstand fragt, wie viel Kontingent an einem Anbieterkonto hängt.
//
// **Lesend, und die Antwort sagt, wie alt sie ist** (A5, #1444).
func (b *Bindungen) Anbieterstand(ctx context.Context, dienst string) Anbieterstandantwort {
	return b.app(ctx).Anbieterstand(b.ruf(ctx), dienst)
}

// Fremdpreisverlauf holt den Preisverlauf eines Drucks bei einer Quelle.
func (b *Bindungen) Fremdpreisverlauf(
	ctx context.Context,
	quelle, printing, finish string, zustaende []string,
) Fremdpreisantwort {
	return b.app(ctx).Fremdpreisverlauf(b.ruf(ctx), quelle, printing, finish, zustaende)
}

// KIModelle nennt, was über KI-Modelle feststeht.
//
// **Seit A7 (#1446) fragt sie die eingeschaltete Laufzeit** — und `einrichtung`
// sagt, welches Stück fehlt, wenn die Liste leer bleibt.
func (b *Bindungen) KIModelle(ctx context.Context) KIKatalogantwort {
	return b.app(ctx).KIModelle(b.ruf(ctx))
}

// KIVorschlagen fragt ein Modell und legt seinen Vorschlag ab.
//
// **Es ändert die Karte nicht** (A7, #1446) — dafür gibt es die Bindung
// darunter.
// **Ein Verbund statt fünf Zeichenketten** — Werkstück A13.4 von Phase 5.3
// (#1452).
// Mit Belegfrage und Freigabe wären es sieben, und wer zwei davon
// vertauscht, merkt es an keiner Stelle: der Typ ist derselbe.
//
// **Die Wails-Naht darf sich ändern** (§16): sie ist eine innere Naht
// zwischen zwei Hälften **derselben** Auslieferung. Die drei Flächen mit
// Verträglichkeitszusage behalten ihre Gestalt und bekommen die zwei Felder
// **zusätzlich** — erweiternd.
func (b *Bindungen) KIVorschlagen(
	ctx context.Context,
	karteID string, bezug KIVorschlagsbezug,
) KIVorschlagAntwort {
	b.tor(ctx)

	kennung, befund := benannteKennungAus("karteId", karteID)
	if befund != nil {
		return KIVorschlagAntwort{Fehler: befund}
	}

	bezug.KarteID = kennung

	return b.app(ctx).KIVorschlagenMit(b.ruf(ctx), bezug)
}

// KIVorschlagUebernehmen schreibt die **ausgewählten** Felder eines
// abgelegten Vorschlags (A11, #1450).
//
// **Die Auswahl steht im Ruf und nicht in einer Vorgabe**: das Issue verlangt
// „nur ausgewählte gültige Felder", und eine Vorgabe „dann eben alles" wäre
// die stillste Art, eine Karte zu überschreiben.
func (b *Bindungen) KIVorschlagUebernehmen(
	ctx context.Context,
	kennung string, felder []string,
) KIUebernahmeAntwort {
	return b.app(ctx).KIVorschlagUebernehmen(b.ruf(ctx), kennung, felder)
}

// KIModellPruefen misst, was ein Modell kann (A9, #1448).
//
// **Es kostet Geld**, wenn der Anbieter anderswo rechnet — und geht deshalb
// durch dieselbe Sperre wie ein Vorschlag.
func (b *Bindungen) KIModellPruefen(
	ctx context.Context,
	anbieter, alias, nachweis string,
) KIProbeAntwort {
	return b.app(ctx).KIModellPruefen(b.ruf(ctx), anbieter, alias, nachweis)
}

// KIModellVergessen nimmt die Messungen eines Modells zurück (A9, #1448).
func (b *Bindungen) KIModellVergessen(ctx context.Context, anbieter, alias string) SchreibAntwort {
	return b.app(ctx).KIModellVergessen(b.ruf(ctx), anbieter, alias)
}

// Portfolio rechnet die Sammlung zusammen.
func (b *Bindungen) Portfolio(ctx context.Context) Portfolioantwort {
	return b.app(ctx).Portfolio(b.ruf(ctx))
}

// KIFeldherkunft sagt, woher die Werte einer Karte stammen (A11, #1450).
//
// **Eine Auskunft und keine Frage nach draußen**: hier geht kein Aufruf an
// einen Anbieter, und es entsteht kein Auftrag.
func (b *Bindungen) KIFeldherkunft(ctx context.Context, karteID string) KIFeldherkunftAntwort {
	b.tor(ctx)

	kennung, befund := benannteKennungAus("karteId", karteID)
	if befund != nil {
		// **Auch hier eine Liste und kein `null`**: was fehlt, fehlt im
		// Fenster als `herkunft.length`.
		return KIFeldherkunftAntwort{
			Herkunft: []KIFeldherkunftDTO{}, Fehler: befund,
		}
	}

	return b.app(ctx).KIFeldherkunft(b.ruf(ctx), kennung)
}
