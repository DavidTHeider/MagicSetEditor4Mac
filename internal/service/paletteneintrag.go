package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// paletteneintraege macht aus dem Einfügemenü der Schrift die Palette.
func paletteneintraege(
	ctx context.Context, menue []msepaket.Menueeintrag, stil *msepaket.Feldstil,
	schrift msepaket.Symbolschrift, name string, umgebung *msescript.Umgebung,
) []Paletteneintrag {
	eintraege := make([]Paletteneintrag, 0, len(menue))

	for i := range menue {
		eintrag := &menue[i]

		// **Ein Code ohne Namen wird weggelassen.** Drei stehen so im Pack:
		// `space-psi-xerent` schreibt ein nacktes `item:` dorthin, wo ein
		// `type: line` hingehörte, zwei weitere `item:` mit einem Leerzeichen
		// dahinter. MSE baut daraus einen Menüeintrag ohne Bild, der nichts
		// einfügt; eine Schaltfläche ohne Aufschrift und ohne Wirkung ist
		// keine.
		if eintrag.Art == msepaket.MenueCode && strings.TrimSpace(eintrag.Name) == "" {
			continue
		}

		eintraege = append(eintraege,
			palettenEintrag(ctx, eintrag, stil, schrift, name, umgebung))
	}

	if len(eintraege) == 0 {
		return nil
	}

	return eintraege
}

// palettenEintrag baut einen einzelnen Eintrag samt seinem Plan.
func palettenEintrag(
	ctx context.Context, eintrag *msepaket.Menueeintrag, stil *msepaket.Feldstil,
	schrift msepaket.Symbolschrift, name string, umgebung *msescript.Umgebung,
) Paletteneintrag {
	gebaut := Paletteneintrag{
		Art:          eintrag.Art,
		Beschriftung: beschriftungAus(eintrag),
		Rueckfrage:   eintrag.Rueckfrage,
	}

	switch eintrag.Art {
	case msepaket.MenueUntermenue:
		gebaut.Kinder = paletteneintraege(
			ctx, eintrag.Kinder, stil, schrift, name, umgebung)

	case msepaket.MenueCode:
		gebaut.Code = eintrag.Name
		gebaut.Plan = symbolplan(ctx, eintrag.Name, stil, schrift, name, umgebung)

	// **Ein `custom` bekommt kein Bild.** MSE zeigt dort das Standardsymbol
	// der Schrift (`makeMenuItem`, Zeile 505, Fassung f9d9356 vom
	// 2021-07-06) — eine Kachel, die für „irgendeine Zahl" steht und keinen
	// Code hat. Eine Schaltfläche, die einen Dialog öffnet, braucht das nicht,
	// und ein Bild, das ein anderes Symbol zeigt als das eingefügte, wäre eine
	// Behauptung.
	case msepaket.MenueEigen, msepaket.MenueTrenner:
	}

	return gebaut
}

// symbolplan plant **ein** Symbol in einem Quadrat von [Palettengroesse].
//
// **`nil`, wenn die Schrift den Code nicht kennt.** Das Menü eines Pakets nennt
// Codes, die nur unter einer Bedingung dastehen — `enabled:` hat sie beim
// Auflösen der Fassungen weggenommen. MSE zeigt dann einen Eintrag ohne Bild
// (`makeMenuItem`, Zeilen 508–518); hier fehlt der Plan, und die Oberfläche
// entscheidet, was sie damit tut.
func symbolplan(
	ctx context.Context, code string, stil *msepaket.Feldstil,
	schrift msepaket.Symbolschrift, name string, umgebung *msescript.Umgebung,
) *render.Plan {
	treffer, unbekannt := zerlegeInSymbole(code, schrift)
	if unbekannt != "" || len(treffer) == 0 {
		return nil
	}

	rahmen := render.Rechteck{Breite: Palettengroesse, Hoehe: Palettengroesse}

	return &render.Plan{
		Breite: Palettengroesse,
		Hoehe:  Palettengroesse,
		Ebenen: []render.Ebene{{
			Art:    render.ArtSymbolzug,
			Feld:   stil.Feld,
			Rahmen: rahmen,
			Abschnitte: []render.Abschnitt{{
				Text:    code,
				Symbole: symbolbilder(ctx, treffer, name, umgebung),

				// **Die Maße der Schrift, aber die Größe der Kachel.** Wer
				// `symbolmasse` nähme, bekäme die Größe, die dieses Feld auf
				// der Karte hat — bei den Manakosten 15 Punkte, bei einem
				// Kastensymbol 22 —, und die Palette sähe je Feld anders aus.
				Symbolmasse: render.Symbolmasse{
					Groesse:           Palettengroesse,
					Bildgroesse:       schrift.Bildgroesse,
					AbstandWaagerecht: schrift.AbstandWaagerecht,
					AbstandSenkrecht:  schrift.AbstandSenkrecht,
					Ausrichtung:       symbolausrichtung(stil.Symbolschrift.Ausrichtung),
				},
			}},
			Satz:    render.Satz{Ausrichtung: render.MitteMitte},
			Deckung: 1,
		}},
	}
}

// beschriftungAus gibt die Aufschrift eines Menüeintrags.
//
//	if (m.label.empty())
//	  m.label.default_ = tr(*symbol_font_for_reading(), "menu_item", m.name, capitalize);
//
// `src/data/symbol_font.cpp`, Zeilen 545–549, Fassung f9d9356 vom 2021-07-06.
// Ohne Sprachpaket ist der Rückfall `capitalize` — **jedes Wort** groß, anders
// als bei `name_to_caption` für ein Feld.
func beschriftungAus(eintrag *msepaket.Menueeintrag) string {
	if eintrag.Beschriftung != "" {
		return eintrag.Beschriftung
	}

	return msescript.Grossgeschrieben(eintrag.Name)
}
