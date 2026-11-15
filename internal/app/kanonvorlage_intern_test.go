//go:build fremdpakete

// Der Abstand **je Vorlage** — die zweite Hälfte von
// #498.
//
// # Warum ein eigener Lauf
//
// `TestDerAbstandZuMSE2AmKanon` nimmt jede n-te Karte. Diese Stichprobe ist
// **proportional**, und damit trifft sie, was häufig ist: über 510 Karten
// waren es 444-mal `m15-altered`, 60-mal `archenemy` und sechs Karten aus
// allen übrigen Vorlagen zusammen. Ein Planeswalker kam nicht vor.
//
// **Für eine Aussage über eine Klasse taugt sie deshalb nicht.** Wer sie
// trotzdem so läse, hätte eine Zahl, deren Reichweite ihre Frage nicht deckt
// (§10).
//
// Hier wird stattdessen **je Vorlage** gezogen: aus jeder Gruppe bis zu
// [KartenJeVorlage] Karten, über die Gruppe verteilt.
//
// # Was dieser Lauf **nicht** ist
//
// Er ist **kein zweites Maß für den Abstand**. Der Mittelwert über alle
// Vorlagen ist mit den 26,573 aus dem Kanonlauf nicht vergleichbar — dort
// zählt jede Karte gleich, hier zählt jede Vorlage gleich. Der Deckel steht
// deshalb dort und nicht hier.
package app

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// KartenJeVorlage ist die Stichprobe je Vorlage.
//
// **Zwölf**, weil eine Klasse mit 33 Karten je Set (`m15-mainframe-planeswalker`
// in `Red`) davon rund ein Drittel abdeckt und der Lauf trotzdem in Minuten
// bleibt. Ein Mittelwert über zwölf Karten sagt etwas über eine Klasse; einer
// über zwei sagt etwas über zwei Karten.
const KartenJeVorlage = 12

// TestDerAbstandJeVorlage beantwortet, ob eine Vorlagenklasse weiter weg ist.
func TestDerAbstandJeVorlage(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	gesamt := vorlagenstand{}
	gemessen := 0

	for _, satz := range saetzeMitReferenz(t) {
		if _, err := os.Stat(mse2ausgabe(satz)); err != nil {
			continue
		}

		if messeVorlagenEines(t, anwendung, satz, gesamt) {
			gemessen++
		}
	}

	if gemessen == 0 {
		t.Skip("keine MSE2-Ausgabe — scripts/mse2-referenz.sh Archenemy White")
	}

	gesamt.melde(t)
	nenneDieWeitesteVorlage(t, gesamt)
}

// messeVorlagenEines misst ein Set nach Vorlagen gruppiert.
func messeVorlagenEines(
	t *testing.T, anwendung *App, satz string, gesamt vorlagenstand,
) bool {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(t.Context(), filepath.Join(kanonverzeichnis(),
		satz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Errorf("%s öffnen: %+v", satz, geoeffnet.Fehler)

		return false
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Errorf("%s: Karten: %+v", satz, karten.Fehler)

		return false
	}

	desSets := gewaehlterStil(t, anwendung, geoeffnet.SetID)

	for vorlage, gewaehlte := range nachVorlage(karten.Karten, desSets) {
		stand := neuerMessstand()
		stand.setstil = desSets

		abstaende, ohneReferenz := vergleicheMitMSE2(t, anwendung, satz, gewaehlte, stand)
		for _, abstand := range abstaende {
			gesamt.nimm(vorlage, abstand)
		}

		// **Was fehlt, wird genannt**
		// (#535).
		// Hier stand `_`, und damit sah eine Gruppe von zwölf aus wie zwölf,
		// obwohl acht still herausgefallen waren.
		if ohneReferenz > 0 {
			t.Logf("  %-10s %-30s %d von %d ohne Referenzbild",
				satz, vorlage, ohneReferenz, len(gewaehlte))
		}
	}

	return true
}

// nachVorlage gruppiert die Karten und zieht aus jeder Gruppe eine Stichprobe.
func nachVorlage(karten []KarteDTO, desSets string) map[string][]KarteDTO {
	gruppen := map[string][]KarteDTO{}
	for i := range karten {
		vorlage := vorlageDerKarte(karten[i], desSets)
		gruppen[vorlage] = append(gruppen[vorlage], karten[i])
	}

	for vorlage, gruppe := range gruppen {
		gruppen[vorlage] = verteilteStichprobe(gruppe, KartenJeVorlage)
	}

	return gruppen
}

// nenneDieWeitesteVorlage stellt die Antwort auf #498 in einen Satz.
//
// **Gegen den Mittelwert der übrigen**, nicht gegen den Gesamtmittelwert: eine
// Klasse in ihrem eigenen Durchschnitt zu spiegeln macht den Unterschied
// kleiner, als er ist.
func nenneDieWeitesteVorlage(t *testing.T, stand vorlagenstand) {
	t.Helper()

	namen := make([]string, 0, len(stand))
	for name, abstaende := range stand {
		if len(abstaende) >= MindestensJeVorlage {
			namen = append(namen, name)
		}
	}

	if len(namen) < 2 {
		t.Logf("zu wenige Vorlagen mit mindestens %d Karten — keine Aussage",
			MindestensJeVorlage)

		return
	}

	sort.Slice(namen, func(i, j int) bool {
		return mittel(stand[namen[i]]) > mittel(stand[namen[j]])
	})

	weiteste := namen[0]

	var uebrige []float64

	for _, name := range namen[1:] {
		uebrige = append(uebrige, stand[name]...)
	}

	t.Logf("am weitesten weg: %q mit %.3f über %d Karten — die übrigen %d "+
		"Vorlagen liegen bei %.3f über %d Karten",
		weiteste, mittel(stand[weiteste]), len(stand[weiteste]),
		len(namen)-1, mittel(uebrige), len(uebrige))
}
