package service_test

// **Die Modi der Suche** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Wogegen hier geprüft wird
//
// **„Ausfall ergibt sichtbaren Moduswechsel nach deklarierter Policy, keine
// andere Quelle/Cloud."** Vier Wege führen zum lexikalischen Modus — kein
// Einbetter, ein fremder Vektorraum, ein schweigendes Modell und ein
// unlesbarer Vektorteil —, und jeder von ihnen sagt es.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Einbetter bleibt die Suche lexikalisch — und sagt es.**
//
// **Das Issue verlangt beides**: „exakte Suche funktioniert ohne installierten
// Encoder" und „Ausfall ergibt sichtbaren Moduswechsel nach deklarierter
// Policy". Ein stiller Rückfall wäre eine Suche, die weniger kann, ohne dass
// jemand es merkt.
func TestOhneEinbetterBleibtDieSucheLexikalisch(t *testing.T) {
	t.Parallel()

	ablage := &retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
		semantisch:  []domain.Retrievaltreffer{einRetrievaltreffer("zweites", 1, domain.QuelleOracle)},
	}

	ergebnis, fehler := einRetrievaldienst(ablage, nil).Suchen(t.Context(),
		service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusLexikalisch {
		t.Fatalf("der Modus ist %q", ergebnis.Modus)
	}

	if ergebnis.Grund == "" {
		t.Error("der Moduswechsel nennt keinen Grund")
	}

	if len(ergebnis.Treffer) != 1 ||
		ergebnis.Treffer[0].Dokument.Kennung != "eines" {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}

// **Ein Encoder, der nicht in diesen Raum passt, sucht nicht darin.**
//
// **Das ist der teuerste Fall des ganzen Werkstücks.** Das Issue: „Der neue
// Encoder darf nicht im alten Raum suchen." Er bekäme Treffer — falsche, ohne
// Meldung.
func TestEinFremderEncoderSuchtNichtImAltenRaum(t *testing.T) {
	t.Parallel()

	ablage := &retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
		semantisch:  []domain.Retrievaltreffer{einRetrievaltreffer("zweites", 1, domain.QuelleOracle)},
	}

	// **Gleiche Dimension, anderer Zerteiler** — genau der Fall, den eine
	// Längenprüfung durchließe.
	fremd := einRetrievalmanifest().Raum()
	fremd.Vertrag.Zerteiler = "ein anderer"

	ergebnis, fehler := einRetrievaldienst(ablage, &retrievaleinbetterdoppel{
		raum: fremd, vektor: []float32{1, 0, 0},
	}).Suchen(t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusLexikalisch {
		t.Fatalf("der Modus ist %q", ergebnis.Modus)
	}

	if !strings.Contains(ergebnis.Grund, "Vektorraum") {
		t.Errorf("der Grund ist %q", ergebnis.Grund)
	}

	// **Und der Einbetter wurde gar nicht erst gefragt** — ein Ruf hätte Zeit
	// gekostet für eine Antwort, die niemand benutzen darf.
	if len(ergebnis.Treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}

// **Die Frage bekommt ihre Anweisung, das Dokument seine.**
//
// **Das Issue trennt sie fachlich**: „`EmbedQuery` und `EmbedDocuments`
// fachlich unterscheiden." Ein Modell, das eine Frageanweisung verlangt,
// antwortet ohne sie anders — und dann liegt die Frage neben dem Korpus.
func TestDieFrageBekommtIhreAnweisung(t *testing.T) {
	t.Parallel()

	manifest := einRetrievalmanifest()
	manifest.Vertrag.Frageanweisung = "search_query: "

	einbetter := &retrievaleinbetterdoppel{
		raum: manifest.Raum(), vektor: []float32{1, 0, 0},
	}

	if _, fehler := einRetrievaldienst(&retrievalablagedoppel{manifest: manifest},
		einbetter).Suchen(t.Context(), service.Suchwunsch{
		Frage: "lightning", Grenze: 10,
	}); fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if len(einbetter.gesehen) != 1 ||
		einbetter.gesehen[0] != "search_query: lightning" {
		t.Fatalf("eingebettet wurde %+v", einbetter.gesehen)
	}
}

// **Scheitert der Einbetter, bleibt die Suche lexikalisch.**
//
// **Und sie sagt es** — das Issue: „Ausfall ergibt sichtbaren Moduswechsel
// nach deklarierter Policy, keine andere Quelle/Cloud."
func TestEinGescheiterterEinbetterLaesstDieSucheLexikalisch(t *testing.T) {
	t.Parallel()

	ablage := &retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
	}

	ergebnis, fehler := einRetrievaldienst(ablage, &retrievaleinbetterdoppel{
		raum:   einRetrievalmanifest().Raum(),
		fehler: domain.MitGrund(domain.ErrConflict, nil, "die Laufzeit schweigt"),
	}).Suchen(t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusLexikalisch {
		t.Fatalf("der Modus ist %q", ergebnis.Modus)
	}

	if !strings.Contains(ergebnis.Grund, "schweigt") {
		t.Errorf("der Grund ist %q", ergebnis.Grund)
	}

	if len(ergebnis.Treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}

// **Scheitert die lexikalische Hälfte, scheitert die Suche.**
//
// **Sie ist die Pflichthälfte** — ohne sie gibt es keine Grundlage, und ein
// Ergebnis aus der halben Suche wäre eines, das niemand einordnen kann.
func TestEineGescheiterteLexikalischeHaelfteIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest:  einRetrievalmanifest(),
		lexfehler: domain.MitGrund(domain.ErrConflict, nil, "der Bestand schweigt"),
	}, nil).Suchen(t.Context(), service.Suchwunsch{
		Frage: "lightning", Grenze: 10,
	}); fehler == nil {
		t.Fatal("die gescheiterte Suche kam als Ergebnis zurück")
	}
}

