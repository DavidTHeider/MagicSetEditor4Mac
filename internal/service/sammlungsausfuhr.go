package service

// **Die Sammlung ausgeben** — Werkstück 4 von Phase 4.5
// (#1292).
//
// **Was hineingeht, muss auch wieder heraus.** Die Abnahme von #1292 nennt es
// ausdrücklich; ohne diesen Weg wäre die Sammlung eine Einbahnstraße, und wer
// sie woanders weiterführen will, tippt sie ab.
//
// # Warum hier und nicht an der Einfuhr
//
// Weil sie **kein Netz** braucht. `Listeneinfuhr` hängt an Scryfall und damit
// an einem Zuhause für den Zwischenspeicher; die Ausfuhr liest den Bestand und
// schreibt Text. Wer sie dort anhängte, machte das Ausgeben von einem Dienst
// abhängig, den es dafür nicht braucht (§16, Regel 5).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// Ausfuhrform ist die Gestalt, in der die Sammlung hinausgeht.
type Ausfuhrform string

const (
	// FormListe ist die Gestalt, die die Einfuhr versteht — **verlustbehaftet**.
	//
	// Zustand, Sprache, Erwerbsdatum, Erwerbspreis, Währung und Notiz stehen
	// nicht darin, weil das Listenformat sie nicht kennt. Wer eine Sammlung
	// **sichern** will, nimmt [FormCSV].
	FormListe Ausfuhrform = "liste"

	// FormCSV trägt jede Spalte.
	FormCSV Ausfuhrform = "csv"
)

// Sammlungsausfuhr gibt die Sammlung als Text.
type Sammlungsausfuhr struct {
	ablage Sammlungsablage

	// herkunft schreibt auf, woher eingelesene Zeilen kamen — Werkstück 12
	// von Phase 4.12 (#1351). **Nachgerüstet und darum optional**: eine
	// Ausfuhr ohne sie bricht keine Zusage, sondern macht keine (dieselbe
	// Bauform wie `Nutzerverwaltung.MitFaktor`).
	herkunft Exemplarherkunft

	// drucke ist das Register der Drucke — Werkstück A2 von Phase 5.3
	// (#1441). **Optional wie die Herkunft**: ohne Register liest die Einfuhr
	// den Alias wie vor 0.5.0, und die Ablage hängt die Zeile des Drucks beim
	// Schreiben an.
	drucke Druckregister
}

// Exemplarherkunft schreibt auf, woher ein Exemplar kam.
type Exemplarherkunft interface {
	ExemplarherkunftSetzen(ctx context.Context, exemplarID int64,
		herkunft domain.Herkunft) error
}

// NeueSammlungsausfuhr baut sie.
func NeueSammlungsausfuhr(ablage Sammlungsablage) *Sammlungsausfuhr {
	return &Sammlungsausfuhr{ablage: ablage}
}

// MitHerkunft setzt die Naht, die aufschreibt, woher Eingelesenes kam.
func (s *Sammlungsausfuhr) MitHerkunft(herkunft Exemplarherkunft) *Sammlungsausfuhr {
	s.herkunft = herkunft

	return s
}

// MitDrucken setzt das Register, über das die Einfuhr eine Referenz auflöst.
func (s *Sammlungsausfuhr) MitDrucken(drucke Druckregister) *Sammlungsausfuhr {
	s.drucke = drucke

	return s
}

// Ausgeben schreibt die ganze Sammlung in der genannten Form.
func (s *Sammlungsausfuhr) Ausgeben(ctx context.Context, form Ausfuhrform) (
	string, error,
) {
	exemplare, err := s.ablage.Liste(ctx)
	if err != nil {
		return "", err
	}

	zeilen := make([]sammelliste.Ausfuhrzeile, 0, len(exemplare))

	for _, exemplar := range exemplare {
		zeile := alsAusfuhrzeile(exemplar)

		// **Ein Druck ohne Alias nennt seine Kennung aus dem Register** — die
		// hat die Zeile nicht bei sich, und ohne sie fände der Empfänger ihn
		// nur über den Rückfall. Für jede Zeile mit Alias kostet das nichts.
		if fehler := s.kennungNachtragen(ctx, exemplar, &zeile); fehler != nil {
			return "", fehler
		}

		zeilen = append(zeilen, zeile)
	}

	switch form {
	case FormListe:
		return sammelliste.SchreibeListe(zeilen), nil

	case FormCSV:
		return sammelliste.SchreibeCSV(zeilen), nil

	default:
		return "", fmt.Errorf(
			"%w: %q ist keine Ausfuhrform — es gibt %q und %q",
			domain.ErrValidation, form, FormListe, FormCSV)
	}
}

