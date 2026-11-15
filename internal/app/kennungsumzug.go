package app

// **Der Umzug von der Kennung der Vorlage auf die eigene** — R02
// (#2334),
// entschieden in Vermerk 0191.
//
// # Was umzieht
//
// Was macOS unter der Bündelkennung ablegt und für Anwendende Wert hat: die
// **Vorgaben** (`~/Library/Preferences`), der **Webspeicher** des Fensters
// (`~/Library/WebKit`, `~/Library/HTTPStorages` — darin etwa die weggelegte
// Fassung der Aktualisierung und die eigenen Lagen am Tisch) und ein
// eingerichteter **Dienst** unter seinem alten Etikett. Der Fensterzustand
// unter `Saved Application State` zieht nicht um; ihn baut macOS neu.
//
// # Wie
//
// **Einmal, beim Start, und nur, was unter der neuen Kennung noch fehlt** —
// ein zweiter Start findet alles vor und tut nichts. Ordner gehen über einen
// Zwischenort, der erst am Ende umbenannt wird: ein abgebrochener Umzug
// hinterlässt keinen Ordner, der wie ein fertiger aussieht. **Das Alte bleibt
// liegen**, bis auf die plist des Dienstes — es ist der Rückweg zum alten
// Bündel, und zwei geladene Dienste stritten um einen Port.
//
// **Bestenfalls-Arbeit**: ein Fehler wird gemeldet und hält den Start nicht
// auf. Wer seine Vorgaben verliert, kann arbeiten; wer nicht starten kann,
// nicht.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Kennung ist die Bündelkennung dieses Programms.
const Kennung = "io.github.davidtheider.magicseteditor4mac"

// altkennungen ist die Kennung vor 0.9.0, in den **beiden** Schreibweisen,
// die auf der Platte stehen (Vermerk 0191).
var altkennungen = []string{"com.wails.MagicSetEditor4Mac", "com.wails.magicseteditor4mac"}

// altdienstlabel ist das Etikett des LaunchAgents vor 0.9.0.
const altdienstlabel = "com.wails.MagicSetEditor4Mac.dienst"

// umzugsmarke ist der Schlüssel, den der Umzug nach den Vorgaben in die neue
// Domäne schreibt — mit der alten Domäne als Wert.
const umzugsmarke = "Kennungsumzug"

// zwischenendung kennzeichnet einen Ordner, dessen Umzug noch läuft.
const zwischenendung = ".umzug"

// beiseitevorsatz kennzeichnet einen neuen Ordner ohne Marke, der vor dem
// Umzug beiseiteging — die Außerbetriebnahme nennt ihn (Zweitblick über
// Phase 4, Befund 13).
const beiseitevorsatz = ".vor-umzug-"

// umzugsdatei liegt in einem umgezogenen Ordner und nennt die alte Kennung —
// die Marke, dass er umgezogen ist (Zweitblick über Phase 3, Befund 10).
const umzugsdatei = ".mse-kennungsumzug"

// webordner sind die Orte unter `~/Library`, an denen WebKit unter der
// Kennung ablegt.
var webordner = []string{"WebKit", "HTTPStorages"}

// kennungsumzug zieht unter einem Zuhause um; `ausfuehren` ruft `launchctl`
// und `defaults`.
type kennungsumzug struct {
	heim       string
	domaene    string
	ausfuehren befehlsausfuehrung
}

// KennungUmziehen zieht um, was unter der alten Kennung liegt — für den Start
// des Fensters.
func KennungUmziehen(ctx context.Context) error {
	heim, fehler := os.UserHomeDir()
	if fehler != nil {
		return fmt.Errorf("das Zuhause bestimmen: %w", fehler)
	}

	// **Ein Testlauf zieht im echten Zuhause nichts um** — dieselbe Wache wie
	// beim Umzug des Bestands (`benutzerorte.go`).
	if testlauf && !zuhauseFreiFuerTestlauf(heim, os.TempDir()) {
		return errors.New("ein Testlauf zieht im echten Zuhause nichts um")
	}

	return kennungsumzug{
		heim: heim, domaene: "gui/" + strconv.Itoa(os.Getuid()),
		ausfuehren: duldetNichtGeladen(befehlAusfuehren),
	}.ziehen(ctx)
}

