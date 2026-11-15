package openai

// **Die zwei Bildwege** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Die eine Regel, aus der alles andere folgt
//
// **Ein Bild erreicht dieses Programm als Bytes oder gar nicht**
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md) §4).
// Für die `gpt-image`-Familie ist genau das der dokumentierte Vertrag: sie
// gibt „always base64-encoded images", `url` ist für sie **nicht**
// unterstützt und `response_format` nimmt sie nicht an. Kommt doch eine
// Adresse, ist das Drift und wird laut — geholt wird sie nicht.
//
// Damit gibt es hier keinen Holweg, keine Frist auf einer fremden Adresse,
// keinen Umleiter und keinen Schlüssel, der zu einem CDN reisen könnte.
//
// # Und derselbe Pfad heißt nicht derselbe Rumpf
//
// `POST /v1/images/edits` gibt es bei OpenAI **und** bei xAI. Das Bild steht
// hier in `images` — einer **Liste** von `{"image_url": …}` —, dort in
// `image`, einem **Objekt** mit `type` und `url` (0101 §8, 0102 §5). Wer den
// einen Vertrag für den anderen hält, bekommt einen Fehler, der nach einem
// kaputten Bild aussieht.
//
// **Nachgelesen am 2026-09-12** unter
// <https://developers.openai.com/api/docs/api-reference/images/create> und
// <https://developers.openai.com/api/docs/api-reference/images/createEdit>.

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die zwei Bildwege der Positivliste.
const (
	// Bildweg erzeugt ein Bild aus Text.
	Bildweg = "/v1/images/generations"

	// Bildeditweg bearbeitet ein vorhandenes.
	Bildeditweg = "/v1/images/edits"
)

// hoechstensBild ist die Größe, ab der eine Bildantwort abgewiesen wird.
//
// **Zwölf Mebibyte**, und die Zahl folgt aus einer anderen: der Vorrat lässt
// acht Mebibyte je Bild durch ([service.UeblicheBildgrenzen]), Base64 bläht
// um ein Drittel auf, und darüber liegt noch der JSON-Rahmen. Wer mehr
// schickt, schickt etwas, das ohnehin nicht abgelegt würde.
const hoechstensBild int64 = 12 << 20

// Bildauftrag ist, was dieses Programm von einem Bild verlangt.
type Bildauftrag struct {
	Alias   string
	Auftrag string

	// Groesse ist `1024x1024`, `1536x1024` oder `1024x1536` — leer heißt:
	// der Dienst entscheidet.
	Groesse string

	// Guete ist `low`, `medium` oder `high` — leer heißt: der Dienst
	// entscheidet.
	//
	// **`auto`, `xhigh` und `max` sind dokumentiert und werden nicht
	// angeboten** (0102 §6): die ersten beiden kosten mehr, `auto` überlässt
	// dem Dienst, wie viel — und über einem Kostendeckel ist das die eine
	// Angabe, die man nicht machen will.
	Guete string

	// Quelle gesetzt heißt: **bearbeiten** statt erzeugen.
	//
	// **Ohne Medientyp daneben.** Er stünde als zweite Auskunft neben der
	// Signatur, und eine davon wäre irgendwann falsch — gelesen wird der
	// Kopf der Bytes.
	Quelle []byte
}

// Rohbild ist, was zurückkam.
type Rohbild struct {
	Inhalt    []byte
	Medientyp string
	Modell    string

	// Umschrift ist der Prompt, wie der Dienst ihn umgeschrieben hat — leer,
	// wenn er keinen nennt.
	//
	// **Das ist Herkunft und keine Randnotiz.** Gezeichnet wurde, was hier
	// steht, und nicht, was ein Mensch getippt hat.
	Umschrift string

	Verbrauch Verbrauch
}

// bildantwort ist die Gestalt, die OpenAI schickt.
type bildantwort struct {
	Erzeugt      int64       `json:"created"`
	Ausgabeform  string      `json:"output_format"`
	Daten        []bilddatum `json:"data"`
	Verbrauch    Verbrauch   `json:"usage"`
	Modellangabe string      `json:"model"`

	// Fehlerrumpf ist der Fehler **im 200er-Rumpf** — derselbe Fall wie bei
	// der Responses-API: die Anfrage ist angekommen, der Lauf misslungen.
	Fehlerrumpf *struct {
		Code    string `json:"code"`
		Meldung string `json:"message"`
	} `json:"error"`
}

// bilddatum ist ein Eintrag in `data`.
type bilddatum struct {
	Base64    string `json:"b64_json"`
	Adresse   string `json:"url"`
	Umschrift string `json:"revised_prompt"`
}

// bildanfrage ist der Rumpf, den dieses Programm schickt.
//
// **Ohne `response_format` und ohne `n`.** Das erste nimmt diese
// Modellfamilie nicht an; das zweite bleibt bei eins, weil ein Vorschlag ein
// Vorschlag ist — und weil jedes weitere Bild noch einmal Geld kostet.
type bildanfrage struct {
	Modell  string       `json:"model"`
	Prompt  string       `json:"prompt"`
	Groesse string       `json:"size,omitempty"`
	Guete   string       `json:"quality,omitempty"`
	Form    string       `json:"output_format,omitempty"`
	Bilder  []bildquelle `json:"images,omitempty"`
}

