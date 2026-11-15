package service_test

// **Der Ausweis eines Profils** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine eigene Datei**, weil `peerhandlung_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, was mit einer
// Identität **geschieht** — hier, wie man an die kommt, mit der unterschrieben
// wird.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Ausweis ist die Identität, mit der unterschrieben wird.**
//
// # Warum er ein eigener Ruf ist und nicht in [service.Peerlage] steht
//
// **Weil die Lage an eine Fläche geht und der Ausweis nicht.** `Peerlage`
// trägt Kennung und Abdruck; der **private** Schlüssel bleibt in dieser
// Schicht, und ein Feld daneben wäre die Stelle, an der er einmal in einer
// Antwort landet.
func TestDerAusweisIstDieIdentitaetZumUnterschreiben(t *testing.T) {
	t.Parallel()

	vermerk := &vermerkprobe{}
	bund := &bundprobe{}

	werk := service.NeuesPeerwerk(vermerk, bund, &kontaktprobe{},
		func() time.Time { return derPeerlauf })

	// **Ohne Identität gibt es keinen Ausweis** — und das ist kein Ausfall,
	// sondern die Antwort: es gibt noch keine.
	if _, fehler := werk.Ausweis(t.Context()); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("ohne Identität kam ein Ausweis: %v", fehler)
	}

	lage, fehler := werk.Sicherstellen(t.Context())
	if fehler != nil {
		t.Fatalf("Sicherstellen: %v", fehler)
	}

	ich, fehler := werk.Ausweis(t.Context())
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	if ich.Kennung() != lage.Kennung {
		t.Fatalf("der Ausweis heißt %s, die Lage %s",
			ich.Kennung(), lage.Kennung)
	}

	// **Und er kann unterschreiben** — das ist der Unterschied zur Lage.
	if fehler := peer.Pruefe(
		ich.Kennung(), []byte("probe"),
		ich.Signiere([]byte("probe"))); fehler != nil {
		t.Fatalf("die Unterschrift des Ausweises gilt nicht: %v", fehler)
	}
}

// **Ein Verlust gibt keinen Ausweis, sondern sagt es.**
//
// **Der Stand `verloren` heißt: die Kennung ist vermerkt, der Schlüssel
// fehlt.** Wer daraus einen Ausweis machte, machte eine neue Identität — und
// genau das schließt das Issue aus.
func TestEinVerlustGibtKeinenAusweis(t *testing.T) {
	t.Parallel()

	vermerk := &vermerkprobe{vermerk: domain.Peervermerk{
		Kennung: "bverloren", Seit: derPeerlauf,
	}}

	werk := service.NeuesPeerwerk(vermerk, &bundprobe{}, &kontaktprobe{},
		func() time.Time { return derPeerlauf })

	if _, fehler := werk.Ausweis(t.Context()); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("der Verlust gab einen Ausweis: %v", fehler)
	}
}

// **Ein Bestand, der nicht antwortet, gibt keinen Ausweis.**
func TestEinStummerBestandGibtKeinenAusweis(t *testing.T) {
	t.Parallel()

	werk := service.NeuesPeerwerk(
		&vermerkprobe{lesefehler: errors.New("kein Bestand")},
		&bundprobe{}, &kontaktprobe{},
		func() time.Time { return derPeerlauf })

	if _, fehler := werk.Ausweis(t.Context()); fehler == nil {
		t.Fatal("der stumme Bestand gab einen Ausweis")
	}
}

// **Und ein Bund, der zwischen zwei Rufen aufgibt, ebenso wenig.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil [service.Peerwerk.Ausweis] zweimal liest**: einmal über die Lage, um
// den Stand zu bestimmen, und einmal für das Geheimnis selbst. Zwischen den
// beiden liegt eine Lücke — und ein Schlüsselbund kann sich in ihr sperren,
// etwa wenn der Bildschirm zufällt.
func TestEinBundDerZwischendurchAufgibtGibtKeinenAusweis(t *testing.T) {
	t.Parallel()

	vermerk := &vermerkprobe{}
	bund := &bundprobe{}

	werk := service.NeuesPeerwerk(vermerk, bund, &kontaktprobe{},
		func() time.Time { return derPeerlauf })

	if _, fehler := werk.Sicherstellen(t.Context()); fehler != nil {
		t.Fatalf("Sicherstellen: %v", fehler)
	}

	// **Ab dem nächsten Ruf schweigt er** — die Lage kommt noch zustande,
	// das Geheimnis nicht mehr.
	bund.lesefehlerAb = bund.gelesen + 2

	if _, fehler := werk.Ausweis(t.Context()); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("es kam etwas anderes: %v", fehler)
	}
}
