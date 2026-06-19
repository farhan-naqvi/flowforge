package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"flowforge/compiler/pkg"
	ir "flowforge/ir/pkg"
)

func main() {
	// Define subcommands
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compile":
		compileCmd()
	case "validate":
		validateCmd()
	case "optimize":
		optimizeCmd()
	case "inspect":
		inspectCmd()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FlowForge Compiler")
	fmt.Println("\nUsage:")
	fmt.Println("  flowforge-compiler compile <ir.json> [OPTIONS]")
	fmt.Println("  flowforge-compiler validate <ir.json>")
	fmt.Println("  flowforge-compiler optimize <ir.json>")
	fmt.Println("  flowforge-compiler inspect <ir.json>")
	fmt.Println("\nCompile Options:")
	fmt.Println("  -executor [argo|airflow]   Target executor (default: argo)")
	fmt.Println("  -output <file>             Output file (default: stdout)")
	fmt.Println("  -namespace <ns>            Kubernetes namespace for Argo (default: default)")
}

func compileCmd() {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	executor := fs.String("executor", "argo", "Target executor (argo|airflow)")
	output := fs.String("output", "", "Output file (default: stdout)")
	namespace := fs.String("namespace", "default", "Kubernetes namespace")

	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Println("Error: IR file required")
		os.Exit(1)
	}

	irFile := fs.Arg(0)

	// Load IR
	spec, err := loadIR(irFile)
	if err != nil {
		log.Fatalf("Failed to load IR: %v", err)
	}

	// Compile
	compiler := pkg.New()
	opts := pkg.CompileOptions{
		Format:    pkg.ExecutorFormat(*executor),
		Namespace: *namespace,
	}

	result, err := compiler.Compile(context.Background(), spec, opts)
	if err != nil {
		log.Fatalf("Compilation failed: %v", err)
	}

	// Output
	artifact := result.Artifact.(string)
	if *output != "" {
		if err := ioutil.WriteFile(*output, []byte(artifact), 0644); err != nil {
			log.Fatalf("Failed to write output: %v", err)
		}
		fmt.Printf("Compiled to: %s\n", *output)
	} else {
		fmt.Println(artifact)
	}

	fmt.Printf("\nMetadata:\n")
	for k, v := range result.Metadata {
		fmt.Printf("  %s: %v\n", k, v)
	}
}

func validateCmd() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Println("Error: IR file required")
		os.Exit(1)
	}

	irFile := fs.Arg(0)

	// Load IR
	spec, err := loadIR(irFile)
	if err != nil {
		log.Fatalf("Failed to load IR: %v", err)
	}

	// Validate
	validator := pkg.NewIRValidator()
	result := validator.ValidateSpec(context.Background(), spec)

	if result.Valid {
		fmt.Println("✓ Pipeline is valid")
	} else {
		fmt.Println("✗ Validation errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
		os.Exit(1)
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warn := range result.Warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}
}

func optimizeCmd() {
	fs := flag.NewFlagSet("optimize", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Println("Error: IR file required")
		os.Exit(1)
	}

	irFile := fs.Arg(0)

	// Load IR
	spec, err := loadIR(irFile)
	if err != nil {
		log.Fatalf("Failed to load IR: %v", err)
	}

	// Optimize
	optimizer := pkg.NewOptimizer()
	optimized := optimizer.Optimize(context.Background(), spec)

	fmt.Println(optimizer.Summary())

	// Output optimized IR
	data, err := json.MarshalIndent(optimized, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize optimized IR: %v", err)
	}

	fmt.Println("\nOptimized IR:")
	fmt.Println(string(data))
}

func inspectCmd() {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Println("Error: IR file required")
		os.Exit(1)
	}

	irFile := fs.Arg(0)

	// Load IR
	spec, err := loadIR(irFile)
	if err != nil {
		log.Fatalf("Failed to load IR: %v", err)
	}

	// Print info
	fmt.Printf("Pipeline: %s (v%s)\n", spec.Metadata.Name, spec.Metadata.Version)
	fmt.Printf("Owner: %s\n", spec.Metadata.Owner)
	fmt.Printf("Description: %s\n\n", spec.Metadata.Description)

	fmt.Printf("Tasks: %d\n", len(spec.Tasks))
	for taskID, task := range spec.Tasks {
		fmt.Printf("  - %s (image: %s)\n", taskID, task.Config.Image)
	}

	fmt.Printf("\nEdges: %d\n", len(spec.Edges))
	for _, edge := range spec.Edges {
		fmt.Printf("  - %s.%s → %s.%s\n", edge.From, edge.FromPort, edge.To, edge.ToPort)
	}

	// Validate
	validator := pkg.NewIRValidator()
	valResult := validator.ValidateSpec(context.Background(), spec)
	fmt.Printf("\nValidation: %s\n", map[bool]string{true: "✓ VALID", false: "✗ INVALID"}[valResult.Valid])
	if !valResult.Valid {
		for _, err := range valResult.Errors {
			fmt.Printf("  Error: %s\n", err)
		}
	}
}

func loadIR(filepath string) (*ir.PipelineSpec, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec ir.PipelineSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IR: %w", err)
	}

	return &spec, nil
}
