package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Bulkeintrag ist eine der Sammeldateien, die Scryfall anbietet.
//
// # Warum nur fünf Felder
//
// §16: „Wer den fremden Verbund direkt in `service` reicht, hat die Struktur
// eines fremden Anbieters in seiner Fachlogik." Die Antwort führt neun Felder;
// hier stehen die, für die es einen Verwender gibt.
//
// Gelesen am 2026-08-24 an der **Antwort selbst** — und das war nötig: die
// Dokumentation nennt `download_uri` und `size`, die Antwort führte nur
// `jsonl_download_uri` und `compressed_size`. Wer der Dokumentation gefolgt
// wäre, hätte eine leere Adresse bekommen.
type Bulkeintrag struct {
	// Art ist `oracle_cards`, `default_cards`, `all_cards`, `rulings` und so
	// fort — der Schlüssel, unter dem man eine Datei meint.
	Art string

	Name         string
	Beschreibung string

	// Adresse ist `jsonl_download_uri` — eine `.jsonl.gz` auf `data.scryfall.io`.
	//
	// **JSONL und nicht JSON**, und das ist der Grund, aus dem sich diese
	// Dateien überhaupt lesen lassen: eine Karte je Zeile, kein umschließendes
	// Array. `all_cards` wiegt 392 MB gepackt — als ein Array müsste man alles
	// laden, um die erste Karte zu sehen.
	Adresse string

	// Groesse ist `compressed_size` in Bytes.
	Groesse int64

	// Stand ist `updated_at`, unverändert und **mit Zone**.
	//
	// Keine Umrechnung — dieselbe Überlegung wie bei jedem anderen Datum hier
	// (§3, „die Wanduhrzeit mit Zone, nicht der UTC-Kalendertag").
	Stand string
}

// rohbulk ist die Gestalt, die Scryfall schickt.
type rohbulk struct {
	Objekt       string `json:"object"`
	Art          string `json:"type"`
	Name         string `json:"name"`
	Beschreibung string `json:"description"`
	Adresse      string `json:"jsonl_download_uri"`
	Groesse      int64  `json:"compressed_size"`
	Stand        string `json:"updated_at"`
}

// bulkgrenze ist die Grenze für eine Sammeldatei.
//
// **Ein Gibibyte.** Die größte angebotene Datei wog am 2026-08-24 gepackt 392
// MB; die Grenze lässt Luft für Wachstum und deckelt trotzdem, was ein Fehler
// oder eine fremde Adresse anrichten könnte.
const bulkgrenze = 1 << 30

// Bulkverzeichnis holt die Liste der angebotenen Sammeldateien.
//
// **Über die allgemeine Drossel**: `/bulk-data` ist kein Kartenweg und darf
// 10/s. Die Antwort selbst ist klein — ein paar Kilobyte für sieben Einträge.
func (k *Klient) Bulkverzeichnis(ctx context.Context) (
	eintraege []Bulkeintrag, alter time.Duration, fehler error,
) {
	roh, alter, holfehler := k.speicher.Hole(ctx, "bulk-data", func() ([]byte, error) {
		return k.holeMit(ctx, k.allgemein, k.grundadresse+"/bulk-data",
			"das Verzeichnis der Sammeldateien")
	})
	if holfehler != nil {
		return nil, 0, holfehler
	}

	gelesen, lesefehler := LiesBulkverzeichnis(roh)
	if lesefehler != nil {
		return nil, 0, lesefehler
	}

	return gelesen, alter, nil
}

