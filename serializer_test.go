package serializer

import (
	"testing"

	"google.golang.org/appengine/datastore"
)

type Post struct {
	ID       string     `datastore:"-"`
	Text     string     `datastore:"text,noindex"`
	Comments []*Comment `datastore:"-" serializer:"relation,comments"`
}

type Comment struct {
	ID   string `datastore:"-"`
	Text string `datastore:"text,noindex"`
}

func TestSaveEntity(t *testing.T) {
	post := &Post{
		ID:   "ag9zfmRhdGFzdG9yZS1rZXlyNQsSEk15UGFyZW50RW50aXR5VHlwZSIIZXVyb3BlLTQMCxIPTXlTdWJFbnRpdHlUeXBlGCoM",
		Text: "My post",
	}

	comment := &Comment{
		ID:   "ag9zfmRhdGFzdG9yZS1rZXlyNQsSEk15UGFyZW50RW50aXR5VHlwZSIIZXVyb3BlLTQMCxIPTXlTdWJFbnRpdHlUeXBlGCoM",
		Text: "My comment",
	}
	post.Comments = append(post.Comments, comment)

	_, err := SaveEntity(post)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
}

func TestLoadEntity(t *testing.T) {
	key := &datastore.Key{}
	encodedKey := key.Encode()

	post := &Post{
		Text: "My post",
	}
	comment := &Comment{
		ID:   encodedKey,
		Text: "My comment",
	}
	post.Comments = append(post.Comments, comment)

	serialized := datastore.PropertyList{
		datastore.Property{
			Name:     "text",
			Value:    "My post",
			NoIndex:  true,
			Multiple: false,
		},
		datastore.Property{
			Name:     "comments.text",
			Value:    "My comment",
			NoIndex:  true,
			Multiple: true,
		},
		datastore.Property{
			Name:     "comments.id",
			Value:    key,
			NoIndex:  false,
			Multiple: true,
		},
	}

	deserializedPost := &Post{}
	err := LoadEntity(deserializedPost, serialized)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
}
