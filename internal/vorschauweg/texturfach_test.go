package vorschauweg_test

// **Das Fach der Spielbilder** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum es neben dem Vorschaufach steht und nicht an seiner Stelle
//
// **Weil beide etwas anderes zählen.** Das Vorschaufach hält **acht** Bilder:
// mehr als eine Vorschau steht nie auf dem Bildschirm. Ein Spieltisch zeigt
// sechzig, und die Frage ist nicht „wie viele", sondern „wie viel Speicher"
// — eine Miniatur und ein Betrachtungsbild unterscheiden sich um den Faktor
// hundert.
//
// # Und warum der Schlüssel nicht der Inhalt ist
//
// **Weil er vor dem Zeichnen feststeht.** Das Vorschaufach adressiert über
// den Abdruck der **Bytes**; die hat es, wenn es ablegt. Hier ist die Frage
// umgekehrt: *muss* überhaupt gezeichnet werden? Ein Schlüssel aus den
// Eingaben beantwortet sie, ein Abdruck der Ausgabe nicht.

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// bild gibt ein Bild der gewünschten Größe.
func bild(groesse int) []byte { return make([]byte, groesse) }

// einPNG gibt ein winziges, echtes PNG.
//
// **Echt und nicht acht Bytes**: der Weg schreibt es unverändert hinaus, und
// ein Lauf, der nur die Länge zählte, ließe eine Datei durch, die kein Bild
// ist (§10, derselbe Satz wie an `istBild`).
func einPNG() []byte {
	flaeche := image.NewRGBA(image.Rect(0, 0, 2, 2))
	flaeche.Set(0, 0, color.RGBA{R: 255, A: 255})

	var gepackt bytes.Buffer
	_ = png.Encode(&gepackt, flaeche) //nolint:errcheck // ein Puffer klagt nicht

	return gepackt.Bytes()
}

// **Was hinterlegt wurde, kommt unter seiner Adresse wieder.**
func TestEinHinterlegtesSpielbildKommtWieder(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)

	adresse := fach.Hinterlege("abc", bild(100), 375, 523)

	if !strings.HasPrefix(adresse, vorschauweg.Spielpfad) ||
		!strings.HasSuffix(adresse, ".png") {
		t.Errorf("die Adresse %q sieht nicht aus wie eine", adresse)
	}

	if _, _, _, da := fach.Adresse("abc"); !da {
		t.Error("der Schlüssel ist nicht wiederzufinden")
	}

	gelesen, da := fach.Hole("abc")
	if !da || len(gelesen) != 100 {
		t.Errorf("gelesen: %d Bytes, da=%v", len(gelesen), da)
	}
}

// **Ein unbekannter Schlüssel sagt es, statt ein leeres Bild zu geben.**
func TestEinUnbekannterSchluesselSagtEs(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)

	if _, _, _, da := fach.Adresse("gibtesnicht"); da {
		t.Error("ein unbekannter Schlüssel hat eine Adresse")
	}

	if _, da := fach.Hole("gibtesnicht"); da {
		t.Error("ein unbekannter Schlüssel gibt ein Bild")
	}
}

// **Das Budget wird gehalten, und das Älteste fällt zuerst hinaus.**
func TestDasBudgetWirdGehalten(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(250)

	fach.Hinterlege("a", bild(100), 375, 523)
	fach.Hinterlege("b", bild(100), 375, 523)
	fach.Hinterlege("c", bild(100), 375, 523)

	if belegt := fach.Belegt(); belegt > 250 {
		t.Errorf("%d Bytes belegt, das Budget ist 250", belegt)
	}

	if _, _, _, da := fach.Adresse("a"); da {
		t.Error("das älteste Bild steht noch da")
	}

	for _, schluessel := range []string{"b", "c"} {
		if _, _, _, da := fach.Adresse(schluessel); !da {
			t.Errorf("%q ist verschwunden", schluessel)
		}
	}
}

