// **Wer arbeitet, ändert sich nicht durch einen Profilwechsel** — Werkstück 17
// von Phase 4.12 (#1400),
// nach [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
//
// # Warum das an den **gebauten** Programmen geprüft wird
//
// **Weil die Zusage an einem Zeiger im Bestand hängt und nicht an einem Feld
// im Prozess.** Ein Lauf gegen ein Doppel bewiese die Rechnung; hier läuft
// jeder Ruf durch einen eigenen Prozess, der den Bestand frisch aufmacht —
// und genau das ist der Fall, in dem der alte Weg die Identität verlor.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// **Wer in ein fremdes Profil wechselt, bleibt, wer er ist — an allen vier
// Flächen.**
func TestWerInEinFremdesProfilWechseltBleibtWerErIst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	_, fremdes := profilEinesFremden(t, angelegt)

	// **Hinein — das geht, denn der Zweite führt keinen Faktor.**
	mussKommandozeile(t, umgebung, "profil", fremdes)

	// **Und danach arbeitet immer noch der Erste.**
	//
	// **An allen drei kopflosen Flächen gefragt**, weil jede den Bestand
	// eigens aufmacht: eine Antwort, die nur ein Prozess gibt, wäre keine
	// Zusage über den Bestand.
	werArbeitetUeberall(t, umgebung, ich)

	mussKommandozeile(t, umgebung, "profil", "1")
}

// werArbeitetUeberall fragt „wer arbeitet" an Kommandozeile, MCP und Dienst.
func werArbeitetUeberall(t *testing.T, umgebung []string, erwartet string) {
	t.Helper()

	an := starteDienst(t, umgebung)

	// **Gelesen und nicht in der Zeichenkette gesucht.** Alle drei Flächen
	// setzen ihr JSON verschieden — mit und ohne Einrückung —, und ein
	// Textvergleich prüfte die Formatierung statt der Antwort.
	for name, roh := range map[string]string{
		"die Kommandozeile": mussKommandozeile(t, umgebung, "nutzer-wechseln", "--json"),
		"MCP":               mussMCP(t, umgebung, "lesen.werarbeitet", map[string]any{}),
		"der Dienst":        mussREST(t, an, http.MethodGet, "/werarbeitet", ""),
	} {
		var gelesen struct {
			NutzerID string `json:"nutzerID"`
		}

		if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
			t.Fatalf("%s: die Antwort lesen: %v — %s", name, err, roh)
		}

		if gelesen.NutzerID != erwartet {
			t.Errorf("%s sagt %q, erwartet war %q", name, gelesen.NutzerID, erwartet)
		}
	}
}

// **Und ein anderer wird man ausdrücklich — der Zeiger zieht mit.**
func TestEinNutzerwechselZiehtDenProfilzeigerMit(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	werArbeitetUeberall(t, umgebung, fremder)

	// **Und der Profilzeiger steht im Profil des neuen Menschen** — sonst
	// stünde er in einem, für das er keine Freigabe hat.
	if steht := profilkennung(t,
		mussKommandozeile(t, umgebung, "profil", "--json")); steht != fremdes {
		t.Errorf("der Zeiger steht auf %q, erwartet war das Profil des neuen "+
			"Menschen (%q)", steht, fremdes)
	}
}
