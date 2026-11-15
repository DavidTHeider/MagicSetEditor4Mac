package service

// **Münzwurf und Würfel** — Werkstück M15 der Regelmaschine
// (#1617).
//
// # Was hier entsteht und was nicht
//
// **Ein Ergebnis im Ereignisstrom, und sonst nichts.** Was ein Wurf bewirkt,
// steht auf einer Karte — und einen Kartentext auszuführen ist `M8`. Diese
// zwei Befehle sagen nur, **was gefallen ist**.
//
// **Und das genügt, weil der Strom eine Hashkette ist** (G1,
// #1458):
// ein nachträglich geänderter Wurf bricht sie. Das ist die Zusage, die das
// Issue verlangt — und sie ist nicht neu gebaut, sondern **geerbt**.
//
// # Was ausdrücklich nicht zugesagt wird
//
// **Keine faire Zufälligkeit gegenüber einem böswilligen Wirt** — dieselbe
// Grenze wie beim Mischen (#1409). Der Wirt zieht, und er sieht das Ergebnis
// zuerst. Wer das nicht will, braucht ein Verfahren mit Festlegung, und das
// steht in keinem Werkstück dieser Phase.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Muenzseiten sind die zwei Seiten einer Münze — CR 705.1.
//
// **„Kopf" und „Zahl" und nicht 0 und 1**: CR 705.1 spricht von zwei
// unterscheidbaren Seiten, und ein Mensch liest am Tisch ein Wort.
const (
	Kopf = "kopf"
	Zahl = "zahl"
)

// Wuerfelgrenze ist die größte Seitenzahl, die dieser Sandkasten annimmt.
//
// # Warum es überhaupt eine gibt
//
// **Weil eine Zahl aus einer Nutzlast von außen kommt** (§8). Ein Würfel mit
// zwei Milliarden Seiten ist kein Würfel, sondern ein Zahlengenerator — und
// die Zuteilung dahinter wäre eine Einladung. **Zwanzig ist der größte
// Würfel, den Magic kennt** (CR 706.1 nennt keine Grenze; die Karten tun es).
// Hundert steht hier, weil ein Sandkasten großzügiger sein darf als das
// Spiel, und weil eine runde Zahl nachlesbar ist.
const Wuerfelgrenze = 100

// zufallsbefehl wirft eine Münze oder würfelt.
func (s *Sandkasten) zufallsbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	if befehl.Art == BefehlMuenzwurf {
		return s.muenzwurf(partie, befehl)
	}

	return s.wuerfeln(partie, befehl)
}

// muenzwurf wirft eine Münze — CR 705.
//
// # Der Ruf ist freiwillig
//
// **CR 705.2 kennt zwei Sorten von Würfen**: solche, bei denen nur zählt, was
// oben liegt, und solche, bei denen jemand **gewinnt**. Welche von beiden es
// ist, steht auf der Karte; hier steht beides im Ereignis — die Seite immer,
// der Ausgang nur, wenn jemand gerufen hat.
func (s *Sandkasten) muenzwurf(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	ruf, _ := befehl.Nutzlast["ruf"].(string) //nolint:errcheck // ohne Ruf zählt nur die Seite

	if ruf != "" && ruf != Kopf && ruf != Zahl {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q ist keine Münzseite — %q oder %q (CR 705.1)",
			domain.ErrValidation, ruf, Kopf, Zahl)
	}

	gezogen, fehler := s.zufall.Zahl(2)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	seite := Kopf
	if gezogen == 1 {
		seite = Zahl
	}

	oeffentlich := map[string]any{
		"sitz": befehl.Absender, "seite": seite,
	}

	if ruf != "" {
		oeffentlich["ruf"] = ruf
		oeffentlich["gewonnen"] = ruf == seite
	}

	return []domain.Ereignis{{
		Art: befehl.Art, Oeffentlich: oeffentlich,
	}}, partie, nil
}

// wuerfeln würfelt — CR 706.
//
// **Die Seitenzahl kommt mit** (CR 706.1), die Zahl der Würfel auch. **Ohne
// Veränderer** (CR 706.2): die kommen aus Effekten, und die sind `M8`.
func (s *Sandkasten) wuerfeln(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	seiten, fehler := ganzeZahl(befehl, "seiten")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	if seiten < 2 || seiten > Wuerfelgrenze {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: ein Würfel hat zwischen 2 und %d Seiten, hier stehen %d "+
				"(CR 706.1)", domain.ErrValidation, Wuerfelgrenze, seiten)
	}

	wieviele := 1

	if _, steht := befehl.Nutzlast["wuerfel"]; steht {
		wieviele, fehler = ganzeZahl(befehl, "wuerfel")
		if fehler != nil {
			return nil, domain.Partie{}, fehler
		}
	}

	if wieviele < 1 || wieviele > Wuerfelgrenze {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: zwischen 1 und %d Würfeln, hier stehen %d (CR 706.1)",
			domain.ErrValidation, Wuerfelgrenze, wieviele)
	}

	// **Die Augen stehen einzeln da und nicht als Summe** — CR 706.4 sagt,
	// dass der Kartentext entscheidet, was mit ihnen geschieht; eine Summe
	// wäre schon eine Auswertung.
	augen := make([]any, 0, wieviele)

	for range wieviele {
		gezogen, fehler := s.zufall.Zahl(seiten)
		if fehler != nil {
			return nil, domain.Partie{}, fehler
		}

		// **Ab eins gezählt** — CR 706.2 spricht von der Zahl auf der
		// oberen Fläche, und die fängt bei eins an.
		augen = append(augen, float64(gezogen+1))
	}

	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz": befehl.Absender, "seiten": float64(seiten),
			"augen": augen,
		},
	}}, partie, nil
}
