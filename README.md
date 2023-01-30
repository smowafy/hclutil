# sre-hclutil
A CLI tool for operations on HCL files

### Build
`go build -o hclutil main.go ast_walker.go traversal.go snode_walker.go`

### Command examples:
`example.tf`
```hcl
variable "abc" {
  default = 1
}
```
```console
a@a:~$ hclutil find example.tf 'variable.abc'
                {
   default = 1
}
a@a:~$ hclutil find example.tf 'variable.abc.default'
1
a@a:~$ echo 3 | hclutil replace example.tf 'variable.abc.default'
variable "abc" {
  default = 3
}
```
