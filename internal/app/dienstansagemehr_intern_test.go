package app

// **Die Ansage mit mehreren Adressen** — G08
// (#2272).
//
// **Die erste Zeile bleibt, wie sie war** — „Dienst auf" mit der ersten
// Adresse —, jede weitere kommt als eigene Zeile dazu, und eine Adresse, auf
// der kein Lauscher entstand, als „Ohne". Wer nur die erste liest, liest
// weiter dasselbe (§16: erweiternd).

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieAnsageNenntJedeAdresseUndJedeFehlende(t *testing.T) {
	t.Parallel()

	gesagt := Dienstansage{
		MitTLS:   true,
		Adresse:  "127.0.0.1:8443",
		Weitere:  []string{"192.168.178.74:8443", "[fe80::1%en0]:8443"},
		Fehlende: []domain.Lauschbefund{{Adresse: "10.8.0.2", Grund: "listen tcp 10.8.0.2:8443: bind: can't assign requested address"}},
		Pfad:     "/mcp",
		Marke:    "geheim",
	}

	text := gesagt.Text()
	if erste, _, _ := strings.Cut(text, "\n"); erste != "Dienst auf https://127.0.0.1:8443/mcp" {
		t.Errorf("die erste Zeile ist %q", erste)
	}

	gelesen, fehler := leseAnsage(context.Background(), strings.NewReader(text), time.Second)
	if fehler != nil {
		t.Fatalf("leseAnsage: %v", fehler)
	}

	if gelesen.adresse != "https://127.0.0.1:8443/mcp" {
		t.Errorf("die erste Adresse kam als %q an", gelesen.adresse)
	}

	erwartet := []string{
		"https://127.0.0.1:8443/mcp", "https://192.168.178.74:8443/mcp", "https://[fe80::1%en0]:8443/mcp",
	}
	if !reflect.DeepEqual(gelesen.adressen, erwartet) {
		t.Errorf("die Adressen kamen als %v an", gelesen.adressen)
	}

	if len(gelesen.fehlende) != 1 || gelesen.fehlende[0] != gesagt.Fehlende[0] {
		t.Errorf("die fehlenden kamen als %+v an", gelesen.fehlende)
	}
}
