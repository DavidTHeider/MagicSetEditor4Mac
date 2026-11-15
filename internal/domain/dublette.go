package domain

// **Derselbe Druck zweimal** — DOM-03, S04.NAT-01 bis 03
// (#1698).
//
// # Was eine Zeile der Sammlung ist
//
// **Die Eindeutigkeit im Bestand sagt es**: Profil, Druck, Zustand, Sprache,
// Ausführung und ob es ein Wunsch ist (Migration 0086) — der Druck über
// Alias **und** Zeile, sonst wären zwei vorläufige Drucke eine. Zwei Exemplare, die
// darin übereinstimmen, sind **eine** Zeile mit einer Menge — und wer das
// zweite aufnimmt, entscheidet über die Menge, statt eine zweite Zeile
// anzulegen. Eine andere Sprache oder Ausführung bleibt eine eigene Zeile
// (S04.NAT-02).
//
// **Neu** und nicht portiert: MSE2 führt keine Sammlung (Vermerk 0052); die
// Regel steht in der Eindeutigkeit dieses Bestands.

// GleicheZeile sagt, ob zwei Exemplare dieselbe Zeile der Sammlung meinen —
// **dieselbe Frage wie die Eindeutigkeit im Bestand**, ohne das Profil, das
// dort der Rahmen ist.
//
// **Ohne Ausführung ist es nonfoil** ([Exemplar.Ausfuehrung]), wie der Bestand
// es beim Schreiben ergänzt; der Druck zählt über [Exemplar.Druckschluessel],
// damit ein vorläufiger Druck nicht einem anderen vorläufigen gleicht.
func (e Exemplar) GleicheZeile(anderes Exemplar) bool {
	return e.Druckschluessel() == anderes.Druckschluessel() &&
		e.Zustand == anderes.Zustand &&
		e.Sprache == anderes.Sprache &&
		e.Ausfuehrung() == anderes.Ausfuehrung() &&
		e.Wunsch == anderes.Wunsch
}

// Dublette ist ein Exemplar, das die Sammlung schon führt — **mit der Zeile,
// die es schon gibt**.
//
// **Sie trägt Kennung und Fassung**, damit das Fenster ihre Menge erhöhen
// kann, ohne sie zu suchen, und gegen den gelesenen Stand (§7, Regel 2). Sie
// bleibt ein [ErrDoppelt] mit dem Schlüssel `doppeltesExemplar`: wer sie
// nicht kennt, sieht dieselbe Absage wie vorher.
type Dublette struct {
	Vorhanden Exemplar

	befund *FieldError
}

// NeueDublette baut sie über der Absage des Bestands.
//
// **Der Satz nennt die Zeile** — Kommandozeile, MCP und REST haben keinen
// Knopf, der zu ihr führt, und ihre Kennung ist dort der Weg.
func NeueDublette(vorhanden Exemplar, grund error) *Dublette {
	return &Dublette{
		Vorhanden: vorhanden,
		befund: NewFieldError("", SchluesselDoppeltesExemplar,
			"Diesen Druck führt die Sammlung schon als Exemplar %d mit %d Stück — im "+
				"selben Zustand, in derselben Sprache und Ausführung. Erhöhe dort die "+
				"Menge oder wähle eine andere Angabe.", vorhanden.ID, vorhanden.Menge).MitGrund(grund),
	}
}

func (d *Dublette) Error() string { return d.befund.Error() }

// Unwrap gibt den Befund — Art, Schlüssel und Grund bleiben so erreichbar.
func (d *Dublette) Unwrap() error { return d.befund }
