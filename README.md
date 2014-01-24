## gofor

Because who can resist a good homonym?

## What is it?

gofor analyzes the types of for loops encountered in Go code.
It is geared towards detecting counting loops -- loops of the form
`for i := min; i < max; i += stride` -- and its categorizations are
mostly steps on the way to ruling out non-counting loops and then detecting
interesting types of counting loops -- special values for min and stride
(0 and 1 respectively) and literal vs non-literal min/max/stride.

It is intended to be used in conjunction with `sort` and `uniq`.

Running this on all the packages that were go gettable from http://godoc.org/-/index
as of Sep 25, 2013, yielded the following counts. (Results accounting for > 1% are annotated for clarity. Beyond that, use the source, Luke.)

```
101238 range: for ?? := range ?? {
31705 counting loop min 0, max non-literal, stride 1: for i := 0; i < ??; i++ {
14508 bare for: for {
12814 cond only: for ?? {
7235 cond not < or <=: for ??; not LT[E]; ?? {
5472 counting loop min 0, max literal, stride 1: for i := 0; i < N; i++ {
4616 counting loop min non-literal, max non-literal, stride 1: for i := ??; i < ??; i++ {
1728 missing init
1535 counting loop min const, max non-const, stride 1
1508 missing post
 908 counting loop min const, max const, stride 1
 671 counting loop min non-const, max const, stride 1
 605 init multiple values
 554 counting loop min 0, max non-const, stride const
 236 cond lhs not identifier
 236 counting loop min 0, max non-const, stride non-const
 227 post assign not += or -= (but might be i = i - 1, oh well)
 191 counting loop min 0, max const, stride const
 162 counting loop min non-const, max non-const, stride non-const
 103 counting loop min non-const, max non-const, stride const
  89 counting loop min const, max const, stride const
  78 counting loop min const, max non-const, stride const
  76 init lhs != cond lhs
  59 counting loop min non-const, max const, stride const
  26 post assign multiple values
  19 counting loop min const, max non-const, stride non-const
  17 init not i := n
  15 counting loop min const, max const, stride non-const
  13 init lhs != post incdec lhs
   6 counting loop min 0, max const, stride non-const
   4 counting loop min non-const, max const, stride non-const
   3 init lhs != post assign lhs
```

## Acknowledgements

* Thanks to godoc.org for providing a lovely index for scraping bunches of Go.
* Thanks for @kr for `github.com/kr/fs`, vendored here.
* Thanks to Brad Fitzpatrick for encouraging me to gather this data.
