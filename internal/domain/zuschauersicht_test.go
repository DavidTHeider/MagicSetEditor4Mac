// **Ein Zuschauer sieht nichts, was ihn verriete** — PLAY-05.AK-02
// (#1713).
//
// # Der Anlass
//
// **Das Artboard `TischSicht` nennt den Grund**, und er ist feiner als
// „verstecke die Hand": *„Beides ergibt dasselbe Bild — ein Rücken. Die
// Anwendung unterscheidet sie trotzdem, weil sonst ein Zuschauerexport
// verriete, was er nicht darf."* Eine **öffentliche Tatsache** (die Karte
// liegt verdeckt, alle wissen das) und eine **Sichtgrenze** (du darfst nicht
// hinsehen) sind zwei Dinge; die Projektion trennt sie seit G1.
//
// # Warum dieser Lauf anders fragt als die daneben
//
// **`partiesicht_test.go` prüft Felder**: dass `Karte` leer ist, dass
// `Ordnung` null ist. Das ist richtig und **veraltet beim nächsten Feld** —
// wer `Sichtobjekt` um eine Eigenschaft erweitert, die die Karte verrät,
// macht keinen dieser Läufe rot.
//
// **Dieser hier fragt nach der Eigenschaft selbst**: zwei Partien, die sich
// **nur** in den verborgenen Karten unterscheiden, sehen für einen
// Zuschauer **gleich** aus. Wer ein Feld hinzufügt, das die Karte
// durchreicht, macht die beiden Projektionen verschieden — und diesen Lauf
// rot, ohne dass ihn jemand dafür ändern müsste.
//
// # Was er nicht prüft
//
// **Die Marke `Geheim` selbst.** Sie ist in beiden Partien gleich; wer sie
// nicht mehr setzte, verriete damit keine Karte, sondern beschriftete nur
// falsch. Die Gegenprobe dazu bleibt folgerichtig grün — **das ist kein
// Loch, sondern eine andere Zusage**, und sie steht in
// `partiesicht_test.go`, das Felder prüft.
//
// Drei Gegenproben greifen: ohne `gezeigt.Karte = ""`, ohne
// `gezeigt.Ordnung = 0` und ohne `gezeigt.Name = ""` wird dieser Lauf rot.
//
// **Dass eine Fläche die Sicht auch anbietet.** Bis zum 2026-09-21 ließ sich
// der Tisch nur aus der Sicht **eines Sitzes** ansehen; die Sichtwahl in
// `Partietisch.tsx` führte allein `partie.sitze`. Der Lauf dazu steht dort.
package domain_test

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitKarten gibt dieselbe Tischlage mit einem anderen **Geheimnis** an den
// verborgenen Stellen.
//
// **Karte und Ordnung zugleich**, und das ist die Lehre einer Gegenprobe:
// der erste Anlauf änderte nur die Karte, und das Entfernen von
// `gezeigt.Ordnung = 0` blieb **grün** — beide Partien trugen dieselbe
// Ordnung, und was in beiden gleich ist, verrät im Vergleich nichts. Eine
// Gegenprobe, die nichts ändert, ist keine (§10); hier war der Lauf zu eng
// und nicht der Eingriff wirkungslos.
func mitKarten(karte string, ordnung int) domain.Partie {
	// **Der Name geht mit**, und auch das hat eine Gegenprobe gelehrt: ohne
	// ihn blieb das Entfernen von `gezeigt.Name = ""` grün, weil beide
	// Partien gar keinen trugen. Was in beiden fehlt, verrät im Vergleich
	// nichts.
	name := "Karte " + karte

	partie := einePartie()
	partie.ID = "p1"
	partie.Stand = domain.StandLaeuft

	partie.Objekte = []domain.Spielobjekt{
		// **Auf zwei fremden Händen und in zwei Bibliotheken** — die beiden
		// Zonen, die das Kriterium ausdrücklich nennt.
		{
			Kennung: "h1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneHand, Karte: karte, Name: name,
		},
		{
			Kennung: "b1", Art: domain.ObjektKarte, Beherrscher: "sitz-2",
			Eigner: "sitz-2", Zone: domain.ZoneBibliothek, Karte: karte, Name: name,
			Ordnung: ordnung,
		},

		// **Und eine verdeckt liegende Kreatur auf dem Tisch** — sie liegt in
		// einer **offenen** Zone und ist trotzdem geheim (CR 708.2a).
		{
			Kennung: "t1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: karte, Name: name,
			Verdeckt: true,
		},
	}

	return partie
}

func TestEinZuschauerKannZweiVerborgeneKartenNichtUnterscheiden(t *testing.T) {
	t.Parallel()

	eine := mitKarten("wald", 1).Sicht("")
	andere := mitKarten("schwarzer-lotus", 7).Sicht("")

	if !reflect.DeepEqual(eine.Objekte, andere.Objekte) {
		t.Errorf("ein Zuschauer sieht einen Unterschied:\n  %+v\n  %+v",
			eine.Objekte, andere.Objekte)
	}
}

// **Und der Beherrscher sieht ihn sehr wohl.**
//
// **Der Steuerlauf** (§10): ohne ihn wäre der Lauf darüber auch dann grün,
// wenn die Projektion **alles** wegwürfe — und ein Tisch, der niemandem
// etwas zeigt, ist kein Datenschutz, sondern ein Fehler.
func TestDerBeherrscherSiehtDenUnterschied(t *testing.T) {
	t.Parallel()

	eine := mitKarten("wald", 1).Sicht("sitz-1")
	andere := mitKarten("schwarzer-lotus", 7).Sicht("sitz-1")

	if reflect.DeepEqual(eine.Objekte, andere.Objekte) {
		t.Error("der Beherrscher sieht keinen Unterschied — die Projektion " +
			"wirft zu viel weg")
	}
}

// **Ein fremder Sitz sieht so viel wie ein Zuschauer** — an den Stellen, die
// keinem von beiden gehören.
//
// **Das Kriterium nennt beide in einem Satz**: „Gegneransicht und
// Zuschaueransicht zeigen keine verdeckten Handkarten,
// Bibliotheksreihenfolgen oder vertraulichen Daten."
func TestEinFremderSitzKannSieEbensoWenigUnterscheiden(t *testing.T) {
	t.Parallel()

	eine := mitKarten("wald", 1).Sicht("sitz-2")
	andere := mitKarten("schwarzer-lotus", 7).Sicht("sitz-2")

	// **Ohne die eigene Bibliothek**: `b1` gehört `sitz-2`, und ihre
	// Reihenfolge darf er sehen. Verglichen wird, was ihm **nicht** gehört.
	fremde := func(sicht domain.Partiesicht) []domain.Sichtobjekt {
		var ohne []domain.Sichtobjekt

		for _, objekt := range sicht.Objekte {
			if objekt.Beherrscher != "sitz-2" {
				ohne = append(ohne, objekt)
			}
		}

		return ohne
	}

	if !reflect.DeepEqual(fremde(eine), fremde(andere)) {
		t.Errorf("ein fremder Sitz sieht einen Unterschied:\n  %+v\n  %+v",
			fremde(eine), fremde(andere))
	}
}
