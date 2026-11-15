package app_test

// **Auch die Texte aus Go sprechen vertraut an** — Dialoge und Menüleiste,
// Vermerk 0196 (AP-11, #2368). Den Katalogen der Oberfläche hält
// `frontend/src/sprache/anrede.test.ts` dieselbe Regel; hier sind es die,
// die macOS zeichnet: Titel der Dateidialoge, Diagnosetexte, Menüeinträge.
//
// # Was er nicht hält
//
// Eine förmliche Verbform ohne Marke. Gefangen werden die Pronomen und die
// Imperative, die in diesen Texten vorkommen.

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

var foermlich = map[string]*regexp.Regexp{
	"fr": regexp.MustCompile(`(?i)\b(vous|votre|vos|réinitialisez|choisissez|ouvrez|lisez|vérifiez|sélectionnez)\b`),
	"ru": regexp.MustCompile(`(?i)(^|[^а-яё])(вы|вас|вам|вами|ваш[а-яё]*|выберите|верните|откройте|прочитайте|проверьте|нажмите)([^а-яё]|$)`),
	"zh": regexp.MustCompile(`您`),
	"es": regexp.MustCompile(`\b(usted|ustedes|Elija|Seleccione|Abra|Lea|Restablézcalo|Compruebe)\b`),
	"de": regexp.MustCompile(`[a-zäöüß,] (Sie|Ihnen|Ihre?[mnrs]?)\b`),
}

// zeichenketten gibt jedes Textfeld eines Katalogs.
func zeichenketten(katalog any) map[string]string {
	felder := map[string]string{}
	wert := reflect.ValueOf(katalog)

	for stelle := range wert.NumField() {
		if feld := wert.Field(stelle); feld.Kind() == reflect.String {
			felder[wert.Type().Field(stelle).Name] = feld.String()
		}
	}

	return felder
}

func TestDieTexteAusGoSprechenVertrautAn(t *testing.T) {
	t.Parallel()

	for sprache, muster := range foermlich {
		for katalog, texte := range map[string]any{
			"Dialogtexte": app.DialogtexteFuer(sprache),
			"Menuetexte":  app.MenuetexteFuer(sprache),
		} {
			for name, text := range zeichenketten(texte) {
				if treffer := muster.FindString(text); treffer != "" {
					t.Errorf("%s/%s.%s: „%s“ in „%s“", sprache, katalog, name, treffer, text)
				}
			}
		}
	}
}
