from setuptools import find_packages, setup

setup(
    name='flowforge',
    version='0.1.0',
    description='FlowForge pipeline compiler CLI',
    packages=find_packages(),
    include_package_data=True,
    install_requires=['click>=8.0.0'],
    entry_points={
        'console_scripts': ['flowforge=flowforge.cli:main'],
    },
    python_requires='>=3.8',
)