// **Wer wieder geholt wird, ist jung.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Ein Spieltisch greift immer wieder auf dieselben Karten.** Verdrängte das
// Fach nach der **Einfügereihenfolge**, fiele nach `A, B, A` das A heraus,
// obwohl es gerade gebraucht wurde — dieselbe Lehre, die das Vorschaufach
// schon einmal gekostet hat
// (#1278,
// Befund 1). Dort ging es ums Hinterlegen; hier zählt auch das **Holen**,
// denn ein Bild, das im Cache liegt, wird nicht noch einmal hinterlegt.
func TestWerGeholtWirdBleibtLaenger(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(250)

	fach.Hinterlege("a", bild(100), 375, 523)
	fach.Hinterlege("b", bild(100), 375, 523)

	// **A wird gebraucht** — und rückt damit ans Ende der Reihe.
	if _, da := fach.Hole("a"); !da {
		t.Fatal("a ist schon weg")
	}

	fach.Hinterlege("c", bild(100), 375, 523)

	if _, _, _, da := fach.Adresse("a"); !da {
		t.Error("a ist hinausgefallen, obwohl es gerade geholt wurde")
	}

	if _, _, _, da := fach.Adresse("b"); da {
		t.Error("b steht noch da, obwohl es das älteste ist")
	}
}

// **Dasselbe zweimal zu hinterlegen belegt den Platz einmal.**
func TestZweimalDasselbeBelegtEinmal(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)

	fach.Hinterlege("a", bild(100), 375, 523)
	fach.Hinterlege("a", bild(100), 375, 523)

	if belegt := fach.Belegt(); belegt != 100 {
		t.Errorf("%d Bytes belegt, erwartet 100", belegt)
	}
}

// **Ein Bild größer als das Budget wird gar nicht erst abgelegt.**
//
// **Und es bekommt trotzdem eine Adresse** — sonst stünde der Aufrufer vor
// der Frage, ob er es noch einmal zeichnen soll, und zeichnete es jedes Mal.
// Die Adresse zeigt dann ins Leere, und der Weg antwortet mit 404: eine
// sichtbare Diagnose statt eines leeren Bildes (#1459).
func TestEinZuGrossesBildFaelltGleichWiederHinaus(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(100)

	adresse := fach.Hinterlege("riesig", bild(200), 375, 523)

	if adresse == "" {
		t.Error("ein zu großes Bild bekommt keine Adresse")
	}

	if belegt := fach.Belegt(); belegt != 0 {
		t.Errorf("%d Bytes belegt, erwartet 0", belegt)
	}
}

// **Ein Profilwechsel räumt das Fach.**
//
// **Weil die Bilder einem Profil gehören.** Wer wechselt, sieht die Karten
// eines anderen — und ein Fach, das sie noch hielte, gäbe sie unter ihrer
// Adresse weiter heraus (#1459: „Profilwechsel behandeln").
func TestEinProfilwechselRaeumtDasFach(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)

	fach.Hinterlege("a", bild(100), 375, 523)
	fach.Raeume()

	if belegt := fach.Belegt(); belegt != 0 {
		t.Errorf("%d Bytes belegt nach dem Räumen", belegt)
	}

	if _, _, _, da := fach.Adresse("a"); da {
		t.Error("nach dem Räumen steht noch ein Bild da")
	}
}

// **Der Weg liefert auch Spielbilder aus — unter ihrem eigenen Vorsatz.**
func TestDerWegLiefertSpielbilderAus(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)
	adresse := fach.Hinterlege("abc", einPNG(), 375, 523)

	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitTexturen(fach)

	auf := httptest.NewRecorder()
	weg.ServeHTTP(auf, httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, adresse, http.NoBody))

	if auf.Code != http.StatusOK {
		t.Fatalf("Status %d", auf.Code)
	}

	if art := auf.Header().Get("Content-Type"); art != "image/png" {
		t.Errorf("Content-Type %q", art)
	}

	if auf.Body.Len() != len(einPNG()) {
		t.Errorf("%d Bytes", auf.Body.Len())
	}
}

// **Ein Spielbild, das hinausgefallen ist, gibt 404 — keine leere Antwort.**
//
// **Das ist die sichtbare Diagnose aus #1459**: ein leeres PNG sähe aus wie
// eine Karte ohne Inhalt, und niemand suchte nach der Ursache.
func TestEinVerdraengtesSpielbildGibt404(t *testing.T) {
	t.Parallel()

	weg := vorschauweg.Neu(&vorschauweg.Fach{}).
		MitTexturen(vorschauweg.NeuesTexturfach(1024))

	auf := httptest.NewRecorder()
	weg.ServeHTTP(auf, httptest.NewRequestWithContext(t.Context(),
		http.MethodGet, vorschauweg.Spielpfad+"gibtesnicht.png", http.NoBody))

	if auf.Code != http.StatusNotFound {
		t.Errorf("Status %d, erwartet 404", auf.Code)
	}
}

