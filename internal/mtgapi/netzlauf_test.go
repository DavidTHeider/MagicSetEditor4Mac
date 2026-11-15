//go:build netzlauf || dienstdrift

// **Der Lauf gegen den echten Dienst** — Werkstück 3 von Phase 4.8
// (#1321).
//
// # Warum es ihn hier gibt und für zwei der drei Dienste nicht
//
// **Weil magicthegathering.io keinen Zugang verlangt** (nachgelesen am
// 2026-09-02: „No API key is required"). MTGJSON braucht ein Patreon-Token,
// Scrydex ein Konto samt Team — beides gibt es hier nicht, und mehr zu
// behaupten wäre der Fehler, gegen den Vermerk 0056 geschrieben ist.
//
// **Das ist der ganze Unterschied zwischen den drei Graden von Gewissheit**,
// und diese Datei ist seine sichtbare Seite: für diesen einen Dienst kann
// dieses Repositorium sagen, dass er so antwortet, wie es annimmt.
//
// # Was er prüft, das kein anderer Test kann
//
// Die Vertragstests halten, dass dieses Programm die **dokumentierte**
// Gestalt liest. Sie können nicht sagen, ob der Dienst sie noch schickt —
// genau das ist die Drift, gegen die sie gebaut sind, und gegen sich selbst
// prüft ein Beispiel nicht.
//
// # Warum er nicht in der Pipeline läuft
//
// Er befragt einen fremden Rechner: er scheitert, wenn jemand anderes ein
// Problem hat, und er verbraucht von einer Grenze, die anderen gehört (§3).
//
//	go test -tags netzlauf ./internal/mtgapi/

package mtgapi_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// echterKlient baut den Klienten gegen den echten Dienst.
func echterKlient() *mtgapi.Klient {
	return mtgapi.NeuerKlient(mtgapi.Adresse,
		netz.NeuerKlient(netz.UeblicheFristen()))
}

// **Der Dienst antwortet, und die Pflichtfelder sind da.**
//
// **»Black Lotus« ist bewusst gewählt**: die Karte gibt es seit 1993, und
// wenn sie einmal verschwände, wäre das eine Nachricht über den Dienst und
// keine über diesen Test.
func TestDerEchteDienstAntwortet(t *testing.T) {
	karten, stand, err := echterKlient().Karten(t.Context(),
		mtgapi.Abfrage("name", "Black Lotus"))
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) == 0 {
		t.Fatal("keine Karte")
	}

	eine := karten[0]

	// **Die drei Pflichtfelder** — sie sind der Vertrag.
	if eine.Kennung == "" || eine.Name == "" || eine.Setcode == "" {
		t.Errorf("Pflichtfeld leer: %+v", eine)
	}

	// **Und der Stand kommt an.** Wäre er unbekannt, hätte der Dienst die
	// beiden Kopfzeilen fallen lassen — eine Drift, die dieses Programm
	// erfahren soll.
	if !stand.Bekannt {
		t.Error("der Dienst nannte keinen Stand — Ratelimit-Kopfzeilen fehlen")
	}

	// **Die gemessene Grenze, nicht die dokumentierte.**
	//
	// `https://docs.magicthegathering.io/` sagt am 2026-09-02 „5000 requests
	// per hour"; der Dienst schickt am selben Tag `ratelimit-limit: 1000`.
	// **Dieser Lauf hat den Unterschied gefunden** — er war rot, als hier
	// noch 5.000 stand, und die Drossel war fünffach zu schnell.
	if stand.Grenze != 1000 {
		t.Errorf("die Grenze ist %d, gemessen waren am 2026-09-02 1000 — "+
			"hat der Dienst sie geändert?", stand.Grenze)
	}
}

// **Und die Zuordnung liest, was er schickt.**
//
// Sie prüft **Feldnamen**, nicht Werte: dass `manaCost` noch so heißt, dass
// `legalities` noch eine Liste aus `format` und `legality` ist.
func TestDieZuordnungLiestDenEchtenDienst(t *testing.T) {
	karten, _, err := echterKlient().Karten(t.Context(),
		mtgapi.Abfrage("name", "Lightning Bolt"))
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	var mitKosten, mitLegal int

	for _, eine := range karten {
		if eine.Manakosten != "" {
			mitKosten++
		}

		if len(eine.Legalitaeten) > 0 {
			mitLegal++
		}
	}

	if mitKosten == 0 {
		t.Error("keine Karte trug Manakosten — heißt das Feld noch manaCost?")
	}

	if mitLegal == 0 {
		t.Error("keine Karte trug Legalitäten — ist legalities noch eine Liste?")
	}
}

// **Ein unbekannter Zustand wäre ein Befund und kein Absturz.**
//
// Der Lauf sagt nichts darüber, **ob** der Dienst einen schickt — er sagt,
// dass die Zuordnung ihn heute nicht meldet. Meldet sie einen, ist das die
// Drift, für die es die Befunde gibt.
func TestDerEchteDienstSchicktKeineUnbekanntenZustaende(t *testing.T) {
	klient := echterKlient()

	karten, _, err := klient.Karten(t.Context(), mtgapi.Abfrage("name", "Counterspell"))
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) == 0 {
		t.Fatal("keine Karte")
	}
}