// LiesBulkverzeichnis liest eine Antwort von `/bulk-data`.
//
// **Die Adresse ist Pflicht.** Ein Eintrag ohne sie ist unbenutzbar, und ihn
// mitzugeben hieße, dem Aufrufer einen Abruf ins Nichts anzubieten.
func LiesBulkverzeichnis(roh []byte) ([]Bulkeintrag, error) {
	var gelesen rohliste[rohbulk]

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		return nil, fmt.Errorf("%w: das Bulk-Verzeichnis ließ sich nicht lesen: %s",
			domain.ErrValidation, fehler.Error())
	}

	eintraege := make([]Bulkeintrag, 0, len(gelesen.Daten))

	for stelle, einer := range gelesen.Daten {
		if einer.Adresse == "" {
			return nil, fmt.Errorf(
				"%w: der Bulk-Eintrag an Stelle %d (%q) nennt keine Adresse",
				domain.ErrValidation, stelle, einer.Name)
		}

		eintraege = append(eintraege, Bulkeintrag{
			Art: einer.Art, Name: einer.Name, Beschreibung: einer.Beschreibung,
			Adresse: einer.Adresse, Groesse: einer.Groesse, Stand: einer.Stand,
		})
	}

	return eintraege, nil
}

// BulkHolen lädt eine Sammeldatei und legt sie ab.
//
// # Ohne Drossel
//
// Die Dateien liegen auf `data.scryfall.io`, und dazu steht ausdrücklich: „The
// direct file origins located at *.scryfall.io" haben keine der genannten
// Grenzen. Der `User-Agent` geht trotzdem mit.
//
// # Erst vollständig, dann sichtbar
//
// Geschrieben wird in eine Nebendatei und **danach** umbenannt. Eine
// abgebrochene Datei, die unter dem endgültigen Namen läge, sähe beim nächsten
// Start aus wie ein vollständiger Bestand — und der Leser fände darin die
// Hälfte der Karten, ohne dass jemand es merkte.
//
// **Was der Zweischritt hält, und was nicht** (§10):
//
// *Gehalten* ist, dass unter dem endgültigen Namen nur eine **vollständige**
// Datei erscheint. `TestBulkfehlerwege/ein Ziel, das ein Verzeichnis ist`
// belegt es: die Nebendatei durch das Ziel zu ersetzen verschiebt den Fehler
// vom Umbenennen ins Ablegen, und der Lauf wird rot.
//
// *Nicht gehalten* ist der Fall, für den es den Zweischritt eigentlich gibt:
// ein `os.WriteFile`, das **teilweise** schreibt und dann scheitert — eine
// volle Platte. Der lässt sich in einem Test nicht herstellen.
//
// **Und was der Zweischritt nicht kann, prüft [pruefeSammeldatei]**: eine
// vollständig übertragene, inhaltlich falsche Antwort geht durch ihn hindurch
// (#1027).
//
// **Zwei Gegenproben davor waren wertlos** und stehen hier als Warnung: die
// erste prüfte am Ausfallfall, wo gar nicht geschrieben wird, die zweite an
// einem Ordner mit genau einer Datei — `os.Rename(ziel, ziel)` glückt, und
// beide blieben grün. Eine Gegenprobe, die nichts ändert, ist keine.
func (k *Klient) BulkHolen(ctx context.Context, eintrag Bulkeintrag, ziel string) error {
	if eintrag.Adresse == "" {
		return fmt.Errorf("%w: der Bulk-Eintrag nennt keine Adresse", domain.ErrValidation)
	}

	// **Einer holt, die anderen warten** (#1163, Befund Z7-03).
	//
	// Die Sammeldatei wiegt 392 MB. Zwei Prozesse mit demselben fehlenden
	// Stand luden sie **beide** vollständig; der Endstand war durch die
	// eindeutige Nachbardatei in Ordnung, das Laden aber zweimal da.
	//
	// **Und danach wird nachgesehen**: wer gewartet hat, findet den Stand, den
	// der andere gerade geschrieben hat — sonst wäre die Wartezeit umsonst
	// gewesen. Angestrichen vom Zweitblick über Phase 3.2.
	fertig := alleinLaden(ctx, ziel)
	defer fertig()

	if schonDa(ziel, eintrag) {
		return nil
	}

	antwort, fehler := netz.Hole(ctx, k.klient, nil, k.kenner, eintrag.Adresse, bulkgrenze)
	if fehler != nil {
		return fehler
	}

	if antwort.Status != netz.StatusGut {
		return fmt.Errorf("%w: die Sammeldatei antwortet mit %d",
			domain.ErrValidation, antwort.Status)
	}

	// **Und der Inhalt, bevor irgendetwas geschrieben wird** (#1027). Eine
	// förmlich geglückte Antwort mit Unsinn darin ersetzte sonst den letzten
	// guten Stand — genau den, aus dem der Betrieb ohne Netz lebt.
	if prueffehler := pruefeSammeldatei(antwort.Rumpf, eintrag); prueffehler != nil {
		return prueffehler
	}

	// **Ein Name, den zwei Läufe nicht teilen** (#1163, Befund Z7-03).
	//
	// Hier stand `ziel + ".teil"` — vorhersagbar und für jeden Prozess
	// derselbe. Zwei gleichzeitige Läufe schrieben in dieselbe Nachbardatei
	// und benannten sie beide um; die Sammeldatei ist 392 MB groß, und was
	// dabei entsteht, ist die Mischung zweier Ladevorgänge.
	neben, err := os.CreateTemp(filepath.Dir(ziel), filepath.Base(ziel)+".*.teil")
	if err != nil {
		return fmt.Errorf("die Sammeldatei ablegen: %w", err)
	}

	defer func() { _ = os.Remove(neben.Name()) }() //nolint:errcheck // nach dem Umbenennen fort

	// **Schreiben, Schließen und Umbenennen auf einem Weg.**
	//
	// Sie scheitern an derselben Ursache — der Ort nimmt die Datei nicht —,
	// und getrennte Zweige wären zwei, die kein Lauf betreten kann: ein Griff
	// auf eine eben angelegte Datei scheitert beim Schreiben und beim
	// Schließen nicht auf eine Art, die sich herbeiführen ließe (§3). Dass
	// **anlegen** eine eigene Meldung hat, bleibt: dort scheitert es an den
	// Rechten des Verzeichnisses, und das ist eine andere Auskunft.
	_, fehler = neben.Write(antwort.Rumpf)
	if schliessfehler := neben.Close(); fehler == nil {
		fehler = schliessfehler
	}

	if fehler == nil {
		fehler = os.Rename(neben.Name(), ziel)
	}

	if fehler != nil {
		return fmt.Errorf("die Sammeldatei einsetzen: %w", fehler)
	}

	return nil
}

