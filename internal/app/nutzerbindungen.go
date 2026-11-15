package app

import "context"

// **Die Wails-Bindungen für Nutzer und Profile** — Werkstück 1 von Phase 4.12
// (#1351).
//
// **Ohne Kontext, weil Wails ihn selbst mitbringt** — dieselbe Form wie jede
// andere Bindung.

// Nutzer gibt alle Nutzer samt ihren Profilen.
func (b *Bindungen) Nutzer(ctx context.Context) NutzerAntwort {
	return b.frei().Nutzer(b.frei().rufkontext(ctx))
}

// NutzerAnlegen legt einen Nutzer samt seinem ersten Profil an.
func (b *Bindungen) NutzerAnlegen(ctx context.Context, name, bild string) EinNutzerAntwort {
	return b.app(ctx).NutzerAnlegen(b.ruf(ctx), name, bild)
}

// ProfilAnlegen legt ein weiteres Profil für einen Nutzer an.
func (b *Bindungen) ProfilAnlegen(ctx context.Context, nutzerID, name string) EinProfilAntwort {
	return b.app(ctx).ProfilAnlegen(b.ruf(ctx), nutzerID, name)
}

// GezeigtesProfil sagt, in welchem Profil dieses Fenster arbeitet.
func (b *Bindungen) GezeigtesProfil(ctx context.Context) EinProfilAntwort {
	return b.app(ctx).GezeigtesProfil(b.ruf(ctx))
}

// ProfilWechseln stellt das Profil um, in dem dieses Fenster arbeitet.
//
// **Der Nachweis ist der zweite Faktor** — Werkstück 7 (#1351): ein Code aus
// der Authenticator-App, ein Wiederherstellungscode, oder **leer** für Touch
// ID. Wer keinen Faktor eingerichtet hat, gibt nichts an und wird nicht
// gefragt.
func (b *Bindungen) ProfilWechseln(ctx context.Context, profilID, nachweis string) EinProfilAntwort {
	return b.app(ctx).ProfilWechseln(b.ruf(ctx), profilID, nachweis)
}

// Anbieter nennt die Anbieter, über die man sich anmelden kann.
func (b *Bindungen) Anbieter(ctx context.Context) AnbieterAntwort {
	return b.app(ctx).Anbieter(b.ruf(ctx))
}

// Anmelden führt den Anmeldefluss über einen Anbieter.
func (b *Bindungen) Anmelden(ctx context.Context, anbieter string) EinNutzerAntwort {
	return b.app(ctx).Anmelden(b.ruf(ctx), anbieter)
}

// Anbieterkonten gibt die Konten eines Nutzers.
func (b *Bindungen) Anbieterkonten(ctx context.Context, nutzerID string) KontenAntwort {
	return b.app(ctx).Anbieterkonten(b.ruf(ctx), nutzerID)
}

// KontoLoesen nimmt eine fremde Kennung von einem Nutzer.
func (b *Bindungen) KontoLoesen(ctx context.Context, nutzerID, anbieter string) SchreibAntwort {
	return b.app(ctx).KontoLoesen(b.ruf(ctx), nutzerID, anbieter)
}

// NutzerAendern schreibt die Angaben eines Nutzers fort.
//
// **Ein Verbund und keine elf Argumente** — Werkstück 8 (#1351). Er trägt
// die **Fassung**, die der Rufer gelesen hat; ohne sie ginge eine gleichzeitige
// Änderung an einer anderen Fläche verloren (§7, Regel 2).
func (b *Bindungen) NutzerAendern(ctx context.Context, angaben NutzerangabenDTO) EinNutzerAntwort {
	return b.app(ctx).NutzerAendern(b.ruf(ctx), angaben)
}

// ProfilAendern schreibt die Angaben eines Profils fort.
//
// **Ein Verbund seit Werkstück 9** (#1351) — dieselbe Bauform wie bei
// [Bindungen.NutzerAendern] und aus demselben Grund. Er trägt die
// **Fassung**, die der Rufer gelesen hat.
func (b *Bindungen) ProfilAendern(ctx context.Context, angaben ProfilangabenDTO) EinProfilAntwort {
	return b.app(ctx).ProfilAendern(b.ruf(ctx), angaben)
}

// NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
func (b *Bindungen) NutzerLoeschen(ctx context.Context, nutzerID, name string) SchreibAntwort {
	return b.app(ctx).NutzerLoeschen(b.ruf(ctx), nutzerID, name)
}

// ProfilLoeschen nimmt ein Profil heraus.
func (b *Bindungen) ProfilLoeschen(ctx context.Context, profilID, name string) SchreibAntwort {
	return b.app(ctx).ProfilLoeschen(b.ruf(ctx), profilID, name)
}

// KontoEintragen legt ein Konto unter einer Adresse oder einer Rufnummer an.
//
// **Das Kennwort geht hier hinein und nie wieder heraus** — Werkstück 6
// (#1351). [NutzerDTO] hat kein Feld dafür, und die Seite bekommt es nie
// zurück; wer es vergessen hat, setzt ein neues, statt das alte zu lesen.
func (b *Bindungen) KontoEintragen(ctx context.Context, kennung, kennwort, name string) EinNutzerAntwort {
	return b.app(ctx).KontoEintragen(b.ruf(ctx), kennung, kennwort, name)
}

// KontoAnmelden lässt einen Nutzer unter seiner Kennung herein.
func (b *Bindungen) KontoAnmelden(ctx context.Context, kennung, kennwort, nachweis string) EinNutzerAntwort {
	return b.app(ctx).KontoAnmelden(b.ruf(ctx), kennung, kennwort, nachweis)
}

// KontoAnonym legt ein Konto ohne jede Angabe an.
func (b *Bindungen) KontoAnonym(ctx context.Context) EinNutzerAntwort {
	return b.app(ctx).KontoAnonym(b.ruf(ctx))
}

// Faktoren gibt die zweiten Faktoren eines Nutzers.
//
// **Samt der Auskunft, ob dieser Rechner Touch ID kann** — Werkstück 7
// (#1351). Beides in einer Antwort, weil die Seite beides nebeneinander zeigt
// (§7, Regel 2).
func (b *Bindungen) Faktoren(ctx context.Context, nutzerID string) FaktorenAntwort {
	return b.app(ctx).Faktoren(b.ruf(ctx), nutzerID)
}

// FaktorEinrichten zieht ein TOTP-Geheimnis.
//
// **Was zurückkommt, ist einmal zu sehen**: die Adresse für den Scanner, das
// Geheimnis zum Abtippen und zehn Wiederherstellungscodes. Danach steht
// nichts davon mehr in einer Antwort.
func (b *Bindungen) FaktorEinrichten(ctx context.Context, nutzerID string) Einrichtungsantwort {
	return b.app(ctx).FaktorEinrichten(b.ruf(ctx), nutzerID)
}

// FaktorBestaetigen trägt den TOTP-Faktor ein, wenn ein Code aus ihm stimmt.
func (b *Bindungen) FaktorBestaetigen(ctx context.Context, nutzerID, code string) SchreibAntwort {
	return b.app(ctx).FaktorBestaetigen(b.ruf(ctx), nutzerID, code)
}

// BiometrieEinrichten trägt Touch ID oder Face ID als zweiten Faktor ein.
func (b *Bindungen) BiometrieEinrichten(ctx context.Context, nutzerID string) SchreibAntwort {
	return b.app(ctx).BiometrieEinrichten(b.ruf(ctx), nutzerID)
}

// FaktorEntfernen nimmt einen zweiten Faktor heraus.
func (b *Bindungen) FaktorEntfernen(ctx context.Context, nutzerID, art string) SchreibAntwort {
	return b.app(ctx).FaktorEntfernen(b.ruf(ctx), nutzerID, art)
}

// Nutzerbilder gibt zu jedem Nutzer mit Bild einen Daten-URI.
//
// **Nur am Fenster** — Werkstück 8 (#1351). Ein Daten-URI ist die
// **Darstellung** einer Liste und kein Weg zu einer Auskunft; an der
// Kommandozeile wäre er eine Zeichenkette, die niemand ansehen kann.
func (b *Bindungen) Nutzerbilder(ctx context.Context) Nutzerbildantwort {
	return b.app(ctx).Nutzerbilder(b.ruf(ctx))
}

