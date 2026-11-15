package anmeldung_test

// **Der Rückruf eines fremden Anbieters** — Werkstück O11 von Phase 5.2
// (#1435).
//
// # Warum
//
// Der Rückruf kommt über die **Schleifenadresse** herein, und wer auf diesem
// Rechner einen Browser aufmachen kann, kann ihn schicken. Er ist damit die
// Fläche, die am wenigsten Vertrauen verdient: **ein falscher `state` ist ein
// Angriffsversuch**, kein Bedienfehler — ohne diese Prüfung nähme die
// Anwendung einen Code an, den jemand anderes untergeschoben hat, und meldete
// den Angreifer als den Nutzer an.
//
// Gefuzzt wird die **rohe Abfrage**, wie sie in der Adresszeile steht — nicht
// eine schon zerlegte `url.Values`. Genau das verlangt O11: Grenzen in den
// Rohdaten prüfen, vor der Normalisierung.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Abfrage, zwei Ergebnisse
//   - **ein angenommener Rückruf ohne den `state` dieses Laufs** — die Zusage,
//     um derentwillen es den `state` gibt
//   - **ein angenommener Rückruf ohne Code**: ohne ihn gibt es nichts
//     einzutauschen
//   - **ein angenommener Rückruf, der eine Absage trägt** — wer „Abbrechen"
//     drückt, ist nicht angemeldet
//
//	go test ./internal/anmeldung -run Fuzz -fuzz FuzzPruefeRueckruf -fuzztime 2m

import (
	"net/url"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

// rueckrufsaat sind die Abfragen, mit denen der Fuzzer beginnt.
//
// **Mit dem echten `state` in der Hälfte davon**: eine Saat, die ihn nie
// trifft, prüft nur den Weg der Abweisung — und der andere ist der, an dem
// etwas kaputtgehen kann.
var rueckrufsaat = []string{
	"",
	"code=abc",
	"state=%s",
	"state=%s&code=abc",
	"state=%s&error=access_denied",
	"state=%s&error=access_denied&error_description=nein",
	"state=%s&error=server_error&error_description=kaputt",
	"state=%s&code=",
	"state=&code=abc",
	"state=fremd&code=abc",
	"state=%s&code=abc&code=zweiter",
	"state=%s&state=fremd&code=abc",
	"state=%s&code=%00",
	"code=abc&state=%s&error=",
	"%s",
	"state=%s&code=abc&scope=repo%20user",
}

// **Kein Rückruf ohne den eigenen `state` kommt durch.**
func FuzzPruefeRueckruf(f *testing.F) {
	lauf, fehler := anmeldung.Beginne(
		anmeldung.GitHub("kennung", "geheim"), "http://127.0.0.1:7777/rueckruf")
	if fehler != nil {
		f.Fatalf("den Lauf beginnen: %v", fehler)
	}

	for _, saat := range rueckrufsaat {
		f.Add(strings.ReplaceAll(saat, "%s", url.QueryEscape(lauf.State)))
	}

	f.Fuzz(func(t *testing.T, abfrage string) {
		werte, zerlegefehler := url.ParseQuery(abfrage)
		if zerlegefehler != nil {
			// Eine Abfrage, die nicht einmal eine ist, erreicht den Lauf nie:
			// `net/http` zerlegt sie, bevor jemand sie zu sehen bekommt.
			return
		}

		ergebnis := lauf.PruefeRueckruf(werte)

		if zweites := lauf.PruefeRueckruf(werte); (ergebnis == nil) != (zweites == nil) {
			t.Fatalf("derselbe Rückruf einmal so und einmal anders: %v gegen %v",
				ergebnis, zweites)
		}

		if ergebnis != nil {
			return
		}

		if werte.Get("state") != lauf.State {
			t.Errorf("angenommen mit fremdem state %q", werte.Get("state"))
		}

		if werte.Get("code") == "" {
			t.Error("angenommen ohne Code")
		}

		if werte.Get("error") != "" {
			t.Errorf("angenommen trotz Absage %q", werte.Get("error"))
		}
	})
}
