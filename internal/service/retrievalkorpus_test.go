package service_test

// **Was aus einer Sammeldatei wird** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Die Daten sind selbst geschrieben** (§3, §11): dieselbe Syntax wie eine
// echte Sammeldatei, eigener Inhalt.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// derAbruf ist die Zeit, zu der die Sammeldatei geholt wurde.
var derAbruf = time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

// derKorpusstand nennt Revision und Gültigkeitsdatum.
func derKorpusstand() domain.Korpusstand {
	return domain.Korpusstand{
		Quelle: domain.QuelleOracle, Revision: "2026-09-01", Datum: "2026-09-01",
	}
}

// eineSammeldatei baut eine JSONL-Datei aus den gegebenen Zeilen.
//
// **Je Karte eine Zeile, und die Umbrüche der Konstanten fallen weg.** Eine
// Sammeldatei ist JSONL; ein Objekt über mehrere Zeilen wären mehrere
// kaputte Zeilen.
func eineSammeldatei(t *testing.T, zeilen ...string) *service.Scryfallkorpus {
	t.Helper()

	einzeilig := make([]string, 0, len(zeilen))
	for _, zeile := range zeilen {
		einzeilig = append(einzeilig, strings.ReplaceAll(zeile, "\n", " "))
	}

	return service.NeuerScryfallkorpus(
		scryfall.NeuerBulkleser(strings.NewReader(strings.Join(einzeilig, "\n"))),
		derKorpusstand(), derAbruf)
}

// eineEinseitige ist eine gewöhnliche Karte.
const eineEinseitige = `{"object":"card","id":"d1","oracle_id":"o1",
	"name":"Funkenschlag","lang":"de","layout":"normal",
	"mana_cost":"{R}","cmc":1,"type_line":"Spontanzauber",
	"oracle_text":"Funkenschlag fügt 3 Schadenspunkte zu.",
	"flavor_text":"Ein Blitz aus heiterem Himmel.",
	"colors":["R"],"color_identity":["R"],
	"legalities":{"modern":"legal","standard":"not_legal"},
	"set":"tst","set_name":"Test","collector_number":"1","rarity":"common",
	"scryfall_uri":"https://beispiel.invalid/d1"}`

// eineZweiseitige hat zwei Regeltexte.
const eineZweiseitige = `{"object":"card","id":"d2","oracle_id":"o2",
	"name":"Wandler // Gewandelter","lang":"de","layout":"transform",
	"cmc":2,"type_line":"Kreatur // Kreatur",
	"colors":[],"color_identity":["G"],
	"set":"tst","set_name":"Test","collector_number":"2","rarity":"rare",
	"scryfall_uri":"https://beispiel.invalid/d2",
	"card_faces":[
	  {"name":"Wandler","mana_cost":"{1}{G}","type_line":"Kreatur — Mensch",
	   "oracle_text":"Verwandle Wandler."},
	  {"name":"Gewandelter","type_line":"Kreatur — Werwolf",
	   "oracle_text":"Gewandelter hat Trampelschaden."}]}`

// **Eine einseitige Karte wird ein Abschnitt mit voller Herkunft.**
func TestEineEinseitigeKarteWirdEinAbschnitt(t *testing.T) {
	t.Parallel()

	stapel, fehler := eineSammeldatei(t, eineEinseitige).Stapel(t.Context())
	if fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	if len(stapel) != 1 {
		t.Fatalf("der Stapel ist %+v", stapel)
	}

	abschnitt := stapel[0]

	if fehler := abschnitt.Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	// **Die Identität kommt aus der Orakelkennung und nicht aus dem Namen.**
	if abschnitt.Kennung != "scryfall-oracle:o1:0" ||
		abschnitt.Fachkennung != "o1" ||
		abschnitt.Semantik != domain.SemantikKonzept ||
		abschnitt.Kennungssystem != domain.SystemScryfall {
		t.Errorf("die Identität ist %q / %q / %q / %q", abschnitt.Kennung,
			abschnitt.Fachkennung, abschnitt.Semantik, abschnitt.Kennungssystem)
	}

	// **Abrufzeit und Gültigkeitsdatum sind zwei verschiedene Angaben.**
	if !abschnitt.Abgerufen.Equal(derAbruf) || abschnitt.Quelldatum != "2026-09-01" {
		t.Errorf("%v / %q", abschnitt.Abgerufen, abschnitt.Quelldatum)
	}

	// **Der Aromatext ist nicht dabei.**
	if strings.Contains(abschnitt.Text, "heiterem Himmel") {
		t.Errorf("der Text ist %q", abschnitt.Text)
	}

	if !strings.Contains(abschnitt.Text, "3 Schadenspunkte") ||
		!strings.Contains(abschnitt.Text, "Spontanzauber") {
		t.Errorf("der Text ist %q", abschnitt.Text)
	}
}

