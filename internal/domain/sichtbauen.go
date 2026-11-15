package domain

// **Aus einer Partie wird eine Sicht** — die Projektion und sonst nichts.
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, *was* eine Sicht ist** — die drei Typen und die Gründe
// dafür, dass sie weniger tragen als die Partie. Hier steht, **wie** sie
// entsteht: was projiziert wird, was jedem gehört und was leer statt `nil`
// hinausgeht. Das ist die zweite Frage, und `partiesicht.go` stand mit
// beidem bei 404 Zeilen (§2) — die wartenden Fähigkeiten aus `M9`
// (#1611)
// waren die Zeilen, die darüber hinausgingen.

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

// Sicht rechnet, was `fuer` sehen darf.
func (p Partie) Sicht(fuer string) Partiesicht {
	return p.SichtVon(fuer, "")
}

// SichtVon rechnet die Sicht so, wie sie im Bestand `hier` gelesen wird — P9
// (#2377),
// Vermerk 0198. **Leer heißt: der Bestand, in dem die Partie liegt**; für
// einen Gast rechnet der Wirt mit dessen Peerkennung.
//
// **Eine Kartenkennung nur aus dem eigenen Bestand**: eine Zeile, die anderswo
// nachgeschlagen würde, fände dort eine andere Karte. Was sich nicht selbst
// zeichnen lässt, erkennt man an der Bildkennung.
func (p Partie) SichtVon(fuer, hier string) Partiesicht {
	// **Keine Kennung geht hinaus** (#2480, Vermerk 0201) — jede wird hier
	// unter ihrer Sichtkennung genannt.
	namen := p.sichtnamen()

	sicht := Partiesicht{
		PartieID:  p.ID,
		Fassung:   p.Fassung,
		Spiel:     p.Spiel,
		Regelwerk: p.Regelwerk,
		Stand:     p.Stand,
		Fuer:      fuer,
		Wirt:      p.Wirt,
		Zug:       p.Zug,
		Ausgang:   p.Ausgang,
		Zuschauer: p.Zuschauer,

		// **Jedem, und ohne Projektion.** Eine Mischpflicht ist die
		// öffentliche Hälfte eines Blicks; wer sie nur dem Betroffenen
		// zeigte, verschwiege dem Gegenüber, was gerade geschehen ist.
		Mischpflichten: p.Mischpflichten,
		Legendenwahlen: p.Legendenwahlen,

		// **Und die wartenden Fähigkeiten ebenso** (M9, #1611) — wer
		// ordnet, ist am Tisch öffentlich.
		Ausgeloest:       p.Ausgeloest,
		Ausloeserreihung: p.Ausloeserreihung,

		// **Und was am Tisch nicht durchgesetzt wird** — CR 702 (M14,
		// #1616).
		// **Für jeden dieselben**: der gedruckte Text eines offenen
		// Permanents ist öffentlich.
		Befunde: p.tischbefunde(),

		Sitze:   make([]Sichtsitz, 0, len(p.Sitze)),
		Objekte: make([]Sichtobjekt, 0, len(p.Objekte)),
	}

	sicht.Zug.Kampf = namen.kampf(p.Zug.Kampf)

	if sicht.Legendenwahlen == nil {
		sicht.Legendenwahlen = []Legendenwahl{}
	}

	if sicht.Ausgeloest == nil {
		sicht.Ausgeloest = []Ausloeser{}
	}

	if sicht.Zuschauer == nil {
		sicht.Zuschauer = []Zuschauer{}
	}

	if sicht.Mischpflichten == nil {
		// **Eine leere Liste und kein `nil`** — derselbe Satz wie eine Zeile
		// darüber, und derselbe Grund: im Fenster stünde sonst `null`, und
		// jede Schleife darüber fiele um.
		sicht.Mischpflichten = []Mischpflicht{}
	}

	for _, sitz := range p.Sitze {
		// **Wen er eingeladen hat, sieht nur der Wirt** (P9b, #2377).
		eingeladen := ""
		if p.IstWirt(fuer) {
			eingeladen = sitz.Eingeladen
		}

		sicht.Sitze = append(sicht.Sitze, Sichtsitz{
			Kennung:     sitz.Kennung,
			Name:        sitz.Name,
			Deckabdruck: sitz.Deck.Abdruck,
			Wartet:      sitz.Wartet(),
			Eingeladen:  eingeladen,
			Karten:      sitz.Deck.Zaehle(),
			ProfilID:    sitz.ProfilID,
			Leben:       sitz.Leben,
			Vergiftung:  sitz.Vergiftung,

			// **Eine eigene Tabelle je Sicht** — eine Sicht ist eine
			// Projektion und kein Griff in den Zustand.
			Manavorrat: sitz.Manavorrat.Kopie(),

			// **Die fünf, die bis zum 2026-09-21 in der Domäne lagen** —
			// PLAY-03 (#1711).
			//
			// Sie sind gerechnet, gespeichert und geprüft gewesen, und
			// **keine** war je im Fenster zu sehen: die Sicht kopierte acht
			// von zwölf Feldern. Keine davon ist geheim — CR 728.1 (Marken
			// am Spieler), CR 903.3 (die Kommandozone ist öffentlich),
			// CR 903.8, CR 903.10a (Kampfschaden) und CR 704.5b (der Versuch
			// geschieht am Tisch).
			Strahlung:       sitz.Strahlung,
			Commander:       sitz.Commander,
			Commandersteuer: sitz.Commandersteuer,
			Leergezogen:     sitz.Leergezogen,

			// **Der Beginn** (#1839) — öffentlich, der Grund steht am Feld.
			Eroeffnung: sitz.Eroeffnung,
			Mulligans:  sitz.Mulligans,
			Unterlegen: sitz.Unterlegen,

			// **Eine eigene Tabelle je Sicht**, aus demselben Grund wie beim
			// Manavorrat darüber — und unter den Sichtkennungen.
			Commanderschaden: namen.schaden(sitz.Commanderschaden),
		})

		if sitz.Kennung == fuer && sitz.Deck.Quelle == hier {
			sicht.Vorrat = vorratAus(sitz.Deck)
		}
	}

	if sicht.Vorrat == nil {
		sicht.Vorrat = []int64{}
	}

	for _, objekt := range p.Objekte {
		gezeigt := objekt.fuerSicht(fuer)

		// **Die Kartenkennung nur, wo hineingesehen werden darf** (G2,
		// #1459).
		//
		// Sie ist der Weg zum **Bild**: ohne sie zeigte der Tisch Namen statt
		// Karten. Und genau deshalb steht sie unter derselben Regel wie der
		// Name — wer sie an einem geheimen Objekt mitgäbe, verriete die Karte
		// über einen Umweg, und die Projektion daneben sähe weiter richtig
		// aus.
		if !gezeigt.Geheim {
			gezeigt.KarteID, gezeigt.Bild = p.kartenbezug(objekt, hier)
			gezeigt.Eigenschaften = p.sichteigenschaften(objekt)
		}

		gezeigt.Kennung = namen.von(objekt.Kennung)
		gezeigt.HaengtAn = namen.von(gezeigt.HaengtAn)
		gezeigt.Haufen = namen.von(gezeigt.Haufen)

		sicht.Objekte = append(sicht.Objekte, gezeigt)
	}

	sicht.Objekte = geordnet(sicht.Objekte)

	return sicht
}

