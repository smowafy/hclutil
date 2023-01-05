package main

import(
   "fmt"
//  "github.com/zclconf/go-cty/cty"
  "github.com/zclconf/go-cty/cty/gocty"
  "github.com/hashicorp/hcl/v2"
//  "github.com/hashicorp/hcl/v2/hclsyntax"
)

func psn(a []*SNode) string {
  s := ""
  for _, n := range a {
    s += fmt.Sprintf(" %T", n.node)
  }
  return s
}


func FormatParentStack(st []*SNode) string {
  s := ""

  for _, v := range st {
    s += fmt.Sprintf(" %T", v.node)
  }

  return s
}

func FormatTraversal(traversal hcl.Traversal) string {
  s := ""
  for _, tr := range traversal {
    switch t := tr.(type){
    case hcl.TraverseRoot:
      s += fmt.Sprintf(" %v", t.Name)
    case hcl.TraverseAttr:
      s += fmt.Sprintf(" %v", t.Name)
    case hcl.TraverseIndex:
      //idxVal, _ := t.Key.AsBigFloat().Int64()
      var idx int
      err := gocty.FromCtyValue(t.Key, &idx)
      if err != nil {
        panic("failed to get index from cty")
      }
      s += fmt.Sprintf(" %v", idx)
    }
  }

  return s
}

// for debugging
func WalkSNode(node *SNode) {
  fmt.Printf("enter traversal: %v\n", FormatTraversal(node.traversal))
  fmt.Printf("type: %T\n", node.node)
  fmt.Printf("children: %v\n", psn(node.children))

  fmt.Printf("\n\n")

  for _, c := range node.children {
    WalkSNode(c)
  }

  fmt.Printf("exit traversal: %v\n", FormatTraversal(node.traversal))
  fmt.Printf("\n\n")
}

func Query(node *SNode, t hcl.Traversal) *SNode {
  if EqualTraversals(node.traversal, t) {
    return node
  }

  var res *SNode

  for _, c := range node.children {
    res = Query(c, t)
    if res != nil {
      return res
    }
  }

  return nil
}
