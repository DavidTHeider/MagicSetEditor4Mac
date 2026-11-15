package mcp

// **Ein Batch von Vorschlägen über MCP** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Was hier geht und was nicht
//
// **Ein Batch an ein Modell auf diesem Rechner geht.** Einer an einen
// Cloudanbieter nicht: er verlangt die Freigabe eines Menschen, und die wird
// **am Fenster** ausgestellt — ein Modell stellt sich keine aus
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md),
// [0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
//
// # Und er schreibt nichts
//
// **Anders als `schreiben.ki.vorschlag`**, der ein `uebernehmen` kennt. Ein
// Batch über zwölf Karten, den ein Modell in einem Zug übernehmen dürfte,
// wäre die Vollmacht, gegen die es Nachweise gibt. Wer übernehmen will, tut
// es je Vorschlag — und sieht dabei, was er schreibt.

import (
	"context"
	"encoding/json"
)

// KIBatchbezug ist, was für einen Batch gebraucht wird.
type KIBatchbezug struct {
	// KarteIDs sind die Zielkarten — höchstens so viele, wie `domain`
	// erlaubt.
	KarteIDs []int64 `json:"karteIds"`

	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
	Auftrag  string `json:"auftrag"`
}

// KIBatchlagensicht ist, was aus **einem** Ziel geworden ist.
type KIBatchlagensicht struct {
	KarteID int64  `json:"karteId"`
	Kennung string `json:"kennung"`
	Modell  string `json:"modell"`

	Felder  []KIFeldwertsicht `json:"felder"`
	Befunde []string          `json:"befunde"`

	Ablauf string `json:"ablauf"`

	// Fehler ist der Grund **an diesem Ziel** — leer, wo es geklappt hat.
	Fehler string `json:"fehler"`
}

// KIBatchsicht trägt die Lagen aller Ziele.
type KIBatchsicht struct {
	Kennung  string              `json:"kennung"`
	Anbieter string              `json:"anbieter"`
	Ziele    []KIBatchlagensicht `json:"ziele"`
}

// kibatchwerkzeug gibt das Werkzeug für einen Batch.
func kibatchwerkzeug(werkstatt Preiswerkstatt) Werkzeug {
	return Werkzeug{
		Name: "schreiben.ki.batch", Art: Schreibend,
		Erklaerung: "Fragt EIN Modell für MEHRERE Zielkarten und gibt je " +
			"Ziel, was daraus wurde. An den KARTEN ÄNDERT es nichts — " +
			"übernommen wird je Vorschlag über »schreiben.ki.vorschlag«. " +
			"ÄNDERT den Bestand trotzdem: es bucht EINEN Kostenposten für " +
			"den ganzen Batch und schließt ihn wieder. Ein Cloudanbieter " +
			"braucht die Freigabe eines Menschen, und die gibt es nur am " +
			"Fenster. " + vorbehalt,
		Schema: objekt(map[string]any{
			"karteIds": zahlenliste(
				"Die Kennungen der Zielkarten. Dieselbe Karte zweimal wird "+
					"abgewiesen: zwei Vorschläge zu einer Karte kennen "+
					"einander nicht.",
				"Die Kennung einer Zielkarte."),
			"alias": text("Das Modell, wie »lesen.ki.modelle« es nennt — es " +
				"muss die strukturierte Ausgabe BELEGT haben."),
			"auftrag": text("Was das Modell für jede Karte tun soll."),
		}, "karteIds", "alias", "auftrag"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var bezug KIBatchbezug

			if err := lies(roh, &bezug); err != nil {
				return "", err
			}

			sicht, fehler := werkstatt.KIBatch(ctx, bezug)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(mitLeerenListen(sicht))
		},
	}
}

// mitLeerenListen macht aus `null` eine leere Liste.
//
// **Für ein Sprachmodell sieht `null` aus wie „konnte nicht beantwortet
// werden"** — dieselbe Überlegung wie beim einzelnen Vorschlag.
func mitLeerenListen(sicht KIBatchsicht) KIBatchsicht {
	if sicht.Ziele == nil {
		sicht.Ziele = []KIBatchlagensicht{}
	}

	for i, lage := range sicht.Ziele {
		if lage.Felder == nil {
			sicht.Ziele[i].Felder = []KIFeldwertsicht{}
		}

		if lage.Befunde == nil {
			sicht.Ziele[i].Befunde = []string{}
		}
	}

	return sicht
}
