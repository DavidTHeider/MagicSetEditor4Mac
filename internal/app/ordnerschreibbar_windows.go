package app

// schreibbar sagt auf Windows ja — beschrieben in `docs/PORTIERUNG.md`,
// nicht gebaut: dort gibt es kein `access(2)`, und die Rechte eines Ordners
// liest man über ACLs (§12).
func schreibbar(string) bool { return true }
