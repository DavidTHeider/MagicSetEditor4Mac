package peer_test

// **Was ein Handschlag abweist** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Wogegen hier geprüft wird
//
// Die Abnahme des Issues zählt sie auf: „falscher Fingerprint, wiederholte
// Nonce, abgelaufene Einladung, fremdes Profil und verweigerte Zweckfreigabe
// werden explizit abgewiesen" — und „Versionsabweichung hat festgelegte
// Ablehnung statt stillen Downgrade".
//
// **Das Nein ist die Hälfte der Zusage** (§10). Ein Lauf, der nur den
// Erfolgsweg prüft, belegt einen Handschlag, den jeder führen kann.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// einRohesAngebot baut ein Angebot von Hand — auch ein unmögliches.
func einRohesAngebot(
	t *testing.T, ich peer.Identitaet, protokoll, art string, rumpf any,
) []byte {
	t.Helper()

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	nachricht, fehler := ich.Verpacke(peer.Huelle{
		Protokoll: protokoll,
		Art:       art,
		ID:        "n-1",
		Sitzung:   strings.Repeat("a", 64),
		Absender:  string(ich.Kennung()),
		Rumpf:     roh,
	})
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	return nachricht
}

// dasAngebot ist ein gültiges, das jede Einzelprüfung besteht.
func dasAngebot(anders func(*peer.Angebot)) peer.Angebot {
	angebot := peer.Angebot{
		Fassung: peer.Handschlagfassung,
		Zweck:   domain.ZweckTeilen,
		Rolle:   "gebend",
		Kanal:   "kanal-1",
		Nonce:   strings.Repeat("a", 64),
		Name:    "Davids Rechner",
	}

	if anders != nil {
		anders(&angebot)
	}

	return angebot
}

// nimmt führt den zweiten Schritt und gibt den Befund.
func nimmt(
	t *testing.T, du peer.Identitaet, angebot []byte, bereit peer.Bereitschaft,
) error {
	t.Helper()

	_, _, fehler := du.Entgegennehmen(
		angebot, bereit, &festerZufall{stand: 200}, eineListe(t))

	return fehler
}

// zweiIdentitaeten gibt Anbieter und Annehmer.
func zweiIdentitaeten(t *testing.T) (anbieter, annehmer peer.Identitaet) {
	t.Helper()

	du, fehler := peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return eineIdentitaet(t), du
}

// **Eine fremde Fassung wird abgewiesen und nicht heruntergehandelt.**
func TestEineFremdeFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	for name, fassung := range map[string]int{
		"älter": peer.Handschlagfassung - 1,
		"neuer": peer.Handschlagfassung + 1,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			angebot := einRohesAngebot(
				t, ich, peer.ProtokollTeilen, peer.ArtAngebot,
				dasAngebot(func(a *peer.Angebot) { a.Fassung = fassung }))

			fehler := nimmt(t, du, angebot, peer.Bereitschaft{
				Rolle: "nehmend", Kanal: "kanal-1",
			})

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("die Fassung %d kam durch: %v", fassung, fehler)
			}

			// **Und das Nein nennt die Fassung.** Ein Befund, der nur
			// „ungültig" sagt, sieht aus wie jeder andere — und dann
			// belegt der Lauf nicht, dass *sie* abgewiesen wurde (§10).
			if !strings.Contains(fehler.Error(), "Fassung") {
				t.Fatalf("abgewiesen hat etwas anderes: %v", fehler)
			}
		})
	}
}

// **Ein Zweck, den es nicht gibt, öffnet nichts.**
func TestEinUnbekannterZweckOeffnetNichts(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot,
		dasAngebot(func(a *peer.Angebot) { a.Zweck = "alles" }))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Zweck kam durch: %v", fehler)
	}
}

// **Ein Zweck gilt in seinem Protokoll und in keinem anderen.**
//
// # Warum das eine eigene Prüfung ist
//
// **Weil beide für sich gültig wären.** `/mse4mac/game/1` ist ein Protokoll
// dieses Programms, `teilen` ein Zweck dieses Programms — und zusammen wären
// sie eine Partie, die in Wahrheit Karten überträgt. Das Issue trennt die drei
// Fachprotokolle; die Trennung hält nur, wenn jemand sie prüft.
func TestEinZweckGiltNurInSeinemProtokoll(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollSpiel, peer.ArtAngebot, dasAngebot(nil))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der verschobene Zweck kam durch: %v", fehler)
	}
}

// **Ein Handschlag gilt für den Kanal, auf dem er stattfand.**
//
// **Sonst ließe er sich weiterreichen**: wer ihn auf einer Verbindung
// mitschneidet und auf einer anderen einspielt, hätte eine Sitzung, die zwei
// andere Peers gebaut haben.
func TestEinHandschlagGiltNurAufSeinemKanal(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-2",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Kanal kam durch: %v", fehler)
	}
}

