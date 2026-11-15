# 0128 — Identität bleibt lokal

**Stand:** entschieden am 2026-09-22 · ID-01
(#1700)

## Was dieser Vermerk tut

**Er entscheidet wenig Neues — er bindet Zusagen an Belege.** Was ID-01
verlangt, galt zum größten Teil schon seit Phase 4.12
([0066](0066-nutzer-profile-und-die-vier-entscheidungen.md),
[0074](0074-wer-arbeitet-steht-fest.md)) und Phase 5.4. Es stand aber
verteilt, und drei der Zusagen waren Verneinungen, die niemand festhielt:
kein Mailversand, kein Nachweisfeld am Nutzer, kein stiller Systemname.

## Vier Zusagen und woran sie hängen

| Zusage | Was gilt | Beleg |
| --- | --- | --- |
| **Nur lokale Nutzer und Profile** (AK-01) | Kein Konto bei einem Dienst dieses Programms, keine zentrale Registrierung, kein bevorzugter Onlinezugang. Nutzer und Profile stehen im Bestand | Einen Dienst dieses Programms gibt es nicht; jeder Netzzugriff steht in der Aufzählung von `internal/schichten/netz_test.go` und ist dort eine Entscheidung |
| **Wiederkehrende Gegenstellen sind bestätigt** (AK-01) | Ein `Peerkontakt` trägt Kennung, Fingerabdruck und einen Stand; `bestaetigt` heißt: der Fingerabdruck wurde verglichen, nicht eine Adresse | `internal/domain/peerkontakt.go`, `internal/peer`; **Gäste bleiben möglich**: `TestEinGastBleibtErSelbst` |
| **Eine Adresse ist eine Kennung, kein Nachweis** (AK-02) | Kein Versand, kein Verifikationsdienst, keine Wiederherstellung über das Netz. Der Rückweg sind zehn Wiederherstellungscodes | `TestKeinDienstVerschicktPost`, `TestEinNutzerTraegtKeinenNachweis` |
| **Ein Systemname ist kein Beweis** (AK-04) | Heute fließt weder der Name des Rechners noch eine Kontaktkarte ein. **ID-04** bringt die Kontakte, **mit Zustimmung** und nur als Vorbelegung eines Anzeigenamens | `TestKeinSystemnameFliesstStillEin` — ID-04 bekommt dort eine benannte Ausnahme |

## Vier Handlungen, die nicht dasselbe sind (AK-03)

| Handlung | Was sie bestätigt | Weg |
| --- | --- | --- |
| **Profil wechseln** | nichts — sie wechselt den **Arbeitskontext**. Wer arbeitet, bleibt, wer er ist ([0074](0074-wer-arbeitet-steht-fest.md)) | `ProfilWechseln`, `mse profil` |
| **Sich anmelden oder entsperren** | den **lokalen Zugang**: Kennwort und zweiter Faktor, Touch ID als **lokale Entsperrung** | `KontoAnmelden`, `NutzerWechseln`; seit ID-02 (#1701) dazu `Sperren`, `Entsperren`, `EntsperrenMitSensor`, `Abmelden`, `LokalAnmelden` — die Sitzung steht im Bestand und überlebt ⌘Q |
| **Eine Gegenstelle koppeln** | eine **Gegenstelle**: ihr Schlüssel, verglichen über den Fingerabdruck | `internal/peer`, Kontakte |
| **Sich bei einem fremden Dienst ausweisen** | ein **Konto dort** — bei einem Anmeldeanbieter (GitHub, Google, Microsoft, X) oder bei einem Datenanbieter mit eigenem Schlüssel (CardTrader, Scrydex, ein Sprachmodell; Scryfall braucht keinen). Es sagt nichts darüber, wer vor diesem Rechner sitzt | `Anmelden` für die Anmeldeanbieter, die Schlüssel der Datenanbieter im Schlüsselbund |

**Wer zwei davon an eine Geste hängt**, macht aus einer Handlung eine
andere, ohne dass es jemand sieht. Die Trennung zwischen Profil- und
Nutzerwechsel steht schon in [0074](0074-wer-arbeitet-steht-fest.md); hier
kommen Kopplung und Anbieteranmeldung dazu.

## Was die Wächter nicht halten

**Ob die Oberfläche eine Adresse irgendwo „bestätigt" nennt.** Das ist ein
Satz in einem Katalog, und den liest ein Mensch. **Und nicht, dass eine
Gegenstelle wirklich über ihren Schlüssel bestätigt wird**: das halten die
Läufe in `internal/peer` und `internal/service`, nicht
`internal/werkzeuge/lokaleidentitaet_test.go`.

## Verworfen

| Verworfen | Weil |
| --- | --- |
| Eine E-Mail-Bestätigung „für später" vorsehen | Ein Zustand, den nichts verlassen kann, ist keine Auskunft, sondern ein Vorwurf — derselbe Grund wie im Katalog der Fähigkeiten beim Konto |
| Den Anzeigenamen aus `os/user` vorbelegen | Still wäre es eine Übernahme ohne Zustimmung, und AK-04 verlangt sie. Wenn, dann mit ID-04 und einem ausdrücklichen Schritt |
