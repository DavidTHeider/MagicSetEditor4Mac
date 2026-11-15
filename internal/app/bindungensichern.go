package app

import "context"

// Die zwei Wege, ein Set zu sichern — ⌘S und „Sichern unter …".
//
// **In einer eigenen Datei**, seit ⌘S bei einem Set ohne Ort den zweiten Weg
// nimmt (FIX-F02): beide gehören zusammen, und `bindungen.go` stand an der
// Zeilengrenze (§2).

// SetSpeichern schreibt ein Set an den Ort, aus dem es kam — ⌘S.
//
// **Ohne Dialog, und das ist der Unterschied zu allem anderen hier.** Der Ort
// steht in der Quelle
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md));
// ein leeres Ziel heißt „dorthin zurück". Wer jedes Mal fragte, hätte kein
// Speichern, sondern ein „Speichern unter".
func (b *Bindungen) SetSpeichern(ctx context.Context, setID string) SpeicherAntwort {
	antwort := b.app(ctx).SetSpeichern(b.ruf(ctx), setID, "")

	// **Ein Set ohne Ort fragt, wohin** — wie jedes Programm auf dem Mac beim
	// ersten Sichern: ⌘S und „Sichern unter …" gehen dann denselben Weg
	// (FIX-F02, S01.NAT-02).
	if antwort.ohneOrt {
		return b.SetSpeichernUnter(ctx, setID)
	}

	return antwort
}

// SetSpeichernUnter fragt nach einem Ort und schreibt das Set dorthin.
func (b *Bindungen) SetSpeichernUnter(ctx context.Context, setID string) SpeicherAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.SetSpeichern,
		Filter:    texte.SetFilter,
		Muster:    "*.mse-set",
		Vorschlag: texte.SetVorschlag,
		Speichern: true,
	})
	if err != nil {
		return SpeicherAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	// **Ein leerer Pfad ist hier ein Abbruch, nicht die Herkunft.** Wer den
	// Dialog wegklickt, will nicht stattdessen über das Original schreiben —
	// derselbe leere Wert bedeutet an den zwei Stellen zwei Dinge, und genau
	// deshalb steht die Unterscheidung hier und nicht im Dienst.
	if pfad == "" {
		return SpeicherAntwort{Befunde: []string{}, Abgebrochen: true}
	}

	return b.app(ctx).SetSpeichern(ctx, setID, pfad)
}
