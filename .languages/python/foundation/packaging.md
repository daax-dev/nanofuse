# Python Packaging Comprehensive Guide

Complete guide to Python packaging, distribution, and dependency management covering traditional and modern approaches. Updated for Python 3.9+ and current best practices as of 2024.

## Quick Start Checklist

✅ **Modern Python Packaging (2024)**
- Use `pyproject.toml` as primary configuration
- Adopt `src/` layout for better isolation
- Use `build` package for building distributions
- Leverage `uv` or `pip-tools` for dependency management
- Implement semantic versioning with automation
- Configure comprehensive CI/CD pipelines
- Include security scanning and vulnerability checks

## Package Structure Fundamentals

### Basic Package Structure

```
my_package/
├── src/
│   └── my_package/
│       ├── __init__.py
│       ├── core.py
│       └── utils.py
├── tests/
│   ├── __init__.py
│   ├── test_core.py
│   └── test_utils.py
├── docs/
│   ├── conf.py
│   └── index.rst
├── .gitignore
├── LICENSE
├── README.md
├── pyproject.toml
├── setup.py (legacy)
└── MANIFEST.in
```

### Package __init__.py

```python
# src/my_package/__init__.py
"""My Package - A sample Python package."""

__version__ = "0.1.0"
__author__ = "Your Name"
__email__ = "your.email@example.com"
__license__ = "MIT"

# Import main functionality
from .core import main_function, MainClass
from .utils import utility_function

# Define public API
__all__ = [
    "main_function",
    "MainClass",
    "utility_function",
]

# Package-level configuration
import logging

# Set up null handler to prevent logging errors
logging.getLogger(__name__).addHandler(logging.NullHandler())

# Optional: Package initialization
def _initialize_package():
    """Initialize package resources."""
    pass

_initialize_package()
```

## Modern Packaging with pyproject.toml

### Complete pyproject.toml Example

```toml
[build-system]
requires = ["setuptools>=68.0", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "my-package"
version = "0.1.0"
description = "A sample Python package"
readme = "README.md"
license = {text = "MIT"}
authors = [
    {name = "Your Name", email = "your.email@example.com"},
]
maintainers = [
    {name = "Maintainer Name", email = "maintainer@example.com"},
]
keywords = ["sample", "package", "python"]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Operating System :: OS Independent",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
    "Programming Language :: Python :: 3.12",
    "Programming Language :: Python :: 3.13",
    "Topic :: Software Development :: Libraries :: Python Modules",
]
requires-python = ">=3.9"
dependencies = [
    "requests>=2.25.0",
    "click>=8.0.0",
    "pydantic>=2.0.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=8.0.0",
    "pytest-cov>=5.0.0",
    "pytest-xdist>=3.5.0",  # Parallel testing
    "ruff>=0.1.0",  # Modern linter/formatter
    "mypy>=1.8.0",
    "pre-commit>=3.6.0",
    "uv>=0.1.0",  # Fast dependency resolver
]
test = [
    "pytest>=8.0.0",
    "pytest-cov>=5.0.0",
    "pytest-mock>=3.12.0",
    "pytest-benchmark>=4.0.0",
    "hypothesis>=6.90.0",  # Property-based testing
]
docs = [
    "sphinx>=7.2.0",
    "sphinx-rtd-theme>=2.0.0",
    "myst-parser>=2.0.0",
    "sphinx-autodoc-typehints>=1.25.0",
    "sphinx-copybutton>=0.5.2",
]
lint = [
    "ruff>=0.1.0",  # Replaces black, isort, flake8
    "mypy>=1.8.0",
    "bandit>=1.7.5",  # Security linting
    "safety>=3.0.0",  # Dependency vulnerability scanning
]

[project.scripts]
my-cli = "my_package.cli:main"
my-tool = "my_package.tools:tool_main"

[project.entry-points."console_scripts"]
my-console = "my_package.console:console_main"

[project.entry-points."my_package.plugins"]
default = "my_package.plugins:default_plugin"

[project.urls]
Homepage = "https://github.com/username/my-package"
Documentation = "https://my-package.readthedocs.io/"
Repository = "https://github.com/username/my-package"
"Bug Tracker" = "https://github.com/username/my-package/issues"
Changelog = "https://github.com/username/my-package/blob/main/CHANGELOG.md"

[tool.setuptools]
package-dir = {"" = "src"}

[tool.setuptools.packages.find]
where = ["src"]

[tool.setuptools.package-data]
my_package = ["data/*.json", "templates/*.html"]

# Ruff configuration (replaces black, isort, flake8)
[tool.ruff]
line-length = 88
target-version = "py39"

[tool.ruff.lint]
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "UP",  # pyupgrade
    "S",   # bandit
]
ignore = [
    "E501",  # line too long (handled by formatter)
    "S101",  # assert usage (OK in tests)
]

[tool.ruff.lint.per-file-ignores]
"tests/**/*" = ["S101", "D"]

[tool.ruff.format]
quote-style = "double"
indent-style = "space"

[tool.mypy]
python_version = "3.9"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
strict = true
show_error_codes = true
namespace_packages = true
explicit_package_bases = true

[tool.pytest.ini_options]
testpaths = ["tests"]
addopts = [
    "--cov=my_package",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-report=xml",
    "--cov-fail-under=90",
    "--strict-markers",
    "--strict-config",
    "-ra",  # Show all test results
]
markers = [
    "slow: marks tests as slow (deselect with '-m \"not slow\"')",
    "integration: marks tests as integration tests",
    "unit: marks tests as unit tests",
]

[tool.coverage.run]
source = ["src"]
omit = [
    "tests/*",
    "*/test_*.py",
]

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "def __repr__",
    "raise AssertionError",
    "raise NotImplementedError",
]
```

