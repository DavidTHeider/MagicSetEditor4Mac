package service_test

// **Die übrigen Wege des Kontaktwerks** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine zweite Datei**, weil `kontaktwerk_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, was ein
// Gegenüber **darf** — hier, wie man es sieht, sperrt und wieder hereinlässt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Liste zeigt, wer bekannt ist.**
func TestDieListeZeigtWerBekanntIst(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, _ := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatalf("Begegnet: %v", fehler)
	}

	liste, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(liste) != 1 || liste[0].Kennung != string(kennung) {
		t.Fatalf("die Liste ist %+v", liste)
	}

	// **Der Abdruck steht dabei und wird aus der Kennung gerechnet.**
	if liste[0].Fingerabdruck != peer.Fingerabdruck(kennung) {
		t.Errorf("der Abdruck ist %q", liste[0].Fingerabdruck)
	}
}

// **Eine Zeichenkette, aus der kein Schlüssel wird, ist kein Kontakt.**
//
// # Der Fund
//
// **Sonst stünde sie mit leerem Abdruck in der Liste** — und ein leerer
// Abdruck ist genau das, was [service.Kontaktwerk.Bestaetigen] abweist. Ein
// Eintrag, den niemand je bestätigen kann, sieht aus wie ein Kontakt und ist
// keiner.
func TestEineKrummeKennungIstKeinKontakt(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)

	for name, kennung := range map[string]peer.Kennung{
		"leer":        "",
		"kein base32": "bxxx!!!",
		"zu kurz":     "babc",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := werk.Begegnet(
				t.Context(), kennung, "Wer"); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("sie kam durch: %v", fehler)
			}
		})
	}
}

// **Eine Sperre nimmt alles, ein Entsperren gibt nichts zurück.**
func TestEineSperreNimmtAllesUndGibtNichtsZurueck(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, abdruck := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatalf("Begegnet: %v", fehler)
	}

	if fehler := werk.Bestaetigen(
		t.Context(), kennung, abdruck); fehler != nil {
		t.Fatalf("Bestaetigen: %v", fehler)
	}

	if fehler := werk.Erlauben(
		t.Context(), kennung, domain.ZweckPartie); fehler != nil {
		t.Fatalf("Erlauben: %v", fehler)
	}

	if fehler := werk.Sperren(t.Context(), kennung); fehler != nil {
		t.Fatalf("Sperren: %v", fehler)
	}

	darf, fehler := werk.Darf(t.Context(), kennung, domain.ZweckPaarung)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("ein gesperrtes Gegenüber darf sich noch vorstellen")
	}

	if befund := werk.Entsperren(t.Context(), kennung); befund != nil {
		t.Fatalf("Entsperren: %v", befund)
	}

	// **Nach dem Entsperren ist es unbestätigt** — nicht so wie vorher.
	darf, fehler = werk.Darf(t.Context(), kennung, domain.ZweckPartie)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("das Entsperren gab die Freigabe zurück")
	}
}

// **Ein Fehler des Bestands kommt an jedem Weg zurück.**
//
// **An jedem**, weil jeder für sich verschluckt werden könnte — und ein
// verschluckter Fehler wäre hier eine Freigabe, die niemand gegeben hat.
func TestEinFehlerDesBestandsKommtAnJedemKontaktwegZurueck(t *testing.T) {
	t.Parallel()

	doppel := neuesKontaktdoppel()
	doppel.scheitert = errors.New("kein Bestand")

	werk := service.NeuesKontaktwerk(doppel, einePeeruhr)
	kennung, abdruck := eineFremdeKennung(t)

	_, listenfehler := werk.Liste(t.Context())
	_, darffehler := werk.Darf(t.Context(), kennung, domain.ZweckTeilen)

	for name, fehler := range map[string]error{
		"Liste":       listenfehler,
		"Darf":        darffehler,
		"Begegnet":    werk.Begegnet(t.Context(), kennung, "Wer"),
		"Bestaetigen": werk.Bestaetigen(t.Context(), kennung, abdruck),
		"Sperren":     werk.Sperren(t.Context(), kennung),
		"Entsperren":  werk.Entsperren(t.Context(), kennung),
		"Erlauben": werk.Erlauben(
			t.Context(), kennung, domain.ZweckTeilen),
		"Entziehen": werk.Entziehen(
			t.Context(), kennung, domain.ZweckTeilen),
		"NachRotation": werk.NachRotation(t.Context()),
	} {
		if fehler == nil {
			t.Errorf("%s schwieg", name)
		}
	}
}
