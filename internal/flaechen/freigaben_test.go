// **Eine Freigabe wirkt — an jeder Fläche** — Werkstück 11 von Phase 4.12
// (#1351).
//
// # Zwei Aussagen, zwei Läufe
//
// **Die Verwaltung** — vergeben, nachlesen, zurückziehen — fährt über jede
// Fläche und muss dasselbe ergeben (§16, letzter Absatz).
//
// **Die Wirkung** — dass eine Freigabe am zweiten Faktor vorbei hereinlässt
// und ein Zurückziehen die Tür wieder schließt — fährt einmal, an der
// Kommandozeile. Sie braucht einen eingerichteten Faktor und damit einen
// eigenen Schlüsselbund; über MCP lässt sich einer gar nicht einrichten, und
// das steht so schon in `zweiterfaktor_test.go`.
//
// **Getrennt, weil ein rotes Ergebnis sonst nicht sagt, welche der beiden
// gebrochen ist** (§10).
//
// # Warum überall erst in das fremde Profil gewechselt wird
//
// **Weil nur sein Eigner Freigaben daran vergibt.** Wer arbeitet, ist der
// Eigner des gezeigten Profils; um für Nutzer 2 zu sprechen, muss man in
// seinem Profil stehen. Solange er keinen zweiten Faktor hat, kommt man dort
// hinein — und genau das ist der Weg, auf dem ein zweiter Mensch an diesem
// Rechner eingerichtet wird.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestFreigabenAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilAnDerZeile(t, umgebung)

	// **Der Nutzerwechsel und nicht der Profilwechsel** — Werkstück 17
	// (#1400).
	// Wer eine Freigabe **auf** ein Profil setzt, muss dessen Verwalter sein.
	// Bis zu diesem Werkstück wurde man das, indem man hineinwechselte; jetzt
	// wird man es, indem man der Mensch **wird**. Ohne diese Zeile antwortet
	// jede der drei Flächen „dafür braucht es mindestens die Rolle
	// verwalter" — und das ist die Abnahme dieses Werkstücks.
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)
	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "mitarbeiter")

	if aus := mussKommandozeile(t, umgebung, "freigaben", fremdes); !strings.Contains(aus, "mitarbeiter") {
		t.Errorf("die Freigabe fehlt in der Liste: %s", aus)
	}

	mussKommandozeile(t, umgebung, "freigabe-loeschen", fremdes, ich)

	if aus := mussKommandozeile(t, umgebung, "freigaben", fremdes); strings.Contains(aus, "mitarbeiter") {
		t.Errorf("die zurückgezogene Freigabe steht noch da: %s", aus)
	}

	// **Zurück als der, der man war.**
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", meines)
}

// **Derselbe Weg über MCP.**
func TestFreigabenUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	eigenes := mussMCP(t, umgebung, "lesen.profil", map[string]any{})
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilUeberMCP(t, umgebung)

	// **Der Nutzerwechsel und nicht der Profilwechsel** — Werkstück 17
	// (#1400).
	// Wer eine Freigabe **auf** ein Profil setzt, muss dessen Verwalter sein.
	// Bis zu diesem Werkstück wurde man das, indem man hineinwechselte; jetzt
	// wird man es, indem man der Mensch **wird**. Ohne diese Zeile antwortet
	// jede der drei Flächen „dafür braucht es mindestens die Rolle
	// verwalter" — und das ist die Abnahme dieses Werkstücks.
	mussMCP(t, umgebung, "schreiben.nutzer.wechseln", map[string]any{"nutzerID": fremder})
	mussMCP(t, umgebung, "schreiben.freigabe", map[string]any{
		"profilID": fremdes, "nutzerID": ich, "rolle": "mitarbeiter",
	})

	if aus := mussMCP(t, umgebung, "lesen.freigaben",
		map[string]any{"profilID": fremdes}); !strings.Contains(aus, "mitarbeiter") {
		t.Errorf("die Freigabe fehlt in der Liste: %s", aus)
	}

	mussMCP(t, umgebung, "schreiben.freigabe.loeschen",
		map[string]any{"profilID": fremdes, "nutzerID": ich})

	if aus := mussMCP(t, umgebung, "lesen.freigaben",
		map[string]any{"profilID": fremdes}); strings.Contains(aus, "mitarbeiter") {
		t.Errorf("die zurückgezogene Freigabe steht noch da: %s", aus)
	}

	mussMCP(t, umgebung, "schreiben.nutzer.wechseln", map[string]any{"nutzerID": ich})
	mussMCP(t, umgebung, "schreiben.profil.wechseln", map[string]any{"profilID": meines})
}

