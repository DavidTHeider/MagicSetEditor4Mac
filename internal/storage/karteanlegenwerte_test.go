package storage_test

// **Eine neue Karte mit ihren gerechneten Werten, in einem Schritt** — #2516.
// MSE rechnet beim Anlegen jeden Wert (`AddCardAction`); wer die Karte
// zurücknimmt, nimmt die Werte mit — ein ⌘Z, nicht zwei.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineNeueKarteTraegtIhreWerteInEinemSchritt(t *testing.T) {
	db, setID, _ := bestandMitDreiKarten(t)

	kennung, fehler := storage.NeuerKartenschreiber(db).KarteAnlegenMitWerten(probekontext(t), setID,
		[]domain.Value{{Field: "typ", Text: "[]"}, {Field: "echo", Text: "[]!"}})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	karte, fehler := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), kennung)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if typ, _ := karte.Value("typ"); typ != "[]" {
		t.Errorf("typ %q, erwartet []", typ)
	}

	if echo, _ := karte.Value("echo"); echo != "[]!" {
		t.Errorf("echo %q, erwartet []!", echo)
	}

	leser := storage.NeuerVerlaufsleser(db)
	if _, gab, err := leser.Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v / %v", gab, err)
	}

	if zeilen, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID); err != nil || len(zeilen) != 3 {
		t.Errorf("nach einem ⌘Z %d Karten (%v), erwartet drei", len(zeilen), err)
	}
}
