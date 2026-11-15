package peer_test

// **Was die zwei letzten Schritte abweisen** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Wogegen hier geprüft wird
//
// **Dass jede Seite die Zufallszahl der anderen wirklich gesehen hat.** Das
// ist der ganze Unterschied zwischen „B besitzt einen Schlüssel" und „B ist
// gerade hier": eine mitgeschnittene, gültig unterschriebene Nachricht belegt
// das erste und nicht das zweite.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// einLauf gibt die Anbahnung und die echte Annahme dazu.
func einLauf(t *testing.T) (peer.Anbahnung, peer.Identitaet, []byte) {
	t.Helper()

	ich, du := zweiIdentitaeten(t)

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen,
		Rolle: "gebend",
		Kanal: "kanal-1",
	}, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	_, annahme, fehler := du.Entgegennehmen(angebot, peer.Bereitschaft{
		Rolle: "nehmend",
		Kanal: "kanal-1",
	}, &festerZufall{stand: 200}, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen: %v", fehler)
	}

	return anbahnung, du, annahme
}

// dieAnnahmeAus liest die Annahme aus ihrer Hülle.
func dieAnnahmeAus(t *testing.T, roh []byte) peer.Annahme {
	t.Helper()

	huelle, fehler := peer.Auspacke(roh)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	var annahme peer.Annahme

	if fehler := json.Unmarshal(huelle.Rumpf, &annahme); fehler != nil {
		t.Fatalf("Unmarshal: %v", fehler)
	}

	return annahme
}

// wiederVerpackt schreibt eine geänderte Annahme neu.
func wiederVerpackt(
	t *testing.T, du peer.Identitaet, sitzung string, annahme peer.Annahme,
) []byte {
	t.Helper()

	rumpf, fehler := json.Marshal(annahme)
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	roh, fehler := du.Verpacke(peer.Huelle{
		Protokoll: peer.ProtokollTeilen,
		Art:       peer.ArtAnnahme,
		ID:        "n-2",
		Sitzung:   sitzung,
		Absender:  string(du.Kennung()),
		Folge:     1,
		Rumpf:     rumpf,
	})
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	return roh
}

// **Eine Antwort, die meine Zufallszahl nicht kennt, ist keine.**
func TestEineAntwortOhneMeineZufallszahlIstKeine(t *testing.T) {
	t.Parallel()

	anbahnung, du, echt := einLauf(t)

	annahme := dieAnnahmeAus(t, echt)
	echtesNonce := annahme.DeinNonce

	// **Geändert wird, was zurückkommen soll** — nicht die eigene Zahl des
	// Annehmers: der Lauf heißt „kennt meine Zufallszahl nicht", und genau
	// das ist `DeinNonce`.
	annahme.DeinNonce = strings.Repeat("b", 64)

	// **Die Hülle nennt weiter die echte Sitzung**, damit der Befund aus
	// dem Rumpf kommt und nicht aus dem Kopf (§10).
	roh := wiederVerpackt(t, du, echtesNonce, annahme)

	if _, _, fehler := anbahnung.Schliessen(roh); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die fremde Zufallszahl kam durch: %v", fehler)
	}
}

// **Und eine, die den Zweck oder den Kanal verschiebt, ebenso wenig.**
//
// **Weil die Antwort sonst etwas anderes zusagte als die Frage.** Wer ein
// Teilen anbietet und eine Partie bestätigt bekommt, hat zwei Sitzungen —
// und beide Seiten glaubten, es sei eine.
func TestEineAntwortDieEtwasVerschiebtIstKeine(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*peer.Annahme){
		"anderer Zweck": func(a *peer.Annahme) { a.Zweck = domain.ZweckPartie },
		"anderer Kanal": func(a *peer.Annahme) { a.Kanal = "kanal-2" },
		"andere Fassung": func(a *peer.Annahme) {
			a.Fassung = peer.Handschlagfassung + 1
		},
		"leere Rolle": func(a *peer.Annahme) { a.Rolle = "" },
	}

	for name, aendere := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anbahnung, du, echt := einLauf(t)

			annahme := dieAnnahmeAus(t, echt)
			aendere(&annahme)

			roh := wiederVerpackt(t, du, annahme.DeinNonce, annahme)

			if _, _, fehler := anbahnung.Schliessen(roh); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Ein Abschluss, der die Zufallszahl des Annehmers nicht kennt, ist keiner.**
func TestEinAbschlussOhneDieAndereZufallszahlIstKeiner(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
	}, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	entgegennahme, annahme, fehler := du.Entgegennehmen(
		angebot, peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"},
		&festerZufall{stand: 200}, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen: %v", fehler)
	}

	_, echt, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	huelle, fehler := peer.Auspacke(echt)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	var abschluss peer.Abschluss

	if befund := json.Unmarshal(huelle.Rumpf, &abschluss); befund != nil {
		t.Fatalf("Unmarshal: %v", befund)
	}

	abschluss.DeinNonce = strings.Repeat("c", 64)

	rumpf, fehler := json.Marshal(abschluss)
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	huelle.Rumpf = rumpf

	gefaelscht, fehler := ich.Verpacke(huelle)
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	if _, fehler := entgegennahme.Abschliessen(gefaelscht); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Abschluss kam durch: %v", fehler)
	}
}

// **Und ein Abschluss von jemand anderem ist keiner.**
//
// # Warum der Abschluss dafür ein echter sein muss
//
// **Weil ein halber nichts belegt.** Der erste Anlauf schickte einen
// Abschluss mit leeren Zufallszahlen — den weist schon die Echoprüfung ab,
// und die Gegenprobe (den Absender nicht mehr prüfen) blieb **grün**. Ein
// Lauf, der seinen eigenen Namen nicht einlöst (§10). Hier steht deshalb der
// **richtige** Abschluss, nur von der falschen Hand unterschrieben.
func TestEinAbschlussVonJemandAnderemIstKeiner(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
	}, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	entgegennahme, annahme, fehler := du.Entgegennehmen(
		angebot, peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"},
		&festerZufall{stand: 200}, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen: %v", fehler)
	}

	_, echt, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	huelle, fehler := peer.Auspacke(echt)
	if fehler != nil {
		t.Fatalf("Auspacke: %v", fehler)
	}

	dritter, fehler := peer.Erzeugen(&festerZufall{stand: 5})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	// **Derselbe Rumpf, dieselbe Sitzung, andere Hand.**
	huelle.Absender = string(dritter.Kennung())

	fremd, fehler := dritter.Verpacke(huelle)
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	if _, fehler := entgegennahme.Abschliessen(fremd); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Absender kam durch: %v", fehler)
	}
}
