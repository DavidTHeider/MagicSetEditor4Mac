package domain

// **Welche Einstellung bestimmt, wohin gesendet wird** — SEC-P4
// (#2000),
// Befund P4 in `docs/PENTEST.md`.
//
// # Warum es die Frage gibt
//
// **Wer die Adresse einer KI, das Gateway eines Anbieters oder den Proxy
// setzt, wählt das Ziel jedes späteren Rufs** — auch eines lesenden, den ein
// Token gehen darf. Ein Token reicht nie bis [WirkungExternSenden]
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)); bis #2000
// reichte es trotzdem bis zu der Einstellung, die das Ziel bestimmt. Wer sie
// setzt, braucht seitdem dieselbe Erlaubnis wie der, der sendet.
//
// # Warum am Schreiben und nicht am Wählen
//
// **Weil die Adresse im eigenen Netz liegen darf.** Ein Ollama auf einem
// zweiten Rechner (`http://gpu.intern:11434`) ist ein beschriebener Weg
// (`docs/NETZEIGENSCHAFTEN.md`), und eine Sperre am Ziel nähme ihn weg. Die
// Grundsperre am Klienten (`internal/netz`) gilt trotzdem — sie trifft
// nichts, was ein Dienst je braucht.
//
// # Was dazugehört
//
// **Jede Adresse, die ein Dienst tragen darf** ([Dienst.AdresseEinstellbar]),
// die der lokalen Laufzeit, und die drei Proxyangaben — auch
// [NetzProxyAus]: sie entscheidet, welche Anfrage am Proxy vorbeigeht.

// IstNetzweg sagt, ob eine Einstellung bestimmt, wohin gesendet wird.
func IstNetzweg(schluessel string) bool {
	switch schluessel {
	case KIAdresse, NetzProxyHTTP, NetzProxyHTTPS, NetzProxyAus:
		return true
	}

	for _, dienst := range Dienste() {
		if dienst.AdresseEinstellbar() && schluessel == DienstAdresse(dienst) {
			return true
		}
	}

	return false
}