// **Eine zweiseitige Karte wird zwei Abschnitte, die ihre Kennung teilen.**
func TestEineZweiseitigeKarteWirdZweiAbschnitte(t *testing.T) {
	t.Parallel()

	stapel, fehler := eineSammeldatei(t, eineZweiseitige).Stapel(t.Context())
	if fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	if len(stapel) != 2 {
		t.Fatalf("der Stapel ist %+v", stapel)
	}

	if stapel[0].Titel != "Wandler" || stapel[1].Titel != "Gewandelter" {
		t.Errorf("die Titel sind %q und %q", stapel[0].Titel, stapel[1].Titel)
	}

	// **Der Bezug bleibt rekonstruierbar** — dieselbe Fachkennung, zwei
	// Abschnittsnummern.
	if stapel[0].Fachkennung != stapel[1].Fachkennung ||
		stapel[0].Abschnitt != 0 || stapel[1].Abschnitt != 1 {
		t.Errorf("die Bezüge sind %+v", stapel)
	}

	if stapel[0].Kennung == stapel[1].Kennung {
		t.Errorf("beide heißen %q", stapel[0].Kennung)
	}

	// **Und die Farbidentität der Karte gilt für beide Seiten.**
	for _, abschnitt := range stapel {
		if len(abschnitt.Merkmale.Farben) != 0 ||
			len(abschnitt.Merkmale.Farbidentitaet) != 1 {
			t.Errorf("die Merkmale sind %+v", abschnitt.Merkmale)
		}
	}
}

// **Die Merkmale kommen aus der Karte und nicht aus ihrem Text.**
func TestDieMerkmaleKommenAusDerKarte(t *testing.T) {
	t.Parallel()

	stapel, fehler := eineSammeldatei(t, eineEinseitige).Stapel(t.Context())
	if fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	merkmale := stapel[0].Merkmale

	if merkmale.Manawert == nil || *merkmale.Manawert != 1 ||
		merkmale.Set != "tst" {
		t.Errorf("die Merkmale sind %+v", merkmale)
	}

	// **Der Datenstand der Legalitäten ist der der Sammeldatei** — nicht der
	// des Sets und nicht heute.
	lage, stand, bekannt := merkmale.Format("modern")
	if !bekannt || lage != "legal" || stand != "2026-09-01" {
		t.Errorf("%q / %q / %v", lage, stand, bekannt)
	}

	if !slices.Contains(merkmale.Typen, "spontanzauber") {
		t.Errorf("die Typen sind %v", merkmale.Typen)
	}
}

// **Eine Karte ohne Orakelkennung kommt nicht in den Index.**
//
// **Weil sie keine Identität hat** — und ein Treffer ohne Identität ist eine
// Behauptung. Das Issue: „keine künstlichen Scryfall-IDs für Eigenkarten",
// und eine erfundene Kennung hier wäre dasselbe.
func TestEineKarteOhneOrakelkennungKommtNichtInDenIndex(t *testing.T) {
	t.Parallel()

	ohne := strings.Replace(eineEinseitige, `"oracle_id":"o1",`, "", 1)

	stapel, fehler := eineSammeldatei(t, ohne).Stapel(t.Context())
	if fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	if len(stapel) != 0 {
		t.Fatalf("der Stapel ist %+v", stapel)
	}
}

// **Eine kaputte Zeile kostet ihre Zeile und wird benannt.**
func TestEineKaputteZeileKostetIhreZeile(t *testing.T) {
	t.Parallel()

	korpus := eineSammeldatei(t, `{"kein":`, eineEinseitige)

	stapel, fehler := korpus.Stapel(t.Context())
	if fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	if len(stapel) != 1 {
		t.Fatalf("der Stapel ist %+v", stapel)
	}

	if len(korpus.Befunde()) != 1 {
		t.Fatalf("die Befunde sind %+v", korpus.Befunde())
	}
}

// **Ein Stand, der keine Oracle-Texte nennt, passt nicht zu dieser Quelle.**
func TestEinFalscherStandPasstNichtZurSammeldatei(t *testing.T) {
	t.Parallel()

	korpus := service.NeuerScryfallkorpus(
		scryfall.NeuerBulkleser(strings.NewReader("")),
		domain.Korpusstand{Quelle: domain.QuelleRegelwerk, Revision: "1"},
		derAbruf)

	if _, fehler := korpus.Staende(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("Staende: %v", fehler)
	}

	richtig := eineSammeldatei(t, eineEinseitige)

	staende, fehler := richtig.Staende(t.Context())
	if fehler != nil || len(staende) != 1 {
		t.Fatalf("%+v, %v", staende, fehler)
	}
}

