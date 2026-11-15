package service_test

// **Die Adresse im Kontaktwerk** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird und was nicht
//
// **Dass die Form geprüft wird, bevor sie den Bestand erreicht** — und dass
// eine Adresse **keine** Bestätigung voraussetzt. Das zweite ist die
// eigentliche Entscheidung: wer sie an die Bestätigung bände, könnte ein
// frisch eingetragenes Gegenüber nicht einmal anrufen, um seinen Abdruck zu
// vergleichen.
//
// **Nicht gemessen wird, ob dort jemand antwortet.** Das ist eine Messung und
// gehört an die Leitung; hier steht eine Regel.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Adresse kommt an und steht am Kontakt.**
func TestEineAdresseStehtDanachAmKontakt(t *testing.T) {
	t.Parallel()

	doppel := einKontaktdoppel("bkennung-eins")
	werk := service.NeuesKontaktwerk(doppel, einePeeruhr)

	if fehler := werk.AdresseSetzen(
		t.Context(), "bkennung-eins", "192.168.1.7:41234"); fehler != nil {
		t.Fatalf("Adresse: %v", fehler)
	}

	liste, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(liste) != 1 || liste[0].Adresse != "192.168.1.7:41234" {
		t.Fatalf("die Liste gibt %+v", liste)
	}
}

// **Ein unbestätigtes Gegenüber bekommt trotzdem eine.**
//
// **Weil eine Adresse nichts erlaubt.** Sie sagt, wohin gewählt wird, und
// nicht, was danach gilt — das entscheidet der Handschlag samt Rechten.
func TestAuchEinUnbestaetigtesGegenueberBekommtEineAdresse(t *testing.T) {
	t.Parallel()

	doppel := einKontaktdoppel("bkennung-zwei")
	werk := service.NeuesKontaktwerk(doppel, einePeeruhr)

	liste, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if liste[0].Stand != domain.KontaktUnbestaetigt {
		t.Fatalf("der Stand ist %q", liste[0].Stand)
	}

	if fehler := werk.AdresseSetzen(
		t.Context(), "bkennung-zwei", "beispiel.invalid:443"); fehler != nil {
		t.Fatalf("Adresse: %v", fehler)
	}
}

// **Eine krumme Adresse erreicht den Bestand nicht.**
//
// **Und das ist am Doppel gemessen und nicht am Fehler**: ein Werk, das die
// Form erst vom Bestand prüfen ließe, wäre für eine zweite Fläche ohne
// Bestand wirkungslos.
func TestEineKrummeAdresseErreichtDenBestandNicht(t *testing.T) {
	t.Parallel()

	doppel := einKontaktdoppel("bkennung-drei")
	werk := service.NeuesKontaktwerk(doppel, einePeeruhr)

	if fehler := werk.AdresseSetzen(
		t.Context(), "bkennung-drei", "frag mal David",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if doppel.kontakte["bkennung-drei"].Adresse != "" {
		t.Fatalf("im Bestand steht %q",
			doppel.kontakte["bkennung-drei"].Adresse)
	}
}

// **Und wen es nicht gibt, der bekommt keine.**
func TestWenDasWerkNichtKenntDerBekommtKeineAdresse(t *testing.T) {
	t.Parallel()

	werk := service.NeuesKontaktwerk(einKontaktdoppel(), einePeeruhr)

	if fehler := werk.AdresseSetzen(
		t.Context(), "bkennung-niemand", "beispiel.invalid:443",
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Die eingetragene Adresse kommt wieder heraus.**
func TestDieEingetrageneAdresseKommtWiederHeraus(t *testing.T) {
	t.Parallel()

	werk := service.NeuesKontaktwerk(
		einKontaktdoppel("bkennung-sechs"), einePeeruhr)

	if fehler := werk.AdresseSetzen(
		t.Context(), "bkennung-sechs", "beispiel.invalid:443"); fehler != nil {
		t.Fatalf("AdresseSetzen: %v", fehler)
	}

	gelesen, fehler := werk.Adresse(t.Context(), "bkennung-sechs")
	if fehler != nil {
		t.Fatalf("Adresse: %v", fehler)
	}

	if gelesen != "beispiel.invalid:443" {
		t.Fatalf("es kam %q", gelesen)
	}
}

// **Ohne Eintrag ist das ein Befund und keine leere Zeichenkette.**
//
// # Warum hier, anders als bei [service.Kontaktwerk.Darf]
//
// **Weil die Frage eine andere ist.** „Darf er handeln?" hat für einen
// Unbekannten ein klares Nein; „wo erreiche ich ihn?" hat keine Antwort. Ein
// `""` wäre eine erfundene — und der Aufrufer wählte danach und bekäme einen
// Fehler von der Leitung statt von der Sache.
func TestOhneEintragIstDieAdresseEinBefund(t *testing.T) {
	t.Parallel()

	// **Eine echte Kennung** — mit einer erfundenen ist der Abdruck leer, und
	// die Prüfung darunter fand den leeren Text in jedem Satz.
	kennung := erzeugt(t).Kennung()
	werk := service.NeuesKontaktwerk(einKontaktdoppel(kennung), einePeeruhr)

	_, fehler := werk.Adresse(t.Context(), kennung)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Mit eigenem Ausgang** (#2491): am Fenster stand sonst „lade die
	// Liste neu" — gefunden mit der Hüllenprobe `p10handel09`.
	if !errors.Is(fehler, domain.ErrPeerOhneAdresse) {
		t.Errorf("es kam %v und nicht ErrPeerOhneAdresse", fehler)
	}

	// **Und der Satz nennt den Abdruck**, damit ein Mensch weiß, wen er
	// meint — eine Kennung ist 52 Zeichen lang und sagt niemandem etwas.
	if abdruck := peer.Fingerabdruck(kennung); abdruck == "" ||
		!strings.Contains(fehler.Error(), abdruck) {
		t.Fatalf("der Satz lautet: %v", fehler)
	}
}

// **Und wen das Werk nicht kennt, dessen Adresse ist auch keine.**
func TestWenDasWerkNichtKenntDessenAdresseIstKeine(t *testing.T) {
	t.Parallel()

	werk := service.NeuesKontaktwerk(einKontaktdoppel(), einePeeruhr)

	if _, fehler := werk.Adresse(
		t.Context(), "bkennung-niemand",
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam %v", fehler)
	}
}

// einKontaktdoppel baut ein Doppel, das die genannten Kontakte kennt.
func einKontaktdoppel(kennungen ...peer.Kennung) *kontaktdoppel {
	doppel := &kontaktdoppel{kontakte: map[string]*domain.Peerkontakt{}}

	for _, kennung := range kennungen {
		doppel.kontakte[string(kennung)] = &domain.Peerkontakt{
			Kennung: string(kennung),
			Stand:   domain.KontaktUnbestaetigt,
		}
	}

	return doppel
}
