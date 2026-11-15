package mcp_test

// **Das Werkzeug für die Wege nach draußen** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier geprüft wird
//
//   - Es kommt durch, mit seinen Argumenten.
//   - **Und es gibt es nicht ohne Naht**: `nil` heißt kein Werkzeug. Der
//     Wächter über der Werkzeugliste sieht sonst einen Namen, hinter dem ein
//     Ruf abstürzt — eine Zusage, die keine ist (§10).
//   - **Es ist als lesend und idempotent geführt.** Die Adressen entstehen
//     auf diesem Rechner aus einer Kennung und einem Namen; zweimal gefragt
//     ist zweimal dieselbe Antwort, und daran kürzt ein vorsichtiges
//     Gegenüber seine Liste.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// linkprobe ist eine Naht, die mitschreibt.
type linkprobe struct {
	printing string
	name     string

	fehler error
}

func (l *linkprobe) Marktlinks(_ context.Context, printing, name string) (
	[]mcp.Marktlink, error,
) {
	l.printing, l.name = printing, name

	if l.fehler != nil {
		return nil, l.fehler
	}

	return []mcp.Marktlink{{
		Anbieter: "cardmarket", Art: "suche", Herkunft: "name",
		Adresse: "https://www.cardmarket.com/en/Magic/Products/Search?searchString=Bolt",
	}}, nil
}

// rufeMarktlinks sucht das Werkzeug und ruft es.
func rufeMarktlinks(t *testing.T, naht mcp.Linkauskunft, argumente string) (
	string, error,
) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Links: naht}) {
		if werkzeug.Name == "lesen.marktlinks" {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatal("das Werkzeug lesen.marktlinks steht nicht in der Liste")

	return "", nil
}

func TestDasMarktlinkwerkzeugKommtDurch(t *testing.T) {
	naht := &linkprobe{}

	aus, err := rufeMarktlinks(t, naht,
		`{"printing":"bolt-m10","name":"Lightning Bolt"}`)
	if err != nil {
		t.Fatalf("rufen: %v", err)
	}

	if naht.printing != "bolt-m10" || naht.name != "Lightning Bolt" {
		t.Errorf("gefragt wurde %q / %q", naht.printing, naht.name)
	}

	if !strings.Contains(aus, `"anbieter": "cardmarket"`) {
		t.Errorf("die Antwort nennt keinen Anbieter: %s", aus)
	}

	// **Die Art steht dabei** — ohne sie verspräche ein Suchlink, was ein
	// Produktlink hält.
	if !strings.Contains(aus, `"art": "suche"`) {
		t.Errorf("die Antwort nennt die Art nicht: %s", aus)
	}
}

// **Ein Fehler der Naht kommt an, ein kaputtes Argument endet davor.**
func TestDasMarktlinkwerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	naht := &linkprobe{fehler: errors.New("den Druck gibt es nicht")}

	if _, err := rufeMarktlinks(t, naht, `{"printing":"weg"}`); err == nil {
		t.Error("der Fehler wurde verschluckt")
	}

	naht = &linkprobe{}

	if _, err := rufeMarktlinks(t, naht, `kein JSON`); err == nil {
		t.Error("das kaputte Argument kam durch")
	}

	if naht.printing != "" {
		t.Errorf("die Naht wurde trotzdem gerufen: %q", naht.printing)
	}
}

// **Ohne Naht gibt es das Werkzeug nicht.**
//
// **Die Gegenprobe zu `rufeMarktlinks`**: es ist dieselbe Liste, und ein
// `nil` lässt sie schrumpfen statt einen Ruf anzubieten, der abstürzt.
func TestOhneLinknahtGibtEsDasMarktlinkwerkzeugNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.marktlinks" {
			t.Error("lesen.marktlinks steht da, obwohl keine Naht eingerichtet ist")
		}
	}
}

// **Es ist als lesend und idempotent geführt.**
func TestDasMarktlinkwerkzeugIstLesendUndIdempotent(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Links: &linkprobe{}}) {
		if werkzeug.Name != "lesen.marktlinks" {
			continue
		}

		if werkzeug.Art != mcp.Lesend {
			t.Errorf("es ist als %v geführt", werkzeug.Art)
		}

		if !werkzeug.Idempotent {
			t.Error("es ist nicht als idempotent geführt")
		}

		return
	}

	t.Fatal("das Werkzeug steht nicht in der Liste")
}
