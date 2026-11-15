package aktualisierung

import (
	"bytes"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// mindestwarten ist, wie lange gewartet wird, wenn GitHub eine Grenze meldet
// und keine Zeit nennt — „wait for at least one minute before retrying"
// (nachgelesen am 2026-10-13, siehe [grenzeErschoepft]).
const mindestwarten = time.Minute

// grenzeErschoepft sagt, ob GitHub mit dieser Antwort eine Grenze meldet —
// und bis wann sie gilt.
//
// **Ein 403 und ein 429 sind bei GitHub dieselbe Auskunft** — R01
// (#2333),
// nachgelesen am 2026-10-13 in „Rate limits for the REST API", Abschnitt
// „Exceeding the rate limit": wer die erste Grenze überschreitet, bekommt
// „a `403` or `429` response", `x-ratelimit-remaining` steht auf 0, und bis
// `x-ratelimit-reset` wird nicht wiederholt. Bei einer **zweiten** Grenze
// gilt `retry-after`, sonst wieder `reset`, wenn `remaining` 0 ist, und sonst
// „wait for at least one minute". **Ohne Marke sind es 60 Anfragen je
// Stunde**, und nach Vermerk 0190 fragt bald jeder ohne Marke.
//
// **Bis zum Zweitblick über Phase 0** setzte ein 429 nur die allgemeine
// Sperre von [netz.Hole], und ein lesbares `reset` schlug `retry-after` auch
// dann, wenn die erste Grenze gar nicht erschöpft war.
//
// **Ein 403 ohne eines dieser Zeichen bleibt ein Verbot.**
func grenzeErschoepft(antwort netz.Antwort, jetzt time.Time) (time.Time, bool) {
	if antwort.Status != netz.StatusVerboten && antwort.Status != netz.StatusGedrosselt {
		return time.Time{}, false
	}

	erschoepft := antwort.Kopf.Get("X-Ratelimit-Remaining") == "0"
	if erschoepft {
		if sekunden, err := strconv.ParseInt(antwort.Kopf.Get("X-Ratelimit-Reset"), 10, 64); err == nil {
			return time.Unix(sekunden, 0), true
		}
	}

	if warte, mitWarten := netz.WarteZeit(antwort.Kopf, jetzt); mitWarten {
		return jetzt.Add(warte), true
	}

	zweiteGrenze := bytes.Contains(bytes.ToLower(antwort.Rumpf), []byte("rate limit"))
	if erschoepft || zweiteGrenze || antwort.Status == netz.StatusGedrosselt {
		return jetzt.Add(mindestwarten), true
	}

	return time.Time{}, false
}
