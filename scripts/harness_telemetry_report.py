def rate(numerator, denominator):
    return "inconnu" if not denominator else f"{numerator / denominator:.3f}"


def print_counter(title, values, limit=None):
    print(f"  {title} :")
    ordered = sorted(values.items(), key=lambda item: (-item[1], item[0]))
    for name, count in ordered[:limit]:
        print(f"    {count:5d}  {name}")


def activated_body_cost(summary, local_sizes):
    activations = {name: summary["skills"].get(name, 0) for name in local_sizes}
    return sum(activations.values()), sum(activations[name] * local_sizes[name] for name in local_sizes)


def print_report(sources, total, by_source, hits, elapsed, local_sizes, unknown, none, eras):
    print("  sources : " + ", ".join(f"{source} {count} sessions" for source, count in sorted(sources.items())))
    if total["days"]:
        print(f"  période : {min(total['days'])} au {max(total['days'])}")
    else:
        print("  période : inconnue")
    print(f"  cache : {hits}/{sum(sources.values())} sessions réutilisées, {elapsed} ms")
    print(
        "  formats : "
        f"{sum(total['records'].values())} enregistrements reconnus, "
        f"{sum(total['unknown_records'].values())} inconnus, "
        f"{sum(total['invalid_records'].values())} invalides"
    )
    print("  signal d'activation des skills :")
    for source in sorted(by_source):
        activations = sum(by_source[source]["skills"].values())
        status = f"{activations} événements explicites" if activations else "inconnu, aucun événement explicite"
        print(f"    {source:8s} {status}")
    print("  coût amorti des corps de skills locales :")
    for source in sorted(by_source):
        activations, body_bytes = activated_body_cost(by_source[source], local_sizes)
        if not sum(by_source[source]["skills"].values()):
            print(f"    {source:8s} inconnu")
        else:
            average = body_bytes / sources[source] if sources[source] else 0
            print(f"    {source:8s} {body_bytes} octets sur {activations} activations, {average:.1f} octets/session")
    print_counter("skills de ce dépôt", {name: total["skills"].get(name, 0) for name in local_sizes})
    external = {name: count for name, count in total["skills"].items() if name not in local_sizes}
    print_counter("autres skills activées, top 5", external, 5)
    print_counter("subagents", total["agents"], 6)
    print(
        "  qualification absente : "
        f"inconnue {total['skills'].get(unknown, 0) + total['agents'].get(unknown, 0)}, "
        f"explicitement nulle {total['skills'].get(none, 0) + total['agents'].get(none, 0)}"
    )
    print("  ponctuation interdite par bloc de texte assistant :")
    for period in eras:
        blocks = total["blocks"].get(period, 0)
        dash = total["dash"].get(period, 0)
        dot = total["middle_dot"].get(period, 0)
        print(f"    {period:8s} {blocks:6d} blocs  cadratin {dash:5d} ({rate(dash, blocks)})  point médian {dot:5d} ({rate(dot, blocks)})")
    print("  lignes de commentaire dans le code écrit :")
    for period in eras:
        lines = total["lines"].get(period, 0)
        comments = total["comments"].get(period, 0)
        print(f"    {period:8s} {lines:6d} lignes  {comments:6d} commentaires  {rate(comments, lines)}")
    print_counter("canaux d'écriture observés", total["write_tools"])
    if total["uninspectable_writes"]:
        print_counter("écritures au contenu inconnu", total["uninspectable_writes"])
