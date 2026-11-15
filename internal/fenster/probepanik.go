package fenster

import "strings"

// nennt sagt, ob die Liste der Naht den Rückruf `was` nennt — ganz, nicht
// als Teil eines Namens.
func nennt(liste, was string) bool {
	if was == "" {
		return false
	}

	for _, eintrag := range strings.Split(liste, "|") {
		if strings.TrimSpace(eintrag) == was {
			return true
		}
	}

	return false
}
