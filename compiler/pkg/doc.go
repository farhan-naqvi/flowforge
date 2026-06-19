package compiler

// Package compiler provides IR-to-executor compilation
//
// The compiler transforms FlowForge Intermediate Representation (IR)
// into executor-specific artifacts:
// - Argo Workflows YAML
// - Apache Airflow Python DAGs
//
// Compilation Pipeline:
// 1. Parse - Load and deserialize IR
// 2. Validate - Check semantic correctness (cycles, edges, etc.)
// 3. Optimize - Detect parallelization, plan resources
// 4. Compile - Generate executor-specific artifact
// 5. Validate Output - Verify artifact format correctness
//
// Usage:
//
//   compiler := pkg.New()
//   opts := pkg.CompileOptions{
//       Format: pkg.ExecutorFormatArgo,
//       Namespace: "default",
//   }
//   result, err := compiler.Compile(ctx, spec, opts)
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(result.Artifact)