// BulkKarteNachName sucht eine Karte in einer abgelegten Sammeldatei.
//
// # Warum zeilenweise
//
// `all_cards` wiegt gepackt 392 MB und entpackt ein Vielfaches. Sie in den
// Arbeitsspeicher zu laden, um **eine** Karte zu finden, wäre auf einem
// gewöhnlichen Rechner ein Ausfall. JSONL macht das unnötig: eine Karte je
// Zeile.
//
// # Was sie nicht tut
//
// **Sie legt keinen Index an.** Eine Suche geht durch die Datei, und das
// kostet bei `all_cards` Sekunden. Für den Fall, für den es sie gibt — eine
// Auskunft ohne Netz —, ist das der richtige Tausch: ein Index wäre ein
// zweiter Stand, der mit der Datei auseinanderlaufen kann.
func BulkKarteNachName(pfad, name string) (*Karte, []domain.Befund, error) {
	leser, fehler := OeffneBulk(pfad)
	if fehler != nil {
		return nil, nil, fehler
	}

	defer leser.Schliessen()

	return sucheInZeilen(leser, name)
}

// zeilengrenze ist die größte Zeile, die gelesen wird.
//
// **Vier Mebibyte.** Ein Kartenobjekt von Scryfall wiegt wenige Kilobyte; die
// Grenze deckelt eine Datei, die keine ist, ohne eine echte abzuweisen.
const zeilengrenze = 4 << 20

