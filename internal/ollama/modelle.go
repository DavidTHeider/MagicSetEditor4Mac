package ollama

// **Die installierten Modelle** (A7, #1446).
//
// # Zwei Rufe je Modell, und beide sind nötig
//
// `/api/tags` nennt, **was liegt** — samt `digest`, und der ist die Revision
// des ausgeführten Artefakts. `/api/show` nennt, **was es kann** und unter
// welcher Lizenz. Wer nur den ersten Weg ginge, hätte Modelle ohne Fähigkeiten;
// wer nur den zweiten ginge, wüsste nicht, welche es gibt.
//
// # Die Fähigkeiten werden übersetzt und nicht erfunden
//
// Ollama meldet `completion`, `vision`, `embedding`, `tools`, `thinking` und
// weiteres. Dieses Programm kennt **sieben** ([domain.KIFaehigkeit]); was sich
// nicht zuordnen lässt, wird übergangen — eine erfundene Zuordnung wäre eine
// Fähigkeit, die niemand belegt hat (A6, 0097).

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// modellDTO ist ein Eintrag aus `/api/tags`.
//
// **Nachgelesen am 2026-09-11** unter <https://docs.ollama.com/api/tags.md>.
type modellDTO struct {
	Name    string `json:"name"`
	Digest  string `json:"digest"`
	Groesse int64  `json:"size"`
	Angaben struct {
		Format        string `json:"format"`
		Familie       string `json:"family"`
		Groesse       string `json:"parameter_size"`
		Quantisierung string `json:"quantization_level"`
	} `json:"details"`
}

// schauDTO ist die Antwort von `/api/show`.
type schauDTO struct {
	Lizenz       string   `json:"license"`
	Faehigkeiten []string `json:"capabilities"`

	// Parameter und Angaben tragen das Kontextfenster (AI-01.AK-02).
	Parameter string                     `json:"parameters"`
	Angaben   map[string]json.RawMessage `json:"model_info"`
}

// Modelle gibt, was auf dieser Laufzeit installiert ist.
//
// **Was keinen Digest hat, kommt als Befund und nicht als Modell** — das Issue
// verlangt es: „unbekannte Herkunft wird abgewiesen".
func (k *Klient) Modelle(ctx context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	fassung, fehler := k.Fassung(ctx)
	if fehler != nil {
		return nil, nil, fehler
	}

	var antwort struct {
		Modelle []modellDTO `json:"models"`
	}

	if fehler := k.hole(ctx, Modellweg, nil, &antwort); fehler != nil {
		return nil, nil, fehler
	}

	var (
		modelle []domain.KIModell
		befunde []domain.Befund
	)

	for stelle, dto := range antwort.Modelle {
		if dto.Digest == "" {
			befunde = append(befunde, domain.Befund{
				Datei: dto.Name, Ort: stelleIn(stelle),
				Art: domain.ArtUmgebung,
				Meldung: "Ollama nennt zu " + dto.Name + " keinen Digest. " +
					"Ohne ihn ist nicht zu sagen, welche Datei dort liegt — " +
					"und eine Fähigkeit ohne Artefakt hängt an nichts.",
			})

			continue
		}

		modell, befund, schaufehler := k.alsModell(ctx, dto, fassung)
		if schaufehler != nil {
			return nil, nil, schaufehler
		}

		if befund != nil {
			befunde = append(befunde, *befund)
		}

		modelle = append(modelle, modell)
	}

	return modelle, befunde, nil
}

// alsModell holt die Fähigkeiten und baut das Modell.
func (k *Klient) alsModell(ctx context.Context, dto modellDTO, fassung string) (
	domain.KIModell, *domain.Befund, error,
) {
	var schau schauDTO

	if fehler := k.hole(ctx, Schauweg,
		map[string]any{"model": dto.Name}, &schau); fehler != nil {
		return domain.KIModell{}, nil, fehler
	}

	laufzeit := Anbieter + " " + fassung
	tag := k.jetzt().UTC().Format("2006-01-02")
	fenster, kontextbeleg := kontextAus(schau.Parameter, schau.Angaben)

	return domain.KIModell{
		Alias: dto.Name, Anbieter: Anbieter, Laufzeit: laufzeit,
		Ort:      ortVon(dto.Name),
		Geprueft: belegeAus(schau.Faehigkeiten, laufzeit, tag, dto.Name),
		Herkunft: herkunftAus(dto, schau.Lizenz),

		Kontextfenster: fenster, KontextBeleg: kontextbeleg,
	}, einbetterbefund(dto.Name, schau.Faehigkeiten), nil
}

