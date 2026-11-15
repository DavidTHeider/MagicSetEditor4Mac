package service_test

// **Wer fragt, bestimmt nicht den Schlüssel seiner Referenzen** — Befund 5
// des Zweitblicks über Phase 8 (#2329).
//
// Die Referenzen einer Sicht sind ein HMAC über die Zeilenkennung des
// Bestands (`domain.Sitzungsreferenz`). Bis hierher war ihr Schlüssel die
// Vorgangskennung der Anfrage — und die schreibt der Fragende selbst. Wer sie
// kannte, rechnete fortlaufende Zeilenkennungen nach und ordnete sie seinen
// Referenzen zu; wer sie wiederholte, bekam über zwei Anfragen hinweg
// dieselben Namen.

import (
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// frageMitVorgang schickt eine Sichtanfrage mit einer Vorgangskennung, die
// der Fragende wählt, und gibt die Sitzung, mit der Bert die Sicht baute.
func frageMitVorgang(t *testing.T, anna, bert peer.Identitaet, vorgang string) string {
	t.Helper()

	hier, dort := leitung.Schleife(service.Handelskanal)
	ctx := t.Context()

	var gefragt []string
	post := bertsPost(anna, bert, sichtgeberprobe{zeilen: freigegebeneKarten(), gefragt: &gefragt})

	anfrage := domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtSichtanfrage,
		NachrichtID: "n-" + vorgang, Vorgang: vorgang, Folge: 1,
		Absender: string(anna.Kennung()), Gesendet: time.Now(),
	}
	anfrage.Unterschrift = hex.EncodeToString(anna.Signiere(anfrage.Unterschriftsform()))

	roh, fehler := json.Marshal(service.Handelsnachricht{Umschlag: anfrage})
	if fehler != nil {
		t.Fatal(fehler)
	}

	beantwortet := make(chan error, 1)
	go func() {
		_, fehler := post.EmpfangeUeber(ctx, dort, string(anna.Kennung()))
		beantwortet <- fehler
	}()

	if fehler := hier.Sende(ctx, roh); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	if _, fehler := hier.Empfange(ctx); fehler != nil {
		t.Fatalf("Empfange: %v", fehler)
	}

	if fehler := <-beantwortet; fehler != nil {
		t.Fatalf("Beantworten: %v", fehler)
	}

	if len(gefragt) != 3 {
		t.Fatalf("gefragt: %v", gefragt)
	}

	return gefragt[1]
}

func TestDerFragendeBestimmtDenSchluesselSeinerReferenzenNicht(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)

	erste := frageMitVorgang(t, anna, bert, "sicht/gewaehlt")
	zweite := frageMitVorgang(t, anna, bert, "sicht/gewaehlt")

	if erste == "sicht/gewaehlt" || erste == "" {
		t.Errorf("der Schlüssel ist die Kennung des Fragenden: %q", erste)
	}

	if erste == zweite {
		t.Errorf("zwei Anfragen mit derselben Kennung bekamen denselben Schlüssel: %q", erste)
	}
}
