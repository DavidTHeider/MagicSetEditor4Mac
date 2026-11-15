package ollama

// **Die Einbettung** — Werkstück A7 von Phase 5.3
// (#1446),
// Ergänzung vom 2026-09-10.
//
// # Warum der Vertrag von außen kommt
//
// **Weil er nicht aus einer Modellliste folgt.** Der [Einbetter] bekommt einen
// [domain.KIEinbettungsvertrag] und prüft jede Antwort dagegen — Zahl der
// Vektoren, Länge, endliche Werte. Woher der Vertrag stammt, ist eine andere
// Frage; [Klient.Vermessen] beantwortet sie mit einem echten Lauf.
//
// # Und warum nichts still gekürzt wird
//
// **Weil ein gekürzter Text einen Vektor gibt, der zu einem anderen Text
// gehört.** Ollamas `truncate` steht auf `true`, wenn man nichts sagt
// (**nachgelesen am 2026-09-11** unter <https://docs.ollama.com/api/embed>) —
// dieses Paket sagt ausdrücklich `false` und nimmt den Fehler.

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Einbettungsweg ist der fünfte Weg der Positivliste.
const Einbettungsweg = "/api/embed"

// HoechstensJeStapel begrenzt eine Anfrage.
//
// **Zweiunddreißig**, und die Zahl ist eine Entscheidung dieses Programms und
// keine Vorgabe von Ollama: ein Stapel wird als **ein** Auftrag geladen, und
// wer tausend Texte auf einmal schickte, hielte die Laufzeit für alle anderen
// besetzt. Das Issue verlangt „begrenzte Batches"; wo die Grenze liegt, sagt
// es nicht.
const HoechstensJeStapel = 32

// Einbetter bettet Texte mit einem Modell und seinem Vertrag ein.
type Einbetter struct {
	klient *Klient
	raum   domain.Vektorraum
}

// Einbetter baut ihn.
//
// # Warum ein ganzer Raum und nicht nur ein Vertrag
//
// **Weil ein Vertrag zwei Modelle nicht unterscheidet.** `all-minilm` und
// `bge-small` sind beide 384-dimensional, mitteln, normalisieren und
// zerteilen mit demselben Zerteiler — ihre Verträge sind gleich, ihre
// Vektorräume nicht. Wer nur den Vertrag mitführte, ließe den einen im Index
// des anderen suchen: Treffer, falsche, ohne Meldung
// ([domain.Vektorraum], A15).
//
// **Der Alias ist `raum.Modell`.** Zwei Felder dafür wären zwei Wahrheiten.
func (k *Klient) Einbetter(raum domain.Vektorraum) *Einbetter {
	return &Einbetter{klient: k, raum: raum}
}

// Raum gibt den Vektorraum, in dem dieser Einbetter rechnet.
func (e *Einbetter) Raum() domain.Vektorraum { return e.raum }

// Vertrag gibt den Vertrag, gegen den dieser Einbetter prüft.
func (e *Einbetter) Vertrag() domain.KIEinbettungsvertrag { return e.raum.Vertrag }

// einbettungsanfrage ist der Rumpf von `/api/embed`.
type einbettungsanfrage struct {
	Modell   string   `json:"model"`
	Eingaben []string `json:"input"`

	// Kuerzen steht auf `false` und ist deshalb **ausgeschrieben**: mit
	// `omitempty` fiele es aus dem Rumpf, und Ollamas Vorgabe `true` griffe.
	Kuerzen bool `json:"truncate"`
}

// einbettungsantwort ist, was zurückkommt.
type einbettungsantwort struct {
	Modell    string      `json:"model"`
	Vektoren  [][]float32 `json:"embeddings"`
	Zeichen   int         `json:"prompt_eval_count"`
	Ladedauer int64       `json:"load_duration"`
}

// Einbetten gibt je Text einen Vektor.
//
// **Es erfüllt `service.Einbettungsquelle`** — und ist damit das, was den
// Vertragslauf aus A6 fährt.
func (e *Einbetter) Einbetten(ctx context.Context, texte []string) (
	[][]float32, error,
) {
	if len(texte) == 0 {
		return nil, nil
	}

	if len(texte) > HoechstensJeStapel {
		return nil, fmt.Errorf(
			"%w: %d Texte auf einmal, höchstens %d — ein Stapel wird als ein "+
				"Auftrag geladen", domain.ErrValidation, len(texte), HoechstensJeStapel)
	}

	var antwort einbettungsantwort

	if fehler := e.klient.hole(ctx, Einbettungsweg, einbettungsanfrage{
		Modell: e.raum.Modell, Eingaben: texte, Kuerzen: false,
	}, &antwort); fehler != nil {
		return nil, fehler
	}

	if len(antwort.Vektoren) != len(texte) {
		return nil, fmt.Errorf(
			"%w: auf %d Texte kamen %d Vektoren zurück — welcher zu welchem "+
				"gehört, entscheidet dieses Programm nicht",
			domain.ErrValidation, len(texte), len(antwort.Vektoren))
	}

	for stelle, vektor := range antwort.Vektoren {
		if fehler := e.raum.Vertrag.PruefeVektor(vektor); fehler != nil {
			return nil, fmt.Errorf("der %d. Vektor von %q: %w",
				stelle+1, e.raum.Modell, fehler)
		}
	}

	return antwort.Vektoren, nil
}

