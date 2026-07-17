"""FlowForge CLI - Command-line interface for pipelines."""

from __future__ import annotations
import sys
import json
from pathlib import Path
from typing import Optional


class CLI:
    """Command-line interface for FlowForge."""

    @staticmethod
    def load_pipeline_module(filepath: str):
        """Load a Python module containing a pipeline.
        
        Args:
            filepath: Path to Python file
        
        Returns:
            Module object
        """
        import importlib.util
        spec = importlib.util.spec_from_file_location("pipeline_module", filepath)
        module = importlib.util.module_from_spec(spec)
        sys.modules["pipeline_module"] = module
        spec.loader.exec_module(module)
        return module

    @staticmethod
    def find_pipeline(module) -> Optional[Any]:
        """Find a Pipeline object in a module.
        
        Args:
            module: Module to search
        
        Returns:
            Pipeline object, or None if not found
        """
        from flowforge.core.pipeline import Pipeline
        
        for name in dir(module):
            obj = getattr(module, name)
            if isinstance(obj, Pipeline):
                return obj
        return None

    @staticmethod
    def compile(filepath: str, executor: str = "argo", output: Optional[str] = None):
        """Compile a pipeline to executor format.
        
        Args:
            filepath: Path to Python pipeline file
            executor: Target executor (argo, airflow, local)
            output: Output file path (optional)
        """
        from flowforge.compiler.ir_exporter import pipeline_to_ir
        
        # Load module
        module = CLI.load_pipeline_module(filepath)
        pipeline_obj = CLI.find_pipeline(module)
        
        if not pipeline_obj:
            print("Error: No pipeline found in file")
            sys.exit(1)
        
        # Export to IR
        spec = pipeline_to_ir(pipeline_obj)
        
        # Output
        ir_json = spec.to_json()
        if output:
            Path(output).write_text(ir_json)
            print(f"Compiled pipeline saved to: {output}")
        else:
            print(ir_json)

    @staticmethod
    def validate(filepath: str):
        """Validate a pipeline.
        
        Args:
            filepath: Path to Python pipeline file
        """
        # Load module
        module = CLI.load_pipeline_module(filepath)
        pipeline_obj = CLI.find_pipeline(module)
        
        if not pipeline_obj:
            print("Error: No pipeline found in file")
            sys.exit(1)
        
        # Validate
        errors = pipeline_obj.validate()
        if errors:
            print("Validation errors:")
            for error in errors:
                print(f"  - {error}")
            sys.exit(1)
        else:
            print("✓ Pipeline is valid")

    @staticmethod
    def run_local(filepath: str):
        """Run pipeline locally.
        
        Args:
            filepath: Path to Python pipeline file
        """
        from flowforge.executor.local import LocalExecutor
        
        # Load module
        module = CLI.load_pipeline_module(filepath)
        pipeline_obj = CLI.find_pipeline(module)
        
        if not pipeline_obj:
            print("Error: No pipeline found in file")
            sys.exit(1)
        
        # Execute
        executor = LocalExecutor()
        try:
            results = executor.execute(pipeline_obj)
            print("✓ Pipeline executed successfully")
            print(f"Results: {json.dumps(results, indent=2, default=str)}")
        except Exception as e:
            print(f"Error: {e}")
            sys.exit(1)

    @staticmethod
    def inspect(filepath: str):
        """Inspect a pipeline.
        
        Args:
            filepath: Path to Python pipeline file
        """
        from flowforge.graph.visualizer import DAGVisualizer
        
        # Load module
        module = CLI.load_pipeline_module(filepath)
        pipeline_obj = CLI.find_pipeline(module)
        
        if not pipeline_obj:
            print("Error: No pipeline found in file")
            sys.exit(1)
        
        # Visualize
        print(DAGVisualizer.visualize(pipeline_obj))


def main():
    """Main CLI entry point."""
    import argparse
    
    parser = argparse.ArgumentParser(
        prog='flowforge',
        description='FlowForge - Declarative data pipeline orchestration',
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Compile command
    compile_parser = subparsers.add_parser('compile', help='Compile pipeline to IR')
    compile_parser.add_argument('file', help='Python pipeline file')
    compile_parser.add_argument('--executor', default='argo', help='Target executor')
    compile_parser.add_argument('-o', '--output', help='Output file')
    
    # Validate command
    validate_parser = subparsers.add_parser('validate', help='Validate pipeline')
    validate_parser.add_argument('file', help='Python pipeline file')
    
    # Run-local command
    run_parser = subparsers.add_parser('run-local', help='Run pipeline locally')
    run_parser.add_argument('file', help='Python pipeline file')
    
    # Inspect command
    inspect_parser = subparsers.add_parser('inspect', help='Inspect pipeline')
    inspect_parser.add_argument('file', help='Python pipeline file')
    
    # Parse arguments
    args = parser.parse_args()
    
    if args.command == 'compile':
        CLI.compile(args.file, args.executor, args.output)
    elif args.command == 'validate':
        CLI.validate(args.file)
    elif args.command == 'run-local':
        CLI.run_local(args.file)
    elif args.command == 'inspect':
        CLI.inspect(args.file)
    else:
        parser.print_help()


if __name__ == '__main__':
    main()
