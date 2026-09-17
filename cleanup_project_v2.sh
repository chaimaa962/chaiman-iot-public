#!/bin/bash
# ═══════════════════════════════════════════════════════════════
# cleanup_project_v3.sh — Nettoyage SÛR et RAPIDE du projet CHAIMAN IoT
# ═══════════════════════════════════════════════════════════════
#
# [FIX v2 → v3] La v2 descendait DANS chaque node_modules/venv trouvé
#   et listait tous les node_modules IMBRIQUÉS un par un (des milliers
#   de lignes, très lent, Ctrl+C nécessaire). Corrigé avec "-prune" :
#   dès qu'un dossier node_modules/venv est trouvé, on ne descend plus
#   dedans (inutile : le supprimer supprime tout son contenu de toute
#   façon). Résultat : scan quasi instantané au lieu de plusieurs minutes.
#
# PRINCIPE (inchangé) :
#   - Mode par défaut = DRY-RUN (affiche ce qu'il ferait, ne touche à rien)
#   - "--apply" = déplace les fichiers "sûrs" vers une archive datée
#     (jamais de suppression définitive directe, sauf catégorie A qui
#     est 100% régénérable : node_modules, venv, __pycache__, logs)
#   - Les fichiers SENSIBLES (clés privées, mots de passe, certs) ne
#     sont JAMAIS déplacés ni supprimés automatiquement — juste listés.
#
# USAGE :
#   cd ~/iot-docker-project
#   bash cleanup_project_v3.sh              # dry-run
#   bash cleanup_project_v3.sh --apply       # applique réellement
# ═══════════════════════════════════════════════════════════════
set -uo pipefail
APPLY=false
if [ "${1:-}" == "--apply" ]; then
    APPLY=true
fi

PROJECT_ROOT="$(pwd)"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ARCHIVE_DIR="${PROJECT_ROOT}/_cleanup_archive_${TIMESTAMP}"

echo "═══════════════════════════════════════════════════════════════"
echo "🧹 Nettoyage du projet — $([ "$APPLY" = true ] && echo 'MODE APPLICATION' || echo 'MODE DRY-RUN (rien ne sera modifié)')"
echo "   Racine : ${PROJECT_ROOT}"
echo "═══════════════════════════════════════════════════════════════"
echo ""

if [ ! -d ".git" ] && [ ! -f "docker-compose.yml" ]; then
    echo "⚠️  Ce dossier ne ressemble pas à la racine du projet."
    echo "   Lance ce script depuis ~/iot-docker-project."
    read -p "   Continuer quand même ? (o/N) " confirm
    if [ "$confirm" != "o" ] && [ "$confirm" != "O" ]; then
        exit 1
    fi
fi

do_action() {
    local action="$1"   # "rm" ou "mv"
    local src="$2"
    local dst="${3:-}"

    if [ "$action" == "rm" ]; then
        local size
        size=$(du -sh "$src" 2>/dev/null | cut -f1)
        if [ "$APPLY" = true ]; then
            rm -rf "$src"
            echo "   ✅ supprimé ($size) : $src"
        else
            echo "   [dry-run] serait supprimé ($size) : $src"
        fi
    elif [ "$action" == "mv" ]; then
        local size
        size=$(du -sh "$src" 2>/dev/null | cut -f1)
        if [ "$APPLY" = true ]; then
            mkdir -p "$(dirname "$dst")"
            mv "$src" "$dst"
            echo "   ✅ déplacé ($size) : $src  →  $dst"
        else
            echo "   [dry-run] serait déplacé ($size) : $src  →  $dst"
        fi
    fi
}

# ═══════════════════════════════════════════════════════════════
echo "── Catégorie A : régénérable automatiquement (suppression directe) ──"
# [FIX] -prune empêche de descendre dans les node_modules/venv trouvés :
# on les traite comme des unités entières, pas fichier par fichier.
find "$PROJECT_ROOT" \
    \( -type d \( -name node_modules -o -name venv -o -name venv_sudo -o -name __pycache__ \) -prune \) \
    -print0 2>/dev/null | while IFS= read -r -d '' dir; do
    do_action rm "$dir"
done

# Logs et artefacts runtime régénérables (fichiers, pas dossiers)
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    \( -type f \( -name "*.log" -o -name "*.pid" \) -print0 \) 2>/dev/null | while IFS= read -r -d '' f; do
    do_action rm "$f"
done

echo ""

# ═══════════════════════════════════════════════════════════════
echo "── Catégorie B : fichiers .bak/.backup — archivés (pas supprimés) ──"
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    \( -type f \( -name "*.bak" -o -name "*.backup.*" -o -name "*.backup" \) -print0 \) 2>/dev/null | while IFS= read -r -d '' f; do
    relpath="${f#$PROJECT_ROOT/}"
    do_action mv "$f" "${ARCHIVE_DIR}/backups/${relpath}"
done
echo ""

