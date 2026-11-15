package werkzeuge_test

// **Die Umfangsmatrix nennt jeden Posten und jeden Teil** — OPS-06.AK-01
// (#1723).
//
// # Was er hält
//
// **Jeder Posten der Platzübersicht** steht in `docs/SICHERUNG.md` mit dem
// Ort, an dem er liegt — ein neuer Posten erzwingt die Entscheidung, ob er
// in die Sicherung gehört. **Und jeder Teil des Archivs** steht dort unter
// dem Namen, den er im Archiv trägt.
//
// # Was er nicht hält
//
// **Ob die Einordnung stimmt.** „ja" oder „nein" ist eine Entscheidung mit
// Grund; sie prüft ein Mensch.

import (
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// ortDesPostens nennt, woran die Matrix einen Posten erkennt.
var ortDesPostens = map[domain.Platzposten]string{
	domain.PostenBestand:        "`bestand.db`",
	domain.PostenBilder:         "`beiwerk/`",
	domain.PostenPakete:         "`pakete/`",
	domain.PostenVorsicherungen: "`bestand-vor-*.db`",
	domain.PostenProtokoll:      "`mse.log*`",
	domain.PostenSonstiges:      "„sonstiges",
}

func TestDieUmfangsmatrixNenntJedenPostenUndTeil(t *testing.T) {
	t.Parallel()

	roh, fehler := os.ReadFile("../../docs/SICHERUNG.md")
	if fehler != nil {
		t.Fatal(fehler)
	}

	matrix := string(roh)

	for _, posten := range domain.PostenAlle() {
		ort, bekannt := ortDesPostens[posten]
		if !bekannt {
			t.Errorf("der Posten %q ist neu — gehört er in die Sicherung? Eintragen in docs/SICHERUNG.md und hier", posten)

			continue
		}

		if !strings.Contains(matrix, ort) {
			t.Errorf("der Posten %q fehlt in der Matrix (gesucht: %s)", posten, ort)
		}
	}

	for _, teil := range []string{
		"`" + service.Bestandsname + "`", "`" + service.Beiwerkordner + "/…`",
		"`" + service.Paketordner + "/<name>/…`", "`" + sicherung.Manifestname + "`",
	} {
		if !strings.Contains(matrix, teil) {
			t.Errorf("der Teil %s steht nicht in der Matrix", teil)
		}
	}
}
