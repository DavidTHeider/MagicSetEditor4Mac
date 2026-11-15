package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein gescheitertes Bestätigen einer Schlüsselwortänderung wird gemeldet**
// (#1077).
//
// Paketintern und über den Stolpertreiber, aus demselben Grund wie überall
// hier: mit einer funktionierenden Datenbank ist dieser Weg nicht erreichbar,
// weil sie funktioniert.
//
// **Und er gehört geprüft, weil ein halber Stand hier ohne Rückweg dastünde.**
// Scheitert das Bestätigen und die Änderung bliebe trotzdem, wäre sie ohne ⌘Z
// nicht mehr zurückzunehmen — genau die Lücke, die dieses Issue schließt.
func TestDieWortaenderungMeldetEinGescheitertesCommit(t *testing.T) {
	db, err := sql.Open("stolper-wort-commit", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := NeueSchluesselwortablage(db).Als(domain.UrheberFenster).
		SchluesselwortAnlegen(probekontext(t), 1, "core"); fehler == nil {
		t.Error("das gescheiterte Commit kam durch")
	}
}