// NutzerbildWaehlen fragt nach einer Bilddatei und legt sie ab.
//
// **Der Dialog gehört hierher und nicht ins Frontend** — derselbe Grund wie
// bei [Bindungen.BildSetzen]: ein Fenster bekäme einen Namen und keinen Pfad.
func (b *Bindungen) NutzerbildWaehlen(ctx context.Context, nutzerID, fassung string) EinNutzerAntwort {
	return b.app(ctx).NutzerbildWaehlen(b.ruf(ctx), nutzerID, fassung)
}

// **`NutzerbildSetzen` steht hier nicht**, obwohl es die Anwendung gibt.
//
// **Weil das Fenster keinen Rufer dafür hätte.** Es kommt über den Dialog
// herein, und der ist [Bindungen.NutzerbildWaehlen]; die Pfadform bedient die
// drei Flächen ohne Fenster über [Nutzerflaeche]. Eine Bindung, die niemand
// ruft, ist eine Zusage ohne Einlösung (§2) — `BildAusDatei` hat einen Rufer,
// nämlich das Ziehen und Ablegen auf ein Kartenfeld; für ein Profilbild gibt
// es das noch nicht.

// NutzerbildEntfernen nimmt das Bild eines Nutzers wieder weg.
func (b *Bindungen) NutzerbildEntfernen(ctx context.Context, nutzerID, fassung string) EinNutzerAntwort {
	return b.app(ctx).NutzerbildEntfernen(b.ruf(ctx), nutzerID, fassung)
}

// Nutzereinstellungen gibt die Einstellungen eines Menschen.
//
// **Neben `Einstellungen` und nicht darin** — Werkstück 8 (#1351): jene
// gelten für das **Profil**, in dem gearbeitet wird, diese für den
// **Menschen**, der sie in jedes seiner Profile mitnimmt.
func (b *Bindungen) Nutzereinstellungen(ctx context.Context, nutzerID string) NutzereinstellungenAntwort {
	return b.app(ctx).Nutzereinstellungen(b.ruf(ctx), nutzerID)
}

// NutzereinstellungSetzen schreibt eine fort und gibt alle zurück.
//
// **Alle, weil das Fenster sonst zweimal fragen müsste** (§7, Regel 2).
func (b *Bindungen) NutzereinstellungSetzen(
	ctx context.Context,
	nutzerID, schluessel, wert string,
) NutzereinstellungenAntwort {
	return b.app(ctx).NutzereinstellungSetzen(b.ruf(ctx), nutzerID, schluessel, wert)
}

// PersonendatenSichern fragt nach einem Ziel und schreibt die Auskunft.
//
// **Der Dialog gehört hierher und nicht ins Frontend** — ein Fenster bekäme
// einen Namen und keinen Pfad. Ein Abbruch ist kein Fehler.
func (b *Bindungen) PersonendatenSichern(ctx context.Context, nutzerID string) SchreibAntwort {
	return b.app(ctx).PersonendatenSichern(b.ruf(ctx), nutzerID)
}

// ProfilbadgeWaehlen fragt nach einer Bilddatei und legt sie als Abzeichen ab.
//
// **Das Abzeichen gehört dem Profil, nicht dem Menschen** — Werkstück 9
// (#1351): es liegt rechts unten **über** dem Nutzerfoto und zeigt die Rolle,
// in der er gerade arbeitet.
func (b *Bindungen) ProfilbadgeWaehlen(ctx context.Context, profilID, fassung string) EinProfilAntwort {
	return b.app(ctx).ProfilbadgeWaehlen(b.ruf(ctx), profilID, fassung)
}

// ProfilbadgeEntfernen nimmt das Abzeichen eines Profils wieder weg.
func (b *Bindungen) ProfilbadgeEntfernen(ctx context.Context, profilID, fassung string) EinProfilAntwort {
	return b.app(ctx).ProfilbadgeEntfernen(b.ruf(ctx), profilID, fassung)
}
