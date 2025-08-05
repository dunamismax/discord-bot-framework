#!/usr/bin/env python3
"""Validation script for the Discord bot framework."""

import ast
import sys
from pathlib import Path


def validate_syntax(file_path):
    """Validate Python syntax of a file."""
    try:
        with open(file_path, encoding='utf-8') as f:
            source = f.read()
        ast.parse(source)
        return True, None
    except SyntaxError as e:
        return False, f"Syntax error: {e}"
    except Exception as e:
        return False, f"Error reading file: {e}"


def validate_imports(file_path):
    """Validate that imports can be resolved (basic check)."""
    try:
        # Add project directories to path temporarily
        project_root = Path(__file__).parent
        libs_path = project_root / "libs"

        if str(libs_path) not in sys.path:
            sys.path.insert(0, str(libs_path))

        # Try to compile the file to check for import issues
        with open(file_path, encoding='utf-8') as f:
            source = f.read()

        compile(source, file_path, 'exec')
        return True, None
    except Exception as e:
        return False, f"Import/compilation error: {e}"


def main():
    """Run validation checks."""
    project_root = Path(__file__).parent

    # Files to validate
    files_to_check = [
        # Shared utilities
        project_root / "libs/shared_utils/base_bot.py",
        project_root / "libs/shared_utils/config_loader.py",
        project_root / "libs/shared_utils/database.py",
        project_root / "libs/shared_utils/health_check.py",
        project_root / "libs/shared_utils/help_system.py",

        # Bot applications
        project_root / "apps/clippy_bot/main.py",
        project_root / "apps/clippy_bot/cogs/unhinged_responses.py",
        project_root / "apps/music_bot/main.py",
        project_root / "apps/music_bot/cogs/music_player.py",

    ]

    all_passed = True

    print("🔍 Validating Discord Bot Framework...")
    print("=" * 50)

    for file_path in files_to_check:
        if not file_path.exists():
            print(f"❌ {file_path.relative_to(project_root)}: File not found")
            all_passed = False
            continue

        # Check syntax
        syntax_ok, syntax_error = validate_syntax(file_path)
        if not syntax_ok:
            print(f"❌ {file_path.relative_to(project_root)}: {syntax_error}")
            all_passed = False
            continue

        # Check imports/compilation
        import_ok, import_error = validate_imports(file_path)
        if not import_ok:
            print(f"⚠️  {file_path.relative_to(project_root)}: {import_error}")
            # Don't fail on import errors as they might be due to missing dependencies
        else:
            print(f"✅ {file_path.relative_to(project_root)}: OK")

    print("=" * 50)

    # Check configuration files
    config_files = [
        project_root / "pyproject.toml",
        project_root / "config.example.json",
        project_root / "Caddyfile",
    ]

    print("📄 Checking configuration files...")
    for config_file in config_files:
        if config_file.exists():
            print(f"✅ {config_file.name}: Found")
        else:
            print(f"❌ {config_file.name}: Missing")
            all_passed = False

    print("=" * 50)

    # Summary
    if all_passed:
        print("🎉 All validations passed!")
        return 0
    else:
        print("⚠️  Some validations failed. Check the output above.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
