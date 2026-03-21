# Python Local Build Toolchain Guide

A comprehensive guide to setting up and using the best local development tools for Python projects.

## Python Version Management

### pyenv - Python Version Manager

```bash
# Install pyenv (macOS)
brew install pyenv

# Install pyenv (Linux)
curl https://pyenv.run | bash

# Add to shell configuration
echo 'export PATH="$HOME/.pyenv/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init --path)"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc

# List available Python versions
pyenv install --list

# Install specific Python versions
pyenv install 3.11.7
pyenv install 3.12.1

# Set global Python version
pyenv global 3.11.7

# Set local Python version for project
pyenv local 3.12.1

# List installed versions
pyenv versions

# Create .python-version file
echo "3.12.1" > .python-version
```

### python-launcher (Windows)

```bash
# Install specific version
py -3.11 -m pip install package

# List installed versions
py -0

# Use specific version
py -3.12 script.py
```

## Virtual Environment Management

### Built-in venv

```bash
# Create virtual environment
python -m venv venv

# Activate (Unix/macOS)
source venv/bin/activate

# Activate (Windows)
venv\Scripts\activate

# Deactivate
deactivate

# Remove virtual environment
rm -rf venv
```

### virtualenvwrapper

```bash
# Install
pip install virtualenvwrapper

# Add to shell configuration
export WORKON_HOME=$HOME/.virtualenvs
source /usr/local/bin/virtualenvwrapper.sh

# Create environment
mkvirtualenv myproject

# Work on environment
workon myproject

# Deactivate
deactivate

# Remove environment
rmvirtualenv myproject

# List environments
lsvirtualenv
```

### pipenv - High-level Dependency Management

```bash
# Install pipenv
pip install pipenv

# Initialize project with Pipfile
pipenv --python 3.11

# Install dependencies
pipenv install requests
pipenv install pytest --dev

# Install from Pipfile
pipenv install

# Activate shell
pipenv shell

# Run commands in environment
pipenv run python script.py
pipenv run pytest

# Generate requirements.txt
pipenv requirements > requirements.txt
pipenv requirements --dev > requirements-dev.txt

# Check for security vulnerabilities
pipenv check

# Update dependencies
pipenv update

# Remove unused dependencies
pipenv clean
```

### Poetry - Modern Dependency Management

```bash
# Install poetry
curl -sSL https://install.python-poetry.org | python3 -

# Initialize new project
poetry new my-project
cd my-project

# Initialize in existing project
poetry init

# Install dependencies
poetry add requests
poetry add pytest --group dev

# Install project dependencies
poetry install

# Run commands in environment
poetry run python script.py
poetry run pytest

# Activate shell
poetry shell

# Build package
poetry build

# Publish package
poetry publish

# Update dependencies
poetry update

# Remove dependency
poetry remove requests

# Export requirements
poetry export -f requirements.txt --output requirements.txt
poetry export --dev -f requirements.txt --output requirements-dev.txt
```

### Example pyproject.toml (Poetry)

```toml
[tool.poetry]
name = "my-project"
version = "0.1.0"
description = "A sample Python project"
authors = ["Your Name <you@example.com>"]
readme = "README.md"
packages = [{include = "my_project"}]

[tool.poetry.dependencies]
python = "^3.9"
requests = "^2.28.0"
click = "^8.1.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.2.0"
black = "^23.1.0"
isort = "^5.12.0"
flake8 = "^6.0.0"
mypy = "^1.0.0"
pre-commit = "^3.0.0"

[tool.poetry.scripts]
my-cli = "my_project.cli:main"

[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"

[tool.black]
line-length = 88
target-version = ['py39']

[tool.isort]
profile = "black"
line_length = 88

[tool.mypy]
python_version = "3.9"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

## Code Formatting and Linting

### Black - Code Formatter

```bash
# Install
pip install black

# Format single file
black script.py

# Format directory
black src/

# Check what would be formatted
black --check src/

# Show diff
black --diff src/

# Configuration in pyproject.toml
[tool.black]
line-length = 88
target-version = ['py39']
include = '\.pyi?$'
exclude = '''
/(
    \.eggs
  | \.git
  | \.mypy_cache
  | \.tox
  | \.venv
  | build
  | dist
)/
'''
```

### isort - Import Sorting

```bash
# Install
pip install isort

# Sort imports in file
isort script.py

# Sort imports in directory
isort src/

# Check what would be changed
isort --check-only src/

# Show diff
isort --diff src/

# Configuration in pyproject.toml
[tool.isort]
profile = "black"
line_length = 88
known_first_party = ["my_project"]
skip = ["migrations"]
```

### Flake8 - Style Guide Enforcement

```bash
# Install
pip install flake8

# Check file
flake8 script.py

# Check directory
flake8 src/

# Configuration in .flake8 or setup.cfg
[flake8]
max-line-length = 88
extend-ignore = E203, W503
exclude =
    .git,
    __pycache__,
    .venv,
    migrations
```

### pylint - Comprehensive Linter

```bash
# Install
pip install pylint

# Check file
pylint script.py

