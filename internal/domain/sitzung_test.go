package domain_test

// **Die lokale Sitzung** — ID-02.AK-01
// (#1701).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJederUebergangDerSitzungIstBenannt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		von      domain.Sitzungszustand
		ereignis domain.Sitzungsereignis
		nach     domain.Sitzungszustand
	}{
		{domain.SitzungAbgemeldet, domain.SitzungAnmelden, domain.SitzungAngemeldet},
		// **Abgemeldet setzt die Wiederherstellung nur das Kennwort** — wer
		// danach arbeitet, meldet sich an (Zweitblick über ID-02, Befund 3).
		{domain.SitzungAbgemeldet, domain.SitzungWiederherstellen, domain.SitzungAbgemeldet},
		{domain.SitzungAngemeldet, domain.SitzungSperren, domain.SitzungGesperrt},
		{domain.SitzungAngemeldet, domain.SitzungAbmelden, domain.SitzungAbgemeldet},
		{domain.SitzungAngemeldet, domain.SitzungKennwortWechseln, domain.SitzungAngemeldet},
		{domain.SitzungGesperrt, domain.SitzungEntsperren, domain.SitzungAngemeldet},
		{domain.SitzungGesperrt, domain.SitzungWiederherstellen, domain.SitzungAngemeldet},
		{domain.SitzungGesperrt, domain.SitzungAbmelden, domain.SitzungAbgemeldet},
	} {
		nach, fehler := fall.von.Nach(fall.ereignis)
		if fehler != nil || nach != fall.nach {
			t.Errorf("%s --%s-->: %s, %v — erwartet %s", fall.von, fall.ereignis, nach, fehler, fall.nach)
		}
	}
}

// **Was nicht geht, ist ein Konflikt** — und die Sitzung bleibt, wo sie war.
func TestEinUnzulaessigerUebergangIstEinKonflikt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		von      domain.Sitzungszustand
		ereignis domain.Sitzungsereignis
	}{
		// **Gesperrt heißt: nichts anderes als entsperren, wiederherstellen,
		// abmelden** — ein Kennwortwechsel hinter der Sperre ginge an ihr
		// vorbei.
		{domain.SitzungGesperrt, domain.SitzungKennwortWechseln},
		{domain.SitzungGesperrt, domain.SitzungSperren},
		{domain.SitzungGesperrt, domain.SitzungAnmelden},
		{domain.SitzungAbgemeldet, domain.SitzungEntsperren},
		{domain.SitzungAbgemeldet, domain.SitzungSperren},
		{domain.SitzungAbgemeldet, domain.SitzungAbmelden},
		{domain.SitzungAbgemeldet, domain.SitzungKennwortWechseln},
		{domain.SitzungAngemeldet, domain.SitzungAnmelden},
		{domain.SitzungAngemeldet, domain.SitzungEntsperren},
		{domain.SitzungAngemeldet, domain.SitzungWiederherstellen},
	} {
		nach, fehler := fall.von.Nach(fall.ereignis)
		if !errors.Is(fehler, domain.ErrConflict) || nach != fall.von {
			t.Errorf("%s --%s-->: %s, %v", fall.von, fall.ereignis, nach, fehler)
		}
	}

	if _, fehler := domain.Sitzungszustand("irgendwie").Nach(domain.SitzungAnmelden); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("unbekannter Zustand: %v", fehler)
	}
}

// **Nur angemeldet erreicht die Daten eines Profils.**
func TestNurEineAngemeldeteSitzungErreichtDenBestand(t *testing.T) {
	t.Parallel()

	for zustand, erreicht := range map[domain.Sitzungszustand]bool{
		domain.SitzungAngemeldet: true, domain.SitzungGesperrt: false, domain.SitzungAbgemeldet: false,
	} {
		if zustand.ErreichtDenBestand() != erreicht {
			t.Errorf("%s erreicht den Bestand: %v", zustand, !erreicht)
		}
	}
}
