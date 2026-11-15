package domain_test

// **„Nur lokal" schaltet Schalter ab, keine Angaben** — Folgebefund F2 des
// Lastlaufs QA-08
// (#2004).
//
// Unter `dienst.` stehen drei Arten von Schlüsseln: der **Schalter** eines
// Dienstes (`dienst.openai`), die **Adresse** dahinter
// (`dienst.openai.adresse`, #1540) und die vier **Pfade** des Dienstzeugnisses
// (`dienst.tls.*`, NET-03). Nur die erste Art sagt „darf hinaus" oder „darf
// nicht" — und nur sie kann „nur lokal" auf `aus` legen. Bis zum 2026-10-02
// traf die Schranke alle drei, und ein Dienst jenseits der Schleife suchte
// danach ein Zertifikat namens „aus".

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// angabenUnterDienst sind die Schlüssel unter `dienst.`, die nichts schalten,
// mit einem Wert, den es für sie geben darf.
func angabenUnterDienst() map[string]string {
	angaben := map[string]string{
		domain.DienstTLSZertifikat: "/etc/mse/dienst.pem",
		domain.DienstTLSSchluessel: "/etc/mse/schluessel.pem",
		domain.DienstTLSKette:      "/etc/mse/kette.pem",
		domain.DienstTLSAnker:      "/etc/mse/anker.pem",
	}

	for _, dienst := range []domain.Dienst{
		domain.DienstAnthropic, domain.DienstOpenAI, domain.DienstXAI,
	} {
		angaben[domain.DienstAdresse(dienst)] = "http://127.0.0.1:8099"
	}

	return angaben
}

func TestNurLokalLaesstDieAngabenUnterEinemDienstStehen(t *testing.T) {
	t.Parallel()

	profil := angabenUnterDienst()
	stand := domain.WirksamerStand(profil,
		map[string]string{domain.RichtlinieNurLokal: domain.NurLokalAn})

	for schluessel, wert := range profil {
		if stand[schluessel].Wirksam != wert {
			t.Errorf("%q gilt unter »nur lokal« als %q, eingetragen war %q",
				schluessel, stand[schluessel].Wirksam, wert)
		}

		if stand[schluessel].Herkunft != domain.GeltungProfil {
			t.Errorf("%q nennt die Herkunft %q — eingetragen hat sie das Profil",
				schluessel, stand[schluessel].Herkunft)
		}
	}
}

// **Die Regel, benannt**: „nur lokal" legt genau die Schalter aus
// [domain.Dienste] auf `aus` — keinen Schlüssel mehr und keinen weniger.
//
// **Über die ganze Reihe und nicht über eine Liste daneben.** Ein neuer
// Schlüssel unter `dienst.` landet hier von selbst, und ist er kein Schalter,
// muss er seinen Wert behalten.
func TestNurLokalTrifftGenauDieDienstschalter(t *testing.T) {
	t.Parallel()

	schalter := make([]string, 0, len(domain.Dienste()))
	for _, dienst := range domain.Dienste() {
		schalter = append(schalter, domain.DienstEinstellung(dienst))
	}

	profil := angabenUnterDienst()
	for _, schluessel := range schalter {
		profil[schluessel] = domain.DienstAn
	}

	stand := domain.WirksamerStand(profil,
		map[string]string{domain.RichtlinieNurLokal: domain.NurLokalAn})

	for _, schluessel := range domain.Einstellungsreihe() {
		if !strings.HasPrefix(schluessel, "dienst.") {
			continue
		}

		abgeschaltet := stand[schluessel].Wirksam == domain.DienstAus
		if istSchalter := slices.Contains(schalter, schluessel); abgeschaltet != istSchalter {
			t.Errorf("%q: abgeschaltet %t, ist ein Schalter %t (gilt %q)",
				schluessel, abgeschaltet, istSchalter, stand[schluessel].Wirksam)
		}
	}
}
