package msescript

// **Was ein Skript zu sagen hat** — Werkstück A von Phase 4.9
// (#1324).
//
// # Warum es das gibt
//
// `trace`, `warning` und `error` melden etwas, ohne die Auswertung anzuhalten:
//
//	SCRIPT_FUNCTION(warning) {
//	  SCRIPT_PARAM_C(String, input);
//	  SCRIPT_PARAM_DEFAULT_C(bool, condition, true);
//	  if (condition) { queue_message(MESSAGE_WARNING, input); }
//	  return script_nil;
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 33–40, Fassung f9d9356 vom
// 2021-07-06. **`queue_message` legt sie in eine Warteschlange**, die die
// Oberfläche leert — sie sind kein Fehler, sondern eine Auskunft.
//
// # Warum an der Umgebung
//
// **Weil sie ohnehin durch jede Auswertung reicht.** Ein zweiter Rückgabewert
// an `Auswerten`, `AuswertenQuelle` und `AuswertenTextskript` änderte zehn
// Aufrufstellen in `internal/service`, ohne mehr zu können — und ein Paket,
// das seine Meldungen in einen Paketwert legte, hätte einen Zustand über
// Auswertungen hinweg.
//
// **Der Sammler wird vererbt und nicht kopiert**: `Darueber` reicht denselben
// Zeiger weiter, damit eine Meldung aus einem Funktionsrumpf oben ankommt.
// Wer keinen anlegt — `Bibliothek()` selbst —, sammelt nichts; das ist der
// Fall »eine eingebaute Funktion von Hand gerufen« aus dem Kopf von [Ruf].

// Meldungsart sagt, wie ernst eine Meldung ist.
type Meldungsart string

const (
	// MeldungHinweis kommt von `trace`.
	MeldungHinweis Meldungsart = "hinweis"

	// MeldungWarnung kommt von `warning`.
	MeldungWarnung Meldungsart = "warnung"

	// MeldungFehler kommt von `error`.
	//
	// **Und hält trotzdem nichts an.** Im Original steht `return script_nil`
	// darunter; wer `error` ruft, sagt etwas über die Daten, nicht über das
	// Skript.
	MeldungFehler Meldungsart = "fehler"
)

// Meldung ist, was ein Skript gemeldet hat.
type Meldung struct {
	Art  Meldungsart
	Text string

	// Zeile und Spalte sind die Stelle des Aufrufs — eins-basiert.
	Zeile  int
	Spalte int
}

// Meldungen gibt, was unter dieser Umgebung gemeldet wurde.
//
// **In der Reihenfolge des Meldens.** Wer sie liest, liest den Verlauf.
func (u *Umgebung) Meldungen() []Meldung {
	for ebene := u; ebene != nil; ebene = ebene.darauf {
		if ebene.meldungen != nil {
			return *ebene.meldungen
		}
	}

	return nil
}

// melde legt eine Meldung in den Sammler — falls es einen gibt.
func (u *Umgebung) melde(meldung Meldung) {
	for ebene := u; ebene != nil; ebene = ebene.darauf {
		if ebene.meldungen != nil {
			*ebene.meldungen = append(*ebene.meldungen, meldung)

			return
		}
	}
}
