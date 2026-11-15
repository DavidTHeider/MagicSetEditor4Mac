package openai

// **Die Responses-API** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Was sie von der Messages-API unterscheidet
//
// **Der Ausgang steht an drei Stellen statt an einer.** Anthropic sagt alles
// in `stop_reason`. Hier steht `status` oben, `incomplete_details.reason`
// daneben — und eine **Ablehnung** steht an keiner von beiden, sondern als
// Inhaltsblock vom Typ `refusal`, während `status` auf `completed` steht.
//
// **Und die Anweisung hat einen eigenen Platz.** `instructions` steht neben
// `input` und hat laut Dokumentation Vorrang vor dem, was dort steht.
//
// **Nachgelesen am 2026-09-12** unter
// <https://developers.openai.com/api/docs/guides/structured-outputs> und
// <https://developers.openai.com/api/docs/guides/text-generation>.

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// hoechstensZeichen ist die Obergrenze der Ausgabe.
//
// **Sie steht hier und nicht beim Aufrufer**, weil sie eine Eigenschaft
// dieses Wegs ist: ein Kartenvorschlag ist ein kurzes JSON. Ohne Grenze
// bezahlte ein Denkmodell beliebig lange nach — `reasoning_tokens` zählen
// mit und stehen in keiner Antwort, die jemand liest.
const hoechstensZeichen = 4096

// Bild ist ein Bild, das mitgeschickt wird.
type Bild struct {
	// Medientyp ist `image/png`, `image/jpeg`, `image/webp` oder `image/gif`.
	Medientyp string

	// Daten sind die Bytes — base64 kodiert.
	Daten string
}

// Frage ist, was ein Modell beantworten soll.
type Frage struct {
	// Alias ist das Modell.
	Alias string

	// System ist die Anweisung — sie geht nach `instructions`.
	System string

	// Auftrag ist, was getan werden soll.
	Auftrag string

	// Bilder reisen **vor** dem Text mit.
	//
	// **Dieselbe Reihenfolge wie bei Anthropic**, und aus demselben Grund:
	// beide Dokumentationen zeigen es so, und ein Auftrag, der auf ein Bild
	// zeigt, das danach kommt, ist eine Anweisung ins Leere.
	Bilder []Bild

	// Schema ist das JSON-Schema, in das die Antwort gezwungen wird — leer
	// heißt: freier Text.
	Schema map[string]any
}

// Ausgang ist, wie eine Antwort geendet hat.
type Ausgang string

// Die vier Ausgänge, und drei davon sind kein Ergebnis.
const (
	// AusgangFertig heißt: die Antwort ist vollständig.
	AusgangFertig Ausgang = "fertig"

	// AusgangUnvollstaendig heißt: sie ist abgeschnitten.
	AusgangUnvollstaendig Ausgang = "unvollstaendig"

	// AusgangAbgelehnt heißt: das Modell wollte nicht — oder ein Filter hat
	// den Lauf angehalten.
	//
	// **Beides zusammen, und das ist eine Entscheidung.** Für den Aufrufer
	// ist es dasselbe: es gibt kein Ergebnis, und ein zweiter Versuch mit
	// demselben Auftrag hilft nicht. Woran es lag, steht in [Antwort.Grund].
	AusgangAbgelehnt Ausgang = "abgelehnt"

	// AusgangUnbekannt heißt: dieses Programm kennt den Grund nicht.
	//
	// **Und hält die Antwort deshalb nicht für fertig.** Ein neuer Stand
	// kommt bei jedem Anbieter irgendwann.
	AusgangUnbekannt Ausgang = "unbekannt"
)

// Antwort ist, was zurückkam.
type Antwort struct {
	// Inhalt ist der Text — bei einem Schema das JSON.
	Inhalt string

	// Ausgang ist, wie sie geendet hat.
	Ausgang Ausgang

	// Grund ist, was der Dienst dazu gesagt hat — `status` und, wo einer
	// dasteht, `incomplete_details.reason`.
	Grund string

	// Naeheres ist der Satz einer Ablehnung — leer, wenn es keinen gibt.
	Naeheres string

	// Modell ist, was wirklich geantwortet hat.
	//
	// **Nicht, wonach gefragt wurde**: ein Alias zeigt auf eine Fassung, und
	// die steht hier.
	Modell string

	// Verbrauch ist, was der Ruf gekostet hat — in Zeichen.
	Verbrauch Verbrauch
}

