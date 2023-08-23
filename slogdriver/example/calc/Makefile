.PHONY: gen example
gen:
	goa gen calc/design

example:
	rm -rf cmd
	rm calc.go
	goa example calc/design