// bildquelle ist ein Eingangsbild — **eine Daten-URI und nie eine Adresse**.
//
// `file_id` gibt es hier nicht: es setzte `/v1/files` voraus, also einen Weg,
// der beim Anbieter etwas ablegt (0102 §5).
type bildquelle struct {
	Adresse string `json:"image_url"`
}

// Bild erzeugt eines oder bearbeitet eines.
func (k *Klient) Bild(ctx context.Context, auftrag Bildauftrag) (Rohbild, error) {
	anfrage, weg, fehler := auftrag.alsAnfrage()
	if fehler != nil {
		return Rohbild{}, fehler
	}

	var roh bildantwort

	if fehler := k.holeGross(ctx, weg, anfrage, &roh, hoechstensBild); fehler != nil {
		return Rohbild{}, fehler
	}

	if roh.Fehlerrumpf != nil {
		return Rohbild{}, fmt.Errorf(
			"%w: OpenAI hat den Bildauftrag abgebrochen (%s): %s",
			domain.ErrConflict, roh.Fehlerrumpf.Code, roh.Fehlerrumpf.Meldung)
	}

	return roh.alsRohbild(auftrag.Alias)
}

// alsRohbild liest das erste Datum — und weist eine Adresse ab.
func (a bildantwort) alsRohbild(alias string) (Rohbild, error) {
	if len(a.Daten) == 0 {
		return Rohbild{}, fmt.Errorf(
			"%w: OpenAI hat kein Bild geschickt", domain.ErrConflict)
	}

	datum := a.Daten[0]

	if datum.Base64 == "" {
		if datum.Adresse != "" {
			return Rohbild{}, fmt.Errorf(
				"%w: OpenAI hat eine Adresse statt der Bilddaten geschickt — "+
					"diese Anwendung holt sie nicht (Vermerk 0102)",
				domain.ErrConflict)
		}

		return Rohbild{}, fmt.Errorf(
			"%w: OpenAI hat ein leeres Bild geschickt", domain.ErrConflict)
	}

	inhalt, fehler := base64.StdEncoding.DecodeString(datum.Base64)
	if fehler != nil {
		return Rohbild{}, domain.MitGrund(domain.ErrConflict, fehler,
			"die Bilddaten von OpenAI auspacken")
	}

	modell := a.Modellangabe
	if modell == "" {
		modell = alias
	}

	return Rohbild{
		Inhalt: inhalt, Medientyp: medientypFuer(a.Ausgabeform, inhalt),
		Modell: modell, Umschrift: datum.Umschrift, Verbrauch: a.Verbrauch,
	}, nil
}

// alsAnfrage baut Rumpf und Weg — und sagt nein, bevor etwas hinausgeht.
func (a Bildauftrag) alsAnfrage() (bildanfrage, string, error) {
	if strings.TrimSpace(a.Alias) == "" {
		return bildanfrage{}, "", fmt.Errorf(
			"%w: ein Bildauftrag braucht ein Modell", domain.ErrValidation)
	}

	if strings.TrimSpace(a.Auftrag) == "" {
		return bildanfrage{}, "", fmt.Errorf(
			"%w: ein Bildauftrag braucht eine Beschreibung", domain.ErrValidation)
	}

	anfrage := bildanfrage{
		Modell: a.Alias, Prompt: a.Auftrag,
		Groesse: a.Groesse, Guete: a.Guete, Form: "png",
	}

	if len(a.Quelle) == 0 {
		return anfrage, Bildweg, nil
	}

	// **Die Quelle wird hier geprüft und nicht beim Anbieter.** Was keine
	// Bildsignatur trägt, ist entweder kaputt oder etwas anderes — und der
	// Ruf kostete in beiden Fällen Geld.
	typ := signaturtyp(a.Quelle)
	if typ == "" {
		return bildanfrage{}, "", fmt.Errorf(
			"%w: das Bild, das bearbeitet werden soll, trägt keine "+
				"Bildsignatur", domain.ErrValidation)
	}

	anfrage.Bilder = []bildquelle{{
		Adresse: "data:" + typ + ";base64," +
			base64.StdEncoding.EncodeToString(a.Quelle),
	}}

	return anfrage, Bildeditweg, nil
}

// bildvorspaenne ordnet den ersten Bytes einen Medientyp zu.
var bildvorspaenne = []struct {
	vorspann []byte
	typ      string
}{
	{[]byte("\x89PNG\r\n\x1a\n"), "image/png"},
	{[]byte("\xff\xd8\xff"), "image/jpeg"},
	{[]byte("RIFF"), "image/webp"},
}

// signaturtyp gibt den Medientyp der Signatur — leer, wenn keine passt.
func signaturtyp(inhalt []byte) string {
	for _, vorspann := range bildvorspaenne {
		if bytes.HasPrefix(inhalt, vorspann.vorspann) {
			return vorspann.typ
		}
	}

	return ""
}

// medientypFuer sagt, was zurückkam.
//
// **Die Signatur schlägt die Angabe.** `output_format` ist, was der Dienst
// sagt; die ersten Bytes sind, was dasteht — und nur eines von beidem lässt
// sich nachsehen.
func medientypFuer(ausgabeform string, inhalt []byte) string {
	if typ := signaturtyp(inhalt); typ != "" {
		return typ
	}

	if ausgabeform == "" {
		return ""
	}

	return "image/" + ausgabeform
}
