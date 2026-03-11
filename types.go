package main

type Commentary struct {
	Author string
	Text   string
	URL    string
	Source string
}

type Discussion struct {
	Title  string
	URL    string
	Source string
}

type Headline struct {
	Title      string
	URL        string
	Source     string
	Time       string
	Discussion []Discussion
	Commentary []Commentary
}

type Feed struct {
	Headlines []Headline
}
