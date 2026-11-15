package service

// **Ein CSV zurück in den Bestand** — Werkstück 4 von Phase 4.5
// (#1292),
// mit Vorschau und Zusammenführung seit DOM-03.AK-01
// (#1698).
//
// **Eine eigene Datei**, weil `sammlungsausfuhr.go` mit der Vorschau über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, was
// **hinausgeht**, hier, was **hereinkommt**.

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// Einfuegen liest eine CSV-Sammlung in den Bestand — Vorhandenes wird
// addiert, wie seit #1292.
//
// # Warum das hier steht und nicht bei der Listeneinfuhr
//
// **Weil ein CSV seine Drucke schon kennt.** Es trägt die Kennung je Zeile;
// es ist nichts zu deuten und nichts zu entscheiden — und damit braucht es
// weder Scryfall noch den zweiten Schritt aus §16 Regel 6.
//
// **Das ist der Rückweg der eigenen Ausfuhr**, nicht der Weg für eine fremde
// Liste. Wer eine Liste mit Namen hat, nimmt `Listeneinfuhr`.
func (s *Sammlungsausfuhr) Einfuegen(ctx context.Context, text, quelle string) (
	Einfuhrbericht, error,
) {
	return s.EinfuegenMit(ctx, text, quelle, ZusammenAddieren)
}

// EinfuegenMit liest ein CSV ein und führt Vorhandenes so zusammen, wie
// gewählt.
//
// # Und die Quelle sagt, woher das CSV kam
//
// **Sie ist eine Angabe des Aufrufers und keine des Bestands** — ein Dateiname
// am Fenster, ein Pfad an der Kommandozeile, eine Bezeichnung über den Dienst.
// **Leer heißt: niemand hat es gesagt**, und dann wird auch nichts behauptet
// ([0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md): eine falsche
// Nennung ist schlimmer als keine).
//
// **Aufgeschrieben wird sie nur an den Zeilen, die neu entstehen.** Eine
// **erhöhte** Zeile war schon vorher da; ihr eine Herkunft zu geben behauptete,
// der ganze Stapel käme von dort.
func (s *Sammlungsausfuhr) EinfuegenMit(
	ctx context.Context, text, quelle string, modus Zusammenfuehrung,
) (Einfuhrbericht, error) {
	geprueft, modusfehler := modus.geprueft()
	if modusfehler != nil {
		return leererBericht(), modusfehler
	}

	zeilen, befunde := sammelliste.LiesCSV(text)

	if len(zeilen) == 0 {
		return Einfuhrbericht{Angelegt: []int64{}, Befunde: befunde}, nil
	}

	einfuhr := &Listeneinfuhr{sammlung: s.ablage}
	bericht := Einfuhrbericht{Angelegt: []int64{}, Befunde: []domain.Befund{}}

	for stelle, zeile := range zeilen {
		// **`stelle+2`**: die Kopfzeile ist Zeile 1.
		exemplar, _, zeilenbefunde, fehler := s.csvExemplar(ctx, zeile, stelle+2, true)
		bericht.Befunde = append(bericht.Befunde, zeilenbefunde...)

		if fehler != nil {
			bericht.Befunde = append(befunde, bericht.Befunde...)

			return bericht, fehler
		}

		if exemplar == nil {
			continue
		}

		fehler = einfuhr.einExemplar(ctx, exemplar, &bericht, geprueft)
		if fehler == nil {
			continue
		}

		// **Ein Systemfehler bricht ab, ein Eingabefehler nicht** — derselbe
		// Schnitt wie bei [Listeneinfuhr.Einlesen], und aus demselben Grund.
		if !errors.Is(fehler, domain.ErrValidation) &&
			!errors.Is(fehler, domain.ErrConflict) {
			bericht.Befunde = append(befunde, bericht.Befunde...)

			return bericht, fehler
		}

		bericht.Befunde = append(bericht.Befunde, domain.Befund{
			Ort: fmt.Sprintf("Zeile %d", stelle+2), Meldung: fehler.Error(), Art: domain.ArtEingabe,
		})
	}

	// **Die Befunde des Zerteilers gehen mit** (§8): eine Zeile, die schon das
	// CSV nicht hergab, ist ein Befund über die Datei und keiner über den
	// Bestand.
	bericht.Befunde = append(befunde, bericht.Befunde...)

	if fehler := s.herkunftSchreiben(ctx, bericht.Angelegt, quelle); fehler != nil {
		return bericht, fehler
	}

	return bericht, nil
}

