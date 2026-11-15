package netz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// Speicher hält, was von einem fremden Dienst kam — und wie alt es ist.
//
// # Warum es ihn gibt
//
// Zwei Gründe, und beide stehen in den Bedingungen. **Höflichkeit**: „We
// encourage you to cache the data you download from Scryfall […] at least for
// 24 hours" (nachgelesen am 2026-08-23). Und **Verlässlichkeit** (§16, Regel
// 5): „Ohne Netz arbeitet die Anwendung weiter. […] Der letzte gute Stand
// bleibt, und er sagt, wie alt er ist."
//
// # Was er nicht ist
//
// **Kein Zwischenspeicher im Arbeitsspeicher.** Er soll einen Neustart
// überleben — sonst fragte jeder Start wieder alles.
//
// **Keine Ablaufsteuerung.** Ein Eintrag verschwindet nicht von selbst; er
// wird überschrieben, wenn jemand ihn neu holt. Aufräumen ist eine eigene
// Frage, und sie stellt sich erst, wenn jemand viele Karten anfasst.
type Speicher struct {
	ort   string
	frist time.Duration
	jetzt func() time.Time
}

// NeuerSpeicher baut einen Zwischenspeicher an einem Ort.
//
// **Die Uhr wird eingesetzt** (§16). Sonst prüfte kein Test die Frist, ohne
// einen Tag zu warten.
func NeuerSpeicher(ort string, frist time.Duration, jetzt func() time.Time) *Speicher {
	return &Speicher{ort: ort, frist: frist, jetzt: jetzt}
}

// eintrag ist, was auf der Platte steht.
//
// **Der Zeitpunkt gehört dazu und nicht an die Datei.** Eine Änderungszeit im
// Dateisystem überlebt ein Kopieren nicht, und sie lässt sich nicht mit dem
// Inhalt zusammen atomar ersetzen.
type eintrag struct {
	Geholt time.Time `json:"geholt"`
	Inhalt []byte    `json:"inhalt"`
}

// Hole gibt den Stand — aus dem Speicher, wenn er frisch ist, sonst neu.
//
// # Was zurückkommt
//
// Der Inhalt, sein **Alter** und ein Fehler. Das Alter ist null für etwas
// gerade Geholtes; wer einen alten Stand bekommt, sieht daran, wie alt er ist,
// und kann es sagen.
//
// # Die Reihenfolge, und warum sie so ist
//
//  1. **Ist ein frischer Stand da?** Dann wird nicht gefragt — das ist die
//     Zusage an die Gegenstelle.
//  2. **Sonst holen.** Gelingt es, wird geschrieben und ausgeliefert.
//  3. **Scheitert es und ein alter Stand ist da**, kommt der alte — mit
//     seinem Alter, ohne Fehler. Das ist die kontrollierte Verschlechterung
//     aus §16, Regel 5.
//  4. **Scheitert es und kein Stand ist da**, kommt der Fehler — und er sagt
//     beides.
func (s *Speicher) Hole(
	ctx context.Context, schluessel string, holen func() ([]byte, error),
) ([]byte, time.Duration, error) {
	gefunden, alter, gabEs := s.lies(schluessel)
	if gabEs && alter < s.frist {
		speicherzaehler.treffer.Add(1)

		return gefunden, alter, nil
	}

	// **Einer holt, die anderen warten** (#1163, Befund Z7-03).
	//
	// Ohne diese Absprache laden **beide** Prozesse denselben fehlenden Stand
	// vollständig — bei einer Scryfall-Sammeldatei sind das zweimal 392 MB.
	// Der Endstand wäre trotzdem in Ordnung; was fehlte, ist das
	// ausdrücklich geforderte **eine** Laden.
	//
	// **Und danach wird noch einmal nachgesehen.** Wer gewartet hat, findet
	// den Stand, den der andere gerade geschrieben hat — sonst wäre die
	// Wartezeit umsonst gewesen und beide hätten doch geladen.
	fertig := s.alleinHolen(ctx)
	defer fertig()

	if nachher, alterNachher, gabEsNachher := s.lies(schluessel); gabEsNachher &&
		alterNachher < s.frist {
		speicherzaehler.treffer.Add(1)

		return nachher, alterNachher, nil
	}

	frisch, fehler := holen()
	if fehler == nil {
		// **Ein Schreibfehler macht die Antwort nicht kaputt.** Sie ist da;
		// dass sie nicht auf die Platte kam, kostet die nächste Anfrage einen
		// Netzgang, mehr nicht.
		_ = s.schreib(schluessel, frisch) //nolint:errcheck // siehe oben: ein Schreibfehler kostet die nächste Anfrage einen Netzgang, mehr nicht

		speicherzaehler.geholt.Add(1)

		return frisch, 0, nil
	}

	if gabEs {
		speicherzaehler.veraltet.Add(1)

		return gefunden, alter, nil
	}

	speicherzaehler.fehlgeschlagen.Add(1)

	return nil, 0, fmt.Errorf("%w — und es gibt keinen Stand im Speicher", fehler)
}

