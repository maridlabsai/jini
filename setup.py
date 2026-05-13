from pathlib import Path

from setuptools import setup


ROOT = Path(__file__).resolve().parent
README = ROOT.joinpath("README.md").read_text(encoding="utf-8")
VERSION = ROOT.joinpath("VERSION").read_text(encoding="utf-8").strip()


setup(
    name="jini-framework",
    version=VERSION,
    description="A framework with a strict protocol core for governed, stateful AI work.",
    long_description=README,
    long_description_content_type="text/markdown",
    author="Sharad Sharma",
    author_email="maridlabsai@gmail.com",
    url="https://github.com/maridlabsai/jini",
    packages=["tools"],
    include_package_data=True,
    install_requires=["PyYAML>=6,<7"],
    python_requires=">=3.9",
    entry_points={
        "console_scripts": [
            "jini=tools.jini_validate:main",
        ]
    },
)
