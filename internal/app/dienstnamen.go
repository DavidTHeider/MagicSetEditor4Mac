package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// EigeneDienstnamen sind die Namen, unter denen ein Dienst **jenseits der
// Schleife** erreichbar ist — seine Adressen und die Namen seines Zeugnisses.
//
// **Für beide Dienste an einer Stelle** (SEC-11,
// #2140):
// `mse-rest` und `mse-mcp` prüfen den `Host` einer Anfrage gegen diese Liste
// und gegen `127.0.0.1` und `localhost`. Bis dahin rechnete sie nur der
// Dienst, und der Werkzeugserver prüfte gar keinen.
//
// **Ein Befund hält nichts auf**: lässt sich die Horchadresse nicht lesen,
// gelten nur die Namen der Schleife — wer über die Netzadresse fragt, bekommt
// eine klare Absage statt eines Dienstes, der nicht startet.
func EigeneDienstnamen(
	ctx context.Context, bindung domain.Bindung, horchadresse string, zeugnisnamen []string,
) []string {
	if !bindung.VerlaesstDenRechner() {
		return nil
	}

	erreichbar, fehler := leitung.ErreichbarUnter(ctx, horchadresse)
	if fehler != nil {
		return nil
	}

	return append(ohnePort(append(erreichbar.Adressen, erreichbar.Namen...)), zeugnisnamen...)
}

// ohnePort nimmt den Port ab — `Host` vergleicht `rest.HerkunftTaugtUnter`
// und `mcp.herkunftTaugt` ohne ihn. Ein Eintrag ohne Port bleibt, wie er ist.
func ohnePort(eintraege []string) []string {
	ohne := make([]string, 0, len(eintraege))

	for _, eintrag := range eintraege {
		// **Über die Domäne und nicht über `net`**: `internal/app` lädt von
		// sich aus nichts und kennt das Netz nicht (`schichten/netz_test.go`).
		rechner, _, fehler := domain.ZerlegeAdresse(eintrag)
		if fehler != nil {
			rechner = eintrag
		}

		ohne = append(ohne, rechner)
	}

	return ohne
}
