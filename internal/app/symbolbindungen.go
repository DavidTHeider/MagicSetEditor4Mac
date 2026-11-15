// Die Bindungen des **Symboleditors** —
// #890.
//
// **Eine eigene Datei**, wie `paketbindungen.go`: `bindungenkarten.go` führt
// die Kartenwege und liefe sonst über die Zeilengrenze (§2).

package app

import "context"

// Grundform gibt eine der vier Grundformen, mittig und in voller Größe.
func (b *Bindungen) Grundform(ctx context.Context, form SymbolgrundformDTO) SymbolAntwort {
	return b.app(ctx).Grundform(form)
}

// SymbolSetzen legt eine gezeichnete Zeichnung neben das Set.
//
// **Ohne Kontext** — es gibt nichts abzuwarten; der Grund steht an
// [App.SymbolSetzen].
func (b *Bindungen) SymbolSetzen(
	ctx context.Context,
	setID string, teile []SymbolteilDTO, unter string,
) SymbolAntwort {
	return b.app(ctx).SymbolSetzen(b.ruf(ctx), setID, teile, unter)
}

// Symbol liest eine abgelegte Zeichnung zurück.
func (b *Bindungen) Symbol(ctx context.Context, setID, name string) SymbolAntwort {
	return b.app(ctx).Symbol(b.ruf(ctx), setID, name)
}

// PunktEinfuegen setzt einen Punkt in die Mitte der Strecke dahinter.
func (b *Bindungen) PunktEinfuegen(
	ctx context.Context,
	teile []SymbolteilDTO, weise SymbolpunktweiseDTO,
) SymbolAntwort {
	return b.app(ctx).PunktEinfuegen(teile, weise)
}

// PunktLoeschen entfernt einen Punkt.
func (b *Bindungen) PunktLoeschen(
	ctx context.Context,
	teile []SymbolteilDTO, weise SymbolpunktweiseDTO,
) SymbolAntwort {
	return b.app(ctx).PunktLoeschen(teile, weise)
}

// GriffZiehen bewegt einen Bézierhandgriff.
func (b *Bindungen) GriffZiehen(
	ctx context.Context,
	teile []SymbolteilDTO, griff SymbolgriffDTO,
) SymbolAntwort {
	return b.app(ctx).GriffZiehen(teile, griff)
}

// SymmetrieAnlegen wandelt ein Teil in eine Symmetrie um
// (#1092).
func (b *Bindungen) SymmetrieAnlegen(ctx context.Context, teile []SymbolteilDTO, stelle int) SymbolAntwort {
	return b.app(ctx).SymmetrieAnlegen(teile, stelle)
}

// SymmetrieAufloesen macht aus einer Symmetrie ihre Kinder.
func (b *Bindungen) SymmetrieAufloesen(ctx context.Context, teile []SymbolteilDTO, stelle int) SymbolAntwort {
	return b.app(ctx).SymmetrieAufloesen(teile, stelle)
}

// SymmetrieSetzen ändert Art, Kopienzahl, Mitte oder Griff.
func (b *Bindungen) SymmetrieSetzen(
	ctx context.Context,
	teile []SymbolteilDTO, weise SymmetrieweiseDTO,
) SymbolAntwort {
	return b.app(ctx).SymmetrieSetzen(teile, weise)
}

// Symmetriekopien gibt die Züge, die eine Symmetrie zeichnet.
func (b *Bindungen) Symmetriekopien(
	ctx context.Context,
	teile []SymbolteilDTO, stelle int,
) SymmetriekopienAntwort {
	return b.app(ctx).Symmetriekopien(teile, stelle)
}

// TeileGruppieren fasst mehrere Teile zu einer Gruppe zusammen
// (#1168).
func (b *Bindungen) TeileGruppieren(
	ctx context.Context,
	teile []SymbolteilDTO, stellen []int,
) SymbolAntwort {
	return b.app(ctx).TeileGruppieren(teile, stellen)
}

// GruppeAufloesen setzt die Kinder einer Gruppe an deren Stelle.
func (b *Bindungen) GruppeAufloesen(
	ctx context.Context,
	teile []SymbolteilDTO, stellen []int,
) SymbolAntwort {
	return b.app(ctx).GruppeAufloesen(teile, stellen)
}

// TeilVerschieben ändert die Reihenfolge — und damit, was über was liegt.
func (b *Bindungen) TeilVerschieben(
	ctx context.Context,
	teile []SymbolteilDTO, von, nach int,
) SymbolAntwort {
	return b.app(ctx).TeilVerschieben(teile, von, nach)
}

// SymbolNachzeichnen fragt nach einer Bilddatei und zeichnet sie nach.
//
// **Der Dialog steckt hier und nicht in [App.SymbolNachzeichnen]** — dieselbe
// Trennung wie bei `BildSetzen`. Nur deshalb können Kommandozeile, MCP und
// Dienst dieselbe Fähigkeit anbieten (§16).
func (b *Bindungen) SymbolNachzeichnen(ctx context.Context) SymbolAntwort {
	ctx = b.ruf(ctx)

	// **Derselbe Filter wie bei `BildSetzen`.** Ein Format, das dort ginge und
	// hier nicht, wäre eine Falle — und umgekehrt.
	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:  DialogtexteFuer(b.app(ctx).sprache).BildOeffnen,
		Filter: DialogtexteFuer(b.app(ctx).sprache).BildFilter,
		Muster: "*.png;*.jpg;*.jpeg;*.gif",
	})
	if err != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	// **Ein Abbruch ist kein Fehler.** Wer den Dialog schließt, hat nichts
	// gewählt — eine Meldung dafür wäre eine Antwort auf eine Frage, die
	// niemand gestellt hat.
	if pfad == "" {
		return SymbolAntwort{Teile: []SymbolteilDTO{}}
	}

	return b.app(ctx).SymbolNachzeichnen(b.ruf(ctx), pfad)
}
