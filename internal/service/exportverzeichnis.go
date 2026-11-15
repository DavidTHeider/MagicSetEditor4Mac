package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`create directory:` entscheidet, ob eine Vorlage schreiben darf** —
// #732.
//
//	void guard_export_info(const String& fun, bool need_template = false) {
//	  if (!export_info() && (!need_template || export_info()->export_template)) {
//	    throw ScriptError(_("Can only use ") + fun + _(" from export templates"));
//	  } else if (export_info()->directory_relative.empty()) {
//	    throw ScriptError(_("Can only use ") + fun +
//	                      _(" when 'create directory' is set to true"));
//	  }
//	}
//
// `src/script/functions/export.cpp`, Zeilen 26–33, Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum die Namen trotzdem dastehen
//
// Weil sie bei MSE dastehen. `init_script_export_functions` trägt sie in
// **jeden** Kontext ein; die Wache sitzt im Rumpf, nicht an der Bindung. Der
// Unterschied ist die Meldung: „der Name `copy_file` ist hier nicht bekannt"
// schickt jemanden auf die Suche nach einer fehlenden Bibliothek — „nur mit
// `create directory: true`" nennt die Zeile, die zu ändern ist.
//
// **Anders als bei den Fähigkeiten** (#216, #657, #658): dort *gibt* es die
// Funktion wirklich nicht, weil der Aufrufer keinen Ort für Dateien genannt
// hat. Hier hat die **Vorlage** gesagt, dass sie keinen will.

// verweigern legt Namen ab, die beim Aufruf ihren Grund nennen.
func verweigern(umgebung *msescript.Umgebung, namen ...string) {
	for _, name := range namen {
		umgebung.Setze(name, msescript.NeueEingebaute(nurMitVerzeichnis(name)))
	}
}

// nurMitVerzeichnis baut die Funktion, die den Grund nennt.
func nurMitVerzeichnis(name string) msescript.Eingebaut {
	return func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		return msescript.Nichts, fmt.Errorf("%w: %s", domain.ErrValidation,
			ruf.FehlerAn(name+" geht nur, wenn die Vorlage "+
				"`create directory: true` sagt").Meldung)
	}
}
