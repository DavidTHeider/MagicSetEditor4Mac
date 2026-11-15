package flaechen_test

// **Blick und Mischung über jede Fläche** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum das einen eigenen Lauf bekommt
//
// **Weil hier ein Geheimnis über eine Fläche geht.** Ein Tappen ist überall
// gleich langweilig; ein Blick in die eigene Bibliothek gibt an **einen**
// Empfänger heraus, was sonst niemand sieht — und jede Fläche muss dabei
// dasselbe tun und dasselbe verschweigen.
//
// **Und weil die Mischpflicht die Gegenrechnung ist.** Sie steht danach für
// **jeden** da, auch für den, der gar nicht hingesehen hat. Eine Fläche, die
// sie wegließe, machte aus einer Abmachung eine Einbahnstraße.

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein Blick gibt den Inhalt nur seinem Absender — an jeder Fläche.**
func TestEinBlickGehtAnJederFlaecheNurAnSeinenAbsender(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	for _, fall := range []struct {
		name    string
		blicken func(partieID string) string
	}{
		{"Kommandozeile", func(partieID string) string {
			return mussKommandozeile(t, umgebung, "partie-befehl",
				partieID, "anna", "b1", "zone.blick",
				`{"zone":"bibliothek"}`, "--json")
		}},
		{"MCP", func(partieID string) string {
			return mussMCP(t, umgebung, "schreiben.partie.befehl",
				map[string]any{
					"partieID": partieID, "absender": "anna",
					"kennung": "b1", "art": "zone.blick",
					"nutzlast": map[string]any{"zone": "bibliothek"},
				})
		}},
		{"REST", func(partieID string) string {
			return mussREST(t, an, http.MethodPost,
				"/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"b1","art":"zone.blick",`+
					`"nutzlast":{"zone":"bibliothek"}}`)
		}},
	} {
		partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

		antwort := fall.blicken(partieID)

		// **Der Inhalt kommt an jeder Fläche zurück.**
		//
		// **Das war der erste Anlauf nicht**: MCP und Dienst gaben ihre
		// Ereignisse ohne `oeffentlich` und `privat`, und damit hatte dort
		// gesucht, wer nichts fand — eine Fähigkeit, die an drei von vier
		// Flächen ins Leere lief (§16). Dieser Lauf hat es angestrichen.
		if !strings.Contains(antwort, "karten") {
			t.Errorf("%s: der Blick gibt keinen Inhalt: %s", fall.name, antwort)
		}

		// **Die Mischpflicht steht danach für jeden da** — auch für Bert,
		// der nicht hingesehen hat.
		for _, sitz := range []string{"anna", "bert"} {
			bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, sitz, "--json"))

			if len(bild.Mischpflichten) != 1 ||
				bild.Mischpflichten[0].Sitz != "anna" ||
				bild.Mischpflichten[0].Zone != "bibliothek" {
				t.Errorf("%s: %s sieht %+v, erwartet Annas Bibliothek",
					fall.name, sitz, bild.Mischpflichten)
			}
		}

		// **Und was Bert sieht, ist nicht der Inhalt.** Er weiß, *dass*
		// gesucht wurde — nicht, *was* gefunden wurde.
		pruefeVerborgenFuerBert(t, fall.name, umgebung, partieID)
	}
}

// pruefeVerborgenFuerBert verlangt, dass Bert keine Karte von Anna liest —
// weder aus ihrer Bibliothek noch von ihrer Hand.
//
// **Je Objekt und nicht im Text der Antwort**: hier stand bis #1851 die
// Suche nach `"karte":"p0-` in allem, was Bert sieht. Seit dem Beginn hält
// er eine eigene Starthand aus demselben Druck, und die liest er zu Recht —
// die Suche im Text fände sie und meldete einen Befund, den es nicht gibt.
func pruefeVerborgenFuerBert(t *testing.T, wer string, umgebung []string, partieID string) {
	t.Helper()

	gesehen := 0

	for _, objekt := range bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "bert", "--json")).Objekte {
		if objekt.Beherrscher != "anna" {
			continue
		}

		gesehen++

		if !objekt.Geheim || objekt.Karte != "" || objekt.Name != "" {
			t.Errorf("%s: Bert liest Annas Karte: %+v", wer, objekt)
		}
	}

	if gesehen != deckmenge {
		t.Errorf("%s: Bert sieht %d von Annas Karten, erwartet %d verdeckt",
			wer, gesehen, deckmenge)
	}
}

// **Und ein Mischen löst sie — an jeder Fläche.**
func TestEinMischenLoestDiePflichtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	for _, fall := range []struct {
		name    string
		mischen func(partieID string)
	}{
		{"Kommandozeile", func(partieID string) {
			mussKommandozeile(t, umgebung, "partie-befehl",
				partieID, "anna", "m1", "zone.mischen",
				`{"zone":"bibliothek"}`)
		}},
		{"MCP", func(partieID string) {
			mussMCP(t, umgebung, "schreiben.partie.befehl",
				map[string]any{
					"partieID": partieID, "absender": "anna",
					"kennung": "m1", "art": "zone.mischen",
					"nutzlast": map[string]any{"zone": "bibliothek"},
				})
		}},
		{"REST", func(partieID string) {
			mussREST(t, an, http.MethodPost,
				"/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"m1","art":"zone.mischen",`+
					`"nutzlast":{"zone":"bibliothek"}}`)
		}},
	} {
		partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

		mussKommandozeile(t, umgebung, "partie-befehl",
			partieID, "anna", "b1", "zone.blick", `{"zone":"bibliothek"}`)

		fall.mischen(partieID)

		bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
			"partie-zeigen", partieID, "bert", "--json"))

		if len(bild.Mischpflichten) != 0 {
			t.Errorf("%s: nach dem Mischen steht %+v",
				fall.name, bild.Mischpflichten)
		}
	}
}
