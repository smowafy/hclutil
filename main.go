package main

import(
  "fmt"
  "os"
  "io/ioutil"
//  "reflect"
//  "github.com/zclconf/go-cty/cty"
  "github.com/hashicorp/hcl/v2"
  "github.com/hashicorp/hcl/v2/hclsyntax"
)

func Find(root *SNode, inputTraversal hcl.Traversal, inpFile []byte) {
  res := Query(root, inputTraversal)

  if res != nil {
    r := res.rng
    s := r.Start.Byte
    e := r.End.Byte
    fmt.Printf("%v\n", string(inpFile[s:e]))
  }
}

func Replace(root *SNode, inputTraversal hcl.Traversal, inpFile []byte, val []byte) {
  res := Query(root, inputTraversal)

  if res != nil {
    r := res.rng
    s := r.Start.Byte
    e := r.End.Byte

    newFile := make([]byte, 0)

    newFile = append(newFile, inpFile[:s]...)
    newFile = append(newFile, val...)
    newFile = append(newFile, inpFile[e:]...)

    fmt.Printf("%v\n", string(newFile))
  }
}

func main() {

  command := os.Args[1]
  filePath := os.Args[2]
  traversalInput := os.Args[3]

  inp, diags := hclsyntax.ParseTraversalAbs([]byte(traversalInput), "input_traversal.hcl", hcl.Pos{ Line: 1, Column: 1 })

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

  switch command {
  case "find":
    Find(rootNode, inp, srcBytes)
  case "replace":
    replaceVal, err := ioutil.ReadAll(os.Stdin)
    
    if err != nil {
      panic(err)
    }

    Replace(rootNode, inp, srcBytes, replaceVal)
  }
}
