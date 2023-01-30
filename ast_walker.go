package main

import(
  "github.com/hashicorp/hcl/v2"
  "github.com/hashicorp/hcl/v2/hclsyntax"
//  "fmt"
)

type Mode int

const (
  ModeNone = iota
  ModeTopLevel

  ModeObj
  ModeObjVal

  ModeTupleVal

  ModeObjKey
)

type SNode struct {
  node hclsyntax.Node
  traversal hcl.Traversal
  parent *SNode
  children []*SNode
  rng hcl.Range
}

func (s *SNode) CreateChild(n hclsyntax.Node, t hcl.Traversal) *SNode{
  c := &SNode {
    node: n,
    traversal: t,
    parent: s,
    children: []*SNode{},
    rng: n.Range(),
  }

  s.children = append(s.children, c)
  c.parent = s

  return c
}

type AstWalker struct {
  parentStack []*SNode
  modeStack []Mode
  tupleIndexStack []int
}

// todo: nil and empty cases etc.

func (w *AstWalker) parentPush(n *SNode) {
  w.parentStack = append(w.parentStack, n)
}

func (w *AstWalker) parentPeek() *SNode {
  return w.parentStack[len(w.parentStack) - 1]
}

func (w *AstWalker) parentPop() *SNode {
  top := w.parentPeek()
  w.parentStack = w.parentStack[:len(w.parentStack) - 1]

  return top
}

func (w *AstWalker) modePush(n Mode) {
  w.modeStack = append(w.modeStack, n)
}

func (w *AstWalker) modePeek() Mode {
  return w.modeStack[len(w.modeStack) - 1]
}

func (w *AstWalker) modePop() Mode {
  top := w.modePeek()
  w.modeStack = w.modeStack[:len(w.modeStack) - 1]

  return top
}

func (w *AstWalker) indexPush(i int) {
  w.tupleIndexStack = append(w.tupleIndexStack, i)
}

func (w *AstWalker) indexPeek() int {
  return w.tupleIndexStack[len(w.tupleIndexStack) - 1]
}

func (w *AstWalker) indexPop() int {
  top := w.indexPeek()
  w.tupleIndexStack = w.tupleIndexStack[:len(w.tupleIndexStack) - 1]

  return top
}

func (w *AstWalker) indexIncrement() {
  w.tupleIndexStack[len(w.tupleIndexStack) - 1]++
}

func (w *AstWalker) createChildAndPush(node hclsyntax.Node, tuple bool) *SNode {
  var idx int
  parent := w.parentPeek()

  if tuple {
    idx = w.indexPeek()
  } else {
    idx = -2
  }

  t := BuildTraversalForNode(parent.traversal, node, tuple, idx)
  sn := parent.CreateChild(node, t)
  w.parentPush(sn)

  return sn
}


func (w *AstWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
  parent := w.parentPeek()
  mode := w.modePeek()

  //fmt.Printf("ENTER\nnode: %T\nmode: %v\nparentStack: %v\nparent traversal: %v\nmode stack: %v\n\n\n", node, mode, FormatParentStack(w.parentStack), FormatTraversal(parent.traversal), w.modeStack)

  switch mode{
  case ModeNone:
    w.modePush(ModeNone)

  case ModeTopLevel:
    switch n := node.(type) {
    case *hclsyntax.Body, hclsyntax.Blocks, hclsyntax.Attributes:
      w.modePush(ModeTopLevel)
    case *hclsyntax.Block:
      sn := w.createChildAndPush(node, false)
      sn.rng = n.Body.Range()
      w.modePush(ModeTopLevel)
    case *hclsyntax.Attribute:
      sn := w.createChildAndPush(node, false)
      sn.rng = n.Expr.Range()
      w.modePush(ModeTopLevel)
    case *hclsyntax.ObjectConsExpr:
      w.modePush(ModeObj)
    case *hclsyntax.TupleConsExpr:
      w.modePush(ModeTupleVal)
      w.indexPush(0)
    default:
      w.modePush(ModeNone)
    }

  case ModeObj:
    switch node.(type) {
    case *hclsyntax.ObjectConsKeyExpr:
      w.createChildAndPush(node, false) // for creating the traversal
      w.modePush(ModeObjKey) // ignore the subtree of the key
    default:
      panic("unexpected case")
    }

  case ModeObjVal:
    switch node.(type) {
    case *hclsyntax.TupleConsExpr:
      parent.node = node // populate node reference with the value expression node
      parent.rng = node.Range()
      w.modePush(ModeTupleVal)
      w.indexPush(0)
    case *hclsyntax.ObjectConsExpr:
      // w.createChildAndPush(node, false)
      parent.node = node
      parent.rng = node.Range()
      w.modePush(ModeObj)
    case *hclsyntax.ObjectConsKeyExpr:
      w.modePop()

      w.parentPop()

      w.createChildAndPush(node, false)
      w.modePush(ModeObjKey)
    default:
      parent.node = node // populate node reference with the value expression node
      parent.rng = node.Range()
      // w.createChildAndPush(node, false)
      w.modePush(ModeNone)
    }

  case ModeTupleVal:
    switch node.(type) {
    case *hclsyntax.TupleConsExpr:
      w.createChildAndPush(node, true)

      w.modePush(ModeTupleVal)
      w.indexIncrement()
      w.indexPush(0)
    case *hclsyntax.ObjectConsExpr:
      w.createChildAndPush(node, true)

      w.indexIncrement()
      w.modePush(ModeObj)
    default:
      w.createChildAndPush(node, true)

      w.indexIncrement()
      w.modePush(ModeNone)
    }

  case ModeObjKey:
    w.modePush(ModeObjKey)
  }

  return nil
}

