package domain

// **Bezeichnungen und Sonderzustände** — Werkstück M20 der Regelmaschine
// (#1621).
//
// # Was eine Bezeichnung ist
//
// **Etwas, das ein Spieler *hat*, ohne dass es ein Objekt wäre.** CR 725.1
// sagt es für den Monarchen, CR 726.1 für die Initiative: es gibt sie in
// einer Partie erst, wenn ein Effekt sie vergibt — und bis dahin hat sie
// niemand.
//
// **Und Tag und Nacht gehören der Partie** (CR 731.1), nicht einem Spieler.
//
// # Was hier nicht steht
//
// **Die eingebauten ausgelösten Fähigkeiten.** CR 725.2 gibt dem Monarchen
// zwei (im Endschritt ziehen, und wer ihm Kampfschaden zufügt, wird
// Monarch), CR 726.2 der Initiative drei, CR 728.1 den
// Radioaktivitätsmarken eine. **Sie haben keine Quelle** — die CR nennt das
// ausdrücklich eine Ausnahme —, und ausgelöste Fähigkeiten gibt es erst mit
// `M9` (#1611).
// Bis dahin vergibt sie ein Mensch, so wie er heute eine Karte umlegt.

import "fmt"

// Tageszeit ist, was die Partie selbst gerade ist — CR 731.1.
//
// **Drei Zustände und nicht zwei**: eine Partie fängt mit **keinem** von
// beidem an, und CR 731.2c sagt, dass dann auch nichts geprüft wird. Wer
// „Tag" zur Vorgabe machte, ließe jede Partie mit einer Tageszeit anfangen,
// die keine Karte gesetzt hat.
type Tageszeit string

const (
	// TagesZeitTag — CR 731.1.
	TageszeitTag Tageszeit = "tag"

	// TageszeitNacht — CR 731.1.
	TageszeitNacht Tageszeit = "nacht"
)

// Monarchwerden gibt einem Sitz die Bezeichnung „Monarch" — CR 725.1.
//
// **Nur einer** (CR 725.3): wer sie bekommt, nimmt sie dem, der sie hatte.
func (p Partie) Monarchwerden(wer string) (Partie, error) {
	if fehler := p.bezeichnungsitz(wer, "Monarch", "725.1"); fehler != nil {
		return p, fehler
	}

	p.Monarch = wer

	return p, nil
}

// Initiativenehmen gibt einem Sitz die Initiative — CR 726.1.
//
// **Nur einer** (CR 726.3), und zwar aus demselben Grund wie beim
// Monarchen.
func (p Partie) Initiativenehmen(wer string) (Partie, error) {
	if fehler := p.bezeichnungsitz(wer, "die Initiative", "726.1"); fehler != nil {
		return p, fehler
	}

	p.Initiative = wer

	return p, nil
}

// bezeichnungsitz prüft, ob dieser Sitz eine Bezeichnung bekommen kann.
//
// **Ein Gegangener nicht** — CR 725.4 und CR 726.4 nehmen sie ihm im
// selben Augenblick wieder ab; sie ihm erst zu geben wäre eine Handlung,
// die sich selbst zurücknimmt.
func (p Partie) bezeichnungsitz(wer, was, nummer string) error {
	if p.SitzVon(wer) == nil {
		return fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
			ErrNotFound, wer)
	}

	if p.istDraussen(wer) {
		return fmt.Errorf(
			"%w: %q ist nicht mehr im Spiel und wird nicht %s (CR %s)",
			ErrValidation, wer, was, nummer)
	}

	return nil
}

// Strahlung ändert die Radioaktivitätsmarken eines Sitzes — CR 728.1.
//
// **Am Spieler und nicht am Objekt** — CR 728.1 nennt sie eine Art Zähler,
// die ein **Spieler** haben kann, und verweist auf CR 122. Dieselbe Gestalt
// wie die Vergiftungsmarken (CR 122.1f).
//
// **Unter null gibt es keine** — dieselbe Regel wie bei jedem Zähler.
func (p Partie) Strahlung(wer string, um int) (Partie, error) {
	if p.SitzVon(wer) == nil {
		return p, fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
			ErrNotFound, wer)
	}

	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)

	for stelle := range sitze {
		if sitze[stelle].Kennung != wer {
			continue
		}

		sitze[stelle].Strahlung = max(0, sitze[stelle].Strahlung+um)
	}

	p.Sitze = sitze

	return p, nil
}

// bezeichnungenNachrichten gibt Monarch und Initiative weiter, wenn ihr
// Träger geht — CR 725.4, CR 726.4.
//
// **Der aktive Spieler bekommt sie**, und geht der selbst, der Nächste in
// Zugreihenfolge. Beide Regeln sagen dasselbe, und deshalb steht es einmal
// hier.
func (p Partie) bezeichnungenNachruecken() Partie {
	p.Monarch = p.naechsterTraeger(p.Monarch)
	p.Initiative = p.naechsterTraeger(p.Initiative)

	return p
}

// naechsterTraeger sagt, wer eine Bezeichnung übernimmt — CR 725.4.
//
// **Leer bleibt leer**: eine Bezeichnung, die niemand hatte, bekommt auch
// niemand. CR 725.1 sagt es ausdrücklich — es gibt keinen Monarchen, bis
// ein Effekt einen macht.
func (p Partie) naechsterTraeger(bisher string) string {
	if bisher == "" || !p.istDraussen(bisher) {
		return bisher
	}

	// **Der aktive Spieler**, es sei denn, er ist selbst gegangen — dann
	// der Nächste in Zugreihenfolge, und [Partie.naechsterSitz]
	// überspringt die Gegangenen ohnehin (CR 800.4).
	if p.Zug.Aktiv != "" && !p.istDraussen(p.Zug.Aktiv) {
		return p.Zug.Aktiv
	}

	naechster := p.naechsterSitz(p.Zug.Aktiv)
	if p.istDraussen(naechster) {
		return ""
	}

	return naechster
}

// naechsteTageszeit gibt, was die Partie nach diesem Zug ist — CR 731.2.
//
// # Warum der zweite Teil des Enttappschritts
//
// **CR 502.2 nennt die Stelle**: als Zweites im Enttappen, und geprüft wird
// der **vorige** Zug. Hier steht er beim Zugwechsel, und das ist dieselbe
// Stelle: der Enttappschritt ist der erste des neuen Zuges, und die Zahl des
// vorigen steht in diesem Augenblick noch da.
//
// **Und ohne Tageszeit findet er nicht statt** — CR 731.2c sagt es
// ausdrücklich: weder Tag noch Nacht bleibt weder Tag noch Nacht, und keine
// Karte hat je etwas anderes gesagt.
func (p Partie) naechsteTageszeit() Tageszeit {
	switch p.Tageszeit {
	// **CR 731.2a** — Tag wird Nacht, wenn der vorige aktive Spieler
	// keinen Zauber gewirkt hat.
	case TageszeitTag:
		if p.Zug.Zauber == 0 {
			return TageszeitNacht
		}

	// **CR 731.2b** — Nacht wird Tag, wenn er **zwei oder mehr** gewirkt
	// hat. Bei einem bleibt es Nacht.
	case TageszeitNacht:
		if p.Zug.Zauber >= 2 {
			return TageszeitTag
		}
	}

	return p.Tageszeit
}
