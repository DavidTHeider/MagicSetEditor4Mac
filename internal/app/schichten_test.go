//go:build fremdpakete

// **Eine Stichprobe, die die Fälle trifft**
// (#610).
//
// # Warum die alte sie nicht traf
//
// Sie nahm jede n-te Karte. Über 25.148 Karten fand der volle Lauf 43 Befunde,
// die Stichprobe von 680 fand **null** — und die 40, um die es ging, saßen an
// Karten, die eine **Stiloption** teilen: `frames: conspiracy` und
// `draft matters` (#664).
//
// Eine Rate hilft gegen unterschiedliche **Größen** der Stichprobe. Sie hilft
// nicht gegen eine Stichprobe, die den Fall gar nicht **enthält**.
//
// # Wonach geschichtet wird — und wonach nicht
//
// Nachgemessen über `Colorless.mse-set`, 3.216 Karten, davon 25 mit
// `conspiracy`:
//
//	Schichtung nach                | Schichten | wo die 25 liegen
//	-------------------------------|-----------|----------------------------
//	Feldern der Karte (`werte`)    |         5 | im größten Topf: 25 von 3.181
//	Stilangaben (`styling_data`)   |        16 | zwei eigene Töpfe, 13 und 12
//
// **Nach den Feldern zu schichten hätte sie nicht getroffen.** Nach den
// Stilangaben trifft eine Ziehung von einer je Schicht sie mit Sicherheit.
//
// # Woher die Stilangaben kommen
//
// Aus der Setdatei, über `msepaket` — dasselbe Werkzeug, mit dem die Anwendung
// sie liest. Sie stehen **nicht** im `KarteDTO`, und dort einen Eintrag
// anzulegen, den nur ein Messlauf braucht, wäre eine Änderung am Erzeugnis für
// einen Test.
package app

import (
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stilschichten gibt je Karte des Sets ihre Stilsignatur, in der Reihenfolge
// der Datei.
//
// **Die Reihenfolge ist der Schlüssel**, und das ist keine Bequemlichkeit: der
// Import behält sie, und `cards.reihenfolge` ist genau sie
// (#623). Eine
// Zuordnung über den Namen ginge daneben, sobald zwei Karten gleich heißen.
//
// **Ohne lesbare Datei gibt es keine Schichten**, und der Aufrufer fällt auf
// die verteilte Stichprobe zurück — ein Messlauf, der an einer Vorrichtung
// scheitert, misst nichts.
func stilschichten(t *testing.T, pfad string) []string {
	t.Helper()

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Logf("  (keine Schichtung: %s ließ sich nicht öffnen: %v)", pfad, err)

		return nil
	}

	defer func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("%s schließen: %v", pfad, fehler)
		}
	}()

	roh, err := paket.Datei("set")
	if err != nil {
		t.Logf("  (keine Schichtung: %s führt keine Datei `set`: %v)", pfad, err)

		return nil
	}

	baum, err := msepaket.Zerteile("set", roh)
	if err != nil {
		t.Logf("  (keine Schichtung: %s ließ sich nicht zerteilen: %v)", pfad, err)

		return nil
	}

	schichten := make([]string, 0, 256)

	for _, knoten := range baum.Wurzel {
		if knoten.Schluessel != "card" {
			continue
		}

		schichten = append(schichten, signaturVon(knoten))
	}

	return schichten
}

// signaturVon fasst zusammen, was an einer Karte den Zeichenweg verzweigt.
//
// **Der Stilbogen und die Stilangaben, nicht die Feldwerte.** Ein Kartenname
// ändert den Weg nicht; `frames: conspiracy` schlägt eine andere Ebene auf.
//
// **`has_styling` bleibt draußen**: es sagt nur, *dass* Angaben folgen, und
// steht bei jeder Karte, die welche hat.
func signaturVon(karte *msepaket.Knoten) string {
	teile := make([]string, 0, 8)

	for _, kind := range karte.Kinder {
		switch kind.Schluessel {
		case "stylesheet":
			teile = append(teile, "stylesheet="+kind.Wert)

		case "styling_data":
			for _, angabe := range kind.Kinder {
				if angabe.Schluessel == "" {
					continue
				}

				teile = append(teile, angabe.Schluessel+"="+angabe.Wert)
			}
		}
	}

	sort.Strings(teile)

	return strings.Join(teile, "|")
}

