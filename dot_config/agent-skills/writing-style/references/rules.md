# Writing Rules

Each list gives the English form, then the French one. Apply the list matching the target text.

## 1. Modes

| Mode | Default |
| --- | --- |
| Technical (docs, README, PR) | clarity first, terms defined, steps shown, no decoration near important details |
| Published (article, post) | short paragraphs, varied rhythm, fragments allowed, `I`/`you` and English contractions when natural |
| Editing | name the problem, give the fix, show the better version; no praise before the edit |
| Message on the user's behalf | direct, no assistant chatter, a follow-up question only if the answer depends on it |
| Sensitive topic | calm and exact over punchy |
| Persuasion | proof over adjectives |

Everywhere: start with the useful answer, state uncertainty plainly ("probably", "je pense"), take a
stance when evidence supports one, stop once the point is made.

## 2. Formatting

- Digits for quantities: 3 ans, 10 tools.
- No em dash, no middle dot: period, comma, colon, parentheses, or a plain `-`.
- Bold at most twice per section; headers and bullets only when they help scanning.
- Sentence case in headers, which is also the French norm.
- No summary paragraph unless the text is long enough to need one.

## 3. Rejected-frame contrast

The most recognisable tic. A sentence, pair of sentences, heading or conclusion fails when it
dismisses X, then asserts Y in its place. The word "not" need not appear.

| English | French |
| --- | --- |
| It isn't X. It's Y. / Not X. Y. | Ce n'est pas X, c'est Y. / Pas X. Y. |
| Not only X, but also Y. | Non seulement X, mais aussi Y. |
| It's not about X, it's about Y. | Il ne s'agit pas de X, mais de Y. |
| The real problem isn't X. It's Y. | Le vrai problème n'est pas X, c'est Y. |
| Less X, more Y. / Forget X. Focus on Y. | Moins de X, plus de Y. / Oubliez X. |
| Most people think X... / While X may seem... | On pense souvent que X... / Si X peut sembler... |
| At first glance, X... / Sure, X... | À première vue, X... / Certes, X... |
| Is it X? No. It's Y. | X ? Non. Y. |

| Many assume X... / Conventional wisdom says X... | Beaucoup pensent que X... / On a coutume de dire que X... |
| People focus on X... / X gets all the attention... | On se focalise sur X... / X attire toute l'attention... |
| X sounds right... / X looks like the problem... | X semble juste... / X a l'air d'être le problème... |
| The question isn't X. The question is Y. | La question n'est pas X, c'est Y. |
| You don't need X. You need Y. | Vous n'avez pas besoin de X, mais de Y. |
| X is dead. Y is the future. / X is overrated. | X est mort, place à Y. / X est surcoté. |
| It was never about X. It was always about Y. | Ça n'a jamais été une question de X. |

The ban holds across sentence boundaries, where it is hardest to spot:

- "Most teams think they have a hiring problem. They have a standards problem." becomes "The team's
  standards are unclear."
- "Le tableau de bord ressemble à un outil de reporting. C'est en fait un filtre de décision."
  becomes "Le tableau de bord filtre les décisions."
- "People blame the algorithm. The input data is broken." becomes "The input data is broken."

A question that rejects one idea to install another fails the same way: "Is this a productivity
problem? No. It's an attention problem." becomes "Attention is the constraint." / "La vraie
question : combien de contrôle avez-vous ?" becomes "Question utile : combien de contrôle
avez-vous ?". Ask a question only when the reader must answer it.

Pivot words are fine in normal use and fail only when they perform the reframe: but, yet, actually,
really, instead, ultimately, the truth is, the real / mais, pourtant, en réalité, en fait, au fond,
finalement, le vrai, la vraie question.

Reframe headings fail too: "Not a tool. A system.", "Beyond X", "From chaos to clarity" / "Plus
qu'un outil", "Au-delà de X", "Du chaos à la clarté", "Ce qui compte vraiment". Name the subject
instead: "Le système", "Règles de décision".

**Fix:** delete the rejected half, then state the claim directly.
"Ce n'est pas le prompt qui compte, c'est le contexte." becomes "Le contexte détermine la sortie."

**Allowed:** correcting a fact, date, number, name, scope, or a legal or technical distinction.
"Le fichier fait 12 Mo, pas 12 Go."

## 4. Vocabulary

Inflated words, replace with the plain one unless it is the exact technical term:

