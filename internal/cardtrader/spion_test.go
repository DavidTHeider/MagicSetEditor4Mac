package cardtrader_test

// **Es wird gelesen und nichts geschrieben** — die Abnahme aus A4
// (#1443):
// „Server-Spion weist genau null POST/PUT/PATCH/DELETE an CardTrader nach".
//
// # Die Reichweite dieser Messung
//
// **Sie deckt jeden Ruf ab, den dieses Programm an CardTrader tun kann** — und
// das ist keine Behauptung, sondern folgt aus zwei Dingen zusammen: [Klient]
// hat genau vier öffentliche Rufe, und [cardtrader.Wege] ist die geschlossene
// Positivliste, gegen die der Spion prüft. Kommt ein fünfter dazu, fällt der
// Lauf über die Zahl der gesehenen Wege (§10).
//
// **Was sie nicht deckt**: was ein anderes Paket an dieselbe Adresse schickt.
// Dafür steht die Netzgrenze in `internal/schichten/netz_test.go`.

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
)

func TestKeinRufSchreibtBeiCardTrader(t *testing.T) {
	t.Parallel()

	spion := &wegspion{}

	server := httptest.NewServer(spion)
	t.Cleanup(server.Close)

	klient := klientFuer(server)

	// **Jeder öffentliche Ruf, den es gibt** — und jeder mit gültigen
	// Argumenten, damit er die Anfrage wirklich stellt.
	if _, fehler := klient.Spiele(t.Context()); fehler != nil {
		t.Fatalf("Spiele: %v", fehler)
	}

	if _, fehler := klient.Erweiterungen(t.Context()); fehler != nil {
		t.Fatalf("Erweiterungen: %v", fehler)
	}

	if _, _, fehler := klient.Blueprints(t.Context(), 8); fehler != nil {
		t.Fatalf("Blueprints: %v", fehler)
	}

	if _, _, fehler := klient.Angebote(t.Context(), 10050,
		cardtrader.Angebotsfilter{}); fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	// **Und die drei Kontowege** (T11, #1476) — sie stehen in derselben
	// Liste und sind aus demselben Holz.
	if _, fehler := klient.Konto(t.Context()); fehler != nil {
		t.Fatalf("Konto: %v", fehler)
	}

	if _, fehler := klient.EigeneAngebote(t.Context(), 8); fehler != nil {
		t.Fatalf("EigeneAngebote: %v", fehler)
	}

	if _, _, fehler := klient.Bestellungen(t.Context(), 1, 20); fehler != nil {
		t.Fatalf("Bestellungen: %v", fehler)
	}

	methoden, wege := spion.gesehen()

	// **Nur GET, und zwar in Zahlen**: ohne die Zahl wäre der Lauf auch dann
	// grün, wenn gar keine Anfrage angekommen wäre.
	if len(methoden) != 1 || methoden[0] != http.MethodGet {
		t.Errorf("gesehen wurden die Methoden %v, erwartet nur GET", methoden)
	}

	erwartet := cardtrader.Wege()
	slices.Sort(erwartet)

	if !slices.Equal(wege, erwartet) {
		t.Errorf("gesehen wurden die Wege %v, erwartet %v", wege, erwartet)
	}
}

// wegspion schreibt jede Methode und jeden Weg auf.
type wegspion struct {
	schloss  sync.Mutex
	methoden []string
	wege     []string
}

func (w *wegspion) ServeHTTP(schreiben http.ResponseWriter, gestellt *http.Request) {
	w.schloss.Lock()

	if !slices.Contains(w.methoden, gestellt.Method) {
		w.methoden = append(w.methoden, gestellt.Method)
	}

	if !slices.Contains(w.wege, gestellt.URL.Path) {
		w.wege = append(w.wege, gestellt.URL.Path)
	}

	w.schloss.Unlock()

	// **Der Marktplatz und die Kontoauskunft geben ein Objekt, die anderen
	// ein Array** — beides muss zerteilbar sein, sonst bricht der Ruf vor
	// der Zählung ab.
	if gestellt.URL.Path == "/marketplace/products" ||
		gestellt.URL.Path == "/info" {
		_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // der Spion

		return
	}

	_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // der Spion
}

// gesehen gibt Methoden und Wege, beide sortiert.
func (w *wegspion) gesehen() (methoden, wege []string) {
	w.schloss.Lock()
	defer w.schloss.Unlock()

	methoden = slices.Clone(w.methoden)
	wege = slices.Clone(w.wege)

	slices.Sort(methoden)
	slices.Sort(wege)

	return methoden, wege
}

// **Und die Gegenprobe zur Positivliste**: sie ist geschlossen.
//
// Ohne diesen Lauf sagte der Spion nur, dass die **heute gerufenen** Wege
// lesend sind — nicht, dass es keine anderen gibt. [cardtrader.Wege] ist die
// Liste, gegen die `Klient` baut; eine Adresse daneben gäbe es nicht.
// **Sieben seit T11** (#1476): vier zum Marktplatz, drei zum eigenen Konto.
func TestDiePositivlisteFuehrtSiebenLesendeWege(t *testing.T) {
	t.Parallel()

	wege := cardtrader.Wege()

	if len(wege) != 7 {
		t.Fatalf("die Positivliste führt %d Wege, erwartet 7: %v", len(wege), wege)
	}

	for _, weg := range wege {
		if weg == "" || weg[0] != '/' {
			t.Errorf("%q ist kein Weg", weg)
		}
	}
}