// kennungNachtragen holt für einen Druck ohne Alias die Kennung aus dem Register.
func (s *Sammlungsausfuhr) kennungNachtragen(
	ctx context.Context, exemplar *domain.Exemplar, zeile *sammelliste.Ausfuhrzeile,
) error {
	if s.drucke == nil || exemplar.Printing != "" || exemplar.DruckID == 0 {
		return nil
	}

	druck, fehler := s.drucke.Fuer(ctx, exemplar.DruckID)
	if fehler != nil {
		return fehler
	}

	verweis := druck.AlsVerweis()
	zeile.Kennungssystem, zeile.Kennung = string(verweis.System), verweis.Wert

	return nil
}

// herkunftSchreiben notiert an jeder **neuen** Zeile, woher sie kam.
//
// **Der Bericht ist schon fertig, wenn das hier läuft**, und er wird auch bei
// einem Fehler zurückgegeben: die Exemplare sind da (§7, Regel 3 —
// „geschrieben ist geschrieben"). Was fehlt, ist die Notiz darüber, und das ist
// ein Fehler, der gemeldet und nicht verschluckt wird.
func (s *Sammlungsausfuhr) herkunftSchreiben(
	ctx context.Context, angelegt []int64, quelle string,
) error {
	if s.herkunft == nil || quelle == "" {
		return nil
	}

	for _, kennung := range angelegt {
		if fehler := s.herkunft.ExemplarherkunftSetzen(ctx, kennung,
			domain.Herkunft{Art: domain.HerkunftEinfuhr, Quelle: quelle},
		); fehler != nil {
			return fehler
		}
	}

	return nil
}

// alsAusfuhrzeile bildet ein Exemplar auf eine Ausfuhrzeile ab.
func alsAusfuhrzeile(exemplar *domain.Exemplar) sammelliste.Ausfuhrzeile {
	return sammelliste.Ausfuhrzeile{
		Printing: exemplar.Printing,
		Menge:    exemplar.Menge,
		Zustand:  string(exemplar.Zustand),
		Sprache:  exemplar.Sprache,
		Finish:   exemplar.Ausfuehrung(),
		Name:     exemplar.Name,
		Setcode:  exemplar.Setcode,
		Nummer:   exemplar.Nummer,

		Erwerbsdatum: exemplar.Erwerbsdatum,

		// **Der Preis als Text**, in der Form, in der jemand ihn tippt — im
		// Bestand ist er eine Ganzzahl in der **kleinsten Einheit seiner
		// Währung** (LOC-03.AK-02,
		// #1689).
		Erwerbspreis:    preistext(exemplar.Erwerbspreis, exemplar.Erwerbswaehrung),
		Erwerbswaehrung: exemplar.Erwerbswaehrung,
		Notiz:           exemplar.Notiz,

		// **Die Ordnung geht mit** (0057): ein CSV, das den Lagerort verlöre,
		// wäre keine Sicherung.
		Lagerort: exemplar.Lagerort,
		Wunsch:   exemplar.Wunsch,

		// **Und mit ihr Rang und Art** (T4, #1469) — aus demselben Grund:
		// wer eine Dringlichkeit von Hand gesetzt hat, bekäme sie sonst nach
		// dem Zurücklesen nicht wieder.
		Wunschrang: string(exemplar.Dringlichkeit()),
		Wunschart:  string(exemplar.Druckbindung()),

		// **Die portable Referenz** (A2, #1441, 0086): der Alias ist die
		// Scryfall-Kennung, und die Fassung sagt, wie die Spalten zu lesen sind.
		Verweisfassung: domain.Verweisfassung,
		Kennungssystem: kennungssystemFuer(exemplar.Printing),
		Kennung:        exemplar.Printing,
	}
}

// kennungssystemFuer nennt das System des Alias — keines, wo keiner ist.
func kennungssystemFuer(printing string) string {
	if printing == "" {
		return ""
	}

	return string(domain.SystemScryfall)
}

// preistext schreibt den Preis in der kanonischen Form — leer bei null.
//
// **Leer und nicht „0.00"**: null heißt „kein Preis erfasst", und eine Null im
// CSV läse sich beim Zurücklesen als geschenkte Karte.
//
// **Mit den Stellen seiner Währung** ([domain.GeldText]): 150 Yen sind
// »150« und nicht »1.50« (LOC-03.AK-02, #1689).
func preistext(betrag int64, waehrung string) string {
	if betrag == 0 {
		return ""
	}

	return domain.GeldText(betrag, waehrung)
}

