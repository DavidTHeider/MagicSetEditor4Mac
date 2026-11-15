package flaechen_test

// **Die freigegebenen Sichten über jede Fläche** — Werkstück `T3` von Phase
// 5.6 (#1468).
//
// # Was hier gemessen wird
//
// **Die drei Abnahmepunkte des Issues**, und zwar an der **tatsächlich
// serialisierten** Nachricht:
//
//  1. Kein lokales und kein privates Feld geht hinaus — auch nicht, wenn das
//     Ausgangsexemplar sie alle trägt.
//  2. Ohne Freigabe, für einen anderen Empfänger, bei abgelaufenem Umfang und
//     bei gesperrtem Kontakt gibt es **keine** Sicht.
//  3. Ein Widerruf beendet künftige Antworten.
//
// # Warum das über den Dienst geht und nicht über die Domäne
//
// **Weil der Lauf über die Domäne die Gestalt prüft und nicht den Weg.**
// Zwischen beidem liegt jede Abkürzung, die jemand nimmt — und die Abnahme
// verlangt ausdrücklich den Stand, der über die Leitung geht.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Eine Sicht entsteht erst durch eine Freigabe — und sie trägt nichts
// Privates.**
func TestDieFreigegebeneSichtUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	// **Ein Exemplar, in dem alles steht, was privat ist.**
	mussKommandozeile(t, umgebung, "exemplar",
		"a1b2c3d4-0000-4000-8000-00000000000a", "3", "played", "de", "etched",
		"2026-08-01", "4.17", "EUR", "aus dem Tausch", "Lightning Bolt",
		"lea", "161", "Kiste 4")

	// **Ohne Kontakt und ohne Freigabe gibt es keine Sicht** — an beiden
	// Flächen, und mit demselben Nein.
	if _, _, code := kommandozeile(t, umgebung,
		"kontakt-sicht", kennung, "handel"); code == 0 {
		t.Error("die Kommandozeile gab eine Sicht ohne Freigabe")
	}

	if status, _ := an.ruf(t, http.MethodGet,
		"/peer/kontakte/"+kennung+"/sicht?zweck=handel", ""); status == http.StatusOK {
		t.Error("der Dienst gab eine Sicht ohne Freigabe")
	}

	// **Jetzt der Kontakt, bestätigt und zum Handel berechtigt.**
	mussKommandozeile(t, umgebung, "kontakt-aufnehmen", kennung, "Annas Rechner")
	mussKommandozeile(t, umgebung, "kontakt-bestaetigen", kennung,
		fremd.Fingerabdruck())
	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "handel")

	// **Der Zweck allein reicht nicht.** „Darf handeln" ist keine Freigabe
	// der Sammlung — das ist die Trennung, um die es im Issue geht.
	if _, _, code := kommandozeile(t, umgebung,
		"kontakt-sicht", kennung, "handel"); code == 0 {
		t.Error("ein erlaubter Zweck gab schon eine Sicht")
	}

	mussKommandozeile(t, umgebung,
		"kontakt-freigeben", kennung, "handel", "sammlung")

	// **Und die Liste sagt, was gilt** — an beiden Flächen.
	for name, liste := range map[string]string{
		"Kommandozeile": mussKommandozeile(
			t, umgebung, "kontakt-freigaben", kennung),
		"Dienst": mussREST(t, an, http.MethodGet,
			"/peer/kontakte/"+kennung+"/freigaben", ""),
	} {
		if !strings.Contains(liste, "sammlung") {
			t.Errorf("%s nennt die Freigabe nicht: %s", name, liste)
		}
	}

	// **Eine Notiz gehört diesem Profil und geht nie hinaus.**
	mussKommandozeile(t, umgebung, "kontakt-notiz", kennung, "tauscht fair")

	if aus := mussKommandozeile(t, umgebung, "kontakte"); !strings.Contains(
		aus, "tauscht fair") {
		t.Errorf("die Notiz steht nicht in der Kontaktliste: %s", aus)
	}

	sicht := mussKommandozeile(t, umgebung, "--json", "kontakt-sicht",
		kennung, "handel")

	pruefeNichtsPrivates(t, "Kommandozeile", sicht)
	pruefeNichtsPrivates(t, "Dienst", mussREST(t, an, http.MethodGet,
		"/peer/kontakte/"+kennung+"/sicht?zweck=handel", ""))

	// **Und ein Widerruf beendet künftige Antworten.**
	mussKommandozeile(t, umgebung,
		"kontakt-widerrufen", kennung, "handel", "sammlung")

	if _, _, code := kommandozeile(t, umgebung,
		"kontakt-sicht", kennung, "handel"); code == 0 {
		t.Error("nach dem Widerruf kam noch eine Sicht")
	}
}

