package app

// **Welche Anbieter es gibt und ob sie gefragt werden dürfen** — Werkstück A8
// von Phase 5.3
// (#1447).
//
// # Eine Liste und keine Kette von `if`
//
// **Weil ein dritter Anbieter sonst an drei Stellen eingetragen werden
// müsste** — A9 (OpenAI) und A10 (xAI) stehen schon im Plan. Was hier steht,
// ist je Anbieter eine Zeile: der Name, sein Schalter, seine Adresse und der
// Bauplan.
//
// # Warum der Schalter ein Dienstschalter ist
//
// **Weil `dienst.anthropic` dieselbe Frage beantwortet und es ihn schon
// gibt.** Ein eigener `ki.anthropic` daneben wäre die zweite Aufzählung, vor
// der der Kopf von `einstellung.go` warnt: zwei Schalter für dieselbe Sache,
// und einer bleibt beim nächsten Umbau stehen. Die Einstellungsseite, der
// Schlüsselbund, der Dienstkatalog und ihre Wächter hängen ohnehin an dieser
// Familie.
//
// **`ki.laufzeit` bleibt daneben stehen**, und das ist kein Widerspruch: es
// ist **kein** fremder Dienst, sondern ein Programm auf diesem Rechner. Es
// hat keine Zugangsdaten, keinen Vertrag und keine Bedingungen — was `dienst.`
// führt, träfe auf nichts davon zu.

