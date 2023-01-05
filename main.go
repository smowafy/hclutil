package main

import(
  "fmt"
  "os"
//  "reflect"
//  "github.com/zclconf/go-cty/cty"
  "github.com/hashicorp/hcl/v2"
  "github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
  filePath := os.Args[1]
  traversalInput := os.Args[2]

  inp, diags := hclsyntax.ParseTraversalAbs([]byte(traversalInput), "dummy.hcl", hcl.Pos{ Line: 1, Column: 1 })

  if diags.HasErrors() {
    panic(diags)
  }

  srcBytes, err := os.ReadFile(filePath)

  if err != nil {
    panic(err)
  }

  srcFile, diags := hclsyntax.ParseConfig(srcBytes, "placeholder.hcl", hcl.Pos{ Line: 1, Column: 1, Byte: 0 })

  if diags.HasErrors() {
    panic(diags)
  }

  srcBody := srcFile.Body.(*hclsyntax.Body)

  initialTraversal := make(hcl.Traversal, 0)
  rootNode := &SNode {traversal: initialTraversal}
  initialStack := []*SNode{rootNode}
  initialModes := []Mode{ModeTopLevel}

  w := AstWalker {
    parentStack: initialStack,
    modeStack: initialModes,
  }

  hclsyntax.Walk(srcBody, &w)

//  WalkSNode(rootNode)

  res := Query(rootNode, inp)

  if res != nil {
    r := res.rng
    s := r.Start.Byte
    e := r.End.Byte
    fmt.Printf("%v\n", string(srcBytes[s:e]))
  }
}