# Check package
pylint my_package/

# Generate config file
pylint --generate-rcfile > .pylintrc

# Configuration in .pylintrc
[MASTER]
extension-pkg-allow-list=pydantic

[MESSAGES CONTROL]
disable=C0114,C0115,C0116  # Missing docstrings

[FORMAT]
max-line-length=88
```

### ruff - Fast Python Linter

```bash
# Install
pip install ruff

# Check files
ruff check .

# Fix auto-fixable issues
ruff --fix .

# Configuration in pyproject.toml
[tool.ruff]
line-length = 88
target-version = "py39"
select = ["E", "F", "I", "N", "W"]
ignore = ["E203"]
exclude = [
    ".git",
    "__pycache__",
    ".venv",
]

[tool.ruff.isort]
known-first-party = ["my_project"]
```

## Type Checking

### mypy - Static Type Checker

```bash
# Install
pip install mypy

# Check file
mypy script.py

# Check package
mypy src/

# Install type stubs
mypy --install-types

# Configuration in mypy.ini or pyproject.toml
[tool.mypy]
python_version = "3.9"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
disallow_incomplete_defs = true
check_untyped_defs = true
disallow_untyped_decorators = true
no_implicit_optional = true
warn_redundant_casts = true
warn_unused_ignores = true

[[tool.mypy.overrides]]
module = ["requests.*", "pytest.*"]
ignore_missing_imports = true
```

### pyright/pylance - Microsoft Type Checker

```bash
# Install (Node.js required)
npm install -g pyright

# Check project
pyright

# Configuration in pyrightconfig.json
{
    "include": ["src"],
    "exclude": ["**/node_modules", "**/__pycache__"],
    "pythonVersion": "3.9",
    "typeCheckingMode": "basic",
    "useLibraryCodeForTypes": true
}
```

## Testing Framework

### pytest - Testing Framework

```bash
# Install
pip install pytest pytest-cov pytest-mock

# Run tests
pytest

# Run with coverage
pytest --cov=src --cov-report=html

# Run specific test
pytest tests/test_example.py::test_function

# Run with markers
pytest -m slow

# Configuration in pyproject.toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
addopts = [
    "--strict-markers",
    "--strict-config",
    "--cov=src",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-fail-under=80",
]
markers = [
    "slow: marks tests as slow",
    "integration: marks tests as integration tests",
]
```

### tox - Testing Multiple Environments

```bash
# Install
pip install tox

# Run tests in all environments
tox

# Run specific environment
tox -e py39

# Recreate environments
tox -r

# Configuration in tox.ini
[tox]
envlist = py39,py310,py311,flake8,mypy

[testenv]
deps =
    pytest
    pytest-cov
commands = pytest {posargs}

[testenv:flake8]
deps = flake8
commands = flake8 src tests

[testenv:mypy]
deps = mypy
commands = mypy src

[testenv:coverage]
deps =
    pytest
    pytest-cov
commands = pytest --cov=src --cov-report=html
```

## Pre-commit Hooks

### Setting up pre-commit

```bash
# Install
pip install pre-commit

# Create .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-merge-conflict
      - id: debug-statements

  - repo: https://github.com/psf/black
    rev: 23.1.0
    hooks:
      - id: black
        language_version: python3

  - repo: https://github.com/pycqa/isort
    rev: 5.12.0
    hooks:
      - id: isort
        args: ["--profile", "black"]

  - repo: https://github.com/pycqa/flake8
    rev: 6.0.0
    hooks:
      - id: flake8

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.0.0
    hooks:
      - id: mypy

# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files

# Update hooks
pre-commit autoupdate
```

## Documentation Tools

### Sphinx - Documentation Generator

```bash
# Install
pip install sphinx sphinx-rtd-theme

# Quick start
sphinx-quickstart docs

# Build documentation
cd docs
make html

# Auto-generate API docs
sphinx-apidoc -o source/ ../src/

# Configuration in conf.py
extensions = [
    'sphinx.ext.autodoc',
    'sphinx.ext.viewcode',
    'sphinx.ext.napoleon',
    'sphinx.ext.intersphinx',
]

html_theme = 'sphinx_rtd_theme'
autodoc_typehints = 'description'
```

### MkDocs - Markdown Documentation

```bash
# Install
pip install mkdocs mkdocs-material

# Create new project
mkdocs new my-project

# Serve locally
mkdocs serve

# Build documentation
mkdocs build

# Deploy to GitHub Pages
mkdocs gh-deploy

# Configuration in mkdocs.yml
site_name: My Project
theme:
  name: material
  features:
    - navigation.tabs
    - navigation.sections

markdown_extensions:
  - codehilite
  - toc:
      permalink: true

nav:
  - Home: index.md
  - User Guide: user-guide.md
  - API Reference: api.md
```

## Development Tools

### IPython and Jupyter

```bash
# Install
pip install ipython jupyter

# Start IPython
ipython

# Start Jupyter
jupyter notebook

# Install extensions
pip install jupyterlab
jupyter lab

