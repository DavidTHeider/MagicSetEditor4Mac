package app_test

import (
	"bytes"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Bildnaht kommt über einem echten Bestand zusammen**
// (#1005).
//
// Geprüft wird die **Verdrahtung** — dass aus einer `*sql.DB` etwas
// Benutzbares wird und dass die Kartenprüfung vor dem Netz greift.
func TestDieBildnahtKommtUeberEinemBestandZusammen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	naht, fehler := app.Fremdbildnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	// **Eine Karte, die es nicht gibt, scheitert vor dem Netz.**
	if _, err := naht.Einsetzen(probekontext(t), 1, 4711, "image", "Prüfstein",
		"normal"); err == nil {
		t.Error("die fehlende Karte kam durch")
	}
}

// **Ohne Zuhause kommt sie nicht zusammen.**
func TestOhneZuhauseKommtDieBildnahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, fehler := app.Fremdbildnaht(nil, domain.UrheberKommandozeile); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Und der Bildweg selbst geht bis zur Gegenstelle.**
//
// # Warum das ein eigener Test ist
//
// [App.ScryfallBildnaht] baut Klient, Drossel und Zwischenspeicher. Ein Test
// über ein Doppel prüfte das Doppel; hier steht eine echte Gegenstelle, und
// damit ist belegt, dass die Adresse, der Kopf und die Bytes ankommen.
func TestDerBildwegGehtBisZurGegenstelle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bild := []byte("\x89PNG\r\n\x1a\nBilddaten")

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.Header.Get("User-Agent") == "" {
				t.Error("es wurde ohne User-Agent gefragt")
			}

			if strings.Contains(anfrage.URL.Path, "/bild.png") {
				sende(t, schreiben, bild)

				return
			}

			schreiben.Header().Set("Content-Type", "application/json")
			sende(t, schreiben, []byte(`{"object":"card","id":"a","name":"Prüfstein",`+
				`"artist":"Wer auch immer","scryfall_uri":"https://scryfall.invalid/x",`+
				`"image_uris":{"normal":"`+adresseFuer(anfrage)+`"}}`))
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	quelle, fehler := app.ScryfallBildnaht(app.ImmerAn{})
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	holung, holfehler := quelle.Holen(probekontext(t), "Prüfstein", "normal")
	if holfehler != nil {
		t.Fatalf("Holen: %v", holfehler)
	}

	if !strings.Contains(holung.Nennung, "Wer auch immer") ||
		!strings.Contains(holung.Nennung, "Wizards of the Coast") {
		t.Errorf("Nennung = %q", holung.Nennung)
	}

	// **Und eine unbekannte Art kommt gar nicht erst bis dorthin.**
	if _, err := quelle.Holen(probekontext(t), "Prüfstein", "riesig"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler = %v, erwartet ErrValidation", err)
	}
}

// sende schreibt eine Antwort und meldet, wenn das nicht geht.
//
// **Ein Helfer statt `_, _ =`**: der Linter liest die Zuweisung als
// verschluckten Fehler, und er hat recht — eine Gegenstelle, die nicht
// schreiben kann, macht aus jedem Befund einen anderen.
func sende(t *testing.T, schreiben http.ResponseWriter, rumpf []byte) {
	t.Helper()

	// **G705 sieht hier eine Ausgabe, die aus der Anfrage stammen könnte** —
	// die Bildadresse trägt `anfrage.Host`. Es ist eine Gegenstelle in einem
	// Test, sie antwortet keinem Browser, und der Host kommt aus `httptest`.
	//nolint:gosec // G705: Prüfgegenstelle, keine Ausgabe an einen Browser
	if _, fehler := schreiben.Write(rumpf); fehler != nil {
		t.Errorf("die Gegenstelle konnte nicht schreiben: %v", fehler)
	}
}

// adresseFuer baut die Bildadresse auf derselben Gegenstelle.
func adresseFuer(anfrage *http.Request) string {
	return "http://" + anfrage.Host + "/bild.png"
}

