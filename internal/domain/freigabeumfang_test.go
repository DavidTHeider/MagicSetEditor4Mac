package domain_test

// **Was ein Umfang zeigt und was nicht** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum jeder Umfang seinen eigenen Fall bekommt
//
// **Weil „alles geschlossen" nur so viel wert ist wie der weiteste Umfang.**
// Fünf Umfänge, die sich gegenseitig nicht überschneiden dürfen: wer Dubletten
// freigibt, gibt nicht die Sammlung frei, und wer Wünsche freigibt, gibt
// **keinen Besitz** frei. Ein Lauf, der nur den weitesten prüfte, sagte über
// die vier engeren nichts.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// besitz ist eine Besitzzeile mit einer Menge.
func besitz(printing string, menge int) domain.Exemplar {
	return domain.Exemplar{
		Printing: printing, Menge: menge,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	}
}

// wunsch ist eine Wunschzeile.
func wunsch(printing string) domain.Exemplar {
	zeile := besitz(printing, 1)
	zeile.Wunsch = true

	return zeile
}

// **Jeder Umfang zeigt genau das Seine.**
func TestJederUmfangZeigtGenauDasSeine(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		freigabe domain.Sammlungsfreigabe
		exemplar domain.Exemplar
		menge    int
		zeigt    bool
	}{
		"die Sammlung zeigt Besitz": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangSammlung},
			besitz("abc", 3), 3, true,
		},
		"die Sammlung zeigt keinen Wunsch": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangSammlung},
			wunsch("abc"), 0, false,
		},

		// **Die Menge ist die überzählige und nicht die ganze.** Wer vier hat
		// und eines behält, bietet drei an — nicht vier.
		"Dubletten zeigen, was über der Behaltemenge liegt": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangDubletten, Behalten: 1},
			besitz("abc", 4), 3, true,
		},
		"Dubletten zeigen nichts, was darunter liegt": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangDubletten, Behalten: 4},
			besitz("abc", 4), 0, false,
		},

		"Wünsche zeigen einen Wunsch": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangWuensche},
			wunsch("abc"), 1, true,
		},
		"Wünsche zeigen keinen Besitz": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangWuensche},
			besitz("abc", 3), 0, false,
		},

		"die Auswahl zeigt, was gewählt ist": {
			domain.Sammlungsfreigabe{
				Umfang: domain.UmfangAuswahl, Auswahl: []string{"abc", "def"},
			},
			besitz("abc", 2), 2, true,
		},
		"die Auswahl zeigt nichts anderes": {
			domain.Sammlungsfreigabe{
				Umfang: domain.UmfangAuswahl, Auswahl: []string{"def"},
			},
			besitz("abc", 2), 0, false,
		},

		// **Ein digitales Objekt ist kein Exemplar.** Der Umfang gibt eigene
		// Karten und Sets frei; über den Bestand sagt er nichts, und wer ihn
		// hier durchließe, zeigte Papier unter dem Namen „digital".
		"digitale Objekte zeigen keinen Bestand": {
			domain.Sammlungsfreigabe{Umfang: domain.UmfangDigitales},
			besitz("abc", 3), 0, false,
		},

		// **Ein erfundener Umfang zeigt nichts.** Der sichere Ausgang ist das
		// Nein: wer ihn offen ließe, machte aus einem Tippfehler im Bestand
		// eine offene Sammlung.
		"ein unbekannter Umfang zeigt nichts": {
			domain.Sammlungsfreigabe{Umfang: "alles"},
			besitz("abc", 3), 0, false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			menge, zeigt := fall.freigabe.Zeigt(fall.exemplar)
			if zeigt != fall.zeigt || menge != fall.menge {
				t.Errorf("es kam (%d, %v), erwartet war (%d, %v)",
					menge, zeigt, fall.menge, fall.zeigt)
			}
		})
	}
}
