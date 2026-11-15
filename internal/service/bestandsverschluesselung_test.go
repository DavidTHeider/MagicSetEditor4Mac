// **Die Verschlüsselung an- und ausschalten** — Werkstück 14 von Phase 4.12
// (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// # Wonach hier gefragt wird
//
// Nicht danach, **ob** die Chiffre taugt — das steht in `internal/storage` und
// wird dort an echten Dateien gemessen. Hier steht die **Reihenfolge**, und
// die ist die fachliche Regel: wann der Schlüssel geschrieben wird, wann die
// Datei getauscht, und was geschieht, wenn ein Schritt dazwischen scheitert.
package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// chiffredoppel merkt sich, was in welcher Reihenfolge geschah.
type chiffredoppel struct {
	// schritte hält jeden Ruf in der Reihenfolge seines Auftretens.
	schritte []string

	// schluessel ist, was im Bund liegt — leer heißt: nichts.
	schluessel string

	// verschluesselt ist, was der Dateikopf sagen soll.
	verschluesselt bool

	// scheitertBei nennt den Schritt, der fehlschlagen soll.
	scheitertBei string

	// ziel merkt sich, wohin geschrieben wurde.
	ziel string

	// pruefefehler lässt das Ansehen der Datei scheitern — ab dem
	// **zweiten** Mal, damit der Weg bis zum Tausch überhaupt läuft.
	pruefefehler error

	// geprueft zählt, wie oft angesehen wurde.
	geprueft int

	// ohneDatei lässt das Umschreiben glücken, **ohne** etwas anzulegen.
	//
	// **Eine Chiffre, die Erfolg meldet und nichts hinterlässt.** Sie ist
	// künstlich — und die Abwehr dagegen steht seit jeher in
	// `bestandErsetzen` („Ein fehlender Quellpfad darf den Bestand nicht
	// kosten"). Ohne diesen Schalter bliebe sie unbetreten.
	ohneDatei bool
}

func (c *chiffredoppel) schritt(name string) error {
	c.schritte = append(c.schritte, name)

	if c.scheitertBei == name {
		return errors.New("dieser Schritt ist in diesem Lauf abgeschaltet")
	}

	return nil
}

func (c *chiffredoppel) IstVerschluesselt(string) (bool, error) {
	c.geprueft++

	if c.pruefefehler != nil && c.geprueft > 1 {
		return false, c.pruefefehler
	}

	return c.verschluesselt, nil
}

