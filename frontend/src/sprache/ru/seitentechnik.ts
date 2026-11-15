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
  gesundheit: 'Рабочее состояние',
  gesundheitErklaerung:
    'Жив ли этот процесс и может ли он работать. Внешний сервис, который не отвечает, не делает его неготовым — редактирование продолжается без сети.',
  lebt: 'Процесс отвечает.',
  bereit: 'Он может работать.',
  nichtBereit: 'Сейчас он не может работать.',
  grundBestandZu: 'Ни одна библиотека не открыта.',
  grundWanderungLaeuft: 'Библиотека мигрирует на новую схему.',
  grundBestandUnlesbar: 'Библиотека не отвечает.',
  fremdeDienste: 'Внешние сервисы',
  erreichbar: 'доступен',
  nichtErreichbar: 'недоступен',
  metriken: 'Воздействия с момента запуска',
  metrikenErklaerung:
    'Что этот процесс сделал вовне — по интерфейсам, классам воздействия и исходам. Считаются только проверенные воздействия; простое чтение здесь не отображается.',
  metrikenGrenze:
    'Не более {grenze} строк — их число не растёт с числом карт, наборов или профилей.',
  metrikenLeer: 'Пока ничто не затронуло внешний мир.',
  spalteFlaeche: 'Интерфейс',
  spalteWirkung: 'Воздействие',
  spalteAusgang: 'Исход',
  spalteAnzahl: 'Число',
  flaechen: {
    fenster: 'Окно',
    kommandozeile: 'Командная строка',
    mcp: 'MCP',
    dienst: 'Запустить службу для скриптов',
  },
  wirkungen: {
    lokalLesen: 'локальное чтение',
    lokalSchreiben: 'локальная запись',
    externSenden: 'отправка вовне',
    externSchreiben: 'изменение снаружи',
    verbindlich: 'обязывающее',
  },
  ausgaenge: { gewaehrt: 'разрешено', abgewiesen: 'отклонено' },
  paket: {
    ueberschrift: 'Диагностический пакет',
    erklaerung:
      'Та же информация по частям, в виде файла с контрольными суммами — для отчёта, которому недостаточно текста. Он создаётся там, куда ты укажешь, и сам никуда не отправляется.',
    inhalt:
      'В него входят версия, платформа, библиотека, пакеты, счётчики процесса, строки журнала и следы сбоев — без карт, изображений, коллекции и учётных данных.',
    zusammenstellen: 'Собрать',
    sichern: 'Сохранить пакет',
    loeschen: 'Удалить пакет',
    groesse: '{bytes} из максимум {grenze}.',
    zeitraum: 'Строки журнала охватывают период с {von} по {bis}.',
    ausgelassen: 'Пропущено — {teil}:',
    teile: {
      fassung: 'Версия и платформа',
      bestand: 'Расположение и схема библиотеки',
      pakete: 'Установленные пакеты шаблонов',
      metriken: 'Счётчики этого процесса',
      protokoll: 'Строки журнала',
    },
    spalteTeil: 'Часть',
    spalteDatei: 'Файл',
    spalteZeilen: 'Строки',
    spalteBytes: 'Размер',
    spalteMitnehmen: 'Включить',
    mitnehmen: 'Включить в пакет: {teil}',
  },
  diagnose: {
    ueberschrift: 'Диагностика для отчёта об ошибке',
    erklaerung:
      'Если что-то не работает, этот текст помогает найти причину. Он показывается здесь — передавать его или нет, решаешь ты.',
    inhalt:
      'В нём версия, компьютер, база данных, журнал и пакеты шаблонов — без текстов карт, без значений полей, без содержимого наборов. Подробно: версия, тип компьютера, сборка, из которой запущена программа, договор привязок между окном и программой, расположение базы данных, состояние схемы, число осиротевших строк, расположение журнала, твои пакеты шаблонов и последние 200 строк журнала. Твоя домашняя папка отображается как «~».',
    anzeigen: 'Показать диагностику',
    kopieren: 'Копировать',
    kopiert: 'Скопировано',
    kopierfehler:
      'Не удалось скопировать — у окна нет разрешения на доступ к буферу обмена. Выдели текст и скопируй его вручную.',
    sichern: 'Сохранить в файл',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: 'Поиск по строкам',
    zeilenGezeigt: '{gezeigt} из {alle} строк',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: 'Технические сведения и диагностика',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'Сеть и доступность',

  ueberschrift: 'Инструменты без окна',
  erklaerung:
    'Это приложение включает три программы, работающие без окна. mse для командной строки, mse-mcp для языковой модели, mse-rest для службы для скриптов; они лежат внутри пакета приложения, в путь поиска попадают, только когда ты настроишь их ниже, а строки здесь содержат полное расположение.',
  keineShell:
    'Копируется, а не выполняется: то, что здесь, предназначено для терминала по твоему выбору.',
  ort: 'Они находятся в {ort}.',
  vorhanden: 'Есть: {namen}.',
  keineWerkzeuge:
    'Ни одной из трёх нет. Это нормально, пока приложение запущено из исходников, — при выпуске их кладут в пакет приложения.',
  kopieren: 'Копировать команду',
  kopiert: 'Скопировано.',
  kopierfehler:
    'Не удалось скопировать — у окна нет разрешения на доступ к буферу обмена. Выдели текст и скопируй его вручную.',
  befehle: {
    hilfe: 'Что умеет командная строка',
    sets: 'Показать наборы в библиотеке',
    verknuepfen: 'Сделать mse доступной в пути поиска (запросит пароль)',
    dienst: 'Запустить сервис на интерфейсе loopback',
    mcp: 'Запустить сервер MCP — для языковой модели, через stdio',
    karten: 'Показать карты открытого набора',
    statistik: 'Подсчитать открытый набор',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: 'Локальный поисковый индекс',
  suchindexErklaerung:
    'Тексты карт и фрагменты правил, доступные для локального поиска — полнотекстово и через векторы, сведённые в один список. Каждый результат указывает, откуда он взят и в какой ревизии был источник.',
  suchindexKeiner:
    'Поисковый индекс не построен. Пока его нет, поиск здесь ничего не находит.',
  suchindexModell: 'Модель',
  suchindexLaufzeit: 'Среда выполнения',
  suchindexVerfahren: 'Метод',
  suchindexStaende: 'Ревизия по источникам',
  suchindexSemantisch: 'Семантический поиск',
  suchindexJa: 'да',
  suchindexNein: 'нет',
  suchindexFrage: 'Что ты ищешь?',
  suchindexSuchen: 'Искать',
  suchindexLeer: 'Ничего не найдено.',
  suchindexModus: 'Искали',
  suchindexModusHybrid: 'полнотекстово и через векторы',
  suchindexModusLexikalisch: 'только полнотекстово',
  suchindexModusOhneIndex: 'вообще не искали — индекса нет',
  suchindexAufbau: 'Перестроить',
  suchindexAufbauErklaerung:
    'Строит новое поколение рядом с работающим и переключается только в конце. Это занимает минуты и требует настроенной модели эмбеддингов.',
  suchindexModellfeld: 'Модель эмбеддингов',
  suchindexQuellfeld: 'Файл массива данных',
  suchindexRevisionfeld: 'Ревизия источника',
  suchindexDelta: 'Читать только изменения',
  suchindexAufbauen: 'Построить',
  suchindexAbbrechen: 'Отмена',
  suchindexLaeuft: 'Идёт построение',
  suchindexGeschrieben: 'разд. записано',
  suchindexFertig: 'поколение построено',
  // **Configuration for agents** — A14, #1453.
  agenten: 'Готово для агента',
  agentenErklaerung:
    'Claude Code и Codex говорят на MCP. Фрагменты ниже содержат полный путь к mse-mcp на этом компьютере — скопируй и вставь. Это приложение не записывает ни в один из этих файлов и не запускает никаких агентов.',
  agentenOhneGeheimnis:
    'Ключа доступа в них нет. mse-mcp читает его сам из связки ключей — всё, что было бы записано здесь, лежало бы открытым текстом в файле, которым делятся.',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'Настроить командную строку',
  befehlszeileSatz:
    'Размещает ссылки на mse и mse-mcp в папке, которая принадлежит тебе, — без прав администратора и без изменения профиля твоей оболочки. Всё, что там уже есть и не отсюда, остаётся.',
  befehlszeileVerzeichnis: 'Папка: {verzeichnis}',
  befehlszeileNichtImPfad:
    'Эта папка не входит в твой PATH. Добавь эту строку в профиль оболочки самостоятельно:',
  befehlszeileStaende: {
    frei: 'ещё не настроено',
    eigen: 'настроено',
    veraltet: 'устарело — указывает на программу, которой больше нет',
    fremd: 'там другая команда, она остаётся',
  } as Record<string, string>,
  befehlszeileVerdeckt: 'Перекрыто {pfad}: при вводе имени запускается она.',
  befehlszeilePruefen: 'Проверить с помощью',
  befehlszeileEinrichten: 'Настроить',
  befehlszeileHerausnehmen: 'Убрать снова',
  befehlszeileKopierenPfad: 'Копировать строку для профиля оболочки',
  befehlszeileKopierenPruefung: 'Копировать команду проверки',
  agentenDatei: 'Должно находиться в {datei}',
  agentenInhalt: 'Содержимое для {datei}',
  agentenBefehl: 'Или через командную строку агента',
  agentenNachgelesen: 'Сверено: {quelle}',
  agentenKopierenDatei: 'Копировать фрагмент',
  agentenKopierenBefehl: 'Копировать команду агента',

  peerTitel: 'Как этот профиль представляется',

  peerZeigen: 'Показать идентичность',
  peerErklaerung:
    'Пара ключей, по которой другие могут убедиться, что говорят именно с тобой. Закрытая половина остаётся в связке ключей этого компьютера и никуда не уходит; наружу выходит только идентификатор.',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'Она принадлежит этому профилю — это не твой Apple Account и не имя этого компьютера.',
  peerStandOhne: 'У этого профиля пока нет идентичности.',
  peerStandBereit: 'Готово.',
  peerStandVerloren: 'Ключ отсутствует.',
  peerStandFremd: 'Связка ключей и база данных расходятся.',
  peerKennung: 'Идентификатор',
  peerFingerabdruck: 'Отпечаток',
  peerFingerabdruckHinweis:
    'Именно его ты зачитываешь другой стороне — отображаемое имя ничего не доказывает.',
  peerSeit: 'С {zeitpunkt}',
  peerVorher: 'Ранее: {kennung}',
  peerAnlegen: 'Создать идентичность',
  peerRotieren: 'Заменить идентичность',
  peerRotierenFrage:
    'Создать новую идентичность? Все, кто подтвердил текущий отпечаток, должны будут проверить его заново.',
  peerRotierenJa: 'Заменить',
  peerRotierenNein: 'Отмена',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: 'Кого знает этот профиль',
  kontakteErklaerung:
    'Удалённые стороны, с которыми этот профиль уже встречался. Одно знакомство ничего не разрешает: только сверка отпечатка делает кого-то проверяемым, и только разрешение для конкретной цели что-то позволяет.',
  kontakteLeer:
    'Пока ни с кем не встречались. Тот, кто выйдет на связь, появится здесь — неподтверждённым.',
  kontakteZeigen: 'Показать контакты',
  kontakteAufnehmen: 'Добавить контакт',
  kontakteAufnehmenErklaerung:
    'Пока нет соединения, идентификаторы обмениваются вручную. Тот, кого добавили здесь, не подтверждён и ничего не может делать.',
  kontakteAufnehmenKennung: 'Идентификатор собеседника',
  kontakteAufnehmenName: 'Как ты его называешь',
  kontakteStandUnbestaetigt: 'Не подтверждён',
  kontakteStandBestaetigt: 'Подтверждён',
  kontakteStandGesperrt: 'Заблокирован',
  kontakteAbdruck: 'Отпечаток',
  kontakteNennt: 'Называет себя: {name}',
  kontakteNenntHinweis: 'Выбранное самим человеком имя не доказывает личность.',
  kontakteGesehen: 'Последний раз замечен: {zeitpunkt}',
  kontakteVergleichen: 'Введи сверенный отпечаток',
  kontakteBestaetigen: 'Подтвердить',
  kontakteSperren: 'Заблокировать',
  kontakteEntsperren: 'Разблокировать',
  kontakteRechte: 'Разрешено',
  kontakteKeineRechte: 'Ничего не разрешено.',
  zweckPaarung: 'Знакомство',
  zweckTeilen: 'Передача объектов',
  zweckHandel: 'Обмен картами',
  zweckPartie: 'Игра в партию',
  kontakteAdresse: 'Доступен по адресу',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    'Либо адрес в той же сети, либо приглашение, которое прислала другая сторона. Пусто означает: нет.',
  kontakteAdresseMerken: 'Запомнить',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'Заметка',
  kontakteNotizBeispiel: 'честно обменивается, отвечает медленно …',
  kontakteNotizHinweis:
    'Только для тебя. Она никогда не передаётся — ни в представление, ни в общий доступ.',
  kontakteNotizMerken: 'Сохранить заметку',
  freigabenTitel: 'Что видит этот пир',
  freigabenErklaerung:
    'Без разрешения он не видит ничего. Несколько областей действуют одновременно — побеждает более широкая.',
  freigabenLeer: 'Ни одно представление не открыто.',
  freigabenOhneZweck:
    'Этому пиру пока ничего нельзя. Сначала разреши ему цель.',
  freigabenZweck: 'Для чего',
  freigabenUmfang: 'Что',
  freigabenErteilen: 'Открыть доступ',
  freigabenWiderrufen: 'Отозвать',
  freigabenVorschau: 'Посмотреть, что он видит',
  freigabenVorschauLeer: 'Он не увидел бы ничего.',
  fernTitel: 'Быть доступным через сети',
  fernErklaerung:
    'Так с тобой можно связаться, даже когда вы не в одной сети, — без проброса портов и без учётной записи. Ты получаешь приглашение и отправляешь его любым удобным способом.',
  fernStill: 'Этот профиль недоступен через сети.',
  fernHinweis:
    'Соединение ищет прямой путь и использует ретранслятор, только если другого нет. Через ретранслятор проходит лишь зашифрованный трафик — ни карт, ни имён.',
  fernStarten: 'Создать приглашение',
  fernBeenden: 'Перестать быть доступным',
  fernWeitergeben:
    'Это то, что нужно другой стороне. Кто перехватит это, получит строку, а не идентичность, — проверка происходит при соединении, по отпечатку.',
  fernStirbt:
    'Приглашение действует, пока этот профиль доступен. Останови его — и приглашение станет недействительным; в следующий раз ты получишь новое.',
  fernKopieren: 'Копировать приглашение',
  fernKopiert: 'Скопировано.',
  horchenTitel: 'Быть доступным для других',
  horchenErklaerung:
    'Пока этот профиль слушает, знакомый пир может открыть к нему ' +
    'соединение. Само по себе этого не происходит — даже после ' +
    'перезапуска, — а смена профиля прекращает это.',
  horchenAdresse: 'Адрес для прослушивания',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'Один порт (:41234) слушает на всех интерфейсах этого компьютера. ' +
    'Если перед ним стоит 127.0.0.1 — только на самом этом компьютере.',
  horchenStarten: 'Стать доступным',
  horchenBeenden: 'Перестать быть доступным',
  horchenLaeuft: 'Слушает на {adresse}',
  horchenNamen: 'Имя (сохраняется при смене адреса)',
  horchenAdressen: 'Адрес',
  horchenNurHier:
    'При такой привязке извне до тебя никто не достучится: 127.0.0.1 слушает только на этом компьютере. Укажи вместо этого один порт (:41234), чтобы этот Mac был доступен в своей сети.',
  horchenNurDiesesNetz:
    'Оба действуют в этой сети. Выданный адрес может измениться, имя обычно сохраняется, — а адрес, действующий везде, появляется только с соединениями между сетями.',
  horchenStill: 'Сейчас этот профиль не слушает.',
  horchenWeitergeben:
    'Это то, что нужно другой стороне, — вместе с ключом этого профиля.',
  sichtbarTitel: 'Быть обнаруженным поблизости',
  sichtbarErklaerung:
    'Объявление сообщает всем в той же сети, что этот профиль ' +
    'существует, — включая его идентификатор пира. Быть доступным и быть ' +
    'обнаруженным — два разных решения, и это второе.',
  sichtbarName: 'Имя для объявления',
  sichtbarNameHinweis:
    'Выбранное самим человеком имя ничего не доказывает. Имя этого Mac в сети наружу не уходит.',
  sichtbarWerden: 'Объявить',
  sichtbarBeenden: 'Прекратить объявление',
  sichtbarLaeuft: 'Этот профиль объявляет о себе.',
  sichtbarStill: 'Этот профиль не объявляет о себе.',
  sichtbarBrauchtHorchen:
    'Сначала нужно стать доступным — объявление без порта было бы ' +
    'приглашением к двери, за которой никого нет.',
  nachbarnTitel: 'Кто рядом',
  nachbarnErklaerung:
    'Пиры, объявляющие о себе в той же сети прямо сейчас. То, что ты ' +
    'видишь, — их собственное утверждение; оно проверяется только при соединении.',
  nachbarnSuchen: 'Посмотреть',
  nachbarnSucht: 'Поиск …',
  nachbarnLeer: 'Сейчас никто не объявляет о себе.',
  nachbarnUnbekannt: 'Неизвестно',
  nachbarnAufnehmen: 'Добавить в контакты',
  nachbarnGemerkt: 'Запомнен как',
  nachbarnSchonBekannt: 'Уже в списке',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'Этот пир находится по новому адресу.',
  nachbarnAdresseBisher: 'Запомненный ранее',
  nachbarnAdresseKeine: 'нет',
  nachbarnAdresseUebernehmen: 'Использовать новый адрес',
  nachbarnAdresseHinweis:
    'Это меняет только то, куда устанавливается соединение. Доверия к нему не становится больше, чем было, а его ключ проверяется при каждом соединении.',
  nachbarnVerworfen:
    'Ответов без идентификатора пира: {zahl} — их нет в списке.',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: 'Из окружения этого процесса',
  netzwegErklaerung:
    'Эти значения берутся из окружения, в котором запущена программа. Они переопределяют поля ниже, и ни одно окно их не меняет.',
  netzwegProxyHTTP: 'Прокси для http',
  netzwegProxyHTTPS: 'Прокси для https',
  netzwegProxyAus: 'Без прокси для',
  netzwegAusUmgebung: '{name}: {wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: 'Интерфейс для скриптов',
  dienstRESTErklaerung:
    'Веб-сервис над той же базой данных — для чтения и изменения из скрипта. Он работает, только пока ты это позволяешь, и завершается вместе с этим окном.',
  dienstMCPTitel: 'Сервер инструментов по HTTP',
  dienstMCPErklaerung:
    'Те же инструменты, что и через стандартный ввод, но по адресу — доступные в том числе с другого устройства.',
  dienstRESTAdresse: 'Адрес интерфейса для скриптов',
  dienstRESTPort: 'Порт интерфейса для скриптов',
  dienstMCPAdresse: 'Адрес сервера инструментов',
  dienstMCPPort: 'Порт сервера инструментов',
  dienstMarke: 'Доступ',
  dienstMarkeKopieren: 'Копировать доступ',
  dienstFingerabdruck: 'Отпечаток',
  dienstFingerabdruckKopieren: 'Копировать отпечаток',
  dienstKopiert: 'Скопировано',
  dienstStarten: 'Запустить',
  dienstBeenden: 'Остановить',
  dienstErreichbar: 'Доступен по адресу:',
  dienstMitTLS:
    'Зашифровано. Сертификат выдан самим себе и содержит идентификатор этого профиля — тот, кто подключается, сверяет его отпечаток.',
  dienstOhneTLS:
    'Без шифрования, и здесь это допустимо: адрес действует только на этом компьютере.',
  dienstHinweis:
    '127.0.0.1 означает: только этот компьютер. 0.0.0.0 означает: все устройства в сети — тогда шифрование и доступ обязательны, и оба настраиваются сами. Порт 0 означает: найти свободный.',
}
