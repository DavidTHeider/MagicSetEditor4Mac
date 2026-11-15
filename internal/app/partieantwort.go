package app

// **Aus einer Sicht wird eine Antwort** — die Übersetzung und sonst nichts.
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, welche Gestalt das Fenster bekommt** — die DTOs und die
// Schnittstelle, über die sie entstehen. Hier steht, **wie** aus einer
// [domain.Partiesicht] eine wird. Das ist die zweite Frage, und `partie.go`
// stand mit beidem bei 410 Zeilen (§2) — die Legendenwahl aus `M12`
// (#1614)
// war die Zeile, die darüber hinausging.

import (
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// partieAus macht aus einer Sicht ein DTO.
func partieAus(sicht domain.Partiesicht) PartieDTO {
	dto := PartieDTO{
		ID:      sicht.PartieID,
		Fassung: strconv.FormatInt(sicht.Fassung, 10),
		Spiel:   sicht.Spiel,
		Stand:   string(sicht.Stand),
		Fuer:    sicht.Fuer,
		Wirt:    sicht.Wirt,
		Regelwerk: RegelwerkDTO{
			Kennung: sicht.Regelwerk.Kennung,
			Fassung: sicht.Regelwerk.Fassung,
			Strenge: string(sicht.Regelwerk.Strenge),

			// **Was die Runde selbst übernimmt, sieht jeder Sitz** —
			// PLAY-02.AK-03 (#1710): es ist Teil der Vereinbarung und kein
			// Geheimnis.
			Abgewaehlt: alsBereiche(sicht.Regelwerk.Abgewaehlt),
		},
		Sitze:     make([]SitzDTO, 0, len(sicht.Sitze)),
		Zuschauer: make([]ZuschauerDTO, 0, len(sicht.Zuschauer)),
		Objekte:   make([]ObjektDTO, 0, len(sicht.Objekte)),
		Mischpflichten: make([]MischpflichtDTO, 0,
			len(sicht.Mischpflichten)),
		Legendenwahlen: make([]LegendenwahlDTO, 0,
			len(sicht.Legendenwahlen)),
		Ausgeloest:       make([]AusloeserDTO, 0, len(sicht.Ausgeloest)),
		Ausloeserreihung: sicht.Ausloeserreihung,
		Vorrat:           kennungstexte(sicht.Vorrat),
		Zug:              zugAus(sicht.Zug),
		Ausgang:          ausgangAus(sicht.Ausgang),
	}

	for _, pflicht := range sicht.Mischpflichten {
		dto.Mischpflichten = append(dto.Mischpflichten, MischpflichtDTO{
			Sitz: pflicht.Sitz, Zone: string(pflicht.Zone),
		})
	}

	for _, wahl := range sicht.Legendenwahlen {
		dto.Legendenwahlen = append(dto.Legendenwahlen, LegendenwahlDTO{
			Sitz: wahl.Sitz, Name: wahl.Name,
		})
	}

	for _, einer := range sicht.Ausgeloest {
		dto.Ausgeloest = append(dto.Ausgeloest, AusloeserDTO{
			Kennung: einer.Kennung, Art: string(einer.Art),
			Beherrscher: einer.Beherrscher, Wen: einer.Wen,
		})
	}

	for _, sitz := range sicht.Sitze {
		dto.Sitze = append(dto.Sitze, SitzDTO{
			Kennung:     sitz.Kennung,
			Name:        sitz.Name,
			Leben:       sitz.Leben,
			Vergiftung:  sitz.Vergiftung,
			Manavorrat:  service.Vorratstext(sitz.Manavorrat),
			Deckabdruck: sitz.Deckabdruck,
			Karten:      sitz.Karten,
			Wartet:      sitz.Wartet,
			Eingeladen:  sitz.Eingeladen,
			ProfilID:    profilkennung(sitz.ProfilID),

			// **Die fünf aus PLAY-03** — sie kamen bis zum 2026-09-21 nicht
			// über diese Naht, obwohl der Kern sie führt.
			Strahlung:        sitz.Strahlung,
			Commander:        sitz.Commander,
			Commandersteuer:  sitz.Commandersteuer,
			Leergezogen:      sitz.Leergezogen,
			Commanderschaden: sitz.Commanderschaden,

			// **Der Beginn** (#1839).
			Eroeffnung: string(sitz.Eroeffnung),
			Mulligans:  sitz.Mulligans,
			Unterlegen: sitz.Unterlegen,
		})
	}

	for _, zuschauer := range sicht.Zuschauer {
		dto.Zuschauer = append(dto.Zuschauer, ZuschauerDTO{
			Kennung: zuschauer.Kennung, Name: zuschauer.Name,
		})
	}

	for _, objekt := range sicht.Objekte {
		dto.Objekte = append(dto.Objekte, objektAus(objekt))
	}

	return dto
}

// objektAus macht aus einem Sichtobjekt ein DTO.
func objektAus(objekt domain.Sichtobjekt) ObjektDTO {
	zaehler := objekt.Zaehler
	if zaehler == nil {
		// **Eine leere Abbildung und kein `nil`** — sonst steht im Fenster
		// `zaehler: null`, und jede Schleife darüber fällt um.
		zaehler = map[string]int{}
	}

	return ObjektDTO{
		Kennung:       objekt.Kennung,
		Art:           string(objekt.Art),
		Beherrscher:   objekt.Beherrscher,
		Eigner:        objekt.Eigner,
		Zone:          string(objekt.Zone),
		Karte:         objekt.Karte,
		Name:          objekt.Name,
		KarteID:       kennungstext(objekt.KarteID),
		Bild:          objekt.Bild,
		Eigenschaften: eigenschaftenAus(objekt.Eigenschaften),
		Ordnung:       objekt.Ordnung,
		Getappt:       objekt.Getappt,
		Verdeckt:      objekt.Verdeckt,
		Geheim:        objekt.Geheim,
		HaengtAn:      objekt.HaengtAn,
		Haufen:        objekt.Haufen,
		Zaehler:       zaehler,
	}
}

// kennungstext macht aus einer Kennung eine Zeichenkette — leer bei null.
//
// **Leer und nicht `"0"`**: eine Null ist keine Zeile im Bestand, und das
// Fenster soll die Frage „gibt es ein Bild?" mit einem Blick beantworten
// können statt mit einem Vergleich.
func kennungstext(kennung int64) string {
	if kennung == 0 {
		return ""
	}

	return strconv.FormatInt(kennung, 10)
}

// ereignisseAus macht aus Ereignissen die DTOs für **einen** Empfänger.
//
// **Die Projektion steht hier und nicht beim Aufrufer.** Wer sie vergäße,
// schickte jedem alles — und die Antwort sähe vollständiger aus, nicht
// falscher.
func ereignisseAus(ereignisse []domain.Ereignis, fuer string) []EreignisDTO {
	gefunden := make([]EreignisDTO, 0, len(ereignisse))

	for _, ereignis := range ereignisse {
		gefunden = append(gefunden, ereignisDTO(ereignis.Fuer(fuer)))
	}

	return gefunden
}

// ereignisDTO macht aus einem **schon projizierten** Ereignis die Gestalt
// fürs Fenster.
//
// # Warum es das neben [ereignisseAus] gibt
//
// **Weil nicht jedes Ereignis hier projiziert werden darf.** Was über die
// Leitung kommt, hat der Wirt schon für **diesen** Sitz gekürzt (#1627) —
// und dieser Rechner weiß seinen eigenen Sitz am fremden Tisch gar nicht.
// Ein zweites `Fuer("")` warf weg, was gerade angekommen war: ein Gast sah
// nie, was er selbst gezogen hatte.
func ereignisDTO(gekuerzt domain.Ereignis) EreignisDTO {
	dto := EreignisDTO{
		Folge:       strconv.FormatInt(gekuerzt.Folge, 10),
		Art:         gekuerzt.Art,
		Oeffentlich: gekuerzt.Oeffentlich,
		Kette:       gekuerzt.Kette,
	}

	if dto.Oeffentlich == nil {
		dto.Oeffentlich = map[string]any{}
	}

	// **Eine Ergänzung je Empfänger, also höchstens eine** — die Projektion
	// lässt nur die seine übrig.
	if len(gekuerzt.Privat) > 0 {
		dto.Privat = gekuerzt.Privat[0].Inhalt
	}

	if dto.Privat == nil {
		dto.Privat = map[string]any{}
	}

	return dto
}

// profilkennung macht aus einer Profilkennung ihren Text — **null wird leer**.
//
// **Und nicht „0"**: eine Oberfläche, die darauf prüfte, prüfte auf eine
// Zeichenkette, die wie eine Kennung aussieht. Leer heißt: dieser Sitz gehört
// zu keinem Profil, und das ist der Normalfall (G8, #1465).
func profilkennung(profilID int64) string {
	if profilID == 0 {
		return ""
	}

	return strconv.FormatInt(profilID, 10)
}
