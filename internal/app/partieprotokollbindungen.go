package app

import "context"

// **Das Protokoll und die Leitung, am Fenster** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum diese fünf zusammenstehen
//
// **Weil sie eine Frage teilen: was bleibt von einer Partie übrig?** Vier
// davon beantworten sie fachlich — wer darf es lesen, und wann ist es weg.
// Die fünfte fragt, wie es um die Leitung steht, und sie steht dabei, weil
// ein Mensch beides an derselben Stelle sucht: „läuft es, und was bleibt?"

// PartieOffenlegungen sagt, wer wem etwas freigegeben hat.
func (b *Bindungen) PartieOffenlegungen(
	ctx context.Context,
	kennung string,
) PartieoffenlegungAntwort {
	return b.app(ctx).PartieOffenlegungen(b.ruf(ctx), kennung)
}

// PartieOffenlegen trägt die Einwilligung eines Sitzes ein.
//
// **Die Liste ersetzt die alte** — wer einen Empfänger herausnimmt, schickt
// die Liste ohne ihn.
func (b *Bindungen) PartieOffenlegen(
	ctx context.Context,
	kennung, von string, an []string,
) PartieoffenlegungAntwort {
	return b.app(ctx).PartieOffenlegen(b.ruf(ctx), kennung, von, an)
}

// PartieOffenlegungZuruecknehmen schließt die Tür wieder.
func (b *Bindungen) PartieOffenlegungZuruecknehmen(
	ctx context.Context,
	kennung, von string,
) PartieoffenlegungAntwort {
	return b.app(ctx).PartieOffenlegungZuruecknehmen(b.ruf(ctx), kennung, von)
}

// PartieVergessen löscht das Protokoll einer beendeten Partie.
func (b *Bindungen) PartieVergessen(ctx context.Context, kennung string) PartieoffenlegungAntwort {
	return b.app(ctx).PartieVergessen(b.ruf(ctx), kennung)
}

// Partiediagnose sagt, wie es um die Leitungen dieses Tisches steht.
func (b *Bindungen) Partiediagnose(ctx context.Context) PartiediagnoseAntwort {
	return b.app(ctx).Partiediagnose(b.ruf(ctx))
}
