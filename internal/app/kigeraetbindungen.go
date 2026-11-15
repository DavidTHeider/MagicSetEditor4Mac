package app

import "context"

// KIGeraetemodellFuerSitzung schaltet Apples Gerätemodell für diese Sitzung
// ab oder wieder an — AI-05.AK-04
// (#2042).
// Nach einem Neustart gilt wieder die Einstellung `ki.geraetemodell`.
func (b *Bindungen) KIGeraetemodellFuerSitzung(ctx context.Context, an bool) KIGeraetSitzungAntwort {
	return b.app(ctx).KIGeraetemodellFuerSitzung(b.ruf(ctx), an)
}
