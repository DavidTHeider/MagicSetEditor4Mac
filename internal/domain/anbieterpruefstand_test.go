package domain_test

// **Wogegen ein Anmeldeanbieter geprüft ist** — ID-05.AK-01
// (#1704).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Alle vier nur gegen eine Nachbildung** — gegen den echten Anbieter lief
// keiner, weil dazu ein eigenes Klientenkonto gehört, das dieses Projekt
// nicht mitliefert. Wer einen Lauf nachträgt, trägt ihn in
// `anbieterlaeufe` ein, und dieser Test wird dort rot.
func TestJederAnmeldeanbieterNenntWogegenErGeprueftIst(t *testing.T) {
	t.Parallel()

	for _, anbieter := range domain.Anbieter() {
		if stand := domain.AnbieterPruefstand(anbieter); stand != domain.PruefstandSimuliert {
			t.Errorf("%s: %q", anbieter, stand)
		}
	}

	if stand := domain.AnbieterPruefstand("apple"); stand != "" {
		t.Errorf("unbekannter Anbieter: %q", stand)
	}
}
