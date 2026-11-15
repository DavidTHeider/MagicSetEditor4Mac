package service_test

// **Das Tischbild, das der Wirt zurückschickt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Aus `partiewirt_test.go` herausgelöst**, als sie an die Zeilengrenze aus
// §2 stieß. Der Schnitt folgt der Frage: dort steht, was der Wirt mit einem
// **Wunsch** tut, hier, was er von sich aus **zurückschickt**.
//
// # Und das ist die Hälfte, die man übersieht
//
// **Ohne Bild sähe der Gast nichts.** Er hat keinen Vollzustand und keinen
// Regelausleger; der Strom allein sagte ihm, **dass** etwas geschah, nicht,
// wie der Tisch danach aussieht — und schon gar nicht, was er selbst noch
// darf.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Tischbild an den Wirt ist eine Anmaßung.**
//
// **Die Spiegelung zum Ereignis von außen**: das Bild ist das Ergebnis seiner
// eigenen Rechnung, und wer es ihm schickte, behauptete, für ihn gerechnet zu
// haben.
func TestEinTischbildAnDenWirtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, _ := einWirt(nehmer)

	if fehler := wirt.Angekommen(context.Background(),
		eineKontrollnachricht(t, domain.KontrolleTischbild,
			domain.Partiesicht{PartieID: "p1", Fuer: "gast"}),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}

	if len(nehmer.bekommen) != 0 {
		t.Error("das Bild wurde ausgelegt")
	}
}

// **Und was das Rechnen des Bildes klagt, kommt zurück.**
//
// **Der Befehl ist trotzdem gewirkt** — die Ereignisse sind hinaus, und was
// hier scheitert, ist die Anzeige beim Gast. Der Fehler wird gemeldet und
// nicht verschluckt (§8): sonst säße der Gast vor einem Tisch, der nicht
// nachgezogen ist, und niemand wüsste warum.
func TestWasDasTischbildKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Partie ließ sich nicht lesen")

	nehmer := &nehmerdoppel{
		gibt:      []domain.Ereignis{{PartieID: "p1", Folge: 1, Art: "x"}},
		bildklage: kaputt,
	}

	wirt, naht := einWirt(nehmer)

	fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}))

	switch {
	case !errors.Is(fehler, kaputt):
		t.Fatalf("es kam %v", fehler)
	case len(naht.gesendet) != 1:
		t.Errorf("%d Nachrichten hinaus — das Ereignis ist gewirkt und geht",
			len(naht.gesendet))
	}
}

// **Was der Dienst klagt, klagt der Wirt weiter — und verkündet nichts.**
func TestWasDerDienstKlagtWirdNichtVerkuendet(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Fassung gilt nicht mehr")
	wirt, naht := einWirt(&nehmerdoppel{klagt: kaputt})

	fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}))

	switch {
	case !errors.Is(fehler, kaputt):
		t.Fatalf("es kam %v", fehler)
	case len(naht.gesendet) != 0:
		t.Error("trotz der Klage ging etwas hinaus")
	}
}

// **Und eine leere Erlaubnis geht als leere Liste hinaus, nicht als `null`.**
//
// **Sie ist eine Antwort**: „gerade geht nichts" sieht anders aus als „es
// wurde nicht gefragt". Ein `null` im Fenster ließe jede Schleife darüber
// umfallen.
func TestEineLeereErlaubnisGehtAlsListeHinaus(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{ohneErlaubnis: true}
	wirt, naht := einWirt(nehmer)

	if fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}),
	); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	letzte := ausgepackt(t, naht.gesendet[len(naht.gesendet)-1])

	if !strings.Contains(string(letzte.Inhalt), `"erlaubt":[]`) {
		t.Errorf("hinaus ging %s", letzte.Inhalt)
	}
}