// Vorschauen sagt je Zeile eines CSV, was das Einlesen täte — **ohne zu
// schreiben**, auch keinen Druck ins Register.
//
// **Eine Referenz, die der Bestand nicht kennt, ist ein neuer Posten** — ob
// zwei solche Zeilen denselben Druck meinen, entscheidet erst das Anlegen,
// und die Vorschau behauptet es nicht.
func (s *Sammlungsausfuhr) Vorschauen(ctx context.Context, text string) (Sammelvorschau, error) {
	zeilen, befunde := sammelliste.LiesCSV(text)

	vorschau := Sammelvorschau{Posten: []Vorschauposten{}, Befunde: append([]domain.Befund{}, befunde...)}
	einfuhr := &Listeneinfuhr{sammlung: s.ablage}

	for stelle, zeile := range zeilen {
		exemplar, zuordnung, zeilenbefunde, fehler := s.csvExemplar(ctx, zeile, stelle+2, false)
		if fehler != nil {
			return Sammelvorschau{}, fehler
		}

		vorschau.Befunde = append(vorschau.Befunde, zeilenbefunde...)

		switch {
		case exemplar == nil:
			continue
		case zuordnung.Neu:
			exemplar.Normalisieren()

			vorschau.Posten = append(vorschau.Posten, Vorschauposten{
				Zeile: stelle + 2, Exemplar: *exemplar, Art: EinfuhrNeu, Nachher: exemplar.Menge,
			})

			continue
		}

		if fehler := vorbereiten(exemplar); fehler != nil {
			vorschau.Befunde = append(vorschau.Befunde, domain.Befund{
				Ort: fmt.Sprintf("Zeile %d", stelle+2), Meldung: fehler.Error(), Art: domain.ArtEingabe,
			})

			continue
		}

		if fehler := einfuhr.einordnen(ctx, &vorschau, stelle+2, exemplar); fehler != nil {
			return Sammelvorschau{}, fehler
		}
	}

	return vorschau, nil
}

// csvExemplar macht aus einer Zeile das Exemplar, das sie anlegte — oder
// sagt, warum nicht.
//
// **Das volle Exemplar und keine [Listenzeile]** (#1292): ein CSV trägt
// Erwerbsdatum, Preis, Währung und Notiz, und der Rundlauf soll sie
// zurückbringen. **`anlegen` sagt, ob ein unbekannter Druck ins Register
// darf** — beim Einlesen ja, in der Vorschau nicht.
func (s *Sammlungsausfuhr) csvExemplar(
	ctx context.Context, zeile sammelliste.Ausfuhrzeile, nummer int, anlegen bool,
) (*domain.Exemplar, Druckzuordnung, []domain.Befund, error) {
	zuordnung, zuordnungsfehler := s.druckZuordnenFalls(ctx, zeile, anlegen)
	if zuordnungsfehler != nil {
		// **Eine Referenz, die die Ablage abweist, ist ein Befund der Zeile;
		// ein klemmendes Register ist ein Systemfehler.**
		if !errors.Is(zuordnungsfehler, domain.ErrValidation) &&
			!errors.Is(zuordnungsfehler, domain.ErrConflict) {
			return nil, Druckzuordnung{}, nil, zuordnungsfehler
		}

		zuordnung = Druckzuordnung{Meldung: zuordnungsfehler.Error()}
	}

	ort := fmt.Sprintf("Zeile %d", nummer)

	if zuordnung.Meldung != "" {
		return nil, zuordnung, []domain.Befund{{Ort: ort, Meldung: zuordnung.Meldung, Art: domain.ArtEingabe}}, nil
	}

	// **Der Preis wird gelesen, bevor das Exemplar entsteht** — ein Befund
	// gehört zu seiner Zeile, und die Zeile geht trotzdem durch.
	var befunde []domain.Befund

	preis, preiswaehrung, preisbefund := preisAus(zeile.Erwerbspreis, zeile.Erwerbswaehrung, nummer)
	if preisbefund != nil {
		befunde = append(befunde, *preisbefund)
	}

	exemplar := &domain.Exemplar{
		Printing: zeile.Printing, DruckID: zuordnung.DruckID, Menge: zeile.Menge,
		Zustand: domain.Zustand(zeile.Zustand), Sprache: zeile.Sprache,
		Finish: zeile.Finish, Name: zeile.Name,
		Setcode: zeile.Setcode, Nummer: zeile.Nummer,
		Erwerbsdatum:    zeile.Erwerbsdatum,
		Erwerbspreis:    preis,
		Erwerbswaehrung: preiswaehrung,
		Notiz:           zeile.Notiz,
		Lagerort:        zeile.Lagerort,
		Wunsch:          zeile.Wunsch,
		Wunschrang:      domain.Wunschrang(zeile.Wunschrang),
		Wunschart:       domain.Wunschart(zeile.Wunschart),
	}

	// **Der Alias kommt vom getroffenen Druck**, wenn die Zeile keinen trug.
	if exemplar.Printing == "" {
		exemplar.Printing = zuordnung.Alias
	}

	return exemplar, zuordnung, append(befunde, erwerbGeprueft(exemplar, nummer)...), nil
}

// druckZuordnenFalls löst die Referenz auf — wenn es ein Register gibt.
//
// **Ohne Register bleibt die Zeile beim Alias**: das ist der Weg von vor
// 0.5.0, und die Ablage hängt die Druckzeile beim Schreiben an.
func (s *Sammlungsausfuhr) druckZuordnenFalls(
	ctx context.Context, zeile sammelliste.Ausfuhrzeile, anlegen bool,
) (Druckzuordnung, error) {
	if s.drucke == nil {
		return Druckzuordnung{}, nil
	}

	return s.druckZuordnen(ctx, zeile, anlegen)
}
