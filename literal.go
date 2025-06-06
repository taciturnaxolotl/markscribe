package main

import "github.com/taciturnaxolotl/markscribe/literal"

func literalClubCurrentlyReading(count int) []literal.Book {
	books, err := literal.CurrentlyReading()
	if err != nil {
		panic(err)
	}
	if len(books) > count {
		return books[:count]
	}
	return books
}
