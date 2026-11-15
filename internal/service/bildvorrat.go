package service

// **Wo ein erzeugtes Bild wartet, bis jemand es nimmt oder verwirft** —
// Werkstück A12 von Phase 5.3
// (#1451).
//
// # Warum neben [Bildablage] und nicht darin
//
// **Weil die zwei verschiedene Zusagen geben.** Die Ablage schreibt neben das
// Set, dauerhaft, und was dort liegt, gehört zum Bestand
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
// Hier liegt ein **Vorschlag**: er hat eine Frist, er gehört einem Profil, er
// verschwindet beim nächsten Start, und er darf unter keinen Umständen in ein
// Set, eine Sicherung oder ein Paket geraten.
//
// Das Issue zählt auf, was das heißt: „unerratbare ID, Profil-/
// Vorschlagsbindung, MIME-/Signaturprüfung, komprimierte Byte- und
// Pixelgrenze, Ablauf, Cachequote und Bereinigung nach Abbruch/Neustart".
// Jedes dieser sieben Wörter steht unten als Zeile Code und in
// `bildvorrat_test.go` als Lauf.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bildkennungslaenge ist die Länge einer Vorratskennung in Zeichen.
//
// **Sechzehn Bytes als Hex.** Das ist dieselbe Breite wie eine UUID und
// reicht: die Kennung ist das Einzige, was ein wartendes Bild vor fremdem
// Zugriff schützt, und sie wird nie hergeleitet, sondern gewürfelt.
const Bildkennungslaenge = 32

// Bildbesitz sagt, zu wem ein wartendes Bild gehört.
//
// **Profil und Auftrag zusammen** — das Issue verlangt „Profil-/
// Vorschlagsbindung", und beides ist nötig: das Profil, weil ein zweiter
// Mensch am selben Rechner nichts davon sehen darf, der Auftrag, weil ein
// Bild zu **einem** Vorschlag gehört und nicht zu einer Sitzung.
type Bildbesitz struct {
	ProfilID int64
	Karte    int64
	Feld     string
	Auftrag  string

	// Erwartet ist der Stand der Karte, den der Auftrag eingefroren hat.
	//
	// **Er reist mit, weil ein Bildauftrag Minuten dauert** (Befund 2 des
	// Nachlaufs über Phase 5.3). [KIAuftragswerk.Anlegen] friert den Stand
	// ein — auch für ein Bild —, und zwischen dem Anlegen und dem Einsetzen
	// kann die Karte sich ändern. Ohne diesen Wert schriebe das Einsetzen
	// über eine Karte, die es so nicht mehr gibt.
	Erwartet string
}

// Bildeintrag ist, was der Vorrat über ein wartendes Bild weiß.
//
// **Ohne die Bytes.** Wer den Eintrag hat, hat die Auskunft; die Bytes gibt
// es nur über [Bildvorrat.Lesen], und der prüft das Profil.
type Bildeintrag struct {
	Kennung   string
	Besitz    Bildbesitz
	Medientyp string
	Bytes     int64
	Breite    int
	Hoehe     int
	Abgelegt  time.Time
	Verfaellt time.Time

	// Herkunft ist, was über die Entstehung feststeht — leer bei einem
	// Bild, das nicht aus einem Modell kommt.
	//
	// **Am Eintrag und nicht daneben** (A12): eine zweite Karte liefe mit
	// dieser auseinander, und dann stünde an einem Bild die Herkunft eines
	// anderen.
	Herkunft domain.KIBildherkunft
}

// Bildvorrat hält erzeugte Bilder, bis jemand sie übernimmt oder verwirft.
type Bildvorrat struct {
	ort     string
	grenzen Bildgrenzen
	jetzt   func() time.Time

	// mu deckt `eintraege` und `belegt`. Das Fenster ruft nebenläufig (§7).
	mu        sync.Mutex
	eintraege map[string]Bildeintrag
	belegt    int64
}

