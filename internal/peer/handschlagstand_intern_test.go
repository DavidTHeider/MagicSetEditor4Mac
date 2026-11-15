package peer

// **Ein Stand, den der Ablauf nicht erzeugt** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum dieser Lauf im Paket steht
//
// **Weil [Anbahnung] ein öffentlicher Typ ist.** Wer sie von [Identitaet.
// Anbieten] bekommt, bekommt eine heile; wer `peer.Anbahnung{…}` selbst baut,
// kann hineinschreiben, was er will — und genau dann greift der Zweig, der den
// dritten Schritt scheitern lässt.
//
// **Von außen ist er nicht zu treffen**, denn dort sind die Felder unsichtbar.
// Das ist die ehrliche Fassung von „unbetretener Zweig": er **ist** betretbar,
// nur nicht über den Weg, den der Ablauf nimmt.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Anbahnung mit einer unmöglichen Sitzung schließt nicht.**
func TestEineUnmoeglicheAnbahnungSchliesstNicht(t *testing.T) {
	t.Parallel()

	ich, fehler := Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	du, fehler := Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	// **Genau so lang, dass die Kennung des Abschlusses darüber liegt.**
	//
	// Die Hülle trägt die Sitzung selbst noch ([Hoechstwort]); ihre `ID`
	// entsteht aus Sitzung und Schrittzahl und ist damit zwei Zeichen
	// länger — und **die** weist [Huelle.pruefeKopf] ab. Genau dort
	// scheitert der dritte Schritt.
	nonce := strings.Repeat("a", Hoechstwort)

	anbahnung := Anbahnung{
		ich:   ich,
		nonce: nonce,
		wunsch: Wunsch{
			Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
		},
	}

	rumpf, fehler := json.Marshal(Annahme{
		Fassung:   Handschlagfassung,
		Zweck:     domain.ZweckTeilen,
		Rolle:     "nehmend",
		Kanal:     "kanal-1",
		DeinNonce: nonce,
		MeinNonce: strings.Repeat("b", 64),
	})
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	// **Die Annahme selbst ist heil** — nur ihre Sitzung ist die unmögliche,
	// und sie wird gekürzt, damit die Hülle sie trägt.
	annahme, fehler := du.Verpacke(Huelle{
		Protokoll: ProtokollTeilen,
		Art:       ArtAnnahme,
		ID:        "n-2",
		Sitzung:   nonce,
		Absender:  string(du.Kennung()),
		Folge:     1,
		Rumpf:     rumpf,
	})
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	if _, _, fehler := anbahnung.Schliessen(annahme); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die unmögliche Anbahnung schloss: %v", fehler)
	}
}
