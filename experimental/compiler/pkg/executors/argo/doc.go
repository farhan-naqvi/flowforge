package argo

// Package argo provides Argo Workflows compilation
//
// This package compiles FlowForge IR to Argo Workflow YAML format.
//
// Example:
//
//   compiler := argo.New("default")
//   result, err := compiler.Compile(ctx, spec)
//   if err != nil {
//       log.Fatal(err)
//   }
//   yaml := result.Artifact.(string)
//   fmt.Println(yaml)