// geschichteteStichprobe zieht **mindestens eine Karte je Schicht**.
//
// # Die Reihenfolge der Ziehung
//
// Erst eine je Schicht — die seltenste zuerst, damit eine Schicht mit zwei
// Karten nicht hinter einer mit dreitausend verhungert. Bleibt Platz, wird nach
// Größe aufgefüllt.
//
// **Reicht `anzahl` nicht für alle Schichten, gewinnen die seltenen.** Das ist
// die Absicht: die häufige Schicht ist die, die eine verteilte Stichprobe
// ohnehin trifft.
//
// **Ohne Schichten bleibt es bei der verteilten Stichprobe.** Sie ist nicht
// falsch — nur blind für Fälle, die sich zusammenrotten.
func geschichteteStichprobe(
	karten []KarteDTO, schichten []string, anzahl int,
) []KarteDTO {
	if len(schichten) != len(karten) || anzahl <= 0 {
		return verteilteStichprobe(karten, anzahl)
	}

	if len(karten) <= anzahl {
		return karten
	}

	nachSchicht := map[string][]int{}
	for stelle, schicht := range schichten {
		nachSchicht[schicht] = append(nachSchicht[schicht], stelle)
	}

	namen := make([]string, 0, len(nachSchicht))
	for name := range nachSchicht {
		namen = append(namen, name)
	}

	// **Nach Größe, bei Gleichstand nach Namen.** Ohne den zweiten Schlüssel
	// hinge die Ziehung an der Laufordnung einer Go-Abbildung, und zwei Läufe
	// über dieselbe Datei zögen verschiedene Karten.
	sort.Slice(namen, func(a, b int) bool {
		links, rechts := nachSchicht[namen[a]], nachSchicht[namen[b]]
		if len(links) != len(rechts) {
			return len(links) < len(rechts)
		}

		return namen[a] < namen[b]
	})

	return ziehe(karten, nachSchicht, namen, anzahl)
}

// ziehe nimmt reihum aus jeder Schicht, bis `anzahl` voll ist.
func ziehe(
	karten []KarteDTO, nachSchicht map[string][]int, namen []string, anzahl int,
) []KarteDTO {
	genommen := make(map[int]bool, anzahl)
	gewaehlt := make([]KarteDTO, 0, anzahl)

	for runde := 0; len(gewaehlt) < anzahl; runde++ {
		vorher := len(gewaehlt)

		for _, name := range namen {
			stellen := nachSchicht[name]
			if runde >= len(stellen) || len(gewaehlt) >= anzahl {
				continue
			}

			// **Verteilt innerhalb der Schicht**, nicht die ersten n: eine
			// Schicht mit dreitausend Karten gäbe sonst dreitausend Nachbarn.
			stelle := stellen[runde*len(stellen)/max(1, min(len(stellen), anzahl))%len(stellen)]
			if genommen[stelle] {
				stelle = stellen[runde]
			}

			if genommen[stelle] {
				continue
			}

			genommen[stelle] = true
			gewaehlt = append(gewaehlt, karten[stelle])
		}

		// **Kommt eine Runde ohne Zuwachs zurück, ist nichts mehr zu holen.**
		// Ohne diesen Ausstieg drehte die Schleife, sobald jede Schicht
		// erschöpft ist und `anzahl` größer ist als die Kartenzahl.
		if len(gewaehlt) == vorher {
			break
		}
	}

	return gewaehlt
}

// zaehleSchichten sagt, über wie viele Schichten gezogen wurde.
//
// **Die Zahl steht in der Ausgabe** (§10): eine Stichprobe ohne ihre Schichten
// ist wieder nur eine Zahl. Null heißt „nicht geschichtet" — dann lag die
// Setdatei nicht lesbar vor, und es galt die verteilte Ziehung.
func zaehleSchichten(schichten []string) int {
	verschieden := map[string]bool{}
	for _, schicht := range schichten {
		verschieden[schicht] = true
	}

	return len(verschieden)
}
