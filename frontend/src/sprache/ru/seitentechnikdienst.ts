// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: 'Собственный сертификат от {aussteller}, действителен до {bis}',
  identitaet:
    'Сертификат на основе идентичности этого профиля — тот, кто подключается, сверяет отпечаток.',
}

export const fernwege = {
  titel: 'Как подключены удалённые стороны',
  keine: 'Этим приглашением пока никто не воспользовался.',
  direkt: 'Напрямую, через {adresse}',
  relais: 'Через ретранслятор {adresse} — прямого пути пока нет',
  aufbau: 'Известна, но пока недоступна ни по одному пути',
  unbestimmt: 'Подключено: {anzahl} — путь пока неизвестен.',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'Настройка командной строки и сервера инструментов',
  satz: 'Одна кнопка вносит то, что нужно программе на этом Mac, чтобы найти командную строку или сервер инструментов, — вторая убирает ровно это. Заранее показано, какой файл и какие строки затрагиваются; записи других остаются как есть.',
  namen: {
    befehle: 'Ссылки на mse и mse-mcp в ~/.local/bin',
    befehlszeile: 'Командная строка в PATH оболочки входа',
    codex: 'Сервер инструментов для Codex',
    claudedesktop: 'Claude Desktop',
    claudecode: 'Claude Code',
    dienst: 'Сервис для скриптов как LaunchAgent',
  },
  staende: {
    offen: 'не настроено',
    eingerichtet: 'настроено',
    veraltet: 'устарело',
    fremd: 'чужая запись',
    nichtGefunden: 'нет на этом Mac',
    unnoetig: 'не нужно',
  },
  staendeSaetze: {
    veraltet:
      'Приложение теперь лежит в другом месте; «Настроить» это исправит.',
    fremd: 'Чужая запись уже есть — она остаётся как есть.',
    unnoetig: 'Уже есть в PATH.',
  },
  zeilen: 'Будут добавлены эти строки',
  vorschau: {
    json: 'Эта запись добавляется в файл',
    befehl: 'Выполняется эта команда',
    agent: 'Создаётся этот файл, и эта команда его загружает',
  },
  nurEintrag:
    '{name} с тех пор изменился — удалена только собственная запись, остальное осталось как было.',
  einrichten: 'Настроить',
  widerrufen: 'Отозвать',
  alle: {
    titel: 'Всё сразу',
    satz: 'Настраивает каждую цель, которая есть на этом Mac и ещё не настроена, или отменяет всё настроенное. Чужие записи остаются как есть.',
    vorschau: 'Что будет записано ({anzahl} целей)',
    vorschauLeer: 'Ничего не недостаёт: настроено то, что здесь есть.',
    einrichten: 'Настроить всё',
    widerrufen: 'Отменить всё',
    ergebnis: 'Что произошло',
    ausgaenge: {
      eingerichtet: 'настроено',
      widerrufen: 'отменено',
      fremd: 'чужая запись, без изменений',
      unveraendert: 'без изменений',
      gescheitert: 'не удалось',
      abgebrochen: 'прервано',
    },
  },
  lauschwahl: {
    titel: 'Где слушают службы',
    satz: 'Относится к службе для скриптов — запущенной в окне и как LaunchAgent — и к серверу инструментов по HTTP, начиная со следующего запуска.',
    wahl: 'Доступно через',
    schleife: 'Только этот Mac (127.0.0.1)',
    auswahl: 'Выбранные адреса',
    alle: 'Все интерфейсы (0.0.0.0)',
    alleSatz: 'Каждый интерфейс, в том числе добавленный позже.',
    aussen:
      'Выбранный адрес выходит за пределы этого Mac: службы тогда требуют TLS и доступ (токен), а macOS может запросить разрешение для локальной сети.',
    inaktiv: 'выключен',
    ohneAdresse: 'нет адреса',
    liste: 'Интерфейсы этого Mac',
    arten: {
      schleife: 'Петля',
      netz: 'Сеть',
      tunnel: 'Туннель (VPN)',
      bruecke: 'Мост',
      direkt: 'Прямая беспроводная связь',
    },
    speichern: 'Применить выбор',
    gespeichert: 'Применено — действует при следующем запуске службы.',
    keineGewaehlt: 'Выбери хотя бы один адрес.',
    platzhalter: 'как в разделе «Где слушают службы»',
    weitere: 'Также слушает на',
    ohne: 'Без {adresse}: {grund}',
  },
}
