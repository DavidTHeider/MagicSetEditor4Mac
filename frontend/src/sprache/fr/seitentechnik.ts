// The texts of the **technical page** — operating state, effect counter and
// the three tools without a window.
//
// Split out of `seiten.ts` when the file grew past 400 lines with the effect
// counter (work item O9, #1433).

import { betrieb, leistungsprofil, zwischenspeicher } from './seitenbetrieb'
import { dienstzeugnis, fernwege, schnittstellen } from './seitentechnikdienst'
import { einladung } from './einladung'
import { gegenueber } from './gegenueber'

export const technik = {
  betrieb,
  dienstzeugnis,
  einladung,
  gegenueber,
  fernwege,
  schnittstellen,
  leistungsprofil,
  zwischenspeicher,
  gesundheit: 'État de fonctionnement',
  gesundheitErklaerung:
    'Indique si ce processus est vivant et s’il peut travailler. Un service externe qui ne répond pas ne le rend pas indisponible — la modification continue sans réseau.',
  lebt: 'Le processus répond.',
  bereit: 'Il peut travailler.',
  nichtBereit: 'Il ne peut pas travailler pour le moment.',
  grundBestandZu: 'Aucune bibliothèque n’est ouverte.',
  grundWanderungLaeuft: 'La bibliothèque migre vers un nouveau schéma.',
  grundBestandUnlesbar: 'La bibliothèque ne répond pas.',
  fremdeDienste: 'Services externes',
  erreichbar: 'joignable',
  nichtErreichbar: 'injoignable',
  metriken: 'Effets depuis le démarrage',
  metrikenErklaerung:
    'Ce que ce processus a produit vers l’extérieur — par interface, classe d’effet et issue. Seuls les effets contrôlés sont comptés ; la simple lecture n’apparaît pas ici.',
  metrikenGrenze:
    'Au plus {grenze} lignes — elles ne grandissent pas avec le nombre de cartes, de sets ou de profils.',
  metrikenLeer: 'Rien n’a encore eu d’effet vers l’extérieur.',
  spalteFlaeche: 'Interface',
  spalteWirkung: 'Effet',
  spalteAusgang: 'Issue',
  spalteAnzahl: 'Nombre',
  flaechen: {
    fenster: 'Fenêtre',
    kommandozeile: 'Ligne de commande',
    mcp: 'MCP',
    dienst: 'Démarrer le service pour scripts',
  },
  wirkungen: {
    lokalLesen: 'lecture locale',
    lokalSchreiben: 'écriture locale',
    externSenden: 'envoi vers l’extérieur',
    externSchreiben: 'modification à l’extérieur',
    verbindlich: 'engageant',
  },
  ausgaenge: { gewaehrt: 'accordé', abgewiesen: 'refusé' },
  paket: {
    ueberschrift: 'Lot de diagnostic',
    erklaerung:
      'Les mêmes informations en plusieurs parties, sous forme de fichier avec sommes de contrôle — pour un rapport qui demande plus qu’un texte. Il est créé là où tu l’indiques, et ne va nulle part de lui-même.',
    inhalt:
      'Il contient la version, la plateforme, la bibliothèque, les paquets, les compteurs du processus, des lignes de journal et des traces de plantage — ni cartes, ni images, ni collection, ni identifiants.',
    zusammenstellen: 'Assembler',
    sichern: 'Enregistrer le lot',
    loeschen: 'Supprimer le lot',
    groesse: '{bytes} sur au plus {grenze}.',
    zeitraum: 'Les lignes de journal vont de {von} à {bis}.',
    ausgelassen: 'Omis — {teil} :',
    teile: {
      fassung: 'Version et plateforme',
      bestand: 'Emplacement et état de la bibliothèque',
      pakete: 'Paquets de modèles installés',
      metriken: 'Compteurs de ce processus',
      protokoll: 'Lignes de journal',
    },
    spalteTeil: 'Partie',
    spalteDatei: 'Fichier',
    spalteZeilen: 'Lignes',
    spalteBytes: 'Taille',
    spalteMitnehmen: 'Inclure',
    mitnehmen: 'Inclure {teil} dans le lot',
  },
  diagnose: {
    ueberschrift: 'Diagnostic pour un rapport d’erreur',
    erklaerung:
      'Quand quelque chose ne fonctionne pas, ce texte aide à chercher. Il est affiché ici — le transmettre relève de toi.',
    inhalt:
      'Il contient la version, la machine, la base de données, le journal et les paquets de modèles — pas de textes de cartes, pas de valeurs de champs, pas de contenu de sets. En détail : la version, le type de machine, la compilation dont provient le programme, le contrat de liaison entre la fenêtre et le programme, l’emplacement de la base de données, l’état du schéma, le nombre de lignes orphelines, l’emplacement du journal, tes paquets de modèles et les 200 dernières lignes du journal. Ton dossier personnel apparaît sous la forme « ~ ».',
    anzeigen: 'Afficher le diagnostic',
    kopieren: 'Copier',
    kopiert: 'Copié',
    kopierfehler:
      'La copie a échoué — la fenêtre n’a pas l’autorisation d’accéder au presse-papiers. Sélectionne le texte et copie-le à la main.',
    sichern: 'Enregistrer comme fichier',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: 'Rechercher dans les lignes',
    zeilenGezeigt: '{gezeigt} sur {alle} lignes',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: 'Technique et diagnostic',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'Réseau et connectivité',

  ueberschrift: 'Outils sans fenêtre',
  erklaerung:
    'Cette application fournit trois programmes qui fonctionnent sans fenêtre. mse pour la ligne de commande, mse-mcp pour un modèle de langage, mse-rest pour un service pour scripts ; ils se trouvent dans le bundle de l’application, n’arrivent dans un chemin de recherche que lorsque tu les installes ci-dessous, et les lignes ici indiquent l’emplacement complet.',
  keineShell:
    'Copié, pas exécuté — ce qui figure ici va dans un terminal de ton choix.',
  ort: 'Ils se trouvent dans {ort}.',
  vorhanden: 'Présents : {namen}.',
  keineWerkzeuge:
    'Aucun des trois n’est là. C’est normal tant que l’application s’exécute depuis le code source — la voie de publication les place dans le bundle.',
  kopieren: 'Copier la commande',
  kopiert: 'Copié.',
  kopierfehler:
    'La copie a échoué — la fenêtre n’a pas l’autorisation d’accéder au presse-papiers. Sélectionne le texte et copie-le à la main.',
  befehle: {
    hilfe: 'Ce que peut faire la ligne de commande',
    sets: 'Lister les sets de la bibliothèque',
    verknuepfen:
      'Rendre mse disponible dans le chemin de recherche (demande ton mot de passe)',
    dienst: 'Démarrer le service sur l’interface de bouclage',
    mcp: 'Démarrer le serveur MCP — pour un modèle de langage, via stdio',
    karten: 'Lister les cartes du set ouvert',
    statistik: 'Compter le set ouvert',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: 'Index de recherche local',
  suchindexErklaerung:
    'Textes de cartes et passages de règles, consultables localement — en texte intégral et par vecteurs, fusionnés en une seule liste. Chaque résultat indique d’où il vient et dans quelle révision se trouvait la source.',
  suchindexKeiner:
    'Aucun index de recherche n’a été construit. D’ici là, la recherche ne trouve rien ici.',
  suchindexModell: 'Modèle',
  suchindexLaufzeit: 'Environnement d’exécution',
  suchindexVerfahren: 'Méthode',
  suchindexStaende: 'Révision par source',
  suchindexSemantisch: 'Recherche sémantique',
  suchindexJa: 'oui',
  suchindexNein: 'non',
  suchindexFrage: 'Que cherches-tu ?',
  suchindexSuchen: 'Rechercher',
  suchindexLeer: 'Rien trouvé.',
  suchindexModus: 'Recherche effectuée',
  suchindexModusHybrid: 'en texte intégral et par vecteurs',
  suchindexModusLexikalisch: 'en texte intégral uniquement',
  suchindexModusOhneIndex: 'pas du tout — aucun index n’est disponible',
  suchindexAufbau: 'Reconstruire',
  suchindexAufbauErklaerung:
    'Construit une nouvelle génération à côté de celle en cours et ne bascule qu’à la fin. Cela prend des minutes et nécessite un modèle d’embedding configuré.',
  suchindexModellfeld: 'Modèle d’embedding',
  suchindexQuellfeld: 'Fichier en masse',
  suchindexRevisionfeld: 'Révision de la source',
  suchindexDelta: 'Ne lire que ce qui a changé',
  suchindexAufbauen: 'Construire',
  suchindexAbbrechen: 'Annuler',
  suchindexLaeuft: 'Construction en cours',
  suchindexGeschrieben: 'sections écrites',
  suchindexFertig: 'génération construite',
  // **Configuration for agents** — A14, #1453.
  agenten: 'Prêt pour un agent',
  agentenErklaerung:
    'Claude Code et Codex parlent MCP. Les fragments ci-dessous contiennent le chemin complet de mse-mcp sur ce Mac — à copier et coller. Cette application n’écrit dans aucun de ces fichiers et ne démarre aucun agent.',
  agentenOhneGeheimnis:
    'Aucune clé d’accès n’y figure. mse-mcp les lit lui-même dans le trousseau — ce qui serait écrit ici se trouverait en clair dans un fichier que l’on transmet.',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'Configurer la ligne de commande',
  befehlszeileSatz:
    'Place des liens vers mse et mse-mcp dans un dossier qui t’appartient — sans droits d’administrateur et sans toucher à ton profil de shell. Ce qui s’y trouve déjà et ne vient pas d’ici est conservé.',
  befehlszeileVerzeichnis: 'Dossier : {verzeichnis}',
  befehlszeileNichtImPfad:
    'Ce dossier ne figure pas dans le PATH. Ajoute toi-même cette ligne à ton profil de shell :',
  befehlszeileStaende: {
    frei: 'pas encore configuré',
    eigen: 'configuré',
    veraltet: 'obsolète — pointe vers un programme qui n’existe plus',
    fremd: 'une autre commande s’y trouve, elle est conservée',
  } as Record<string, string>,
  befehlszeileVerdeckt:
    'Masqué par {pfad} : c’est celui-là qui s’exécute lorsque tu tapes le nom.',
  befehlszeilePruefen: 'Vérifier avec',
  befehlszeileEinrichten: 'Configurer',
  befehlszeileHerausnehmen: 'Retirer',
  befehlszeileKopierenPfad: 'Copier la ligne pour le profil de shell',
  befehlszeileKopierenPruefung: 'Copier la commande de vérification',
  agentenDatei: 'À placer dans {datei}',
  agentenInhalt: 'Contenu pour {datei}',
  agentenBefehl: 'Ou via la ligne de commande de l’agent',
  agentenNachgelesen: 'Vérifié : {quelle}',
  agentenKopierenDatei: 'Copier le fragment',
  agentenKopierenBefehl: 'Copier la commande de l’agent',

  peerTitel: 'Comment ce profil s’identifie',

  peerZeigen: 'Afficher l’identité',
  peerErklaerung:
    'Une paire de clés qui permet aux autres de reconnaître qu’ils parlent bien avec toi. La moitié privée reste dans le trousseau de ce Mac et ne va nulle part ; seul l’identifiant sort.',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'Elle appartient à ce profil — ce n’est ni ton compte Apple ni le nom de cet ordinateur.',
  peerStandOhne: 'Ce profil n’a pas encore d’identité.',
  peerStandBereit: 'Prêt.',
  peerStandVerloren: 'La clé est manquante.',
  peerStandFremd: 'Le trousseau et la base de données ne concordent pas.',
  peerKennung: 'Identifiant',
  peerFingerabdruck: 'Empreinte',
  peerFingerabdruckHinweis:
    'C’est ce que tu lis à ton correspondant — un nom d’affichage n’est pas une preuve.',
  peerSeit: 'Depuis {zeitpunkt}',
  peerVorher: 'Auparavant : {kennung}',
  peerAnlegen: 'Créer une identité',
  peerRotieren: 'Remplacer l’identité',
  peerRotierenFrage:
    'Créer une nouvelle identité ? Toute personne ayant confirmé l’empreinte actuelle doit la vérifier de nouveau.',
  peerRotierenJa: 'Remplacer',
  peerRotierenNein: 'Annuler',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: 'Qui ce profil connaît',
  kontakteErklaerung:
    'Des correspondants que ce profil a déjà rencontrés. Une rencontre seule ne permet rien : seule la comparaison de l’empreinte rend quelqu’un vérifiable, et seule une autorisation par usage permet quoi que ce soit.',
  kontakteLeer:
    'Personne rencontré pour l’instant. Quiconque prend contact apparaît ici ensuite — non confirmé.',
  kontakteZeigen: 'Afficher les contacts',
  kontakteAufnehmen: 'Ajouter un contact',
  kontakteAufnehmenErklaerung:
    'Tant qu’il n’y a pas de connexion, les identifiants s’échangent à la main. Toute personne ajoutée ici est non confirmée et ne peut rien faire.',
  kontakteAufnehmenKennung: 'Identifiant du correspondant',
  kontakteAufnehmenName: 'Comment tu l’appelles',
  kontakteStandUnbestaetigt: 'Non confirmé',
  kontakteStandBestaetigt: 'Confirmé',
  kontakteStandGesperrt: 'Bloqué',
  kontakteAbdruck: 'Empreinte',
  kontakteNennt: 'Se fait appeler : {name}',
  kontakteNenntHinweis:
    'Un nom choisi par la personne elle-même n’est pas une preuve d’identité.',
  kontakteGesehen: 'Vu pour la dernière fois : {zeitpunkt}',
  kontakteVergleichen: 'Saisir l’empreinte comparée',
  kontakteBestaetigen: 'Confirmer',
  kontakteSperren: 'Bloquer',
  kontakteEntsperren: 'Débloquer',
  kontakteRechte: 'Autorisé',
  kontakteKeineRechte: 'Rien d’autorisé.',
  zweckPaarung: 'Faire connaissance',
  zweckTeilen: 'Transférer des objets',
  zweckHandel: 'Échanger des cartes',
  zweckPartie: 'Jouer une partie',
  kontakteAdresse: 'Joignable à l’adresse',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    'Soit une adresse sur le même réseau — soit une invitation que ton correspondant t’a envoyée. Vide signifie : aucune.',
  kontakteAdresseMerken: 'Mémoriser',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'Note',
  kontakteNotizBeispiel: 'échange équitablement, répond lentement …',
  kontakteNotizHinweis:
    'Pour toi seul. Elle ne voyage jamais — ni dans une vue ni dans un partage.',
  kontakteNotizMerken: 'Enregistrer la note',
  freigabenTitel: 'Ce que ce pair voit',
  freigabenErklaerung:
    'Sans partage, il ne voit rien. Plusieurs périmètres s’appliquent côte à côte — le plus large l’emporte.',
  freigabenLeer: 'Aucune vue accordée.',
  freigabenOhneZweck:
    'Ce pair ne peut encore rien faire. Accorde-lui d’abord un usage.',
  freigabenZweck: 'Pour quoi',
  freigabenUmfang: 'Quoi',
  freigabenErteilen: 'Accorder la vue',
  freigabenWiderrufen: 'Révoquer',
  freigabenVorschau: 'Voir ce qu’il voit',
  freigabenVorschauLeer: 'Il ne verrait rien.',
  fernTitel: 'Être joignable depuis d’autres réseaux',
  fernErklaerung:
    'Cela permet à quelqu’un de te joindre même s’il n’est pas sur le même réseau — sans redirection de port, sans compte. Tu obtiens une invitation et la lui envoies, par le moyen de ton choix.',
  fernStill: 'Ce profil n’est pas joignable depuis d’autres réseaux.',
  fernHinweis:
    'La connexion cherche le chemin direct et n’utilise un relais que s’il n’y a pas d’autre solution. Seul du trafic chiffré transite par le relais — ni cartes, ni noms.',
  fernStarten: 'Créer une invitation',
  fernBeenden: 'Ne plus être joignable',
  fernWeitergeben:
    'Voici ce dont le correspondant a besoin. Quiconque l’intercepte obtient une ligne, pas une identité — la vérification a lieu à la connexion, au moyen de l’empreinte.',
  fernStirbt:
    'L’invitation est valable tant que ce profil est joignable. Si tu l’arrêtes, elle devient invalide — et tu en obtiendras une nouvelle la prochaine fois.',
  fernKopieren: 'Copier l’invitation',
  fernKopiert: 'Copié.',
  horchenTitel: 'Être joignable pour les autres',
  horchenErklaerung:
    'Tant que ce profil est à l’écoute, un correspondant connu peut établir ' +
    'une connexion vers lui. Cela ne se produit jamais de soi-même — pas même ' +
    'après un redémarrage — et un changement de profil y met fin.',
  horchenAdresse: 'Adresse d’écoute',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'Le port seul (:41234) écoute sur chaque interface de ce Mac. ' +
    'Avec 127.0.0.1 devant, uniquement sur ce Mac lui-même.',
  horchenStarten: 'Devenir joignable',
  horchenBeenden: 'Ne plus être joignable',
  horchenLaeuft: 'À l’écoute sur {adresse}',
  horchenNamen: 'Nom (subsiste après un changement d’adresse)',
  horchenAdressen: 'Adresse',
  horchenNurHier:
    'Avec cette liaison, personne ne peut te joindre de l’extérieur : 127.0.0.1 écoute uniquement sur ce Mac. Indique plutôt seulement le port (:41234) pour que ce Mac soit joignable sur son réseau.',
  horchenNurDiesesNetz:
    'Les deux sont valables sur ce réseau. Une adresse attribuée peut changer, un nom y survit généralement — et une adresse valable partout n’arrive qu’avec la connexion depuis d’autres réseaux.',
  horchenStill: 'Ce profil n’est pas à l’écoute pour le moment.',
  horchenWeitergeben:
    'Voici ce dont le correspondant a besoin — avec l’identifiant de ce profil.',
  sichtbarTitel: 'Être découvert à proximité',
  sichtbarErklaerung:
    'S’annoncer dit à tous ceux du même réseau que ce profil existe ' +
    '— identifiant de pair compris. Être joignable et être découvert ' +
    'sont deux décisions, et voici la seconde.',
  sichtbarName: 'Nom sous lequel s’annoncer',
  sichtbarNameHinweis:
    'Un nom choisi par toi ne prouve rien. Le nom d’hôte de ce Mac ' +
    'n’est pas diffusé.',
  sichtbarWerden: 'S’annoncer',
  sichtbarBeenden: 'Ne plus s’annoncer',
  sichtbarLaeuft: 'Ce profil s’annonce.',
  sichtbarStill: 'Ce profil ne s’annonce pas.',
  sichtbarBrauchtHorchen:
    'Deviens d’abord joignable — une annonce sans port serait ' +
    'une invitation à une porte derrière laquelle il n’y a personne.',
  nachbarnTitel: 'Qui est à proximité',
  nachbarnErklaerung:
    'Des pairs qui s’annoncent en ce moment sur le même réseau. Ce qui ' +
    'est affiché est leur déclaration — elle n’est vérifiée qu’à la connexion.',
  nachbarnSuchen: 'Regarder',
  nachbarnSucht: 'Recherche en cours …',
  nachbarnLeer: 'Personne ne s’annonce en ce moment.',
  nachbarnUnbekannt: 'Inconnu',
  nachbarnAufnehmen: 'Ajouter comme contact',
  nachbarnGemerkt: 'Mémorisé sous',
  nachbarnSchonBekannt: 'Déjà dans la liste',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'Ce pair se trouve à une nouvelle adresse.',
  nachbarnAdresseBisher: 'Mémorisée jusqu’ici',
  nachbarnAdresseKeine: 'aucune',
  nachbarnAdresseUebernehmen: 'Utiliser la nouvelle adresse',
  nachbarnAdresseHinweis:
    'Cela ne change que la destination de la connexion. On ne lui fait pas plus confiance qu’avant, et sa clé est vérifiée à chaque connexion.',
  nachbarnVerworfen:
    '{zahl} réponses ne portaient pas d’identifiant de pair et ne figurent pas dans la liste.',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: 'Depuis l’environnement de ce processus',
  netzwegErklaerung:
    'Ces valeurs proviennent de l’environnement dans lequel le programme a été démarré. Elles remplacent les champs ci-dessous, et aucune fenêtre ne les modifie.',
  netzwegProxyHTTP: 'Proxy pour http',
  netzwegProxyHTTPS: 'Proxy pour https',
  netzwegProxyAus: 'Sans proxy pour',
  netzwegAusUmgebung: '{name} : {wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: 'L’interface pour les scripts',
  dienstRESTErklaerung:
    'Un service web sur la même base de données — pour lire et modifier depuis un script. Il ne tourne que tant que tu le laisses tourner, et s’arrête avec cette fenêtre.',
  dienstMCPTitel: 'Le serveur d’outils via HTTP',
  dienstMCPErklaerung:
    'Les mêmes outils que via l’entrée standard, ici sous une adresse — joignable aussi depuis un autre appareil.',
  dienstRESTAdresse: 'Adresse de l’interface de script',
  dienstRESTPort: 'Port de l’interface de script',
  dienstMCPAdresse: 'Adresse du serveur d’outils',
  dienstMCPPort: 'Port du serveur d’outils',
  dienstMarke: 'Accès',
  dienstMarkeKopieren: 'Copier l’accès',
  dienstFingerabdruck: 'Empreinte',
  dienstFingerabdruckKopieren: 'Copier l’empreinte',
  dienstKopiert: 'Copié',
  dienstStarten: 'Démarrer',
  dienstBeenden: 'Arrêter',
  dienstErreichbar: 'Joignable à l’adresse :',
  dienstMitTLS:
    'Chiffré. Le certificat est auto-émis et porte l’identifiant de ce profil — quiconque se connecte compare son empreinte.',
  dienstOhneTLS:
    'Non chiffré, et c’est acceptable ici : l’adresse ne vaut que sur ce Mac.',
  dienstHinweis:
    '127.0.0.1 signifie : uniquement ce Mac. 0.0.0.0 signifie : tous les appareils du réseau — le chiffrement et l’accès sont alors obligatoires, et les deux se configurent d’eux-mêmes. Le port 0 signifie : en chercher un libre.',
}
