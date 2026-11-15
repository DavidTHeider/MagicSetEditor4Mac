package rest_test

// **Kein Pfad kommt unter einem anderen Namen am Vertrag vorbei** —
// #2073.
// Der Dekodierer legt `"Ziel"` in das Feld `json:"ziel"`, und `quelle` beim
// Einbinden eines Pakets ist ein Verzeichnis, obwohl der Name nach keinem
// klingt. Beides kam ungeprüft an den Weg.

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitschreibender merkt jeden Pfad, nach dem er gefragt wird, und weist ihn
// ab — so zeigt ein Test, **dass** der Vertrag gefragt wurde, und nicht nur,
// dass irgendetwas ein `400` gab.
type mitschreibender struct {
	sperre  sync.Mutex
	gefragt []string
}

func (m *mitschreibender) urteil(_ context.Context, pfad string) error {
	m.sperre.Lock()
	defer m.sperre.Unlock()

	m.gefragt = append(m.gefragt, pfad)

	return domain.NewFieldError("pfad", domain.SchluesselValidierung, "»%s« liegt außerhalb", pfad)
}

func (m *mitschreibender) wurdeGefragt(pfad string) bool {
	m.sperre.Lock()
	defer m.sperre.Unlock()

	return slices.Contains(m.gefragt, pfad)
}

func TestKeinPfadKommtUnterEinemAnderenNamenDurch(t *testing.T) {
	t.Parallel()

	for name, ziel := range map[string]endpunkt{
		"Ziel groß geschrieben":   {methode: http.MethodPost, weg: "/sets/1/sicherung", rumpf: `{"Ziel":"/etc/fremd"}`},
		"PFAD in Großbuchstaben":  {methode: http.MethodPost, weg: "/sets", rumpf: `{"PFAD":"/etc/fremd"}`},
		"quelle beim Einbinden":   {methode: http.MethodPut, weg: "/pakete/eingebunden", rumpf: `{"quelle":"/etc/fremd"}`},
		"Quelle beim Ersetzen":    {methode: http.MethodPut, weg: "/pakete/ersetzt", rumpf: `{"Quelle":"/etc/fremd"}`},
		"quelle im Plan":          {methode: http.MethodGet, weg: "/pakete/plan?vorgang=einbinden&quelle=/etc/fremd"},
		"ORT beim Übernehmen":     {methode: http.MethodPost, weg: "/pakete/uebernommen", rumpf: `{"ORT":"/etc/fremd"}`},
		"quelle beim Indexaufbau": {methode: http.MethodPost, weg: "/index/aufbau", rumpf: `{"quelle":"/etc/fremd"}`},
		"ort in der Adresse auch": {methode: http.MethodPost, weg: "/pakete/uebernommen?ort=/etc/fremd", rumpf: `{}`},
	} {
		vertrag := &mitschreibender{}
		dienst := dienstMitLeeremSet(t).MitPfadvertrag(vertrag.urteil)

		antwort := ruf(t, dienst, ziel, gueltig)
		if !vertrag.wurdeGefragt("/etc/fremd") {
			t.Errorf("%s: der Vertrag wurde nicht gefragt — %d %s", name, antwort.Code, antwort.Body.String())

			continue
		}

		if antwort.Code != http.StatusBadRequest || alsBefund(t, antwort).Code != "validierung" {
			t.Errorf("%s: %d %s", name, antwort.Code, antwort.Body.String())
		}
	}
}

// **Und unter einem Namen, der dort keinen Pfad trägt, fragt er nicht** —
// `quelle` beim Einlesen eines CSV ist eine Bezeichnung, `ziel` eines
// Angriffs ein Spieler. Dass derselbe Vertrag zuhört, belegt der Test
// darüber.
//
// **Die Rümpfe scheitern danach am Dekodieren** (`"text":1`, `"art":1`): der
// Vertrag steht davor, und der Weg braucht so keine Naht, die diese
// Vorrichtung nicht hat.
func TestEinNameOhnePfadGehtAmVertragVorbei(t *testing.T) {
	t.Parallel()

	for name, ziel := range map[string]endpunkt{
		"quelle eines CSV": {methode: http.MethodPost, weg: "/sammlung/einlesen", rumpf: `{"text":1,"quelle":"/etc/fremd"}`},
		"ziel eines Angriffs": {
			methode: http.MethodPost, weg: "/partien/p/befehle",
			rumpf: `{"art":1,"nutzlast":{"angriffe":[{"kreatur":"k","ziel":"/etc/fremd"}]}}`,
		},
		"quelle eines Pakets in einem fremden Weg": {
			methode: http.MethodPost, weg: "/pakete/neu", rumpf: `{"quelle":"/etc/fremd"}`,
		},
	} {
		vertrag := &mitschreibender{}
		dienst := dienstMitLeeremSet(t).MitPfadvertrag(vertrag.urteil)

		ruf(t, dienst, ziel, gueltig)

		if vertrag.wurdeGefragt("/etc/fremd") {
			t.Errorf("%s: der Vertrag wurde gefragt", name)
		}
	}
}

// **Ein groß geschriebenes Ziel ist ein genanntes** — sonst legte der Dienst
// seinen Ausgang daneben, und weil der Dekodierer den letzten Schlüssel
// nimmt, schriebe der Bogen still dorthin statt an den Ort des Aufrufers.
func TestEinGrossGeschriebenesZielBleibt(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())

	mit := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`{"Ziel":"/tmp/eigen.pdf"}`))
	if mit.Code != http.StatusCreated || mit.Header().Get("Location") != "" || ausfuhr.gelegt != "/tmp/eigen.pdf" {
		t.Errorf("%d %v %q", mit.Code, mit.Header(), ausfuhr.gelegt)
	}
}