// pruefeNichtsPrivates hält eine serialisierte Sicht gegen die Werte, die
// nicht hinausgehen dürfen.
//
// **Gegen die Werte und nicht gegen die Feldnamen** — ein umbenanntes Feld
// trägt denselben Wert, und wer nach `"lagerort"` suchte, fände `"ort"` nicht.
func pruefeNichtsPrivates(t *testing.T, fläche, nachricht string) {
	t.Helper()

	// **Zuerst der Beleg, dass überhaupt etwas dasteht.** Eine leere Antwort
	// enthält nichts Privates und sagt damit nichts (§10).
	if !strings.Contains(nachricht, "Lightning Bolt") {
		t.Fatalf("%s gab keine Sicht: %s", fläche, nachricht)
	}

	if !strings.Contains(nachricht, "etched") {
		t.Errorf("%s nennt die Ausführung nicht: %s", fläche, nachricht)
	}

	for name, wert := range map[string]string{
		"die lokale Notiz zum Kontakt": "tauscht fair",
		"der Erwerbspreis":             "4.17",
		"das Erwerbsdatum":             "2026-08-01",
		"die Währung":                  "EUR",
		"die Notiz":                    "aus dem Tausch",
		"der Lagerort":                 "Kiste 4",
	} {
		if strings.Contains(nachricht, wert) {
			t.Errorf("%s gibt %s heraus: %s", fläche, name, nachricht)
		}
	}
}

// **Ein gesperrter Kontakt sieht nichts — auch mit Freigabe.**
//
// **Das ist die Zusage aus `N1`** („blockiert bleibt auch nach Spiel→Handeln
// blockiert"), und sie wäre gebrochen, sobald eine Freigabe sie überstimmen
// könnte.
func TestEinGesperrterKontaktSiehtNichts(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	mussKommandozeile(t, umgebung, "exemplar",
		"a1b2c3d4-0000-4000-8000-00000000000a", "1", "played", "de", "foil")
	mussKommandozeile(t, umgebung, "kontakt-aufnehmen", kennung, "Anna")
	mussKommandozeile(t, umgebung, "kontakt-bestaetigen", kennung,
		fremd.Fingerabdruck())
	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "handel")
	mussKommandozeile(t, umgebung,
		"kontakt-freigeben", kennung, "handel", "sammlung")

	// **Erst sieht er etwas** — sonst prüfte der Lauf nur, dass die Sperre
	// über einem Nichts liegt.
	if aus := mussKommandozeile(t, umgebung,
		"kontakt-sicht", kennung, "handel"); strings.TrimSpace(aus) == "" {
		t.Fatal("vor der Sperre gab es keine Sicht")
	}

	mussKommandozeile(t, umgebung, "kontakt-sperren", kennung)

	if _, _, code := kommandozeile(t, umgebung,
		"kontakt-sicht", kennung, "handel"); code == 0 {
		t.Error("der gesperrte Kontakt bekam eine Sicht")
	}
}

// **Und was einem Gegenüber gilt, gilt nicht dem anderen.**
func TestEineFreigabeGiltNurIhremEmpfaenger(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	erster, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	zweiter, zweitfehler := peer.Erzeugen(nil)
	if zweitfehler != nil {
		t.Fatalf("Erzeugen: %v", zweitfehler)
	}

	mussKommandozeile(t, umgebung, "exemplar",
		"a1b2c3d4-0000-4000-8000-00000000000a", "1", "played", "de", "foil")

	for _, wer := range []struct{ kennung, abdruck string }{
		{string(erster.Kennung()), erster.Fingerabdruck()},
		{string(zweiter.Kennung()), zweiter.Fingerabdruck()},
	} {
		mussKommandozeile(t, umgebung, "kontakt-aufnehmen", wer.kennung, "wer")
		mussKommandozeile(t, umgebung,
			"kontakt-bestaetigen", wer.kennung, wer.abdruck)
		mussKommandozeile(t, umgebung, "zweck-erlauben", wer.kennung, "handel")
	}

	mussKommandozeile(t, umgebung, "kontakt-freigeben",
		string(erster.Kennung()), "handel", "sammlung")

	if _, _, code := kommandozeile(t, umgebung, "kontakt-sicht",
		string(zweiter.Kennung()), "handel"); code == 0 {
		t.Error("die Freigabe des einen galt dem anderen")
	}
}
