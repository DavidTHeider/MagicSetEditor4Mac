package schriften

// **Wo eine Schrift liegt und was sich daran geändert hat** — aus `suche.go`
// herausgelöst, als sie über die Zeilengrenze aus §2 ging (G2,
// #1459).
//
// **Der Schnitt folgt der Frage**: dort steht, **welche** Schrift gemeint ist
// und wie sie gefunden wird, hier, **wo** sie liegt und **woran** man merkt,
// dass sich etwas geändert hat.

import (
	"os"
	"path/filepath"
	"strings"
)

// ohneEndung gibt einen Dateinamen ohne seine Endung.
func ohneEndung(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// Fundort sagt, welche Datei zu einem Schriftnamen gehört.
//
// **Für die Prüfung der Reihenfolge**, und nur dafür: `Suche` gibt die Schrift
// und verrät nicht, woher sie kam — womit „System vor Paket" eine Behauptung
// bliebe, die kein Test einlösen kann.
func (s *Suche) Fundort(name string, fett bool) (string, error) {
	// **Unter demselben Schloss wie `Suche`** — `schnitt` baut die
	// Namensindizes, und die teilen beide.
	s.schloss.Lock()
	defer s.schloss.Unlock()

	ort, datei, err := s.schnitt(name, fett)
	if err != nil {
		return "", err
	}

	return filepath.Join(ort, datei), nil
}

func passendeEndung(endung string) bool {
	for _, erlaubt := range Endungen {
		if endung == erlaubt {
			return true
		}
	}

	return false
}

// ortestand gibt die Änderungszeiten der Orte als eine Zeichenkette.
//
// **Ein `os.Stat` je Ort**, nicht ein Lauf über ihren Inhalt: die Orte sind
// eine kurze Liste von Verzeichnissen, und wer eine Schrift installiert,
// ändert das Verzeichnis, in das er sie legt.
//
// **Ein unlesbarer Ort zählt als leer und nicht als unverändert** — dieselbe
// Regel wie bei [Suche.paketstand], und aus demselben Grund: teurer, aber nie
// falsch.
func (s *Suche) ortestand() string {
	var stand strings.Builder

	for _, ort := range s.orte {
		stand.WriteString(ort)
		stand.WriteByte(0)

		if angaben, err := os.Stat(ort); err == nil {
			stand.WriteString(angaben.ModTime().String())
		}

		stand.WriteByte(0)
	}

	return stand.String()
}

// Stand fasst zusammen, wovon ein Fund abhängt: die Änderungszeiten der
// Schriftorte und der Paketwurzeln — derselbe Stand, mit dem die Suche ihre
// gemerkten Fehlschläge verwirft ([Suche.fehlschlaege]). Eine Problemprüfung
// über ein Set (PERF02, #2352) trägt ihn in ihrem Abdruck: eine Schrift, die
// danach dazukommt, macht ihr Ergebnis veraltet.
func (s *Suche) Stand() string {
	s.schloss.RLock()
	defer s.schloss.RUnlock()

	return s.paketstand() + "\x00" + s.ortestand()
}
