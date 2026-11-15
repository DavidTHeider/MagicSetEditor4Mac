package Gast;
# **Was jeder Treiber im Gast braucht** — aus `budgets.pl` herausgelöst, als
# die Untergrenzenprobe (I02,
# #2260)
# dieselben Griffe brauchte: einen Ort mit eigenem Zuhause, Bestand und Bund,
# die Kommandozeile darin, den Start über LaunchServices, die Abnahmeleitung
# und den MCP von Wails.
#
# **Perl und nicht Node**, weil der Gast keins hat und keins bekommen soll:
# Perl 5.34 liegt jedem macOS bei, samt `JSON::PP`, `HTTP::Tiny` und
# `IO::Socket::UNIX`. **Kein Kennwort des Gastes** — der Bund trägt das
# Probekennwort der Hüllenprobe.
use strict;
use warnings;
use utf8;
use Exporter 'import';
use File::Temp qw(tempdir);
use HTTP::Tiny;
use IO::Socket::UNIX;
use JSON::PP;
use POSIX qw(WNOHANG);
use Time::HiRes qw(time sleep);

our @EXPORT_OK = qw(lesen jetzt buendel messungen ort mse frage taste seite warte starte beende klicke textziel aufraeumen laeuft);

my $J = JSON::PP->new->utf8->canonical;
my $W = JSON::PP->new->allow_nonref;
my $http = HTTP::Tiny->new(timeout => 60);
my ($APP, $MSE, $m);
my @orte;
my ($laeuft, $sock, $port, $schluessel, $aufruf) = (0, '', 0, '', 0);

# eigene sind die Prozesse, die den Abnahmesocket eines **eigenen** Orts
# offen halten — die gestartete Anwendung selbst; `$laeuft` ist nur das
# wartende `open -W` (Befund 2 des Zweitblicks über Phase I). **Gefragt wird
# der Socket und keine gemerkte PID** (Befunde 1 und 2 des Nachlaufs): eine
# PID kam erst nach dem Start in die Liste, ein gescheiterter Start ließ die
# Anwendung laufen, und eine freigewordene PID kann ein fremdes Programm
# tragen. Ein Ort und sein Socket gehören diesem Lauf.
sub eigene {
  my %ziel;
  for my $o (@orte) {
    my $pfad = "$o->{ort}/abnahme.sock";
    $ziel{$pfad} = $ziel{ $pfad =~ s{^/private}{}r } = $ziel{"/private$pfad"} = 1;
  }
  my ($pid, %pids);
  for (`lsof -U -F pn 2>/dev/null`) {
    chomp;
    if (/^p(\d+)$/) { $pid = $1 }
    elsif (/^n(.+)$/ && $ziel{$1} && $pid != $$) { $pids{$pid} = 1 }
  }
  return keys %pids;
}

sub lesen { local $/; open my $f, '<', $_[0] or die "$_[0]: $!\n"; my $x = <$f>; close $f; return $x }
sub jetzt { return time() * 1000 }

# buendel nennt das Programm, das gestartet wird — **auf der Platte des
# Gastes**, nicht im geteilten Verzeichnis: ein Start darüber mäße dessen Weg
# mit.
sub buendel { ($APP) = @_; $MSE = "$APP/Contents/MacOS/mse"; return }

# messungen sind die JavaScript-Stücke aus dem Drehbuch (`messungen.json`).
sub messungen { ($m) = @_; return }

sub laeuft { return $laeuft }

sub ort {
  my $o = tempdir('/tmp/huelle-XXXXXX');
  mkdir "$o/zuhause";
  my $bund = "$o/bund.keychain-db";
  system('security', 'create-keychain', '-p', 'leer', $bund) == 0 or die "Bund\n";
  system('security', 'unlock-keychain', '-p', 'leer', $bund);
  system('security', 'set-keychain-settings', $bund);
  my $r = { ort => $o, heim => "$o/zuhause", bestand => "$o/bestand.db", bund => $bund };
  push @orte, $r;
  return $r;
}

sub mse {
  my ($o, @argumente) = @_;
  local $ENV{HOME} = $o->{heim};
  local $ENV{MSE_BESTAND} = $o->{bestand};
  local $ENV{MSE_SCHLUESSELBUND} = $o->{bund};
  open my $aus, '-|', $MSE, @argumente or die "mse: $!\n";
  local $/;
  my $text = <$aus>;
  close $aus or die "mse @argumente: Rückgabe $?\n";
  return $text // '';
}