// sucheInZeilen geht die JSONL-Zeilen durch.
//
// **Über denselben Leser wie der Korpusaufbau in A15** ([Bulkleser]): zwei
// Deutungen desselben Formats liefen irgendwann auseinander.
func sucheInZeilen(leser *Bulkleser, name string) (*Karte, []domain.Befund, error) {
	gesucht := strings.TrimSpace(name)

	var befunde []domain.Befund

	for {
		karte, zeilenbefunde, fehler := leser.Naechste()
		if fehler != nil {
			return nil, nil, fehler
		}

		if karte == nil && zeilenbefunde == nil {
			break
		}

		// **Eine kaputte Zeile kostet ihre Zeile** (§8). In einer Datei
		// mit 400.000 Karten darf eine davon nicht die übrigen
		// unbrauchbar machen.
		if karte == nil {
			befunde = append(befunde, zeilenbefunde...)

			continue
		}

		// **Ohne Rücksicht auf Groß- und Kleinschreibung**, wie
		// `/cards/named?exact=` — und über `EqualFold`, das auch für
		// nicht-lateinische Schriften das Richtige tut.
		if strings.EqualFold(karte.Name, gesucht) {
			return karte, append(befunde, zeilenbefunde...), nil
		}
	}

	return nil, nil, fmt.Errorf("%w: %q steht nicht in dieser Sammeldatei",
		domain.ErrNotFound, name)
}

// Ladeendung hängt an den Pfad der Sammeldatei.
//
// **Eine eigene Nachbardatei und nicht die Sammeldatei selbst**: die wird
// umbenannt, und eine Sperre auf einer Datei, die verschwindet, ist keine.
// Sie bleibt liegen; sie zu löschen, während ein anderer sie hält, gäbe jedem
// einen eigenen Inode.
const Ladeendung = ".laedt"

// alleinLaden sorgt dafür, dass nur einer dieselbe Sammeldatei holt.
//
// **Ein Fehler hält niemanden auf.** Wer die Nachbardatei nicht bekommt, lädt
// selbst — ein doppelter Ladevorgang ist besser als kein Ergebnis.
func alleinLaden(ctx context.Context, ziel string) (fertig func()) {
	griff, err := dateisperre.Oeffnen(ziel + Ladeendung)
	if err != nil {
		return func() {}
	}

	if fehler := griff.AlleinBis(ctx, netz.Ladefrist, netz.Ladetakt); fehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // ohne Absprache geht es weiter

		return func() {}
	}

	return func() { _ = griff.Schliessen() } //nolint:errcheck // die Sperre fällt mit dem Griff
}

// schonDa sagt, ob die Sammeldatei schon in diesem Stand vorliegt.
//
// **Nach der Wartezeit gefragt und nicht davor.** Wer den Ladevorgang eines
// anderen abgewartet hat, hat ihn abgewartet, damit er ihn nicht wiederholen
// muss.
//
// **Zwei Kennzeichen, und beide müssen stimmen**: der Eintrag nennt
// `compressed_size` **und** `updated_at`. Der erste Anlauf verglich nur die
// Größe — und übersah damit den Zeitpunkt, den der Eintrag längst führte
// (Nachlauf über Phase 3.2). Zwei Fassungen gleicher gepackter Größe wären
// dauerhaft verwechselt worden.
//
// **Die Zeit entscheidet, die Größe bestätigt**: eine Datei, die **nach** dem
// angesagten Stand entstanden ist, trägt ihn — und wenn sie zusätzlich so viel
// wiegt, wie er ansagt, ist sie vollständig.
//
// **Ein unlesbarer Zeitpunkt heißt: nicht da.** Lieber einmal zu viel geladen
// als einen alten Stand für den neuen gehalten.
func schonDa(ziel string, eintrag Bulkeintrag) bool {
	if eintrag.Groesse <= 0 || eintrag.Stand == "" {
		return false
	}

	stand, fehler := time.Parse(time.RFC3339, eintrag.Stand)
	if fehler != nil {
		return false
	}

	angaben, err := os.Stat(ziel)

	return err == nil && angaben.Size() == eintrag.Groesse &&
		angaben.ModTime().After(stand)
}