### Dynamic Versioning

```python
# src/my_package/_version.py
"""Version information for my_package."""

__version__ = "0.1.0"

# Dynamic versioning with setuptools_scm
# pyproject.toml addition:
"""
[project]
dynamic = ["version"]

[tool.setuptools_scm]
write_to = "src/my_package/_version.py"
"""

# Using version in __init__.py
from ._version import __version__
```

## Legacy setup.py

### Complete setup.py Example

```python
# setup.py (legacy approach)
from setuptools import setup, find_packages
import os

# Read long description from README
def read_long_description():
    here = os.path.abspath(os.path.dirname(__file__))
    with open(os.path.join(here, 'README.md'), encoding='utf-8') as f:
        return f.read()

# Read version from package
def read_version():
    here = os.path.abspath(os.path.dirname(__file__))
    with open(os.path.join(here, 'src', 'my_package', '__init__.py')) as f:
        for line in f:
            if line.startswith('__version__'):
                return line.split('=')[1].strip().strip('"\'')
    raise RuntimeError('Unable to find version string.')

setup(
    name="my-package",
    version=read_version(),
    author="Your Name",
    author_email="your.email@example.com",
    description="A sample Python package",
    long_description=read_long_description(),
    long_description_content_type="text/markdown",
    url="https://github.com/username/my-package",
    project_urls={
        "Bug Tracker": "https://github.com/username/my-package/issues",
        "Documentation": "https://my-package.readthedocs.io/",
        "Source Code": "https://github.com/username/my-package",
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    package_dir={"": "src"},
    packages=find_packages(where="src"),
    python_requires=">=3.9",
    install_requires=[
        "requests>=2.25.0",
        "click>=8.0.0",
        "pydantic>=2.0.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "black>=23.0.0",
            "isort>=5.10.0",
            "flake8>=6.0.0",
            "mypy>=1.0.0",
        ],
        "test": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
        ],
        "docs": [
            "sphinx>=5.0.0",
            "sphinx-rtd-theme>=1.2.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "my-cli=my_package.cli:main",
            "my-tool=my_package.tools:tool_main",
        ],
        "my_package.plugins": [
            "default=my_package.plugins:default_plugin",
        ],
    },
    package_data={
        "my_package": ["data/*.json", "templates/*.html"],
    },
    include_package_data=True,
    zip_safe=False,
)
```

## Advanced Packaging Features

