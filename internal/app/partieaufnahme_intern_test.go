package app

// **Was mit einer hereinkommenden Partiesitzung geschieht** — Werkstück G5
// von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Sitzung ohne offene Partie niedergelegt wird** — und nicht
// stehen bleibt. Der Gast wartet sonst auf einen Tisch, der nie kommt; eine
// Absage ist die ehrlichere Antwort (§8).
//
// **Und dass ein Gast den Sitz bekommt, den der Wirt vergeben hat** — nicht
// den, den er nennt. Der Grund steht bei `service.Partiefreigabe`: eine
// Sitzung belegt einen Rechner und keinen Sitz.

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// angenommenendoppel ist eine Sitzung, die nichts hergibt und sich merkt, ob
// sie geschlossen wurde.
type angenommenendoppel struct {
	schloss sync.Mutex

	stand       service.Sitzungsstand
	geschlossen bool
	lesefehler  error

	// gibt sind die Zustellungen, die diese Sitzung hergibt — danach wartet
	// sie auf den Kontext.
	gibt    [][]byte
	gelesen int

	schliessfehler error

	// sendefehler lässt jedes Senden scheitern (G5, #1462).
	sendefehler error

	// leerfehler lässt das Leeren scheitern — die **zweite** Hälfte.
	leerfehler error

	// gesendet hält fest, was hinausging (#2506).
	gesendet [][]byte
}

func (d *angenommenendoppel) Sende(
	_ context.Context, nachricht []byte, _ service.Zustellung,
) error {
	d.schloss.Lock()
	defer d.schloss.Unlock()

	if d.sendefehler == nil {
		d.gesendet = append(d.gesendet, nachricht)
	}

	return d.sendefehler
}

// tischbilder gibt die Tischbilder, die hinausgingen.
func (d *angenommenendoppel) tischbilder(t *testing.T) []domain.Tischbild {
	t.Helper()

	d.schloss.Lock()
	defer d.schloss.Unlock()

	var bilder []domain.Tischbild

	for _, roh := range d.gesendet {
		var nachricht domain.Partienachricht
		if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
			t.Fatalf("keine Nachricht: %v", fehler)
		}

		if nachricht.Art != domain.ArtKontrolle {
			continue
		}

		art, inhalt, fehler := domain.LiesKontrolle(nachricht.Inhalt, nachricht.Von,
			domain.KontrolleTischbild)
		if fehler != nil || art != domain.KontrolleTischbild {
			continue
		}

		var bild domain.Tischbild
		if fehler := json.Unmarshal(inhalt, &bild); fehler != nil {
			t.Fatalf("kein Tischbild: %v", fehler)
		}

		bilder = append(bilder, bild)
	}

	return bilder
}

// Empfange gibt erst die vorbereiteten Zustellungen, dann den Lesefehler.
//
// **Die Reihenfolge ist tragend**: ein Lauf, der eine Zustellung prüfen will,
// braucht danach ein **Ende** — sonst läuft die Schleife weiter, und der Lauf
// hängt bis zu seiner Frist. Genau das ist beim ersten Anlauf passiert: 600
// Sekunden für eine Zusicherung, die nach einer Millisekunde feststand.
func (d *angenommenendoppel) Empfange(ctx context.Context) ([]byte, error) {
	if d.gelesen < len(d.gibt) {
		d.gelesen++

		return d.gibt[d.gelesen-1], nil
	}

	if d.lesefehler != nil {
		return nil, d.lesefehler
	}

	// **Warten statt sofort enden** — eine Leitung, die nichts hergibt und
	// trotzdem zurückkehrt, ließe die Schleife durchdrehen.
	<-ctx.Done()

	return nil, ctx.Err()
}

func (d *angenommenendoppel) Empfangen([]byte) bool { return true }

func (d *angenommenendoppel) Leere(context.Context) error { return d.leerfehler }

func (d *angenommenendoppel) Stand() service.Sitzungsstand { return d.stand }

