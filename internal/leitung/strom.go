package leitung

// **Eine Leitung über einen Strom, den jemand anderes gebaut hat** —
// Werkstück N3 von Phase 5.4
// (#1456).
//
// # Warum es das gibt
//
// **Weil tailcat einen `net.Conn` gibt und keine Adresse.** [Waehle] und
// [Horche] bauen ihre Verbindung selbst — über ein fremdes Netz kommt sie von
// woanders, durch einen WireGuard-Tunnel, den ein anderes Blatt aufgemacht
// hat.
//
// **Alles darüber bleibt dasselbe**: TLS 1.3, das Festnageln auf die erwartete
// Kennung, die Kanalbindung aus `ExportKeyingMaterial`, die Rahmung. Eine
// zweite Fassung davon wäre die Stelle, an der zwei Wege auseinanderlaufen —
// und die eine wäre dann geprüft und die andere nicht.
//
// # Warum `Waehle` und `Horche` jetzt darüber liegen
//
// **Damit es wirklich derselbe Weg ist und nicht zweimal derselbe Text.**
// Wer hier etwas ändert, ändert es für beide Netze; wer es nur für eines
// änderte, hätte genau den Bruch gebaut, den dieses Paket verhindern soll.

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// UeberStrom sichert einen fertigen Strom von der wählenden Seite.
//
// **Die erwartete Kennung ist Pflicht**, wie bei [Waehle]: wer wählt, weiß,
// wen er erreichen will. Ohne sie bekäme ein Fremder an derselben Leitung
// dieselbe Sitzung.
//
// **Der Strom gehört danach der Leitung.** Scheitert der Handschlag, wird er
// geschlossen; glückt er, schließt ihn [Leitung.Schliesse].
func UeberStrom(
	ctx context.Context, strom net.Conn,
	ich peer.Identitaet, erwartet peer.Kennung,
) (*Leitung, error) {
	if erwartet == "" {
		return nil, fmt.Errorf(
			"%w: ohne erwartete Kennung wird nicht gewählt",
			domain.ErrValidation)
	}

	zeugnis, fehler := zeugnisFuer(
		ich.Kennung(), ich.Oeffentlich(), ich.Signierer())
	if fehler != nil {
		return nil, fehler
	}

	ctx, abbrechen := context.WithTimeout(ctx, Handschlagfrist)
	defer abbrechen()

	var gegenstelle peer.Kennung

	gesichert := tls.Client(strom, &tls.Config{
		Certificates: []tls.Certificate{zeugnis},

		// **Nicht „unsicher", sondern „anders geprüft"** — der Grund steht
		// bei [Waehle], und er ist hier derselbe.
		InsecureSkipVerify: true, //#nosec G402 -- siehe VerifyConnection

		VerifyConnection: pruefeGegenueber(erwartet, &gegenstelle),
		MinVersion:       tls.VersionTLS13,
	})

	return schliesseHandschlagAb(ctx, gesichert, strom, &gegenstelle)
}

// NimmStrom sichert einen fertigen Strom von der horchenden Seite.
//
// **Ohne erwartete Kennung**, wie bei [Horcher.Nimm]: wer horcht, erfährt erst
// aus dem Handschlag, wer da ist, und was der darf, sagen die Kontakte.
func NimmStrom(
	ctx context.Context, strom net.Conn, ich peer.Identitaet,
) (*Leitung, error) {
	zeugnis, fehler := zeugnisFuer(
		ich.Kennung(), ich.Oeffentlich(), ich.Signierer())
	if fehler != nil {
		return nil, fehler
	}

	ctx, abbrechen := context.WithTimeout(ctx, Handschlagfrist)
	defer abbrechen()

	var gegenstelle peer.Kennung

	return schliesseHandschlagAb(ctx, tls.Server(strom, serverwerk(zeugnis, &gegenstelle)), strom, &gegenstelle)
}

// serverwerk ist die TLS-Einstellung der horchenden Seite.
//
// **An einer Stelle**, weil [Horcher.Nimm] und [NimmStrom] dieselbe brauchen —
// und eine zweite Fassung liefe beim ersten Zusatz auseinander.
func serverwerk(zeugnis tls.Certificate, gegenstelle *peer.Kennung) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{zeugnis},

		// **Jedes Gegenüber zeigt ein Zeugnis vor.** Ohne diese Zeile käme
		// die Gegenseite ohne Schlüssel durch, und der Kanal bände an
		// niemanden.
		ClientAuth: tls.RequireAnyClientCert,

		// **Und wer es ist, wird hier nicht entschieden.** Das sagt der
		// Handschlag; was er darf, sagen die Kontakte. **Festgehalten wird
		// es trotzdem** (SEC-16): der Handschlag darf dem Zeugnis nicht
		// widersprechen ([service.Peersitzungen]).
		InsecureSkipVerify: true, //#nosec G402 -- siehe VerifyConnection

		VerifyConnection: pruefeGegenueber("", gegenstelle),
		MinVersion:       tls.VersionTLS13,
	}
}