// **Ein Ablagefehler beim Manifest ist ein Fehler.**
//
// **Nicht jeder ist es**: ein fehlendes Manifest heißt „kein Index" und ist
// eine Auskunft ([TestOhneAktivenIndexSagtDieSucheEs]). Was hier geprüft
// wird, ist der andere Fall — der Bestand antwortet gar nicht.
func TestEinUnlesbaresManifestIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifestfehler: domain.MitGrund(domain.ErrConflict, nil,
			"der Bestand ist zu"),
	}, nil).Suchen(t.Context(), service.Suchwunsch{
		Frage: "lightning", Grenze: 10,
	}); fehler == nil {
		t.Fatal("der Bestandsfehler kam als leeres Ergebnis zurück")
	}
}

// **Ein unlesbarer Vektorteil lässt die Suche lexikalisch.**
func TestEinUnlesbarerVektorteilLaesstDieSucheLexikalisch(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
		semfehler:   domain.MitGrund(domain.ErrConflict, nil, "der BLOB ist halb"),
	}, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}).Suchen(t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusLexikalisch ||
		!strings.Contains(ergebnis.Grund, "halb") {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Zu viele Vektoren auf eine Frage lassen die Suche lexikalisch.**
//
// **Und das ist kein erfundener Fall**: `internal/ollama` weist ihn schon ab,
// aber der Dienst kennt beliebige Einbetter — ein zweiter Vektor ohne zweite
// Frage gehört zu nichts, und welcher der richtige ist, entscheidet dieses
// Programm nicht.
func TestZuVieleVektorenLassenDieSucheLexikalisch(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
	}, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
		zweimal: true,
	}).Suchen(t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusLexikalisch ||
		!strings.Contains(ergebnis.Grund, "2 Vektoren") {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Ein fremder Fehler bringt keinen Satz mit.**
//
// **Das ist die Zusage von [domain.SatzVon]**, hier eingelöst: was kein
// eigener Satz ist, erreicht die Oberfläche nicht (§8).
func TestEinFremderFehlerBringtKeinenSatzMit(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
	}, &retrievaleinbetterdoppel{
		raum:   einRetrievalmanifest().Raum(),
		fehler: errors.New("/Users/jemand/geheim.db: no such table"),
	}).Suchen(t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if strings.Contains(ergebnis.Grund, "geheim") {
		t.Fatalf("der fremde Text steht im Grund: %q", ergebnis.Grund)
	}

	if !strings.Contains(ergebnis.Grund, "nicht geantwortet") {
		t.Fatalf("der Grund ist %q", ergebnis.Grund)
	}
}

// **Mehr Treffer als gewünscht werden abgeschnitten.**
func TestMehrTrefferAlsGewuenschtWerdenAbgeschnitten(t *testing.T) {
	t.Parallel()

	viele := make([]domain.Retrievaltreffer, 0, 5)
	for i := range 5 {
		viele = append(viele, einRetrievaltreffer(
			"karte"+string(rune('a'+i)), i+1, domain.QuelleOracle))
	}

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest: einRetrievalmanifest(), lexikalisch: viele,
	}, nil).Suchen(t.Context(), service.Suchwunsch{Frage: "x", Grenze: 2})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if len(ergebnis.Treffer) != 2 {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}

// **Der Ausschnitt kommt aus der Liste, die einen hat.**
func TestDerAusschnittKommtAusDerListeDieEinenHat(t *testing.T) {
	t.Parallel()

	ohne := einRetrievaltreffer("beide", 1, domain.QuelleOracle)

	mit := einRetrievaltreffer("beide", 1, domain.QuelleOracle)
	mit.Ausschnitt = "…Lightning…"

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{ohne},
		semantisch:  []domain.Retrievaltreffer{mit},
	}, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}).Suchen(t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if len(ergebnis.Treffer) != 1 ||
		ergebnis.Treffer[0].Ausschnitt != "…Lightning…" {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}
