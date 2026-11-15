package xai_test

// **Was hinausgeht** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Dies ist der Lauf, um den es im Issue geht
//
// Das Issue verlangt xAI „ohne Annahmen aus OpenAI-API-Ähnlichkeiten". Der
// Unterschied ist nicht theoretisch: **derselbe Anbieter** hat an
// `/v1/responses` und `/v1/chat/completions` **zwei verschiedene** Bildformen.
//
//	/v1/responses         {"type":"input_image","image_url":"data:…"}
//	/v1/chat/completions  {"type":"image_url","image_url":{"url":"data:…"}}
//
// Wer die erste schickte, bekäme einen 400 — **nach** dem Ruf und damit
// nachdem die Kosten entstanden sind. Dieser Lauf liest den Rumpf mit und
// sieht die Form, bevor irgendetwas hinausgeht.

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Ein Bild reist in der Form dieses Weges — und vor dem Text.**
func TestEinBildReistInDerFormDiesesWeges(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"model":"m","choices":[{"message":` +
		`{"content":"x"},"finish_reason":"stop"}]}`}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "grok-probe", System: "du bist knapp", Auftrag: "was ist das?",
		Bilder: []xai.Bild{{Medientyp: "image/png", Daten: "AAAA"}},
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	gestellt := gelesen(t, fuhrwerk.rumpf)

	nachrichten := alsListe(t, gestellt["messages"], "messages")
	if len(nachrichten) != 2 {
		t.Fatalf("die Nachrichten sind %v", nachrichten)
	}

	// **Die Systemanweisung ist hier eine Nachricht** — `instructions` gibt
	// es an diesem Weg nicht.
	system := alsKarte(t, nachrichten[0], "messages[0]")
	if system["role"] != "system" || system["content"] != "du bist knapp" {
		t.Errorf("die erste Nachricht ist %v", system)
	}

	nutzer := alsKarte(t, nachrichten[1], "messages[1]")

	teile := alsListe(t, nutzer["content"], "content")
	if len(teile) != 2 {
		t.Fatalf("die Teile sind %v", teile)
	}

	// **Das Bild zuerst**: „place the image(s) before the text".
	bild := alsKarte(t, teile[0], "content[0]")
	if bild["type"] != "image_url" {
		t.Errorf("die Bildart ist %v — an diesem Weg heißt sie image_url", bild["type"])
	}

	// **Und `image_url` ist hier ein Objekt und keine Zeichenkette.**
	adresse, istObjekt := bild["image_url"].(map[string]any)
	if !istObjekt {
		t.Fatalf("image_url ist %T — an diesem Weg ist es ein Objekt", bild["image_url"])
	}

	if adresse["url"] != "data:image/png;base64,AAAA" ||
		adresse["detail"] != "high" {
		t.Errorf("die Bildadresse ist %v", adresse)
	}

	text := alsKarte(t, teile[1], "content[1]")
	if text["type"] != "text" || text["text"] != "was ist das?" {
		t.Errorf("der Textteil ist %v", text)
	}
}

// **Ohne Systemanweisung gibt es keine Systemnachricht.**
//
// **Und keine leere**: eine Rolle `system` mit leerem Inhalt wäre eine
// Anweisung, die nichts anweist — und sie zählt beim Anbieter mit.
func TestOhneSystemanweisungGibtEsKeineSystemnachricht(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"model":"m","choices":[{"message":` +
		`{"content":"x"},"finish_reason":"stop"}]}`}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", System: "   ", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	nachrichten := alsListe(t, gelesen(t, fuhrwerk.rumpf)["messages"], "messages")
	if len(nachrichten) != 1 {
		t.Fatalf("die Nachrichten sind %v", nachrichten)
	}
}

// **Das Schema liegt hier eine Ebene tiefer als bei OpenAI.**
//
//	OpenAI  text.format         {type, name, schema, strict}
//	xAI     response_format     {type, json_schema:{name, schema, strict}}
func TestDasSchemaLiegtEineEbeneTiefer(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"model":"m","choices":[{"message":` +
		`{"content":"x"},"finish_reason":"stop"}]}`}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
		Schema: map[string]any{"type": "object"},
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	gestellt := gelesen(t, fuhrwerk.rumpf)

	form := alsKarte(t, gestellt["response_format"], "response_format")
	if form["type"] != "json_schema" {
		t.Fatalf("die Form ist %v", form)
	}

	schema, istObjekt := form["json_schema"].(map[string]any)
	if !istObjekt {
		t.Fatalf("json_schema ist %T — hier ist es ein eigenes Objekt",
			form["json_schema"])
	}

	if schema["name"] == "" || schema["strict"] != true {
		t.Errorf("die Schemaangabe ist %v", schema)
	}

	// **Und ohne Schema steht das Feld gar nicht da** — sonst verlangte
	// jede Frage eine feste Form.
	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if strings.Contains(fuhrwerk.rumpf, "response_format") {
		t.Errorf("ohne Schema steht %q im Rumpf", fuhrwerk.rumpf)
	}
}

// **Eine Frage, die niemand beantworten kann, geht nicht hinaus.**
func TestEineUnbeantwortbareFrageGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name  string
		frage xai.Frage
	}{
		{"ohne Modell", xai.Frage{Auftrag: "x"}},
		{"ohne Auftrag", xai.Frage{Alias: "m", Auftrag: "  "}},

		// **Webp nimmt OpenAI, nicht xAI.** Genau die Annahme, die dieses
		// Werkstück nicht machen darf.
		{"mit einem Bild in webp", xai.Frage{
			Alias: "m", Auftrag: "x",
			Bilder: []xai.Bild{{Medientyp: "image/webp", Daten: "AAAA"}},
		}},

		{"mit einem Bild ohne Daten", xai.Frage{
			Alias: "m", Auftrag: "x",
			Bilder: []xai.Bild{{Medientyp: "image/png"}},
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fuhrwerk := &mitschrift{}

			klient := xai.NeuerKlient("",
				netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
				xai.Zugang{Schluessel: "geheim"})

			_, fehler := klient.Rede(t.Context(), fall.frage)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if fuhrwerk.adresse != "" {
				t.Errorf("gerufen wurde trotzdem %q", fuhrwerk.adresse)
			}
		})
	}
}

// gelesen macht aus dem mitgeschriebenen Rumpf eine Abbildung.
func gelesen(t *testing.T, rumpf string) map[string]any {
	t.Helper()

	if rumpf == "" {
		t.Fatal("es ist kein Rumpf hinausgegangen — dieser Lauf prüft nichts")
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal([]byte(rumpf), &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v (%q)", fehler, rumpf)
	}

	return gestellt
}

// alsKarte liest ein Objekt aus dem Rumpf — und sagt, wenn keines dasteht.
func alsKarte(t *testing.T, wert any, wo string) map[string]any {
	t.Helper()

	karte, istKarte := wert.(map[string]any)
	if !istKarte {
		t.Fatalf("%s ist %T und kein Objekt", wo, wert)
	}

	return karte
}

// alsListe liest eine Liste aus dem Rumpf — und sagt, wenn keine dasteht.
func alsListe(t *testing.T, wert any, wo string) []any {
	t.Helper()

	liste, istListe := wert.([]any)
	if !istListe {
		t.Fatalf("%s ist %T und keine Liste", wo, wert)
	}

	return liste
}