// preisAus liest den Preis einer Zeile — und sagt, wenn er nicht zu lesen
// war.
//
// # Warum ein Befund und kein Fehler
//
// **Hier stand „ohne Fehler"**, und der Grund daneben war gut: „ein
// unlesbarer Preis ist ein fehlender, und die Zeile geht deshalb mit ihrer
// Karte durch. Ihn zum Befund zu machen hieße, eine ganze Sammlung an einer
// Spalte scheitern zu lassen, die niemand braucht."
//
// **Dieser Grund bleibt** — die Zeile geht weiter durch. Was nicht bleibt,
// ist das **Schweigen**: `strconv.ParseFloat` gibt für »1,23« einen Fehler,
// und daraus wurde eine Null. Wer eine deutsch formatierte Liste einlas,
// verlor **jeden** Preis, ohne ein Wort zu lesen (gemessen am 2026-09-22).
// LOC-03.AK-03 verlangt genau das Gegenteil: „mehrdeutige Eingaben werden
// erklärt und nicht still falsch interpretiert."
// **Und die Währung fällt mit** (gefunden beim ersten Lauf): ein Exemplar
// mit Währung und ohne Preis ist ungültig — [domain.Exemplar.Pruefe] sagt
// „zu einer Währung gehört ihr Preis", und die Zeile wäre damit doch
// gescheitert. Eine Währung ohne Betrag sagt ohnehin nichts.
func preisAus(text, waehrung string, stelle int) (
	betrag int64, gilt string, befund *domain.Befund,
) {
	betrag, fehler := domain.GeldAus(text, waehrung)
	if fehler == nil {
		return betrag, waehrung, nil
	}

	return 0, "", &domain.Befund{
		Ort: fmt.Sprintf("Zeile %d", stelle),
		Meldung: fehler.Error() +
			" Preis und Währung dieser Zeile bleiben leer; die Karte kommt an.",
		Art: domain.ArtEingabe,
	}
}

// erwerbGeprueft leert, was die Prüfung am Erwerb abwiese — und sagt, warum.
//
// # Lesbar ist nicht gültig
//
// **Befund G16 des Zweitblicks über Phase 6.3**
// (#1666):
// [preisAus] fing nur, was sich nicht **lesen** ließ. »-1.00«, »1.00« ohne
// Währung, »EURO« oder eine Währung ohne Betrag lasen sich — und die
// Prüfung des Exemplars verwarf danach die **ganze** Zeile, gegen Vermerk
// 0127: „Beim Einlesen einer Sammlung bleibt die Zeile trotzdem".
//
// **Und der Tag gehört dazu**: die Prüfung kam mit UX-06 (#1678) für die
// Formulare und nahm am Einlesen seither ebenso die Zeile mit. Preis,
// Währung und Tag sind die Angaben zum **Erwerb** (`domain/exemplarerwerb.go`)
// — sie beschreiben den Kauf, nicht die Karte, und ohne sie ist die Karte
// trotzdem da.
//
// **Was nicht dazugehört, bleibt, wie es war**: ein Zustand oder eine
// Ausführung, die es nicht gibt, beschreiben die Karte selbst, und die Zeile
// scheitert weiter an ihnen.
func erwerbGeprueft(exemplar *domain.Exemplar, stelle int) []domain.Befund {
	var befunde []domain.Befund

	if abgewiesen := exemplar.Preisbefunde(); len(abgewiesen) > 0 {
		exemplar.Erwerbspreis, exemplar.Erwerbswaehrung = 0, ""
		befunde = append(befunde, domain.Befund{
			Ort: fmt.Sprintf("Zeile %d", stelle),
			Meldung: abgewiesen[0].Error() +
				" Preis und Währung dieser Zeile bleiben leer; die Karte kommt an.",
			Art: domain.ArtEingabe,
		})
	}

	if abgewiesen := exemplar.Datumsbefunde(); len(abgewiesen) > 0 {
		exemplar.Erwerbsdatum = ""
		befunde = append(befunde, domain.Befund{
			Ort: fmt.Sprintf("Zeile %d", stelle),
			Meldung: abgewiesen[0].Error() +
				" Der Tag des Erwerbs bleibt leer; die Karte kommt an.",
			Art: domain.ArtEingabe,
		})
	}

	return befunde
}
