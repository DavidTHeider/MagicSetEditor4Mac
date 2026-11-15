package app

// **Eine Adresse gehört zu ihrer Bindung, auch wenn eine davor ausfällt** —
// Befund 3 des Zweitblicks über Phase G
// (#2272).
//
// Der Dienst sagt nur die Adressen an, auf denen ein Lauscher entstand; das
// Fenster paarte sie über die Stelle in der **gewählten** Liste. Scheiterte
// die erste Bindung, rückte jede weitere eine Stelle auf — und die
// Erreichbarkeit fragte an einer Adresse, die zu einer anderen Bindung
// gehört.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieErreichbarkeitFragtDieAdresseIhrerBindung(t *testing.T) {
	t.Parallel()

	// **Die Schleife scheitert, zwei Adressen im Netz laufen.** Nach außen
	// fragt die erste, die den Rechner verlässt: 192.0.2.10.
	bindungen := []domain.Bindung{
		{Adresse: "127.0.0.1", Port: 8080},
		{Adresse: "192.0.2.10", Port: 8080},
		{Adresse: "198.51.100.7", Port: 8080},
	}
	lauf := dienstlauf{
		adresse:  "https://192.0.2.10:8080",
		adressen: []string{"https://192.0.2.10:8080", "https://198.51.100.7:8080"},
		fehlende: []domain.Lauschbefund{{Adresse: "127.0.0.1", Grund: "belegt"}},
	}

	draussen, gibt := ersteNachAussen(bindungen, lauf)
	if !gibt || draussen != "https://192.0.2.10:8080" {
		t.Errorf("gefragt wird an %q (%v) — erwartet die Adresse von 192.0.2.10", draussen, gibt)
	}
}

// **Und eine mittlere, die ausfällt**, verschiebt die nach ihr nicht: `::1`
// scheitert zwischen Schleife und Netz, und die Netzadresse rückte vorher
// über das Ende der Liste hinaus — gefragt wurde dann an der Schleife.
func TestEineMittlereAusfallendeBindungVerschiebtNichts(t *testing.T) {
	t.Parallel()

	bindungen := []domain.Bindung{
		{Adresse: "127.0.0.1", Port: 8080},
		{Adresse: "::1", Port: 8080},
		{Adresse: "192.0.2.10", Port: 8080},
	}
	lauf := dienstlauf{
		adresse:  "https://127.0.0.1:8080",
		adressen: []string{"https://127.0.0.1:8080", "https://192.0.2.10:8080"},
		fehlende: []domain.Lauschbefund{{Adresse: "::1", Grund: "nicht da"}},
	}

	draussen, gibt := ersteNachAussen(bindungen, lauf)
	if !gibt || draussen != "https://192.0.2.10:8080" {
		t.Errorf("gefragt wird an %q (%v) — erwartet die Adresse von 192.0.2.10", draussen, gibt)
	}
}
