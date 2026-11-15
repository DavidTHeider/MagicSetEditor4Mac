// **Die Null-Netz-Probe** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Der Satz, für den es diesen Lauf gibt
//
// > Anbieter fällt aus/all extern aus/Local-only mit konfigurierter Cloud:
// > Editor bleibt verwendbar, Null-Netz-Probe umfasst Modelle/Probe/Assets/
// > Retry.
//
// **Der Lauf in `kicloudstoff_test.go` deckt einen der vier** — den
// Modellkatalog. Hier stehen die übrigen, soweit sie ohne Fenster zu messen
// sind.
//
// # Was er hält und was daneben steht
//
//   - **Modelle** und **Probe** gehen über denselben Riegel und werden beide
//     gezählt.
//   - **Retry** ist die Frage, ob eine abgewiesene Anfrage **wiederholt**
//     wird. Sie ist am Zähler zu beantworten: null heißt null, auch über
//     mehrere Rufe.
//   - **Der Editor bleibt verwendbar** — lesen und schreiben am Bestand geht
//     weiter, während nach draußen nichts geht.
//   - **Assets stehen nicht hier**: ein Bildvorschlag lebt im Prozess und
//     gibt es nur am Fenster ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md)).
//     Ohne Fenster ist die Frage nicht zu stellen, und ein Lauf, der so täte,
//     prüfte eine Fläche, die es nicht gibt.
package flaechen_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Unter »nur lokal« geht über keinen der Wege etwas hinaus — und der
// Bestand bleibt benutzbar.**
func TestUnterNurLokalGehtUeberKeinenWegEtwasHinaus(t *testing.T) {
	t.Parallel()

	umgebung, wolke := eineEingerichteteWolke(t)
	setID, karteID := eineAbnahmekarte(t, umgebung)

	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.RichtlinieNurLokal, domain.NurLokalAn)

	agent := starteMCPProzess(t, umgebung)

	// **Der Katalog.**
	agent.muss("lesen.ki.modelle", map[string]any{})

	// **Die Messung** — sie kostet Geld und geht durch denselben Riegel.
	agent.scheitert("schreiben.ki.modell.pruefen", map[string]any{
		"anbieter": "anthropic", "alias": wolkenmodell,
	}, "nur lokal")

	// **Der Vorschlag.**
	agent.scheitert("schreiben.ki.vorschlag", map[string]any{
		"karteId": karteID, "anbieter": "anthropic", "alias": wolkenmodell,
		"auftrag": derAuftrag,
	}, "nur lokal")

	// **Und die Einschätzung.**
	agent.scheitert("schreiben.ki.einschaetzung", map[string]any{
		"karteId": karteID, "art": string(domain.EinschaetzungRegel),
		"anbieter": "anthropic", "alias": wolkenmodell, "frage": "Was gilt?",
	}, "nur lokal")

	// **Über vier Rufe null Anfragen** — das ist zugleich die Antwort auf
	// „Retry": eine abgewiesene Anfrage wird nicht wiederholt, weil sie
	// keine wird.
	if anfragen := wolke.anfragen(); anfragen != 0 {
		t.Errorf("%d Anfragen gingen hinaus", anfragen)
	}

	// **Und der Bestand bleibt benutzbar.**
	agent.muss("schreiben.feld", map[string]any{
		"setID": zahl(t, setID), "karteID": karteID,
		"feld": dasFeld, "wert": "Ohne Netz",
	})

	if jetzt := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID); jetzt != "Ohne Netz" {
		t.Errorf("die Karte trägt %q", jetzt)
	}
}

// **Und eine Einbettung an eine fremde Adresse wird aufgehalten, bevor sie
// eine Anfrage wird.**
//
// # Woran das zu erkennen ist
//
// **An der Meldung und nicht an einem Zähler.** Ein Zähler misst hier nichts:
// die Gegenstelle gäbe es gar nicht. Was die beiden Fälle unterscheidet, ist
// der **Grund** — wer erst hinausgeht und scheitert, meldet eine Verbindung;
// wer vorher aufgehalten wird, meldet die Regel.
//
// **Die Adresse ist bewusst unauflösbar** (`.invalid`, RFC 2606): geht der
// Riegel verloren, endet der Ruf an der Namensauflösung, und die Meldung
// spricht von etwas anderem. Genau daran fällt es auf.
func TestEineEinbettungNachDraussenWirdVorherAufgehalten(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)

	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.KIAdresse, "http://beispiel.invalid:1234")
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.KILaufzeit, domain.KILaufzeitAn)
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.RichtlinieNurLokal, domain.NurLokalAn)

	agent := starteMCPProzess(t, umgebung)

	text := agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "Blitz",
	})

	// **Sie antwortet mit einem Grund und nicht mit einer Verbindung.**
	if strings.Contains(text, "beispiel.invalid") {
		t.Errorf("der Ruf ging an die Adresse: %s", text)
	}

	if !strings.Contains(text, "leere") {
		t.Errorf("die Antwort nennt keinen Grund: %s", text)
	}
}

// **Der Fortsetzungszeiger steht nicht hier, und das ist gemessen.**
//
// # Was das Issue verlangt
//
// > echter fortlaufender MCP-Prozess mit Token-/Quellenwiderruf, **abgelaufenem
// > Cursor**, laufendem Rebuild und Offline-Modell
//
// **Der Zeiger wird ohne aktiven Index gar nicht gelesen.** `Recherchedienst`
// prüft `stand.Aktiv` **vor** `stelleAus` und gibt „ohne Index" zurück; ein
// Zeiger, der nicht dazu passt, fällt dort nicht auf. Das ist richtig — ohne
// Liste gibt es nichts fortzusetzen —, und es heißt, dass dieser Lauf die
// Zusage nicht erreichen kann.
//
// **Gehalten wird sie in `internal/domain/recherchezeiger_test.go`**: Ablauf
// nach fünfzehn Minuten, jede Änderung an Frage, Profil oder Indexstand macht
// ihn fremd, ein unlesbarer ist ein Fehler der Anfrage. Und im
// `Recherchedienst` gegen einen aktiven Index.
//
// **Dasselbe gilt für den laufenden Rebuild**: er braucht eine echte
// Einbettungslaufzeit und einen echten Korpus — beides steht hinter
// `netzlauf` und nicht in der Pipeline (§3).

// **Und ein fehlendes Einbettungsmodell ist ein eigener Grund.**
//
// # Warum das nicht dasselbe ist wie „nichts gefunden"
//
// **Weil es sich einrichten lässt.** Das Issue verlangt die Unterscheidung
// ausdrücklich, und A16 führt sechs Gründe: nur einer davon heißt, dass es
// die Antwort nicht gibt. Ein Agent, der `ohne-modell` als „gibt es nicht"
// läse, antwortete aus dem Gedächtnis statt zu sagen, was fehlt.
func TestOhneEinbettungsmodellNenntDieRechercheDenGrund(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)
	agent := starteMCPProzess(t, umgebung)

	text := agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "Blitz",
	})

	// **Der Grund steht da, und es ist nicht »nichts-gefunden«.**
	if strings.Contains(text, `"leere": "nichts-gefunden"`) {
		t.Errorf("der Grund ist »nichts gefunden«, obwohl nichts eingerichtet ist: %s", text)
	}

	if !strings.Contains(text, `"grund"`) {
		t.Errorf("die Antwort nennt keinen Grund im Klartext: %s", text)
	}
}
