package service_test

// **Ein bekannter Peer an einer neuen Adresse** — NET-02.NAT-02 und
// S14.NAT-04 (#1706).
//
// Die Suche fand ihn und sagte „steht schon in der Liste" — und die Adresse,
// unter der die Kontaktliste ihn führte, stimmte nicht mehr. Wer verbinden
// wollte, tippte die neue ab. Jetzt nennt der Fund die gemerkte Adresse mit,
// und ob sie abweicht, entscheidet der Dienst und nicht das Fenster.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinBekannterFundSagtObSichSeineAdresseGeaendertHat(t *testing.T) {
	t.Parallel()

	kontakte := einKontaktdoppel("bumgezogen", "bgeblieben", "bohne")
	kontakte.kontakte["bumgezogen"].Stand = domain.KontaktBestaetigt
	kontakte.kontakte["bumgezogen"].Adresse = "annas-mac.local:41234"
	kontakte.kontakte["bgeblieben"].Adresse = "bens-mac.local:41234"

	werk := service.NeuesNachbarwerk(
		nachbardoppel{funde: []service.Nachbarfund{
			{Kennung: "bumgezogen", Adresse: "10.0.0.8:41234", Merkadresse: "annas-neuer.local:41234"},
			{Kennung: "bgeblieben", Adresse: "10.0.0.9:41234", Merkadresse: "bens-mac.local:41234"},
			{Kennung: "bohne", Adresse: "10.0.0.10:41234", Merkadresse: "10.0.0.10:41234"},
			{Kennung: "bfremd", Adresse: "10.0.0.11:41234", Merkadresse: "10.0.0.11:41234"},
		}},
		service.NeuesKontaktwerk(kontakte, einePeeruhr))

	nachbarn, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	nach := map[peer.Kennung]service.Nachbar{}
	for _, n := range nachbarn.Nachbarn {
		nach[n.Kennung] = n
	}

	// **Umgezogen**: die gemerkte steht dabei, und sie weicht ab.
	if umgezogen := nach["bumgezogen"]; umgezogen.GemerkteAdresse != "annas-mac.local:41234" ||
		!umgezogen.AdresseGeaendert() {
		t.Errorf("umgezogen: %+v", umgezogen)
	}

	// **Geblieben**: dieselbe Merkadresse — nichts zu übernehmen.
	if geblieben := nach["bgeblieben"]; geblieben.AdresseGeaendert() {
		t.Errorf("geblieben: %+v", geblieben)
	}

	// **Ohne gemerkte Adresse** gibt es sehr wohl etwas zu übernehmen — die
	// Kontaktliste weiß nicht, wo er ist.
	if ohne := nach["bohne"]; ohne.GemerkteAdresse != "" || !ohne.AdresseGeaendert() {
		t.Errorf("ohne: %+v", ohne)
	}

	// **Unbekannt**: da gibt es keine Adresse zu übernehmen, nur aufzunehmen.
	if fremd := nach["bfremd"]; fremd.AdresseGeaendert() {
		t.Errorf("fremd: %+v", fremd)
	}
}
