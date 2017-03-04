package sites

type Site interface {
	Create(id int) *Site
	Parse()
}
