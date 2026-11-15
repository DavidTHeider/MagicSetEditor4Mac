// **Eine Sicherung, die nicht (mehr) daliegt, ist ein fachlicher Befund** —
// gefunden bei C01
// (#2226).
//
// Bis dahin galt eine fehlende Datei als nackte Datenbank aus 0.1.0, und erst
// das Kopieren scheiterte — mit dem rohen Fehler des Systems, den das Fenster
// als „unerwarteter Fehler" zeigte. Zwischen Dialog und Zurückspielen kann
// eine Datei verschwinden, und Kommandozeile, MCP und REST nehmen einen
// Pfad ohne Dialog.
package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

func TestEineFehlendeSicherungIstNichtGefunden(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	for _, pfad := range []string{
		filepath.Join(ort, "weg"+sicherung.Endung),
		filepath.Join(ort, "weg.db"),
	} {
		_, schaufehler := weg.Beschauen(t.Context(), pfad)
		if !errors.Is(schaufehler, domain.ErrNotFound) {
			t.Errorf("Beschauen %s: %v", filepath.Base(pfad), schaufehler)
		}

		_, rueckfehler := weg.Zurueckspielen(t.Context(), pfad)
		if !errors.Is(rueckfehler, domain.ErrNotFound) {
			t.Errorf("Zurückspielen %s: %v", filepath.Base(pfad), rueckfehler)
		}

		// **Und die Lage vor einem Rücksprung** — derselbe Weg zum Auspacken;
		// der Vertragsrundgang fand auch sie als „unerwarteter Fehler".
		_, lagefehler := weg.Ruecksprunglage(t.Context(), pfad)
		if !errors.Is(lagefehler, domain.ErrNotFound) {
			t.Errorf("Rücksprunglage %s: %v", filepath.Base(pfad), lagefehler)
		}
	}
}

// **Und eine, die daliegt und sich nicht lesen lässt, scheitert beim
// Kopieren** — nicht als „nicht gefunden": ein Verzeichnis unter dem Namen
// einer alten Datenbank öffnet sich, liest sich aber nicht.
func TestEineUnlesbareNackteSicherungScheitertBeimKopieren(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	pfad := filepath.Join(ort, "verzeichnis.db")
	if fehler := os.Mkdir(pfad, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	_, rueckfehler := weg.Zurueckspielen(t.Context(), pfad)
	if rueckfehler == nil || errors.Is(rueckfehler, domain.ErrNotFound) {
		t.Errorf("Zurückspielen eines Verzeichnisses: %v", rueckfehler)
	}
}