- English: delve, realm, unlock, tapestry, paradigm, cutting-edge, revolutionize, intricate,
  showcase, crucial, pivotal, meticulous, vibrant, unparalleled, underscore, leverage, synergy,
  innovative, game-changer, testament, highlight, emphasize, boast, groundbreaking, foster, enhance,
  holistic, garner, unleash, versatile, transformative, redefine, seamless, streamline,
  frictionless, elevate, effortless, empower, visionary, disruptive, reimagine, unprecedented,
  state-of-the-art, immersive, turnkey, future-proof, supercharge, interplay, valuable, captivate.
- French: crucial, essentiel, primordial, incontournable, véritable, au cœur de, clé (adjectif),
  levier, écosystème, synergie, révolutionner, transformer en profondeur, sublimer, booster,
  décupler, fluide, sans couture, clé en main, novateur, innovant, holistique, puissant, riche,
  précieux, mettre en lumière, souligner, témoigner de, s'inscrire dans, repenser.

Bloated verbs that dodge "is" or "has":

- English: serves as, stands as, marks a, represents a, boasts, features, offers, plays a role in,
  helps to, aims to, seeks to.
- French: sert de, constitue, représente, se veut, fait office de, joue un rôle dans, permet de,
  vise à, a pour vocation de, se positionne comme.

Use: is, has, uses, shows, causes, removes / est, a, utilise, montre, provoque, supprime.

Keep the verb when it states a real capability or purpose: "`--dry-run` permet de vérifier sans
écrire" is exact; "cet outil permet de gagner en efficacité" is not.

## 5. Dead phrases

- Openings: In today's..., It is important to note, It's worth noting, Let's dive in, Let's
  explore, In this article I will / À l'heure où, Il est important de noter, Il convient de
  souligner, Plongeons dans, Dans cet article, nous allons.
- Transitions: Furthermore, Additionally, Moreover, That said, With that in mind, On top of that /
  De plus, En outre, Par ailleurs, Cela dit, Dans cette optique, Qui plus est. Use a real logical
  link or none.
- Filler: In order to, At the end of the day, Moving forward, In other words, It goes without
  saying / Afin de, Au final, À terme, En d'autres termes, Il va sans dire.
- Engagement bait: Let that sink in, Read that again, Full stop, This changes everything / Relisez
  bien, Point final, Ça change tout.
- Assistant chatter: Certainly, Great question, Happy to help, I hope this helps / Bien sûr,
  Excellente question, J'espère que cela vous aide.
- Cutoff disclaimers: As of my last update / À ma connaissance à ce jour. Verify instead.

## 6. Analogies and metaphors

Default: none. An analogy passes only if the subject is unfamiliar, the analogy makes it easier,
is shorter than the literal version, is exact enough not to mislead, and sounds normal aloud.

- Under 800 words: 0. Up to 1,500 words: at most 1. Beyond: 1 per 1,500 words, never 2 in a section,
  never extended over paragraphs.
- Setups to audit: think of it as, imagine, picture, it's like, acts as, the backbone of, the DNA of
  / imaginez, c'est comme, agit comme, la colonne vertébrale de, l'ADN de, le socle de, la boussole.
- Metaphor families to avoid for abstract subjects: journey, battle, engine and fuel, map and
  compass, bridge, north star, flywheel, iceberg, ecosystem, toolbox, sport, chess.
- Metaphor verbs for abstract work: baked in, woven, distilled, unpacked, sharpened, surfaced,
  anchored, fueled / tissé, distillé, décortiqué, ancré, nourri, cristallisé, catalysé. Use literal
  verbs: cut, added, removed, joined, explained, fixed / coupé, ajouté, retiré, expliqué, corrigé.

"Your onboarding is a leaky bucket" becomes "42% of users leave at step 2, where the form asks for
billing details before showing the product."

## 7. Other machine patterns

- **Inflated significance** - "a pivotal moment", "paving the way for" / "un tournant majeur",
  "ouvre la voie à". State the fact and let the reader weigh it.
- **Rule of three** - list the number of items that is true, not three by reflex.
- **False range** - "from ancient traditions to modern innovation" / "de X à Y" with no middle.
- **Elegant variation** - repeat the name or use a pronoun rather than a new epithet for the same
  person or thing.
- **Participle fake depth** - ", highlighting its importance" / ", soulignant ainsi son
  importance". Give real analysis its own sentence.
- **Meta commentary** - "In this section", "Here is a comprehensive overview" / "Dans cette
  section", "Voici un tour d'horizon". Say the thing.
- **Metronome rhythm** - same-length sentences and paragraphs in a row.

## 8. Specificity

"The company faced challenges" becomes "The company missed payroll twice in 6 months."
"L'outil améliore le processus" becomes "L'outil supprime 4 e-mails de validation par facture."
Prefer a real example to "imagine a scenario where".
