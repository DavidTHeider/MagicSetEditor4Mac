package service

import (
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kontingente sind die Guthaben je Principal und das globale, mit
// Schutzreserve fürs Fenster — Werkstück O5 von Phase 5.2 (#1429).
//
// # Zwei Guthaben je Ruf
//
// Ein Maschinenzugang bucht bei **seinem** Guthaben und beim **globalen**.
// Das erste hält einen Principal von den anderen fern, das zweite hält alle
// zusammen von der Maschine fern. Das Fenster hat kein eigenes Guthaben —
// der Mensch ist kein Maschinenzugang —, bucht aber global mit, und **nur es
// darf in die Reserve**: was unter `Reserve` liegt, bekommt kein Principal,
// damit der Editor arbeitet, wenn ein Skript das Übrige aufbraucht.
//
// # Eine begrenzte Tabelle
//
// Ein Principal ist eine Zeichenkette — die Kennung eines Tokens, eine
// MCP-Sitzung. Wer beliebige vorzeigt, darf den Speicher nicht füllen:
// die Tabelle hält [HoechstensPrincipals] Einträge und verdrängt den am
// längsten nicht gesehenen. Der beginnt wieder voll, wenn er wiederkommt —
// ein Guthaben ist keinen Speicher wert, den ein Angreifer bestimmt.
type Kontingente struct {
	grenzen domain.Haushaltsgrenzen
	uhr     func() time.Time

	mu        sync.Mutex
	global    domain.Guthaben
	eintraege map[string]*principalguthaben
	// folge zählt jeden Zugriff — der kleinste Wert ist der am längsten
	// nicht gesehene Principal.
	folge uint64
}

// HoechstensPrincipals ist die Größe der Tabelle.
const HoechstensPrincipals = 256

type principalguthaben struct {
	guthaben domain.Guthaben
	zuletzt  uint64
}

// NeueKontingente baut die Guthaben — die Uhr wird eingesetzt (§16).
func NeueKontingente(grenzen domain.Haushaltsgrenzen, uhr func() time.Time) *Kontingente {
	return &Kontingente{
		grenzen:   grenzen,
		uhr:       uhr,
		global:    domain.NeuesGuthaben(grenzen.Global, uhr()),
		eintraege: make(map[string]*principalguthaben),
	}
}

// Umfang ist, wie viele Principals die Tabelle gerade führt.
func (k *Kontingente) Umfang() int {
	k.mu.Lock()
	defer k.mu.Unlock()

	return len(k.eintraege)
}

// Abbuchen nimmt die Kosten einer Klasse — beim Principal und global, oder
// bei keinem von beiden: erst wird geprüft, dann gebucht.
//
// Die Wartezeit ist die bis zum nächsten möglichen Ruf: die längere von
// beiden, wenn beide fehlen.
func (k *Kontingente) Abbuchen(principal string, klasse domain.Kostenklasse, fenster bool) (bool, time.Duration) {
	// **Abgeschaltet geht jeder Ruf durch** — und vor dem Schloss und der
	// Tabelle: wer nichts zählt, merkt sich niemanden (QA-08.AK-03, 0161).
	if k.grenzen.OhneKontingente {
		return true, 0
	}

	// **Was nichts kostet, geht immer durch** (#2003) — auch unter der
	// Reserve: sie schützt Guthaben, und ein Abbruch nimmt keines. Er ist die
	// eine Anfrage, die Last nimmt.
	kosten := klasse.Kosten()
	if kosten <= 0 {
		return true, 0
	}

	jetzt := k.uhr()

	k.mu.Lock()
	defer k.mu.Unlock()

	reserve := k.grenzen.Reserve
	if fenster {
		reserve = 0
	}

	globalOk, globalWarten := k.reicht(k.global, kosten, reserve, jetzt)

	if fenster {
		if !globalOk {
			return false, globalWarten
		}

		k.global.Abbuchen(kosten, jetzt)

		return true, 0
	}

	eintrag := k.eintragFuer(principal, jetzt)
	probe := eintrag.guthaben
	principalOk, principalWarten := probe.Abbuchen(kosten, jetzt)

	if !principalOk || !globalOk {
		return false, max(principalWarten, globalWarten)
	}

	eintrag.guthaben = probe
	k.global.Abbuchen(kosten, jetzt)

	return true, 0
}

// reicht sagt, ob ein Guthaben `kosten` hergibt, ohne unter `reserve` zu
// fallen — und wie lange es sonst dauert.
func (k *Kontingente) reicht(guthaben domain.Guthaben, kosten, reserve int, jetzt time.Time) (bool, time.Duration) {
	stand := guthaben.Stand(jetzt)
	fehlend := float64(kosten+reserve) - stand

	if fehlend <= 0 {
		return true, 0
	}

	return false, k.grenzen.Global.Dauer(fehlend)
}

// eintragFuer gibt das Guthaben eines Principals — frisch, wenn er neu oder
// verdrängt ist — und merkt den Zugriff.
//
// **Verdrängt wird durch Suchen, nicht über eine Liste**: die Tabelle hat
// höchstens [HoechstensPrincipals] Einträge, und ein Gang darüber je
// Verdrängung ist billiger als eine Struktur, die Typzusicherungen braucht.
func (k *Kontingente) eintragFuer(principal string, jetzt time.Time) *principalguthaben {
	k.folge++

	if eintrag, bekannt := k.eintraege[principal]; bekannt {
		eintrag.zuletzt = k.folge

		return eintrag
	}

	for len(k.eintraege) >= HoechstensPrincipals {
		delete(k.eintraege, aeltester(k.eintraege))
	}

	eintrag := &principalguthaben{guthaben: domain.NeuesGuthaben(k.grenzen.Principal, jetzt), zuletzt: k.folge}
	k.eintraege[principal] = eintrag

	return eintrag
}

// aeltester ist der Schlüssel mit dem kleinsten Zugriffszähler.
func aeltester(eintraege map[string]*principalguthaben) string {
	var (
		schluessel string
		kleinster  = ^uint64(0)
	)

	for kandidat, eintrag := range eintraege {
		if eintrag.zuletzt < kleinster {
			kleinster = eintrag.zuletzt
			schluessel = kandidat
		}
	}

	return schluessel
}
