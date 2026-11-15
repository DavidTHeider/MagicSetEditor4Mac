package flaechen_test

// **Der Beginn, bevor eine Vorrichtung spielt** —
// #1851.
//
// # Warum es diese Hilfe gibt
//
// **Weil der Beginn vom Aufbau an läuft** — CR 103.8 (Fassung 2026-09-25):
// bis jeder Sitz seine Starthand behalten hat, nimmt eine Partie an keiner
// Fläche etwas anderes an. Die Läufe, die eine Handlung an jeder Fläche
// prüfen, bringen ihn deshalb zuerst zu Ende — **über die Kommandozeile**,
// wie sie die Partie auch anlegen: beides ist Vorbereitung und nicht das
// Geprüfte ([frischePartie]). Den Stand im Bestand umzuschreiben, prüfte die
// Vorrichtung statt des Wegs (§3).
//
// # Und warum die Objekte danach gesucht werden
//
// **Weil die Starthand die Bibliothek wirklich mischt** (CR 103.3): danach
// liegt `anna/p0-1/0` irgendwo, auf der Hand oder in der Bibliothek. Gesucht
// wird deshalb im Stand, nach Sitz und Zone ([objektIn]).

import (
	"slices"
	"strconv"
	"testing"
)

// spielbereitePartie legt eine Partie aus den beiden Decks an und bringt
// ihren Beginn zu Ende.
func spielbereitePartie(
	t *testing.T, umgebung []string, erster, zweiter string,
) partiebild {
	t.Helper()

	return beginnAbschliessen(t, umgebung,
		frischePartie(t, umgebung, erster, zweiter))
}

// beginnAbschliessen zieht für jeden Sitz die Starthand und behält sie —
// CR 103.5 — und gibt den Stand danach.
//
// **Jeder Sitz, den die Partie nennt**, und nicht zwei feste: auch der Tisch
// zu dritt beginnt erst, wenn der letzte behalten hat. **Gelesen wird als
// Anna** — sie ist an jedem Tisch dieser Läufe der Wirt.
func beginnAbschliessen(t *testing.T, umgebung []string, partieID string) partiebild {
	t.Helper()

	vorher := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	for _, sitz := range vorher.Sitze {
		for _, art := range []string{"spieler.starthand", "spieler.behalten"} {
			mussKommandozeile(t, umgebung, "partie-befehl", partieID,
				sitz.Kennung, art+"/"+sitz.Kennung, art, `{}`)
		}
	}

	nachher := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	// **Dass der Beginn wirklich vorbei ist** — sonst prüfte jeder Lauf
	// danach die Sperre statt seiner Handlung.
	if !slices.Contains(nachher.Erlaubt, "zug.passen") ||
		slices.Contains(nachher.Erlaubt, "spieler.starthand") {
		t.Fatalf("nach dem Beginn ist Anna erlaubt: %v", nachher.Erlaubt)
	}

	return nachher
}

// objektIn gibt die Kennung eines Objekts, das `wer` in dieser Zone
// beherrscht.
func objektIn(t *testing.T, bild partiebild, wer, zone string) string {
	t.Helper()

	for _, objekt := range bild.Objekte {
		if objekt.Beherrscher == wer && objekt.Zone == zone {
			return objekt.Kennung
		}
	}

	t.Fatalf("%s beherrscht in %q kein Objekt: %+v", wer, zone, bild.Objekte)

	return ""
}

// naechsteFassung ist die Fassung, die ein Befehl nach dieser hinterlässt.
func naechsteFassung(t *testing.T, fassung string) string {
	t.Helper()

	zahl, fehler := strconv.Atoi(fassung)
	if fehler != nil {
		t.Fatalf("die Fassung %q ist keine Zahl", fassung)
	}

	return strconv.Itoa(zahl + 1)
}
