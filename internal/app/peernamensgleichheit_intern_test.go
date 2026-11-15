package app

// **Zwei gleichnamige Gegenüber sind zu unterscheiden** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Zwei Sätze der Abnahme, und beide an der Naht**, weil dort die Antwort
// entsteht, die eine Fläche zeigt:
//
//   - „Zwei gleichnamige Peers sind unterscheidbar."
//   - „Alle Identitäts-/Handshake-Nachrichten und Exporte enthalten keine
//     privaten Schlüssel, Kennwörter, Zugriffstoken oder **lokale
//     Profil-IDs**."
//
// **Der zweite ist der leisere.** Eine Profilkennung ist eine kleine Zahl;
// sie fällt in einer Antwort nicht auf, und wer sie hinausgibt, sagt einem
// Fremden, wie dieser Rechner innen sortiert ist.

import (
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Zwei Gegenüber mit demselben Namen bleiben zwei.**
func TestZweiGleichnamigeGegenueberBleibenZwei(t *testing.T) {
	t.Parallel()

	anwendung, erste, ersterAbdruck := eineKontaktwelt(t)
	ctx := anwendung.imProfil(riegelkontext(t))

	zweiter, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	for _, kennung := range []peer.Kennung{erste, zweiter.Kennung()} {
		if antwort := anwendung.PeerkontaktAufnehmen(
			ctx, string(kennung), "Annas Rechner"); antwort.Fehler != nil {
			t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
		}
	}

	antwort := anwendung.Peerkontakte(ctx)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Kontakte) != 2 {
		t.Fatalf("es stehen %d darin", len(antwort.Kontakte))
	}

	// **Derselbe Name, zwei Kennungen, zwei Abdrücke.** Wäre der Name der
	// Schlüssel, stünde hier einer — und die Freigabe des einen gälte dem
	// anderen.
	if antwort.Kontakte[0].SelbstgewaehlterName !=
		antwort.Kontakte[1].SelbstgewaehlterName {
		t.Fatal("die Läufe messen nicht denselben Namen")
	}

	if antwort.Kontakte[0].Kennung == antwort.Kontakte[1].Kennung ||
		antwort.Kontakte[0].Fingerabdruck ==
			antwort.Kontakte[1].Fingerabdruck {
		t.Fatalf("die beiden sind nicht zu unterscheiden: %+v",
			antwort.Kontakte)
	}

	// **Und eine Freigabe gilt genau einem.**
	if antwort := anwendung.PeerkontaktBestaetigen(
		ctx, string(erste), ersterAbdruck); antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerzweckErlauben(
		ctx, string(erste), string(domain.ZweckHandel),
	); antwort.Fehler != nil {
		t.Fatalf("Erlauben: %+v", antwort.Fehler)
	}

	darf, fehler := anwendung.kontaktwerk().Darf(
		ctx, zweiter.Kennung(), domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("die Freigabe des einen gilt dem gleichnamigen anderen")
	}
}

// **Und keine Antwort trägt die lokale Profilkennung.**
//
// **Über den ganzen Weg**: die Identität, die Kontakte und die Zwecke — das
// sind die drei, die an einer Fläche stehen und in eine Nachricht wandern
// könnten.
func TestKeineAntwortTraegtDieLokaleProfilkennung(t *testing.T) {
	t.Parallel()

	// **Eine Kennung, die sich nicht zufällig ergibt.** `1` stünde in jeder
	// zweiten Zeichenkette; `4711` steht nur da, wo sie hingeschrieben wurde.
	// Das Profil dazu wird angelegt — eine Kennung ohne Zeile wiese der
	// Fremdschlüssel ab, und der Lauf prüfte dann etwas anderes.
	const profilID = 4711

	bestand := bestandFuerDenRiegel(t)

	if _, err := bestand.ExecContext(t.Context(),
		`INSERT INTO profile (id, nutzer_id, name) VALUES (?, 1, ?)`,
		profilID, "Vierzehnsiebenelf"); err != nil {
		t.Fatalf("Profil anlegen: %v", err)
	}

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung, abdruck := fremd.Kennung(), fremd.Fingerabdruck()
	anwendung := PeerflaecheUeber(bestand)

	anwendung.profil.Store(int64(profilID))

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(kennung), "Annas Rechner"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerkontaktBestaetigen(
		ctx, string(kennung), abdruck); antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	antwort := anwendung.PeerzweckErlauben(
		ctx, string(kennung), string(domain.ZweckTeilen))
	if antwort.Fehler != nil {
		t.Fatalf("Erlauben: %+v", antwort.Fehler)
	}

	gesucht := strconv.Itoa(profilID)

	for _, kontakt := range antwort.Kontakte {
		zusammen := kontakt.Kennung + kontakt.Fingerabdruck +
			kontakt.SelbstgewaehlterName + kontakt.EigenerName +
			kontakt.Stand + kontakt.Gesehen + strings.Join(kontakt.Rechte, "")

		if strings.Contains(zusammen, gesucht) {
			t.Errorf("ein Kontakt trägt die Profilkennung: %+v", kontakt)
		}
	}

	if strings.Contains(strings.Join(antwort.Zwecke, ""), gesucht) {
		t.Error("die Zwecke tragen die Profilkennung")
	}
}
