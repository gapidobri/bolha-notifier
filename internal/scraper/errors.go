package scraper

type ErrNoPosts struct{}

func (ErrNoPosts) Error() string {
	return "no posts found"
}