# ═══════════════════════════════════════════════════════════════
echo "── Catégorie C : variantes deploy-*.js/.sh — archivées en bloc ──"
echo "   ⚠️  Vérifie lequel est utilisé RÉELLEMENT avant suppression finale"
echo "   (grep \"deploy\" iot-backend/package.json / README.md pour le savoir)"
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    \( -type f \( -iname "deploy-*.js" -o -iname "deploy-*.sh" -o -iname "deploy_*.js" -o -iname "deploy.js" \) -print0 \) 2>/dev/null | while IFS= read -r -d '' f; do
    relpath="${f#$PROJECT_ROOT/}"
    do_action mv "$f" "${ARCHIVE_DIR}/old_deploy_scripts/${relpath}"
done
echo ""

# ═══════════════════════════════════════════════════════════════
echo "── Catégorie D : fichiers scratch/debug — archivés pour relecture ──"
SCRATCH_NAMES=(res.txt resultas.txt cm.txt check.txt lance.txt lanc.txt block_res.txt INFO n)
SCRATCH_PATTERNS=("*.sh")  # sera filtré par nom court suspect ci-dessous
for name in "${SCRATCH_NAMES[@]}"; do
    find "$PROJECT_ROOT" \
        \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
        -type f -name "$name" -print0 2>/dev/null | while IFS= read -r -d '' f; do
        relpath="${f#$PROJECT_ROOT/}"
        do_action mv "$f" "${ARCHIVE_DIR}/scratch_review/${relpath}"
    done
done
# scripts à noms courts/cryptiques (1-3 lettres avant .sh) — souvent du scratch
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    -type f -regextype posix-extended -iregex '.*/[a-z]{1,4}\.sh$' -print0 2>/dev/null | while IFS= read -r -d '' f; do
    relpath="${f#$PROJECT_ROOT/}"
    do_action mv "$f" "${ARCHIVE_DIR}/scratch_review/${relpath}"
done
echo ""

# ═══════════════════════════════════════════════════════════════
echo "── Catégorie G : fichiers .bad (versions abandonnées) — archivés ──"
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    -type f -name "*.bad" -print0 2>/dev/null | while IFS= read -r -d '' f; do
    relpath="${f#$PROJECT_ROOT/}"
    do_action mv "$f" "${ARCHIVE_DIR}/bad_files/${relpath}"
done
echo ""

# ═══════════════════════════════════════════════════════════════
echo "── Catégorie H : fichiers avec typo évidente dans le nom ──"
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    -type f -name "pakage.json" -print0 2>/dev/null | while IFS= read -r -d '' f; do
    relpath="${f#$PROJECT_ROOT/}"
    do_action mv "$f" "${ARCHIVE_DIR}/typos/${relpath}"
done
echo ""

# ═══════════════════════════════════════════════════════════════
echo "═══════════════════════════════════════════════════════════════"
echo "🟠 CATÉGORIE F — DOSSIERS À VÉRIFIER MANUELLEMENT (NON TOUCHÉS)"
echo "═══════════════════════════════════════════════════════════════"
for suspect in "project" "iot-project.git" "backup_20260426" "ai-sandbox"; do
    found=$(find "$PROJECT_ROOT" -maxdepth 3 -type d -name "$suspect" 2>/dev/null)
    if [ -n "$found" ]; then
        echo "   ⚠️  $found existe — inspecte : ls -la \"$found\""
    fi
done
echo ""

# ═══════════════════════════════════════════════════════════════
echo "═══════════════════════════════════════════════════════════════"
echo "🔴 CATÉGORIE E — FICHIERS SENSIBLES DÉTECTÉS (JAMAIS TOUCHÉS)"
echo "═══════════════════════════════════════════════════════════════"
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    -type f \( -iname "*password*" -o -iname "*private_key*" -o -iname "*.pem" -o -iname "jwt-secret" -o -iname "*.key" -o -iname "bootnode.key" \) -print0 2>/dev/null | while IFS= read -r -d '' f; do
    echo "   ⚠️  $f"
done
find "$PROJECT_ROOT" \
    \( -path "*/node_modules" -o -path "*/venv" -o -path "*/venv_sudo" -o -path "*/_cleanup_archive_*" \) -prune -o \
    -type d \( -iname "my_account" -o -iname "extracted_keys" -o -iname "certs" \) -prune -print 2>/dev/null | while IFS= read -r d; do
    echo "   ⚠️  $d/"
done
echo ""
echo "   Ces fichiers ne sont ni déplacés ni supprimés par ce script."
echo "   Ils restent en place — à toi de décider (chiffrement, .gitignore, Vault...)."
echo ""

echo "═══════════════════════════════════════════════════════════════"
if [ "$APPLY" = true ]; then
    echo "✅ Nettoyage appliqué. Archive créée dans :"
    echo "   ${ARCHIVE_DIR}"
    echo "   Vérifie que le projet fonctionne toujours avant de supprimer"
    echo "   cette archive définitivement."
else
    echo "ℹ️  Ceci était un DRY-RUN — rien n'a été modifié."
    echo "   Relance avec --apply pour exécuter réellement ces actions :"
    echo "     bash cleanup_project_v3.sh --apply"
fi
echo "═══════════════════════════════════════════════════════════════"
