//go:build vmlauf

package app_test

// **Zwei Maschinen finden einander und sprechen gesichert** — SEC-22
// (#2163),
// die Abnahme zu ABN-01.AK-05.
//
// `zweimac_netzlauf_intern_test.go` stellt beide Enden auf **einen** Rechner.
// Dieser Lauf gehört auf **zwei**: je ein macOS-Gast, beide im selben Netz.
// Er läuft in beiden zugleich, mit `MSE_VM_ROLLE=a` und `b`:
//
//	go test -c -tags vmlauf -o vmlauf.test ./internal/app/
//	MSE_VM_ROLLE=a ./vmlauf.test -test.run ZweiMaschinen -test.v   # Gast A
//	MSE_VM_ROLLE=b ./vmlauf.test -test.run ZweiMaschinen -test.v   # Gast B
//
// Jede Seite kündigt sich über mDNS an (`entdeckung.Melde`), horcht mit TLS
// auf der eigenen Peeridentität (`leitung.Horche`), sucht die andere
// (`entdeckung.Suche`) und wählt sie an — festgenagelt auf die gefundene
// Kennung (`leitung.Waehle`). Belegt wird je Richtung:
//
//   - **der Fund**: Kennung, Name und Port der anderen Seite, und der
//     `.local`-Name ist ihr Fingerabdruck und **nicht** der Rechnername,
//   - **das Echo** über TLS 1.3 und eine **nichtleere Kanalbindung** aus
//     `ExportKeyingMaterial`.
//
// **Was er nicht belegt**: zwei physische Macs, ein WLAN, NAT oder einen
// Router zwischen zwei Netzen (Ausgangslage §9). Und er gibt die Erlaubnis
// *Lokales Netzwerk* nicht — fehlt sie, endet die Suche mit
// `EHOSTUNREACH`, und das ist ein roter Befund, kein grüner.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

const vmHafen = 47120

func TestZweiMaschinenFindenEinanderUndSprechenGesichert(t *testing.T) {
	rolle := os.Getenv("MSE_VM_ROLLE")
	gegenueber := map[string]string{"a": "b", "b": "a"}[rolle]

	if gegenueber == "" {
		t.Skip("MSE_VM_ROLLE ist weder a noch b — dieser Lauf gehört in zwei Gäste")
	}

	ctx, abbrechen := context.WithTimeout(t.Context(), 3*time.Minute)
	defer abbrechen()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("eine Peeridentität erzeugen: %v", fehler)
	}

	rechner, _ := os.Hostname() //nolint:errcheck // nur zum Vergleich mit dem Fund
	t.Logf("Rolle %s, Rechner %q, Kennung %s, Fingerabdruck %s", rolle, rechner, ich.Kennung(), ich.Fingerabdruck())

	horcher, fehler := leitung.Horche(ctx, fmt.Sprintf(":%d", vmHafen), ich)
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	defer func() { _ = horcher.Schliesse() }() //nolint:errcheck // beim Aufräumen

	// **Wer anruft, bekommt sein Echo** — und die Bindung seines Kanals.
	bedient := make(chan string, 4)

	go func() {
		for {
			eingang, nimmfehler := horcher.Nimm(ctx)
			if nimmfehler != nil {
				return
			}

			nachricht, empfangsfehler := eingang.Empfange(ctx)
			if empfangsfehler == nil {
				_ = eingang.Sende(ctx, nachricht) //nolint:errcheck // die Gegenseite prüft das Echo
				bedient <- fmt.Sprintf("von %s, Kanal %d Zeichen", eingang.Gegenstelle(), len(eingang.Kanal()))
			}

			_ = eingang.Schliesse() //nolint:errcheck // beim Aufräumen
		}
	}()

	anmeldung, fehler := entdeckung.Melde(entdeckung.Angebot{
		Kennung: ich.Kennung(), Name: "VM-" + rolle, Port: vmHafen,
	})
	if fehler != nil {
		t.Fatalf("ankündigen: %v — fehlt die Erlaubnis Lokales Netzwerk?", fehler)
	}

	defer func() { _ = anmeldung.Beende() }() //nolint:errcheck // beim Aufräumen

	fund := sucheDieAndere(ctx, t, "VM-"+gegenueber)

	t.Logf("gefunden: Kennung %s, Name %q, Wirt %s, Port %d, Wirtname %q", fund.Kennung, fund.Name, fund.Wirt, fund.Port, fund.Wirtname)

	if fund.Port != vmHafen {
		t.Errorf("der Port ist %d, angekündigt war %d", fund.Port, vmHafen)
	}

	abdruck := peer.Fingerabdruck(fund.Kennung)
	if !strings.Contains(fund.Wirtname, abdruck) {
		t.Errorf("der .local-Name %q trägt nicht den Fingerabdruck %q", fund.Wirtname, abdruck)
	}

	if rechner != "" && strings.Contains(strings.ToLower(fund.Wirtname), strings.ToLower(strings.TrimSuffix(rechner, ".local"))) {
		t.Errorf("der .local-Name %q verrät den Rechnernamen %q", fund.Wirtname, rechner)
	}

	gesichert, fehler := leitung.Waehle(ctx, fund.Adresse(), ich, fund.Kennung)
	if fehler != nil {
		t.Fatalf("die andere Seite unter %s anwählen: %v", fund.Adresse(), fehler)
	}

	defer func() { _ = gesichert.Schliesse() }() //nolint:errcheck // beim Aufräumen

	gruss := []byte("Gruß von " + rolle)
	if sendefehler := gesichert.Sende(ctx, gruss); sendefehler != nil {
		t.Fatalf("senden: %v", sendefehler)
	}

	echo, fehler := gesichert.Empfange(ctx)
	if fehler != nil || !bytes.Equal(echo, gruss) {
		t.Fatalf("das Echo: %q, %v", echo, fehler)
	}

	if gesichert.Kanal() == "" {
		t.Error("die Kanalbindung ist leer")
	}

	t.Logf("gewählt: Echo %q, Gegenstelle %s, Kanal %d Zeichen", echo, gesichert.Gegenstelle(), len(gesichert.Kanal()))

	// **Und die andere Richtung**: die andere Seite wählt hier an. Ohne
	// diesen Eingang hätte nur eine Richtung getragen.
	select {
	case eingang := <-bedient:
		t.Logf("bedient: %s", eingang)
	case <-ctx.Done():
		t.Fatal("die andere Seite hat hier nicht angerufen")
	}
}

// sucheDieAndere fragt, bis die andere Seite unter ihrem Namen auftaucht.
func sucheDieAndere(ctx context.Context, t *testing.T, name string) entdeckung.Fund {
	t.Helper()

	for runde := 1; ; runde++ {
		ergebnis, fehler := entdeckung.Suche(ctx, entdeckung.Frage{Frist: 3 * time.Second})
		if fehler != nil {
			t.Fatalf("suchen (Runde %d): %v — fehlt die Erlaubnis Lokales Netzwerk?", runde, fehler)
		}

		for _, fund := range ergebnis.Funde {
			if fund.Name == name {
				return fund
			}
		}

		t.Logf("Runde %d: %d Funde, %d verworfen, %s noch nicht dabei", runde, len(ergebnis.Funde), ergebnis.Verworfen, name)

		if ctx.Err() != nil {
			t.Fatalf("%s nicht gefunden", name)
		}
	}
}
