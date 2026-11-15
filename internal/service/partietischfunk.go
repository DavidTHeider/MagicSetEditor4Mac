package service

// **Der Tisch hat Funk** — Werkstück aus
// #1627.
//
// # Warum es das gibt
//
// **Weil ein Zug bis zum 2026-09-19 nur den Sitz erreichte, der ihn gestellt
// hat.** [Partiewirt] schickt über **seine** Sitzung, und ein Wirt gehört zu
// **einem** Gastsitz; gerufen wird er nur, wenn dieser Gast gewünscht hat.
// Zog der Wirt selbst oder ein anderer Gast, ging nichts hinaus — null
// Ereignisse, null Tischbilder, und das andere Fenster zeigte den Tisch von
// vorhin.
//
// # Warum je Sitz gerechnet wird und nicht einmal für alle
//
// **Weil jeder etwas anderes sehen darf.** Ein Tischbild ist sichtbezogen
// ([Befehlsnehmer.Partie] nimmt den Sitz), und ein Ereignis trägt private
// Ergänzungen, die einem einzelnen gehören ([domain.Ergaenzung]). Dieselben
// Bytes an alle wären zwei Lücken in einer: ein falsches Bild und ein
// verratenes Geheimnis.
//
// **Und genau das stand offen.** [Partiewirt.verkuenden] hat das **rohe**
// Ereignis verpackt — samt der Ergänzungen aller Sitze. Die Projektion
// [domain.Ereignis.Fuer] gab es längst, `internal/app` benutzte sie, der Weg
// über die Leitung nicht.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Tischfunk trägt an den ganzen Tisch, was ein Befehl bewirkt hat.
//
// **Er steht neben dem Wirt und nicht in ihm**, weil er von **zwei** Wegen
// gerufen wird: vom Wirt, wenn ein Gast gestellt hat, und von der Anwendung,
// wenn dieser Rechner selbst gestellt hat. Ein Funk im Wirt erreichte den
// zweiten Fall nie — und genau der ist der häufigere.
type Tischfunk struct {
	partie string
	nehmer Befehlsnehmer
	runde  *Partierunde

	// fenster ist das Fenster **dieses** Rechners, fenstersitz sein Platz.
	//
	// **Beide oder keiner.** Ein Fenster ohne Sitz bekäme ein Ereignis, das
	// für niemanden projiziert ist; ein Sitz ohne Fenster hätte niemanden,
	// dem er etwas sagt.
	fenster     Ereignisnehmer
	fenstersitz string
}

// NeuerTischfunk baut ihn für **eine** Partie.
func NeuerTischfunk(
	partie string, nehmer Befehlsnehmer, runde *Partierunde,
) *Tischfunk {
	return &Tischfunk{partie: partie, nehmer: nehmer, runde: runde}
}

// MitFenster sagt, welcher Sitz an diesem Rechner sitzt und wohin er hört.
func (f *Tischfunk) MitFenster(
	sitz string, nehmer Ereignisnehmer,
) *Tischfunk {
	f.fenster = nehmer
	f.fenstersitz = sitz

	return f
}

// Verkuende schickt jedem Sitz außer `ausser` seine eigene Sicht.
//
// **`ausser` ist der Sitz, der gestellt hat.** Er bekommt das Ergebnis über
// den Weg, über den er gestellt hat — als Antwort am Fenster oder über seinen
// eigenen Wirt. Ein zweites Mal wäre nicht falsch, aber doppelt: dasselbe
// Ereignis mit derselben Folge zweimal, und das Fenster müsste unterscheiden.
//
// **Ein klagender Platz hält die anderen nicht auf.** Dieselbe Gestalt wie in
// [Partierunde.anJeden]: wer am Tisch sitzt, soll nicht davon abhängen, ob
// ein anderer noch eine Leitung hat.
func (f *Tischfunk) Verkuende(
	ctx context.Context, ausser string, ausgang Befehlsausgang,
) error {
	var klagen []error

	for _, sitz := range f.sitze() {
		if sitz == ausser {
			continue
		}

		if fehler := f.anEinenSitz(ctx, sitz, ausgang); fehler != nil {
			klagen = append(klagen, fmt.Errorf("an %q: %w", sitz, fehler))
		}
	}

	if fehler := f.ansFenster(ctx, ausser, ausgang); fehler != nil {
		klagen = append(klagen, fehler)
	}

	return errors.Join(klagen...)
}

