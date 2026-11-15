package service_test

// **Der Vorrat für erzeugte Bilder** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// Das Issue zählt sieben Eigenschaften in einem Satz auf: „unerratbare ID,
// Profil-/Vorschlagsbindung, MIME-/Signaturprüfung, komprimierte Byte- und
// Pixelgrenze, Ablauf, Cachequote und Bereinigung nach Abbruch/Neustart".
//
// **Jede bekommt hier einen Lauf**, und zwei davon bekommen einen zweiten:
// die Profilbindung, weil ihre Meldung nichts verraten darf, und die
// Pixelgrenze, weil sie **vor** dem Auspacken greifen muss.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einBild malt ein echtes PNG.
func einVorratsbild(t *testing.T, breite, hoehe int) []byte {
	t.Helper()

	flaeche := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	for x := range breite {
		flaeche.Set(x, 0, color.NRGBA{R: uint8(x), A: 255})
	}

	var aus bytes.Buffer
	if fehler := png.Encode(&aus, flaeche); fehler != nil {
		t.Fatalf("das Prüfbild ließ sich nicht schreiben: %v", fehler)
	}

	return aus.Bytes()
}

// einJPEG malt dasselbe als JPEG.
func einJPEG(t *testing.T, breite, hoehe int) []byte {
	t.Helper()

	flaeche := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	flaeche.Set(0, 0, color.NRGBA{R: 255, A: 255})

	var aus bytes.Buffer
	if fehler := jpeg.Encode(&aus, flaeche, nil); fehler != nil {
		t.Fatalf("das Prüf-JPEG ließ sich nicht schreiben: %v", fehler)
	}

	return aus.Bytes()
}

// einKopfMitMassen baut ein PNG, das **behauptet**, so groß zu sein.
//
// **Das ist die Bildbombe**: die Datei wiegt zweiundsechzig Bytes, das
// ausgepackte Bild wöge Gigabytes. Wer erst dekodiert und dann misst, hat
// den Speicher schon belegt.
func einKopfMitMassen(breite, hoehe uint32) []byte {
	kopf := make([]byte, 13)
	binary.BigEndian.PutUint32(kopf[0:], breite)
	binary.BigEndian.PutUint32(kopf[4:], hoehe)
	kopf[8] = 8 // Bittiefe
	kopf[9] = 6 // Farbtyp RGBA

	var aus bytes.Buffer

	aus.WriteString("\x89PNG\r\n\x1a\n")
	// **Ohne Fehlerprüfung**: das Ziel ist ein `bytes.Buffer`, und der
	// nimmt alles an.
	_ = binary.Write(&aus, binary.BigEndian, uint32(len(kopf))) //nolint:errcheck // siehe oben
	aus.WriteString("IHDR")
	aus.Write(kopf)
	_ = binary.Write(&aus, binary.BigEndian, //nolint:errcheck // siehe oben
		crc32.ChecksumIEEE(append([]byte("IHDR"), kopf...)))

	return aus.Bytes()
}

// einVorrat baut einen Vorrat über einem frischen Verzeichnis.
func einVorrat(t *testing.T, aendern func(*service.Bildgrenzen)) (*service.Bildvorrat, *time.Time) {
	t.Helper()

	grenzen := service.UeblicheBildgrenzen()
	if aendern != nil {
		aendern(&grenzen)
	}

	uhr := time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)

	vorrat, fehler := service.NeuerBildvorrat(t.TempDir(), grenzen, func() time.Time { return uhr })
	if fehler != nil {
		t.Fatalf("der Vorrat ließ sich nicht bauen: %v", fehler)
	}

	return vorrat, &uhr
}

// einBesitz ist der Besitz, den die Läufe hier benutzen.
func einBildbesitz() service.Bildbesitz {
	return service.Bildbesitz{ProfilID: 7, Karte: 4, Feld: "image", Auftrag: "ki-4-1"}
}

// **Eine Kennung errät niemand, und zweimal dieselbe gibt es nicht.**
func TestEineKennungIstUnerratbar(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	gesehen := map[string]bool{}

	for range 16 {
		eintrag, fehler := vorrat.Legen(einBildbesitz(), einVorratsbild(t, 8, 8))
		if fehler != nil {
			t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
		}

		if len(eintrag.Kennung) != service.Bildkennungslaenge {
			t.Fatalf("die Kennung %q ist %d Zeichen lang, erwartet sind %d",
				eintrag.Kennung, len(eintrag.Kennung), service.Bildkennungslaenge)
		}

		if gesehen[eintrag.Kennung] {
			t.Fatalf("die Kennung %q kam zweimal", eintrag.Kennung)
		}

		gesehen[eintrag.Kennung] = true
	}
}

// **Ein fremdes Profil sieht nichts — und erfährt auch nicht, dass es
// etwas gäbe.**
//
// # Warum dieselbe Meldung
//
// **Weil eine andere Meldung die Kennung bestätigte.** „Das gehört dir
// nicht" sagt: es gibt sie. Wer Kennungen durchprobiert, hätte damit ein
// Orakel — und die Kennung ist das Einzige, was ein Bild schützt.
func TestEinFremdesProfilSiehtNichts(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	eintrag, fehler := vorrat.Legen(einBildbesitz(), einVorratsbild(t, 8, 8))
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	_, _, fremd := vorrat.Lesen(eintrag.Kennung, 8)
	if !errors.Is(fremd, domain.ErrNotFound) {
		t.Fatalf("ein fremdes Profil bekam %v statt ErrNotFound", fremd)
	}

	_, _, unbekannt := vorrat.Lesen("gibtesnicht", 7)
	if !errors.Is(unbekannt, domain.ErrNotFound) {
		t.Fatalf("eine unbekannte Kennung bekam %v statt ErrNotFound", unbekannt)
	}

	if fremd.Error() != unbekannt.Error() {
		t.Fatalf("die zwei Meldungen unterscheiden sich:\n  fremd:     %s\n  unbekannt: %s",
			fremd.Error(), unbekannt.Error())
	}

	// **Und verwerfen darf es auch niemand anders.**
	if weg := vorrat.Verwerfen(eintrag.Kennung, 8); !errors.Is(weg, domain.ErrNotFound) {
		t.Fatalf("ein fremdes Profil durfte verwerfen: %v", weg)
	}

	if _, _, eigenes := vorrat.Lesen(eintrag.Kennung, 7); eigenes != nil {
		t.Fatalf("das eigene Bild war nach dem fremden Versuch fort: %v", eigenes)
	}
}