// Vorrat gibt, was unter einem Schlüssel liegt — **ohne zu fragen und ohne
// Frist** (P10, #2378). Die Frist ist eine Zusage an die Gegenstelle; wer nur
// zeigen will, was schon bekannt ist, braucht sie nicht, und das Alter steht
// dabei.
func (s *Speicher) Vorrat(schluessel string) (inhalt []byte, alter time.Duration, gabEs bool) {
	return s.lies(schluessel)
}

// Ablegen legt einen Eintrag ab, den niemand geholt hat — etwas, das eine
// andere Antwort mitbrachte (P10, #2378: die Drucke einer Suche). **Atomar
// wie jedes Schreiben**, und ein Fehler sagt nur, dass er nicht auf die Platte
// kam.
func (s *Speicher) Ablegen(schluessel string, inhalt []byte) error {
	return s.schreib(schluessel, inhalt)
}

// lies holt einen Eintrag von der Platte.
func (s *Speicher) lies(schluessel string) (inhalt []byte, alter time.Duration, gabEs bool) {
	roh, err := os.ReadFile(s.pfad(schluessel)) // #nosec G304 -- der Pfad entsteht aus einer Prüfsumme
	if err != nil {
		return nil, 0, false
	}

	var gelesen eintrag
	if err := json.Unmarshal(roh, &gelesen); err != nil {
		// **Ein kaputter Eintrag gilt als keiner.** Er wird beim nächsten
		// geglückten Holen überschrieben; ihn zu melden hülfe niemandem.
		return nil, 0, false
	}

	return gelesen.Inhalt, s.jetzt().Sub(gelesen.Geholt), true
}

// schreib legt einen Eintrag ab — atomar.
//
// **Daneben schreiben und umbenennen.** Ein halb geschriebener Eintrag sähe
// aus wie ein gültiger; auf demselben Dateisystem ist `os.Rename` atomar, und
// ein Leser sieht deshalb entweder den alten Stand oder den neuen.
//
// # Warum der Zwischenname aus dem Ziel entsteht und nicht aus `os.CreateTemp`
//
// Weil sich beide Fehlerwege dann **betreten** lassen: ein Verzeichnis am
// Zwischennamen lässt das Schreiben scheitern, eines am Ziel das Umbenennen.
// Mit `CreateTemp` läge dazwischen ein Griff, dessen `Write` und `Close` kein
// Test zum Scheitern bringt — und §3 nimmt eine unbetretene Verzweigung nicht
// als Ausnahme an.
//
// # Was das nicht abdeckt
//
// **Zwei, die denselben Schlüssel gleichzeitig holen.** Beide schreiben auf
// denselben Zwischennamen. Was ein **Leser** sieht, bleibt trotzdem heil —
// dafür sorgt das Umbenennen —, aber der Zwischenstand kann von beiden
// stammen. Das ist hinnehmbar, solange derselbe Schlüssel dieselbe Antwort
// gibt; wer das ändert, braucht hier einen eindeutigen Namen.
func (s *Speicher) schreib(schluessel string, inhalt []byte) error {
	if err := os.MkdirAll(s.ort, 0o700); err != nil {
		return fmt.Errorf("den Speicherort anlegen: %w", err)
	}

	// **Ohne Fehlerprüfung, denn sie wäre unerreichbar.**
	//
	// `json.Marshal` scheitert an Werten, die sich nicht kodieren lassen —
	// Kanäle, Funktionen, Zeiger im Kreis. [eintrag] hat einen `time.Time` und
	// ein `[]byte`, und beide kodieren immer. §3 nimmt eine unbetretene
	// Verzweigung nicht als Ausnahme an; sie gehört weg, nicht in `.covignore`.
	//
	// **Der Fehler wird trotzdem nicht verschwiegen**: er kann keiner sein.
	// Wer [eintrag] um ein Feld erweitert, das kodieren kann scheitern, holt
	// die Prüfung zurück — und der Test dafür fällt ihm dann ein.
	roh, _ := json.Marshal(eintrag{Geholt: s.jetzt(), Inhalt: inhalt}) //nolint:errcheck,errchkjson // siehe oben

	ziel := s.pfad(schluessel)

	// **Ein Name, den zwei Läufe nicht teilen** (#1163, Befund Z7-03).
	//
	// Hier stand `ziel + ".teil"` — vorhersagbar und damit für **jeden**
	// Prozess derselbe. Zwei gleichzeitige Läufe auf denselben Schlüssel
	// schrieben in dieselbe Nachbardatei und benannten sie beide um; was am
	// Ende dastand, war die Mischung zweier Antworten oder die halbe einer.
	daneben, err := os.CreateTemp(filepath.Dir(ziel), filepath.Base(ziel)+".*.teil")
	if err != nil {
		return fmt.Errorf("den Eintrag schreiben: %w", err)
	}

	// **Aufgeräumt wird auf jedem Weg.** Nach einem geglückten Umbenennen
	// gibt es die Nachbardatei nicht mehr, und `os.Remove` sagt das als
	// `ErrNotExist` — kein Grund für eine Meldung.
	defer func() { _ = os.Remove(daneben.Name()) }() //nolint:errcheck // siehe darüber

	// **Schreiben, Schließen und Umbenennen auf einem Weg.**
	//
	// Sie scheitern an derselben Ursache — der Ort nimmt die Datei nicht —,
	// und getrennte Zweige wären zwei, die kein Lauf betreten kann: ein Griff
	// auf eine eben angelegte Datei scheitert beim Schreiben und beim
	// Schließen nicht auf eine Art, die sich herbeiführen ließe (§3). Dass
	// **anlegen** eine eigene Meldung hat, bleibt: dort scheitert es an den
	// Rechten des Verzeichnisses, und das ist eine andere Auskunft.
	_, fehler := daneben.Write(roh)
	if schliessfehler := daneben.Close(); fehler == nil {
		fehler = schliessfehler
	}

	if fehler == nil {
		fehler = os.Rename(daneben.Name(), ziel)
	}

	if fehler != nil {
		return fmt.Errorf("den Eintrag einsetzen: %w", fehler)
	}

	return nil
}

