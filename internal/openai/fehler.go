package openai

// **Was ein Fehler von OpenAI heißt** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Die Gestalt steht fest
//
// **Nachgelesen am 2026-09-12** unter
// <https://developers.openai.com/api/docs/guides/error-codes>:
//
//	{"error":{"message":…,"type":…,"code":…,"param":…}}
//
// # Der Satz, der die Recherche wert war
//
// > „Retrying billing, spend, or quota errors won't restore API access."
//
// **Drei verschiedene Dinge tragen den Status 429**, und nur eines davon geht
// vorüber:
//
//   - `slow_down` — die Drossel. Sie trägt `Retry-After` und vergeht.
//   - `credit_balance_exhausted` — das Guthaben ist alle.
//   - `organization_spend_limit_exceeded` / `project_spend_limit_exceeded` —
//     eine gesetzte Ausgabengrenze ist erreicht.
//
// Ein blindes Wiederholen läuft bei den letzten dreien ins Leere. Sie sind
// **Erlaubnisfragen** und keine Geduldsfragen (§8).

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// dieGeldgruende sind die Schlüssel, an denen ein 429 ohne Wiederkehr
// erkennbar ist.
//
// **Eine Aufzählung und kein Präfix.** „Enthält `spend`" träfe auch einen
// künftigen Schlüssel, der etwas anderes heißt — und was dieses Programm
// nicht kennt, soll als unbekannt durchkommen und nicht als erraten.
var dieGeldgruende = []string{
	"credit_balance_exhausted",
	"organization_spend_limit_exceeded",
	"project_spend_limit_exceeded",
}

// fehlerhuelle ist die Gestalt, in der OpenAI Fehler schickt.
type fehlerhuelle struct {
	Fehler struct {
		Meldung string `json:"message"`
		Art     string `json:"type"`
		Code    string `json:"code"`

		// Feld ist `param` — der Name der Angabe, an der es lag.
		//
		// **Bei einem 400 ist er die halbe Auskunft**: „unbrauchbare
		// Anfrage" sagt nichts, „`text.format.schema`" sagt alles.
		Feld string `json:"param"`
	} `json:"error"`
}

// fehlerAus macht aus einer abgelehnten Antwort einen Fehler dieses Programms.
//
// **Die Meldung des Dienstes reist als Grund mit und nicht als Satz** (§8,
// [domain.MitGrund]): sie ist englisch und gehört ins Protokoll.
func fehlerAus(antwort netz.Antwort, weg string) error {
	var huelle fehlerhuelle

	// **Ein unlesbarer Rumpf ist kein Grund, den Status zu verschweigen.**
	_ = json.Unmarshal(antwort.Rumpf, &huelle) //nolint:errcheck // siehe oben

	grund := fmt.Errorf("%s: %s (%s/%s%s)", weg, huelle.Fehler.Meldung,
		huelle.Fehler.Art, huelle.Fehler.Code, feldhinweis(huelle))

	return domain.MitGrund(artFuer(antwort.Status, huelle), grund,
		"OpenAI hat die Anfrage abgelehnt (%d)", antwort.Status)
}

// feldhinweis nennt die beanstandete Angabe — leer, wenn keine dasteht.
func feldhinweis(huelle fehlerhuelle) string {
	if huelle.Fehler.Feld == "" {
		return ""
	}

	return ", an " + huelle.Fehler.Feld
}

// artFuer ordnet einen Status der Fehlerart dieses Programms zu.
func artFuer(status int, huelle fehlerhuelle) error {
	switch status {
	case netz.StatusNichtGefunden:
		return domain.ErrNotFound

	case netz.StatusOhneZugang, netz.StatusOhneZahlung, netz.StatusVerboten:
		// **Erlaubnis und nicht Validierung**: der Aufruf war richtig, und
		// der Zugang trägt ihn nicht. Ein 403 heißt hier zudem etwas
		// Eigenes — „unsupported country/region" —, und auch das ist keine
		// Frage der Eingabe.
		return domain.ErrErlaubnis

	case netz.StatusGedrosselt:
		return gedrosselteArt(huelle)

	case netz.StatusWiderspruch:
		return domain.ErrConflict
	}

	return domain.ErrValidation
}

// gedrosselteArt unterscheidet die Drossel vom leeren Konto.
//
// **Drei Dinge mit demselben Status.** Die Drossel geht vorüber und nennt
// `retry-after`; ein leeres Guthaben und eine erreichte Ausgabengrenze tun
// beides nicht — sie sind **Erlaubnisfragen** und keine Geduldsfragen.
func gedrosselteArt(huelle fehlerhuelle) error {
	if slices.Contains(dieGeldgruende, huelle.Fehler.Code) {
		return domain.ErrErlaubnis
	}

	return domain.ErrConflict
}