func (c *chiffredoppel) NeuerSchluessel() string {
	c.schritte = append(c.schritte, "würfeln")

	return "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

func (c *chiffredoppel) Verschluesseln(_ context.Context, ziel, schluessel string) error {
	c.ziel = ziel

	if schluessel == "" {
		return errors.New("ohne Schlüssel")
	}

	return c.schreibeZiel(ziel, "verschlüsseln")
}

func (c *chiffredoppel) Entschluesseln(_ context.Context, ziel string) error {
	c.ziel = ziel

	return c.schreibeZiel(ziel, "entschlüsseln")
}

// schreibeZiel legt die Datei wirklich an — der Tausch danach sucht sie.
func (c *chiffredoppel) schreibeZiel(ziel, name string) error {
	if err := c.schritt(name); err != nil {
		return err
	}

	if c.ohneDatei {
		return nil
	}

	return os.WriteFile(ziel, []byte("umgeschriebener Bestand"), 0o600)
}

func (c *chiffredoppel) SchluesselSchreiben(_ context.Context, schluessel string) error {
	if err := c.schritt("Schlüssel ablegen"); err != nil {
		return err
	}

	c.schluessel = schluessel

	return nil
}

func (c *chiffredoppel) SchluesselLoeschen(context.Context) error {
	if err := c.schritt("Schlüssel löschen"); err != nil {
		return err
	}

	c.schluessel = ""

	return nil
}

// verschluesselungsweg baut einen Bestandsweg mit einem echten Bestandspfad.
func verschluesselungsweg(t *testing.T, chiffre *chiffredoppel) (
	weg *service.Bestandsweg, bestandsort string,
) {
	t.Helper()

	return wegMitLebenszyklus(t, chiffre,
		func() error { chiffre.schritte = append(chiffre.schritte, "schließen"); return nil },
		func(context.Context) error {
			chiffre.schritte = append(chiffre.schritte, "öffnen")

			return nil
		},
		nil)
}

// wegMitLebenszyklus baut ihn mit einem eigenen Lebenszyklus.
//
// **Getrennt, weil die Fehlerwege genau ihn brauchen**: was geschieht, wenn
// das Schließen scheitert, wenn jemand anders den Bestand hält, wenn das
// Wiederöffnen nicht glückt. Ein `nil` für `allein` heißt: wir sind allein.
func wegMitLebenszyklus(
	t *testing.T, chiffre *chiffredoppel,
	schliessen func() error, oeffnen func(context.Context) error,
	allein service.Alleinpruefung,
) (weg *service.Bestandsweg, bestandsort string) {
	t.Helper()

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	if err := os.WriteFile(bestand, []byte("der bisherige Bestand"), 0o600); err != nil {
		t.Fatalf("Bestand anlegen: %v", err)
	}

	if allein == nil {
		allein = func(context.Context, string) (func() error, error) {
			return func() error { return nil }, nil
		}
	}

	weg = service.NeuerBestandsweg(
		service.Sicherungsorte{
			Datenbank: bestand,
			Bilder:    func() (string, error) { return filepath.Join(ort, "beiwerk"), nil },
		},
		nil, allein, nil, schliessen, oeffnen, "probe",
	)

	weg.MitBestandschiffre(chiffre)

	return weg, bestand
}

// **Der Schlüssel liegt im Bund, bevor die Datei getauscht wird.**
//
// # Warum das die eine Regel ist, auf die es ankommt
//
// Zwischen „die Datei ist verschlüsselt" und „der Schlüssel liegt im Bund"
// darf **kein** Augenblick liegen, in dem das erste gilt und das zweite nicht
// — dort verliert jemand seine Daten, und zwar endgültig.
//
// Andersherum ist die Lücke harmlos: ein Schlüssel im Bund, während die Datei
// noch Klartext ist, stört niemanden. `storage.schluesselFuer` fragt den Bund
// gar nicht erst, solange der Dateikopf die Klartextmarke trägt.
func TestDerSchluesselLiegtVorDemTausch(t *testing.T) {
	chiffre := &chiffredoppel{}
	weg, _ := verschluesselungsweg(t, chiffre)

	if _, err := weg.VerschluesselungEinschalten(t.Context()); err != nil {
		t.Fatalf("einschalten: %v", err)
	}

	folge := strings.Join(chiffre.schritte, " → ")

	ablegen := indexVon(chiffre.schritte, "Schlüssel ablegen")
	if ablegen < 0 {
		t.Fatalf("der Schlüssel wurde nie abgelegt: %s", folge)
	}

	verschluesseln := indexVon(chiffre.schritte, "verschlüsseln")
	if verschluesseln < 0 {
		t.Fatalf("es wurde nie verschlüsselt: %s", folge)
	}

	if ablegen > verschluesseln {
		t.Errorf("der Schlüssel wurde nach dem Umschreiben abgelegt: %s", folge)
	}
}

// **Und beim Ausschalten umgekehrt: erst tauschen, dann löschen.**
//
// Dieselbe Regel von der anderen Seite — der Schlüssel muss so lange da sein,
// wie die Datei ihn brauchen könnte.
func TestDerSchluesselFaelltNachDemTausch(t *testing.T) {
	chiffre := &chiffredoppel{verschluesselt: true, schluessel: "abc"}
	weg, _ := verschluesselungsweg(t, chiffre)

	if _, err := weg.VerschluesselungAusschalten(t.Context()); err != nil {
		t.Fatalf("ausschalten: %v", err)
	}

	folge := strings.Join(chiffre.schritte, " → ")

	loeschen := indexVon(chiffre.schritte, "Schlüssel löschen")
	entschluesseln := indexVon(chiffre.schritte, "entschlüsseln")

	if loeschen < 0 || entschluesseln < 0 {
		t.Fatalf("ein Schritt fehlt: %s", folge)
	}

	if loeschen < entschluesseln {
		t.Errorf("der Schlüssel wurde vor dem Umschreiben gelöscht: %s", folge)
	}
}

// **Die Verbindung wird geschlossen, bevor getauscht wird, und danach wieder
// geöffnet** — dieselbe Regel wie beim Zurückspielen.
func TestDieVerbindungSchliesstUndOeffnetWieder(t *testing.T) {
	chiffre := &chiffredoppel{}
	weg, _ := verschluesselungsweg(t, chiffre)

	if _, err := weg.VerschluesselungEinschalten(t.Context()); err != nil {
		t.Fatalf("einschalten: %v", err)
	}

	folge := strings.Join(chiffre.schritte, " → ")

	if indexVon(chiffre.schritte, "schließen") < 0 ||
		indexVon(chiffre.schritte, "öffnen") < 0 {
		t.Fatalf("Lebenszyklus fehlt: %s", folge)
	}

	if indexVon(chiffre.schritte, "öffnen") < indexVon(chiffre.schritte, "schließen") {
		t.Errorf("geöffnet wurde vor dem Schließen: %s", folge)
	}
}

// **Was schon verschlüsselt ist, wird nicht zweimal verschlüsselt.**
func TestZweimalEinschaltenIstEinFehler(t *testing.T) {
	chiffre := &chiffredoppel{verschluesselt: true}
	weg, _ := verschluesselungsweg(t, chiffre)

	_, fehler := weg.VerschluesselungEinschalten(t.Context())

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war eine Abweisung", fehler)
	}

	if len(chiffre.schritte) != 0 {
		t.Errorf("es wurde gearbeitet: %v", chiffre.schritte)
	}
}