func (d *angenommenendoppel) Schliesse() error {
	d.schloss.Lock()
	defer d.schloss.Unlock()

	d.geschlossen = true

	return d.schliessfehler
}

func (d *angenommenendoppel) istZu() bool {
	d.schloss.Lock()
	defer d.schloss.Unlock()

	return d.geschlossen
}

// eineAngenommene gibt ein Doppel mit Zweck und einem **echten** Gegenüber.
//
// **Echt, weil der Fingerabdruck aus dem Schlüssel gerechnet wird.** Eine
// erfundene Kennung hat keinen — `peer.Fingerabdruck` gibt dafür bewusst eine
// leere Zeichenkette zurück, statt zwei kaputte gleich aussehen zu lassen. Ein
// Lauf über einen erfundenen Wert prüfte also die Anzeige eines Fehlerfalls
// und nicht den Normalbetrieb.
func eineAngenommene(t *testing.T, zweck domain.Zweck) *angenommenendoppel {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return &angenommenendoppel{stand: service.Sitzungsstand{
		Zweck: zweck, Gegenueber: ich.Kennung(),
	}}
}

// **Eine Sitzung für einen anderen Zweck wird nicht angefasst.**
//
// **Und nicht geschlossen**: ein Handschlag oder ein Deckhandel hat mit einer
// Partie nichts zu tun, und wer ihn hier niederlegte, beendete eine Sitzung,
// die jemand anderes führt.
func TestEineFremdeSitzungWirdNichtAngefasst(t *testing.T) {
	t.Parallel()

	sitzung := eineAngenommene(t, domain.ZweckHandel)

	freigabeanwendung(t).partiesitzungAufnehmen(t.Context(), sitzung)

	if sitzung.istZu() {
		t.Error("die fremde Sitzung wurde geschlossen")
	}
}

// **Ohne offene Partie wird die Sitzung niedergelegt.**
func TestOhneOffenePartieWirdDieSitzungNiedergelegt(t *testing.T) {
	t.Parallel()

	sitzung := eineAngenommene(t, domain.ZweckPartie)

	freigabeanwendung(t).partiesitzungAufnehmen(t.Context(), sitzung)

	if !sitzung.istZu() {
		t.Error("die Sitzung blieb offen, obwohl nichts freisteht")
	}
}

// **Ist der Tisch voll, wird sie ebenfalls niedergelegt.**
func TestEinVollerTischLegtDieSitzungNieder(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	// **Der erste nimmt den einzigen Sitz.**
	erster := eineAngenommene(t, domain.ZweckPartie)
	anwendung.partiesitzungAufnehmen(t.Context(), erster)

	zweiter := eineAngenommene(t, domain.ZweckPartie)

	anwendung.partiesitzungAufnehmen(t.Context(), zweiter)

	switch {
	case erster.istZu():
		t.Error("der erste wurde abgewiesen")
	case !zweiter.istZu():
		t.Error("der zweite bekam einen Sitz an einem vollen Tisch")
	}
}

// **Und ein aufgenommener Gast steht im Stand — mit seinem Sitz.**
//
// **Das ist die Zusage, die das Fenster zeigt**: wer sitzt schon, und auf
// welchem Platz.
func TestEinAufgenommenerGastStehtImStand(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	anwendung.partiesitzungAufnehmen(
		t.Context(), eineAngenommene(t, domain.ZweckPartie))

	antwort := anwendung.PartieFreigabeStand(t.Context())

	switch {
	case len(antwort.Vergeben) != 1:
		t.Fatalf("%d vergeben, erwartet 1", len(antwort.Vergeben))
	case antwort.Vergeben[0].Sitz != "bert":
		t.Errorf("Sitz %q", antwort.Vergeben[0].Sitz)
	case antwort.Vergeben[0].Fingerabdruck == "":
		t.Error("ohne Fingerabdruck kann niemand vergleichen")
	}
}

