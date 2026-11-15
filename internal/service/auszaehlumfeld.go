// **Was die Auszählung neben der Karte braucht** — die Kartenliste des Sets
// und das Zerlegen einer Antwort.
//
// # Warum eine eigene Datei
//
// `auszaehldimension.go` stand an der Zeilengrenze aus §2, als `set.cards`
// dazukam
// (#1356).
// Der Schnitt folgt der Frage: *wie eine Dimension gerechnet wird*, steht
// dort; *was um sie herum gehört*, hier.

package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// mitKartenliste hängt `cards` an die Setabbildung.
//
// **Die vorhandenen Setfelder bleiben.** Sie stehen dort, weil das Spielskript
// sie liest; `cards` ist ein Glied mehr und kein Ersatz.
//
// **Angehängt und nicht vorangestellt**, und der Unterschied ist eine
// Entscheidung: [Wert.Eintrag] nimmt den **ersten** Treffer, also gewänne ein
// Setfeld namens `cards` gegen die Liste.
//
// **Ob MSE es genauso hielte, ist hier nicht geprüft.** `IMPLEMENT_REFLECTION(Set)`
// führt `set_info` (Zeile 196) vor `reflect_cards` (Zeile 202), und `GetMember`
// hält beim ersten Treffer — das *deutet* auf dieselbe Reihenfolge, ist aber
// aus Makros erschlossen und nicht am laufenden Programm nachgemessen (§15).
//
// **Der Fall kommt nicht vor**: gemessen am 2026-08-31 über die fünf
// `.mse-game`-Pakete des Packs führt **kein** Spiel ein Setfeld namens `cards`
// oder `card`; das einzige mit „card" im Namen heißt `automatic card numbers`.
// Deshalb steht hier keine Regel, die niemand prüfen kann, sondern die
// einfachere Form — und dieser Absatz.
func mitKartenliste(
	umgebung *msescript.Umgebung, karten []msescript.Eintrag,
) msescript.Wert {
	vorhanden, _ := umgebung.Suche("set")

	eintraege := make([]msescript.Eintrag, 0, len(vorhanden.Eintraege)+1)
	eintraege = append(eintraege, vorhanden.Eintraege...)
	eintraege = append(eintraege, msescript.Eintrag{
		Name: "cards",
		Wert: msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: karten},
	})

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// zerlegteListe zerlegt einen Wert an Kommas — MSEs `splitList`.
//
//	size_t comma = v.find_first_of(',');
//	while (comma != String::npos) {
//	  … e2->values[axis] = v.substr(0,comma);
//	  if (is_substr(v, comma, ", ")) ++comma; // skip space after it
//	  v = v.substr(comma + 1);
//	  comma = v.find_first_of(',');
//	}
//
// **Das übersprungene Leerzeichen ist der Punkt.** Ohne es hieße die zweite
// Gruppe `" blau"` und stünde neben `"blau"` als eigener Balken.
func zerlegteListe(wert string) []string {
	glieder := strings.Split(wert, ",")

	gebaut := make([]string, 0, len(glieder))
	for stelle, glied := range glieder {
		if stelle > 0 {
			glied = strings.TrimPrefix(glied, " ")
		}

		gebaut = append(gebaut, glied)
	}

	return gebaut
}

// Auszaehlangaben übersetzt die Dimensionen eines Pakets in Angaben.
//
// **Exportiert seit dem 2026-08-31** (#1355).
// Der Lauf hinter `fremdpakete` baute die Angabe **von Hand** und ließ dabei
// `Globalskript` weg — eine Vorrichtung, die weniger trägt als der Betrieb,
// und genau die Falle aus §3. Jetzt gibt es die Abbildung einmal.
func Auszaehlangaben(dimensionen []msepaket.Auszaehldimension) []Auszaehlangabe {
	angaben := make([]Auszaehlangabe, 0, len(dimensionen))

	for _, eine := range dimensionen {
		angaben = append(angaben, Auszaehlangabe{
			Name:          eine.Name,
			Beschreibung:  eine.Beschreibung,
			Skript:        eine.Skript,
			Globalskript:  eine.Globalskript,
			Numerisch:     eine.Numerisch,
			Klassenbreite: eine.Klassenbreite,
			LeereZeigen:   eine.LeereZeigen,
			Zerlegen:      eine.Zerlegen,
			Gruppen:       eine.Gruppen,
			Farben:        eine.Farben,
		})
	}

	return angaben
}
