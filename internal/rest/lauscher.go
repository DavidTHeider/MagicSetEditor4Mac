// Package rest ist der Dienst auf der Rückschleife
// ([0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)).
//
// # Was hier liegt und was nicht
//
// **Ein Blatt** (AGENTS.md §2): dieses Paket kennt `internal/domain` und die
// Schnittstellen, die es hier braucht — sonst nichts aus dem Projekt. Die
// Anwendungsfälle spricht es über Schnittstellen an, die **hier** definiert
// sind; wer sie erfüllt, entscheidet `cmd/mse-rest`.
//
// **Keine Fachlogik.** Ein `if` über einen Domänenzustand gehört nicht
// hierher, und zwar nicht als Stilfrage: 0046 nennt genau das als eine der
// drei Bedingungen, unter denen die Entscheidung für diesen Dienst wieder
// fiele. Der Grund, warum ein vierter Zugang vertretbar ist, ist der **eine**
// Fachkern dahinter.
package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Lauschen macht einen Lauscher auf der Rückschleife auf.
//
// **Port 0 heißt: das Betriebssystem sucht einen freien.** Das ist der Weg für
// Tests und für einen Start ohne feste Angabe; welcher es wurde, sagt
// `Addr()`.
func Lauschen(ctx context.Context, port int) (net.Listener, error) {
	return LauschenMit(ctx,
		domain.Bindung{Adresse: domain.Schleifenadresse, Port: port}, nil)
}

// LauschenMit nimmt eine Bindung entgegen — und **jenseits der Schleife ein
// Zeugnis**.
//
// # Was sich hier am 2026-09-15 geändert hat
//
// **Bis N6 band dieser Dienst an genau eine Adresse und wies jede andere ab.**
// Das war die Zusage, um derentwillen
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
// überhaupt gegeben werden konnte, und sie steht in dessen Satz „er bindet
// nirgendwo sonst hin".
//
// **Dieser Halbsatz ist mit [0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)
// gefallen** — fortgeschrieben, nicht umgangen. Die Vorgabe bleibt die
// Schleife; was sich ändert, ist, dass sie eine **Vorgabe** ist.
//
// # Was an ihre Stelle tritt
//
// **Eine Bindung, die den Rechner verlässt, braucht ein Zeugnis.** Ohne eines
// endet der Ruf hier — und zwar **vor** dem Lauscher, damit es keinen
// Augenblick gibt, in dem der Port offen und ungesichert ist.
//
// **Was diese Funktion nicht hält, steht daneben**: den **Zugangszwang**. Ob
// eine Anfrage ein Token trägt, entscheidet [Dienst] und nicht der Lauscher —
// er sieht keine Anfrage. Dafür steht [Dienst.MitZugangszwang], und ein
// Wächter hält, dass beide zusammen gesetzt werden.
func LauschenMit(
	ctx context.Context, bindung domain.Bindung, zeugnis *tls.Config,
) (net.Listener, error) {
	if fehler := domain.PruefeBindung(bindung); fehler != nil {
		return nil, fehler
	}

	if bindung.VerlaesstDenRechner() && zeugnis == nil {
		return nil, fmt.Errorf(
			"%w: eine Bindung an %s verlässt diesen Rechner und braucht "+
				"deshalb TLS — ein Zugangstoken ginge sonst im Klartext "+
				"über das Netz",
			domain.ErrValidation, bindung.Wirkadresse())
	}

	// **Über `ListenConfig` und nicht über `net.Listen`**: nur so lässt sich
	// der Aufbau abbrechen. Ein Lauscher, der beim Start hängt, hinge sonst,
	// bis das Betriebssystem aufgibt.
	var aufbau net.ListenConfig

	roh, fehler := aufbau.Listen(ctx, "tcp", bindung.Adresstext())
	if fehler != nil {
		return nil, fehler //nolint:wrapcheck // der Aufrufer erwartet den Fehler des Lauschers
	}

	if zeugnis == nil {
		return roh, nil
	}

	// **Der TLS-Mantel sitzt um den Lauscher und nicht um den Server.** So
	// trägt jede Verbindung ihn, auch die, die ein anderer Weg annähme — und
	// `Addr()` sagt weiterhin, wo wirklich gelauscht wird.
	return tls.NewListener(roh, zeugnis), nil
}