### MANIFEST.in for Additional Files

```
# MANIFEST.in
include README.md
include LICENSE
include CHANGELOG.md
include requirements*.txt
include pyproject.toml

# Include data files
recursive-include src/my_package/data *.json *.yaml *.txt
recursive-include src/my_package/templates *.html *.jinja2

# Include tests
recursive-include tests *.py

# Include documentation
recursive-include docs *.rst *.md *.png *.jpg
include docs/Makefile
include docs/conf.py

# Exclude unwanted files
exclude .gitignore
exclude .pre-commit-config.yaml
recursive-exclude * __pycache__
recursive-exclude * *.py[co]
recursive-exclude * *.so
recursive-exclude * .DS_Store
```

### Entry Points and Console Scripts

```python
# src/my_package/cli.py
import click
from . import __version__

@click.group()
@click.version_option(version=__version__)
def main():
    """My Package CLI tool."""
    pass

@main.command()
@click.option('--name', default='World', help='Name to greet.')
def hello(name):
    """Say hello."""
    click.echo(f'Hello {name}!')

@main.command()
@click.argument('input_file', type=click.File('r'))
@click.option('--output', '-o', type=click.File('w'), default='-')
def process(input_file, output):
    """Process a file."""
    content = input_file.read()
    processed = content.upper()
    output.write(processed)

if __name__ == '__main__':
    main()

# Plugin system example
# src/my_package/plugins.py
class PluginBase:
    """Base class for plugins."""

    def execute(self, *args, **kwargs):
        raise NotImplementedError

def default_plugin():
    """Default plugin implementation."""

    class DefaultPlugin(PluginBase):
        def execute(self, data):
            return f"Default processing: {data}"

    return DefaultPlugin()

# Using entry points to discover plugins
import pkg_resources

def load_plugins():
    """Load all registered plugins."""
    plugins = {}
    for entry_point in pkg_resources.iter_entry_points('my_package.plugins'):
        plugin = entry_point.load()
        plugins[entry_point.name] = plugin()
    return plugins
```

### Resource Management

```python
# src/my_package/resources.py
"""Resource management utilities."""

import os
from pathlib import Path
try:
    from importlib.resources import files, as_file
except ImportError:
    # Python < 3.9 fallback
    from importlib_resources import files, as_file

def get_data_file(filename):
    """Get path to data file."""
    package_files = files('my_package')
    return package_files / 'data' / filename

def read_config():
    """Read configuration from package data."""
    config_file = get_data_file('config.json')
    with as_file(config_file) as path:
        with open(path, 'r') as f:
            import json
            return json.load(f)

def get_template_path():
    """Get path to templates directory."""
    package_files = files('my_package')
    return package_files / 'templates'

# Legacy resource access (for older Python versions)
import pkg_resources

def get_data_legacy(filename):
    """Get data using pkg_resources (legacy)."""
    return pkg_resources.resource_string('my_package', f'data/{filename}')

def get_data_path_legacy(filename):
    """Get data file path using pkg_resources (legacy)."""
    return pkg_resources.resource_filename('my_package', f'data/{filename}')
```

## Building and Distribution

### Building Packages

```bash
# Install build tools (modern approach with uv)
pip install uv
uv tool install build
uv tool install twine

# Alternative: traditional pip approach
pip install build twine

# Build package (creates dist/ directory)
python -m build

# Fast build with uv
uv build

# Build only wheel
python -m build --wheel

# Build only source distribution
python -m build --sdist

# Clean build
rm -rf dist/ build/ *.egg-info/
python -m build

# Verify package contents
unzip -l dist/my_package-0.1.0-py3-none-any.whl
tar -tzf dist/my_package-0.1.0.tar.gz
```

### Publishing to PyPI

```bash
# Check package before upload
twine check dist/*

# Upload to TestPyPI first
twine upload --repository testpypi dist/*

# Test installation from TestPyPI
pip install --index-url https://test.pypi.org/simple/ my-package

# Upload to PyPI
twine upload dist/*

# Upload with API token
twine upload --username __token__ --password pypi-... dist/*
```

