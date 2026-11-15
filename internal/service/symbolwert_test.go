package service_test

// **Der Wert eines Symbolfeldes** — Werkstück A von Phase 4.11
// (#1339).
//
// # Was nachgelesen wurde (§15)
//
// `src/data/field/symbol.hpp` (Fassung f9d9356 vom 2021-07-06):
//
//	ValueType filename;    ///< Filename of the symbol (in the current package)
//
// **Der Wert ist ein Dateiname im Paket.** Und die Serialisierung schreibt ihn
// **namenlos** — `src/data/field/symbol.cpp`, Zeilen 51–53:
//
//	IMPLEMENT_REFLECTION_NO_GET_MEMBER(SymbolValue) {
//	  if (fieldP->save_value || !handler.isWriting) REFLECT_NAMELESS(filename);
//	}
//
// Das heißt: in der `.mse-set` steht unter dem Feldnamen schlicht der
// Dateiname, wie bei einem Textfeld auch.
//
// # Dieser Lauf war von Anfang an grün, und das gehört dazugesagt
//
// §3 verlangt den roten Lauf vor der Zeile Produktivcode. **Hier gab es keine
// Zeile**: `werteAus` liest jeden Eintrag als `domain.Value{Field, Text}`,
// ohne die Art zu kennen, und deshalb läuft ein Symbolwert seit jeher rund.
//
// **Er steht trotzdem hier**, weil die Zusage leicht zu brechen ist: wer der
// Einfuhr eine Sonderbehandlung je Feldart gibt — und Werkstück A von Phase
// 4.11 kommt dieser Versuchung nahe —, macht diesen Lauf rot. Ein Lauf, der
// eine bestehende Eigenschaft festhält, ist kein Beleg für neue Arbeit; er ist
// ein Zaun um eine, die man sonst unbemerkt einreißt.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Symbolwert überlebt den Rundlauf durch die `.mse-set`.**
//
// Die zweite Hälfte der Abnahme aus #1339. Die erste — die Karte zeigt ihre
// Zeichnung — hängt an der Rasterung und steht in ihrem eigenen Lauf.
func TestEinSymbolwertLaeuftRund(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "symbol.mse-set")
	if err := os.MkdirAll(quelle, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Der Dateiname steht namenlos unter dem Feld** — so schreibt MSE2 ihn.
	const inhalt = `mse version: 2.0.0
game: probe
card:
	name: Prüfstein
	kartensymbol: zeichnung.mse-symbol
`

	if err := os.WriteFile(filepath.Join(quelle, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("set schreiben: %v", err)
	}

	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), quelle); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.set == nil || len(speicher.set.Cards) != 1 {
		t.Fatalf("Set: %+v", speicher.set)
	}

	var gefunden string

	for _, wert := range speicher.set.Cards[0].Values {
		if wert.Field == "kartensymbol" {
			gefunden = wert.Text
		}
	}

	if gefunden != "zeichnung.mse-symbol" {
		t.Errorf("der Symbolwert lautet %q", gefunden)
	}

	// **Und er steht als einziger Wert neben dem Namen da** — nicht als
	// Rohblock, nicht verschluckt, nicht zerlegt.
	if len(speicher.set.Cards[0].Values) != 2 {
		t.Errorf("die Werte lauten %+v", speicher.set.Cards[0].Values)
	}
}
