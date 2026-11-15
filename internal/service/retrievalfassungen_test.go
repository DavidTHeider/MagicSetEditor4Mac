package service_test

// **Ein Index nennt die Fassungen, unter denen er entstand** — ENG-06
// (#1731).
//
// # Ein Etikett und keine Prüfung
//
// **Nichts vergleicht sie beim Suchen.** Ein Index aus einer älteren
// Zerteilfassung bleibt aktiv, bis jemand neu aufbaut — `docs/FASSUNGEN.md`
// nennt das als Grenze. Was hier gehalten wird, ist nur, dass das Etikett
// stimmt: ein Manifest, das eine andere Fassung nennte als die, unter der
// gebaut wurde, wäre eine falsche Auskunft über den eigenen Stand.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinIndexNenntDieFassungenSeinesAufbaus(t *testing.T) {
	t.Parallel()

	verwaltung := eineVerwaltung(&retrievalablagedoppel{}, &schreiberdoppel{},
		dieQuelleMit("all-minilm"), &korpusoeffnerdoppel{})

	bericht, fehler := verwaltung.Aufbauen(t.Context(), service.Aufbauauftrag{
		Modell: "all-minilm", Quelle: "oracle.jsonl.gz",
	})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if manifest := bericht.Manifest; manifest.Schemafassung != domain.RetrievalschemaFassung ||
		manifest.Zerteilfassung != service.Zerteilfassung ||
		manifest.Indexfassung != domain.RetrievalindexFassung {
		t.Errorf("das Manifest nennt Schema %d, Zerteilung %q, Index %d",
			manifest.Schemafassung, manifest.Zerteilfassung, manifest.Indexfassung)
	}
}