func (w *AstWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
  mode := w.modePeek()
  //parent := w.parentPeek()

  //fmt.Printf("EXIT\nnode: %T\nmode: %v\nparentStack: %v\nparent traversal: %v\nmode stack: %v\n\n\n", node, mode, FormatParentStack(w.parentStack), FormatTraversal(parent.traversal), w.modeStack)

  switch mode {
  case ModeObjKey:
    switch node.(type) {
      case *hclsyntax.ObjectConsKeyExpr:
        w.modePop()
        w.modePush(ModeObjVal)
      default:
        w.modePop()
    }

  case ModeNone:
    w.modePop()
    if w.modePeek() == ModeTupleVal {
      w.parentPop()
    }

  case ModeTopLevel:
    switch node.(type) {
    case *hclsyntax.Block, *hclsyntax.Attribute:
      w.parentPop()
      w.modePop()
    case hclsyntax.Blocks, hclsyntax.Attributes, *hclsyntax.Body:
      w.modePop()
    default:
      panic("unexpected case")
    }

  case ModeObjVal:
    switch node.(type) {
    case *hclsyntax.ObjectConsExpr:
      w.modePop() // pop out modeobjval
      w.parentPop() // pop out val node
      if w.modePeek() == ModeObj {
        w.modePop() // pop out modeobj
      }

      if w.modePeek() == ModeTupleVal {
        w.parentPop() // this item is a tuple element, remove
      }
      /*
      if w.modePeek() == ModeTupleVal {
        for { // keep popping parents till we reach the object cons
          if _, ok := w.parentPeek().node.(*hclsyntax.ObjectConsExpr); ok {
            break
          }
          w.parentPop()
        }
      }
      */
    case *hclsyntax.Body, *hclsyntax.Block, *hclsyntax.Attribute, hclsyntax.Blocks, hclsyntax.Attributes, *hclsyntax.ObjectConsKeyExpr:
      panic("unexpected case")
    case *hclsyntax.TupleConsExpr:
      w.parentPop()
      w.modePop()
    default:
      w.modePop()
    }

  case ModeTupleVal:
    switch node.(type) {
    case *hclsyntax.TupleConsExpr:
      w.modePop()
      w.indexPop()
      // w.parentPop() // why would I parent pop here?

      /*
      if w.modePeek() != ModeObjVal {
        w.parentPop()
      }
      */
    default:
      w.parentPop()
      // w.indexPop()
      w.modePop()
    }

  default: // ModeObj only so far
    switch node.(type) {
    case *hclsyntax.Body, *hclsyntax.Block, *hclsyntax.Attribute, hclsyntax.Blocks, hclsyntax.Attributes, *hclsyntax.ObjectConsKeyExpr:
      panic("unexpected case")
    case *hclsyntax.ObjectConsExpr:
      w.modePop()
      // w.parentPop()
      
      if w.modePeek() == ModeTupleVal {
        w.parentPop()
        w.modePop()
      }
    default:
      w.modePop()
    }
  }

  return nil
}
