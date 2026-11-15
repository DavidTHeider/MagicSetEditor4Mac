// **Ein Widerruf von außen greift in der laufenden Sitzung** — Werkstück A14
// von Phase 5.3
// (#1453).
//
// # Was hier gemessen wird, und warum es etwas anderes ist als daneben
//
// **`tokenrotation_test.go` widerruft aus der Sitzung heraus** — der
// MCP-Prozess ruft `schreiben.token.widerrufen.alle` selbst. Das belegt, dass
// er seinen eigenen Widerruf sieht; es belegt **nicht**, dass er den eines
// anderen sieht. Ein Prozess, der die Berechtigung beim Start läse und danach
// nur seine eigenen Änderungen nachzöge, wäre in jenem Lauf grün.
//
// **Hier widerruft die Kommandozeile**, während der Prozess wartet. Das ist
// die Zusage aus dem Issue in ihrer schärfsten Form:
//
// > Keine einmalig gültige Tokenprüfung als dauerhafte Berechtigung.
//
// # Die Gegenprobe
//
// **Die Prüfung aus der Schleife herausgezogen** (`cmd/mse-mcp/server.go`:
// das Token einmal vor `Bediene` geprüft statt in `Ausstattung`): der zweite
// Ruf geht durch, und der Lauf wird rot. Sie steht hier, weil genau dieser
// Umbau naheliegt — er spart je Ruf eine Abfrage.
package flaechen_test

import (
	"strings"
	"testing"
)

// **Der zweite Ruf desselben Prozesses ist nach einem fremden Widerruf tot.**
func TestEinFremderWiderrufGreiftInDerLaufendenMCPSitzung(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	profilID := technischesProfil(t, umgebung)

	geheim := tokenAnDerZeile(t, umgebung, profilID, "schreiben")
	mitToken := append(append([]string{}, umgebung...),
		tokendateivariable+"="+tokendatei(t, geheim))

	agent := starteMCPProzess(t, mitToken)

	// **1. Die Sitzung steht.**
	agent.muss("lesen.sets", map[string]any{})

	// **2. Ein anderer Prozess zieht das Token.**
	//
	// **Über die Kommandozeile und nicht über die Sitzung**: der Widerruf
	// soll von außen kommen, sonst prüft der Lauf nur, ob ein Prozess seine
	// eigene Änderung kennt.
	mussKommandozeile(t, umgebung, "token-alle-widerrufen", profilID)

	// **3. Und der nächste Ruf gilt nicht mehr.**
	agent.scheitert("lesen.sets", map[string]any{}, "gilt nicht")
}

