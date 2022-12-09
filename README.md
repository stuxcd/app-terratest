# Terratest Go Modules :grinning:

Go modules used across Terratests.

## Usage

```go
package main

import (
  "fmt"
  "context"

  "github.com/stuxcd/app-terratest/pkg/k8s"
)

func main() {
  eksClusterName := "example"
  awsRegion := "eu-west-2"

  clientSet, err := k8s.NewEKSClientset(eksClusterName, awsRegion)
  if err != nil {
    log.Fatal(err)
  }

  pods, err := clientSet.GetPods(context.TODO(), "")
  if err != nil {
    log.Fatal(err)
  }

  for _, pod := range pods {
    fmt.Println(pod.Name)
  }
}
```