// **Und derselbe Weg über den Dienst.**
func TestFreigabenUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	eigenes := mussREST(t, an, http.MethodGet, "/profil", "")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilUeberDenDienst(t, an)

	// **Der Nutzerwechsel und nicht der Profilwechsel** — Werkstück 17
	// (#1400).
	// Wer eine Freigabe **auf** ein Profil setzt, muss dessen Verwalter sein.
	// Bis zu diesem Werkstück wurde man das, indem man hineinwechselte; jetzt
	// wird man es, indem man der Mensch **wird**. Ohne diese Zeile antwortet
	// jede der drei Flächen „dafür braucht es mindestens die Rolle
	// verwalter" — und das ist die Abnahme dieses Werkstücks.
	mussREST(t, an, http.MethodPut, "/nutzer", `{"nutzerID":"`+fremder+`"}`)
	mussREST(t, an, http.MethodPut,
		"/profile/"+fremdes+"/freigaben/"+ich, `{"rolle":"mitarbeiter"}`)

	if aus := mussREST(t, an, http.MethodGet,
		"/profile/"+fremdes+"/freigaben", ""); !strings.Contains(aus, "mitarbeiter") {
		t.Errorf("die Freigabe fehlt in der Liste: %s", aus)
	}

	mussREST(t, an, http.MethodDelete, "/profile/"+fremdes+"/freigaben/"+ich, "")

	if aus := mussREST(t, an, http.MethodGet,
		"/profile/"+fremdes+"/freigaben", ""); strings.Contains(aus, "mitarbeiter") {
		t.Errorf("die zurückgezogene Freigabe steht noch da: %s", aus)
	}

	mussREST(t, an, http.MethodPut, "/nutzer", `{"nutzerID":"`+ich+`"}`)
	mussREST(t, an, http.MethodPut, "/profil", `{"profilID":"`+meines+`"}`)
}