// **Und ein Weg ohne Texturfach antwortet ebenso, statt zu stürzen.**
//
// **Der Fall gibt es wirklich**: die Kommandozeile und der Dienst bauen keinen
// Spieltisch, und ihr Weg trägt deshalb kein Texturfach.
func TestEinWegOhneTexturfachStuerztNicht(t *testing.T) {
	t.Parallel()

	weg := vorschauweg.Neu(&vorschauweg.Fach{})

	auf := httptest.NewRecorder()
	weg.ServeHTTP(auf, httptest.NewRequestWithContext(t.Context(),
		http.MethodGet, vorschauweg.Spielpfad+"abc.png", http.NoBody))

	if auf.Code != http.StatusNotFound {
		t.Errorf("Status %d, erwartet 404", auf.Code)
	}
}

// **Das Fach gibt die Maße zurück, die es bekommen hat.**
//
// # Wofür dieser Lauf da ist
//
// **Weil das Fach sie seit dem 2026-09-15 mitführt, und zwar aus einem
// gemessenen Grund.** Vorher gab es nur die Adresse; wer im Fach fündig
// wurde, musste die Maße **ausrechnen** — und die gerechnete Zahl ging von
// der gemessenen ab: derselbe Kartenrücken meldete beim ersten Ruf 512 × 512
// und beim zweiten 512 × 714 (#1459:
// „unveränderter Inhalt trifft ihn").
func TestDasFachGibtDieMasseZurueck(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)
	fach.Hinterlege("abc", bild(100), 375, 523)

	_, breite, hoehe, da := fach.Adresse("abc")
	if !da {
		t.Fatal("nicht gefunden")
	}

	if breite != 375 || hoehe != 523 {
		t.Errorf("%d×%d, erwartet 375×523", breite, hoehe)
	}

	// **Und was nicht da ist, hat keine Maße** — statt der Null eines
	// Bildes, das niemand abgelegt hat.
	adresse, breite, hoehe, da := fach.Adresse("fehlt")
	if da || adresse != "" || breite != 0 || hoehe != 0 {
		t.Errorf("%q %d×%d da=%v", adresse, breite, hoehe, da)
	}
}

// **Und ein zu großes Bild bekommt eine Adresse, aber keine Maße.**
//
// **Weil es nicht im Fach liegt.** Die Adresse zeigt ins Leere und [Weg]
// antwortet mit 404 — eine sichtbare Diagnose statt eines leeren Bildes.
// Wer die Maße trotzdem bekäme, hielte ein Bild für abgelegt, das es nicht
// ist.
func TestEinZuGrossesBildHatKeineMasse(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(50)

	adresse := fach.Hinterlege("riesig", bild(200), 375, 523)
	if adresse == "" {
		t.Error("keine Adresse")
	}

	if _, breite, hoehe, da := fach.Adresse("riesig"); da {
		t.Errorf("gefunden mit %d×%d — es liegt nicht im Fach", breite, hoehe)
	}
}

// **Ein Spielbild bleibt nicht im Speicher der Webansicht** (E06, #2278) —
// derselbe Grund wie bei der Vorschau: jede geänderte Karte bekommt eine neue
// Adresse, und `immutable` ließ WebKit jede alte behalten.
func TestEinSpielbildBleibtNichtImSpeicherDerWebansicht(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1024)
	fach.Hinterlege("abc", bild(10), 1, 1)

	weg := vorschauweg.Neu(&vorschauweg.Fach{}).MitTexturen(fach)

	auf := httptest.NewRecorder()
	weg.ServeHTTP(auf, httptest.NewRequestWithContext(t.Context(),
		http.MethodGet, vorschauweg.Spielpfad+"abc.png", http.NoBody))

	if auf.Code != http.StatusOK || auf.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Status %d, Cache-Control %q, erwartet 200 und no-store",
			auf.Code, auf.Header().Get("Cache-Control"))
	}
}