// Raum vermisst ein Modell und nennt den Vektorraum, in dem es rechnet.
//
// # Wofür es das gibt
//
// **Damit Aufbau und Suche denselben Raum meinen.** Ein Index trägt Modell,
// Revision, Laufzeit und Vertrag in seinem Manifest; wer später sucht, muss
// **dieselben** Angaben nennen können, sonst ist jeder Vergleich ein Raten.
// Beide holen sie hier.
//
// # Die drei Quellen, und warum es drei sind
//
// **Der Vertrag kommt aus [Klient.Vermessen]** — einem echten Lauf, nicht
// aus Metadaten. **Die Revision kommt aus `/api/tags`**: Ollamas `digest` ist
// der Abdruck der Gewichte, und ein Modell mit demselben Namen und anderem
// Digest ist ein anderes Modell. **Die Fassung kommt aus `/api/version`**,
// weil eine andere Laufzeit andere Zahlen liefern kann.
//
// **Befunde sind kein Fehler.** Wer sie bekommt, hat kein Modell — aber auch
// keinen Ausfall, und der Unterschied gehört dem Aufrufer.
func (k *Klient) Raum(ctx context.Context, alias string) (
	domain.Vektorraum, []domain.Befund, error,
) {
	vertrag, befunde, fehler := k.Vermessen(ctx, alias)
	if fehler != nil || len(befunde) > 0 {
		return domain.Vektorraum{}, befunde, fehler
	}

	fassung, fehler := k.Fassung(ctx)
	if fehler != nil {
		return domain.Vektorraum{}, nil, fehler
	}

	modelle, modellbefunde, fehler := k.Modelle(ctx)
	if fehler != nil {
		return domain.Vektorraum{}, modellbefunde, fehler
	}

	for _, modell := range modelle {
		if !derselbeAlias(modell.Alias, alias) ||
			len(modell.Herkunft.Glieder) == 0 {
			continue
		}

		// **Der Digest ist Revision und Artefakthash zugleich**, und das ist
		// keine Nachlässigkeit: bei Ollama *ist* der Abdruck der Gewichte die
		// Revision. Ein zweites Feld mit demselben Wert sagte nichts mehr.
		digest := modell.Herkunft.Glieder[0].Revision

		return domain.Vektorraum{
			Modell: alias, Modellrevision: digest, Artefakthash: digest,
			Laufzeit: Anbieter + " " + fassung,
			// **Cosinus, weil der Vertrag normalisierte Vektoren zusagt** —
			// und wo er es nicht tut, ordnet das Skalarprodukt anders. Das
			// steht im Raum und nicht in einer Fußnote.
			Abstandsmass: abstandFuer(vertrag),
			Vertrag:      vertrag,
		}, nil, nil
	}

	return domain.Vektorraum{}, []domain.Befund{{
		Datei: alias, Art: domain.ArtUmgebung,
		Meldung: fmt.Sprintf("%q ist vermessen, steht aber nicht in der Liste "+
			"der installierten Modelle — ohne seinen Digest ist nicht zu "+
			"sagen, welche Gewichte gerechnet haben", alias),
	}}, nil
}

// abstandFuer sagt, womit in diesem Raum verglichen wird.
//
// **Normalisiert heißt Cosinus, sonst Skalarprodukt.** Bei normalisierten
// Vektoren ordnen beide gleich; bei unnormalisierten nicht, und dann ist der
// Cosinus die Rechnung, die die Länge herausdividiert — was hier niemand
// will, weil sie dann Bedeutung trägt.
func abstandFuer(vertrag domain.KIEinbettungsvertrag) domain.Abstandsmass {
	if vertrag.Normalisiert {
		return domain.AbstandCosinus
	}

	return domain.AbstandPunkt
}

// derselbeAlias sagt, ob zwei Modellnamen dasselbe Modell meinen.
//
// # Warum das nötig ist, und woher die Regel kommt
//
// **Gemessen am 2026-09-12 an Ollama 0.33.3.** `/api/show` und `/api/embed`
// nehmen `all-minilm` entgegen und antworten; `/api/tags` führt dasselbe
// Modell als `all-minilm:latest`. Wer beide Namen zeichenweise verglich,
// bekam einen Befund für ein Modell, das gerade gerechnet hatte — genau das
// ist im ersten Messlauf passiert.
//
// **Die Regel ist die Marke `latest`**, und sie steht nur auf einer Seite:
// ein Name ohne Doppelpunkt meint die Marke `latest`. Ein Name **mit** Marke
// bleibt, was er ist — `gemma3:270m` und `gemma3:latest` sind zwei Modelle.
func derselbeAlias(gefuehrt, gesucht string) bool {
	return gefuehrt == gesucht ||
		(!strings.Contains(gesucht, ":") && gefuehrt == gesucht+":latest")
}
