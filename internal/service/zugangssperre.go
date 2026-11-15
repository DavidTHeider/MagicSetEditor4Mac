package service

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zugangssperre ist die Sperre je vorgezeigtem Nachweis — Werkstück O5 von
// Phase 5.2 (#1429).
//
// # Je Nachweis, nicht je Rechner
//
// Bis O5 zählte REST Fehlversuche in **einer** Drossel: sechzig falsche
// Marken, und der richtige Aufrufer war für eine Minute draußen — ohne dass
// der Angreifer je etwas Gültiges gebraucht hätte. Hier hängt die Sperre am
// Nachweis, den jemand vorzeigt: wer `mse_falsch` rät, sperrt `mse_falsch`,
// und wer `mse_richtig` hat, merkt nichts.
//
// # Nichts Klartext, nichts Unbegrenztes
//
// Der Schlüssel ist die SHA-256 des Nachweises — ein Klartext-Token hat im
// Speicher nichts verloren, auch kein falsches. Und die Tabelle hält
// [HoechstensNachweise] Einträge: wer Zufall vorzeigt, füllt sie nicht, er
// verdrängt nur seine eigenen früheren Versuche.
type Zugangssperre struct {
	uhr func() time.Time

	mu    sync.Mutex
	lagen map[string]*nachweislage
	folge uint64
}

// HoechstensNachweise ist die Größe der Tabelle.
const HoechstensNachweise = 1024

type nachweislage struct {
	lage    domain.Sperrlage
	zuletzt uint64
}

// NeueZugangssperre baut die Tabelle — die Uhr wird eingesetzt (§16).
func NeueZugangssperre(uhr func() time.Time) *Zugangssperre {
	return &Zugangssperre{uhr: uhr, lagen: make(map[string]*nachweislage)}
}

// Umfang ist, wie viele Nachweise die Tabelle gerade führt.
func (z *Zugangssperre) Umfang() int {
	z.mu.Lock()
	defer z.mu.Unlock()

	return len(z.lagen)
}

// Gesperrt sagt, ob dieser Nachweis warten muss — und wie lange noch.
func (z *Zugangssperre) Gesperrt(nachweis string) (bool, time.Duration) {
	z.mu.Lock()
	defer z.mu.Unlock()

	eintrag, bekannt := z.lagen[schluesselFuer(nachweis)]
	if !bekannt {
		return false, 0
	}

	return eintrag.lage.Gesperrt(z.uhr())
}

// Fehlgeschlagen zählt einen Fehlversuch dieses Nachweises.
func (z *Zugangssperre) Fehlgeschlagen(nachweis string) {
	z.mu.Lock()
	defer z.mu.Unlock()

	eintrag := z.eintragFuer(schluesselFuer(nachweis))
	eintrag.lage = eintrag.lage.Fehlgeschlagen(z.uhr())
}

// Gelungen löscht die Lage eines Nachweises — er hat sich bewiesen.
func (z *Zugangssperre) Gelungen(nachweis string) {
	z.mu.Lock()
	defer z.mu.Unlock()

	delete(z.lagen, schluesselFuer(nachweis))
}

// eintragFuer gibt die Lage eines Schlüssels — neu, wenn er fehlt — und
// merkt den Zugriff; der am längsten nicht gesehene fällt, wenn die Tabelle
// voll ist (gesucht, nicht verkettet — wie bei den Kontingenten).
func (z *Zugangssperre) eintragFuer(schluessel string) *nachweislage {
	z.folge++

	if eintrag, bekannt := z.lagen[schluessel]; bekannt {
		eintrag.zuletzt = z.folge

		return eintrag
	}

	for len(z.lagen) >= HoechstensNachweise {
		delete(z.lagen, aeltesterNachweis(z.lagen))
	}

	eintrag := &nachweislage{zuletzt: z.folge}
	z.lagen[schluessel] = eintrag

	return eintrag
}

// aeltesterNachweis ist der Schlüssel mit dem kleinsten Zugriffszähler.
func aeltesterNachweis(lagen map[string]*nachweislage) string {
	var (
		schluessel string
		kleinster  = ^uint64(0)
	)

	for kandidat, eintrag := range lagen {
		if eintrag.zuletzt < kleinster {
			kleinster = eintrag.zuletzt
			schluessel = kandidat
		}
	}

	return schluessel
}

// schluesselFuer ist die Prüfsumme des Nachweises.
func schluesselFuer(nachweis string) string {
	summe := sha256.Sum256([]byte(nachweis))

	return hex.EncodeToString(summe[:])
}
