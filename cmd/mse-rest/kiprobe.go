// **Ein Modell messen über REST** — Werkstück A9 von Phase 5.3
// (#1448).
//
// **Eine eigene Datei**, weil `sammlung.go` mit ihr über die Zeilengrenze aus
// §2 ginge.

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// KIModellPruefen misst, was ein Modell kann.
//
// **Es kostet Geld**, wenn der Anbieter anderswo rechnet — und geht deshalb
// durch dieselbe Sperre wie ein Vorschlag: Schalter, »nur lokal«, Journal
// und Freigabe.
func (p *preise) KIModellPruefen(ctx context.Context, bezug rest.KIProbenbezug) (
	rest.KIProbensicht, error,
) {
	antwort := p.kianwendung().KIModellPruefen(
		ctx, bezug.Anbieter, bezug.Alias, "")
	if antwort.Fehler != nil {
		return rest.KIProbensicht{}, antwort.Fehler.AlsFehler()
	}

	proben := make([]rest.KIProbesicht, 0, len(antwort.Proben))
	for _, probe := range antwort.Proben {
		proben = append(proben, rest.KIProbesicht(probe))
	}

	return rest.KIProbensicht{Proben: proben}, nil
}

// KIModellVergessen nimmt die Messungen eines Modells zurück.
func (p *preise) KIModellVergessen(
	ctx context.Context, bezug rest.KIProbenbezug,
) error {
	antwort := p.kianwendung().KIModellVergessen(
		ctx, bezug.Anbieter, bezug.Alias)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// KIBatch lässt ein Modell für mehrere Karten vorschlagen (A13.1, #1452).
//
// **Ohne Nachweis**: einen stellt nur das Fenster aus, und ein Cloudanbieter
// scheitert hier mit genau diesem Satz
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
func (p *preise) KIBatch(ctx context.Context, bezug rest.KIBatchbezug) (
	rest.KIBatchsicht, error,
) {
	antwort := p.kianwendung().KIBatchVorschlagen(ctx,
		bezug.KarteIDs, bezug.Anbieter, bezug.Alias, bezug.Auftrag, "")
	if antwort.Fehler != nil {
		return rest.KIBatchsicht{}, antwort.Fehler.AlsFehler()
	}

	ziele := make([]rest.KIBatchlagensicht, 0, len(antwort.Ziele))
	for _, lage := range antwort.Ziele {
		ziele = append(ziele, alsRestbatchlage(lage))
	}

	return rest.KIBatchsicht{
		Kennung: antwort.Kennung, Anbieter: antwort.Anbieter, Ziele: ziele,
	}, nil
}

// alsRestbatchlage übersetzt die Lage eines Ziels.
func alsRestbatchlage(lage app.KIBatchlage) rest.KIBatchlagensicht {
	sicht := rest.KIBatchlagensicht{
		KarteID: lage.Zielkarte, Kennung: lage.Kennung, Modell: lage.Modell,
		Befunde: lage.Befunde, Ablauf: lage.Ablauf,
		Felder: make([]rest.KIFeldwertsicht, 0, len(lage.Felder)),
	}

	if lage.Fehler != nil {
		sicht.Fehler = lage.Fehler.Meldung
	}

	for _, feld := range lage.Felder {
		sicht.Felder = append(sicht.Felder, rest.KIFeldwertsicht{
			Feld: feld.Feld, Werte: feld.Werte, Bisher: feld.Bisher,
		})
	}

	return sicht
}
