// **Drei weitere Wege** — umgestalten, hinausgeben, übernehmen.
//
// Eigene Datei, weil `abnahmewege_test.go` mit ihnen über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort die Wege, auf denen eine **Karte**
// entsteht und sich ändert, hier die, auf denen ein **Set** sein Aussehen
// wechselt oder die Anwendung verlässt.
//
// Was diese Läufe prüfen und was nicht, steht im Kopf von
// `abnahmewege_test.go`.
package app

import (
	"os"
	"path/filepath"
	"testing"
)

// **Weg 4: ein Set umgestalten.**
//
// Der Weg, für den es die Anwendung gibt: einen anderen Stil wählen, eine
// Option umstellen — und sehen, dass es wirkt.
func TestWegEinSetUmgestalten(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	schritt(t, "welche Stile gibt es?")

	stile := anwendung.Stile(t.Context(), geoeffnet.SetID)
	if stile.Fehler != nil {
		t.Fatalf("Stile: %+v", stile.Fehler)
	}

	if len(stile.Stile) < 2 {
		t.Fatalf("die Saat bietet %d Stile, für einen Wechsel braucht es zwei",
			len(stile.Stile))
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if len(karten.Karten) == 0 {
		t.Fatal("das Set trägt keine Karte")
	}

	vorher := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorher.Fehler != nil || vorher.Vorschau == nil {
		t.Fatalf("Vorschau: %+v", vorher.Fehler)
	}

	schritt(t, "der Stil wird gewechselt")

	anderer := andererStil(stile.Stile)

	if antwort := anwendung.StilWechseln(
		probekontext(t), geoeffnet.SetID, anderer,
	); antwort.Fehler != nil {
		t.Fatalf("Stil wechseln: %+v", antwort.Fehler)
	}

	schritt(t, "und die Karte sieht anders aus")

	nachher := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if nachher.Fehler != nil || nachher.Vorschau == nil {
		t.Fatalf("Vorschau: %+v", nachher.Fehler)
	}

	// **Über das Bild und nicht über den Namen des Stils.** Dass der Bestand
	// einen anderen Namen führt, sagt nichts darüber, ob jemand es sieht.
	if nachher.Vorschau.Bild == vorher.Vorschau.Bild {
		t.Error("nach dem Stilwechsel kommt dasselbe Bild heraus")
	}
}

// andererStil gibt einen Stil, der nicht der gewählte ist.
func andererStil(stile []StilDTO) string {
	for _, stil := range stile {
		if !stil.Gewaehlt {
			return stil.Paket
		}
	}

	return ""
}

// **Weg 5: ein Set hinausgeben.**
//
// Drei Ausgänge in einem Lauf — Kartenbild, Druckbogen, Exportvorlage. Sie
// stehen zusammen, weil sie **eine** Frage beantworten: kommt aus dem Set
// etwas heraus, das jemand weiterverwenden kann
// (#677)?
func TestWegEinSetHinausgeben(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if len(karten.Karten) == 0 {
		t.Fatal("das Set trägt keine Karte")
	}

	ziel := t.TempDir()

	schritt(t, "ein Kartenbild wird geschrieben")

	bild := anwendung.BildSchreiben(
		probekontext(t), karten.Karten[0].ID, ziel, "karte.png", "png", 375,
		"ueberschreiben", "")
	if bild.Fehler != nil {
		t.Fatalf("Bild schreiben: %+v", bild.Fehler)
	}

	pruefeDaLiegtEtwas(t, filepath.Join(ziel, "karte.png"))

	schritt(t, "ein Druckbogen entsteht")

	bogenpfad := filepath.Join(ziel, "bogen.pdf")

	bogen := anwendung.BogenLegen(t.Context(), geoeffnet.SetID, bogenpfad, "a4", 0, 150, true, nil)
	if bogen.Fehler != nil {
		t.Fatalf("Bogen legen: %+v", bogen.Fehler)
	}

	pruefeDaLiegtEtwas(t, bogenpfad)

	// **Ein Bogen ist eine Ausgabe** — gefunden am nativen Fenster
	// (Hüllenprobe `fmt03b`, FMT-03): die Startseite sagte nach einem
	// gelegten Bogen „Noch nichts ausgegeben", weil sich nur MTG JSON und
	// das Speichern vermerkten.
	schritt(t, "der Bogen steht unter »Zuletzt ausgegeben«")

	if ausgegeben := anwendung.Weiterarbeiten(t.Context()).Ausgegeben; len(ausgegeben) != 1 ||
		ausgegeben[0].Objekt.Kennung != geoeffnet.SetID {
		t.Errorf("zuletzt ausgegeben: %+v", ausgegeben)
	}

	schritt(t, "und eine Exportvorlage läuft")

	vorlagen := anwendung.Exportvorlagen(t.Context(), geoeffnet.SetID)
	if vorlagen.Fehler != nil {
		t.Fatalf("Exportvorlagen: %+v", vorlagen.Fehler)
	}

	if len(vorlagen.Vorlagen) == 0 {
		t.Skip("die Saat bringt keine Exportvorlage mit")
	}

	textpfad := filepath.Join(ziel, "ausgabe.txt")

	ausfuhr := anwendung.VorlageAusfuehren(
		probekontext(t), geoeffnet.SetID, vorlagen.Vorlagen[0].Paket, textpfad, nil, nil)
	if ausfuhr.Fehler != nil {
		t.Fatalf("Vorlage ausführen: %+v", ausfuhr.Fehler)
	}

	pruefeDaLiegtEtwas(t, textpfad)
}

// **Auch eine Exportvorlage allein ist eine Ausgabe** — ein eigener Lauf,
// denn im Weg darüber hätte der Bogen schon vermerkt, und ein Fehlen hier
// fiele nicht auf.
func TestEineExportvorlageStehtUnterZuletztAusgegeben(t *testing.T) {
	anwendung := mitSaat(t)

	geoeffnet := anwendung.SetOeffnen(t.Context(), saatset(t))
	vorlagen := anwendung.Exportvorlagen(t.Context(), geoeffnet.SetID)

	if geoeffnet.Fehler != nil || vorlagen.Fehler != nil || len(vorlagen.Vorlagen) == 0 {
		t.Fatalf("öffnen %+v, Vorlagen %+v", geoeffnet.Fehler, vorlagen)
	}

	ausfuhr := anwendung.VorlageAusfuehren(probekontext(t), geoeffnet.SetID, vorlagen.Vorlagen[0].Paket,
		filepath.Join(t.TempDir(), "ausgabe.txt"), nil, nil)

	if ausgegeben := anwendung.Weiterarbeiten(t.Context()).Ausgegeben; ausfuhr.Fehler != nil ||
		len(ausgegeben) != 1 || ausgegeben[0].Objekt.Kennung != geoeffnet.SetID {
		t.Errorf("Ausfuhr %+v, zuletzt ausgegeben %+v", ausfuhr.Fehler, ausgegeben)
	}
}

// **Weg 6: aus MTG JSON übernehmen.**
//
// Der Weg für jemanden, der seine Karten anderswo hat: eine Datei von
// mtgjson.com aufmachen und ansehen, was daraus geworden ist.
func TestWegAusMTGJSONUebernehmen(t *testing.T) {
	anwendung := mitSaat(t)

	quelle, err := filepath.Abs(
		filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json"))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	schritt(t, "eine MTG-JSON-Datei wird geöffnet")

	geoeffnet := anwendung.MTGJSONOeffnen(t.Context(), quelle, "magic")
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	schritt(t, "die Karten stehen im Bestand")

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if len(karten.Karten) == 0 {
		t.Fatal("aus der Datei ist keine einzige Karte geworden")
	}

	// **Mit Werten und nicht als leere Hüllen.** Ein Import, der die richtige
	// Zahl Karten anlegt und keinen Namen trägt, sieht in einer Zählung wie
	// ein Erfolg aus.
	if kartennameAus(karten.Karten[0]) == "" {
		t.Error("die erste Karte hat keinen Namen")
	}
}

// pruefeDaLiegtEtwas sichert zu, dass eine Datei entstanden ist und nicht leer
// ist.
//
// **Beides und nicht nur das erste.** Eine Datei mit null Bytes entsteht auch
// dann, wenn das Schreiben mittendrin aufgibt — und sie sieht im
// Verzeichnislisting aus wie ein Erfolg.
func pruefeDaLiegtEtwas(t *testing.T, pfad string) {
	t.Helper()

	angaben, err := os.Stat(pfad)
	if err != nil {
		t.Fatalf("%s: %v", pfad, err)
	}

	if angaben.Size() == 0 {
		t.Errorf("%s ist leer", pfad)
	}
}
