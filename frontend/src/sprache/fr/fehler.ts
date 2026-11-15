export const fehler = {
  unbekannt:
    'Une erreur inattendue s’est produite. Les détails figurent dans le journal — sous « Réglages › Diagnostic », tu peux en voir la fin et l’enregistrer.',
  unbekanntMitVorgang:
    'Une erreur inattendue s’est produite. Les détails figurent dans le journal — sous « Réglages › Diagnostic », tu peux en voir la fin et l’enregistrer. Cherches-y {vorgang} : toutes les lignes de cette opération portent cet identifiant.',
  nichtGefunden:
    'L’entrée recherchée est introuvable. Quelqu’un d’autre l’a peut-être supprimée — recharge la liste.',
  ohnePeeridentitaet:
    'Ce profil n’a pas encore d’identité. Crées-en une dans l’encadré « Comment ce profil s’identifie » — sans elle, personne ne peut prouver qui il est.',
  peerNichtErreichbar:
    'Ton correspondant est injoignable pour le moment. Est-il à l’écoute, et son adresse est-elle juste ? Sinon, réessaie plus tard.',
  peerOhneAdresse:
    'Aucune adresse n’est enregistrée pour ce correspondant. Saisis-la dans son contact sous « Adresse et note » — sans elle, ce programme ne le trouve que sur le même réseau.',
  ohneNetzerlaubnis:
    'Rien n’est parti vers le réseau local. Si macOS vient de demander l’autorisation, accorde-la et réessaie. La cause ne peut pas être établie avec certitude d’ici : soit l’autorisation manque — Réglages Système → Confidentialité et sécurité → Réseau local —, soit le réseau ne laisse pas passer le multicast, par exemple derrière un VPN.',
  ohneNetzroute:
    'Il n’y a pour le moment aucune route vers le réseau local. Une interface réseau est-elle connectée ? Sous Réglages → Connexions, tu peux en choisir une précise.',
  schluesselbundAblage:
    'La clé de ce profil n’a pas pu être enregistrée dans le trousseau. Le trousseau de session est-il déverrouillé ? Réessaie ensuite.',
  einladungZuNeu:
    'Cette invitation provient d’une version plus récente de ce programme. Demande à l’autre partie une invitation issue de sa version plus ancienne — ou procure-toi la plus récente.',
  fernhorchenLaeuft:
    'Ce profil est déjà joignable via des réseaux distants. Arrête-le d’abord si tu veux une nouvelle invitation — l’ancienne deviendra invalide.',
  anmeldungAbgelaufen:
    'La connexion n’est pas revenue dans les dix minutes. Recommence-la quand tu es prêt dans le navigateur.',
  umschlagFassung:
    'Cette enveloppe provient d’une autre version de ce programme. Demande à l’autre partie d’en envoyer une issue de sa version plus ancienne — ou procure-toi la plus récente.',
  umschlagUnterschrift:
    'La signature de cette enveloppe n’est pas valide. Elle a été modifiée en chemin ou provient de quelqu’un d’autre que ce qu’elle prétend — ne l’acceptez pas et vérifie auprès de l’autre partie.',
  umschlagInhalt:
    'Un élément de cette enveloppe ne correspond pas à son manifeste. Demande à l’autre partie de l’envoyer de nouveau — quelque chose a été endommagé ou modifié en chemin.',
  fremderNutzer:
    'Seul cet utilisateur peut le faire lui-même. Passe d’abord à cet utilisateur — l’application demandera alors le second facteur.',
  nichtFreigegeben:
    'Ton partenaire ne te partage rien pour l’échange. Demande-lui de te partager une vue — ou échangez par fichier.',
  zeileNichtVerbuchbar:
    'La ligne {zeile} ne pourrait pas être comptabilisée du côté qui la reçoit — l’état ou la langue manque, ou une indication n’existe pas. Retire-la et ajoute-la de nouveau avec état, langue et finition.',
  zusageUnpassend:
    'Cet exemplaire ne remplit aucune des lignes que tu donnes — l’état, la langue ou la finition diffèrent. Choisis-en un qui correspond ou ajoute-le à ta collection.',
  erlaubnis:
    'Tu n’as pas le rôle nécessaire. La personne qui gère ce profil peut te le donner.',
  unbekannteHerkunftsart:
    'Le programme ne connaît pas cette origine. Choisis « repris d’un profil » ou « importé ».',
  // The command centre (#1423).
  befehlBestaetigen:
    'Cette commande modifie quelque chose qui ne peut pas être annulé facilement — lis l’aperçu, puis confirme.',
  befehlArgumente:
    'Les arguments ne forment pas un objet JSON. Saisis un objet entre accolades — un champ par paramètre.',
  herkunftOhneProfil:
    'Une origine issue d’un profil le nomme, un import n’en nomme aucun — indiques-en un ou omets-le.',
  nurLesend:
    'Dans ce profil, tu peux seulement lire. Il appartient à quelqu’un d’autre et ton rôle y est « lecteur ». Passe à ton propre profil pour travailler — ou demande à la personne qui gère celui-ci le rôle « collaborateur ».',
  bestandUnbrauchbar:
    'Ce fichier est endommagé ou n’est pas une bibliothèque de MagicSetEditor4Mac — il ne peut pas être ouvert. Restaure une sauvegarde ; la dernière se trouve là où tu l’as enregistrée.',
  bestandNeuereFassung:
    'Cette bibliothèque provient d’une version plus récente de MagicSetEditor4Mac. Le fichier est intact — ne le remplace pas. Procure-toi la version plus récente, ou restaure une sauvegarde antérieure.',
  bestandAeltereFassung:
    'Cette bibliothèque provient d’une version antérieure à 0.7.0, et cette version ne peut plus la mettre à niveau. Le fichier est intact — ne le remplace pas. Ouvre-la une fois avec MagicSetEditor4Mac 0.7.0 ; ensuite, cette version l’ouvrira aussi.',
  bestandNurLesbar:
    'La bibliothèque peut être lue mais pas écrite — le fichier ou son dossier n’a pas l’autorisation d’écriture. Le fichier est intact — ne le remplace pas. Accorde-lui l’écriture dans le Finder, sous « Lire les informations », puis réessaie.',
  bestandEA:
    'Le disque a signalé une erreur pendant la lecture de la bibliothèque. On ne sait pas si le fichier est intact — ne le remplace pas tout de suite. Vérifie le disque, par exemple avec l’Utilitaire de disque, puis réessaie.',
  kaumPlatz:
    'La bibliothèque n’a pas pu être ouverte ni écrite, et il ne reste presque plus de place sur le disque. Le fichier est probablement intact — ne le remplace pas. Libère de la place et réessaie ; si cela ne suffit pas, vérifie le disque, par exemple avec l’Utilitaire de disque, et le droit d’écriture.',
  startUngeklaert:
    'La bibliothèque n’a pas pu être ouverte, et la raison n’est pas claire — elle figure ci-dessous. Réessaie. Ne restaure une sauvegarde que si tu as la certitude que ce fichier ne peut plus être ouvert.',
  sicherungVerworfen:
    'La sauvegarde n’a pas réussi sa vérification et a été écartée — rien n’a été sauvegardé. Réessaie, ou sauvegarde à un autre endroit.',
  sicherungUnvermerkt:
    'La sauvegarde est écrite et vérifiée — seule son heure n’a pas pu être enregistrée, et l’âge de la dernière sauvegarde indique toujours la précédente. Garde le fichier, et sauvegarde à nouveau quand la bibliothèque sera libre.',
  schluesselFehlt:
    'Cette bibliothèque est chiffrée et sa clé ne se trouve pas dans le trousseau de ce Mac. Le fichier est intact — ne le remplace pas. La clé a été créée à l’activation du chiffrement et ne peut pas être recalculée ; sans elle, personne ne peut plus accéder aux données.',
  belegt:
    'La bibliothèque est occupée pour le moment — probablement une autre fenêtre ou une commande est en train d’écrire. Réessaie dans un instant.',
  bestandWirdGetauscht:
    'La bibliothèque est en cours de remplacement — par la restauration d’une sauvegarde, le changement du chiffrement ou un nouveau départ. Patiente un instant, puis réessaie.',
  anwendungEndet:
    'MagicSetEditor4Mac est en train de se fermer — cette étape n’a pas été exécutée. Relance l’application pour continuer.',
  vorgangAbgebrochen:
    'L’opération a été interrompue avant la fin. Relance-la si tu en as encore besoin.',
  konflikt: 'L’entrée a été modifiée entre-temps. Recharge.',
  konfliktgrund:
    'Cela n’a pas fonctionné : {grund}. Vérifie l’état actuel avant de réessayer.',
  doppelt:
    'Une entrée avec ces informations existe déjà. Modifie l’une d’elles — ou modifie l’entrée existante.',
  doppelterName: 'Ce nom existe déjà ici — choisis-en un autre.',
  doppeltesExemplar:
    'La collection contient déjà cette impression — dans le même état, la même langue et la même finition. Augmentes-y la quantité ou modifie l’une des informations.',
  gesperrt:
    'Ce programme n’a pas le droit d’accéder à cet emplacement — il est verrouillé. Choisis-en un autre ou autorise l’accès, dans les permissions du fichier ou sous Réglages Système › Confidentialité et sécurité › Fichiers et dossiers, puis réessaie.',
  voll: 'Il n’y a plus d’espace sur le disque. Libère de l’espace et réessaie.',
  netzName:
    'Le nom « {wirt} » n’a pas pu être résolu. Vérifie la connexion à Internet et, s’il y en a un, le proxy sous « Réglages › Connexions ». Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  netzNichtErreichbar:
    'Personne n’accepte de connexion à « {wirt} ». Le service fonctionne-t-il, et son adresse est-elle juste ? Sinon, réessaie plus tard. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  netzKeineAntwort:
    '« {wirt} » n’a pas répondu en {sekunden} secondes. Réessaie plus tard. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  netzAntwortStockt:
    'La réponse de « {wirt} » est restée bloquée {sekunden} secondes. Vérifie la connexion et réessaie. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  netzAbgerissen:
    'La connexion à « {wirt} » a été interrompue avant que la réponse soit complète. Réessaie ; si cela se reproduit, le problème vient du service ou d’un proxy intermédiaire. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  netzProxyAnmeldung:
    'Le proxy « {wirt} » exige une connexion. Indique ses identifiants via HTTPS_PROXY dans l’environnement — sous « Réglages › Connexions », le proxy ne prend volontairement aucun mot de passe. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  gedrosselt:
    'L’autre partie signale trop de requêtes. Attends un instant, puis réessaie. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  gedrosseltBis:
    '« {wirt} » n’accepte aucune requête pour le moment. Réessaie à partir de {uhrzeit} — d’ici là, ce programme ne redemandera pas. Ce que la connexion a signalé figure sous « Réglages › Diagnostic » à {vorgang}.',
  kontingent:
    'Trop d’appels en peu de temps. Attends {sekunden} s, puis réessaie.',
  prozessVeraltet:
    'Ce processus provient d’une version plus ancienne que la bibliothèque et s’arrête — rien n’a été écrit. Un service installé redémarre de lui-même, avec le programme pour lequel il est installé ; redémarre tout autre processus avec la nouvelle version, puis réessaie.',
  bestandVomDienstGehalten:
    'Le service pour scripts garde la bibliothèque ouverte. Révoque-le dans Réglages › Diagnostic › « Configurer la ligne de commande et le serveur d’outils », puis restaure à nouveau.',
  falschesSpiel:
    'Le deck appartient à un autre jeu que ce tournoi. Choisis un deck du même jeu.',
  turnierformatSpiel:
    'Le format appartient à un autre jeu que ce tournoi. Choisis un format du même jeu — ou aucun.',
  paketGeaendert:
    'Ce paquet diffère de son état lors de l’installation. Confirme le remplacement — la modification est conservée comme version précédente.',
  symbolZuVieleTeile:
    'Ce symbole comporte {anzahl} parties, {grenze} sont autorisées. Fusionne des parties ou choisis un dessin plus simple.',
  symbolZuVielePunkte:
    'Ce symbole comporte {anzahl} points, {grenze} sont autorisés. Simplifie les formes ou choisis un dessin plus simple.',
  symbolZuTief:
    'Les groupes de ce symbole sont imbriqués sur {anzahl} niveaux, {grenze} sont autorisés. Choisis un dessin plus plat ou dissocie les groupes internes.',
  symbolZuGross:
    'Ce fichier dépasse {grenze} et n’est donc pas un fichier de symbole. Choisis un fichier .mse-symbol.',
  symbolGespeichertZuGross:
    'Enregistré, ce symbole dépasserait {grenze}, plus que ne peut contenir un fichier de symbole. Simplifie les formes ou utilise moins de points.',
  bildZuGross:
    'Cette image est trop grande pour être décompressée. Choisis-en une plus petite ou réduis-la d’abord.',
  setOhneSymbol:
    'Ce set n’a pas encore de symbole. Dessines-en un et enregistre-le — il pourra alors être ajouté.',
  // **Die eigenen Übersetzungen** — LOC-02 (#1688).
  sprachkennungUngueltig:
    '« {kennung} » n’est pas un identifiant de langue conforme à BCP 47. Choisis-en un comme pt-BR ou zh-Hant.',
  sprachkennungEingebaut:
    '« {kennung} » est une langue intégrée. Choisis un identifiant personnel comme {kennung}-x-meine.',
  platzhalterAbweichend:
    'Cette traduction utilise d’autres espaces réservés que la phrase allemande. Saisis exactement ceux-ci : {erwartet}.',
  auszeichnungUnsicher:
    'Cette traduction contient ce qui n’est pas du texte, ou une mise en forme qui ne se ferme pas. Supprime les balises et ferme ** et `.',
  sprachpaketNeuer:
    'Cette traduction provient d’un programme plus récent ({format}). Procure-toi la nouvelle version.',
  sprachpaketOhneLizenz:
    'Sans licence, aucune traduction ne peut être diffusée. Saisis-en une, par exemple CC0-1.0.',
  bereichAbgeschaltet:
    'Ce domaine est désactivé ({schalter}). Réactive-le dans Réglages › Domaines et fonctions — rien n’a été perdu.',
  dienstAus:
    'Le service {dienst} est désactivé. Active-le dans Réglages › Services si tu veux l’interroger.',
  fremdeEigeneKarten:
    'Les cartes personnelles de l’expéditeur ne sont pas arrivées — leur identifiant ne vaut que dans sa bibliothèque : {karten}. Si tu en as besoin, ouvre son set et ajoute-les ici.',
  nichtGelegtePlaetze:
    'Ces entrées n’ont pas pu être ajoutées ici : {plaetze}. Le reste du deck a été accepté ; vérifie-les et ajoute-les à la main.',
  fremdeKartenUndNichtGelegte:
    'Les cartes personnelles de l’expéditeur ne sont pas arrivées — leur identifiant ne vaut que dans sa bibliothèque : {karten}. Et ces entrées n’ont pas pu être ajoutées ici : {plaetze}. Vérifie les deux et ajoute-les à la main.',
  symbolVerwendet:
    'Ces sets portent le symbole : {sets}. Confirme la suppression — ils conservent leur copie.',
  symbolOhneVorige:
    'Ce symbole n’a pas de version précédente. Enregistre d’abord une nouvelle version — la version actuelle pourra alors être restaurée.',
  turnierAbgeschlossen:
    'Ce tournoi est clos. Choisis « Rouvrir » pour modifier quoi que ce soit.',
  sitzGegangen:
    'Ce siège a quitté la partie et ne peut plus agir. Choisis un siège encore en jeu.',
  fremderWert:
    'Quelqu’un d’autre a modifié ce champ entre-temps. Vérifie la nouvelle valeur et décide si tu gardes ton brouillon.',
  vorschlagNurGegenseite:
    'Seule l’autre partie peut modifier une offre ouverte. Attends sa réponse, ou annule l’échange et propose à nouveau.',
  turnierNichtFertig:
    'Saisis d’abord tous les résultats : un tournoi peut être clos dès qu’il compte au moins une ronde et que chaque appariement a un résultat.',
  turniernameLeer: 'Un tournoi a besoin d’un nom — saisis-en un.',
  turnierspielLeer:
    'Un tournoi appartient à un jeu — sinon il n’y a aucun format à vérifier. Choisis-en un.',
  validierungsgrund:
    'Cela n’a pas fonctionné : {grund}. Corrige la saisie et réessaie.',

  validierung:
    'La saisie n’est pas valide. L’indication sur le champ explique pourquoi — corrige-la là.',

  // **The rule itself is the remedy** (#1350).
  // **The password of a backup** (OPS-06, #1723).
  sicherungskennwortKurz:
    'Un mot de passe de sauvegarde comporte au moins douze caractères. Choisis-en un plus long — une phrase de quatre mots est bonne et facile à retenir.',
  sicherungskennwortFehlt:
    'Cette sauvegarde est chiffrée. Saisis son mot de passe pour la consulter ou la restaurer.',
  sicherungskennwortFalsch:
    'Le mot de passe ne correspond pas à cette sauvegarde. Vérifie-le et saisis-le à nouveau.',
  // **Eine MTG-JSON-Datei, die sich nicht lesen lässt** (G01, #2248).
  mtgjsonKeinJSON:
    '{datei} n’est pas un JSON lisible — à l’octet {stelle}, le fichier s’arrête ou est mal formé. Choisis un autre fichier.',
  mtgjsonKeinSet:
    '{datei} est du JSON, mais ne contient aucun set au format MTG JSON (il diverge à l’octet {stelle}). Choisis un fichier avec un set — ou un dossier entier.',
  sicherungImKlartext:
    'La bibliothèque est chiffrée — sans mot de passe, la sauvegarde serait en clair. Définis un mot de passe sous Réglages › Données et sauvegarde ou choisis-y explicitement « sauvegarder sans mot de passe ».',

  planVerschluesselt:
    'La bibliothèque est chiffrée et le plan sauvegarde sans mot de passe — le plan ne la sauvegarde donc pas. Sauvegarde-la manuellement avec un mot de passe.',

  paketnameUngueltig:
    'Un nom de paquet se compose de lettres, de chiffres, de traits d’union et de tirets bas — il devient le nom du répertoire. Choisis-en un avec ces caractères.',
  ordnerangabe:
    '« {wert} » n’est pas un emplacement complet. Saisis-en un qui commence par / ou ~/, sans .. au milieu.',
  ordnerFehlt:
    'Le dossier « {ordner} » est inaccessible. Branche le lecteur ou choisis un autre dossier.',
  ordnerIstDatei:
    '« {ordner} » est un fichier et non un dossier. Choisis le dossier qui le contient, ou un autre.',
  ordnerGesperrt:
    'Cette application n’a pas le droit d’écrire dans « {ordner} ». Choisis un autre dossier ou autorise l’accès à celui-ci.',
  ordnerVoll:
    'Il ne reste que {frei} sur le lecteur de « {ordner} ». Libère de l’espace ou choisis un autre dossier.',
  karteInDeck:
    'Cette carte est utilisée dans {anzahl} deck(s) — {decks}. La supprimer l’en retire aussi ; confirme si c’est bien voulu.',
  pflichtfeld: 'Ce champ ne doit pas être vide — saisis quelque chose.',
  nichtEditierbar:
    'Ce champ ne peut pas être modifié — le modèle ne fait que l’afficher. Modifie la valeur dont il est dérivé.',
  berechnetesFeld:
    'Ce champ est dérivé d’autres champs et ne peut pas être défini sur de nombreuses cartes à la fois. Ses parties sont listées séparément — choisis-en une.',
  keinSetMitStil:
    'Aucun set ouvert n’utilise ce style. Ouvre un set qui l’utilise — ou change le style du set ouvert.',
  vorlageFehlt:
    'Aucun modèle n’est installé pour cette carte. Vérifie dans la page Paquets que le jeu et la feuille de style du set y figurent — sans eux, une carte n’a ni champs ni zones.',
  sicherungNeuereFassung:
    'Cette sauvegarde provient d’une version plus récente de MagicSetEditor4Mac. Elle contient {unbekannt}, et cette version va jusqu’à {stand}. Procure-toi la version plus récente — ou choisis une sauvegarde plus ancienne.',
  sicherungAeltereFassung:
    'Cette sauvegarde provient d’une version antérieure à 0.7.0 — il lui manque des étapes du schéma jusqu’à {grundlinie}, et cette version ne peut plus la mettre à niveau. Restaure-la avec MagicSetEditor4Mac 0.7.0 ; ensuite, cette version ouvrira la bibliothèque.',
  sicherungUnbrauchbar:
    'Ce fichier n’est pas une sauvegarde de MagicSetEditor4Mac. Choisis-en une écrite par cette application.',
  // The parser's own sentence is German for now — see issue #1178.
  skriptFehler:
    'L’expression ne peut pas être évaluée : {meldung} Corrige-la à l’endroit indiqué.',

  // **The missing second factor** — piece 7 of phase 4.12 (#1351).
  // **The local session** — ID-02 (#1701).
  kennwortFalsch:
    'Ce mot de passe est incorrect. Saisis-le à nouveau — ou définis-en un nouveau avec la feuille de récupération.',
  kennwortLaenge:
    'Un mot de passe comporte au moins {mindestens} et au plus {hoechstens} caractères. Choisis-en un de cette longueur.',
  sperreOhneRueckweg:
    'Sans mot de passe et sans déverrouillage local, le verrou ne pourrait pas être levé. Définis d’abord un mot de passe.',
  lokaleEntsperrungFehlt:
    'Le déverrouillage local n’est pas configuré pour cet utilisateur sur ce Mac. Saisis plutôt le mot de passe.',
  codeFalsch:
    'Ce code est incorrect — ou il a déjà été utilisé. Vérifie le code sur la feuille — chacun ne vaut qu’une fois.',
  zweiterFaktor:
    'Cet utilisateur exige un second facteur. Saisis le code de ton app d’authentification ou confirme avec Touch ID.',
  faktorGesperrt: 'Trop de codes erronés. Patiente un instant puis réessaie.',
  sicherungFremdeFassung:
    'Cette sauvegarde ne correspond pas à cette version de MagicSetEditor4Mac. Elle mentionne {dort}, et cette application est {hier}. Restaure-la avec la version qui l’a écrite.',

  // **Die Sätze des Zerteilers und des Auswerters** —
  // #1178.
  //
  // **`durchgereicht` ist die Ausnahme, und sie hat einen Grund**: die 48
  // Sätze der Standardbibliothek nennen Funktions- und Argumentnamen. Ein
  // Eintrag je Funktion wäre eine zweite Abschrift der Bibliothek in zwei
  // Sprachen; dort steht der deutsche Satz, wie er kommt.
  skript: {
    fremdesZeichen: 'le caractère {0} ne fait pas partie du langage',
    einzelnesGleich: 'un seul `=` ; on affecte avec `:=`, on compare avec `==`',
    offeneZeichenkette: 'une chaîne sans guillemet fermant',
    zuTiefGeschachtelt: 'l’expression est imbriquée sur plus de {0} niveaux',
    keinSkript: 'il n’existe aucun script nommé {0}',
    skriptkreis: 'ces champs se lisent mutuellement en boucle : {0}',
    keinVerweis: '`+` ne peut pas concaténer une référence à un autre script',
    keineZahl: '{0} n’est pas un nombre',
    formUnbekannt: 'la forme {0} ne peut pas être évaluée',
    nameUnbekannt: 'le nom {0} n’est pas connu ici',
    keineFunktion:
      '{0} est appelé ici, ce qui ne fonctionne que pour une fonction',
    ohneRumpf: 'cette fonction n’a pas de corps',
    nichtWahrheitswert:
      '`not` nécessite une valeur booléenne, ici se trouve {0}',
    vorzeichenZahl: 'un signe nécessite un nombre, ici se trouve {0}',
    zaehlschleifeZahl:
      'dans une boucle de comptage, {0} nécessite un nombre, ici se trouve {1}',
    rechnenLinks: '`{0}` calcule avec des nombres, à gauche {1}',
    rechnenRechts: '`{0}` calcule avec des nombres, à droite {1}',
    vergleichLinks: '`{0}` compare des nombres, à gauche {1}',
    vergleichRechts: '`{0}` compare des nombres, à droite {1}',
    keineListe: 'sort_list trie une liste, ici se trouve {0}',
    zuTief:
      'l’évaluation descend à plus de {0} niveaux — quelque chose s’appelle-t-il lui-même ici ?',
    schleifeSteht:
      'une boucle de comptage n’avance pas à {0} — le nombre est trop grand',
    funktionVerbinden:
      'une fonction ne peut être combinée qu’avec une fonction, ici se trouvent {0} et {1}',
    fehltNach: 'après {0}, il manque {1}',
    fehltHier: 'il manque {0} ici',
    klammerOffen: 'ce {0} n’est pas fermé — il manque {1}',
    ausdruckGehtWeiter: 'l’expression s’arrête ici, mais le texte continue',
    zuweisungBrauchtNamen: 'l’affectation se fait vers un nom',
    durchgereicht: '{0}',
    bedingungWahrheit:
      'une condition nécessite une valeur booléenne, ici se trouve {0}',
    fuerJedenListe: '`for each` parcourt une liste, ici se trouve {0}',
    wahrheitLinks:
      '`{0}` nécessite des valeurs booléennes, à gauche se trouve {1}',
    wahrheitRechts:
      '`{0}` nécessite des valeurs booléennes, à droite se trouve {1}',
    bindenFunktion: '`@` lie des arguments à une fonction, ici se trouve {0}',
  },
} as const
