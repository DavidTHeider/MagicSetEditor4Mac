package app

// **Das Schema einer Karte** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Aus den Feldern des Pakets, nicht aus einer Liste daneben
//
// A6 verlangt es ([domain.KISchema]): „Kanonisches Schema aus **tatsächlichen
// Paketfeldern**." Hier wird es geholt — über denselben Stilwähler, den der
// Karteneditor benutzt. Eine zweite Quelle liefe beim ersten neuen Stil
// auseinander.
//
// # Welche Felder eine KI setzen darf
//
// **Text und Wahl, und sonst nichts.** Die Zuordnung unten ist eine
// Aufzählung mit einem Grund je Zeile — nicht „alles außer":
//
//	| MSE-Art             | Warum sie schreibbar ist oder nicht         |
//	| ------------------- | ------------------------------------------- |
//	| `text`              | ein Sprachmodell schreibt Text              |
//	| `choice`            | eine Wahl aus einer aufgezählten Liste      |
//	| `multiple choice`   | dasselbe, mehrfach                          |
//	| `image`, `symbol`   | eine **Datei** im Paket — kein Text         |
//	| `color`             | eine Farbe; ein Name dafür ist kein Wert    |
//	| `package choice`    | wählt ein **Paket**, keine Karteneigenschaft |
//	| `boolean`, `info`   | kein Text, und `info` ist nicht beschreibbar |
//
// **Ein Feld, das es gibt und das niemand setzen darf, bleibt im Schema
// stehen** — mit `Schreibbar: false`. Das ist der Unterschied zwischen „kenne
// ich nicht" und „darfst du nicht", und A6 hat ihn ausdrücklich getrennt.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kischemaquelle holt das Schema einer Karte über Leser und Stile.
type kischemaquelle struct{ app *App }

// Schema gibt die Felder des Pakets, zu dem diese Karte gehört.
//
// **Ein unlesbares Paket gibt kein leeres Schema, sondern einen Fehler.** A6
// sagt es scharf: „Fehlendes/unlesbares Paket erzeugt für KI **keinen
// beliebigen Feldfreibrief**." Ein leeres Schema weist zwar alles ab — aber es
// sähe aus wie eine Antwort, und die Meldung, die weiterhilft, fehlte.
func (k kischemaquelle) Schema(ctx context.Context, karte int64) (domain.KISchema, error) {
	if k.app.kartenleser == nil || k.app.stile == nil {
		return domain.KISchema{}, fehlenderBestand()
	}

	zeile, fehler := k.app.kartenleser.KarteLesen(ctx, karte)
	if fehler != nil {
		return domain.KISchema{}, fehler
	}

	felder, fehler := k.app.stile.Kartenfelder(ctx, zeile.SetID)
	if fehler != nil {
		return domain.KISchema{}, fehler
	}

	if len(felder) == 0 {
		return domain.KISchema{}, fmt.Errorf(
			"%w: das Paket dieser Karte nennt kein einziges Feld — ohne "+
				"Schema gibt es keine Prüfung und deshalb keinen Vorschlag",
			domain.ErrNotFound)
	}

	return domain.KISchema{
		Paket:   fmt.Sprintf("set:%d", zeile.SetID),
		Fassung: fmt.Sprintf("%d Felder", len(felder)),
		Felder:  schemafelderAus(felder),
	}, nil
}

// schemafelderAus macht aus den Paketfeldern die Schemafelder.
func schemafelderAus(felder []domain.Field) []domain.KISchemafeld {
	umgesetzt := make([]domain.KISchemafeld, 0, len(felder))

	for i := range felder {
		umgesetzt = append(umgesetzt, schemafeldAus(felder[i]))
	}

	return umgesetzt
}

// schemafeldAus macht aus einem Paketfeld ein Schemafeld.
func schemafeldAus(feld domain.Field) domain.KISchemafeld {
	switch feld.Kind {
	case domain.KindText, domain.KindUnknown:
		return domain.KISchemafeld{
			Name: feld.Name, Typ: domain.KIText, Schreibbar: true,
		}

	case domain.KindChoice:
		return domain.KISchemafeld{
			Name: feld.Name, Typ: domain.KIAuswahl,
			Werte: feld.Choices, Schreibbar: true,
		}

	case domain.KindMultipleChoice:
		return domain.KISchemafeld{
			Name: feld.Name, Typ: domain.KIAuswahl, Werte: feld.Choices,
			Mehrfach: true, Schreibbar: true,
		}

	case domain.KindImage, domain.KindSymbol, domain.KindColor,
		domain.KindPackageChoice, domain.KindBoolean, domain.KindInfo:
		return domain.KISchemafeld{Name: feld.Name, Typ: domain.KIText}

	default:
		// **Eine unbekannte Art ist nicht schreibbar.** Sie kommt aus einem
		// Paket, dessen Feldart dieses Programm nicht kennt — und was man
		// nicht kennt, lässt man eine KI nicht setzen.
		return domain.KISchemafeld{Name: feld.Name, Typ: domain.KIText}
	}
}

// alsJSONSchema macht aus dem Schema die Form, die Ollama halten kann.
//
// **Nur die schreibbaren Felder** — ein Schema, das `casting cost` anböte,
// bekäme einen Vorschlag dafür und wiese ihn danach ab. Das ist kein Schutz,
// sondern eine vergeudete Antwort und eine Meldung, die niemand versteht.
//
// **Ohne `required`**: was ein Modell nicht weiß, soll es weglassen dürfen.
// Ein erzwungenes Feld bekäme eine erfundene Antwort — und das Issue sagt,
// dass Form geprüft wird und nicht Wahrheit.
func alsJSONSchema(schema domain.KISchema) map[string]any {
	eigenschaften := map[string]any{}

	for _, feld := range schema.Felder {
		if !feld.Schreibbar {
			continue
		}

		eigenschaften[feld.Name] = feldschema(feld)
	}

	return map[string]any{"type": "object", "properties": eigenschaften}
}

// feldschema gibt die Form eines einzelnen Feldes.
func feldschema(feld domain.KISchemafeld) map[string]any {
	einer := map[string]any{"type": "string"}

	if feld.Typ == domain.KIAuswahl && len(feld.Werte) > 0 {
		einer["enum"] = feld.Werte
	}

	if feld.Mehrfach {
		return map[string]any{"type": "array", "items": einer}
	}

	return einer
}
