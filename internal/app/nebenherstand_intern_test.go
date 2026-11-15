package app

// **Was die Anwendung nebenher schreibt, sieht das Fenster** — gefunden mit
// zwei Fenstern (Hüllenprobe `nat04handel`, S12.NAT-04,
// #1672):
// Bert bot einen Tausch an, er kam bei Anna über die Leitung an — und ihr
// Fenster zeigte ihn erst, als sie die Seite neu öffnete.
//
// # Warum das Fenster es nicht bemerkte
//
// **Weil es nur fremde Prozesse bemerkt.** `Datenstand` liest
// `PRAGMA data_version`, und das steigt nur, wenn eine **andere** Verbindung
// bestätigt (#1485). Was der Horchposten annimmt, schreibt derselbe Prozess
// über dieselbe Verbindung — für das Fenster geschah nichts.

import (
	"errors"
	"testing"
	"time"
)

func TestEineHandelsnachrichtBewegtDenDatenstand(t *testing.T) {
	t.Parallel()

	anwendung, sitzung := eineHandelsanwendung(t)
	anwendung.MitStand(festerStand(7))

	vorher := anwendung.Datenstand(t.Context())

	sitzung.gibt = [][]byte{
		eineHandelsnachricht(t, sitzung.stand.Gegenueber, eigeneKennungVon(t, anwendung)),
	}
	sitzung.lesefehler = errors.New("die Leitung ist zu")

	anwendung.sitzungAufnehmen(anwendung.imProfil(t.Context()), sitzung)

	if nachher := anwendung.Datenstand(t.Context()); nachher.Stand == vorher.Stand {
		t.Errorf("der Handel kam an, und der Datenstand blieb %d", nachher.Stand)
	}
}

// **Eine Sitzung, die nichts brachte, bewegt nichts** — sonst lüde das
// Fenster nach jedem Anklopfen. Dass jemand zuhört, belegt der Lauf darüber
// (§3).
func TestEineLeereHandelssitzungLaesstDenDatenstand(t *testing.T) {
	t.Parallel()

	anwendung, sitzung := eineHandelsanwendung(t)
	anwendung.MitStand(festerStand(7))

	vorher := anwendung.Datenstand(t.Context())

	sitzung.lesefehler = errors.New("die Leitung ist zu")

	anwendung.sitzungAufnehmen(anwendung.imProfil(t.Context()), sitzung)

	if nachher := anwendung.Datenstand(t.Context()); nachher.Stand != vorher.Stand {
		t.Errorf("nichts kam an, und der Datenstand ging von %d auf %d",
			vorher.Stand, nachher.Stand)
	}
}

// **Dieselbe Klasse an der Einlösung** — wer eine Einladung einlöst, steht
// danach in Annas Kontakten, geschrieben von ihrem Horchposten.
func TestEineEinloesungBewegtDenDatenstand(t *testing.T) {
	t.Parallel()

	anna, bert := eineEinladungsseite(t), eineEinladungsseite(t)
	anna.MitStand(festerStand(7))

	ctx := riegelkontext(t)
	vorher := anna.Datenstand(ctx)

	annaHorcht(t, anna)

	ausgestellt := anna.KontakteinladungErstellen(ctx, "Anna", "handel")
	if ausgestellt.Fehler != nil {
		t.Fatalf("Erstellen: %+v", ausgestellt.Fehler)
	}

	angenommen := bert.KontakteinladungAnnehmen(
		ctx, ausgestellt.Einladung.Code, "", "", false, "Bert")
	if angenommen.Gemeldet == nil || angenommen.Gemeldet.Stand != "gemeldet" {
		t.Fatalf("Annehmen: %+v", angenommen)
	}

	// **Auf die Wirkung gewartet** (§3): Bert hat seine Antwort, bevor Annas
	// Posten die Sitzung schließt.
	frist := time.Now().Add(5 * time.Second)
	for anna.Datenstand(ctx).Stand == vorher.Stand {
		if time.Now().After(frist) {
			t.Fatalf("Bert steht in Annas Kontakten, und der Datenstand blieb %d",
				vorher.Stand)
		}

		time.Sleep(10 * time.Millisecond)
	}
}
