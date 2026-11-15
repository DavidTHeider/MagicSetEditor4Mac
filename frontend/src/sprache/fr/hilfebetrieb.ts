// **Operation and trust** — OPS-07
// (#1724).
export const betrieb = {
  ueberschrift: 'Exploitation et confiance',
  einleitung:
    'Ce qui est installé, où cela se trouve, ce qui a été vérifié et ce qui ne l’a pas été — lu dans cette application elle-même.',
  fassungName: 'Version',
  fassungSatz: '{fassung}, compilée à partir de {stand}, l’état du {gebaut}.',
  fassungVeraendert:
    'La compilation contenait des modifications qui n’étaient pas validées.',
  datenName: 'Données',
  datenSatz: '{zahl} migrations appliquées, la dernière : {letzte}.',
  datenOhne: 'Aucune bibliothèque ouverte.',
  signaturName: 'Signature',
  signaturSatz:
    'Signée ad hoc, **non notariée**. Au premier lancement, macOS refuse l’application ; tu peux l’autoriser une fois sous Réglages Système › Confidentialité et sécurité › **Ouvrir quand même**.',
  aktualisierungName: 'Mises à jour',
  aktualisierung: {
    aus: 'Désactivées — cette application ne demande pas de nouvelles versions.',
    melden:
      'Elle signale une nouvelle version depuis GitHub et ne récupère rien.',
    holen:
      'Elle récupère une nouvelle version depuis GitHub et demande avant de remplacer.',
  },
  orteName: 'Où tout se trouve',
  orte: {
    bestand: 'Bibliothèque',
    pakete: 'Paquets de modèles',
    bilder: 'Images insérées',
    zwischenspeicher: 'Caches',
    protokoll: 'Journal',
    leistungsprofile: 'Profils de performance',
  },
  lizenzenName: 'Licences',
  lizenzenSatz:
    'Le code source est placé sous la GNU General Public License 2.0 ou ultérieure. Les composants de tiers, avec leurs licences, sont listés dans `DRITTANBIETER.md` à côté du code source ; un paquet de modèles apporte ses propres conditions, et un modèle d’IA celles de son fournisseur.',
  empfaengerName: 'Où vont les données',
  empfaengerSatz:
    'Uniquement vers les services que tu actives — lesquels et ce qui part figurent sous **Aide › Ce que cette application sait de toi**. Aucune donnée d’utilisation n’est envoyée.',
  grenzenName: 'Limites de sécurité',
  grenzenSatz:
    'Le service et le serveur d’outils n’écoutent, tant que rien d’autre n’est réglé, que sur `127.0.0.1` et exigent un jeton. Si tu donnes à l’un d’eux une adresse sur le réseau sous **Réglages › Diagnostic**, il ne fonctionne qu’avec TLS, et la page indique qu’il quitte l’ordinateur ; via stdio, le serveur d’outils communique par ses propres flux. La bibliothèque peut être chiffrée, sous **Réglages › Données et sauvegarde**.',
  kontaktName: 'Quand quelque chose ne fonctionne pas',
  kontaktSatz:
    'Sous **Réglages › Diagnostic** se trouvent le diagnostic, le paquet de diagnostic et les chiffres de ce processus. Les rapports d’erreur sont reçus par github.com/DavidTHeider/MagicSetEditor4Mac — rien ne part de soi-même.',
  pruefstandName: 'Ce qui est vérifié',
  pruefstandSatz:
    'Vérifié signifie : une exécution l’a prouvé — un essai dont la preuve, avec son commit, se trouve auprès du code source, ou un test que chaque pipeline exécute ; seul ce qui l’a réussi est publié. Présent signifie : construit, mais jamais exécuté face à l’original ou à la plateforme.',
  spalteBereich: 'Domaine',
  spalteGegenstand: 'Objet',
  spalteStand: 'Statut',
  spalteBeleg: 'Preuve',
  belegIssue: 'ticket #{nummer}',
  bereiche: {
    plattform: 'Plateforme',
    sprache: 'Langues',
    barrierefreiheit: 'Accessibilité',
    regeln: 'Règles',
    integration: 'Services tiers',
  },
  staende: {
    geprueft: 'vérifié',
    vorhanden: 'présent',
    beschrieben: 'décrit',
    blockiert: 'attend une personne',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS 15.3 à 26',
    'macOS 15.2 auf Apple Silicon': 'macOS 15.2 sur Apple Silicon',
    'macOS 27 auf Apple Silicon': 'macOS 27 sur Apple Silicon',
    'Tastatur und Vorleser im WebKit': 'Clavier et lecteur d’écran dans WebKit',
  },
}