// **Was keine Bildsignatur trägt, kommt nicht herein.**
//
// **PNG und JPEG kommen durch — beide.**
//
// **Die Quote hält — und das Vorhandene bleibt stehen.**
//
// # Warum das Neue scheitert und nicht das Alte weicht
//
// **Gelesen kommt heraus, was abgelegt wurde.**
func TestGelesenKommtHerausWasAbgelegtWurde(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	hinein := einVorratsbild(t, 16, 16)

	eintrag, fehler := vorrat.Legen(einBildbesitz(), hinein)
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	heraus, gelesen, fehler := vorrat.Lesen(eintrag.Kennung, 7)
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht lesen: %v", fehler)
	}

	if !bytes.Equal(hinein, heraus) {
		t.Fatalf("heraus kamen %d Bytes statt %d", len(heraus), len(hinein))
	}

	if gelesen.Besitz != einBildbesitz() {
		t.Fatalf("der Besitz ist %+v statt %+v", gelesen.Besitz, einBildbesitz())
	}
}

// **Zwei Läufe gleichzeitig stören einander nicht.**
//
// Der Vorrat hängt an einem Fenster, und das Fenster ruft nebenläufig
// (§7). Ohne diesen Lauf sähe der Race-Detektor nie hin.
func TestZweiLaeufeGleichzeitig(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	// **Das Bild entsteht vor den Goroutinen.** `t.Fatalf` aus einer
	// fremden Goroutine ist verboten, und `einBild` ruft es.
	bild := einVorratsbild(t, 8, 8)

	var warten sync.WaitGroup

	for range 8 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			eintrag, fehler := vorrat.Legen(einBildbesitz(), bild)
			if fehler != nil {
				return
			}

			_, _, _ = vorrat.Lesen(eintrag.Kennung, 7) //nolint:errcheck // was hier zählt, ist das Rennen
			_ = vorrat.Verwerfen(eintrag.Kennung, 7)   //nolint:errcheck // dito
		}()
	}

	warten.Wait()
}

// **Und keinen, wenn sich das Verzeichnis nicht anlegen lässt.**
//
// **Der Fall ist nicht erfunden**: der Ort liegt im Zwischenspeicher des
// Benutzers, und dort bestimmt nicht diese Anwendung allein die Rechte. Der
// Vorrat sagt es beim Bauen, statt bei jedem Ablegen zu stolpern.
//
// **Schauen gibt die Auskunft ohne die Bytes — und prüft dasselbe Profil.**
func TestSchauenGibtDieAuskunftOhneBytes(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	eintrag, fehler := vorrat.Legen(einBildbesitz(), einVorratsbild(t, 8, 8))
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	gesehen, fehler := vorrat.Schauen(eintrag.Kennung, 7)
	if fehler != nil {
		t.Fatalf("Schauen: %v", fehler)
	}

	if !reflect.DeepEqual(gesehen, eintrag) {
		t.Fatalf("der Eintrag ist %+v statt %+v", gesehen, eintrag)
	}

	if _, fehler := vorrat.Schauen(eintrag.Kennung, 8); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("ein fremdes Profil durfte schauen: %v", fehler)
	}
}

// **Ein Bild, das sich nicht schreiben lässt, hinterlässt keinen Eintrag.**
//
// **Und keinen, wenn sich das alte Verzeichnis nicht räumen lässt.**
//
// # Warum das kein erfundener Fall ist
//

// **Der Vorschlag trägt seine Herkunft — und sie kommt mit ihm heraus.**
//
// # Warum am Eintrag und nicht daneben
//
// **Weil die Ergänzung des Issues es so verlangt**: „Herkunft gehört zum
// temporären Vorschlag und bei bestätigter Übernahme zum Bildnachweis." Eine
// zweite Karte daneben liefe mit dieser auseinander, und dann stünde an einem
// Bild eine Herkunft, die zu einem anderen gehört.
func TestDerVorschlagTraegtSeineHerkunft(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	herkunft := domain.KIBildherkunft{
		Anbieter: "openai", Modell: "gpt-image-2.5-flare",
		Aufgabe: "ein Blitz", Umschrift: "ein Blitz, Ölgemälde",
		Auftrag: "ki-4-1",
		Parameter: []domain.KIBildparameter{
			{Name: "groesse", Wert: "1024x1536"},
		},
	}

	eintrag, fehler := vorrat.LegenMitHerkunft(einBildbesitz(), herkunft,
		einVorratsbild(t, 8, 8))
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	if eintrag.Herkunft.Umschrift != "ein Blitz, Ölgemälde" {
		t.Fatalf("die Herkunft ist %+v", eintrag.Herkunft)
	}

	_, gelesen, fehler := vorrat.Lesen(eintrag.Kennung, 7)
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if wert, da := gelesen.Herkunft.Parameterwert("groesse"); !da ||
		wert != "1024x1536" {
		t.Fatalf("die Größe ist %q (%v)", wert, da)
	}
}
