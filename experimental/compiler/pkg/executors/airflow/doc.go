package airflow

// Package airflow provides Apache Airflow compilation
//
// This package compiles FlowForge IR to Apache Airflow DAG Python code.
//
// Example:
//
//   compiler := airflow.New("my_dag")
//   result, err := compiler.Compile(ctx, spec)
//   if err != nil {
//       log.Fatal(err)
//   }
//   code := result.Artifact.(string)
//   fmt.Println(code)