# Jupyter configuration
jupyter --config-dir
jupyter notebook --generate-config
```

### Python Debugger (pdb)

```python
# Using pdb
import pdb

def problem_function(x):
    pdb.set_trace()  # Set breakpoint
    result = x * 2
    return result

# Using breakpoint() (Python 3.7+)
def better_function(x):
    breakpoint()  # Modern breakpoint
    result = x * 2
    return result

# IPython debugger
from IPython import embed
embed()  # Drop into IPython shell
```

### Rich - Rich Text and Beautiful Formatting

```bash
# Install
pip install rich

# Example usage
from rich.console import Console
from rich.table import Table
from rich.progress import track

console = Console()
console.print("Hello", style="bold red")

table = Table()
table.add_column("Name")
table.add_column("Age")
table.add_row("Alice", "25")
console.print(table)
```

## Build Tools and Task Runners

### invoke - Task Runner

```bash
# Install
pip install invoke

# Create tasks.py
from invoke import task

@task
def test(c):
    """Run tests"""
    c.run("pytest")

@task
def lint(c):
    """Run linting"""
    c.run("flake8 src tests")
    c.run("black --check src tests")
    c.run("isort --check-only src tests")

@task
def format(c):
    """Format code"""
    c.run("black src tests")
    c.run("isort src tests")

@task(pre=[test, lint])
def ci(c):
    """Run CI tasks"""
    print("All CI tasks completed!")

# Run tasks
inv test
inv lint
inv format
inv ci
```

### make - Traditional Build Tool

```makefile
# Makefile
.PHONY: help test lint format clean install

help:
	@echo "Available commands:"
	@echo "  install    Install dependencies"
	@echo "  test       Run tests"
	@echo "  lint       Run linting"
	@echo "  format     Format code"
	@echo "  clean      Clean up"

install:
	pip install -e .[dev]

test:
	pytest

lint:
	flake8 src tests
	black --check src tests
	isort --check-only src tests
	mypy src

format:
	black src tests
	isort src tests

clean:
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -delete
	rm -rf build/ dist/ *.egg-info/
	rm -rf .pytest_cache/ .mypy_cache/ .coverage htmlcov/

ci: test lint
	@echo "All CI tasks completed!"
```

## Performance Profiling

### cProfile - Built-in Profiler

```bash
# Profile script
python -m cProfile -o profile.stats script.py

# View profile results
python -c "import pstats; pstats.Stats('profile.stats').sort_stats('cumulative').print_stats(10)"
```

### line_profiler - Line-by-Line Profiling

```bash
# Install
pip install line_profiler

# Add @profile decorator to functions
@profile
def slow_function():
    # code here
    pass

# Run profiler
kernprof -l -v script.py
```

### memory_profiler - Memory Usage

```bash
# Install
pip install memory_profiler psutil

# Add @profile decorator
@profile
def memory_intensive_function():
    # code here
    pass

# Run profiler
python -m memory_profiler script.py
```

## Project Template Structure

```
my-project/
├── .github/
│   └── workflows/
│       └── ci.yml
├── docs/
│   ├── conf.py
│   └── index.rst
├── src/
│   └── my_project/
│       ├── __init__.py
│       ├── main.py
│       └── utils.py
├── tests/
│   ├── __init__.py
│   ├── test_main.py
│   └── test_utils.py
├── .gitignore
├── .pre-commit-config.yaml
├── .python-version
├── Makefile
├── README.md
├── pyproject.toml
├── requirements.txt
├── requirements-dev.txt
└── tox.ini
```

### Complete pyproject.toml Example

```toml
[build-system]
requires = ["setuptools>=61.0", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "my-project"
version = "0.1.0"
description = "A sample Python project"
authors = [{name = "Your Name", email = "you@example.com"}]
license = {text = "MIT"}
readme = "README.md"
requires-python = ">=3.9"
classifiers = [
    "Development Status :: 3 - Alpha",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
]
dependencies = [
    "click>=8.0.0",
    "requests>=2.25.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.0.0",
    "pytest-cov>=4.0.0",
    "black>=23.0.0",
    "isort>=5.10.0",
    "flake8>=6.0.0",
    "mypy>=1.0.0",
    "pre-commit>=3.0.0",
]

[project.scripts]
my-cli = "my_project.main:cli"

[project.urls]
Homepage = "https://github.com/username/my-project"
Repository = "https://github.com/username/my-project"
"Bug Tracker" = "https://github.com/username/my-project/issues"

[tool.setuptools.packages.find]
where = ["src"]

[tool.black]
line-length = 88
target-version = ['py39']

[tool.isort]
profile = "black"
line_length = 88
known_first_party = ["my_project"]

[tool.mypy]
python_version = "3.9"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true

[tool.pytest.ini_options]
testpaths = ["tests"]
addopts = [
    "--cov=src",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-fail-under=80",
]

[tool.coverage.run]
source = ["src"]
omit = ["tests/*"]

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "def __repr__",
    "raise AssertionError",
    "raise NotImplementedError",
]
```

This comprehensive local build toolchain setup provides everything needed for professional Python development, from code formatting and linting to testing and documentation generation.