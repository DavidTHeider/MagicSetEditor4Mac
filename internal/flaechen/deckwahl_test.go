package flaechen_test

// **Die Decks eines anderen Profils zur Wahl — über beide Flächen** —
// UX-08.AK-04 (#1680).
//
// Der fremde Mensch legt ein Deck an und gibt mir sein Profil frei; danach
// nennen Kommandozeile und Dienst dasselbe Deck. §16: dieselbe Fähigkeit,
// überall dieselbe Antwort. **Die Ablehnung ohne Freigabe** hält
// `internal/service/deckuebernahmewahl_test.go` — dieselbe Prüfung wie bei
// der Übernahme, und deren Ablehnung fährt `herkunft_test.go` über die Zeile.

import (
	"net/http"
	"strings"
	"testing"
)

// derDeckwahlweg ist der Endpunkt, wie er im Katalog steht.
const derDeckwahlweg = "GET /profile/{profilID}/decks"

func TestDieDecksVonDruebenSindUeberBeideFlaechenDieselben(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilAnDerZeile(t, umgebung)

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	deck := deckkennung(t,
		mussKommandozeile(t, umgebung, "deck", "neu", "Nachtlauf", "probe", "--json"))

	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "mitarbeiter")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", meines)

	if aus := mussKommandozeile(t, umgebung, "decks-uebernehmbar",
		fremdes); !strings.Contains(aus, deck+"\tNachtlauf") {
		t.Errorf("die Kommandozeile nennt das Deck nicht: %s", aus)
	}

	weg := strings.ReplaceAll(pfadAus(derDeckwahlweg), "{profilID}", fremdes)

	if aus := mussREST(t, an, http.MethodGet, weg,
		""); !strings.Contains(aus, `"name":"Nachtlauf"`) {
		t.Errorf("der Dienst nennt das Deck nicht: %s", aus)
	}
}