// ortVon sagt, wo ein Modell rechnet.
//
// **Am Suffix und nicht an der Adresse**: Cloudmodelle liegen auf demselben
// Port wie die lokalen, und Ollama lagert sie an seinen eigenen Dienst aus.
// Die Adresse unterscheidet sie **nicht** — das ist genau der Satz „Loopback
// allein beweist keine lokale Inferenz" in seiner konkretesten Form.
func ortVon(alias string) domain.KIOrt {
	if strings.HasSuffix(alias, Cloudsuffix) {
		return domain.KIOrtCloud
	}

	return domain.KIOrtLokal
}

// dieFaehigkeiten ordnet Ollamas Wörter denen dieses Programms zu.
//
// **Eine Aufzählung und keine Ableitung**: `tools` und `thinking` haben hier
// keine Entsprechung, und ihnen eine zu geben hieße, eine Fähigkeit zu
// behaupten, die niemand belegt hat.
//
// # Warum `embedding` **nicht** darin steht
//
// **Weil eine gemeldete Fähigkeit noch kein Vertrag ist.** A6 verlangt von
// einem Einbettungsmodell vier Angaben ([domain.KIEinbettungsvertrag]), und
// `/api/tags` nennt keine davon. Wer sie hier eintrüge, hätte die Fähigkeit
// aus einer Liste abgeleitet — genau die Namensheuristik, die das Issue
// ausschließt.
//
// **Sie kommt über [Klient.Vermessen]**, und der fährt einen echten Lauf:
// Zusammenfassung und Zerteiler aus den Modellangaben, Dimension und
// Normalisierung gemessen. Bis dahin steht der Einbetter im Katalog — **mit
// null Fähigkeiten und einem Befund**, der den Weg nennt.
func dieFaehigkeiten() map[string]domain.KIFaehigkeit {
	return map[string]domain.KIFaehigkeit{
		derTextkoenner: domain.KITextgenerierung,
		"vision":       domain.KIVision,
	}
}

// derEinbetter ist Ollamas Wort für die Fähigkeit, die hier nicht trägt.
const derEinbetter = "embedding"

// derTextkoenner ist Ollamas Wort für die Texterzeugung.
const derTextkoenner = "completion"

// belegeAus macht aus Ollamas Wörtern belegte Fähigkeiten.
//
// **Der Beleg nennt Ollama, die Fassung und den Tag** — nicht „geprüft".
// Was hier steht, ist die **Erklärung der Laufzeit über ein installiertes
// Artefakt**: mehr als eine Modellliste eines Anbieters, weniger als ein
// eigener Lauf (A6, 0097).
func belegeAus(gemeldet []string, laufzeit, tag, alias string) []domain.KIGeprueft {
	var belege []domain.KIGeprueft

	zuordnung := dieFaehigkeiten()

	for _, wort := range gemeldet {
		faehigkeit, bekannt := zuordnung[wort]
		if !bekannt {
			continue
		}

		belege = append(belege, domain.KIGeprueft{
			Faehigkeit: faehigkeit, Datum: tag,
			Beleg: "Ollama meldet " + wort + " unter " + Schauweg +
				" für dieses Artefakt (" + laufzeit + ")",
		})

		belege = append(belege, formbelege(wort, laufzeit, tag, alias)...)
	}

	return belege
}