// Rede fragt ein Modell.
func (k *Klient) Rede(ctx context.Context, frage Frage) (Antwort, error) {
	if fehler := frage.pruefe(); fehler != nil {
		return Antwort{}, fehler
	}

	var roh antwortrumpf

	if fehler := k.hole(ctx, Antwortweg, frage.alsAnfrage(), &roh); fehler != nil {
		return Antwort{}, fehler
	}

	// **Ein gescheiterter Lauf ist ein Fehler und kein Ausgang.**
	//
	// Der Status ist dabei 200: die Anfrage ist angekommen, der Lauf ist
	// misslungen. Wer das als Ausgang führte, gäbe einen leeren Vorschlag
	// zurück — und die Schemaprüfung spräche danach von fehlenden Feldern.
	if roh.Fehler != nil {
		return Antwort{}, fmt.Errorf(
			"%w: OpenAI hat den Lauf abgebrochen (%s): %s", domain.ErrConflict,
			roh.Fehler.Code, roh.Fehler.Meldung)
	}

	text, ablehnung := inhaltAus(roh)

	return Antwort{
		Inhalt: text, Ausgang: ausgangFuer(roh, ablehnung),
		Grund: grundAus(roh), Naeheres: ablehnung,
		Modell: roh.Modell, Verbrauch: roh.Verbrauch,
	}, nil
}

// antwortrumpf ist die Gestalt einer Antwort.
type antwortrumpf struct {
	Status string `json:"status"`
	Modell string `json:"model"`

	Fehler *struct {
		Code    string `json:"code"`
		Meldung string `json:"message"`
	} `json:"error"`

	Unfertig *struct {
		Grund string `json:"reason"`
	} `json:"incomplete_details"`

	Ausgabe []struct {
		Art    string `json:"type"`
		Inhalt []struct {
			Art       string `json:"type"`
			Text      string `json:"text"`
			Ablehnung string `json:"refusal"`
		} `json:"content"`
	} `json:"output"`

	Verbrauch Verbrauch `json:"usage"`
}

// inhaltAus liest Text und Ablehnung aus den Ausgabeblöcken.
//
// **Alle Textblöcke zusammen.** Die Ausgabe ist eine Liste, und ein Modell
// darf sie teilen; wer nur den ersten läse, bekäme ein halbes JSON.
func inhaltAus(roh antwortrumpf) (text, ablehnung string) {
	var teile []string

	for _, block := range roh.Ausgabe {
		if block.Art != "message" {
			continue
		}

		for _, inhalt := range block.Inhalt {
			switch inhalt.Art {
			case "output_text":
				teile = append(teile, inhalt.Text)
			case "refusal":
				ablehnung = inhalt.Ablehnung
			}
		}
	}

	return strings.Join(teile, ""), ablehnung
}

// grundAus baut den Grund aus Status und Näherem.
func grundAus(roh antwortrumpf) string {
	if roh.Unfertig != nil && roh.Unfertig.Grund != "" {
		return roh.Status + "/" + roh.Unfertig.Grund
	}

	return roh.Status
}

// dieAbbruchgruende sind die Gründe, die ein Ergebnis unvollständig lassen.
var dieAbbruchgruende = []string{"max_output_tokens"}

// dieFiltergruende sind die Gründe, die es ganz verhindern.
var dieFiltergruende = []string{"content_filter"}

// ausgangFuer ordnet Status und Grund einem der vier Ausgänge zu.
//
// **Die Ablehnung zuerst**, denn sie steht neben `status: completed`: wer
// den Status zuerst läse, hielte sie für ein Ergebnis.
func ausgangFuer(roh antwortrumpf, ablehnung string) Ausgang {
	if ablehnung != "" {
		return AusgangAbgelehnt
	}

	if roh.Status == "completed" {
		return AusgangFertig
	}

	if roh.Status != "incomplete" || roh.Unfertig == nil {
		return AusgangUnbekannt
	}

	switch {
	case slices.Contains(dieAbbruchgruende, roh.Unfertig.Grund):
		return AusgangUnvollstaendig
	case slices.Contains(dieFiltergruende, roh.Unfertig.Grund):
		return AusgangAbgelehnt
	}

	return AusgangUnbekannt
}
