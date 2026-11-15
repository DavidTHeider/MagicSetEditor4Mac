package flaechen_test

// **Auch eine neue Karte und eine Massenänderung speichern das Gerechnete** —
// #2516 (Rest aus #2396). MSE rechnet beim Anlegen jeden Wert der neuen Karte
// (`AddCardAction` in `SetScriptManager::onAction`,
// `src/script/script_manager.cpp`, Zeilen 168–178, Fassung f9d9356 vom
// 2021-07-06); die Massenänderung gibt es in MSE2 nicht, sie rechnet hier je
// Karte wie das einzelne Feld. **Dieselbe Antwort über Kommandozeile und
// Dienst** (§16).

import (
	"net/http"
	"testing"
)

func TestNeueKarteUndMassenaenderungSpeichernUeberJedeFlaecheDasGerechnete(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	mitWertskripten(t, umgebung)
	an := starteDienst(t, umgebung)
	setID := neuesSet(t, umgebung, "Nachzug beim Anlegen")

	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")
	if status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/karten", ""); status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("über den Dienst anlegen: %d (%s)", status, rumpf)
	}

	karten := kartenkennungenAus(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet zwei", len(karten))
	}

	// **Die neue Karte trägt, was ihre Skripte rechnen** — `zeile` aus dem
	// leeren Wert, `echo` aus ihr.
	for flaeche, karteID := range map[string]string{"Kommandozeile": karten[0], "Dienst": karten[1]} {
		if werte := gespeichert(t, umgebung, karteID); werte["zeile"] != "[]" || werte["echo"] != "[]!" {
			t.Errorf("über %s angelegt: zeile=%q, echo=%q — erwartet [] und []!", flaeche, werte["zeile"], werte["echo"])
		}
	}

	mussKommandozeile(t, umgebung, "felder-setzen", setID, "zeile", "Glut", karten[0])
	if status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/felder",
		`{"karteIDs":[`+karten[1]+`],"feld":"zeile","wert":"Asche"}`); status != http.StatusOK {
		t.Fatalf("Massenänderung über den Dienst: %d (%s)", status, rumpf)
	}

	for karteID, wert := range map[string]string{karten[0]: "Glut", karten[1]: "Asche"} {
		if werte := gespeichert(t, umgebung, karteID); werte["zeile"] != "["+wert+"]" || werte["echo"] != "["+wert+"]!" {
			t.Errorf("nach der Massenänderung zeile=%q, echo=%q — erwartet [%s] und [%s]!", werte["zeile"], werte["echo"], wert, wert)
		}
	}
}