import (
	"context"
	"maps"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// kianbieter gibt die Anbieter samt ihrem Stand aus den Einstellungen.
func (a *App) kianbieter(ctx context.Context) ([]kianbieter, error) {
	if a.einstellungen == nil {
		return nil, fehlenderBestand()
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return nil, fehler
	}

	adresse := stand[domain.KIAdresse].Wirksam
	lokal := stand[domain.KILaufzeit]
	alle := []kianbieter{{
		name:          ollamaanbieter,
		eingeschaltet: lokal.Wirksam == domain.KILaufzeitAn,
		gewaehlt:      lokal.Gewaehlt == domain.KILaufzeitAn,
		adresse:       adresse,
		bauen:         a.ollamabauplan(adresse),
	}}

	// **Apples Gerätemodell** — AI-05 (#2042), Vermerk 0166. Eingeschaltet,
	// wenn die Einstellung automatisch sagt und es für diese Sitzung nicht
	// abgeschaltet ist; **ob es auf dem Mac aktiviert ist**, sagt seine Quelle
	// selbst (`Fassung`), samt dem, was fehlt.
	alle = append(alle, a.geraeteanbieterAus(ctx, stand[domain.KIGeraetemodell]))

	// **Die Cloudanbieter stehen in einer Aufzählung** (A9, #1448; A10,
	// #1449). Was sie unterscheidet, ist der Name, der Dienst und die
	// Adresse — und genau diese drei stehen hier.
	//
	// **Ihre Adresse steht da und ist nicht leer.** Leer heißt im Fenster
	// „die Vorgabeadresse auf **diesem** Rechner" — für Ollama richtig, für
	// jeden Cloudanbieter das Gegenteil dessen, was zutrifft. Gefunden am
	// 2026-09-12 in der Sichtprüfung am laufenden Fenster; kein Test sah es,
	// weil die Go-Seite ein leeres Feld trug und leer für den anderen
	// Anbieter stimmt.
	for _, cloud := range cloudanbieter() {
		wahl := stand[domain.DienstEinstellung(cloud.dienst)]

		// **Die eingestellte Adresse schlägt die Vorgabe** — Werkstück 3 von
		// A18 (#1533).
		// Leer heißt weiterhin: die Vorgabe des Anbieters, und die steht im
		// Blatt und nicht hier.
		gewaehlteAdresse := cloud.adresse
		if eingestellt := strings.TrimSpace(
			stand[domain.DienstAdresse(cloud.dienst)].Wirksam); eingestellt != "" {
			gewaehlteAdresse = eingestellt
		}

		alle = append(alle, kianbieter{
			name:          cloud.name,
			eingeschaltet: wahl.Wirksam == domain.DienstAn,
			gewaehlt:      wahl.Gewaehlt == domain.DienstAn,
			adresse:       gewaehlteAdresse,
			bauen:         a.cloudbauplan(cloud.name, gewaehlteAdresse),
		})
	}

	return alle, nil
}

// ollamabauplan gibt den Bauplan der lokalen Laufzeit — nil, wenn keiner
// verdrahtet ist.
func (a *App) ollamabauplan(adresse string) func() KILaufzeitquelle {
	if a.kilaufzeit == nil {
		return nil
	}

	return func() KILaufzeitquelle { return a.kilaufzeit(adresse) }
}

// cloudzeile ist ein Cloudanbieter, wie die Liste ihn führt.
type cloudzeile struct {
	name    string
	dienst  domain.Dienst
	adresse string
}

// cloudanbieter gibt die Cloudanbieter samt Dienst und Adresse.
//
// **Drei Angaben je Anbieter, und keine davon ist erschlossen**: der Name
// steht im Katalog, der Dienst schaltet ihn, die Adresse kennt sein Blatt.
func cloudanbieter() []cloudzeile {
	return []cloudzeile{
		{anthropicanbieter, domain.DienstAnthropic, anthropic.Vorgabeadresse},
		{openaianbieter, domain.DienstOpenAI, openai.Vorgabeadresse},
		{xaianbieter, domain.DienstXAI, xai.Vorgabeadresse},
	}
}

// cloudbauplan gibt den Bauplan eines Cloudanbieters — nil, wenn keiner
// verdrahtet ist.
//
// **Mit der Adresse, wie bei [App.ollamabauplan]** — Werkstück 3 von A18
// (#1533). Vorher trug er keine, und damit konnte die Einstellung nirgends
// ankommen.
func (a *App) cloudbauplan(name, adresse string) func() KILaufzeitquelle {
	bauen, verdrahtet := a.kicloud[name]
	if !verdrahtet {
		return nil
	}

	return func() KILaufzeitquelle { return bauen(adresse) }
}

// MitKIAnbietern setzt die Baupläne — für Läufe und den Composition Root.
//
// **Die Cloudanbieter kommen als Karte und nicht als Stellungsparameter**
// (A9, #1448). Mit zweien ginge beides; mit dem dritten stünden drei
// namenlose Funktionen nebeneinander, und wer sie vertauscht, merkt es an
// keiner Stelle — der Typ ist derselbe.
//
// **Und sie wird übernommen statt ausgelesen** (A10, #1449). Wer sie Eintrag
// für Eintrag ausläse, müsste für jeden neuen Anbieter hierher — die Zeile,
// die dann fehlte, gäbe einen Anbieter, der in der Liste steht und nie
// gefragt wird.
func (a *App) MitKIAnbietern(
	lokal func(adresse string) KILaufzeitquelle,
	fremde map[string]func(adresse string) KILaufzeitquelle,
) *App {
	a.kilaufzeit = lokal
	a.kicloud = maps.Clone(fremde)

	return a
}

// cloudnaehte gibt je Cloudanbieter den Bauplan seiner Naht.
//
// **Eine Stelle und nicht zwei** (A10, #1449). Sie stand bis dahin zweimal
// da — im Composition Root für die Flächen ohne Fenster und in
// `kiVerdrahten` für das Fenster —, und die zweite hätte beim nächsten
// Anbieter vergessen werden können: dann stünde er am Fenster und nicht an
// MCP, REST und der Kommandozeile. Das ist genau der Fall, den §16 mit
// „dieselbe Fähigkeit, überall dieselbe Antwort" meint.
func (a *App) cloudnaehte() map[string]func(adresse string) KILaufzeitquelle {
	// **Die Naht bekommt die Adresse durchgereicht** (A18, #1533): leer
	// heißt weiterhin „die Vorgabe des Anbieters", und die steht im Blatt.
	return map[string]func(adresse string) KILaufzeitquelle{
		anthropicanbieter: a.anthropicnaht,
		openaianbieter:    a.openainaht,
		xaianbieter:       a.xainaht,
	}
}
