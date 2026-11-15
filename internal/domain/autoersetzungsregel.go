package domain

// **Eine Autoersetzungsregel, wie der Bestand sie führt**
// (#1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
//
// # ported: `class AutoReplace`
//
// `src/data/word_list.hpp`, Zeilen 46–59, Fassung f9d9356 vom 2021-07-06:
//
//	bool   enabled;
//	bool   whole_word;
//	bool   custom;      ///< Is this a custom auto replace?
//	String match;
//	String replace;
//
// **Fünf Angaben, und `custom` ist die, an der alles hängt.** Sie entscheidet,
// ob das Muster geändert und die Regel gelöscht werden darf
// (`AutoReplaceWindow::refreshItem`, `src/gui/auto_replace_window.cpp`, Zeilen
// 253–264).

import (
	"fmt"
	"strings"
)

// Autoersetzungsregel ist eine Regel, wie eine Fläche sie sieht.
type Autoersetzungsregel struct {
	// Muster ist der Text, auf den die Regel anspringt — **und ihr
	// Schlüssel**. Zwei Regeln mit demselben Muster sind dieselbe Regel
	// (`GameSettings::initDefaults`, `src/data/settings.cpp`, Zeile 93,
	// Fassung f9d9356 vom 2021-07-06).
	Muster string

	Ersatz string

	// GanzesWort sagt, ob nur ein ganzes Wort ersetzt wird — **ohne Angabe
	// wahr**.
	GanzesWort bool

	// Aktiv sagt, ob die Regel greift — MSEs `enabled` je Regel.
	Aktiv bool

	// Eigen sagt, dass die Regel hier angelegt wurde — MSEs `custom`.
	//
	// **Eine Regel des Pakets ist nicht eigen.** Ihr Muster gehört dem Paket:
	// es lässt sich nicht ändern, und die Regel lässt sich nicht löschen.
	// Was sich ändern lässt, ist ihr Ersatz, ihr Schalter und ihre
	// Wortgrenze — genau das, was MSE erlaubt.
	Eigen bool
}

// PruefeAutoersetzung sagt, ob eine Regel angelegt werden darf.
//
// **Ein leeres Muster ist keine Regel** — MSE schaltet sie ab:
//
//	void after_reading(AutoReplace& ar, Version) {
//	  if (ar.match.empty()) ar.enabled = false;
//	}
//
// `src/data/word_list.cpp`, Zeilen 51–55, Fassung f9d9356 vom 2021-07-06.
// Hier wird sie **abgewiesen** statt
// abgeschaltet: eine Regel, die niemand einschalten kann, ist ein Zustand
// ohne Ausweg, und `msepaket` nimmt sie aus demselben Grund gar nicht erst auf.
func PruefeAutoersetzung(muster string) error {
	if strings.TrimSpace(muster) == "" {
		return fmt.Errorf("%w: eine Regel ohne Muster greift nie", ErrValidation)
	}

	return nil
}
