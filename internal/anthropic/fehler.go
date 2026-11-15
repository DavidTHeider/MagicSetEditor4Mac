package anthropic

// **Was ein Fehler von Anthropic heißt** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Die Gestalt steht fest
//
// **Nachgelesen am 2026-09-11** unter
// <https://platform.claude.com/docs/en/api/errors>:
//
//	{"type":"error","error":{"type":…,"message":…},"request_id":"req_…"}
//
// Elf Statuscodes mit je einem `error.type`. Die Zuordnung unten übersetzt
// sie auf die drei Fehlerarten dieses Programms (§8) — und auf `ErrErlaubnis`,
// wo der Dienst sagt, dass jemand darf, aber nicht zahlt.
//
// # Der Satz, der die Recherche wert war
//
// > „A tier spend-cap 429 has no `retry-after` header and keeps failing until
// > access resumes."
//
// **Ein blindes Wiederholen auf 429 läuft dort ins Leere** — Minuten lang,
// und dann ist immer noch Monatsanfang nötig. Erkannt wird es an
// `error.details.error_code == "enforced_spend_limit_reached"`.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// dieAusgabensperre ist der Schlüssel, an dem ein 429 ohne Wiederkehr
// erkennbar ist.
const dieAusgabensperre = "enforced_spend_limit_reached"

// dieEigeneGrenze ist der Anfang der Meldung, mit der eine **selbst
// gesetzte** Ausgabengrenze antwortet.
//
// **Nachgelesen am 2026-09-12**: sie kommt als **400** mit
// `invalid_request_error` und nicht als 429 — „The message begins `You have
// reached your specified API usage limits`, or `You have reached your
// specified workspace API usage limits` for a workspace limit". Beide Formen
// teilen diesen Anfang, und deshalb steht hier er und nicht zwei Sätze.
const dieEigeneGrenze = "You have reached your specified"

// fehlerhülle ist die Gestalt, in der Anthropic Fehler schickt.
type fehlerhuelle struct {
	Fehler struct {
		Art      string `json:"type"`
		Meldung  string `json:"message"`
		Naeheres struct {
			Schluessel string `json:"error_code"`
		} `json:"details"`
	} `json:"error"`

	// Vorgang ist die `request_id` — sie steht auch als Kopfzeile da.
	Vorgang string `json:"request_id"`
}

// fehlerAus macht aus einer abgelehnten Antwort einen Fehler dieses Programms.
//
// **Die Meldung des Dienstes reist als Grund mit und nicht als Satz** (§8,
// [domain.MitGrund]): sie ist englisch, nennt Kontingente und Fristen und
// gehört ins Protokoll — nicht in ein Fenster.
func fehlerAus(antwort netz.Antwort, weg string) error {
	var huelle fehlerhuelle

	// **Ein unlesbarer Rumpf ist kein Grund, den Status zu verschweigen.**
	_ = json.Unmarshal(antwort.Rumpf, &huelle) //nolint:errcheck // siehe oben

	grund := fmt.Errorf("%s: %s (%s, %s)", weg, huelle.Fehler.Meldung,
		huelle.Fehler.Art, vorgangAus(antwort, huelle))

	return domain.MitGrund(artFuer(antwort.Status, huelle), grund,
		"Anthropic hat die Anfrage abgelehnt (%d)", antwort.Status)
}

// vorgangAus gibt die Vorgangskennung — aus dem Rumpf oder der Kopfzeile.
//
// **Beide Wege, weil beide vorkommen**: der Fehlerrumpf trägt `request_id`,
// **jede** Antwort trägt die Kopfzeile `request-id`. Wer nur eine läse,
// stünde bei einem unlesbaren Rumpf ohne die Kennung da, die der Support
// verlangt.
func vorgangAus(antwort netz.Antwort, huelle fehlerhuelle) string {
	if huelle.Vorgang != "" {
		return huelle.Vorgang
	}

	if aus := antwort.Kopf.Get("request-id"); aus != "" {
		return aus
	}

	return "ohne Vorgangskennung"
}

// artFuer ordnet einen Status der Fehlerart dieses Programms zu.
func artFuer(status int, huelle fehlerhuelle) error {
	switch status {
	case netz.StatusNichtGefunden:
		return domain.ErrNotFound

	case netz.StatusOhneZugang, netz.StatusOhneZahlung, netz.StatusVerboten:
		// **Erlaubnis und nicht Validierung**: der Aufruf war richtig, und
		// der Zugang trägt ihn nicht. Wer das als Eingabefehler zeigte,
		// schickte jemanden auf die Suche nach einem Tippfehler.
		return domain.ErrErlaubnis

	case netz.StatusGedrosselt:
		return gedrosselteArt(huelle)

	case netz.StatusWiderspruch:
		return domain.ErrConflict

	case netz.StatusUnbrauchbar:
		return unbrauchbareArt(huelle)
	}

	return domain.ErrValidation
}

// unbrauchbareArt unterscheidet die eigene Ausgabengrenze vom Eingabefehler.
//
// **Derselbe Status und derselbe `error.type` wie ein Tippfehler im Schema**,
// und eine völlig andere Abhilfe: nicht „berichtige die Anfrage", sondern
// „hebe die Grenze an oder warte auf den Monatsanfang". Erkannt wird sie am
// Anfang der Meldung, denn ein `error_code` steht hier nicht dabei.
func unbrauchbareArt(huelle fehlerhuelle) error {
	if strings.HasPrefix(huelle.Fehler.Meldung, dieEigeneGrenze) {
		return domain.ErrErlaubnis
	}

	return domain.ErrValidation
}

// gedrosselteArt unterscheidet die Drossel von der Ausgabensperre.
//
// **Zwei Dinge mit demselben Status und demselben `error.type`.** Die Drossel
// geht vorüber und nennt `retry-after`; die Sperre am Monatsdeckel tut beides
// nicht — sie ist eine **Erlaubnisfrage** und keine Geduldsfrage.
func gedrosselteArt(huelle fehlerhuelle) error {
	if huelle.Fehler.Naeheres.Schluessel == dieAusgabensperre {
		return domain.ErrErlaubnis
	}

	return domain.ErrConflict
}