// ziehen führt jeden Schritt aus, auch wenn ein früherer scheitert.
func (u kennungsumzug) ziehen(ctx context.Context) error {
	fehler := make([]error, 0, len(webordner)+2)
	fehler = append(fehler, u.vorgabenZiehen(ctx))

	for _, ordner := range webordner {
		fehler = append(fehler, u.ordnerZiehen(filepath.Join(u.heim, "Library", ordner)))
	}

	return errors.Join(append(fehler, u.dienstZiehen(ctx))...)
}

// alterName findet unter `ordner` den Eintrag einer Altkennung — **nach dem
// Namen auf der Platte**: auf APFS ohne Unterscheidung der Schreibweise
// träfe ein `Stat` beide.
func alterName(ordner, endung string) string {
	eintraege, _ := os.ReadDir(ordner) //nolint:errcheck // ein fehlender Ordner hat nichts umzuziehen

	for _, eintrag := range eintraege {
		for _, alt := range altkennungen {
			if eintrag.Name() == alt+endung {
				return eintrag.Name()
			}
		}
	}

	return ""
}

// da sagt, ob es einen Pfad gibt.
func da(pfad string) bool {
	_, fehler := os.Lstat(pfad)

	return fehler == nil
}

// ordnerZiehen kopiert den Ordner der alten Kennung unter die neue.
//
// **Umgezogen heißt: die Markendatei steht im neuen Ordner** ([umzugsdatei];
// Zweitblick über Phase 3, Befund 10) — und nicht: es gibt den neuen Ordner.
// Scheiterte ein Umzug und startete das Fenster trotzdem, legt WebKit den
// neuen Ordner selbst an, und dessen Dasein hätte den Umzug für immer
// übersprungen. **Die Marke kommt mit dem Umbenennen** — sie liegt schon im
// Zwischenort —, und ein Schritt danach, der scheitern könnte, gibt es nicht.
// **Ein neuer Ordner ohne Marke geht beiseite** (`….vor-umzug-…`), nicht
// weg: er hält, was seit dem gescheiterten Umzug entstand.
func (u kennungsumzug) ordnerZiehen(unter string) error {
	alt := alterName(unter, "")
	neu := filepath.Join(unter, Kennung)

	if alt == "" || da(filepath.Join(neu, umzugsdatei)) {
		return nil
	}

	if da(neu) {
		beiseite := neu + beiseitevorsatz + strconv.FormatInt(time.Now().UnixNano(), 36)
		if fehler := os.Rename(neu, beiseite); fehler != nil {
			return fmt.Errorf("den neuen Ordner unter %s beiseitelegen: %w", unter, fehler)
		}
	}

	zwischen := neu + zwischenendung
	if fehler := os.RemoveAll(zwischen); fehler != nil {
		return fmt.Errorf("den halben Umzug unter %s räumen: %w", unter, fehler)
	}

	if fehler := os.CopyFS(zwischen, os.DirFS(filepath.Join(unter, alt))); fehler != nil {
		return fmt.Errorf("%s unter %s kopieren: %w", alt, unter, fehler)
	}

	if fehler := os.WriteFile(filepath.Join(zwischen, umzugsdatei), []byte(alt+"\n"), 0o600); fehler != nil {
		return fmt.Errorf("den Umzug unter %s vermerken: %w", unter, fehler)
	}

	return os.Rename(zwischen, neu)
}

