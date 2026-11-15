package mcp_test

// Das Doppel der Einschätzung — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// **Es antwortet mit `null`-Listen**, und das ist der Zweck: der Weg, auf dem
// daraus leere werden, wird sonst nicht betreten.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func (*preisprobe) KIEinschaetzung(
	_ context.Context, bezug mcp.KIEinschaetzungsbezug,
) (mcp.KIEinschaetzungssicht, error) {
	return mcp.KIEinschaetzungssicht{
		Art: bezug.Art, Satz: "Das steht in 702.19b und in [B1].",
		Belegt: false,
		Vorbehalt: "Das ist die Einschätzung eines Modells und keine " +
			"amtliche Regelauskunft.",

		// **Ein Befund und nicht keiner**: eine Antwort, an der nichts zu
		// beanstanden ist, ließe den Weg ungeprüft, auf dem ein erfundenes
		// Zitat sichtbar wird — und der ist der Zweck dieses Werkzeugs.
		Befunde: []mcp.KIBelegbefundsicht{{
			Art: "regelnummer", Stelle: "702.19b",
			Meldung: "Diese Regelnummer steht in keinem Beleg.",
		}},
		Einschraenkungen: []string{"Nur die mitgegebenen Stellen."},
		Belege: []mcp.KIBelegsichtQuelle{{
			Kennung: "B1", Quelle: "regelwerk",
			Titel: "509.1 Blocks erklären", Quelldatum: "2026-06-07",
		}},
		Korpusstaende: []mcp.KIKorpusstandsicht{{
			Quelle: "regelwerk", Revision: "2026-06-07",
		}},
		Anbieter: bezug.Anbieter, Modell: bezug.Alias,
	}, nil
}

func (stolpernderPreis) KIEinschaetzung(
	context.Context, mcp.KIEinschaetzungsbezug,
) (mcp.KIEinschaetzungssicht, error) {
	return mcp.KIEinschaetzungssicht{}, errGestolpertMCP
}

// **Und das Doppel, das nichts füllt** — es gibt jede Liste als `null`.
func (nilprobe) KIEinschaetzung(
	context.Context, mcp.KIEinschaetzungsbezug,
) (mcp.KIEinschaetzungssicht, error) {
	return mcp.KIEinschaetzungssicht{}, nil
}

// **Und das Doppel der Balance** — Werkstück A13.2b desselben Issues.
//
// Es füllt den Verbund und lässt **jede Liste darin `null`**: der Weg, auf
// dem daraus leere werden, liegt eine Ebene tiefer als bei den anderen und
// wäre sonst unbetreten.
type balanceprobe struct{ nilprobe }

func (balanceprobe) KIEinschaetzung(
	_ context.Context, bezug mcp.KIEinschaetzungsbezug,
) (mcp.KIEinschaetzungssicht, error) {
	return mcp.KIEinschaetzungssicht{
		Art:  bezug.Art,
		Satz: "Sie steht zwischen den beiden.",
		Balance: &mcp.KIBalancesicht{
			Strategie: "hybrid, 20 Treffer, kein Reranker",
			Format:    bezug.Format,
		},
	}, nil
}
