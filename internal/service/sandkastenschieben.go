// **Wohin eine Karte kommt und in welchem Zustand** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Warum das Schieben eine eigene Datei bekommt
//
// **Weil es als Einziges der neun Befehle die ganze Partie angeht.** Tappen,
// Umdrehen und Zähler ändern **ein** Ding; wer eine Karte an die zweite
// Stelle einer Bibliothek legt, rückt jede darunter. Der Schnitt folgt der
// Frage (§2): dort, was mit einem Objekt geschieht, hier, was dabei mit
// seiner Nachbarschaft geschieht.
//
// # Und warum eine Stelle überhaupt nötig ist
//
// **Weil „ins Deck zurück" ohne sie unbestimmt ist.** Oben, unten oder
// eingesteckt sind drei verschiedene Handlungen mit drei verschiedenen
// Folgen. Bis zum 2026-09-15 setzte das Schieben die Ordnung schlicht auf
// null — eine Karte, die in die Bibliothek zurückging, lag damit an keiner
// Stelle, und zwei davon lagen an derselben.

package service

import (
	"fmt"
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schieben legt ein Objekt in eine andere Zone — an eine Stelle und in einen
// Zustand.
//
// **Der Zustand bleibt, wo nichts gesagt ist.** Wer eine getappte Kreatur auf
// den Friedhof legt, legt eine getappte Karte dorthin; dass ein Objekt beim
// Zonenwechsel nach den Regeln ein **neues** wird, ist eine Regelfrage und
// gehört in die Matrix von G4
// (#1461).
func (s *Sandkasten) schieben(
	danach *domain.Partie, stelle int, befehl domain.Spielbefehl,
) (map[string]any, error) {
	ziel, fehler := gemeinteZone(befehl)
	if fehler != nil {
		return nil, fehler
	}

	wohin, fehler := gemeinteStelle(befehl, ziel)
	if fehler != nil {
		return nil, fehler
	}

	objekt := &danach.Objekte[stelle]
	vorher := objekt.Zone
	kennung := objekt.Kennung

	// **Ein Zonenwechsel macht ein neues Objekt** — CR 400.7 (M18,
	// #1619).
	// Marken, Zustand, markierter Schaden und was daran hing, bleiben
	// zurück.
	*objekt = domain.NeuesObjekt(*objekt, danach.Zug.Nummer)
	objekt.Zone = ziel

	// **Und danach, was der Befehl sagt.** „Lege sie getappt aufs Feld" ist
	// eine Angabe über das **neue** Objekt; wer sie vor CR 400.7 setzte,
	// setzte sie an das alte.
	zustaendeSetzen(objekt, befehl)

	// **Erst schließt sich die Lücke, dann wird eingefügt.** Andersherum
	// zählte das Einfügen gegen eine Reihe mit einem Loch, und die Stelle
	// „unten" landete eine zu früh.
	if vorher.Geordnet() {
		durchnummerieren(danach, stelle, vorher)
	}

	if ziel.Geordnet() {
		if fehler := s.einsortieren(danach, stelle, wohin); fehler != nil {
			return nil, fehler
		}
	}

	// **Und was daran hing, hängt an nichts mehr** — dieselbe Regel aus der
	// anderen Richtung.
	*danach = danach.Abgeloest(kennung)

	objekt = &danach.Objekte[stelle]

	geschehen := map[string]any{"von": string(vorher), "nach": string(ziel)}

	if ziel.Geordnet() {
		// **Die Stelle steht nur im öffentlichen Teil, wenn die Zone offen
		// ist.** In einer Bibliothek wäre sie das Geheimnis selbst — und
		// eine Zahl reicht, um den nächsten Zug zu kennen.
		if ziel.Sicht() == domain.SichtOffen {
			geschehen["stelle"] = float64(objekt.Ordnung)
		}
	}

	return geschehen, nil
}

// gemeinteStelle liest die Stelle und hält sie gegen die Zone.
//
// **Eine Stelle in einer ungeordneten Zone ist ein Irrtum und kein Wunsch.**
// Wer „oben auf den Friedhof" sagt, meint etwas, das es nicht gibt; eine
// Antwort, die einfach ablegt, ließe ihn glauben, es sei geschehen.
//
// **Und ohne Angabe gilt oben.** Nicht, weil es das Häufigste wäre, sondern
// weil es das **Sichtbarste** ist: was oben liegt, kommt als Nächstes, und
// eine falsche Vorgabe fällt sofort auf. „Unten" fiele vierzig Züge lang
// niemandem auf.
func gemeinteStelle(
	befehl domain.Spielbefehl, zone domain.Zone,
) (domain.Zonenstelle, error) {
	genannt, _ := befehl.Nutzlast["stelle"].(string) //nolint:errcheck // siehe [Sandkasten.objektstelle]

	if genannt == "" {
		return domain.ZonenstelleOben, nil
	}

	if !zone.Geordnet() {
		return "", fmt.Errorf(
			"%w: in %q gibt es keine Stelle, an die sich etwas legen ließe",
			domain.ErrValidation, zone)
	}

	if stelle := domain.Zonenstelle(genannt); domain.IstZonenstelle(stelle) {
		return stelle, nil
	}

	return "", fmt.Errorf(
		"%w: %q ist keine Stelle — es gibt oben, unten und zufaellig",
		domain.ErrValidation, genannt)
}

// zustaendeSetzen setzt, was der Befehl über den Zustand sagt.
//
// **Nur, was dasteht.** Ein fehlendes Feld heißt „lass, wie es ist" und nicht
// „setze auf falsch" — sonst enttappte jedes Schieben.
func zustaendeSetzen(objekt *domain.Spielobjekt, befehl domain.Spielbefehl) {
	if getappt, dabei := befehl.Nutzlast["getappt"].(bool); dabei {
		objekt.Getappt = getappt
	}

	if verdeckt, dabei := befehl.Nutzlast["verdeckt"].(bool); dabei {
		objekt.Verdeckt = verdeckt
	}
}

// einsortieren legt das Objekt an seine Stelle und rückt die anderen.
func (s *Sandkasten) einsortieren(
	danach *domain.Partie, stelle int, wohin domain.Zonenstelle,
) error {
	andere := ordnungsgruppe(*danach, stelle)

	platz, fehler := s.platzFuer(wohin, len(andere))
	if fehler != nil {
		return fehler
	}

	for _, nachbar := range andere {
		if danach.Objekte[nachbar].Ordnung >= platz {
			danach.Objekte[nachbar].Ordnung++
		}
	}

	danach.Objekte[stelle].Ordnung = platz

	return nil
}

// platzFuer gibt die Stelle, an die gelegt wird — 1 bis anzahl+1.
func (s *Sandkasten) platzFuer(
	wohin domain.Zonenstelle, anzahl int,
) (int, error) {
	switch wohin {
	case domain.ZonenstelleOben:
		return 1, nil

	case domain.ZonenstelleUnten:
		return anzahl + 1, nil

	case domain.ZonenstelleZufaellig:
	}

	// **Gefragt wird nach einer Mischung über alle Plätze, genommen ihr
	// erster.** So braucht diese Naht kein zweites Verfahren für „eine
	// Zahl" — und eine Quelle, die für Läufe eingesetzt wird, hat nur eines
	// zu erfüllen.
	mischung, fehler := s.zufall.Mische(anzahl + 1)
	if fehler != nil {
		return 0, fehler
	}

	if fehler := pruefeMischung(mischung, anzahl+1); fehler != nil {
		return 0, fehler
	}

	return mischung[0] + 1, nil
}

// durchnummerieren schließt die Lücke, die ein Objekt hinterlässt.
//
// **Ohne das bliebe ein Loch in der Reihe.** Wer die oberste Karte zieht,
// ließe eine Bibliothek zurück, die bei zwei anfängt — und „oben" hieße
// danach etwas anderes als vorher.
func durchnummerieren(danach *domain.Partie, stelle int, zone domain.Zone) {
	gruppe := []int{}

	for i := range danach.Objekte {
		if i != stelle && danach.Objekte[i].Zone == zone &&
			teiltOrdnung(danach.Objekte[i], danach.Objekte[stelle], zone) {
			gruppe = append(gruppe, i)
		}
	}

	sort.Slice(gruppe, func(eins, zwei int) bool {
		return danach.Objekte[gruppe[eins]].Ordnung <
			danach.Objekte[gruppe[zwei]].Ordnung
	})

	for lauf, nachbar := range gruppe {
		danach.Objekte[nachbar].Ordnung = lauf + 1
	}
}

// ordnungsgruppe gibt die Stellen, die sich eine Reihenfolge mit diesem
// Objekt teilen — es selbst nicht dabei.
func ordnungsgruppe(partie domain.Partie, stelle int) []int {
	objekt := partie.Objekte[stelle]
	gruppe := []int{}

	for i := range partie.Objekte {
		if i != stelle && partie.Objekte[i].Zone == objekt.Zone &&
			teiltOrdnung(partie.Objekte[i], objekt, objekt.Zone) {
			gruppe = append(gruppe, i)
		}
	}

	return gruppe
}

// teiltOrdnung sagt, ob zwei Objekte in derselben Reihe liegen.
//
// # Der Stapel gehört allen, die Bibliothek einem
//
// **Und das ist kein Detail.** Zwei Bibliotheken sind zwei Reihen: Annas
// oberste Karte hat nichts mit Berts zu tun. Der Regelstapel ist **eine**
// Reihe für den ganzen Tisch — wer dort je Sitz nummerierte, hätte zwei
// „obersten" Zauber, und keiner löste sich zuerst auf.
func teiltOrdnung(eines, anderes domain.Spielobjekt, zone domain.Zone) bool {
	if zone == domain.ZoneStapel {
		return true
	}

	return eines.Beherrscher == anderes.Beherrscher
}
