package service

import (
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// keinArchiv weist ein Ziel ab, das eine Datei statt eines Verzeichnisses ist.
//
// # Warum ein Archiv nicht zurückgeschrieben wird
//
// Ein `.mse-set` darf eine ZIP-Datei sein — [msepaket.Oeffne] liest beides seit
// Phase 2. Zurückgeschrieben wird es trotzdem nicht, und die Entscheidung samt
// ihrem Preis steht als
// [0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md).
//
// **Der Grund für die Prüfung hier und nicht tiefer**: `os.MkdirAll` trifft auf
// eine reguläre Datei und meldet „not a directory". Das ist wahr und sagt
// niemandem, was los ist — der Zweitblick über Phase 6 hat es angestrichen
// (#208).
//
// **Ein Ziel, das es nicht gibt, ist keines**: „Sichern unter" nennt einen
// neuen Ort, und der existiert vorher nicht.
func keinArchiv(ziel string) error {
	// **Ein Ziel, das sich nicht befragen lässt, ist keines für diese Frage.**
	// Es gibt es entweder nicht — „Sichern unter" nennt einen neuen Ort — oder
	// es ist unlesbar, und daran scheitert gleich darauf `os.MkdirAll` mit
	// seiner eigenen, genaueren Meldung. Hier daraus einen Fehler zu machen
	// hieße, denselben Umstand zweimal zu melden.
	angaben, fehler := os.Stat(ziel)
	if fehler != nil || angaben.IsDir() {
		//nolint:nilerr // der Grund steht darüber
		return nil
	}

	// **Als Feldbefund und nicht als blankes `ErrValidation`** —
	// #704.
	//
	// [0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md)
	// sagt zu, ⌘S abzuweisen „mit einem Satz, der das Archiv nennt und auf
	// ‚Sichern unter' zeigt". Ein Befund trägt seine Meldung mit.
	//
	// **Ohne Feldbezug**, denn es gibt keinen: der Umstand hängt am Ziel,
	// nicht an einem Feld der Karte.
	return domain.NewFieldError("", domain.SchluesselValidierung,
		"%s ist ein Archiv; ein Archiv wird nicht zurückgeschrieben — "+
			"\u201eSichern unter\u201c gibt ihm einen neuen Ort als Verzeichnis",
		filepath.Base(ziel))
}