// **Ist die Datei zu Ende, kommt nichts mehr.**
func TestNachDemEndeKommtNichtsMehr(t *testing.T) {
	t.Parallel()

	korpus := eineSammeldatei(t, eineEinseitige)

	if _, fehler := korpus.Stapel(t.Context()); fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	zweiter, fehler := korpus.Stapel(t.Context())
	if fehler != nil || len(zweiter) != 0 {
		t.Fatalf("%+v, %v", zweiter, fehler)
	}
}

// **Ein abgebrochener Kontext hält den Korpus an.**
func TestEinAbgebrochenerKontextHaeltDenKorpusAn(t *testing.T) {
	t.Parallel()

	kontext, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := eineSammeldatei(t, eineEinseitige).Stapel(kontext); !errors.Is(
		fehler, context.Canceled) {
		t.Fatalf("Stapel: %v", fehler)
	}
}

// **Eine Karte ohne Typzeile bekommt keine Typwörter.**
//
// **Und nicht ein leeres.** Eine leere Liste im Merkmal hieße „geprüft und
// keine gefunden"; `nil` heißt „unbekannt", und der Filter behandelt beides
// gleich — der Unterschied steht trotzdem im Bestand, und dort zählt er.
func TestEineKarteOhneTypzeileBekommtKeineTypwoerter(t *testing.T) {
	t.Parallel()

	ohne := strings.Replace(eineEinseitige, `"type_line":"Spontanzauber",`,
		`"type_line":"",`, 1)

	stapel, fehler := eineSammeldatei(t, ohne).Stapel(t.Context())
	if fehler != nil {
		t.Fatalf("Stapel: %v", fehler)
	}

	if len(stapel) != 1 || stapel[0].Merkmale.Typen != nil {
		t.Fatalf("die Merkmale sind %+v", stapel[0].Merkmale)
	}
}

// **Ein voller Stapel hört bei seiner Größe auf.**
//
// **Und der nächste holt den Rest.** Ohne diesen Lauf stünde die Grenze da,
// ohne je erreicht zu werden — und ein Aufbau über 400.000 Karten liefe in
// einem Zug in den Speicher.
func TestEinVollerStapelHoertBeiSeinerGroesseAuf(t *testing.T) {
	t.Parallel()

	zeilen := make([]string, 0, service.KorpusstapelGroesse+5)
	for i := range service.KorpusstapelGroesse + 5 {
		zeilen = append(zeilen, strings.ReplaceAll(eineEinseitige,
			`"oracle_id":"o1"`, fmt.Sprintf(`"oracle_id":"o%d"`, i)))
	}

	korpus := eineSammeldatei(t, zeilen...)

	erster, fehler := korpus.Stapel(t.Context())
	if fehler != nil || len(erster) != service.KorpusstapelGroesse {
		t.Fatalf("der erste Stapel hat %d Abschnitte: %v", len(erster), fehler)
	}

	zweiter, fehler := korpus.Stapel(t.Context())
	if fehler != nil || len(zweiter) != 5 {
		t.Fatalf("der zweite Stapel hat %d Abschnitte: %v", len(zweiter), fehler)
	}
}

// **Ein Strom, der abbricht, bricht den Korpus ab.**
//
// **Und gibt keine halben Abschnitte heraus.** Eine Sammeldatei, die mitten
// im Lesen endet, ist kein gültiger Stand — das Issue: „beschädigte/inkomplette
// Quelle ersetzt keinen gültigen Stand."
func TestEinAbbrechenderStromBrichtDenKorpusAb(t *testing.T) {
	t.Parallel()

	korpus := service.NeuerScryfallkorpus(
		scryfall.NeuerBulkleser(io.MultiReader(
			strings.NewReader(strings.ReplaceAll(eineEinseitige, "\n", " ")+"\n"),
			iotest.ErrReader(errors.New("die Platte antwortet nicht")))),
		derKorpusstand(), derAbruf)

	stapel, fehler := korpus.Stapel(t.Context())
	if fehler == nil {
		t.Fatalf("der Strom kam durch: %+v", stapel)
	}

	if stapel != nil {
		t.Errorf("es kamen trotzdem %d Abschnitte heraus", len(stapel))
	}
}
