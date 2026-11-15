package domain

// **Woher ein Paket stammt und wer es braucht** — DOM-01.AK-01
// (#1696):
// „Herkunft, Version, Prüfsumme, Lizenz, Abhängigkeiten und lokale Änderungen
// sind am Objekt sichtbar und filterbar."
//
// # Warum eine Art und nicht nur die Herkunftsangabe
//
// **Weil man nach ihr filtert**, und eine Adresse ist kein Filter. Das
// Artboard `Pakete` zeichnet drei Knöpfe — mitgeliefert, bezogen, eigene; die
// vierte Art, **übernommen**, kommt aus dem Bestand: ein Paket, das aus einem
// Verzeichnis oder einer MSE2-Installation eingebunden wurde, hat eine
// Herkunft und keine Prüfsumme, und es ist weder bezogen noch selbst gemacht.
//
// Portierungsmarke **adapted**: MSE2 unterscheidet lokal und global
// installierte Pakete und solche, die ein Installer brachte
// (`PackageVersion::STATUS_LOCAL`, `STATUS_GLOBAL`, `STATUS_BLESSED`,
// `src/util/io/package_manager.hpp`, Fassung f9d9356 vom 2021-07-06). Hier
// gibt es kein globales Verzeichnis und keinen Installer, wohl aber einen
// Bezug an einem Stand; die Einteilung folgt dem, was dieser Bestand weiß.

import (
	"path/filepath"
	"slices"
	"strings"
)

// Paketherkunftsart sagt, woher ein Paket stammt.
type Paketherkunftsart string

const (
	// PaketMitgeliefert liegt außerhalb des verwalteten Verzeichnisses und
	// hat keine Herkunft — die Saat im Bündel.
	PaketMitgeliefert Paketherkunftsart = "mitgeliefert"

	// PaketBezogen kam über einen Bezug an einem Stand und trägt dessen
	// Prüfsumme.
	PaketBezogen Paketherkunftsart = "bezogen"

	// PaketUebernommen wurde aus einem Verzeichnis oder einer
	// MSE2-Installation eingebunden: eine Herkunft ohne Prüfsumme.
	PaketUebernommen Paketherkunftsart = "uebernommen"

	// PaketEigen liegt im verwalteten Verzeichnis ohne Herkunft — hier
	// angelegt oder von Hand hineingelegt.
	PaketEigen Paketherkunftsart = "eigen"
)

// HerkunftsartIn sagt, woher dieses Paket stammt — gemessen an dem, was der
// Katalog weiß, und an dem Verzeichnis, das diese Anwendung verwaltet.
//
// **Die Prüfsumme zuerst**: sie entsteht nur beim Bezug, und ein bezogenes
// Paket bleibt bezogen, gleich wo es liegt.
func (p Paket) HerkunftsartIn(verwaltet string) Paketherkunftsart {
	switch {
	case p.Pruefsumme != "":
		return PaketBezogen
	case p.Herkunft != "":
		return PaketUebernommen
	case liegtIn(p.Pfad, verwaltet):
		return PaketEigen
	}

	return PaketMitgeliefert
}

// Dateiname ist der Name, unter dem andere Pakete dieses in `depends on:`
// nennen — `magic.mse-game`.
func (p Paket) Dateiname() string { return p.Name + "." + p.Art }

// GebrauchtVon nennt die Pakete des Katalogs, die dieses brauchen —
// sortiert, damit zwei Blicke dieselbe Liste sehen.
func GebrauchtVon(katalog []Paket, gebraucht Paket) []string {
	name := gebraucht.Dateiname()
	von := []string{}

	for _, paket := range katalog {
		if slices.ContainsFunc(paket.Braucht, func(bedarf Bedarf) bool { return bedarf.Paket == name }) {
			von = append(von, paket.Dateiname())
		}
	}

	slices.Sort(von)

	return von
}

// liegtIn sagt, ob ein Pfad unter einem Verzeichnis liegt — **nicht nur mit
// demselben Anfang**: `/pakete-alt` liegt nicht in `/pakete`.
func liegtIn(pfad, verzeichnis string) bool {
	if verzeichnis == "" {
		return false
	}

	rest, err := filepath.Rel(verzeichnis, pfad)

	return err == nil && rest != ".." && !strings.HasPrefix(rest, ".."+string(filepath.Separator))
}