### PyPI Configuration

```toml
# ~/.pypirc
[distutils]
index-servers =
    pypi
    testpypi

[pypi]
username = __token__
password = pypi-AgEIcHl...

[testpypi]
repository = https://test.pypi.org/legacy/
username = __token__
password = pypi-AgEIcHl...
```

## Dependency Management

### Dependency Specification

```toml
# pyproject.toml dependency specification
[project]
dependencies = [
    # Exact version
    "requests==2.28.1",

    # Minimum version
    "click>=8.0.0",

    # Compatible version
    "pydantic~=2.0.0",  # >=2.0.0, <2.1.0

    # Version range
    "numpy>=1.20.0,<2.0.0",

    # Pre-release versions
    "some-package>=1.0.0a1",

    # VCS dependencies
    "git+https://github.com/user/repo.git@tag#egg=package",

    # Local dependencies
    "local-package @ file:///path/to/local/package",

    # Platform-specific dependencies
    "pywin32>=227; sys_platform == 'win32'",
    "colorama>=0.4.0; sys_platform == 'win32'",

    # Python version specific
    "importlib-metadata>=1.0; python_version < '3.8'",
    "typing-extensions>=3.7; python_version < '3.8'",
]

[project.optional-dependencies]
# Development dependencies
dev = [
    "pytest>=7.0.0",
    "pytest-cov>=4.0.0",
    "black[d]>=23.0.0",  # With optional dependency
    "isort>=5.10.0",
]

# Extra dependencies for specific use cases
crypto = [
    "cryptography>=3.0.0",
    "pyopenssl>=20.0.0",
]

performance = [
    "cython>=0.29.0",
    "numba>=0.56.0",
]
```

### Requirements Files

```txt
# requirements.txt - Production dependencies
requests>=2.25.0,<3.0.0
click>=8.0.0
pydantic>=2.0.0,<3.0.0

# requirements-dev.txt - Development dependencies
-r requirements.txt
pytest>=7.0.0
pytest-cov>=4.0.0
black>=23.0.0
isort>=5.10.0
flake8>=6.0.0
mypy>=1.0.0
pre-commit>=3.0.0

# requirements-docs.txt - Documentation dependencies
-r requirements.txt
sphinx>=5.0.0
sphinx-rtd-theme>=1.2.0
myst-parser>=0.18.0

# Constraints file (constraints.txt)
# Pin transitive dependencies for reproducible builds
urllib3==1.26.12
certifi==2022.9.24
charset-normalizer==2.1.1
```

### Lock Files and Reproducible Builds

```bash
# Modern approach with uv (fastest)
uv pip compile pyproject.toml -o requirements.txt
uv pip compile --extra dev pyproject.toml -o requirements-dev.txt
uv pip sync requirements.txt

# Traditional pip-tools approach
pip install pip-tools
pip-compile pyproject.toml
pip-compile --extra dev pyproject.toml -o requirements-dev.txt
pip-compile --upgrade pyproject.toml
pip-sync requirements.txt

# Using Poetry for lock files
poetry lock --no-update
poetry install --sync
poetry export -f requirements.txt --output requirements.txt
poetry export --only=dev -f requirements.txt --output requirements-dev.txt

# Using PDM for lock files
pdm lock
pdm install --frozen-lockfile
pdm export -f requirements --output requirements.txt
pdm export -G dev -f requirements --output requirements-dev.txt
```

## Advanced Packaging Patterns

### Namespace Packages

```python
# PEP 420 implicit namespace packages
# No __init__.py in namespace directory

# Directory structure:
# namespace/
#   package_a/
#     __init__.py
#     module.py
#   package_b/
#     __init__.py
#     module.py

# pyproject.toml
[tool.setuptools.packages.find]
where = ["src"]
include = ["namespace.*"]

# Usage
from namespace.package_a import module
from namespace.package_b import module
```

### Extension Modules (C Extensions)

