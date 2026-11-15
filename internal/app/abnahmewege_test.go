// **Sechs Wege, die jemand wirklich geht** —
// #680.
//
// # Was diese Datei prüft, das keine andere prüft
//
// Jede Prüfebene dieses Projekts sieht ihren Ausschnitt: `domain` seine
// Entitäten, `service` seine Anwendungsfälle, `frontend` seine Bausteine gegen
// `test/api.ts`. **Keine geht einen Weg von Anfang bis Ende.**
//
// §3 nennt die Lücke und stellt eine Anweisung an einen Menschen daneben —
// „vor jeder Freigabe die gebaute Anwendung starten". Was das findet, ist
// belegt: eine einzige Sitzung an der laufenden Anwendung brachte
// #666,
// #667 und
// #668 —
// drei Befunde, keinen davon in irgendeinem Test.
//
// # Was sie **nicht** prüft
//
// **Sie geht durch die Bindungen, nicht durch das Fenster.** Ob ein Knopf zu
// sehen ist, ob die Vorschau haftet, ob ein Kontrast trägt — davon weiß sie
// nichts. Der Weg ist gangbar; ob er sich gut anfühlt, sagt nur die
// Sichtprüfung (§3).
//
// **Und sie sagt nicht, ob das Ergebnis stimmt.** „Das Bild ist entstanden"
// ist eine andere Zusage als „das Bild sieht aus wie bei MSE" — dafür gibt es
// den Kanonlauf und den Abstandslauf.
//
// # Warum sie in der Pipeline läuft
//
// Weil sie mit der **Saat** auskommt. Die Messläufe brauchen bezogene Pakete
// und stehen deshalb hinter `fremdpakete`; ein Weg, der nur mit fremden Daten
// gangbar ist, wäre auch keine Zusage über die Anwendung.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// schritt schreibt einen Wegschritt in den Lauf.
//
// **Der Lauf liest sich als Weg.** Eine Zusicherung, die scheitert, nennt
// sonst nur ihre Zeile; hier steht darüber, an welcher Stelle des Weges jemand
// stünde.
func schritt(t *testing.T, was string) {
	t.Helper()
	t.Logf("  → %s", was)
}

// **Weg 1: von nichts zur ersten Karte.**
//
// Der Weg eines Menschen, der die Anwendung zum ersten Mal öffnet: kein Set,
// keine bezogenen Pakete, nur das, was mitkommt.
func TestWegVonNichtsZurErstenKarte(t *testing.T) {
	anwendung := mitSaat(t)

	schritt(t, "die Anwendung startet ohne ein einziges Set")

	sets := anwendung.Sets(t.Context())
	if sets.Fehler != nil {
		t.Fatalf("Sets: %+v", sets.Fehler)
	}

	if len(sets.Sets) != 0 {
		t.Fatalf("der Bestand ist nicht leer: %d Sets", len(sets.Sets))
	}

	schritt(t, "ein Set wird geöffnet")

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	schritt(t, "eine neue Karte wird angelegt")

	angelegt := anwendung.KarteAnlegen(t.Context(), geoeffnet.SetID)
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	schritt(t, "sie bekommt einen Namen")

	gesetzt := anwendung.FeldSetzen(
		probekontext(t), geoeffnet.SetID, angelegt.KarteID, "name", "Mein Feuerball")
	if gesetzt.Fehler != nil {
		t.Fatalf("Feld setzen: %+v", gesetzt.Fehler)
	}

	schritt(t, "und sie wird gezeichnet")

	vorschau := anwendung.Vorschau(t.Context(), angelegt.KarteID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	if vorschau.Vorschau == nil || vorschau.Vorschau.Bild == "" {
		t.Fatal("die neue Karte hat kein Bild")
	}

	// **Der Name steht auf der Karte** — nicht nur im Bestand. Ohne diese
	// Zusicherung wäre „ein Bild ist entstanden" mit einer leeren Karte
	// zufrieden.
	if len(vorschau.Vorschau.Flaechen) == 0 {
		t.Error("die gezeichnete Karte trägt kein einziges Feld")
	}
}

// **Weg 2: ein fremdes Set öffnen, ändern, sichern, wieder aufmachen.**
//
// Die Zusage aus [0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)
// von der Seite dessen, der sie braucht: was ich ändere, steht danach in der
// Datei, und was ich nicht ändere, steht unverändert daneben.
func TestWegEinSetOeffnenAendernUndSichern(t *testing.T) {
	anwendung := mitSaat(t)

	ort := saatset(t)

	schritt(t, "ein Set wird geöffnet")

	geoeffnet := anwendung.SetOeffnen(t.Context(), ort)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if len(karten.Karten) == 0 {
		t.Fatal("das Set trägt keine Karte")
	}

	schritt(t, "ein Feld wird geändert")

	if antwort := anwendung.FeldSetzen(
		probekontext(t), geoeffnet.SetID, karten.Karten[0].ID, "name", "Eisball",
	); antwort.Fehler != nil {
		t.Fatalf("Feld setzen: %+v", antwort.Fehler)
	}

	// **Und ein mehrzeiliger dazu** —
	// #704.
	// Ein Wert mit Umbrüchen steht in der Datei **unter** seinem Schlüssel;
	// wer ihn dahinter schriebe, beendete mit der zweiten Zeile den
	// `card:`-Block, und alles danach läse sich als oberste Ebene.
	if antwort := anwendung.FeldSetzen(
		probekontext(t), geoeffnet.SetID, karten.Karten[0].ID, "notes",
		"Neue Zeile\n\nNeuer Rest",
	); antwort.Fehler != nil {
		t.Fatalf("Feld setzen: %+v", antwort.Fehler)
	}

	schritt(t, "das Set wird gesichert")

	if antwort := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, ort); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	schritt(t, "und die Datei trägt die Änderung")

	roh, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- eigenes Temp
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(roh), "Eisball") {
		t.Error("die Änderung steht nicht in der Datei")
	}

	// **Und was niemand angefasst hat, steht unverändert da** — die Zusage,
	// die einen Rundlauf von einem Neuschreiben unterscheidet.
	if !strings.Contains(string(roh), "Füge drei Schaden zu.") {
		t.Error("ein unberührtes Feld ist beim Sichern verlorengegangen")
	}

	schritt(t, "und beim Wiederöffnen ist jeder Feldwert wieder da")

	// **Jeder, nicht nur der geänderte** —
	// #704.
	// Ein mehrzeiliger Wert wurde einzeilig geschrieben; seine zweite Zeile
	// beendete den `card:`-Block, und alles danach las sich als oberste Ebene.
	// Über ein echtes Set kamen von 18 Feldwerten **zwei** zurück.
	zweite := mitSaat(t)

	wieder := zweite.SetOeffnen(t.Context(), ort)
	if wieder.Fehler != nil {
		t.Fatalf("wieder öffnen: %+v", wieder.Fehler)
	}

	nochmal := zweite.Karten(t.Context(), wieder.SetID)
	if nochmal.Fehler != nil {
		t.Fatalf("Karten: %+v", nochmal.Fehler)
	}

	if len(nochmal.Karten) != 1 {
		t.Fatalf("%d Karten", len(nochmal.Karten))
	}

	if len(nochmal.Karten[0].Werte) != len(karten.Karten[0].Werte) {
		t.Errorf("%d Feldwerte vorher, %d nachher",
			len(karten.Karten[0].Werte), len(nochmal.Karten[0].Werte))
	}

	// **Und der mehrzeilige Wert kommt ganz zurück, genau einmal.**
	//
	// Zwei Fallen, und die Zahl der Felder sieht keine von beiden: geschrieben
	// als `notes: Neue Zeile\n\nNeuer Rest` steht die zweite Zeile ohne
	// Einrückung — die Karte behält vier Werte, und `notes` heißt danach nur
	// noch „Neue Zeile", während „Neuer Rest" als oberste Ebene danebensteht.
	// Bleibt der alte Block stehen, verdoppelt sich der Wert stattdessen.
	if zahl := strings.Count(string(roh), "Neuer Rest"); zahl != 1 {
		t.Errorf("»Neuer Rest« steht %d-mal in der Datei", zahl)
	}

	if notiz := wertAus(nochmal.Karten[0], "notes"); notiz != "Neue Zeile\n\nNeuer Rest" {
		t.Errorf("notes lautet %q", notiz)
	}
}

