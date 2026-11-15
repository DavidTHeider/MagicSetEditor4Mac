package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Listenname des Pakets ist die zweite Aufschrift** —
// #699.
//
// # Der Befund
//
// Gezählt an `magic.mse-game/card_fields`: **249** Kartenfelder, davon
//
// | Angabe | Zahl |
// | --- | --- |
// | `caption:` | 0 |
// | `localized caption:` | 0 |
// | `card list name:` | 243 |
// | `localized card list name:` | 25 — **jede davon mit `de:`** |
//
// Die Aufschrift kam damit für **jedes** Feld aus [NameZuBeschriftung] und war
// englisch: „Rule text", „Casting cost", „Type". Das Paket bringt für 25 davon
// ein deutsches Wort mit — „Manakosten", „Typ", „Seltenheit", „S/W" —, und es
// stand ungelesen da.
//
// # Warum das eine Abweichung ist
//
//	if (caption.default_.empty()) caption.default_ = name_to_caption(name);
//	if (card_list_name.default_.empty()) card_list_name.default_ = capitalize(caption.default_);
//
// `src/data/field.cpp`, Zeilen 69–70, Fassung f9d9356 vom 2021-07-06. Bei MSE
// hängt der **Listenname an der Aufschrift**, nicht umgekehrt — sein
// Formularfeld hieße auch auf einem deutschen MSE „Type".
//
// **Die Abweichung ist die Übersetzung, nicht der Text.** Genommen wird allein
// der Eintrag in der gemeldeten Sprache; ohne einen bleibt alles, wie MSE es
// hat. Eine erfundene Übersetzung wäre etwas anderes und ist nach §9 verboten:
// der Feldname ist der Schlüssel, unter dem der Wert in der Datei steht.
func TestBeschriftetNimmtDenListennamenDerSprache(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		feld    domain.Field
		sprache string
		wollen  string
	}{
		{
			name: "der übersetzte Listenname gilt",
			feld: domain.Field{
				Name:                 "type",
				Listenname:           "Type",
				Listenuebersetzungen: map[string]string{"de": "Typ", "fr": "Type"},
			},
			sprache: "de", wollen: "Typ",
		},
		{
			// **Der unübersetzte Listenname zählt nicht** — er ist eine
			// Spaltenüberschrift und darf kürzer sein: `cmc` heißt in der
			// Liste `MV`, bei MSE im Formular aber `Cmc`.
			name: "ohne Eintrag in dieser Sprache bleibt es bei MSEs Aufschrift",
			feld: domain.Field{
				Name:                 "cmc",
				Listenname:           "MV",
				Listenuebersetzungen: map[string]string{"ch-s": "规则"},
			},
			sprache: "de", wollen: "Cmc",
		},
		{
			name:    "gar kein Listenname heißt: aus dem Namen",
			feld:    domain.Field{Name: "flavor text"},
			sprache: "de", wollen: "Flavor text",
		},
		{
			name: "`caption:` schlägt beide",
			feld: domain.Field{
				Name:                 "type",
				Beschriftung:         "Kartentyp",
				Listenuebersetzungen: map[string]string{"de": "Typ"},
			},
			sprache: "de", wollen: "Kartentyp",
		},
		{
			name: "und `localized caption:` schlägt auch die",
			feld: domain.Field{
				Name:                 "type",
				Beschriftung:         "Kartentyp",
				Uebersetzungen:       map[string]string{"de": "Typ der Karte"},
				Listenuebersetzungen: map[string]string{"de": "Typ"},
			},
			sprache: "de", wollen: "Typ der Karte",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if hat := fall.feld.Beschriftet(fall.sprache); hat != fall.wollen {
				t.Errorf("Beschriftet(%q) = %q, erwartet %q", fall.sprache, hat, fall.wollen)
			}
		})
	}
}
