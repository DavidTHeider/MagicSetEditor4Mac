// **Betrieb und Vertrauen** — OPS-07
// (#1724).
// Eine eigene Datei, weil `hilfe.ts` an die Zeilengrenze käme (§2); sie wird
// dort eingesetzt und kommt als **ein** Katalog an.
export const betrieb = {
  ueberschrift: 'Betrieb und Vertrauen',
  einleitung:
    'Was installiert ist, wo es liegt, was geprüft wurde und was nicht — aus dieser Anwendung selbst gelesen.',
  fassungName: 'Fassung',
  fassungSatz: '{fassung}, gebaut aus {stand}, dem Stand vom {gebaut}.',
  fassungVeraendert:
    'Der Bau enthielt Änderungen, die nicht eingecheckt waren.',
  datenName: 'Daten',
  datenSatz: '{zahl} Migrationen verbucht, zuletzt {letzte}.',
  datenOhne: 'Kein Bestand geöffnet.',
  signaturName: 'Signatur',
  signaturSatz:
    'Ad-hoc signiert, **nicht notarisiert**. Beim ersten Start weist macOS die Anwendung ab; freigeben lässt sie sich einmal unter Systemeinstellungen › Datenschutz & Sicherheit › **Dennoch öffnen**.',
  aktualisierungName: 'Aktualisierung',
  aktualisierung: {
    aus: 'Aus — diese Anwendung fragt nicht nach neuen Fassungen.',
    melden: 'Sie meldet eine neue Fassung von GitHub und holt nichts.',
    holen: 'Sie holt eine neue Fassung von GitHub und fragt vor dem Tausch.',
  },
  orteName: 'Wo alles liegt',
  orte: {
    bestand: 'Bestand',
    pakete: 'Vorlagenpakete',
    bilder: 'Eingesetzte Bilder',
    zwischenspeicher: 'Zwischenspeicher',
    protokoll: 'Protokoll',
    leistungsprofile: 'Leistungsprofile',
  },
  lizenzenName: 'Lizenzen',
  lizenzenSatz:
    'Der Quelltext steht unter der GNU General Public License 2.0 oder später. Die Programmteile anderer mit ihren Lizenzen führt `DRITTANBIETER.md` beim Quelltext; ein Vorlagenpaket bringt seine eigenen Bedingungen mit, und ein KI-Modell die seines Anbieters.',
  empfaengerName: 'Wohin Daten gehen',
  empfaengerSatz:
    'Nur zu den Diensten, die du einschaltest — welche es sind und was hinausgeht, steht unter **Hilfe › Was diese Anwendung über dich weiß**. Es gibt keinen Versand von Nutzungsdaten.',
  grenzenName: 'Sicherheitsgrenzen',
  grenzenSatz:
    'Dienst und Werkzeugserver lauschen, solange nichts anderes eingestellt ist, nur auf `127.0.0.1` und verlangen eine Marke. Wer einem unter **Einstellungen › Diagnose** eine Adresse im Netz gibt, bekommt ihn nur mit TLS, und die Seite sagt, dass er den Rechner verlässt; über stdio spricht der Werkzeugserver über seine eigenen Ströme. Der Bestand lässt sich verschlüsseln, unter **Einstellungen › Daten und Sicherung**.',
  kontaktName: 'Wenn etwas nicht geht',
  kontaktSatz:
    'Unter **Einstellungen › Diagnose** stehen die Diagnose, das Diagnosepaket und die Zahlen dieses Prozesses. Einen Fehlerbericht nimmt github.com/DavidTHeider/MagicSetEditor4Mac entgegen — hinaus geht nichts von selbst.',
  pruefstandName: 'Was geprüft ist',
  pruefstandSatz:
    'Geprüft heißt: ein Lauf hat es belegt — eine Probe, deren Beleg mit Commit beim Quelltext liegt, oder ein Test, den jede Pipeline fährt; veröffentlicht wird nur, was sie bestanden hat. Vorhanden heißt: gebaut, aber nie gegen das Original oder die Plattform gelaufen.',
  spalteBereich: 'Bereich',
  spalteGegenstand: 'Gegenstand',
  spalteStand: 'Stand',
  spalteBeleg: 'Beleg',
  belegIssue: 'Issue #{nummer}',
  bereiche: {
    plattform: 'Plattform',
    sprache: 'Sprachen',
    barrierefreiheit: 'Barrierefreiheit',
    regeln: 'Regeln',
    integration: 'Fremde Dienste',
  },
  staende: {
    geprueft: 'geprüft',
    vorhanden: 'vorhanden',
    beschrieben: 'beschrieben',
    blockiert: 'wartet auf einen Menschen',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS 15.3 bis 26',
    'macOS 15.2 auf Apple Silicon': 'macOS 15.2 auf Apple Silicon',
    'macOS 27 auf Apple Silicon': 'macOS 27 auf Apple Silicon',
    'Tastatur und Vorleser im WebKit': 'Tastatur und Vorleser im WebKit',
  },
}