// wertAus liest einen Feldwert aus einer Karte.
func wertAus(karte KarteDTO, feld string) string {
	for _, wert := range karte.Werte {
		if wert.Feld == feld {
			return wert.Wert
		}
	}

	return ""
}

// **Weg 3: ein Missgeschick zurücknehmen.**
//
// Der Weg, der über den Wert der Anwendung entscheidet: wer etwas löscht, muss
// es zurückholen können. Ohne ⌘Z ist Löschen Datenverlust mit einem Klick
// (#562).
func TestWegEinMissgeschickZurueckNehmen(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	vorher := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if len(vorher.Karten) == 0 {
		t.Fatal("das Set trägt keine Karte")
	}

	name := kartennameAus(vorher.Karten[0])

	schritt(t, "eine Karte wird gelöscht")

	if antwort := anwendung.KarteLoeschen(
		probekontext(t), geoeffnet.SetID, vorher.Karten[0].ID, false,
	); antwort.Fehler != nil {
		t.Fatalf("löschen: %+v", antwort.Fehler)
	}

	if nachher := anwendung.Karten(t.Context(), geoeffnet.SetID); len(nachher.Karten) != len(vorher.Karten)-1 {
		t.Fatalf("nach dem Löschen sind es %d Karten, erwartet %d",
			len(nachher.Karten), len(vorher.Karten)-1)
	}

	schritt(t, "⌘Z holt sie zurück")

	if antwort := anwendung.Zuruecknehmen(t.Context(), geoeffnet.SetID, ""); antwort.Fehler != nil {
		t.Fatalf("zurücknehmen: %+v", antwort.Fehler)
	}

	zurueck := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if len(zurueck.Karten) != len(vorher.Karten) {
		t.Fatalf("nach ⌘Z sind es %d Karten, erwartet %d",
			len(zurueck.Karten), len(vorher.Karten))
	}

	// **Mit ihren Werten und nicht als leere Hülle.** Eine Karte, die
	// zurückkommt und nichts trägt, ist kein Zurücknehmen.
	if kartennameAus(zurueck.Karten[0]) != name {
		t.Errorf("die zurückgeholte Karte heißt %q, erwartet %q",
			kartennameAus(zurueck.Karten[0]), name)
	}
}

// kartennameAus liest den Namen einer Karte aus ihren Werten.
//
// **Eigene Fassung und nicht die aus `mse2kanon_intern_test.go`**: jene steht
// hinter `fremdpakete` und wäre hier nicht da. Zwei Zeilen doppelt sind
// billiger als ein Bautag, der zwei Dinge bedeutet.
func kartennameAus(karte KarteDTO) string {
	for _, wert := range karte.Werte {
		if wert.Feld == "name" {
			return wert.Wert
		}
	}

	return ""
}