// dienstZiehen lädt einen eingerichteten Dienst unter dem neuen Etikett —
// **erst entladen, dann laden**, und lädt der neue nicht, der alte wieder.
func (u kennungsumzug) dienstZiehen(ctx context.Context) error {
	agenten := filepath.Join(u.heim, "Library", "LaunchAgents")
	alt := filepath.Join(agenten, altdienstlabel+".plist")

	roh, fehler := plistLesen(agenten, altdienstlabel+".plist")
	if errors.Is(fehler, os.ErrNotExist) {
		return nil
	}

	if fehler != nil {
		return fmt.Errorf("die alte plist des Dienstes lesen: %w", fehler)
	}

	neu := filepath.Join(agenten, dienstlabel+".plist")

	// **Stehen beide da und ist der neue geladen, bleibt er** — er wurde nach
	// dem alten eingerichtet. Der alte geht nur noch. **Ist er nicht geladen**,
	// brach ein Umzug zwischen dem Schreiben der plist und dem Laden ab
	// (Zweitblick über Phase 3, Befund 7): dann wird er jetzt geladen, mit
	// demselben Rückweg wie beim ersten Mal.
	if da(neu) {
		geladen, fehler := u.geladen(ctx, dienstlabel)
		if fehler != nil {
			return fmt.Errorf("fragen, ob der Dienst unter dem neuen Etikett geladen ist: %w", fehler)
		}

		if !geladen {
			return u.umladen(ctx, alt, neu, false)
		}

		if fehler := u.ausfuehren(ctx, []string{"launchctl", "bootout", u.domaene + "/" + altdienstlabel}); fehler != nil {
			return fmt.Errorf("den alten Dienst entladen: %w", fehler)
		}

		return os.Remove(alt)
	}

	if fehler := service.SchreibeAtomar(neu, []byte(strings.ReplaceAll(string(roh), altdienstlabel, dienstlabel))); fehler != nil {
		return fmt.Errorf("die neue plist des Dienstes schreiben: %w", fehler)
	}

	return u.umladen(ctx, alt, neu, true)
}

// geladen fragt `launchctl print` nach einem Etikett. **Nur der Ausgang 113
// heißt „nicht geladen"** (Nachlauf über Phase 3, Befund 4) — dieselbe Lesart
// wie [launchctlPID]: ein Fragen, das sonst scheitert, ist kein Nein, und auf
// ein falsches Nein folgte ein Umladen, das eine vorhandene plist nahm.
func (u kennungsumzug) geladen(ctx context.Context, etikett string) (bool, error) {
	fehler := u.ausfuehren(ctx, []string{"launchctl", "print", u.domaene + "/" + etikett})

	var beendet interface{ ExitCode() int }

	switch {
	case fehler == nil:
		return true, nil
	case errors.As(fehler, &beendet) && beendet.ExitCode() == nichtGeladenCode:
		return false, nil
	}

	return false, fehler
}

// umladen entlädt den alten Dienst und lädt den neuen — **lädt der neue
// nicht, der alte wieder**. Die neue plist geht dabei nur, wenn **dieser**
// Umzug sie schrieb (`geschrieben`); eine, die er vorfand, bleibt (Befund 4
// des Nachlaufs).
func (u kennungsumzug) umladen(ctx context.Context, alt, neu string, geschrieben bool) error {
	zurueck := func() error {
		if !geschrieben {
			return nil
		}

		return os.Remove(neu)
	}

	if fehler := u.ausfuehren(ctx, []string{"launchctl", "bootout", u.domaene + "/" + altdienstlabel}); fehler != nil {
		return errors.Join(fmt.Errorf("den alten Dienst entladen: %w", fehler), zurueck())
	}

	if fehler := u.ausfuehren(ctx, []string{"launchctl", "bootstrap", u.domaene, neu}); fehler != nil {
		return errors.Join(fmt.Errorf("den Dienst unter dem neuen Etikett laden: %w", fehler), zurueck(),
			u.ausfuehren(ctx, []string{"launchctl", "bootstrap", u.domaene, alt}))
	}

	return os.Remove(alt)
}

// plistgrenze ist, wie viel einer plist gelesen wird — die des Dienstes hat
// unter einem Kilobyte.
const plistgrenze = 64 << 10

// plistLesen liest eine plist unter `ordner` — **in einer Wurzel und
// begrenzt**: der Name steht fest, und ein fremder Eintrag an seiner Stelle
// soll weder aus dem Ordner führen noch beliebig viel Speicher nehmen.
func plistLesen(ordner, name string) ([]byte, error) {
	wurzel, fehler := os.OpenRoot(ordner)
	if fehler != nil {
		return nil, fehler
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // gelesen ist gelesen

	datei, fehler := wurzel.Open(name)
	if fehler != nil {
		return nil, fehler
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	return io.ReadAll(io.LimitReader(datei, plistgrenze))
}
