# ADR-020 - La barrière rejouée en CI, avec la clé `age` en secret

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : `2e03b13` (workflow et action composite)

## Contexte

`AGENTS.md` et l'en-tête de [scripts/verify.sh](../../scripts/verify.sh) affirmaient
« rien ne tourne en CI, donc tout se vérifie ici ». C'était vrai et suffisant
tant que le dépôt n'avait qu'un poste et un opérateur. Deux limites l'ont rendu
insuffisant.

**La barrière ne prouve pas le dépôt.** Elle rend les templates
(`chezmoi execute-template`) sur trois combinaisons de profil, ce qui vérifie
qu'ils se rendent, pas que les fichiers atterrissent où il faut avec les
attributs voulus. `chezmoi diff`, seul contrôle du dépôt effectif, est une
lecture manuelle - un des « à vérifier à la main » de `AGENTS.md`.

**Elle ne tourne que sur la machine de l'auteur.** Le dépôt cible macOS, Linux
et DSM ([ADR-008](008-dsm-cible-de-premier-rang.md)), et la règle
« nommer l'environnement » de `harness/AGENTS.md` reconnaît qu'un vert sur macOS
ne dit rien de Linux. En pratique, aucune vérification Linux n'avait lieu avant
qu'une machine ne casse.

Le contrôle des noms sensibles complique la mise en CI : sa liste de motifs est
elle-même la donnée à protéger, donc elle vient d'un fragment chiffré déployé, et
son absence rend le contrôle *non fait* - `verify.sh` sort rouge, à dessein
([ADR-016](016-depot-public-sensible-chiffre.md)). Une CI sans clé `age`
n'aurait donc pas pu être verte.

## Décision

`.github/workflows/verify.yml` tourne sur `push` vers `main`, sur
`pull_request` vers `main` et sur `workflow_dispatch`, en `permissions:
contents: read`. Il porte deux jobs :

- **`barriere`** rejoue `bash scripts/verify.sh` sur `ubuntu-latest`, avec
  `fetch-depth: 0` pour que le contrôle des messages de commit ait
  `origin/main..HEAD`, et `SENSIBLE_LIST` pointé sur un
  `chezmoi cat ~/.config/dotfiles/sensible.txt`.
- **`depot`** fait un vrai `chezmoi apply --exclude=scripts,externals` sur la
  matrice `{ubuntu-latest, macos-latest} × {perso, pro}`, dans le `$HOME` du
  runner - jetable par construction - puis exige un `chezmoi status` vide, ce
  qui teste l'idempotence.

**Aucune commande de la CI n'écrit le contenu rendu d'une cible.** Les logs d'un
dépôt public sont lisibles par tous, et un fragment déchiffré en sort aussi
sûrement que d'un fichier versionné. `apply --verbose` et `chezmoi diff`
émettent un diff unifié des fichiers écrits, donc le clair des sept fragments
`age` sur un `$HOME` vierge : ni l'un ni l'autre n'est admis, et `verify.sh`
refuse désormais leur retour dans `.github/`. L'idempotence se lit dans
`chezmoi status`, qui ne donne que des chemins.

L'action composite `.github/actions/setup-chezmoi` installe le binaire `chezmoi`
amont à une version épinglée, écrit la clé privée, sème un `chezmoi.toml`
réduit à `[data]` puis lance `chezmoi init`. Semer les données plutôt que
passer `--promptChoice` : la clé de ce flag est le **texte de la question** et
non le nom du champ, ce qui couperait la CI à la moindre reformulation, tandis
que `promptChoiceOnce` rend sans rien demander une valeur déjà présente dans la
configuration - le mécanisme dont `verify.sh` se sert déjà. Chaque appel porte
`--source` : `init` n'inscrit pas la source dans la configuration, et
`CHEZMOI_SOURCE_DIR` est une variable que chezmoi produit, pas une qu'il lit.