// **Dieselbe Zufallszahl ein zweites Mal ist ein Wiedereinspielen.**
func TestDieselbeNonceEinZweitesMalWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)
	liste := eineListe(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	bereit := peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"}

	if _, _, fehler := du.Entgegennehmen(
		angebot, bereit, &festerZufall{stand: 200}, liste); fehler != nil {
		t.Fatalf("der erste Lauf scheiterte: %v", fehler)
	}

	if _, _, fehler := du.Entgegennehmen(
		angebot, bereit, &festerZufall{stand: 200}, liste,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die zweite Nachricht kam durch: %v", fehler)
	}
}

// **Wer jemand anderen erwartet, macht nicht weiter.**
//
// Das ist der „falsche Fingerprint" aus der Abnahme — geprüft wird die
// **Kennung**, denn der Abdruck ist eine Lesehilfe über 128 Bit.
func TestWerJemandAnderenErwartetMachtNichtWeiter(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	dritter, fehler := peer.Erzeugen(&festerZufall{stand: 5})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1", Erwartet: dritter.Kennung(),
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der falsche Peer kam durch: %v", fehler)
	}
}

// **Und die anbietende Seite erwartet ebenso.**
//
// **Beide Richtungen, nicht eine.** Wer nur den Annehmer prüfen lässt, hat die
// Bestätigung für die Seite gebaut, die ohnehin wartet — und die einladende
// Seite spräche mit jedem, der antwortet.
func TestAuchDerAnbieterErwartetSeinGegenueber(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	dritter, fehler := peer.Erzeugen(&festerZufall{stand: 5})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck:    domain.ZweckTeilen,
		Rolle:    "gebend",
		Kanal:    "kanal-1",
		Erwartet: dritter.Kennung(),
	}, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	_, annahme, fehler := du.Entgegennehmen(angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}, &festerZufall{stand: 200}, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen: %v", fehler)
	}

	if _, _, fehler := anbahnung.Schliessen(annahme); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("die falsche Antwort kam durch: %v", fehler)
	}
}

// **Mit sich selbst gibt es keinen Handschlag.**
func TestMitSichSelbstGibtEsKeinenHandschlag(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	if fehler := nimmt(t, ich, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Handschlag mit mir selbst kam durch: %v", fehler)
	}
}

// **Eine andere Art an dieser Stelle ist keine.**
func TestEineAndereArtAnDieserStelleIstKeine(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAbschluss, dasAngebot(nil))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die falsche Art kam durch: %v", fehler)
	}
}

// **Eine zu kurze Zufallszahl ist keine.**
//
// **Weil eine Zufallszahl von vier Zeichen zu raten ist** — und dann ist der
// ganze Schutz gegen das Wiedereinspielen ein Buchhaltungsposten.
func TestEineZuKurzeZufallszahlIstKeine(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	for name, nonce := range map[string]string{
		"leer":     "",
		"zu kurz":  strings.Repeat("a", 30),
		"kein Hex": strings.Repeat("z", 64),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			angebot := einRohesAngebot(
				t, ich, peer.ProtokollTeilen, peer.ArtAngebot,
				dasAngebot(func(a *peer.Angebot) { a.Nonce = nonce }))

			if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
				Rolle: "nehmend", Kanal: "kanal-1",
			}); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("die Zufallszahl kam durch: %v", fehler)
			}
		})
	}
}

// **Ein Rumpf, der kein Angebot ist, ist keines.**
func TestEinRumpfDerKeinAngebotIstIstKeines(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, []string{"kein", "Verbund"})

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der fremde Rumpf kam durch: %v", fehler)
	}
}

// **Eine verfälschte Nachricht endet schon an der Hülle.**
func TestEineVerfaelschteNachrichtEndetAnDerHuelle(t *testing.T) {
	t.Parallel()

	_, du := zweiIdentitaeten(t)

	if fehler := nimmt(t, du, []byte("kein JSON"), peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Unsinn kam durch: %v", fehler)
	}
}

// **Jeder Zweck nennt sein Protokoll — und ein fremder keines.**
//
// **Die Zuordnung geht nach draußen** ([peer.ProtokollZu]), damit `service`
// eine Verbindung im **richtigen** Protokoll aufbaut. Wer sie dort nachbaute,
// hätte beim fünften Zweck zwei Listen.
func TestJederZweckNenntSeinProtokoll(t *testing.T) {
	t.Parallel()

	for zweck, erwartet := range map[domain.Zweck]string{
		domain.ZweckPaarung: peer.ProtokollPaarung,
		domain.ZweckTeilen:  peer.ProtokollTeilen,
		domain.ZweckHandel:  peer.ProtokollTeilen,
		domain.ZweckPartie:  peer.ProtokollSpiel,
	} {
		if protokoll, gilt := peer.ProtokollZu(zweck); !gilt ||
			protokoll != erwartet {
			t.Errorf("%q gab %q (%v)", zweck, protokoll, gilt)
		}
	}

	if protokoll, gilt := peer.ProtokollZu("leihgabe"); gilt {
		t.Errorf("der fremde Zweck gab %q", protokoll)
	}
}
