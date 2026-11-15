// Der Probelauf einer Exportvorlage — UXR-F7
// (#2015).
//
// # Was er ist
//
// Die Vorlage wird **ausgewertet** — über denselben Weg wie im echten Lauf
// ([App.rechneVorlage]) —, und es wird nichts geschrieben. Gesagt wird, was
// dabei herauskäme: für wie viele Karten, wie groß die Hauptdatei wird, wie
// viele Dateien daneben entstünden, und was die Auswertung zu melden hat.
//
// # Was er nicht ist
//
// **Er fragt den Schreibriegel nicht** (0078): geschrieben wird nichts, und
// ein Leser darf wissen, was eine Vorlage ergäbe. Der echte Lauf fragt ihn,
// und das ist der eine Unterschied in den Fehlern. Er legt auch kein
// Verzeichnis an — ein Ziel, das sich nicht anlegen ließe, zeigt erst der
// echte Lauf.

package app

import (
	"context"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// VorlagenprobelaufAntwort sagt, was eine Exportvorlage ergäbe.
type VorlagenprobelaufAntwort struct {
	// Karten sind die Karten, über die die Vorlage lief — nach der Auswahl.
	Karten int `json:"karten"`

	// Dateiname ist der Name der Hauptdatei: der des Ziels, sonst der
	// Vorschlag des Dialogs (`ausfuhr` samt Endung aus `file type:`).
	Dateiname string `json:"dateiname"`

	// Bytes ist die Größe der Hauptdatei. **Sie ist genau**: die Vorlage ist
	// ausgewertet, und ihr Ergebnis ist der Text, der geschrieben würde.
	Bytes int `json:"bytes"`

	// Beigaben sind die Dateien, die die Vorlage daneben schreiben würde —
	// Bilder, Texte, das Set —, und BeigabenBytes die bekannte Größe davon.
	Beigaben      int `json:"beigaben"`
	BeigabenBytes int `json:"beigabenBytes"`

	// Verzeichnis ist der Name des Ordners, in den sie kämen — leer, wenn die
	// Vorlage keinen anlegt.
	Verzeichnis string `json:"verzeichnis"`

	// Geschaetzt sagt, dass mindestens eine Beigabe dabei ist, deren Größe erst
	// das Rastern oder Packen kennt: BeigabenBytes ist dann eine Untergrenze.
	Geschaetzt bool `json:"geschaetzt"`

	// Befunde sind die Meldungen der Auswertung — und bei einem Fehler der
	// Satz der Vorlage, wie im echten Lauf (#209).
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// vorlagenvorschlag ist der Name, den der Speicherdialog vorschlägt.
//
// **Eine Stelle für Dialog und Probelauf**: der Name, den der Probelauf
// nennt, ist der, den der Dialog danach vorschlägt.
func vorlagenvorschlag(dateityp string) string {
	return "ausfuhr" + endungVon(erstesMuster(dateityp))
}

// VorlageProbelauf wertet eine Vorlage aus und sagt, was sie ergäbe — ohne zu
// schreiben.
//
// **Das Ziel ist wahlfrei**: das Fenster kennt es erst nach dem Dialog und
// fragt ohne; die anderen Flächen nennen es und bekommen die Beigaben unter dem
// Namen, den der echte Lauf ihnen gäbe.
func (a *App) VorlageProbelauf(
	ctx context.Context, setID, paket, ziel string, wahl map[string]string,
	auswahl []string,
) VorlagenprobelaufAntwort {
	ctx = a.imProfil(ctx)

	leer := VorlagenprobelaufAntwort{Befunde: []string{}}

	kennung, befund := kennungAus(setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	vorlage, gewaehlt, fehler := a.vorlageWaehlen(ctx, kennung, paket, auswahl)
	if fehler != nil {
		leer.Fehler = fehler

		return leer
	}

	if ziel == "" {
		ziel = vorlagenvorschlag(vorlage.Dateityp)
	}

	ablage := service.NeueProbeablage()

	rechnung, err := a.rechneVorlage(ctx, kennung, vorlage, ziel, wahl, gewaehlt, ablage)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		// **Der Satz des Skriptfehlers reist mit, wie im echten Lauf** (#209).
		if rechnung.gelaufen {
			leer.Befunde = append(meldungen(rechnung.befunde), err.Error())
		}

		return leer
	}

	stand := ablage.Stand()
	befunde := meldungen(rechnung.befunde)

	// **Eine leere Ausgabe wird gesagt**, wie im echten Lauf (#1347).
	if rechnung.ausgabe == "" {
		befunde = append(befunde, befundLeereAusgabe)
	}

	return VorlagenprobelaufAntwort{
		Karten:        rechnung.karten,
		Dateiname:     filepath.Base(ziel),
		Bytes:         len(rechnung.ausgabe),
		Beigaben:      stand.Dateien,
		BeigabenBytes: stand.Bytes,
		Verzeichnis:   rechnung.verzeichnis,
		Geschaetzt:    stand.Geschaetzt,
		Befunde:       befunde,
	}
}