**Le secret `AGE_KEY` porte la clé privée `age`, pas la passphrase de
`age-key.txt.age`** : `age` ne déchiffre une clé protégée par passphrase qu'au
travers d'un terminal, absent d'un runner. Une clé absente ou inopérante fait
échouer le job.

`.github` est déjà listé dans `.chezmoiignore` : rien de tout cela n'est
déployé.

## Conséquences

- Linux et macOS sont vérifiés à chaque poussée, sur les deux profils. C'est le
  gain principal : le rendu des fragments `pro` n'était jusqu'ici jamais exercé
  depuis une machine `perso`.
- **DSM reste hors couverture** : aucun runner. La règle « nommer
  l'environnement » continue de s'appliquer, et un changement qui touche DSM
  reste une vérification manuelle.
- La clé privée `age` vit dans un secret GitHub. Elle ouvre tous les fragments
  d'un dépôt public : la compromettre équivaut à les publier. En conséquence,
  aucun `pull_request_target`, aucune action tierce, et `actions/checkout` seul
  y a accès.
- **Le canal des logs est aussi dangereux que l'exfiltration de la clé**, et
  moins visible : la première version de ce workflow portait `--verbose` sur
  `apply`, et le run
  [33739550006](https://github.com/wallforfry/dotfiles/actions/runs/33739550006)
  a publié en clair les six cibles chiffrées avant que la revue ne l'arrête. Le
  masquage de secrets de GitHub ne couvre que la valeur de `AGE_KEY`, jamais les
  clairs qu'elle produit. Toute commande ajoutée à ce workflow se juge sur ce
  critère.
- **Une PR de fork sortira rouge**, faute de secret. Assumé sur un dépôt à un
  auteur : mieux vaut un rouge lisible qu'un vert obtenu en sautant le contrôle.
- Le job `depot` ne teste pas l'amorçage d'une machine : `--exclude=scripts`
  écarte l'installation d'outils, `--exclude=externals` le téléchargement
  d'oh-my-zsh. L'ADR-007 « dégradation, jamais échec » reste vérifiée à la main.
- Deux versions épinglées de plus à suivre : celle de `chezmoi` dans l'action,
  celle de `actions/checkout`.

## Alternatives écartées

- **Statu quo, la barrière locale seule.** C'est ce que le dépôt affirmait
  vouloir. Rejeté parce qu'elle ne couvre ni le dépôt effectif ni un autre OS
  que celui de l'auteur, précisément les deux zones où les défauts observés se
  produisent.
- **Une CI sans clé `age`.** Elle éviterait le secret, mais laisserait le
  contrôle des noms sensibles non fait, donc le job rouge en permanence
  (ADR-016) - ou bien exigerait de rendre ce contrôle facultatif, ce qui
  reviendrait à publier un vert sur le contrôle qui protège un dépôt public.
- **La passphrase en secret plutôt que la clé.** Le premier réflexe, et
  l'expression littérale du besoin. Rejeté après vérification :
  `age --decrypt` sur un fichier protégé par passphrase réclame un terminal, que
  le runner n'a pas ; le contourner par un pseudo-terminal ajouterait de la
  mécanique pour une valeur strictement aussi sensible que la clé.
- **`chezmoi apply` complet, scripts et externals compris.** Couverture
  maximale, mais chaque job téléchargerait la chaîne d'outils entière et
  oh-my-zsh, pour tester l'amorçage - un besoin distinct de celui-ci, et payé en
  minutes macOS.
- **`chezmoi init --promptChoice profile=… --promptBool gui=false`.** Première
  version, rouge sur les cinq jobs du premier run : `could not open a new TTY:
  open /dev/tty: no such device or address`. La clé attendue par ces flags est
  le texte de la question - `--promptChoice "Profil de cette machine=perso"`
  fonctionne, mais lie la CI à la formulation d'une invite.
- **Une action tierce d'installation de chezmoi.** Un `curl` d'une version
  épinglée fait la même chose sans donner accès à un job qui manipule la clé
  privée.
