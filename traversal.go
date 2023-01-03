package main

import(
  "github.com/zclconf/go-cty/cty"
  "github.com/zclconf/go-cty/cty/gocty"
  "github.com/hashicorp/hcl/v2"
  "github.com/hashicorp/hcl/v2/hclsyntax"
)

func BuildTraversalForNode(t hcl.Traversal, node hclsyntax.Node, tuple bool, idx int) hcl.Traversal {
  switch n := node.(type) {
  case *hclsyntax.ObjectConsKeyExpr:
    if tuple {
      return BuildTraversalForTupleElement(t, n, idx)
    } else {
      return BuildTraversalForObjectConsKey(t, n)
    }
  case *hclsyntax.Attribute:
    return BuildTraversalForAttribute(t, n)
  case *hclsyntax.Block:
    return BuildTraversalForBlock(t, n)
  default:
    if tuple {
      return BuildTraversalForTupleElement(t, n, idx)
    } else {
      panic("unexpected type to build traversal")
    }
  }
}

func BuildTraversalForObjectConsKey(t hcl.Traversal, n *hclsyntax.ObjectConsKeyExpr) hcl.Traversal {
  res := make(hcl.Traversal, 0)
  switch k := n.Wrapped.(type) {
  case *hclsyntax.ScopeTraversalExpr:
    res = append(res, t...)
    res = append(res, k.Traversal...)
    return res
  case *hclsyntax.TemplateExpr:
    if !k.IsStringLiteral() {
      panic("not handled")
    }
    val, _ := k.Value(nil)
    str := val.AsString()
    tnew := hcl.TraverseAttr {
      Name: str,
      SrcRange: k.Range(),
    }
    res = append(res, t...)
    res = append(res, tnew)
    return res
  default:
    panic("not handled")
  }
}

func BuildTraversalForTupleElement(t hcl.Traversal, n hclsyntax.Node, index int) hcl.Traversal {
  res := make(hcl.Traversal, 0)
  idx := index
  k, err := gocty.ToCtyValue(idx, cty.Number)

  if err != nil {
    panic(err)
  }

  tnew := hcl.TraverseIndex {
    Key: k,
    SrcRange: n.Range(),
  }

  res = append(res, t...)
  res = append(res, tnew)
  return res
}

func BuildTraversalForAttribute(t hcl.Traversal, n *hclsyntax.Attribute) hcl.Traversal {
  res := make(hcl.Traversal, 0)
  tnew := hcl.TraverseAttr {
    Name: n.Name,
    SrcRange: n.Range(),
  }

  res = append(res, t...)
  res = append(res, tnew)
  return res
}

func BuildTraversalForBlock(t hcl.Traversal, n *hclsyntax.Block) hcl.Traversal {
  res := make(hcl.Traversal, 0)
  tnew := make(hcl.Traversal, 0)

  for _, v := range append([]string{n.Type}, n.Labels...) {
    tnew = append(
      tnew,
      hcl.TraverseAttr {
        Name: v,
        SrcRange: n.Range(),
      },
    )
  }

  res = append(res, t...)
  res = append(res, tnew...)
  return res
}

func EqualTraversals(t1 hcl.Traversal, t2 hcl.Traversal) bool {
  if len(t1) == 0 && len(t2) == 0 {
    return true
  }

  if len(t1) != len(t2) {
    return false
  }

  switch a := t1[0].(type) {
  case hcl.TraverseRoot:
    switch b := t2[0].(type) {
    case hcl.TraverseRoot:
      if a.Name != b.Name {
        return false
      }
      return EqualTraversals(t1[1:], t2[1:])
    
    case hcl.TraverseAttr:
      if a.Name != b.Name {
        return false
      }
      return EqualTraversals(t1[1:], t2[1:])
    }
  case hcl.TraverseAttr:
    switch b := t2[0].(type) {
    case hcl.TraverseRoot:
      if a.Name != b.Name {
        return false
      }
      return EqualTraversals(t1[1:], t2[1:])
    case hcl.TraverseAttr:
      if a.Name != b.Name {
        return false
      }
      return EqualTraversals(t1[1:], t2[1:])
    default:
      return false
    }
  case hcl.TraverseIndex:
    switch b := t2[0].(type) {
    case hcl.TraverseIndex:
      if a.Key.Equals(b.Key) == cty.False {
        return false
      }
      return EqualTraversals(t1[1:], t2[1:])
    default:
      return false
    }
  }

  return false
}
