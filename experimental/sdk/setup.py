"""FlowForge SDK setup configuration."""

from setuptools import setup, find_packages

setup(
    name='flowforge',
    version='0.1.0',
    description='Declarative data pipeline orchestration framework',
    author='FlowForge Contributors',
    python_requires='>=3.11',
    packages=find_packages(exclude=['tests', 'examples']),
    install_requires=[
        'flowforge-ir>=0.1.0',  # IR module (local reference)
    ],
    extras_require={
        'dev': [
            'pytest>=7.0',
            'pytest-cov>=3.0',
            'black>=22.0',
            'pylint>=2.0',
        ],
        'aws': [
            'boto3>=1.20',
        ],
        'spark': [
            'pyspark>=3.0',
        ],
    },
    entry_points={
        'console_scripts': [
            'flowforge=flowforge.cli.cli:main',
        ],
    },
    classifiers=[
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: Apache Software License',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.11',
        'Programming Language :: Python :: 3.12',
    ],
)