// **Und dass sie wirkt** — der ganze Bogen an einem Bestand.
//
// # Was hier die Abnahme ist
//
// **Nicht, dass sich eine Freigabe speichern lässt** — das steht oben. Hier
// steht das Verhalten, für das es Werkstück 11 gibt:
//
//  1. Nutzer 2 richtet einen zweiten Faktor ein; sein Profil ist verschlossen.
//  2. Ohne Code und ohne Freigabe: **abgewiesen**.
//  3. Mit Freigabe und **ohne** Code: es geht.
//  4. Zurückgezogen: wieder abgewiesen.
//
// **Schritt 4 gehört dazu**, weil eine Erlaubnis, die man nicht zurücknehmen
// kann, keine ist — ein Lauf, der bei 3 endet, ließe ein Zurückziehen
// durchgehen, das den Eintrag entfernt und die Tür offen lässt (§10).
//
// # Warum die Freigabe **vor** dem Faktor eingerichtet wird
//
// **Weil sie danach niemand mehr vergeben könnte.** Nur der Eigner verwaltet
// sein Profil, und um für ihn zu sprechen, steht man in seinem Profil — was
// nach dem Faktor eben den Code verlangt. Das ist keine Lücke, sondern die
// Reihenfolge, in der ein Mensch das auch täte.
func TestEineFreigabeLaesstAmFaktorVorbei(t *testing.T) {
	t.Parallel()

	umgebung := bestandMitBund(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	// **Solange Nutzer 2 kein Schloss hat, wird man ihn ohne Nachweis.**
	//
	// **Das ist die Zusage des zweiten Faktors und keine Lücke** (Vermerk
	// 0074): wer sich auf einem geteilten Rechner schützen will, richtet einen
	// ein — und die Zeilen darunter tun genau das.
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)
	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "mitarbeiter")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", meines)

	// **Und jetzt bekommt er eines.**
	eingerichtet := einrichtungAus(t,
		mussKommandozeile(t, umgebung, "faktor-neu", fremder, "totp", "--json"))
	mussKommandozeile(t, umgebung, "faktor-bestaetigen", fremder,
		bestaetigungskennwort(t, eingerichtet.Geheimnis))

	// **Die Freigabe führt am Code vorbei** — hinein und wieder heraus.
	mussKommandozeile(t, umgebung, "profil", fremdes)
	mussKommandozeile(t, umgebung, "profil", meines)

	// **Zurückgezogen wird sie vom Eigner — und dafür braucht es seinen
	// Code.**
	//
	// **Seit Werkstück 17 sind das zwei Dinge** (#1400): der Wechsel *in* sein
	// Profil macht mich nicht zu ihm, und eine Freigabe zurückzuziehen darf
	// nur, wer das Profil **verwaltet**. Ein `mitarbeiter`, der die Freigabe
	// entfernen könnte, die ihn hereinlässt, wäre kein Mitarbeiter.
	//
	// **Und genau hier zeigt der zweite Faktor, wozu es ihn gibt**: der Eigner
	// führt jetzt einen, und ohne dessen Code wird niemand zu ihm.
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder,
		einmalkennwort(t, eingerichtet.Geheimnis))
	mussKommandozeile(t, umgebung, "freigabe-loeschen", fremdes, ich)
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", meines)

	// **Und ohne sie ist die Tür zu.**
	_, fehler, code := kommandozeile(t, umgebung, "profil", fremdes)
	if code == 0 {
		t.Fatal("der Wechsel glückte ohne Freigabe und ohne Code")
	}

	if !strings.Contains(fehler, "zweiten Faktor") {
		t.Errorf("die Ablehnung lautet %q", fehler)
	}
}

// fremdesProfilAnDerZeile legt einen zweiten Nutzer an und gibt sein Profil.
func fremdesProfilAnDerZeile(t *testing.T, umgebung []string) (nutzerID, profilID string) {
	t.Helper()

	return profilEinesFremden(t,
		mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json"))
}

// fremdesProfilUeberMCP tut dasselbe über MCP.
func fremdesProfilUeberMCP(t *testing.T, umgebung []string) (nutzerID, profilID string) {
	t.Helper()

	return profilEinesFremden(t,
		mussMCP(t, umgebung, "schreiben.nutzer.neu", map[string]any{"name": "Zweiter"}))
}

// fremdesProfilUeberDenDienst tut dasselbe über den Dienst.
func fremdesProfilUeberDenDienst(t *testing.T, an dienst) (nutzerID, profilID string) {
	t.Helper()

	return profilEinesFremden(t,
		mussREST(t, an, http.MethodPost, "/nutzer", `{"name":"Zweiter"}`))
}

// profilEinesFremden liest Kennung **und** erstes Profil aus einer
// Nutzerantwort.
//
// **Beides seit Werkstück 17 von Phase 4.12**
// (#1400):
// wer eine Freigabe **auf** ein fremdes Profil setzen will, muss dessen
// Verwalter sein — und das wird man nicht mehr, indem man hineinwechselt,
// sondern indem man der Mensch **wird**. Dafür braucht der Lauf seine
// Nutzerkennung.
func profilEinesFremden(t *testing.T, roh string) (nutzerID, profilID string) {
	t.Helper()

	var gelesen struct {
		Nutzer struct {
			ID      string `json:"id"`
			Profile []struct {
				ID string `json:"id"`
			} `json:"profile"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Nutzerantwort lesen: %v — %s", err, roh)
	}

	if len(gelesen.Nutzer.Profile) == 0 {
		t.Fatalf("der zweite Nutzer hat kein Profil: %s", roh)
	}

	return gelesen.Nutzer.ID, gelesen.Nutzer.Profile[0].ID
}