sub frage {
  my ($befehl) = @_;
  my $s = IO::Socket::UNIX->new(Type => SOCK_STREAM(), Peer => $sock) or die "Leitung: $!\n";
  print {$s} $J->encode($befehl), "\n";
  my $zeile = <$s>;
  close $s;
  die "$befehl->{was}: keine Antwort\n" unless defined $zeile;
  my $a = $J->decode($zeile);
  die "$befehl->{was}: " . ($a->{fehler} // '?') . "\n" unless $a->{ok} || $a->{zu};
  return $a->{daten};
}

sub taste { my ($t, @mit) = @_; return frage({ was => 'taste', taste => $t, mit => [@mit] }) }

sub seite {
  my ($js) = @_;
  my $rumpf = $J->encode({ jsonrpc => '2.0', id => ++$aufruf, method => 'tools/call',
    params => { name => 'js_eval', arguments => { js => "return JSON.stringify(await (async () => { $js })())" } } });
  my $r = $http->post("http://127.0.0.1:$port/mcp",
    { headers => { 'content-type' => 'application/json', authorization => "Bearer $schluessel" }, content => $rumpf });
  die "MCP: $r->{status} $r->{reason}\n" unless $r->{success};
  my $a = $J->decode($r->{content});
  die "MCP: $a->{error}{message}\n" if $a->{error};
  my $inhalt = $a->{result}{content}[0]{text};
  die "js_eval: $inhalt\n" if $a->{result}{isError};
  return undef if $inhalt eq 'undefined' || $inhalt eq 'ok';
  return $W->decode($inhalt);
}

# warte fragt, bis die Bedingung wahr ist — auf die Wirkung, mit Frist.
sub warte {
  my ($was, $bedingung, $frist, $takt) = @_;
  my $ende = time() + $frist;
  for (;;) {
    my $w = eval { $bedingung->() };
    return $w if $w;
    my $grund = $@ ? $@ =~ s/\s+$//r : 'noch nicht';
    die "$was (nach $frist s): $grund\n" if time() > $ende;
    sleep($takt // 0.15);
  }
}

# starte gibt die Zeit vom `open` bis zur stehenden Oberfläche — alle 20 ms
# gefragt wie am Wirt, und ohne dessen halbe Sekunde danach.
sub starte {
  my ($o) = @_;
  $sock = "$o->{ort}/abnahme.sock";
  $port = 9190 + int(rand(400));
  unlink $sock;
  my $beginn = jetzt();
  $laeuft = fork // die "fork\n";
  if (!$laeuft) {
    exec('open', '-n', '-W', '-a', $APP, '--env', "HOME=$o->{heim}", '--env', "MSE_BESTAND=$o->{bestand}",
      '--env', "MSE_SCHLUESSELBUND=$o->{bund}", '--env', "MSE_ABNAHME=$sock", '--env', "WAILS_MCP_PORT=$port",
      '--env', 'WAILS_MCP_HIDE_CURSOR=1');
    exit 127;
  }
  warte('die Leitung', sub { -S $sock }, 30, 0.02);
  $schluessel = warte('der Schlüssel', sub { frage({ was => 'schluessel' }) }, 30, 0.02);
  warte('die Oberfläche steht', sub { seite($m->{oberflaecheSteht}) }, 30, 0.02);
  my $dauer = jetzt() - $beginn;
  my $pid = frage({ was => 'fenster' })->{prozess};
  return ($dauer, $pid);
}

sub beende {
  eval { taste('q', 'befehl') };
  my $bis = time() + 15;
  while (time() < $bis) {
    return 1 if waitpid($laeuft, WNOHANG) > 0;
    sleep 0.1;
  }
  return 0;
}

# klicke klickt nativ in die Mitte eines Elements — gesucht über JavaScript,
# geklickt über die Ereignisschlange, nachdem die Stelle zweimal gleich war.
sub klicke {
  my ($ausdruck) = @_;
  my $messen = "const e = ($ausdruck); if (!e || !e.checkVisibility()) return false; e.scrollIntoView({ block: 'center', behavior: 'instant' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }";
  my $stelle = warte("da ist: $ausdruck", sub { seite($messen) }, 15);
  for (1 .. 10) {
    sleep 0.05;
    my $neu = seite($messen) || $stelle;
    last if $neu->{x} == $stelle->{x} && $neu->{y} == $stelle->{y};
    $stelle = $neu;
  }
  frage({ was => 'klicke', x => $stelle->{x} + 0, y => $stelle->{y} + 0 });
  sleep 0.2;
}

sub textziel { my ($text) = @_; my $t = $m->{textziel}; my $q = $W->encode($text); $t =~ s/__TEXT__/$q/g; return $t }

# aufraeumen beendet, was noch läuft, und nimmt Bünde und Orte weg —
# **wiederholbar**: ein zweiter Ruf, etwa aus `END` nach einem Abbruch,
# findet nichts mehr vor.
sub aufraeumen {
  my @laufend = eigene();
  kill 'TERM', @laufend if @laufend;
  # **Bis sie weg sind**, höchstens zehn Sekunden — dann hart.
  my $bis = time() + 10;
  while (time() < $bis && (@laufend = eigene())) { sleep 0.2 }
  kill 'KILL', @laufend if @laufend;
  kill 'TERM', $laeuft if $laeuft && waitpid($laeuft, WNOHANG) == 0;
  system('security', 'delete-keychain', $_->{bund}) for grep { -e $_->{bund} } @orte;
  system('rm', '-rf', $_->{ort}) for @orte;
  @orte = ();
  $laeuft = 0;
  return;
}

# **Auch nach einem Abbruch** — `die` läuft durch `END`, ein Signal durch den
# Handler; beide räumen auf (Befund 2 des Zweitblicks über Phase I).
$SIG{INT} = $SIG{TERM} = sub { aufraeumen(); exit 130 };
END { aufraeumen() }

1;