// NeuerBildvorrat baut ihn über einem Verzeichnis — und **leert es**.
//
// # Warum geleert und nicht gelesen
//
// **Weil ein Abbruch Dateien ohne Eintrag hinterlässt.** Der Vorrat führt
// seine Einträge im Prozess; nach einem Absturz liegen die Bilder da und
// niemand weiß mehr, wem sie gehörten. Sie zu lesen hieße, Besitz zu raten —
// sie zu löschen ist die einzige Antwort, die nichts behauptet.
//
// Das ist zugleich die „Bereinigung nach Abbruch/Neustart" aus dem Issue.
func NeuerBildvorrat(ort string, grenzen Bildgrenzen, jetzt func() time.Time) (
	*Bildvorrat, error,
) {
	if ort == "" {
		return nil, fmt.Errorf(
			"%w: ein Bildvorrat braucht ein Verzeichnis", domain.ErrValidation)
	}

	if fehler := os.RemoveAll(ort); fehler != nil {
		return nil, domain.MitGrund(domain.ErrConflict, fehler,
			"den Bildvorrat unter %q räumen", ort)
	}

	if fehler := os.MkdirAll(ort, 0o700); fehler != nil {
		return nil, domain.MitGrund(domain.ErrConflict, fehler,
			"den Bildvorrat unter %q anlegen", ort)
	}

	return &Bildvorrat{
		ort: ort, grenzen: grenzen, jetzt: jetzt,
		eintraege: map[string]Bildeintrag{},
	}, nil
}

// Ort gibt das Verzeichnis, in dem die wartenden Bilder liegen.
func (v *Bildvorrat) Ort() string { return v.ort }

// Legen nimmt ein Bild entgegen und gibt seinen Eintrag.
//
// **Die Reihenfolge ist die Zusage** (0102 §6): erst die Bytezahl, dann die
// Signatur, dann die Maße aus dem Kopf — und **erst dann** wird geschrieben.
// Wer zuerst schriebe, hätte eine Bildbombe schon auf der Platte.
func (v *Bildvorrat) Legen(besitz Bildbesitz, inhalt []byte) (Bildeintrag, error) {
	return v.LegenMitHerkunft(besitz, domain.KIBildherkunft{}, inhalt)
}

// LegenMitHerkunft ist dasselbe und schreibt die Herkunft mit.
func (v *Bildvorrat) LegenMitHerkunft(
	besitz Bildbesitz, herkunft domain.KIBildherkunft, inhalt []byte,
) (Bildeintrag, error) {
	medientyp, breite, hoehe, fehler := v.pruefe(inhalt)
	if fehler != nil {
		return Bildeintrag{}, fehler
	}

	kennung := neueBildkennung()
	jetzt := v.jetzt()

	eintrag := Bildeintrag{
		Kennung: kennung, Besitz: besitz, Medientyp: medientyp,
		Bytes: int64(len(inhalt)), Breite: breite, Hoehe: hoehe,
		Abgelegt: jetzt, Verfaellt: jetzt.Add(v.grenzen.Frist),
		Herkunft: herkunft,
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// **Abgelaufenes zuerst weg** — sonst hält ein vergessener Vorschlag
	// die Quote besetzt, obwohl ihn niemand mehr bekommt.
	v.raeumeAbgelaufene(jetzt)

	if v.belegt+eintrag.Bytes > v.grenzen.Insgesamt {
		return Bildeintrag{}, fmt.Errorf(
			"%w: der Bildvorrat ist voll (%d von %d Bytes belegt) — verwirf "+
				"einen wartenden Vorschlag", domain.ErrConflict,
			v.belegt, v.grenzen.Insgesamt)
	}

	// **Atomar, obwohl es ein Vorschlag ist.** Der Wächter über die
	// dauerhaften Dateien hat es angestrichen, und er hat recht: ein halbes
	// PNG unter einem Eintrag, der es für ganz hält, ist schlimmer als gar
	// keines — der Vorschlagskasten zeigt dann ein kaputtes Bild und niemand
	// weiß, warum.
	if fehler := schreibeZwischenspeicher(v.pfad(eintrag), inhalt); fehler != nil {
		return Bildeintrag{}, domain.MitGrund(domain.ErrConflict, fehler,
			"das Bild im Vorrat ablegen")
	}

	v.eintraege[kennung] = eintrag
	v.belegt += eintrag.Bytes

	return eintrag, nil
}

// Lesen gibt die Bytes eines wartenden Bildes.
//
// **Eine unbekannte Kennung und ein fremdes Profil bekommen denselben
// Satz.** Eine eigene Meldung für „das gehört dir nicht" bestätigte die
// Kennung, und die Kennung ist der ganze Schutz.
func (v *Bildvorrat) Lesen(kennung string, profil int64) ([]byte, Bildeintrag, error) {
	v.mu.Lock()

	eintrag, da := v.eigener(kennung, profil)
	pfad := v.pfad(eintrag)

	v.mu.Unlock()

	if !da {
		return nil, Bildeintrag{}, keinWartendesBild()
	}

	// #nosec G304 -- der Pfad entsteht aus einer gewürfelten Kennung in einem
	// Verzeichnis, das diese Anwendung selbst angelegt hat
	inhalt, fehler := os.ReadFile(pfad)
	if fehler != nil {
		return nil, Bildeintrag{}, domain.MitGrund(domain.ErrConflict, fehler,
			"das wartende Bild lesen")
	}

	return inhalt, eintrag, nil
}

// Schauen gibt den Eintrag ohne die Bytes.
func (v *Bildvorrat) Schauen(kennung string, profil int64) (Bildeintrag, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	eintrag, da := v.eigener(kennung, profil)
	if !da {
		return Bildeintrag{}, keinWartendesBild()
	}

	return eintrag, nil
}

// Verwerfen nimmt ein wartendes Bild weg.
func (v *Bildvorrat) Verwerfen(kennung string, profil int64) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	eintrag, da := v.eigener(kennung, profil)
	if !da {
		return keinWartendesBild()
	}

	v.vergiss(eintrag)

	return nil
}

