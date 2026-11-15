package peer

// **Die kanonische Form steht fest** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum dieser Lauf im Paket steht und nicht daneben
//
// **Weil die Form kein Rückgabewert ist.** Von außen ist nur zu sehen, *dass*
// eine Unterschrift hält — nicht, **worüber** sie geht. Der Bereich am Anfang
// hat heute gar keinen Gegenspieler: N1.3 bringt den Handschlag, und erst dann
// gäbe es eine zweite unterschriebene Gestalt, die man verwechseln könnte.
// Eine Zusage, die erst in drei Werkstücken zu widerlegen wäre, ist bis dahin
// keine — dieser Lauf macht sie heute prüfbar.

import (
	"encoding/json"
	"testing"
)

// **Die Form ist Byte für Byte diese und keine andere.**
func TestDieUnterschriftsformStehtFest(t *testing.T) {
	t.Parallel()

	huelle := Huelle{
		Protokoll: ProtokollTeilen,
		Art:       "hallo",
		ID:        "n-1",
		Sitzung:   "s-1",
		Absender:  "bwer",
		Folge:     7,
		Rumpf:     json.RawMessage(`{"was":"wert"}`),
	}

	erwartet := "16:mse4mac/huelle/1" +
		"16:/mse4mac/share/1" +
		"5:hallo" +
		"3:n-1" +
		"3:s-1" +
		"4:bwer" +
		"1:7" +
		`14:{"was":"wert"}`

	if form := string(huelle.unterschriftsform()); form != erwartet {
		t.Fatalf("die Form ist\n%q\nerwartet war\n%q", form, erwartet)
	}
}
