package service

import "testing"

// **Die GitHub-Adresse wird in die Archiv- und die Markenadresse übersetzt.**
//
// `codeload` liefert einen Stand als `tar.gz`, ohne dass wir einen Zugang
// brauchen; eine Freigabe wird über die API nachgeschlagen. Ohne diese
// Übersetzung müsste jemand beide Adressen selbst kennen.
//
// **Ein interner Test und kein Lauf gegen GitHub.** Ein Test, der das Netz
// braucht, prüft die Verfügbarkeit einer fremden Gegenstelle mit — und ist rot,
// wenn sie es nicht ist. Was hier zählt, ist die Übersetzung.
func TestDieGitHubAdresseWirdUebersetzt(t *testing.T) {
	faelle := map[string]struct {
		adresse string
		archiv  string
		marken  string
	}{
		"eine gewöhnliche Adresse": {
			adresse: "https://github.com/MagicSetEditorPacks/M15-Magic-Pack",
			archiv:  "https://codeload.github.com/MagicSetEditorPacks/M15-Magic-Pack/tar.gz",
			marken:  "https://api.github.com/repos/MagicSetEditorPacks/M15-Magic-Pack/git/ref/tags/",
		},
		"mit .git und Schrägstrich": {
			adresse: "https://github.com/MagicSetEditorPacks/M15-Magic-Pack.git/",
			archiv:  "https://codeload.github.com/MagicSetEditorPacks/M15-Magic-Pack/tar.gz",
			marken:  "https://api.github.com/repos/MagicSetEditorPacks/M15-Magic-Pack/git/ref/tags/",
		},

		// **Eine andere Gegenstelle wird nicht zu GitHub gemacht.** Wer eine
		// eigene Adresse angibt, meint sie.
		"etwas anderes als GitHub": {
			adresse: "https://beispiel.test/pakete/probe",
			archiv:  "https://beispiel.test/pakete/probe/tar.gz",
			marken:  "https://beispiel.test/pakete/probe/tags/",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if gefunden := archivadresse(fall.adresse); gefunden != fall.archiv {
				t.Errorf("Archiv: %q, erwartet %q", gefunden, fall.archiv)
			}

			if gefunden := markenadresse(fall.adresse); gefunden != fall.marken {
				t.Errorf("Marken: %q, erwartet %q", gefunden, fall.marken)
			}
		})
	}
}
