// **Die Nähte, mit denen die Tests einen Server bauen** — abgetrennt am
// 2026-09-02 (#1391).
package mcp_test

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// probenaehte sind die Nähte, mit denen die Tests einen Server bauen.
//
// **Eine Stelle und nicht zwei.** `sprich` und `werkzeugMitNamen` hielten
// dieselbe Aufzählung, und beim Symboleditor (#1391) ist genau das passiert,
// wovor die Doppelung warnt: die eine bekam die neue Naht, die andere nicht —
// und der Wächter fiel mit einem Nullzeiger statt mit einem Satz.
//
// **Wer eine Naht ergänzt, ergänzt sie hier.** Fehlt sie, ruft ein Werkzeug
// auf `nil`, und das ist kein Fehler, den ein Test lesen kann.
func probenaehte(bestand *bestandsprobe) mcp.Naehte {
	return naehteMitSymbolprobe(bestand, &symbolprobe{})
}

// naehteMitSymbolprobe baut dieselben Nähte mit einer eigenen Symbolprobe.
//
// **Damit ein Lauf sie scheitern lassen kann.** Die Fehlerwege der
// Symbolwerkzeuge sind je zwei Zeilen, und ohne eine Probe, die nein sagt,
// betritt sie kein Lauf (§3).
func naehteMitSymbolprobe(bestand *bestandsprobe, symbole *symbolprobe) mcp.Naehte {
	return mcp.Naehte{
		// **Die Nutzernaht ist belegt** — Werkstück 17 von Phase 4.12
		// (#1400).
		//
		// **Sie stand hier nicht**, und das war folgenlos, solange kein
		// Nutzerwerkzeug Idempotenz zusagte: der Wächter darüber ruft nur, was
		// sie zusagt. `lesen.werarbeitet` tut es — und der erste Lauf endete
		// mit einem Absturz an einer Naht, die niemand gesetzt hatte.
		//
		// **Das ist §3, „eine Vorrichtung, die großzügiger ist als der
		// Betrieb"**, von der anderen Seite: sie war **kärger**, und niemand
		// sah es, weil kein Lauf hinlangte.
		Nutzer:      &nutzerprobe{},
		Leser:       bestand,
		Schreiber:   bestand,
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
		Sammlung:    neueSammlungsprobe(),
		Preise:      neuePreisprobe(),

		// **Die Paketpflege gehört in die Vorrichtung** (#1291) — sonst sähen
		// die Wächter über der Werkzeugliste ihre vier Werkzeuge nicht, und
		// Idempotenzfall, Zusage und README blieben ungeprüft. Genau die
		// Gestalt aus §10: ein Wächter, der die halbe Zusage hält.
		Pflege:       &pflegeprobe{},
		Kartenbilder: &pflegeprobe{},
		Vorlagen:     &pflegeprobe{},

		// **Und der Symboleditor** (#1391), aus demselben Grund.
		Symbolwerkstatt: symbole,

		// **Die Symbolbibliothek gehört in die Vorrichtung** (DOM-02, #1697)
		// — sonst sähe der Wächter über die Zusagen ihre sechs schreibenden
		// Werkzeuge nie.
		Symbolbibliothek: &bibliotheksprobe{},

		// **Und die Zugriffstoken** (#1351, Werkstück 10) — ebenfalls aus
		// demselben Grund: ohne sie liefe der Idempotenzlauf gegen `nil`.
		Token: neueTokenprobe(),

		// **Und die Profilvorgaben** (#1351, Werkstück 15), aus demselben
		// Grund.
		Vorgaben: neueVorgabenprobe(),

		// **Und die Freigaben** (#1351, Werkstück 11), aus demselben Grund.
		Freigaben: neueFreigabenprobe(),

		// **Und die Deckübernahme** (#1351, Werkstück 12), aus demselben
		// Grund.
		Deckuebernahme: neueUebernahmeprobe(),

		// **Und die zwei von T8** (#1473), aus demselben Grund. Ohne sie
		// stünden ihre Werkzeuge nicht in der Liste, und kein Wächter
		// darüber sähe sie — Idempotenzfall, Zusage und README blieben
		// ungeprüft.
		Links:       &linkprobe{},
		Konto:       &kontoprobe{},
		Handel:      &handelsprobe{},
		Deckeinfuhr: deckeinfuhrprobe{},
	}
}