// **Eine Leitung, die sofort klagt, beendet die Schleife ohne Aufhebens —
// und gibt die Sitzung frei.**
//
// **Ein Ende ist kein Fehlschlag**: eine Leitung, die zugeht, ist der
// Normalfall. Was daraus für die Partie folgt, sagt `App.tischVerloren`
// — und dieser Rechner **ist** der Wirt.
//
// **Hier stand bis zum 2026-10-01 das Gegenteil**: „die Schleife hat die
// Sitzung geschlossen — das tut sie nicht". Ohne Aufhebens hieß aber nie
// „ohne aufzuräumen" — die tote Sitzung blieb in der Übersicht, und jeder
// neue Versuch desselben Gastes scheiterte an ihr (Messlauf über Phase 6.7,
// `partiewiederkehr_intern_test.go`). **Und auch ein Schließen, das klagt,
// ändert daran nichts**: die Klage steht im Protokoll.
func TestEineKlagendeLeitungBeendetDieSchleife(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name           string
		schliessfehler error
	}{
		{"das Schließen gelingt", nil},
		{"das Schließen klagt", errors.New("schon zu")},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			anwendung := freigabeanwendung(t)
			anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

			sitzung := eineAngenommene(t, domain.ZweckPartie)
			sitzung.lesefehler = errors.New("die Leitung ist weg")
			sitzung.schliessfehler = fall.schliessfehler

			anwendung.partieBedienen(t.Context(), sitzung, "p1", "bert",
				einePartiesitzung(sitzung), einGastplatz(sitzung))

			if !sitzung.istZu() {
				t.Error("die Schleife hat die beendete Sitzung nicht freigegeben")
			}
		})
	}
}

// **Eine unlesbare Zustellung wird gemeldet und hält die Sitzung nicht an.**
//
// **Die Meldung ist der Punkt** (§8): sie ist ein Befund über **eine**
// Nachricht. Wer dafür die Leitung niederlegte, ließe ein fremdes Programm
// die Partie beenden — mit einem einzigen kaputten Byte.
func TestEineUnlesbareZustellungWirdGemeldetUndNichtVerschluckt(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	sitzung := eineAngenommene(t, domain.ZweckPartie)
	sitzung.gibt = [][]byte{[]byte("{kein json")}
	sitzung.lesefehler = errors.New("danach ist Schluss")

	anwendung.partieBedienen(t.Context(), sitzung, "p1", "bert",
		einePartiesitzung(sitzung), einGastplatz(sitzung))

	if sitzung.gelesen != 1 {
		t.Errorf("gelesen wurde %dmal — die Schleife lief nicht weiter",
			sitzung.gelesen)
	}
}

// **Eine Sitzung, die sich nicht schließen lässt, wird gemeldet statt
// verschluckt.**
//
// **Sie kommt vor**: eine Leitung, die schon weg ist, klagt beim Niederlegen.
// Der Weg endet trotzdem — was hier fehlte, wäre nur die Auskunft.
func TestEineUnschliessbareSitzungWirdGemeldet(t *testing.T) {
	t.Parallel()

	sitzung := eineAngenommene(t, domain.ZweckPartie)
	sitzung.schliessfehler = errors.New("die Leitung ist schon weg")

	freigabeanwendung(t).partiesitzungAufnehmen(t.Context(), sitzung)

	if !sitzung.istZu() {
		t.Error("es wurde gar nicht erst versucht")
	}
}

// einePartiesitzung baut die Sendeseite, die `partiesitzungAufnehmen` sonst
// baut.
//
// **Sie wird hier von Hand gebaut, weil diese Läufe die Schleife prüfen und
// nicht das Zusammenstecken.** Wer beides in einem Lauf prüfte, wüsste bei
// einem roten Ergebnis nicht, welches der beiden gefallen ist.
func einePartiesitzung(griff service.Angenommene) *service.Partiesitzung {
	return service.NeuePartiesitzung(
		"p1", "anna", griff, service.NeueReihenfolge(erstePartieepoche))
}

// einGastplatz baut den Platz dazu.
func einGastplatz(griff service.Angenommene) *service.Gastplatz {
	return service.NeuerGastplatz("bert", einePartiesitzung(griff), griff)
}
