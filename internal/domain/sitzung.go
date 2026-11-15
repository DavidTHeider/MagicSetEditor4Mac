package domain

// **Die lokale Sitzung** — ID-02.AK-01
// (#1701),
// Marke `new`: MSE2 kennt keinen Nutzer.
//
// # Drei Zustände und acht Übergänge
//
// **Angemeldet** heißt: ein Mensch arbeitet, und die Anwendung erreicht die
// Daten seines Profils. **Gesperrt** heißt: derselbe Mensch, aber bis zum
// Entsperren erreicht keine Bindung ein Profil. **Abgemeldet** heißt:
// niemand — bis sich einer anmeldet.
//
// **Kein Zustand „registriert", keine E-Mail-Verifikation.** Ein Konto
// entsteht auf diesem Rechner; eine Adresse ist eine Kennung und kein
// nachgewiesener Kanal (`service/kontoanmeldung.go`).
//
// # Warum ein Kennwortwechsel nur angemeldet geht
//
// **Hinter der Sperre ginge er an ihr vorbei**: wer ein fremdes, gesperrtes
// Fenster vor sich hat, soll nicht dessen Kennwort ändern können. Wer sein
// Kennwort vergessen hat, nimmt die Wiederherstellung — mit einem Code, den
// er beim Einrichten des zweiten Faktors bekommen hat.

import "fmt"

// Sitzungszustand ist, wo die Sitzung steht.
type Sitzungszustand string

// Die Zustände.
const (
	SitzungAngemeldet Sitzungszustand = "angemeldet"
	SitzungGesperrt   Sitzungszustand = "gesperrt"
	SitzungAbgemeldet Sitzungszustand = "abgemeldet"
)

// Sitzungsereignis ist, was die Sitzung bewegt.
type Sitzungsereignis string

// Die Ereignisse.
const (
	SitzungAnmelden         Sitzungsereignis = "anmelden"
	SitzungEntsperren       Sitzungsereignis = "entsperren"
	SitzungSperren          Sitzungsereignis = "sperren"
	SitzungAbmelden         Sitzungsereignis = "abmelden"
	SitzungKennwortWechseln Sitzungsereignis = "kennwortWechseln"
	SitzungWiederherstellen Sitzungsereignis = "wiederherstellen"
)

// sitzungsuebergaenge sind die zulässigen Übergänge — alles andere ist ein
// Konflikt.
var sitzungsuebergaenge = map[Sitzungszustand]map[Sitzungsereignis]Sitzungszustand{
	// **Abgemeldet setzt die Wiederherstellung nur das Kennwort** —
	// angemeldet wird danach, mit Kennwort und zweitem Faktor (Zweitblick
	// über ID-02, Befund 3).
	SitzungAbgemeldet: {
		SitzungAnmelden:         SitzungAngemeldet,
		SitzungWiederherstellen: SitzungAbgemeldet,
	},
	SitzungAngemeldet: {
		SitzungSperren:          SitzungGesperrt,
		SitzungAbmelden:         SitzungAbgemeldet,
		SitzungKennwortWechseln: SitzungAngemeldet,
	},
	SitzungGesperrt: {
		SitzungEntsperren:       SitzungAngemeldet,
		SitzungWiederherstellen: SitzungAngemeldet,
		SitzungAbmelden:         SitzungAbgemeldet,
	},
}

// Nach gibt den Zustand nach einem Ereignis — oder einen Konflikt, und die
// Sitzung bleibt, wo sie war.
func (z Sitzungszustand) Nach(ereignis Sitzungsereignis) (Sitzungszustand, error) {
	nach, erlaubt := sitzungsuebergaenge[z][ereignis]
	if !erlaubt {
		return z, fmt.Errorf("%w: im Zustand »%s« geht »%s« nicht", ErrConflict, z, ereignis)
	}

	return nach, nil
}

// ErreichtDenBestand sagt, ob eine Bindung in diesem Zustand die Daten eines
// Profils erreicht.
func (z Sitzungszustand) ErreichtDenBestand() bool { return z == SitzungAngemeldet }
