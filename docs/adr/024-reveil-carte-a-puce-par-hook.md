# ADR-024 - Réveil de la carte à puce par hook plutôt que par instruction

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : `27a50f0`

## Contexte

Sur le poste macOS, `scdaemon` dialogue avec la YubiKey via PC/SC, donc via le
service système `com.apple.ctkpcscd` :

```
39743 gpg-agent --homedir /Users/…/.gnupg --use-standard-socket --daemon
39824 scdaemon --multi-server
39825 …/com.apple.ctkpcscd.xpc/Contents/MacOS/com.apple.ctkpcscd
```

Une élévation temporaire de privilèges suivie du retour en utilisateur
standard fait repartir le contexte de sécurité de la session. Le `SCardContext` détenu par `scdaemon`
devient mort et `scdaemon` ne le rétablit jamais : la clé est branchée, mais
toute signature échoue jusqu'à ce qu'on le tue. Aucune option de
`scdaemon.conf` ne le fait se reconnecter - c'est une limite amont de GnuPG,
sur un logiciel que ce dépôt ne possède pas.

La reprise tenait dans un script manuel non versionné, `~/.wakeup`, qui faisait
`pkill -9 gpg-agent` puis relançait l'agent de MacGPG2. Il détruisait donc aussi
le cache SSH et les sockets, et il réinstallait un agent 2.2.41 sous un `gpg`
Homebrew 2.5.22, écart que `gpg --card-status` signale lui-même.

## Décision

La reprise est une commande du CLI, `dotfiles smartcard-wakeup` : elle sonde
`gpg --card-status` et, seulement en cas d'échec, exécute `gpgconf --kill
scdaemon` avant de re-sonder. Elle n'appelle que les binaires du `PATH`, jamais
ceux de MacGPG2. Elle est exposée en `~/.local/bin/smartcard-wakeup`.

Les agents la déclenchent par un hook `PostToolUse` sur `^Bash$`, enregistré par
`dotfiles register-claude-hook` : le hook n'examine que la réponse de l'outil,
ne fait rien si elle ne porte pas de signature d'échec carte, et ne parle que
s'il a réellement réparé - en sortant alors sur le code 2, seul code où Claude
Code transmet la sortie d'erreur d'un hook à l'agent, pour qu'il relance sa
commande. Aucune ligne n'est ajoutée aux fichiers toujours chargés.

## Conséquences

- Le coût en contexte est nul tant que rien n'échoue : le hook est un binaire
  qui sort immédiatement sur une charge utile sans signature.
- La liste des signatures est une heuristique, et `gpg` est localisé : elle doit
  couvrir le français et l'anglais, et elle manquera un message inédit. Deux
  garde-fous rendent un faux positif inoffensif : seule la réponse de l'outil est
  lue - sans quoi lire ce fichier suffirait à déclencher le hook - et le silence
  en cas d'échec, sans lequel une machine sans carte, Linux ou DSM, verrait
  échouer chaque commande dont la sortie porte une signature.
- `gpgconf --kill scdaemon` préserve `gpg-agent`, donc le cache SSH et les
  sockets, contrairement au `pkill -9` remplacé.
- `register-claude-hook` n'est plus figé sur `Stop`/`agent-handoff` : il porte
  une liste de hooks - événement, matcher, commande, délai - écrits en une seule
  réécriture de `settings.json` pour que la sauvegarde `.bak` reste l'état
  d'origine. Un hook déjà présent est reconnu à sa commande seule : un matcher
  divergent posé à la main n'est pas corrigé.
- Le réveil ne corrige pas la cohabitation MacGPG2 / Homebrew : la désinstaller
  reste un geste manuel de l'opérateur, hors périmètre de ce dépôt.

## Alternatives écartées

- **Une ligne dans `harness/AGENTS.md`.** Elle ferait payer à chaque session, sur
  toutes les machines, le coût permanent d'un incident rare et propre à un seul
  poste. Le hook agit au moment exact où l'information sert.
- **Un réglage de `scdaemon`.** Rien dans `scdaemon.conf` ne rétablit un contexte
  PC/SC mort ; le pilote CCID interne contournerait PC/SC mais macOS réserve
  l'interface CCID de la clé, donc il n'est pas fiable ici.
- **Versionner `~/.wakeup` tel quel.** Il perpétuait le mélange de versions et
  tuait l'agent entier ; le versionner aurait figé les deux défauts.