// **Und was nicht verschlüsselt ist, lässt sich nicht entschlüsseln.**
func TestAusschaltenOhneVerschluesselungIstEinFehler(t *testing.T) {
	chiffre := &chiffredoppel{}
	weg, _ := verschluesselungsweg(t, chiffre)

	_, fehler := weg.VerschluesselungAusschalten(t.Context())

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war eine Abweisung", fehler)
	}
}

// **Scheitert das Umschreiben, ist der Lebenszyklus nie angefasst worden.**
//
// # Die Zusicherung war zuerst schwächer, und der Lauf hat es angestrichen
//
// Hier stand „der Bestand bleibt offen" und darunter die Frage, ob wieder
// **geöffnet** wurde. Der Lauf wurde rot, und zwar zu Recht: umgeschrieben
// wird, **solange die Verbindung offen ist** — sie ist die Quelle. Scheitert
// das, war nie geschlossen, und ein Wiederöffnen wäre nicht Aufräumen,
// sondern ein zweiter Vorgang.
//
// **Die schärfere Aussage ist deshalb, dass gar nichts geschah**: kein
// Schließen, kein Öffnen, kein Tausch. Sie prüft, was die Reihenfolge
// wirklich zusagt, statt eines Rückwegs, den es hier nicht braucht.
//
// **Und der abgelegte Schlüssel bleibt liegen.** Das ist Absicht: ihn hier
// wegzuräumen hieße, in dem Augenblick zu löschen, in dem am wenigsten
// feststeht, was auf der Platte liegt.
func TestEinGescheitertesUmschreibenLaesstDenBestandStehen(t *testing.T) {
	chiffre := &chiffredoppel{scheitertBei: "verschlüsseln"}
	weg, bestand := verschluesselungsweg(t, chiffre)

	if _, fehler := weg.VerschluesselungEinschalten(t.Context()); fehler == nil {
		t.Fatal("das Einschalten glückte, obwohl das Umschreiben abgeschaltet war")
	}

	// #nosec G304 -- ein Pfad aus t.TempDir()
	inhalt, err := os.ReadFile(bestand)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if string(inhalt) != "der bisherige Bestand" {
		t.Errorf("der Bestand wurde angefasst: %q", inhalt)
	}

	if indexVon(chiffre.schritte, "schließen") >= 0 {
		t.Errorf("die Verbindung wurde geschlossen: %v", chiffre.schritte)
	}

	if chiffre.schluessel == "" {
		t.Error("der Schlüssel wurde weggeräumt — siehe den Kopf dieses Laufs")
	}
}

// **Und der Stand wird gemeldet, wie er ist.**
func TestDerStandSagtObVerschluesseltIst(t *testing.T) {
	for name, an := range map[string]bool{"verschlüsselt": true, "im Klartext": false} {
		t.Run(name, func(t *testing.T) {
			chiffre := &chiffredoppel{verschluesselt: an}
			weg, _ := verschluesselungsweg(t, chiffre)

			stand, err := weg.Verschluesselungsstand()
			if err != nil {
				t.Fatalf("Stand: %v", err)
			}

			if stand != an {
				t.Errorf("Stand %v, erwartet war %v", stand, an)
			}
		})
	}
}

// indexVon gibt die Stelle eines Schritts — oder -1.
func indexVon(schritte []string, gesucht string) int {
	for stelle, schritt := range schritte {
		if schritt == gesucht {
			return stelle
		}
	}

	return -1
}