```python
# setup.py for C extensions
from setuptools import setup, Extension
from pybind11.setup_helpers import Pybind11Extension, build_ext

ext_modules = [
    Pybind11Extension(
        "my_package._core",
        [
            "src/my_package/_core.cpp",
        ],
        include_dirs=[
            "include",
        ],
        language='c++',
    ),
]

setup(
    name="my-package",
    ext_modules=ext_modules,
    cmdclass={"build_ext": build_ext},
    zip_safe=False,
)

# pyproject.toml for extensions
[build-system]
requires = ["setuptools", "wheel", "pybind11"]

[project]
name = "my-package"
# ... other configuration
```

### Data Files and Resources

```python
# Including data files
# pyproject.toml
[tool.setuptools.package-data]
my_package = [
    "data/*.json",
    "data/*.yaml",
    "templates/*.html",
    "static/css/*.css",
    "static/js/*.js",
]

# Alternative approach
[tool.setuptools]
include-package-data = true

# MANIFEST.in (used with include-package-data)
recursive-include src/my_package/data *.json *.yaml
recursive-include src/my_package/templates *.html
recursive-include src/my_package/static *.css *.js *.png
```

## Modern Dependency Management Tools Comparison

### Tool Selection Matrix

| Tool | Speed | Features | Ecosystem | Best For |
|------|-------|----------|-----------|----------|
| **uv** | ⚡⚡⚡⚡⚡ | Basic but growing | Growing | Speed-first, simple projects |
| **Poetry** | ⚡⚡⚡ | ⭐⭐⭐⭐⭐ | Mature | Full-featured development |
| **PDM** | ⚡⚡⚡⚡ | ⭐⭐⭐⭐ | Growing | PEP 582, modern standards |
| **pipenv** | ⚡⚡ | ⭐⚡ | Declining | Legacy projects |
| **pip-tools** | ⚡⚡⚡ | ⭐⭐ | Stable | Minimalist approach |

### UV Workflow (Fastest - 2024)

```bash
# Install uv
pip install uv

# Create virtual environment
uv venv
source .venv/bin/activate  # Linux/Mac
# .venv\\Scripts\\activate  # Windows

# Install dependencies from pyproject.toml
uv pip install -e .

# Add new dependency
echo 'requests>=2.31.0' >> requirements.txt
uv pip install -r requirements.txt

# Development workflow
uv pip compile pyproject.toml -o requirements.txt
uv pip compile --extra dev pyproject.toml -o requirements-dev.txt
uv pip sync requirements-dev.txt

# Fast package building
uv build
```

## Modern Tools and Workflows

### Poetry Workflow

```bash
# Initialize new project
poetry new my-project
cd my-project

# Initialize existing project
poetry init

# Add dependencies
poetry add requests
poetry add --group dev pytest

# Install project
poetry install

# Build and publish
poetry build
poetry publish

# Version management
poetry version patch  # 0.1.0 -> 0.1.1
poetry version minor  # 0.1.1 -> 0.2.0
poetry version major  # 0.2.0 -> 1.0.0

# Export lock file
poetry export -f requirements.txt --output requirements.txt
```

### PDM Workflow

```bash
# Install PDM
pip install pdm

# Initialize project
pdm init

# Add dependencies
pdm add requests
pdm add -d pytest  # Development dependency

# Install dependencies
pdm install

# Run scripts
pdm run python main.py
pdm run pytest

# Build package
pdm build

# Publish package
pdm publish
```

### pipenv Workflow

```bash
# Initialize project
pipenv --python 3.11

# Install dependencies
pipenv install requests
pipenv install pytest --dev

# Install from Pipfile
pipenv install

# Activate shell
pipenv shell

# Run commands
pipenv run python script.py
pipenv run pytest

# Generate requirements
pipenv requirements > requirements.txt
```

## CI/CD Integration

### GitHub Actions for Publishing

```yaml
# .github/workflows/publish.yml
name: Publish to PyPI

on:
  release:
    types: [published]

jobs:
  build-and-publish:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install build twine

    - name: Build package
      run: python -m build

    - name: Check package
      run: twine check dist/*

    - name: Publish to PyPI
      env:
        TWINE_USERNAME: __token__
        TWINE_PASSWORD: ${{ secrets.PYPI_API_TOKEN }}
      run: twine upload dist/*
```

