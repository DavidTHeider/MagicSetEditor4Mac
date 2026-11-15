// **Schreibrecht ist keine Kostenfreigabe** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Der Satz, für den es diesen Lauf gibt
//
// Er steht wörtlich im Umfang des Issues:
//
// > autorisierte Übernahme nur auf konkreten Vorschlag; **Schreibrecht allein
// > ersetzt keine erforderliche Daten-/Kostenfreigabe**.
//
// **Der Lauf über den lang laufenden Agenten sagt selbst, dass er ihn nicht
// deckt** (`agentenlauf_test.go`: „Was er nicht deckt: einen Anbieter, der
// Geld kostet"). Hier steht er.
//
// # Und warum er an einem Zähler hängt
//
// **Weil eine Absage nicht sagt, ob vorher schon etwas hinausging.** Der
// Anbieter wird wirklich erreicht — seine Modellliste kommt an —, und genau
// deshalb sagt der zweite Zähler etwas: an `/v1/messages` darf **nichts**
// ankommen, obwohl derselbe Prozess an dieser Karte schreiben darf.
package flaechen_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Wer schreiben darf, darf deshalb noch kein Geld ausgeben.**
func TestSchreibrechtErsetztKeineKostenfreigabe(t *testing.T) {
	t.Parallel()

	umgebung, wolke := eineEingerichteteWolke(t)
	setID, karteID := eineAbnahmekarte(t, umgebung)

	agent := starteMCPProzess(t, umgebung)

	// **1. Der Prozess darf schreiben** — sonst sagte Schritt 3 nichts über
	// das Geld, sondern über das Recht.
	agent.muss("schreiben.feld", map[string]any{
		"setID": zahl(t, setID), "karteID": karteID,
		"feld": dasFeld, "wert": "Von Hand",
	})

	// **2. Und der Anbieter ist wirklich da.**
	if katalog := agent.muss("lesen.ki.modelle", map[string]any{}); !strings.Contains(
		katalog, wolkenmodell) {
		t.Fatalf("das Wolkenmodell steht nicht im Katalog: %s", katalog)
	}

	// **3. Der Auftrag an dieses Modell geht trotzdem nicht hinaus.**
	text := agent.scheitert("schreiben.ki.vorschlag", map[string]any{
		"karteId": karteID, "anbieter": "anthropic", "alias": wolkenmodell,
		"auftrag": derAuftrag, "uebernehmen": true, "felder": []string{dasFeld},
	}, "Freigabe eines Menschen")

	if strings.Contains(text, "nicht eingeschaltet") {
		t.Errorf("abgewiesen wurde am Schalter statt an der Freigabe: %s", text)
	}

	// **Und es ist wirklich nichts hinausgegangen.**
	if auftraege := wolke.auftraege(); auftraege != 0 {
		t.Errorf("%d Aufträge gingen an den Anbieter", auftraege)
	}

	// **Die Karte trägt, was von Hand geschrieben wurde.**
	if jetzt := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID); jetzt != "Von Hand" {
		t.Errorf("die Karte trägt %q", jetzt)
	}
}

// **Und eine Einschätzung ebenso** — sie kostet dasselbe Geld.
//
// **Ein eigener Lauf und keine zweite Zeile im ersten**: Vorschlag und
// Einschätzung gehen über **zwei** Wege durch `internal/app`, und einer kann
// die Freigabe halten, während der andere sie vergisst. Genau diese Gestalt
// hat #1539 gehabt.
func TestEineEinschaetzungBrauchtDieselbeFreigabe(t *testing.T) {
	t.Parallel()

	umgebung, wolke := eineEingerichteteWolke(t)
	_, karteID := eineAbnahmekarte(t, umgebung)

	agent := starteMCPProzess(t, umgebung)

	agent.scheitert("schreiben.ki.einschaetzung", map[string]any{
		"karteId": karteID, "art": string(domain.EinschaetzungRegel),
		"anbieter": "anthropic", "alias": wolkenmodell,
		"frage": "Wie wirkt sie beim Blocken?",
	}, "Freigabe eines Menschen")

	if auftraege := wolke.auftraege(); auftraege != 0 {
		t.Errorf("%d Aufträge gingen an den Anbieter", auftraege)
	}
}