// **Ein Token bindet das Profil der Sitzung — und lässt den Zeiger stehen.**
//
// # Die Zusage aus dem Issue
//
// > lokale Profilauswahl nicht unbeabsichtigt globale GUI-Auswahl
//
// **Der erste Anlauf maß das Falsche** und wurde zu Recht rot: er ließ die
// Sitzung `schreiben.profil.wechseln` rufen und erwartete, dass der Zeiger
// stehen bleibt. Der wandert dabei — und **das ist Absicht**: ohne Token
// läuft `mse-mcp` im Altbetrieb, arbeitet im **gezeigten** Profil und
// benutzt dasselbe Werkzeug wie ein Mensch. Ein Wechsel, der dort nichts
// bewegte, wäre ein Werkzeug ohne Wirkung.
//
// **Getrennt wird über den Ausweis.** Ein Token gehört einem Profil
// (`app.Tokenwache.Pruefe` setzt es je Ruf), und die Wache läuft **nach** der
// Ausstattung — ihr Profil gewinnt also über das gezeigte. Damit arbeitet
// eine Agentensitzung dort, wo ihr Token hingehört, und der Mensch am Fenster
// merkt nichts davon.
//
// # Was der Lauf zeigt
//
// Zwei Profile, ein Token auf das zweite, der Zeiger auf dem ersten: der
// Agent sieht das zweite, die Kommandozeile weiter das erste.
func TestEinTokenBindetDasProfilUndLaesstDenZeigerStehen(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	vorher := profilkennung(t, mussKommandozeile(t, umgebung, "profil", "--json"))
	zweites := technischesProfil(t, umgebung)

	if zweites == vorher {
		t.Fatalf("beide Profile sind %q — der Lauf misst nichts", vorher)
	}

	geheim := tokenAnDerZeile(t, umgebung, zweites, "schreiben")
	mitToken := append(append([]string{}, umgebung...),
		tokendateivariable+"="+tokendatei(t, geheim))

	// **Je ein Set, damit die Antwort etwas unterscheidet.**
	//
	// **Ohne sie sagte der Lauf nur, welchen Namen die Sitzung nennt** — nicht,
	// ob er zu ihren Antworten passt. Und genau das ist die Zusage der
	// Werkzeugbeschreibung: „Jede andere Antwort dieses Servers … gilt für
	// DIESES Profil und kein anderes."
	mussKommandozeile(t, umgebung, "set-neu", "Beim Zeiger", "probe", "probe-schlicht")
	mussKommandozeile(t, umgebung, "--profil="+zweites,
		"set-neu", "Beim Token", "probe", "probe-schlicht")

	agent := starteMCPProzess(t, mitToken)

	// **Der Agent arbeitet im Profil seines Tokens.**
	if steht := profilkennung(t, agent.muss("lesen.profil",
		map[string]any{})); steht != zweites {
		t.Errorf("die Sitzung steht in %q statt in %q", steht, zweites)
	}

	// **Und seine Sets kommen aus demselben.**
	sets := agent.muss("lesen.sets", map[string]any{})
	if !strings.Contains(sets, "Beim Token") || strings.Contains(sets, "Beim Zeiger") {
		t.Errorf("die Sitzung sieht die falschen Sets: %s", sets)
	}

	// **Und der Zeiger daneben steht, wo er stand.**
	if steht := profilkennung(t, mussKommandozeile(t, umgebung,
		"profil", "--json")); steht != vorher {
		t.Errorf("der Zeiger wanderte von %q nach %q", vorher, steht)
	}
}

// **Und nach dem Entzug kommt auch aus dem Wissensbestand nichts mehr.**
//
// # Warum das eine eigene Zeile ist
//
// **Weil das Issue es eigens verlangt** (Ergänzung vom 2026-09-10): „echter
// fortlaufender MCP-Prozess mit Token-/Quellenwiderruf … keine privaten
// Snippets/Belege nach Entzug." Ein Widerruf, der das Lesen von Sets
// aufhält, aber die Recherche durchließe, gäbe genau das her, was diese
// Zeile ausschließt: Ausschnitte aus fremden Inhalten.
//
// **Gemessen an derselben Sitzung wie oben** — und nicht an einem zweiten
// Prozess: die Frage ist, ob der Riegel **vor jedem Werkzeug** steht und
// nicht nur vor dem ersten.
func TestNachDemWiderrufKommtAuchAusDerRechercheNichts(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	profilID := technischesProfil(t, umgebung)

	geheim := tokenAnDerZeile(t, umgebung, profilID, "lesen")
	mitToken := append(append([]string{}, umgebung...),
		tokendateivariable+"="+tokendatei(t, geheim))

	agent := starteMCPProzess(t, mitToken)

	// **Die Recherche antwortet** — ohne Index mit einem Grund, und das ist
	// eine Antwort. Ohne diesen Schritt sagte der nächste nichts.
	if antwort := agent.muss("lesen.recherche",
		map[string]any{"fall": "karten", "frage": "Blitz"}); !strings.Contains(
		antwort, "leere") {
		t.Fatalf("die Recherche nennt keinen Grund: %s", antwort)
	}

	mussKommandozeile(t, umgebung, "token-alle-widerrufen", profilID)

	agent.scheitert("lesen.recherche",
		map[string]any{"fall": "karten", "frage": "Blitz"}, "gilt nicht")
}