// **Der ganze Bildweg, einmal gefahren: Gegenstelle → Datei → Bestand**
// (#1005).
//
// # Warum dieser Test sein muss
//
// Er ist die Antwort auf die Frage, ob der Bezug **verdrahtet** ist. Alles
// darunter ist mit Doubles geprüft; keiner davon zeigt, dass ein Bild von
// Scryfall wirklich an einer Karte landet — und damit in Vorschau, Druckbogen
// und `.mse-set`-Ausfuhr.
//
// **Und dass der Künstler mitgeht.** Ohne ihn verletzt jede Ausgabe die
// Bedingung, zu der sich diese Anwendung mit jedem Abruf verpflichtet.
func TestDerGanzeBildwegLandetImBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bild := einBildpunkt(t)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if strings.HasSuffix(anfrage.URL.Path, "/kunst.png") {
				sende(t, schreiben, bild)

				return
			}

			sende(t, schreiben, []byte(`{
				"object": "card",
				"id": "00000000-0000-4000-8000-000000000002",
				"name": "Prüfstein",
				"layout": "normal",
				"artist": "Christopher Rush",
				"scryfall_uri": "https://beispiel.invalid/karte/2",
				"image_uris": {"normal": "http://`+anfrage.Host+`/kunst.png"}
			}`))
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	saatInsPaketverzeichnis(t)

	db := frischerBestand(t)
	setID, karteID := setMitKunstfenster(t, db)

	naht, fehler := app.Fremdbildnaht(db, domain.UrheberKommandozeile)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	eingesetzt, bildfehler := naht.Einsetzen(
		probekontext(t), setID, karteID, "kunst", "Prüfstein", "normal")
	if bildfehler != nil {
		t.Fatalf("Einsetzen: %v", bildfehler)
	}

	if !strings.Contains(eingesetzt.Nennung, "Christopher Rush") ||
		!strings.Contains(eingesetzt.Nennung, "Wizards of the Coast") {
		t.Errorf("Nennung = %q", eingesetzt.Nennung)
	}

	// **Nachgesehen im Bestand** — das ist der Punkt: was hier steht, sehen
	// Vorschau, Druckbogen, Ausfuhr, MCP und die Kommandozeile.
	danach, _, standfehler := app.Kartenstand(probekontext(t), db, setID, karteID)
	if standfehler != nil {
		t.Fatalf("Kartenstand: %v", standfehler)
	}

	if danach["kunst"] == "" {
		t.Error("im Bestand steht kein Bildname — das Bild ist nirgends angekommen")
	}

	if !eingesetzt.IllustratorGesetzt || danach["illustrator"] != "Christopher Rush" {
		t.Errorf("illustrator = %q (gesetzt: %v) — ohne ihn verlässt das Bild die "+
			"Anwendung ohne seine Nennung", danach["illustrator"],
			eingesetzt.IllustratorGesetzt)
	}

	// **Und ein fremdes Set kommt nicht durch.**
	//
	// Der Abruf glückt — er kennt nur die Karte —, das Einsetzen scheitert.
	// Das ist die Wache aus #808 an dieser Stelle: die Karte gegen ihr eigenes
	// Set zu halten ginge immer auf.
	_, fremdfehler := naht.Einsetzen(
		probekontext(t), setID+1, karteID, "kunst", "Prüfstein", "normal")
	if fremdfehler == nil {
		t.Error("das fremde Set kam durch")
	} else if !strings.Contains(fremdfehler.Error(), "einsetzen") {
		t.Errorf("der Fehler kommt aus dem falschen Schritt: %v", fremdfehler)
	}
}