// formbelege gibt die strukturierte Ausgabe, wo sie trägt.
//
// # Warum sie nicht in der Zuordnung steht
//
// **Weil sie keine Meldung der Laufzeit ist, sondern ein Schluss aus zweien.**
// Ollama meldet `completion`; dass die Antwort in eine **Form** gezwungen
// werden kann, steht in der Dokumentation und nicht in `capabilities`.
// **Nachgelesen am 2026-09-11** unter
// <https://docs.ollama.com/capabilities/structured-outputs>: der Parameter
// `format` nimmt ein JSON-Schema, und die Laufzeit hält es.
//
// **Und für Cloudmodelle gilt er nicht** — ebenda: „Ollama's Cloud currently
// does not support structured outputs." Eine Fähigkeit, die dieses Programm
// dort behauptete, führte zu einer Antwort ohne Form; sie fiele erst an der
// Schemaprüfung auf, und dann sähe es aus wie ein Fehler des Modells.
func formbelege(wort, laufzeit, tag, alias string) []domain.KIGeprueft {
	if wort != derTextkoenner || ortVon(alias) != domain.KIOrtLokal {
		return nil
	}

	return []domain.KIGeprueft{{
		Faehigkeit: domain.KIStrukturierteGenerierung, Datum: tag,
		Beleg: "Ollama meldet " + wort + " für dieses Artefakt (" + laufzeit +
			") und hält über den Parameter »format« ein JSON-Schema — " +
			"dokumentiert für die lokale Laufzeit, ausdrücklich nicht für " +
			"ihre Cloud",
	}}
}

// einbetterbefund sagt, warum ein Einbetter hier noch nichts kann.
//
// **Sichtbar und nicht still**: wer ein Einbettungsmodell installiert hat und
// es im Katalog ohne Fähigkeit sieht, soll den Grund lesen können — und den
// Weg dorthin.
func einbetterbefund(alias string, gemeldet []string) *domain.Befund {
	for _, wort := range gemeldet {
		if wort != derEinbetter {
			continue
		}

		return &domain.Befund{
			Datei: alias, Art: domain.ArtUmgebung,
			Meldung: "Ollama meldet für " + alias + " die Einbettung. Eine " +
				"Meldung ist noch kein Vertrag: A6 verlangt Zusammenfassung, " +
				"Zerteiler, Dimension und Normalisierung. Die beiden letzten " +
				"werden gemessen, die beiden ersten aus den Modellangaben " +
				"gelesen — beides tut Vermessen, und erst danach steht die " +
				"Fähigkeit fest.",
		}
	}

	return nil
}

// herkunftAus baut die Kette aus dem, was Ollama sagt.
//
// **Ein Glied und nicht vier**: Ollama nennt das **ausgeführte Artefakt** und
// nicht, woraus es gebaut wurde. Eine Kette, die Basis und Adapter erfände,
// wäre genau die Herkunftsangabe, die A6 verbietet — und eine leere wäre zu
// wenig: den Digest gibt es.
func herkunftAus(dto modellDTO, lizenz string) domain.KIHerkunft {
	if lizenz == "" {
		lizenz = "von Ollama nicht genannt"
	}

	return domain.KIHerkunft{Glieder: []domain.KIGlied{{
		Art: domain.KIArtefakt, Kennung: "ausgefuehrt",
		Quelle:   Anbieter + ":" + dto.Name,
		Revision: dto.Digest,
		Dateien: []domain.KIDateihash{{
			Name: dto.Name, Hash: dto.Digest,
		}},
		Lizenz: lizenz,
	}}}
}

// stelleIn benennt einen Eintrag der Modellliste über seine Stelle.
//
// # Warum das dasteht
//
// **Weil ein Eintrag ohne `name` sonst nicht zu benennen wäre** — Werkstück
// A19 von Phase 5.3
// (#1550),
// gefunden vom Fuzzer bei der Saat `{"models":[{}]}`. Zwei solche Einträge
// gaben zwei **wortgleiche** Befunde, und §8 verlangt für jeden Fehler die
// Stelle. Derselbe Fund stand zugleich in `anthropic` und `xai` — es ist eine
// Klasse und kein Fall (§10).
func stelleIn(stelle int) string {
	return "models[" + strconv.Itoa(stelle) + "]"
}