// kartenbezug gibt die Zeile im Bestand `hier`, aus der ein Objekt stammt —
// null, wenn sein Deck in einen anderen zeigt —, und seine Bildkennung.
//
// **Aus dem Deck des Eigners, nicht des Beherrschers** (P9, #2377): eine
// Karte, die den Beherrscher wechselt, bleibt die Karte aus dem Deck, mit dem
// ihr Eigner kam. Hier stand der Beherrscher; dessen Deck kann dieselbe
// Platzkennung für eine andere Karte führen. **Ohne Eigner kein Bezug**:
// leer ist er nur bei einem Spielstein, und der hat keine Karte.
func (p Partie) kartenbezug(
	objekt Spielobjekt, hier string,
) (zeile int64, bild string) {
	sitz := p.SitzVon(objekt.Eigner)
	if sitz == nil {
		return 0, ""
	}

	for _, karte := range sitz.Deck.Karten {
		if karte.Kennung != objekt.Karte || len(karte.Seiten) == 0 {
			continue
		}

		bild = bildkennung(p.ID, sitz.Kennung, sitz.Deck.Quelle,
			sitz.Deck.Abdruck, karte.Kennung)
		if sitz.Deck.Quelle != hier {
			return 0, bild
		}

		return karte.Seiten[0].KarteID, bild
	}

	return 0, ""
}

// bildkennung ist undurchsichtig und an jedem Ende dieselbe: der Hash über
// Partie, Sitz, Quelle, Abzug und Platz — keine Zeile eines Bestands.
//
// **Sitz und Quelle gehören dazu** (Zweitblick über Phase 7, Befund 3): der
// Abdruck fasst Platz, Name, Druck und Menge, nicht die Gestaltung. Zwei
// Sitze mit gleich gebautem Deck teilten sonst einen Zwischenspeicher, und
// das zuerst geladene Bild stünde für beide.
func bildkennung(partie, sitz, quelle, abdruck, platz string) string {
	summe := sha256.Sum256([]byte(strings.Join(
		[]string{partie, sitz, quelle, abdruck, platz}, "\x1f")))

	return "b" + hex.EncodeToString(summe[:12])
}

// Vollzustand gibt die Partie ohne Projektion.
//
// **Der benannte Weg des Wirts**, und er heißt so, damit man ihn suchen kann.
// Wer ihn ruft, hält den Zustand — er zeigt ihn nicht her.
func (p Partie) Vollzustand() Partiesicht {
	sicht := p.Sicht(p.Wirt)

	sicht.Objekte = sicht.Objekte[:0]

	for _, objekt := range p.Objekte {
		sicht.Objekte = append(sicht.Objekte, objekt.offen())
	}

	return sicht
}

// vorratAus sammelt die Kartenkennungen eines Abzugs.
//
// **Sortiert und ohne Wiederholung.** Die Reihenfolge des Abzugs ist die, in
// der jemand sein Deck getippt hat; sie hier durchzureichen sagte etwas über
// die Liste, ohne dass es jemand wollte. Vier Wälder sind ein Bild, also ein
// Eintrag.
//
// **Und das Seitendeck geht mit.** Wer es einwechselt, braucht die Bilder im
// selben Augenblick — und dass es existiert, sagt [Partiedeck.Zaehle] ohnehin.
func vorratAus(abzug Partiedeck) []int64 {
	gesehen := map[int64]bool{}
	gesammelt := []int64{}

	for _, karte := range abzug.Karten {
		for _, seite := range karte.Seiten {
			if seite.KarteID == 0 || gesehen[seite.KarteID] {
				continue
			}

			gesehen[seite.KarteID] = true
			gesammelt = append(gesammelt, seite.KarteID)
		}
	}

	slices.Sort(gesammelt)

	return gesammelt
}