### Automated Version Bumping

```yaml
# .github/workflows/version-bump.yml
name: Version Bump

on:
  push:
    branches: [main]

jobs:
  version-bump:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        token: ${{ secrets.GITHUB_TOKEN }}

    - name: Python Semantic Release
      uses: relekang/python-semantic-release@master
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        repository_username: __token__
        repository_password: ${{ secrets.PYPI_API_TOKEN }}
```

## Security and Vulnerability Management

### Security Scanning Workflow

```bash
# Install security tools
pip install safety bandit semgrep pip-audit

# Check for known vulnerabilities
safety check
safety check --json  # JSON output

# Audit all dependencies
pip-audit
pip-audit --format=json

# Static security analysis
bandit -r src/
bandit -r src/ -f json -o security-report.json

# Advanced security scanning with Semgrep
semgrep --config=auto src/

# Check for secrets in code
pip install detect-secrets
detect-secrets scan --all-files --force-use-all-plugins

# GitHub Actions integration
# .github/workflows/security.yml
```

### Automated Security in CI/CD

```yaml
# .github/workflows/security.yml
name: Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'

    - name: Install security tools
      run: |
        pip install safety bandit pip-audit

    - name: Run safety check
      run: safety check

    - name: Run bandit security scan
      run: bandit -r src/ -f json -o bandit-report.json

    - name: Run pip-audit
      run: pip-audit

    - name: Upload security reports
      uses: actions/upload-artifact@v3
      with:
        name: security-reports
        path: "*-report.json"
```

### Dependency Vulnerability Monitoring

```toml
# pyproject.toml - Security configuration
[tool.bandit]
exclude_dirs = ["tests", "build", "dist"]
skips = ["B101", "B601"]  # Skip assert_used, shell usage

[tool.safety]
# Ignore specific vulnerabilities if needed
ignore = ["12345", "67890"]
```

## Package Testing and Quality

### Testing Package Installation

```python
# tests/test_installation.py
import subprocess
import sys
import tempfile
import os

def test_package_installation():
    """Test that package can be installed and imported."""
    with tempfile.TemporaryDirectory() as tmpdir:
        # Create virtual environment
        venv_path = os.path.join(tmpdir, 'venv')
        subprocess.check_call([sys.executable, '-m', 'venv', venv_path])

        # Get paths
        if sys.platform == 'win32':
            pip_path = os.path.join(venv_path, 'Scripts', 'pip')
            python_path = os.path.join(venv_path, 'Scripts', 'python')
        else:
            pip_path = os.path.join(venv_path, 'bin', 'pip')
            python_path = os.path.join(venv_path, 'bin', 'python')

        # Install package
        subprocess.check_call([pip_path, 'install', '.'])

        # Test import
        result = subprocess.run([
            python_path, '-c', 'import my_package; print(my_package.__version__)'
        ], capture_output=True, text=True)

        assert result.returncode == 0
        assert result.stdout.strip() == '0.1.0'

def test_console_scripts():
    """Test that console scripts work."""
    result = subprocess.run(['my-cli', '--help'], capture_output=True, text=True)
    assert result.returncode == 0
    assert 'Usage:' in result.stdout
```

### Package Security

```bash
# Check for known vulnerabilities
pip install safety
safety check

# Audit dependencies
pip-audit

# Check package for common issues
pip install check-manifest
check-manifest

# Validate package description
pip install readme_renderer
python setup.py check --restructuredtext --strict
```

## Distribution Strategies

### Multiple Distribution Channels

```bash
# PyPI (public)
twine upload dist/*

# Private PyPI server
twine upload --repository-url https://private-pypi.company.com dist/*

# Conda packaging
conda build recipe/
anaconda upload /path/to/package.tar.bz2

# Docker image
docker build -t my-package:latest .
docker push registry.company.com/my-package:latest
```

### Platform-Specific Packages

