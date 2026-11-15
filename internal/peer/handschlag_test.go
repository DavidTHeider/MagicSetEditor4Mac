package peer_test

// **Der signierte Handschlag** — Werkstück N1 von Phase 5.4
// (#1454),
// dritter Schnitt.
//
// # Wogegen hier geprüft wird
//
// **Dass beide Seiten am Ende dasselbe meinen.** Nicht nur „B hat einen
// Schlüssel", sondern: derselbe Zweck, derselbe Kanal, dieselbe Sitzung, und
// jede Seite hat die Zufallszahl der anderen **gesehen**.
//
// **Und dass ein Name nichts belegt.** Das Issue sagt es wörtlich:
// „Anzeigenamen allein sind kein Identitätsnachweis".

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineListe gibt ein Noncegedächtnis mit stehender Uhr.
func eineListe(t *testing.T) *peer.Nonceliste {
	t.Helper()

	jetzt := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

	return peer.NeueNonceliste(func() time.Time { return jetzt }, time.Hour)
}

// einHandschlag führt die drei Schritte und gibt beide Sitzungen.
func einHandschlag(
	t *testing.T, ich, du peer.Identitaet,
) (meine, deine peer.Sitzung) {
	t.Helper()

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen,
		Rolle: "gebend",
		Kanal: "kanal-1",
		Name:  "Davids Rechner",
	}, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	entgegennahme, annahme, fehler := du.Entgegennehmen(
		angebot, peer.Bereitschaft{
			Rolle: "nehmend",
			Kanal: "kanal-1",
			Name:  "Fremder Rechner",
		}, &festerZufall{stand: 200}, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen: %v", fehler)
	}

	// **Die annehmende Seite weiß schon jetzt, wer da ist und was er will.**
	//
	// Das ist der Griff, an dem `service` sein Recht prüft — **zwischen**
	// Annahme und Antwort, und deshalb muss er hier schon stimmen.
	if entgegennahme.Gegenueber() != ich.Kennung() {
		t.Errorf("das Gegenüber ist %s", entgegennahme.Gegenueber())
	}

	if entgegennahme.Zweck() != domain.ZweckTeilen {
		t.Errorf("der Zweck ist %q", entgegennahme.Zweck())
	}

	var abschluss []byte

	meine, abschluss, fehler = anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	deine, fehler = entgegennahme.Abschliessen(abschluss)
	if fehler != nil {
		t.Fatalf("Abschliessen: %v", fehler)
	}

	return meine, deine
}

// **Beide Seiten kommen auf dieselbe Sitzung.**
func TestBeideSeitenKommenAufDieselbeSitzung(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	du, fehler := peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	meine, deine := einHandschlag(t, ich, du)

	if meine.Kennung == "" || meine.Kennung != deine.Kennung {
		t.Fatalf("zwei Sitzungen: %q und %q", meine.Kennung, deine.Kennung)
	}

	// **Jede Seite sieht die andere** — und zwar als Kennung, nicht als Name.
	if meine.Gegenueber != du.Kennung() {
		t.Errorf("mein Gegenüber ist %s", meine.Gegenueber)
	}

	if deine.Gegenueber != ich.Kennung() {
		t.Errorf("dein Gegenüber ist %s", deine.Gegenueber)
	}

	// **Die Rollen stehen über Kreuz.**
	if meine.MeineRolle != "gebend" || meine.SeineRolle != "nehmend" {
		t.Errorf("meine Rollen sind %+v", meine)
	}

	if deine.MeineRolle != "nehmend" || deine.SeineRolle != "gebend" {
		t.Errorf("deine Rollen sind %+v", deine)
	}

	if meine.Zweck != domain.ZweckTeilen || deine.Zweck != domain.ZweckTeilen {
		t.Errorf("die Zwecke sind %q und %q", meine.Zweck, deine.Zweck)
	}
}

// **Der Fingerabdruck kommt aus der Kennung und nicht aus der Nachricht.**
//
// # Der Fund
//
// **Ein selbstgewählter Name belegt nichts.** Wer ihn neben den Abdruck
// stellt, hat zwei Angaben — und nur eine davon hat jemand unterschrieben.
// Deshalb heißt das Feld hier, wie es heißt.
func TestDerAbdruckKommtAusDerKennung(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	du, fehler := peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	meine, deine := einHandschlag(t, ich, du)

	if meine.Fingerabdruck != peer.Fingerabdruck(du.Kennung()) {
		t.Errorf("der Abdruck ist %q", meine.Fingerabdruck)
	}

	if meine.SelbstgewaehlterName != "Fremder Rechner" {
		t.Errorf("der Name ist %q", meine.SelbstgewaehlterName)
	}

	if deine.SelbstgewaehlterName != "Davids Rechner" {
		t.Errorf("der Name ist %q", deine.SelbstgewaehlterName)
	}
}

// **Zwei Läufe mit demselben Paar ergeben zwei Sitzungen.**
//
// **Weil die Zufallszahlen beider Seiten eingehen** — eine Sitzungskennung,
// die sich aus Zweck und Kennungen allein ergäbe, wäre für dasselbe Paar
// immer dieselbe, und eine mitgeschnittene Nachricht passte in den nächsten
// Lauf.
func TestZweiLaeufeErgebenZweiSitzungen(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	du, fehler := peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	erste, _ := einHandschlag(t, ich, du)

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen,
		Rolle: "gebend",
		Kanal: "kanal-1",
	}, &festerZufall{stand: 31})
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

	zweite, _, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	if erste.Kennung == zweite.Kennung {
		t.Fatal("zwei Läufe teilen eine Sitzungskennung")
	}
}
