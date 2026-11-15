package app

// **Den Ausgang zu einem Gegenüber tragen** —
// #1634.
//
// # Wogegen hier geprüft wird
//
// **Gegen die drei Absagen und den Normalfall.** Was wirklich getragen wird,
// misst der Lauf über einen echten Draht
// (`internal/flaechen/handelsdraht_test.go`); hier steht, was geschieht,
// bevor eine Leitung zustande kommt — und was der Bericht dann sagt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nimmtAn lässt die Gegenseite den Handschlag beantworten.
//
// **Nebenher, weil beide Seiten aufeinander warten**: wer hier der Reihe nach
// vorginge, hätte einen Lauf, der auf sich selbst wartet.
func nimmtAn(t *testing.T, dort *seite) <-chan error {
	t.Helper()

	fertig := make(chan error, 1)

	go func() {
		_, fehler := dort.anwendung.sitzungsverwaltung().Annehmen(
			dort.ctx, dort.leitung)
		fertig <- fehler
	}()

	return fertig
}

// **Ohne die Wirkung geht keine Leitung auf.**
//
// **`externSenden` und nicht `lokalSchreiben`** — der Unterschied zu
// `HandelSchicken`: jenes legt in den eigenen Ausgang, dieses baut eine
// Verbindung zu einem fremden Rechner. Wer das gleichsetzte, ließe einen
// Leser hinaussenden.
func TestOhneDieWirkungGehtKeineLeitungAuf(t *testing.T) {
	t.Parallel()

	wache := &abweisendeWache{}
	anwendung := freigabeanwendung(t).MitWache(wache)

	bericht := anwendung.Handelszustellung(t.Context(), "quatsch")

	switch {
	case bericht.Fehler == nil:
		t.Fatal("die Abweisung kam nicht an")
	case wache.gefragt != domain.WirkungExternSenden:
		t.Errorf("gefragt wurde nach %q", wache.gefragt)
	}
}

// **Was keine Peerkennung ist, bekommt keine Leitung.**
//
// **Und zwar bevor gewählt wird**: eine Zeichenkette, die nach nichts
// aussieht, soll keinen Verbindungsversuch auslösen.
func TestEineZustellungBrauchtEinePeerkennung(t *testing.T) {
	t.Parallel()

	bericht := freigabeanwendung(t).Handelszustellung(t.Context(), "quatsch")

	switch {
	case bericht.Fehler == nil:
		t.Fatal("»quatsch« ging als Peerkennung durch")
	case bericht.Fehler.Code != CodeValidierung:
		t.Errorf("der Code ist %q", bericht.Fehler.Code)
	case bericht.Getragen != 0:
		t.Errorf("getragen wurden %d", bericht.Getragen)
	}
}

// **Ohne Kontakt kommt keine Leitung zustande.**
//
// **Die Absage kommt aus der Sitzungsverwaltung** und nicht von hier: wer
// keine Adresse eingetragen hat, soll das mit dem Namen der Sache hören.
func TestOhneKontaktKommtKeineLeitungZustande(t *testing.T) {
	t.Parallel()

	jener, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	bericht := freigabeanwendung(t).Handelszustellung(
		t.Context(), string(jener.Kennung()))

	if bericht.Fehler == nil {
		t.Fatal("die Leitung kam zu einem Unbekannten zustande")
	}
}

// **Ein leerer Ausgang trägt nichts — und das ist kein Fehler.**
//
// **Die Zahl ist die Aussage**: sie sagt einem Menschen, ob etwas unterwegs
// war, und einem Skript, ob es etwas zu warten gibt.
func TestEinLeererAusgangTraegtNichtsUeberDieLeitung(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)
	fertig := nimmtAn(t, dort)

	bericht := hier.anwendung.Handelszustellung(hier.ctx, string(dort.kennung))

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("die Gegenseite: %v", fehler)
	}

	switch {
	case bericht.Fehler != nil:
		t.Fatalf("Fehler: %+v", bericht.Fehler)
	case bericht.Getragen != 0:
		t.Errorf("getragen wurden %d", bericht.Getragen)
	case bericht.Gegenueber != string(dort.kennung):
		t.Errorf("der Bericht nennt %q", bericht.Gegenueber)
	}
}

// **Ein Ausgang, der sich nicht lesen lässt, steht im Bericht.**
//
// **Und die Leitung wird trotzdem niedergelegt** — sonst bliebe nach jedem
// solchen Ruf eine offene Sitzung stehen.
func TestEinUnlesbarerAusgangStehtImBericht(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	if _, fehler := hier.anwendung.db.ExecContext(
		hier.ctx, "DROP TABLE handelsausgang"); fehler != nil {
		t.Fatalf("den Ausgang wegnehmen: %v", fehler)
	}

	fertig := nimmtAn(t, dort)

	bericht := hier.anwendung.Handelszustellung(hier.ctx, string(dort.kennung))

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("die Gegenseite: %v", fehler)
	}

	if bericht.Fehler == nil {
		t.Fatal("der fehlende Ausgang kam nicht im Bericht an")
	}
}

// **Eine Sitzung ist eine Strecke** — die Naht, an der beides zusammenkommt.
//
// **Gegenprobe zur Schnittstelle**: fiele eine der drei Methoden weg, sagte
// der Übersetzer es hier und nicht erst beim ersten Ruf.
func TestEineSitzungErfuelltDieHandelsstrecke(t *testing.T) {
	t.Parallel()

	var _ handelssitzung = (*service.Sitzungsgriff)(nil)
	var _ service.Handelsstrecke = handelsstrecke{}
}