```toml
# pyproject.toml - Platform-specific dependencies
[project]
dependencies = [
    "requests>=2.25.0",
    "pywin32>=227; sys_platform == 'win32'",
    "fcntl; sys_platform != 'win32'",
]

# Platform-specific optional dependencies
[project.optional-dependencies]
windows = [
    "pywin32>=227",
    "wmi>=1.5.1",
]
unix = [
    "python-daemon>=2.3.0",
]
```

## Performance and Optimization

### Package Size Optimization

```bash
# Analyze package size
pip install wheeldiff
wheeldiff my-package-0.1.0-py3-none-any.whl my-package-0.2.0-py3-none-any.whl

# Minimize package size
echo "*.pyc" >> .gitignore
echo "__pycache__/" >> .gitignore
echo "*.so" >> .gitignore
echo ".DS_Store" >> .gitignore

# Use data compression
gzip -9 data/large_file.json
```

### Import Performance

```python
# Lazy imports for better startup time
def expensive_function():
    import heavy_module  # Import only when needed
    return heavy_module.process()

# Use __all__ to control imports
__all__ = ["main_function", "MainClass"]

# Conditional imports
try:
    import optimized_module as processor
except ImportError:
    import fallback_module as processor
```

## Troubleshooting Common Issues

### Build Problems

```bash
# Clear build artifacts
rm -rf build/ dist/ *.egg-info/
pip cache purge

# Fix missing build dependencies
pip install --upgrade setuptools wheel

# Debug build process
python -m build --verbose

# Check for missing files
check-manifest

# Validate package
twine check dist/*
```

### Import Issues

```python
# Debug import problems
import sys
print(sys.path)
print(sys.modules.keys())

# Check package installation
pip show my-package
pip list | grep my-package

# Verify entry points
python -c "import pkg_resources; print([ep for ep in pkg_resources.iter_entry_points('console_scripts')])"
```

## Best Practices Summary

### ✅ Modern Python Packaging Checklist

**Project Structure:**
- [ ] Use `src/` layout for package isolation
- [ ] Include comprehensive `pyproject.toml`
- [ ] Add proper `__init__.py` with `__all__`
- [ ] Include `MANIFEST.in` for data files
- [ ] Set up `pre-commit` hooks

**Dependencies:**
- [ ] Pin dependency versions appropriately
- [ ] Use lock files for reproducible builds
- [ ] Separate dev/test/doc dependencies
- [ ] Regular security audits with `safety`/`pip-audit`

**Testing:**
- [ ] Achieve 90%+ test coverage
- [ ] Test package installation
- [ ] Test console scripts and entry points
- [ ] Include integration tests
- [ ] Use property-based testing with Hypothesis

**Documentation:**
- [ ] Write comprehensive README.md
- [ ] Include API documentation
- [ ] Provide usage examples
- [ ] Document breaking changes
- [ ] Keep CHANGELOG.md updated

**Quality:**
- [ ] Configure `ruff` for linting and formatting
- [ ] Use `mypy` for type checking
- [ ] Set up automated CI/CD
- [ ] Security scanning in CI
- [ ] Automated dependency updates

**Distribution:**
- [ ] Test on TestPyPI first
- [ ] Use semantic versioning
- [ ] Automated releases with GitHub Actions
- [ ] Sign releases when possible
- [ ] Monitor download statistics

### Common Anti-Patterns to Avoid

❌ **Don't:**
- Use `setup.py` for new projects (use `pyproject.toml`)
- Import from `__init__.py` in entry points
- Include test files in distributed package
- Use `pip install .` in CI (use `pip install -e .`)
- Ignore dependency pinning for libraries
- Skip testing package installation
- Commit virtual environments to git
- Use relative imports in console scripts

✅ **Do:**
- Use modern build tools (`build`, `uv`)
- Implement proper error handling
- Use type hints and validate with `mypy`
- Keep dependencies minimal and up-to-date
- Test across multiple Python versions
- Use semantic versioning consistently
- Document breaking changes clearly
- Automate releases and testing

This comprehensive packaging guide covers all aspects of modern Python packaging from basic project structure to advanced distribution strategies, security practices, and performance optimization, enabling you to create professional, maintainable Python packages that follow current best practices.