// sitze sind die besetzten Sitze — **keine, wenn kein Tisch offen ist**.
//
// **Ein Wirt ohne offene Gastsitzung führt die Partie trotzdem**, und dann
// gibt es keine Runde. Derselbe Satz steht bei [Partiewirt.runde]; hier
// steht er noch einmal, weil ein `nil` an dieser Stelle einen `panic` gäbe
// und keinen Fehler.
func (f *Tischfunk) sitze() []string {
	if f.runde == nil {
		return nil
	}

	return f.runde.Sitze()
}

// anEinenSitz schickt einem Gast seine Ereignisse und sein Bild.
func (f *Tischfunk) anEinenSitz(
	ctx context.Context, sitz string, ausgang Befehlsausgang,
) error {
	for _, ereignis := range ausgang.Ereignisse {
		inhalt, fehler := json.Marshal(ereignis.Fuer(sitz))
		if fehler != nil {
			return fmt.Errorf("das Ereignis ließ sich nicht verpacken: %w",
				fehler)
		}

		if fehler := f.runde.Schicke(
			ctx, sitz, domain.ArtEreignis, inhalt); fehler != nil {
			return fehler
		}
	}

	// **Mit dem Bestand des Gastes** (P9, #2377): die Zeilen des Wirts
	// fänden drüben andere Karten.
	bild, fehler := tischbildFuer(
		MitBestandssicht(ctx, f.runde.Peer(sitz)), f.nehmer, f.partie, sitz)
	if fehler != nil {
		return fehler
	}

	// **Kein Fehlerzweig, und das ist dieselbe Aussage wie bei
	// [Partiewirt.tischbildSchicken]**: die Art ist eine Konstante, und ein
	// Tischbild trägt nur Zeichenketten, Zahlen und Wahrheitswerte.
	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // der Grund steht darüber
		domain.KontrolleTischbild, bild)

	return f.runde.Schicke(ctx, sitz, domain.ArtKontrolle, inhalt)
}

// ansFenster gibt dem Fenster dieses Rechners, was es sehen darf.
func (f *Tischfunk) ansFenster(
	ctx context.Context, ausser string, ausgang Befehlsausgang,
) error {
	if f.fenster == nil || f.fenstersitz == ausser {
		return nil
	}

	for _, ereignis := range ausgang.Ereignisse {
		if fehler := f.fenster.Ereignis(
			ctx, ereignis.Fuer(f.fenstersitz)); fehler != nil {
			return fehler
		}
	}

	bild, fehler := tischbildFuer(ctx, f.nehmer, f.partie, f.fenstersitz)
	if fehler != nil {
		return fehler
	}

	return f.fenster.Tischbild(ctx, bild)
}

// funk baut den Träger an den Tisch für einen Wirt.
//
// **Je Ruf und nicht gemerkt**: er hält keinen Zustand, und die Runde ist
// dieselbe. Ein gemerkter Funk müsste bei [Partiewirt.MitRunde] nachgezogen
// werden — eine zweite Stelle für dieselbe Angabe (§10).
//
// **Und er steht hier und nicht in `partiewirt.go`**, weil jene Datei an der
// Zeilengrenze steht (§2).
func (w *Partiewirt) funk() *Tischfunk {
	funk := NeuerTischfunk(w.partie, w.nehmer, w.runde)

	if w.fenster != nil {
		funk = funk.MitFenster(w.fenstersitz, w.fenster)
	}

	return funk
}

// tischbildFuer rechnet, was **ein** Sitz sehen darf.
//
// **Eine Stelle und nicht zwei.** Der Wirt hat es für seinen Gast gerechnet,
// der Funk rechnet es für jeden anderen — zwei Stellen liefen beim ersten
// neuen Feld auseinander, und genau das steht schon im Kopf von
// [Partiewirt.tischbild].
//
// **Aus einem Stand** (#2481): Sicht, Erlaubnis und Handlungen rechnet der
// Nehmer aus derselben gelesenen Fassung.
func tischbildFuer(
	ctx context.Context, nehmer Befehlsnehmer, partie, fuer string,
) (domain.Tischbild, error) {
	bild, fehler := nehmer.Tischbild(ctx, partie, fuer)
	if fehler != nil {
		return domain.Tischbild{}, fehler
	}

	if bild.Erlaubt == nil {
		// **Eine leere Liste und kein `nil`** — sie ist eine Antwort und
		// sieht anders aus als keine. Hier und nicht nur beim Nehmer: hier
		// geht sie hinaus.
		bild.Erlaubt = []string{}
	}

	return bild, nil
}
