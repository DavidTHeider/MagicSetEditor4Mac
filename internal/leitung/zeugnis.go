package leitung

// **Womit sich eine Leitung ausweist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine eigene Datei**, weil `tls.go` mit ihr über die Zeilengrenze aus §2
// ginge. Der Schnitt folgt der Frage: dort, **wie** eine Leitung entsteht —
// hier, **womit** sich beide Seiten dabei ausweisen und woran sie einander
// erkennen.

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// pruefeGegenueber hält das vorgezeigte Zeugnis gegen die erwartete Kennung.
//
// **Eine leere Erwartung heißt: jede gültige Kennung.** So steht es auf der
// horchenden Seite — sie erfährt erst aus dem Handschlag, wer da ist. Geprüft
// wird trotzdem, dass überhaupt ein Ed25519-Schlüssel vorliegt: ohne ihn bände
// der Kanal an niemanden.
//
// # Warum `VerifyConnection` und nicht `VerifyPeerCertificate`
//
// **Weil das zweite bei einer wiederaufgenommenen Sitzung nicht gerufen
// wird.** TLS 1.3 nimmt über ein Ticket wieder auf, und dann prüft niemand
// mehr das Zeugnis — das Festnageln wäre still übergangen. `VerifyConnection`
// läuft bei **jedem** Handschlag. Angestrichen von gosec (G123), und der
// Befund war richtig.
//
// **merke** hält danach die Kennung des vorgezeigten Zeugnisses fest —
// SEC-16 (#2145):
// die Leitung trägt sie weiter, und der Handschlag darf ihr nicht
// widersprechen. Jede Verbindung bekommt ihre eigene Einstellung und damit
// ihren eigenen Platz; `VerifyConnection` läuft im Handschlag, also vor
// dem Lesen.
func pruefeGegenueber(erwartet peer.Kennung, merke *peer.Kennung) func(tls.ConnectionState) error {
	return func(stand tls.ConnectionState) error {
		if len(stand.PeerCertificates) == 0 {
			return fmt.Errorf(
				"%w: die Gegenstelle hat kein Zeugnis vorgezeigt",
				domain.ErrValidation)
		}

		oeffentlich, taugt := stand.PeerCertificates[0].PublicKey.(ed25519.PublicKey)
		if !taugt {
			return fmt.Errorf(
				"%w: das Zeugnis trägt einen %T und keinen Ed25519-Schlüssel",
				domain.ErrValidation, stand.PeerCertificates[0].PublicKey)
		}

		gesehen, fehler := peer.KennungFuer(oeffentlich)
		if fehler != nil {
			return fehler
		}

		if erwartet != "" && gesehen != erwartet {
			return fmt.Errorf(
				"%w: an dieser Adresse sitzt %s und nicht %s",
				domain.ErrValidation,
				peer.Fingerabdruck(gesehen), peer.Fingerabdruck(erwartet))
		}

		*merke = gesehen

		return nil
	}
}

// Zeugnisfrist ist, wie lange ein Zeugnis gilt.
//
// **Sie ist eine Formsache und keine Zusage.** Geprüft wird der Schlüssel,
// nicht die Gültigkeit — aber ein Zeugnis ohne Zeitraum weisen manche
// Gegenstellen ab, und ein Zeugnis wird hier je Verbindung frisch gebaut.
const Zeugnisfrist = 24 * time.Hour

// zeugnisFuer baut ein selbst unterschriebenes Zeugnis auf einer Identität.
//
// **Der Schlüssel bleibt, wo er ist.** `x509.CreateCertificate` nimmt einen
// [crypto.Signer] und ruft ihn; das Geheimnis geht dabei nirgendwohin.
//
// # Warum die drei Teile einzeln hereinkommen
//
// **Weil eine [peer.Identitaet] nie ein krummes Paar ist.** `CreateCertificate`
// weist ein Zeugnis zurück, dessen öffentlicher Schlüssel nicht zum Signierer
// gehört — und diesen Weg kann nur betreten, wer die beiden getrennt
// hereinreichen darf. Im Betrieb kommen sie aus derselben Identität.
func zeugnisFuer(
	kennung peer.Kennung, oeffentlich ed25519.PublicKey,
	signierer crypto.Signer,
) (tls.Certificate, error) {
	nummer, fehler := rand.Int(zufallsquelle, new(big.Int).Lsh(big.NewInt(1), 128))
	if fehler != nil {
		return tls.Certificate{}, domain.MitGrund(domain.ErrValidation, fehler,
			"eine Nummer für das Zeugnis ziehen")
	}

	jetzt := time.Now()

	entwurf := &x509.Certificate{
		SerialNumber: nummer,

		// **Der Name ist die Kennung**, und er wird nicht geprüft. Er steht
		// da, damit ein Mensch, der das Zeugnis ansieht, weiß, wessen es
		// ist.
		Subject:   pkix.Name{CommonName: string(kennung)},
		NotBefore: jetzt.Add(-Zeugnisfrist),
		NotAfter:  jetzt.Add(Zeugnisfrist),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	roh, fehler := x509.CreateCertificate(
		zufallsquelle, entwurf, entwurf, oeffentlich, signierer)
	if fehler != nil {
		return tls.Certificate{}, domain.MitGrund(domain.ErrValidation, fehler,
			"ein Zeugnis für diese Identität bauen")
	}

	return tls.Certificate{
		Certificate: [][]byte{roh},
		PrivateKey:  signierer,
	}, nil
}