// Aufraeumen wirft alles Abgelaufene weg und gibt die Zahl der Einträge.
func (v *Bildvorrat) Aufraeumen() int {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.raeumeAbgelaufene(v.jetzt())
}

// eigener gibt den Eintrag, wenn es ihn gibt, er nicht abgelaufen ist **und**
// er diesem Profil gehört. Der Aufrufer hält `mu`.
func (v *Bildvorrat) eigener(kennung string, profil int64) (Bildeintrag, bool) {
	eintrag, da := v.eintraege[kennung]
	if !da || eintrag.Besitz.ProfilID != profil {
		return Bildeintrag{}, false
	}

	if !v.jetzt().Before(eintrag.Verfaellt) {
		return Bildeintrag{}, false
	}

	return eintrag, true
}

// raeumeAbgelaufene wirft weg, was nicht mehr gilt. Der Aufrufer hält `mu`.
func (v *Bildvorrat) raeumeAbgelaufene(jetzt time.Time) int {
	weg := 0

	for _, eintrag := range v.eintraege {
		if jetzt.Before(eintrag.Verfaellt) {
			continue
		}

		v.vergiss(eintrag)

		weg++
	}

	return weg
}

// vergiss löscht Datei und Eintrag. Der Aufrufer hält `mu`.
//
// **Ein Fehler beim Löschen hält nichts auf.** Die Datei liegt in einem
// Verzeichnis, das der nächste Start ohnehin leert; den Eintrag stehen zu
// lassen, weil `os.Remove` gescheitert ist, hieße, die Quote dauerhaft zu
// belegen.
func (v *Bildvorrat) vergiss(eintrag Bildeintrag) {
	_ = os.Remove(v.pfad(eintrag)) //nolint:errcheck // siehe oben

	delete(v.eintraege, eintrag.Kennung)

	v.belegt -= eintrag.Bytes
}

// pfad gibt den Ort der Datei eines Eintrags.
func (v *Bildvorrat) pfad(eintrag Bildeintrag) string {
	return filepath.Join(v.ort, eintrag.Kennung+bildendungen[eintrag.Medientyp])
}

// keinWartendesBild ist der eine Satz für „gibt es nicht" und „gehört dir
// nicht".
func keinWartendesBild() error {
	return fmt.Errorf(
		"%w: zu dieser Kennung wartet kein Bild", domain.ErrNotFound)
}

// neueBildkennung würfelt eine.
//
// **Ohne Fehlerrückgabe**, wie an den vier anderen Stellen dieses
// Repositoriums, die würfeln: `crypto/rand.Read` „never returns an error, and
// always fills b entirely" — ein Zweig dafür wäre einer, den kein Lauf
// betreten kann, und sähe aus wie eine geprüfte Zusage.
func neueBildkennung() string {
	roh := make([]byte, Bildkennungslaenge/2)

	_, _ = rand.Read(roh) //nolint:errcheck // siehe oben

	return hex.EncodeToString(roh)
}