// pfad gibt den Dateinamen zu einem Schlüssel.
//
// # Warum eine Prüfsumme und nicht der Schlüssel selbst
//
// Ein Schlüssel kommt aus einer Adresse — `cards/named?exact=Lightning Bolt`.
// Wer den ungeprüft als Dateinamen nähme, schriebe bei `../` außerhalb seines
// Verzeichnisses, stolperte über Schrägstriche und über Zeichen, die ein
// Dateisystem nicht mag.
//
// **Eine Prüfsumme löst alle drei Fragen auf einmal**, und sie ist kurz genug
// für jede Namenslänge. Was sie kostet: man sieht dem Dateinamen nicht mehr
// an, was drinsteht — und das ist der Preis, den ein Verzeichnis wert ist, das
// nicht ausbrechen kann.
func (s *Speicher) pfad(schluessel string) string {
	summe := sha256.Sum256([]byte(schluessel))

	return filepath.Join(s.ort, hex.EncodeToString(summe[:])+".json")
}

// Ladefrist ist, wie lange auf einen fremden Ladevorgang gewartet wird.
//
// **Großzügig**: ein Ladevorgang darf lange dauern — die Sammeldatei von
// Scryfall wiegt 392 MB, und über eine langsame Leitung sind fünf Minuten
// dafür knapp. Wer länger braucht, hat ein anderes Problem, und dann ist
// selbst laden die bessere Antwort als Stehenbleiben.
//
// **Eine halbe Stunde und nicht fünf Minuten.** Der Nachlauf über Phase 3.2
// hat angestrichen, dass ein zulässiger 392-MB-Abruf die fünf überschreiten
// kann — und dann lädt ein zweiter Prozess doch parallel.
var Ladefrist = 30 * time.Minute

// Ladetakt ist der Abstand zwischen zwei Versuchen.
var Ladetakt = 50 * time.Millisecond

// Ladedatei ist die Nachbardatei, über die sich Ladevorgänge absprechen.
//
// # Eine je Speicher und nicht eine je Schlüssel
//
// **Weil sie liegen bleiben muss.** Eine Sperrdatei zu löschen, während ein
// anderer sie hält, gibt jedem einen eigenen Inode — und damit zwei Sperren,
// die einander nicht sehen. Eine je Schlüssel hieße also: eine leere Datei je
// jemals geholtem Eintrag.
//
// **Und weil sie nichts kostet.** Die Drossel lässt ohnehin nur eine Anfrage
// zur Zeit hinaus; feiner zu sperren als sie brächte nichts.
const Ladedatei = ".laedt"

// alleinHolen sorgt dafür, dass nur einer holt.
//
// **Ein Fehler hält niemanden auf.** Wer die Nachbardatei nicht bekommt, holt
// selbst — das ist der Stand vor Phase 3.2, und ein doppelter Ladevorgang ist
// besser als kein Ergebnis.
func (s *Speicher) alleinHolen(ctx context.Context) (fertig func()) {
	if fehler := os.MkdirAll(s.ort, 0o700); fehler != nil {
		return func() {}
	}

	griff, err := dateisperre.Oeffnen(filepath.Join(s.ort, Ladedatei))
	if err != nil {
		return func() {}
	}

	// **Mit dem Kontext des Aufrufers** (Nachlauf über Phase 3.2). Der erste
	// Anlauf nahm `context.Background()`: ein längst abgebrochener MCP-,
	// REST- oder Fensteraufruf hing damit bis zu fünf Minuten **vor** dem
	// eigentlichen Netzgang.
	if fehler := griff.AlleinBis(ctx, Ladefrist, Ladetakt); fehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // ohne Absprache geht es weiter

		return func() {}
	}

	return func() { _ = griff.Schliessen() } //nolint:errcheck // die Sperre fällt mit dem Griff
}