// einBildpunkt gibt ein PNG, das wirklich eines ist.
//
// **Erzeugt und nicht abgeschrieben.** Der erste Anlauf setzte einen
// base64-Block hierher, den ich für ein 1×1-PNG hielt; `image.Decode` sagte
// dazu „not enough pixel data", und der Testlauf meldete eine Validierung, die
// wie ein Fehler des Bildwegs aussah. Ein `png.Encode` kann diesen Fehler
// nicht machen.
func einBildpunkt(t *testing.T) []byte {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := range 300 {
		for x := range 400 {
			bild.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	var gepackt bytes.Buffer
	if fehler := png.Encode(&gepackt, bild); fehler != nil {
		t.Fatalf("das Prüfbild ließ sich nicht schreiben: %v", fehler)
	}

	return gepackt.Bytes()
}

// setMitKunstfenster legt ein Set an, dessen Stil ein Bildfeld hat.
//
// **Aus der Saat und nicht erfunden**: `probe-kunst.mse-style` hat eine
// `kunst`-Fläche mit `render style: image`, und ohne eine solche Fläche findet
// [App.bildziel] kein Ziel.
func setMitKunstfenster(t *testing.T, db *sql.DB) (setID, karteID int64) {
	t.Helper()

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Kunstprobe", Game: "probe", Style: "probe-kunst"}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	karteID, kartenfehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID)
	if kartenfehler != nil {
		t.Fatalf("Karte anlegen: %v", kartenfehler)
	}

	return setID, karteID
}

// saatInsPaketverzeichnis legt die Saatvorlagen dorthin, wo eine Fläche ohne
// Fenster sie sucht.
//
// # Warum das nötig ist und was es über den Betrieb sagt
//
// [Zeichennaht] sucht über [Paketverzeichnis], [Vorlagenverzeichnis] und
// [Saatverzeichnis]. Die letzten beiden rechnen vom **Programmpfad** — im
// Bündel liegt die Saat daneben, unter `go test` ist der Programmpfad eine
// temporäre Datei, und beide sind leer. Das Paketverzeichnis dagegen hängt am
// Zuhause, und das setzt dieser Test.
//
// **Es ist also kein Kunstgriff, sondern derselbe Weg**, den ein bezogenes
// Paket geht.
func saatInsPaketverzeichnis(t *testing.T) {
	t.Helper()

	ziel, fehler := app.Paketverzeichnis()
	if fehler != nil {
		t.Fatalf("Paketverzeichnis: %v", fehler)
	}

	wurzel, wegfehler := filepath.Abs(filepath.Join("..", "..", "saat"))
	if wegfehler != nil {
		t.Fatalf("Saatwurzel: %v", wegfehler)
	}

	for _, paket := range []string{"probe.mse-game", "probe-kunst.mse-style"} {
		kopiereBaum(t, filepath.Join(wurzel, paket), filepath.Join(ziel, paket))
	}
}

// kopiereBaum kopiert ein Paketverzeichnis.
func kopiereBaum(t *testing.T, von, nach string) {
	t.Helper()

	eintraege, fehler := os.ReadDir(von)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", von, fehler)
	}

	if err := os.MkdirAll(nach, 0o750); err != nil {
		t.Fatalf("%s anlegen: %v", nach, err)
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			kopiereBaum(t, filepath.Join(von, eintrag.Name()), filepath.Join(nach, eintrag.Name()))

			continue
		}

		inhalt, lesefehler := os.ReadFile(filepath.Join(von, eintrag.Name())) //nolint:gosec // Pfad aus der Saat dieses Repositoriums
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), lesefehler)
		}

		//nolint:gosec // G703: Ziel und Name kommen aus t.TempDir() und der Saat dieses Repositoriums
		if err := os.WriteFile(
			filepath.Join(nach, eintrag.Name()), inhalt, 0o600); err != nil {
			t.Fatalf("%s schreiben: %v", eintrag.Name(), err)
		}
	}
}

// **Ein Bildordner, der als Datei dasteht, wird gemeldet.**
//
// Der Fall ist nicht abwegig: `~/Library/Application Support/…/scryfall/`
// gehört diesem Programm, aber es ist nicht das einzige, das dort schreiben
// darf. Ein `bilder` als Datei wäre sonst ein stiller Ausfall des Bildwegs.
func TestEinBelegterBildordnerWirdGemeldet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, fehler := app.Scryfallverzeichnis()
	if fehler != nil {
		t.Fatalf("Scryfallverzeichnis: %v", fehler)
	}

	if err := os.WriteFile(filepath.Join(ort, "bilder"), []byte("keine Ordner"), 0o600); err != nil {
		t.Fatalf("belegen: %v", err)
	}

	if _, err := app.ScryfallBildnaht(app.ImmerAn{}); err == nil {
		t.Error("der belegte Bildordner blieb ohne Fehler")
	}
}

// **Ohne Scryfall gibt es keinen Bildweg** — Phase 4.8.
//
// Die Sperre sitzt an der Anfrage; das Bauen glückt. Was hier geprüft wird,
// ist der **Fehlerweg des Bauens**: ein Klient, der nicht entsteht, darf
// keinen halben Bildweg zurücklassen.
func TestOhneZuhauseKommtDerBildwegNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, err := app.ScryfallBildnaht(app.ImmerAn{}); err == nil {
		t.Error("ohne Zuhause kam ein Bildweg")
	}
}
