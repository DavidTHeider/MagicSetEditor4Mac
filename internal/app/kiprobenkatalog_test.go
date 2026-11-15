package app_test

// **Ein gemessenes Modell steht danach im Katalog — auch bei OpenAI** —
// #1541.
//
// # Der Befund
//
// **Bei OpenAI kam es nie an.** `domain.MitProben` ergänzt nur, was schon in
// der Liste steht, und `openai.Klient.Modelle` gibt **immer eine leere
// Liste** zurück — jedes Modell wird zu einem Befund, denn „eine Modellliste
// ist keine belegte Fähigkeit" ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)).
// Beides für sich richtig; zusammen heißt es, dass eine bezahlte Messung
// folgenlos bleibt.
//
// # Warum der Lauf daneben es nicht sah
//
// **Weil sein Doppel großzügiger ist als der Betrieb** (§3).
// `TestEinGemessenesModellStehtDanachImKatalog` gibt das Modell aus
// `Modelle()` zurück — der echte Klient tut das nicht. Der Name des Laufs
// verspricht genau das, was im Betrieb nicht galt (§10).
//
// **Hier antwortet die Quelle wie der echte**: keine Modelle, ein Befund je
// Modell.

import (
	"context"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schweigsamequelle antwortet wie OpenAI: **keine** Modelle, ein Befund je
// Modell.
type schweigsamequelle struct {
	*laufzeitdoppel

	nennt []string
}

func (s *schweigsamequelle) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	befunde := make([]domain.Befund, 0, len(s.nennt))
	for _, name := range s.nennt {
		befunde = append(befunde, domain.Befund{
			Datei: name, Art: domain.ArtUmgebung,
			Meldung: "nennt keine Fähigkeiten — auch nicht für " + name,
		})
	}

	return []domain.KIModell{}, befunde, nil
}

// **Eine Messung bringt das Modell in den Katalog.**
func TestEineMessungBringtDasModellInDenKatalog(t *testing.T) {
	t.Parallel()

	const alias = "gpt-image-langsam"

	quelle := &schweigsamequelle{
		laufzeitdoppel: &laufzeitdoppel{
			fassung: "Responses-API /v1/responses",
			antwort: app.KIRohantwort{
				Inhalt: `{"probe": "bereit"}`, Modell: alias,
			},
		},
		nennt: []string{alias},
	}

	welt := eineOpenAIWelt(t, quelle)

	// **Vorher steht von OpenAI nichts da** — der Steuerlauf, ohne den die
	// Zeile darunter nichts sagte (§10).
	//
	// **Nur von OpenAI**: die Welt trägt daneben ein lokales Modell, und ein
	// Blick auf den ganzen Katalog prüfte das falsche.
	if vorher := vonOpenAI(welt.anwendung.KIModelle(probekontext(t))); len(vorher) != 0 {
		t.Fatalf("vor der Messung steht schon %+v im Katalog", vorher)
	}

	if antwort := welt.anwendung.KIModellPruefen(probekontext(t), "openai",
		alias, ""); antwort.Fehler != nil {
		t.Fatalf("KIModellPruefen: %+v", antwort.Fehler)
	}

	katalog := welt.anwendung.KIModelle(probekontext(t))

	seine := vonOpenAI(katalog)

	stelle := slices.IndexFunc(seine,
		func(modell app.KIModellDTO) bool { return modell.Alias == alias })
	if stelle < 0 {
		t.Fatalf("nach der Messung steht %q nicht im Katalog: %+v",
			alias, katalog.Modelle)
	}

	pruefeGemessenesModell(t, katalog, seine[stelle], alias)
}

// vonOpenAI gibt die Modelle dieses einen Anbieters.
func vonOpenAI(katalog app.KIKatalogantwort) []app.KIModellDTO {
	seine := make([]app.KIModellDTO, 0, len(katalog.Modelle))

	for _, modell := range katalog.Modelle {
		if modell.Anbieter == "openai" {
			seine = append(seine, modell)
		}
	}

	return seine
}

// pruefeGemessenesModell sieht an, was am ergänzten Modell steht.
func pruefeGemessenesModell(
	t *testing.T, katalog app.KIKatalogantwort, modell app.KIModellDTO,
	alias string,
) {
	t.Helper()

	// **Der Ort kommt vom Anbieter und nicht aus der Messung.** Ein
	// Cloudanbieter rechnet in der Cloud, und unter „nur lokal" darf dieses
	// Modell nicht plötzlich als lokal gelten.
	if modell.Ort != string(domain.KIOrtCloud) {
		t.Errorf("der Ort ist %q, erwartet %q", modell.Ort, domain.KIOrtCloud)
	}

	if len(modell.Faehigkeiten) == 0 {
		t.Error("das ergänzte Modell trägt keine Fähigkeit — dann ist es " +
			"im Katalog und doch zu nichts zu gebrauchen")
	}

	// **Und es steht nicht mehr unter »messbar«** — dort gehört hin, worüber
	// **nichts** feststeht.
	for _, einer := range katalog.Anbieter {
		if einer.Name != "openai" {
			continue
		}

		if slices.Contains(einer.Messbar, alias) {
			t.Errorf("%q steht noch unter messbar: %v", alias, einer.Messbar)
		}

		if einer.Einrichtung != app.KIEinrichtungBereit {
			t.Errorf("die Einrichtung ist %q, erwartet %q",
				einer.Einrichtung, app.KIEinrichtungBereit)
		}
	}
}
